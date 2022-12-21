######################################################
# config
######################################################
# setting SHELL to bash allows bash commands to be executed by recipes
# options are set to exit when a recipe line exits non-zero or a piped command fails
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

## location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# image url to use all building/pushing image targets
IMG ?= controller:latest

# controller generator binary and version
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_TOOLS_VERSION ?= v0.9.2

# crdoc binary and version
CRDOC ?= $(LOCALBIN)/crdoc
CRDOC_VERSION ?= v0.6.1

# golang ci linter version to use for linting targets
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.50.1

# github_changelog_generator binary
GITHUB_CHANGELOG_GENERATOR ?= github_changelog_generator

######################################################
# misc
######################################################
.PHONY: clean
clean:
	rm -rf build

.PHONY: crdoc
crdoc:
	test -s $(CRDOC) || GOBIN=$(LOCALBIN) go install fybrik.io/crdoc@$(CRDOC_VERSION)

.PHONY: controller-gen
controller-gen:
	test -s $(CONTROLLER_GEN) || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: golangci-lint
golangci-lint:
	test -s $(GOLANGCI_LINT) || GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: github_changelog_generator
github_changelog_generator:
	gem install --conservative github_changelog_generator -v 1.16.4

######################################################
# checksum
######################################################
.PHONY: checksum
checksum:
	./scripts/checksum.sh


######################################################
# go
######################################################
.PHONY: tidy
tidy: ## clean up go.mod and go.sum
	go mod tidy

.PHONY: download
download: ## downloads the dependencies
	go mod download -x

.PHONY: build
build: generate ## build manager binary.
	go build -o build/manager main.go

.PHONY: run
run: manifests generate fmt vet ## run a controller from your host
	go run ./main.go


######################################################
# lint
######################################################
.PHONY: lint
lint: golangci-lint  ## lint all code with golangci-lint
	$(GOLANGCI_LINT) run ./... --timeout 15m0s


######################################################
# test
######################################################
.PHONY: test
test:
	go test ./... -coverprofile cover.out


######################################################
# docker
######################################################
.PHONY: docker-build
docker-build: test ## build docker image with the manager
	docker build -t ${IMG} .

PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: test
	docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	docker buildx build --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile .
	docker buildx rm project-v3-builder


######################################################
# generate
######################################################
# get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# all
.PHONY: generate
generate: generate-manifests generate-crd-documentation generate-client generate-changelog

# manifests
.PHONY: generate-manifests
generate-manifests: controller-gen ## generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	kustomize build config/default > deploy/bundle.yaml

# client
.PHONY: generate-client
generate-client: controller-gen ## generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# documentation
.PHONY: generate-crd-documentation
generate-crd-documentation: crdoc generate-manifests
	$(CRDOC) --resources config/crd/bases/postgres.kubepost.io_instances.yaml --output docs/instance.md
	$(CRDOC) --resources config/crd/bases/postgres.kubepost.io_roles.yaml --output docs/role.md
	$(CRDOC) --resources config/crd/bases/postgres.kubepost.io_databases.yaml --output docs/database.md

.PHONY: generate-changelog
generate-changelog: github_changelog_generator
	github_changelog_generator -u orbatschow -p kubepost


######################################################
# deploy
######################################################
ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: generate ## install CRDs into the K8s cluster specified in ~/.kube/config
	kustomize build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: generate ## uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion
	kustomize build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -


######################################################
# release
######################################################
.PHONY: generate-release-manifests
generate-release-manifests: clean generate-manifests ## generate a complete bundle, that can be released
	mkdir -p build
	cp -r config build
	python scripts/release.py
	kustomize build build/config/default > build/config/bundle.yaml

