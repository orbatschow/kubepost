# Getting started

The goal of kubepost is to make PostgreSQL management on top of Kubernetes
as easy as possible, while preserving Kubernetes-native configuration options.

This guide will show you how to deploy the kubepost operator, set up a
PostgreSQL cluster and configure a `Connection`, that can be used to manage
`Roles` and `Databases`.

# Pre-requisites

To follow this guide, you will need a Kubernetes cluster with admin permissions and a running
PostgreSQL cluster with superuser permissions.

## Kubernetes

You can use [kind](https://kind.sigs.k8s.io/) to create a fully functional Kubernetes cluster.

## PostgreSQL

In order to complete this guide you need an already existing PostgreSQL cluster. You can use this
command in order to create a fully functional cluster:

```sh
kubectl apply -k config/samples/postgres
```

# Installing the operator

The first step is to install the operator's Custom Resource Definitions (CRDs) as well
as the operator itself with the required RBAC resources.

Run the following commands to install the CRDs and deploy the operator:

```sh
kubectl apply -f deploy/bundle.yaml
```

The kubepost operator introduces custom resources in Kubernetes to declare
the desired state of PostgreSQL clusters as well as its configuration.

# Custom Resource Definitions

## Connection

The `Connection` resource declaratively describes the connection details for one or more PostgreSQL
clusters. Those clusters are not managed by kubepost and therefore have to exist already.

First, let's create a new Kubernetes secret, that will hold confidential data regarding the connection details:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: connection-credentials
type: Opaque
stringData:
  username: postgres
  password: postgres
```

> If you haven't used the sample postgres cluster provided beforehand you have to modify this secret to reflect
the actual credentials used for your PostgreSQL cluster.

After we have provided our secret to the cluster we can create our first kubepost resource:

```yaml
apiVersion: postgres.kubepost.io/v1alpha1
kind: Connection
metadata:
  name: default
  namespace: default
  labels:
    instance: default
spec:
  host: postgres.svc.cluster.local
  port: 5432
  database: postgres
  username:
    name: connection-credentials
    key: username
  password:
    name: connection-credentials
    key: password
```

> The provided user will be used by kubepost to log into the database and manage roles and databases. Please
> ensure, that the user has all necessary permissions. If you are unsure regarding the permissions you can start
> with a superuser and gradually remove permissions.

This will create a new kubepost `Connection`, that can be used at a later stage. Note that every connection
requires a label to be useful. Whenever we create another kubepost resource at a later point we can reference
the connection above via the label and kubepost will connect to the configured PostgreSQL cluster.

A more detailed specification of the `Connection` resource can be found within the [connection](connection.md)
documentation.

## Role

You can use the `Role` resource in order to manage PostgreSQL roles.

```yaml
apiVersion: postgres.kubepost.io/v1alpha1
kind: Role
metadata:
  name: kubepost
  namespace: default
spec:
  connectionSelector:
    matchLabels:
      instance: default
  connectionNamespaceSelector:
    matchLabels:
      kubernetes.io/metadata.name: default
  cascadeDelete: false
  grants:
    - database: kubepost
      objects:
        - identifier: public
          privileges:
            - ALL
          schema: public
          type: SCHEMA
          withGrantOption: true
  options:
    - SUPERUSER
  preventDeletion: false
```

This resource will cause kubepost to search for an connection with label `default` within all namespaces,
that have an assigned label `default`. For all matching connections it will grab the connection details, connect
to the `postgres` database and create the role. After creating the role, kubepost will check if the desired
permissions are equal to the current ones. If there are differences kubepost will try to resolve those issues
and grant/revoke the differences.

> **Note*:* There are situations, where kubepost won't be able to resolve conflicts. For example removing a role,
> that still owns a database will cause kubepost to fail. The operator will log these errors and you can remove the
> database beforehand.

A more detailed specification of the `Role` resource can be found within the [role](role.md) documentation.

## Database

You can use the `Database` resource in order to manage PostgreSQL databases.

```yaml
apiVersion: postgres.kubepost.io/v1alpha1
kind: Database
metadata:
  name: kubepost
spec:
  connectionSelector:
    matchLabels:
      instance: default
  connectionNamespaceSelector:
    matchLabels:
      kubernetes.io/metadata.name: default
  owner: kubepost
  extensions:
    - name: pg_stat_statements
      version: "1.7"
    - name: postgres_fdw
```

This resource will cause kubepost to search for a connection with label `default` within all namespaces,
that have an assigned label `default`. For all matching connections it will grab the connection details, connect
to the `postgres` database and create the database and extensions.

A more detailed specification of the `Database` resource can be found within the [database](database.md) documentation.
