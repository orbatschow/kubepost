package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleSpec defines the desired state of Role
type RoleSpec struct {
	// Define which connections shall be used by kubepost for this role.
	ConnectionSelector metav1.LabelSelector `json:"connectionSelector"`
	// Narrow down the namespaces for the previously matched connections.
	ConnectionNamespaceSelector metav1.LabelSelector `json:"connectionNamespaceSelector"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=true
	// Define whether the PostgreSQL role deletion is skipped when the CR is deleted.
	Protected bool `json:"protected"`

	// +kubebuilder:validation:Optional
	// Options that shall be applied to this role. Important: Options that are simply removed from the kubepost role
	// will not be removed from the PostgreSQL role.
	// E.g.: Granting "SUPERUSER" and then removing the option won't cause kubepost to remove this option from
	// the role. You have to set the option "NOSUPERUSER".
	Options []string `json:"options"`

	// +kubebuilder:validation:Optional
	// Kubernetes secret reference, that is used to set a password for the role.
	Password *v1.SecretKeySelector `json:"password"`

	// +kubebuilder:validation:Optional
	// Grants that shall be applied to this role.
	Grants []Grant `json:"grants"`

	// +kubebuilder:validation:Optional
	// Groups that shall be applied to this role.
	Groups []GroupGrantObject `json:"groups"`
}

type Grant struct {
	// Define which database shall the grant be applied to.
	Database string `json:"database"`
	// Define the granular grants within the database.
	Objects []GrantObject `json:"objects"`
}

type GroupGrantObject struct {
	// +kubebuilder:validation:Required
	// Define the name of the group.
	Name string `json:"name"`
	// +kubebuilder:default:=false
	// Define whether the `WITH ADMIN OPTION` shall be granted. More information can be found within
	// the official [PostgreSQL](https://www.postgresql.org/docs/current/sql-grant.html) documentation.
	WithAdminOption bool `json:"withAdminOption"`
}

type GrantObject struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=VIEW;COLUMN;TABLE;SCHEMA;FUNCTION;SEQUENCE
	// Define the type that the grant shall be applied to.
	Type string `json:"type"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=public
	// Define the schema that the grant shall be applied to.
	Schema string `json:"schema"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=''
	// TODO
	Table string `json:"table"`

	// +kubebuilder:validation:Required
	// Name of the PostgreSQL object (VIEW;COLUMN;TABLE;SCHEMA;FUNCTION;SEQUENCE) that the grant shall be applied to.
	Identifier string `json:"identifier"`

	// +kubebuilder:validation:Optional
	// Define the privileges for the grant.
	Privileges []Privilege `json:"privileges"`
	// +kubebuilder:validation:Optional
	// Define whether the `WITH GRANT OPTION` shall be granted. More information can be found within
	// the official [PostgreSQL](https://www.postgresql.org/docs/current/sql-grant.html) documentation.
	WithGrantOption bool `json:"withGrantOption"`
}

// +kubebuilder:validation:Enum=ALL;SELECT;INSERT;UPDATE;DELETE;TRUNCATE;REFERENCES;TRIGGER;USAGE;CREATE;CONNECT;TEMPORARY;TEMP;EXECUTE

type Privilege string

// RoleStatus defines the observed state of Role
type RoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Role is the Schema for the roles API
type Role struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoleSpec   `json:"spec,omitempty"`
	Status RoleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RoleList contains a list of Role
type RoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Role `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Role{}, &RoleList{})
}
