package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RoleSpec defines the desired state of Role
type RoleSpec struct {
	InstanceSelector          metav1.LabelSelector `json:"instanceSelector"`
	InstanceNamespaceSelector metav1.LabelSelector `json:"instanceNamespaceSelector"`

	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=true
	PreventDeletion bool `json:"preventDeletion"`

	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=false
	CascadeDelete bool `json:"cascadeDelete"`

	//+kubebuilder:validation:Optional
	Options []string `json:"options"`

	//+kubebuilder:validation:Optional
	Password *v1.SecretKeySelector `json:"password"`

	//+kubebuilder:validation:Optional
	Grants []Grant `json:"grants"`

	//+kubebuilder:validation:Optional
	Groups []GroupGrantObject `json:"groups"`
}

type Grant struct {
	Database string        `json:"database"`
	Objects  []GrantObject `json:"objects"`
}

type GroupGrantObject struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	//+kubebuilder:default:=false
	WithAdminOption bool `json:"withAdminOption"`
}

type GrantObject struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=VIEW;COLUMN;TABLE;SCHEMA;FUNCTION;SEQUENCE
	Type string `json:"type"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=public
	Schema string `json:"schema"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=''
	Table string `json:"table"`

	// +kubebuilder:validation:Required
	Identifier string `json:"identifier"`

	// +kubebuilder:validation:Optional
	Privileges []Privilege `json:"privileges"`
	//+kubebuilder:validation:Optional
	WithGrantOption bool `json:"withGrantOption"`
}

// +kubebuilder:validation:Enum=ALL;SELECT;INSERT;UPDATE;DELETE;TRUNCATE;REFERENCES;TRIGGER;USAGE;CREATE;CONNECT;TEMPORARY;TEMP;EXECUTE
type Privilege string

// RoleStatus defines the observed state of Role
type RoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Role is the Schema for the roles API
type Role struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoleSpec   `json:"spec,omitempty"`
	Status RoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RoleList contains a list of Role
type RoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Role `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Role{}, &RoleList{})
}
