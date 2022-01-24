package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Role is the Schema for the roles API
type Role struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoleSpec   `json:"spec"`
	Status RoleStatus `json:"status,omitempty"`
}

type RoleSpec struct {
	InstanceRef InstanceRef `json:"instanceRef"`
	RoleName    string      `json:"roleName"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=true
	PreventDeletion bool `json:"preventDeletion"`
	//+kubebuilder:validation:Optional
	//+kubebuilder:default:=false
	CascadeDelete bool `json:"cascadeDelete"`
	//+kubebuilder:validation:Optional
	Options []string `json:"options"`
	//+kubebuilder:validation:Optional
	Password string `json:"password"`
	//+kubebuilder:validation:Optional
	PasswordRef PasswordRef `json:"passwordRef"`
	//+kubebuilder:validation:Optional
	Groups []GroupGrantObject `json:"groups"`
	//+kubebuilder:validation:Optional
	Grants []Grant `json:"grants"`
}

type PasswordRef struct {
	Name        string `json:"name"`
	PasswordKey string `json:"passwordKey"`
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

type RoleStatus struct {
	Status string `json:"status"`
}

//+kubebuilder:object:root=true

// RoleList contains a list of Role
type RoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Role `json:"items"`
}
