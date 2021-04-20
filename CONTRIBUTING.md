# Prerequisites

## Kind

This guide uses kind to demonstrate deployment and operation of kubepost in a multi/single-node Kubernetes cluster
running locally on Docker.

### Configure kind

Configuring kind cluster creation is done using a YAML configuration file. Create a `config.yaml` file based on the
following template. You can add/remove nodes as you want The example will create a cluster with 3 worker nodes and 1
control-plane node. Note: kubepost will also work with a single node only.

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
  - role: worker
  - role: worker
  - role: worker
```

To start the cluster you have to run:

```shell
kind create cluster --config=config.yaml
```

If you just want to run a simple single-node cluster without any configuration you can execute:

```shell
kind create cluster
```

## Metacontroller

kubepost uses Metacontroller under the hood, so this component hast to be installed within the cluster. You can view
detailed instructions for the installation process here.

```shell
kubectl apply -k https://github.com/metacontroller/metacontroller/manifests/production
```

Alternatively create a new kustomization file with the following content:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: metacontroller

resources:
  - github.com/metacontroller/metacontroller/manifests//production
```

Apply the file to the cluster:

```shell
kustomize build . | k
```

## Kubepost

After you have installed metacontroller you have to deploy kubepost:

```shell
kubectl apply -k https://github.com/metacontroller/metacontroller/manifests/production
```

Alternatively create a new kustomization file with the following content:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: metacontroller

resources:
  - github.com/orbatschow/kubepost//manifests
```

Apply the file to the cluster:

```shell
kustomize build . | k
```

## Telepresence

Telepresence can be used to proxy all requests to the previously deployed kubepost Pod to our locally running golang
application. To proxy all requests to our locally running kubepost application you can run:

```shell
telepresence --swap-deployment kubepost-controller:controller --namespace metacontroller --expose 8080 --method inject-tcp
```

## Postgres

To provide kubepost with a functional Postgres instance and also gain some further debugging capabilities we have to
start a local Postgres instance.

```shell
docker run -p 5432:5432 -e POSTGRES_PASSWORD=root postgres
```

Note: Although there is a Postgres deployment within the `examples` directory, it won't be functional when we are
proxying the requests to our locally running kubepost controller.

## Testing

At this stage you should have a locally running Kubernetes cluster with Metacontroller and kubepost installed. There
should also be a local Postgres container and all requests should be proxied to localhost:8080.

# Development

## Repository

Start with cloning the repository

```shell
git clone git@github.com:orbatschow/kubepost.git
```

## Running

After you have cloned the repository you can run the application by executing

```shell
cd kubepost
go run main.go
```

## Testing

You can now test your new code by applying some examples within the `examples` directory to the cluster. All
requests should be redirected from the cluster to the local application.