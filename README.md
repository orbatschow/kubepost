# kubepost

**Project status: *beta*** Not all planned features are completed. The API, spec, status and other user facing objects
may change, but in a backward compatible way.

## Overview

The kubepost operator provides [Kubernetes](https://kubernetes.io/) native deployment and management of
[PostgreSQL](https://www.postgresql.org/) objects. The purpose of this project is to
simplify and automate the configuration of PostgreSQL objects.

The kubepost operator implements, but is not limited to, the following features:

* **Role**: Manage PostgreSQL [roles](https://www.postgresql.org/docs/current/user-manag.htm).

* **Database**: Manage PostgresSQL [databases](https://www.postgresql.org/docs/current/managing-databases.html)

* **Extensions** Manage PostgreSQL [extensions](https://www.postgresql.org/docs/current/external-extensions.html)

## Prerequisites

**Note:** This compatibility matrix is at the moment not tested, it is only provided based on personal experience.
Therefore, it may also be possible, that other combinations are working properly.

| Kubernetes | PostgreSQL | kubepost |
|------------|------------|----------|
| 1.24       | 12,13,14   | 1.0.0    |
| 1.25       | 12,13,14   | 1.0.0    |

## CustomResourceDefinitions

A core feature of kubepost is to monitor the Kubernetes API server for changes
to specific resources and ensure that the desired PostgreSQL match these resources.

* **`Instance`**, which defines one or multiple PostgreSQL instances, that shall be managed by kubepost.

* **`Role`**, which defines a PostgreSQL role, that shall be managed by kubepost.

* **`Database`**, which defines a PostgreSQL database, that shall be managed by kubepost.

The Prometheus operator automatically detects changes in the Kubernetes API server to any of the above objects, and
ensures that matching PostgreSQL objects are kept in sync.

To learn more about the CRDs introduced by kubepost have a look at the [specification](docs/getting-started.md).

## Quickstart

**Note:** this quickstart does provision the kubepost stack, required to access all features of kubepost.

```sh
kubectl apply -f deploy/bundle.yaml
```

> Note: `deploy/bundle.yaml` may be unstable, if you plan to run kubepost in production please use a tagged release.

Tagged versions can be installed using the following command:

```sh
kubectl apply -f deploy/bundle.yaml TODO
```

> Note: The tag used above might not be pointing to the latest release. Check the git tags within this repository to
> get the latest tag.

## Removal

To remove the operator, first delete any custom resources you created in each namespace.

```sh
for n in $(kubectl get namespaces -o jsonpath={..metadata.name}); do
  kubectl delete --all --namespace=$n roles.postgres.kubepost.io,instances.postgres.kubepost.io,databases.postgres.kubepost.io
done
```

After a couple of minutes you can go ahead and remove the operator itself.

```sh
kubectl delete -f deploy/bundle.yaml
```

## Contributing

Contributions are always welcome, have a look at the [contributing](docs/contributing.md) guidelines to get started.


## Troubleshooting

Before creating a new issue please check the whole [documentation](docs).