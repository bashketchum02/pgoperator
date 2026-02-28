package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

var _ runtime.Object = &PostgresDB{}
var _ runtime.Object = &PostgresDBList{}

type PostgresDB struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec PostgresDBSpec `json:"spec,omitempty"`
	Status PostgresDBStatus `json:"status,omitempty"`
}

type PostgresDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items []PostgresDB `json:"items"`
}

type BackupSpec struct {
	Enabled bool `json:"enabled,omitempty"`
	Schedule string `json:"schedule,omitempty"`
	RetentionDays int32 `json:"retentionDays,omitempty"`
}

type PostgresDBSpec struct {
	Version string `json:"version"`
	Replicas int32 `json:"replicas"`
	Storage string `json:"storage"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	Backup BackupSpec `json:"backup,omitempty"`
}

type PostgresDBStatus struct {
	Phase string `json:"phase,omitempty"`
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
	LastBackupTime *metav1.Time `json:"lastBackupTime,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

func (in *PostgresDB) DeepCopyObject() runtime.Object {
	out := new(PostgresDB)
	*out = *in
	out.ObjectMeta = *in.ObjectMeta.DeepCopy()
	return out
}

func (in *PostgresDBList) DeepCopyObject() runtime.Object {
	out := new(PostgresDBList)
	*out = *in
	out.ListMeta = *in.ListMeta.DeepCopy()
	return out
}
