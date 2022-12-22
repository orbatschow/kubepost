# API Reference

Packages:

- [postgres.kubepost.io/v1alpha1](#postgreskubepostiov1alpha1)

# postgres.kubepost.io/v1alpha1

Resource Types:

- [Role](#role)




## Role
<sup><sup>[↩ Parent](#postgreskubepostiov1alpha1 )</sup></sup>






Role is the Schema for the roles API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>postgres.kubepost.io/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Role</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#rolespec">spec</a></b></td>
        <td>object</td>
        <td>
          RoleSpec defines the desired state of Role<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>object</td>
        <td>
          RoleStatus defines the observed state of Role<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Role.spec
<sup><sup>[↩ Parent](#role)</sup></sup>



RoleSpec defines the desired state of Role

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#rolespecconnectionnamespaceselector">connectionNamespaceSelector</a></b></td>
        <td>object</td>
        <td>
          Narrow down the namespaces for the previously matched connections.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#rolespecconnectionselector">connectionSelector</a></b></td>
        <td>object</td>
        <td>
          Define which connections shall be used by kubepost for this role.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>cascadeDelete</b></td>
        <td>boolean</td>
        <td>
          TODO<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#rolespecgrantsindex">grants</a></b></td>
        <td>[]object</td>
        <td>
          Grants that shall be applied to this role.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#rolespecgroupsindex">groups</a></b></td>
        <td>[]object</td>
        <td>
          Groups that shall be applied to this role.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>options</b></td>
        <td>[]string</td>
        <td>
          Options that shall be applied to this role. Important: Options that are simply removed from the kubepost role will not be removed from the PostgreSQL role. E.g.: Granting "SUPERUSER" and then removing the option won't cause kubepost to remove this option from the role. You have to set the option "NOSUPERUSER".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#rolespecpassword">password</a></b></td>
        <td>object</td>
        <td>
          Kubernetes secret reference, that is used to set a password for the role.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>preventDeletion</b></td>
        <td>boolean</td>
        <td>
          TODO<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Role.spec.connectionNamespaceSelector
<sup><sup>[↩ Parent](#rolespec)</sup></sup>



Narrow down the namespaces for the previously matched connections.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#rolespecconnectionnamespaceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Role.spec.connectionNamespaceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#rolespecconnectionnamespaceselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Role.spec.connectionSelector
<sup><sup>[↩ Parent](#rolespec)</sup></sup>



Define which connections shall be used by kubepost for this role.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#rolespecconnectionselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Role.spec.connectionSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#rolespecconnectionselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values. Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Role.spec.grants[index]
<sup><sup>[↩ Parent](#rolespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>database</b></td>
        <td>string</td>
        <td>
          Define which database shall the grant be applied to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#rolespecgrantsindexobjectsindex">objects</a></b></td>
        <td>[]object</td>
        <td>
          Define the granular grants within the database.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Role.spec.grants[index].objects[index]
<sup><sup>[↩ Parent](#rolespecgrantsindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>identifier</b></td>
        <td>string</td>
        <td>
          Name of the PostgreSQL object (VIEW;COLUMN;TABLE;SCHEMA;FUNCTION;SEQUENCE) that the grant shall be applied to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Define the type that the grant shall be applied to.<br/>
          <br/>
            <i>Enum</i>: VIEW, COLUMN, TABLE, SCHEMA, FUNCTION, SEQUENCE<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>privileges</b></td>
        <td>[]enum</td>
        <td>
          Define the privileges for the grant.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>schema</b></td>
        <td>string</td>
        <td>
          Define the schema that the grant shall be applied to.<br/>
          <br/>
            <i>Default</i>: public<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>table</b></td>
        <td>string</td>
        <td>
          TODO<br/>
          <br/>
            <i>Default</i>: ''<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>withGrantOption</b></td>
        <td>boolean</td>
        <td>
          Define whether the `WITH GRANT OPTION` shall be granted. More information can be found within the official [PostgreSQL](https://www.postgresql.org/docs/current/sql-grant.html) documentation.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Role.spec.groups[index]
<sup><sup>[↩ Parent](#rolespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Define the name of the group.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>withAdminOption</b></td>
        <td>boolean</td>
        <td>
          Define whether the `WITH ADMIN OPTION` shall be granted. More information can be found within the official [PostgreSQL](https://www.postgresql.org/docs/current/sql-grant.html) documentation.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Role.spec.password
<sup><sup>[↩ Parent](#rolespec)</sup></sup>



Kubernetes secret reference, that is used to set a password for the role.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>