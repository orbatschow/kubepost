package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InstanceSpec defines the desired state of Instance
type InstanceSpec struct {
	// Host of the PostgreSQL instance
	Host string `json:"host"`
	// Port of the PostgreSQL instance
	Port int `json:"port"`
	// Database of the PostgreSQL instance. This database is used by kubepost to connect to the PostgreSQL instance.
	Database string `json:"database"`
	// Kubernetes secret reference for the username, that will be used by kubepost to connect to the PostgreSQL instance.
	Username *v1.SecretKeySelector `json:"username"`
	// Kubernetes secret reference for the password, that will be used by kubepost to connect to the PostgreSQL instance.
	Password *v1.SecretKeySelector `json:"password"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=prefer
	// Connection mode that kubepost will use to connect to the instance.
	SSLMode string `json:"sslMode"`
}

// InstanceStatus defines the observed state of Instance
type InstanceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Instance is the Schema for the instances API
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec,omitempty"`
	Status InstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InstanceList contains a list of Instance
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}
