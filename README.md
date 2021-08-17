**This is alpha software, use it on your own risk!**

# kubepost Operator

The kubepost Operator manages various Postgres objects via standard Kubernetes CRDs. It requires
[Metacontroller](https://github.com/metacontroller/metacontroller) to be installed within the cluster.

## Features

- Role lifecycle management
- Database lifecycle management
- Database extension lifecycle management
- Permission lifecycle management

## PostgreSQL

At the moment only Postgres 13 is supported, backwards compatibility might be added at a later stage.

## Installation

### Metacontroller

kubepost uses Metacontroller under the hood, so this component hast to be installed within the cluster. You can view
detailed instructions for the installation
process [here](https://metacontroller.github.io/metacontroller/guide/install.html).

Create a new kustomization file with the following content:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: metacontroller

resources:
  - github.com/metacontroller/metacontroller/manifests//production
```

And apply it to the desired cluster:

```shell
kustomize build . | kubectl apply -f -
```

### kubepost

After you have installed metacontroller you have to deploy kubepost. Create a new kustomization file with the following
content:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: metacontroller

resources:
  - github.com/orbatschow/kubepost//manifests
```

And apply it to the desired cluster:

```shell
kustomize build . | kubectl apply -f -
```

## Usage

### Instance

The instance is used by other CRDs to connect to the desired database instance. It allows a clear segregation between
roles, databases and the instance itself. To connect to the databse it uses a secret that should be available within the
Kubernetes cluster.

```yaml
apiVersion: kubepost.io/v1alpha1
kind: Instance
metadata:
  name: kubepost # this is the instanceRef, used by other CRDs
spec:
  host: localhost
  port: 5432
  database: postgres
  secretRef:
    name: kubepost-instance-credentials
    userKey: username
    passwordKey: password
```

The corresponding secret might look like this:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kubepost-instance-credentials
data:
  username: cG9zdGdyZXM=
  password: cm9vdA==

```

### Database

This database uses the previously mentioned instance CRD to connect to the database instance and creates a database with
the name `kubepost`. After the database is created the extension `pg_stat_statements` will be installed.

```yaml
apiVersion: kubepost.io/v1alpha1
kind: Database
metadata:
  name: kubepost
spec:
  databaseName: kubepost
  databaseOwner: kubepost
  preventDeletion: false
  instanceRef:
    name: kubepost
  extensions:
    - name: pg_stat_statements
      version: "1.8"
      # if no version is specified, latest will be used
    - name: postgres_fdw
```

### Role

This role uses the previously mentioned instance CRD to connect to the database instance and creates a role with the
name `kubepost`. It then grants this role `ALL PRIVILEGES` on schema `public` in database`kubepost`. The grant section
is optional.

```yaml
apiVersion: kubepost.io/v1alpha1
kind: Role
metadata:
  name: kubepost
spec:
  roleName: kubepost
  preventDeletion: false
  passwordRef:
    name: kubepost-role-credentials
    passwordKey: password
  instanceRef:
    name: kubepost
  options:
    - SUPERUSER
    - LOGIN
  grants:
    # This field specifies the database to which kubepost will connect for all following grants.
    - database: kubepost

      # This is an array of database-objects and belonging user previliges.
      objects:
         # the identifiert can be chosen like the corresponding identifier in postgres
         # for example one table with schema: public.test_table
        - identifier: public
          
          # possible options: ["TABLE", "SCHEMA", "FUNCTION", "SEQUENCE", "ROLE"]
          # SCHEMA will result in an 'GRANT PREVILIGES TO ALL TABLES IN SCHEMA'
          # every other option will result in GRANT-Querys similar to:
          # https://www.postgresql.org/docs/current/sql-grant.html
          type: SCHEMA
          
          # possible options: ["ALL", "INSERT", "SELECT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"]
          privileges: ["ALL"]
          
          withGrantOption: true
```
