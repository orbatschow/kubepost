######################################################
# config
######################################################
# setting SHELL to bash allows bash commands to be executed by recipes
# options are set to exit when a recipe line exits non-zero or a piped command fails
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

## location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)


######################################################
# variables
######################################################
# image URL to use all building/pushing image targets
IMG ?= controller:latest

# controller generator version
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_TOOLS_VERSION ?= v0.9.2

# golang ci linter version to use for linting targets
GOLANGCI_VERSION = 1.49.0


######################################################
# misc
######################################################
.PHONY: clean
clean:
	rm -rf build


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
# validate
######################################################
.PHONY: validate-image-tag
validate-image-tag: ## validate the contents of the bundle.yaml
	python scripts/validate-tag.py

######################################################
# lint
######################################################
bin/golangci-lint: bin/golangci-lint-$(GOLANGCI_VERSION)
	@ln -sf golangci-lint-$(GOLANGCI_VERSION) $@

bin/golangci-lint-$(GOLANGCI_VERSION):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v$(GOLANGCI_VERSION)
	@mv bin/golangci-lint $@

.PHONY: lint
lint: bin/golangci-lint out download ## lint all code with golangci-lint
	bin/golangci-lint run ./... --timeout 15m0s


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

.PHONY: generate
generate: generate-manifests generate-client

.PHONY: generate-manifests
generate-manifests: controller-gen ## generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	kustomize build config/default > deploy/bundle.yaml

.PHONY: generate-client
generate-client: controller-gen ## generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## download controller-gen locally if necessary
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)


######################################################
# deploy
######################################################
ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## install CRDs into the K8s cluster specified in ~/.kube/config
	kustomize build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion
	kustomize build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion
	kustomize build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -


######################################################
# release
######################################################
.PHONY: generate-release-manifests
generate-release-manifests: clean generate-manifests ## generate a complete bundle, that can be released
	mkdir -p build
	cp -r config build
	python scripts/release.py
	kustomize build build/config/default > build/config/bundle.yaml

