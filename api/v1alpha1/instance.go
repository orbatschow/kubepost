package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Instance is the Schema for the instances API
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec"`
	Status InstanceStatus `json:"status,omitempty"`
}

type InstanceSpec struct {
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Database  string    `json:"database"`
	SecretRef SecretRef `json:"secretRef"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=prefer
	SSLMode string `json:"sslMode"`
}

type SecretRef struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	UserKey     string `json:"userKey"`
	PasswordKey string `json:"passwordKey"`
}

type InstanceStatus struct {
	Status string `json:"status"`
}

type InstanceRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

//+kubebuilder:object:root=true

// RoleList contains a list of Role
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}
