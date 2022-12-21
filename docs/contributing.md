# Contributing

This document will tell you how to get the project up and running as smoothly as possible.

## Prerequisites

### Kubernetes

In order to test your changes you need a running Kubernetes cluster. There are multiple
options for you, but maybe the easiest one is [kind](https://kind.sigs.k8s.io/).

### PostgreSQL

Most likely you want to test your changes within a running PostgreSQL instance. You can
set up a fully functional stack by applying the provided manifests within this repository:

```sh
kubectl apply -k config/samples/postgres
```

> Note: You have to install [kustomize](https://github.com/kubernetes-sigs/kustomize) in order to execute the command
> mentioned above.

### Go

In order to modify the code you need a working [Go](https://github.com/golang/go) installation.

### Python

In order to modify the workflows you need a working [Python](https://www.python.org/) installation.

### Ruby

In order to generate the changelogs you need a working [Ruby](https://www.ruby-lang.org/en/) installation.

### ASDF

As multiple languages are required for this project it is recommended to use some kind of version manager
like [asdf](https://github.com/asdf-vm/asdf).

## Custom Resource Definitions

In order to test changes to the Custom Resources Definitions you have to apply these to your running
Kubernetes cluster. Whenever you make changes you can just execute:

```sh
make install
```

When you are done with testing the changes you can remove the CRDs by executing:

```sh
make uninstall
```

## Operator

To test changes within the operator you have to point your `KUBECONFIG` to the Kubernetes
cluster you configured before. After setting up the `KUBECONFIG` you can just run the
operator locally and it will interact with the configured Kubernetes cluster. You should
be able to develop and debug locally without any issues by running:

```sh
make run
```

## Release

Many files (documentation, manifests, ...) in this repository are
auto-generated. E.g. `deploy/bundle.yaml` originates from the *yaml* files in
`/config`. Before proposing a pull request:

1. Commit your changes.
2. Run `make generate`.
3. Commit the generated changes.
