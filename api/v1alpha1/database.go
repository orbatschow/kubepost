package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Database is the Schema for the databases API
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSpec   `json:"spec"`
	Status DatabaseStatus `json:"status,omitempty"`
}

type DatabaseSpec struct {
	InstanceRef  InstanceRef `json:"instanceRef"`
	DatabaseName string      `json:"databaseName"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=true
	PreventDeletion bool `json:"preventDeletion"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=false
	// TODO: aso
	CascadeDelete bool `json:"cascadeDelete"`
	//+kubebuilder:validation:Optional
	Extensions []Extension `json:"extensions"`
}

type Extension struct {
	Name string `json:"name"`
	//+kubebuilder:default:=latest
	Version string `json:"version,omitempty"`
}

type DatabaseStatus struct {
	Status string `json:"status"`
}

//+kubebuilder:object:root=true

// DatabaseList contains a list of Database
type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Database `json:"items"`
}
