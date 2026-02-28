package postgresdb

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/apranto/pgoperator/pkg/apis/postgresdb/v1alpha1"
)

func buildSecret(pgdb *v1alpha1.PostgresDB, existingPassword string) *corev1.Secret {
	password := existingPassword
	if password == "" {
		password = generatePassword(16)
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-credentials", pgdb.Name),
			Namespace: pgdb.Namespace,
			Labels:    buildLabels(pgdb.Name),
			OwnerReferences: []metav1.OwnerReference{
				buildOwnerReference(pgdb),
			},
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": password,
			"POSTGRES_DB":       pgdb.Name,
		},
	}
}

func generatePassword(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "changeme-fallback-password"
	}
	return hex.EncodeToString(bytes)
}

func buildLabels(name string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "postgresql",
		"app.kubernetes.io/instance":   name,
		"app.kubernetes.io/managed-by": "pgoperator",
	}
}

func buildOwnerReference(pgdb *v1alpha1.PostgresDB) metav1.OwnerReference {
	isController := true
	blockDelete := true
	return metav1.OwnerReference{
		APIVersion:         fmt.Sprintf("%s/%s", v1alpha1.GroupName, v1alpha1.GroupVersion),
		Kind:               "PostgresDB",
		Name:               pgdb.Name,
		UID:                pgdb.UID,
		Controller:         &isController,
		BlockOwnerDeletion: &blockDelete,
	}
}
