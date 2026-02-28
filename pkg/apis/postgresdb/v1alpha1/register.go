package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	GroupName    = "bashketchum02.github.io"
	GroupVersion = "v1alpha1"
)

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: GroupVersion}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func addKnownTypes(scheme *runtime.Scheme) error {
	newDB := PostgresDB{}
	newDBList := PostgresDBList{}
	scheme.AddKnownTypes(SchemeGroupVersion, &newDB, &newDBList)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
