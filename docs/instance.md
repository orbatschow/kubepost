# API Reference

Packages:

- [postgres.kubepost.io/v1alpha1](#postgreskubepostiov1alpha1)

# postgres.kubepost.io/v1alpha1

Resource Types:

- [Instance](#instance)




## Instance
<sup><sup>[↩ Parent](#postgreskubepostiov1alpha1 )</sup></sup>






Instance is the Schema for the instances API

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
      <td>Instance</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#instancespec">spec</a></b></td>
        <td>object</td>
        <td>
          InstanceSpec defines the desired state of Instance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>object</td>
        <td>
          InstanceStatus defines the observed state of Instance<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Instance.spec
<sup><sup>[↩ Parent](#instance)</sup></sup>



InstanceSpec defines the desired state of Instance

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
          Database of the PostgreSQL instance. This database is used by kubepost to connect to the PostgreSQL instance.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          Host of the PostgreSQL instance<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#instancespecpassword">password</a></b></td>
        <td>object</td>
        <td>
          Kubernetes secret reference for the password, that will be used by kubepost to connect to the PostgreSQL instance.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          Port of the PostgreSQL instance<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#instancespecusername">username</a></b></td>
        <td>object</td>
        <td>
          Kubernetes secret reference for the username, that will be used by kubepost to connect to the PostgreSQL instance.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>sslMode</b></td>
        <td>string</td>
        <td>
          Connection mode that kubepost will use to connect to the instance.<br/>
          <br/>
            <i>Default</i>: prefer<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Instance.spec.password
<sup><sup>[↩ Parent](#instancespec)</sup></sup>



Kubernetes secret reference for the password, that will be used by kubepost to connect to the PostgreSQL instance.

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


### Instance.spec.username
<sup><sup>[↩ Parent](#instancespec)</sup></sup>



Kubernetes secret reference for the username, that will be used by kubepost to connect to the PostgreSQL instance.

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