# kubepost

<img align="right" alt="kubepost" width="180px" src="assets/gopher.png">

<p>
    <a href="https://github.com/orbatschow/kubepost/actions/workflows/default.yaml" target="_blank" rel="noopener"><img src="https://img.shields.io/github/actions/workflow/status/orbatschow/kubepost/default.yaml" alt="build" /></a>
    <a href="https://github.com/orbatschow/kubepost/releases" target="_blank" rel="noopener"><img src="https://img.shields.io/github/release/orbatschow/kubepost.svg" alt="Latest releases" /></a>
    <a href="https://github.com/orbatschow/kubepost/blob/master/LICENSE" target="_blank" rel="noopener"><img src="https://img.shields.io/github/license/orbatschow/kubepost" /></a>
</p>

The kubepost operator provides [Kubernetes](https://kubernetes.io/) native deployment and management of
<a href="https://www.postgresql.org/">PostgreSQL</a> objects. The purpose of this project is to
simplify and automate the configuration of PostgreSQL objects.

**Project status: *beta*** Not all planned features are completed. The API, spec, status and other user facing objects
may change, but in a backward compatible way.

## Features

The kubepost operator implements, but is not limited to, the following features:

* **Role**: Manage PostgreSQL [roles](https://www.postgresql.org/docs/current/user-manag.htm).

* **Database**: Manage PostgresSQL [databases](https://www.postgresql.org/docs/current/managing-databases.html)

* **Extensions** Manage PostgreSQL [extensions](https://www.postgresql.org/docs/current/external-extensions.html)

## Prerequisites

**Note:** This compatibility matrix is not tested at the moment, it is only provided based on personal experience.
Therefore, it may also be possible, that other combinations are working properly.

| Kubernetes | PostgreSQL | kubepost |
|------------|------------|----------|
| 1.24       | 12,13,14   | >=1.0.0  |
| 1.25       | 12,13,14   | >=1.0.0  |

## CustomResourceDefinitions

A core feature of kubepost is to monitor the Kubernetes API server for changes
to specific resources and ensure that the desired PostgreSQL match these resources.

* **`Connection`**, which defines connections for one or multiple PostgreSQL clusters, that shall be managed by
  kubepost.

* **`Role`**, which defines a PostgreSQL role and its permissions, that shall be managed by kubepost.

* **`Database`**, which defines a PostgreSQL database and its extensions, that shall be managed by kubepost.

The kubepost operator automatically detects changes in the Kubernetes API server to any of the above objects, and
ensures that matching PostgreSQL objects are kept in sync.

## Quickstart

### Installation

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

To learn more about the CRDs introduced by kubepost have a look at the [getting started](docs/getting-started.md) guide.

### Removal

To remove the operator, first delete any custom resources you created in each namespace.

```sh
for n in $(kubectl get namespaces -o jsonpath={..metadata.name}); do
  kubectl delete --all --namespace=$n roles.postgres.kubepost.io,connections.postgres.kubepost.io,databases.postgres.kubepost.io
done
```

After a couple of minutes you can go ahead and remove the operator itself.

```sh
kubectl delete -f deploy/bundle.yaml
```

## Troubleshooting

Before creating a new issue please check the whole [documentation](docs).

## Contributing

Contributions are always welcome, have a look at the [contributing](docs/contributing.md) guidelines to get started.

## Sponsors

Support this project by becoming a sponsor. Your logo will show up here with a link to your website.

<a href="https://github.com/stackitcloud" target="_blank"><img width="64px" src="https://avatars.githubusercontent.com/u/55577607?s=200&v=4"></a>