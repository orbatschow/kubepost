package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DatabaseSpec struct {
	// Define which instances shall be managed by kubepost for this database.
	InstanceSelector metav1.LabelSelector `json:"instanceSelector"`
	// Narrow down the namespaces for the previously matched instances.
	InstanceNamespaceSelector metav1.LabelSelector `json:"instanceNamespaceSelector"`

	//+kubebuilder:validation:Optional
	// Define the owner of the database.
	Owner string `json:"owner"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=true
	// TODO
	PreventDeletion bool `json:"preventDeletion"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=false
	// TODO
	CascadeDelete bool `json:"cascadeDelete"`
	//+kubebuilder:validation:Optional
	// List of extensions for this database.
	Extensions []Extension `json:"extensions"`
}

type Extension struct {
	// Name of the extensions that shall be managed within the database.
	Name string `json:"name"`
	//+kubebuilder:default:=latest
	//+kubebuilder:validation:Optional
	// Version of the extension.
	Version string `json:"version,omitempty"`
}

// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Database is the Schema for the databases API
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSpec   `json:"spec,omitempty"`
	Status DatabaseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DatabaseList contains a list of Database
type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Database `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Database{}, &DatabaseList{})
}
