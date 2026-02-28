package postgresdb

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	v1alpha1 "github.com/apranto/pgoperator/pkg/apis/postgresdb/v1alpha1"
)

type Reconciler struct {
	client kubernetes.Interface
}

func NewReconciler(client kubernetes.Interface) *Reconciler {
	return &Reconciler{client: client}
}

func (r *Reconciler) Reconcile(ctx context.Context, pgdb *v1alpha1.PostgresDB) error {
	klog.Infof("Reconciling PostgresDB %s/%s (version=%s, replicas=%d)",
		pgdb.Namespace, pgdb.Name, pgdb.Spec.Version, pgdb.Spec.Replicas)

	if err := r.ensureSecret(ctx, pgdb); err != nil {
		return fmt.Errorf("failed to ensure secret: %w", err)
	}

	if err := r.ensureHeadlessService(ctx, pgdb); err != nil {
		return fmt.Errorf("failed to ensure headless service: %w", err)
	}

	if err := r.ensureClientService(ctx, pgdb); err != nil {
		return fmt.Errorf("failed to ensure client service: %w", err)
	}
	if err := r.ensureStatefulSet(ctx, pgdb); err != nil {
		return fmt.Errorf("failed to ensure stateful set: %w", err)
	}
	klog.Infof("Successfully reconciled PostgresDB %s/%s", pgdb.Namespace, pgdb.Name)
	return nil
}

func (r *Reconciler) ensureSecret(ctx context.Context, pgdb *v1alpha1.PostgresDB) error {
	secretName := fmt.Sprintf("%s-credentials", pgdb.Name)

	existing, err := r.client.CoreV1().Secrets(pgdb.Namespace).Get(ctx, secretName, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		desired := buildSecret(pgdb, "")
		klog.Infof("Creating Secret %s/%s", pgdb.Namespace, secretName)
		_, err = r.client.CoreV1().Secrets(pgdb.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create secret: %w", err)
		}
		klog.Infof("Created Secret %s/%s", pgdb.Namespace, secretName)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}

	klog.V(2).Infof("Secret %s/%s already exists", pgdb.Namespace, secretName)
	_ = existing // We could check/update labels here if needed
	return nil
}

func (r *Reconciler) ensureHeadlessService(ctx context.Context, pgdb *v1alpha1.PostgresDB) error {
	serviceName := fmt.Sprintf("%s-headless", pgdb.Name)
	return r.ensureService(ctx, pgdb, serviceName, buildHeadlessService(pgdb))
}

func (r *Reconciler) ensureClientService(ctx context.Context, pgdb *v1alpha1.PostgresDB) error {
	return r.ensureService(ctx, pgdb, pgdb.Name, buildClientService(pgdb))
}

func (r *Reconciler) ensureService(ctx context.Context, pgdb *v1alpha1.PostgresDB, name string, desired *corev1.Service) error {
	_, err := r.client.CoreV1().Services(pgdb.Namespace).Get(ctx, name, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		klog.Infof("Creating Service %s/%s", pgdb.Namespace, name)
		_, err = r.client.CoreV1().Services(pgdb.Namespace).Create(ctx, desired, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create service %s: %w", name, err)
		}
		klog.Infof("Created Service %s/%s", pgdb.Namespace, name)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get service %s: %w", name, err)
	}

	klog.V(2).Infof("Service %s/%s already exists", pgdb.Namespace, name)
	return nil
}

func (r *Reconciler) ensureStatefulSet(ctx context.Context, pgdb *v1alpha1.PostgresDB) error {
	desired := buildStatefulSet(pgdb)
	existing, err := r.client.AppsV1().StatefulSets(pgdb.Namespace).Get(
		ctx, pgdb.Name, metav1.GetOptions{},
	)
	if errors.IsNotFound(err){
		klog.Infof("Creating statefulset %s/%s", pgdb.Namespace, pgdb.Name)
		_, err := r.client.AppsV1().StatefulSets(pgdb.Namespace).Create(
			ctx, desired, metav1.CreateOptions{},
		)
		if err != nil {
			return fmt.Errorf("Failed to create statefulset %s/%s: %w", pgdb.Namespace, pgdb.Name, err)
		}
		klog.Infof("Successfully created statefulset %s/%s", pgdb.Namespace, pgdb.Name)
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting statefulsets %s/%s: %w", pgdb.Namespace, pgdb.Name, err)
	}
	
	if *existing.Spec.Replicas != *desired.Spec.Replicas || desired.Spec.Template.Spec.Containers[0].Image != existing.Spec.Template.Spec.Containers[0].Image {
		*existing.Spec.Replicas = *desired.Spec.Replicas
		existing.Spec.Template.Spec.Containers[0].Image = desired.Spec.Template.Spec.Containers[0].Image
		_, err = r.client.AppsV1().StatefulSets(pgdb.Namespace).Update(
			ctx, existing, metav1.UpdateOptions{},
		)
		if err != nil {
			return fmt.Errorf("Error updating statefulsets %s/%s: %w", pgdb.Namespace, pgdb.Name, err)
		}
		klog.Infof("Successfully updated statefulsets %s/%s", pgdb.Namespace, pgdb.Name)
	}

	return nil
}

func (r *Reconciler) updateStatus(ctx context.Context, pgdb *v1alpha1.PostgresDB, sts *appsv1.StatefulSet) {
	_ = sts
}
