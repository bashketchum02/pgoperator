package postgresdb

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	v1alpha1 "github.com/apranto/pgoperator/pkg/apis/postgresdb/v1alpha1"
)

func buildHeadlessService(pgdb *v1alpha1.PostgresDB) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-headless", pgdb.Name),
			Namespace: pgdb.Namespace,
			Labels:    buildLabels(pgdb.Name),
			OwnerReferences: []metav1.OwnerReference{
				buildOwnerReference(pgdb),
			},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone,
			Selector: buildLabels(pgdb.Name),
			Ports: []corev1.ServicePort{
				{
					Name:       "postgresql",
					Port:       5432,
					TargetPort: intstr.FromInt32(5432),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			PublishNotReadyAddresses: true,
		},
	}
}

func buildClientService(pgdb *v1alpha1.PostgresDB) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: pgdb.Name,
			Namespace: pgdb.Namespace,
			Labels: buildLabels(pgdb.Name),
			OwnerReferences: []metav1.OwnerReference{buildOwnerReference(pgdb)},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: buildLabels(pgdb.Name),
			Ports: []corev1.ServicePort{
				{
					Name: "postgresql",
					Port: 5432,
					TargetPort: intstr.FromInt32(5432),
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}	
}
