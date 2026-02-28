package postgresdb

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/apranto/pgoperator/pkg/apis/postgresdb/v1alpha1"
)

func buildStatefulSet(pgdb *v1alpha1.PostgresDB) *appsv1.StatefulSet {
	labels := buildLabels(pgdb.Name)

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pgdb.Name,
			Namespace: pgdb.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{buildOwnerReference(pgdb)},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: fmt.Sprintf("%s-headless", pgdb.Name),
			Replicas:    int32Ptr(pgdb.Spec.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "postgresql",
							Image: fmt.Sprintf("postgres:%s", pgdb.Spec.Version),
							Ports: []corev1.ContainerPort{
								{
									Name:          "postgresql",
									ContainerPort: 5432,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: fmt.Sprintf("%s-credentials", pgdb.Name),
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/var/lib/postgresql/data",
									SubPath:   "pgdata",
								},
							},
							Resources: pgdb.Spec.Resources,
						},
					},
				},
			},

			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "data"},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse(pgdb.Spec.Storage),
							},
						},
					},
				},
			},
		},
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}
