package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConnectionSpec defines the desired state of Connection
type ConnectionSpec struct {
	// Host of the PostgreSQL connection
	Host string `json:"host"`
	// Port of the PostgreSQL connection
	Port int `json:"port"`
	// Database of the PostgreSQL connection. This database is used by kubepost to connect to the PostgreSQL connection.
	Database string `json:"database"`
	// Kubernetes secret reference for the username, that will be used by kubepost to connect to the PostgreSQL connection.
	Username *v1.SecretKeySelector `json:"username"`
	// Kubernetes secret reference for the password, that will be used by kubepost to connect to the PostgreSQL connection.
	Password *v1.SecretKeySelector `json:"password"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=prefer
	// Connection mode that kubepost will use to connect to the connection.
	SSLMode string `json:"sslMode"`
}

// ConnectionStatus defines the observed state of Connection
type ConnectionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Connection is the Schema for the connections API
type Connection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectionSpec   `json:"spec,omitempty"`
	Status ConnectionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConnectionList contains a list of Connection
type ConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Connection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Connection{}, &ConnectionList{})
}
