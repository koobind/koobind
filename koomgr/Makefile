
include ../Makefile.conf

# You can switch between simple (faster) docker build or multiplatform one.
# For multiplatform build on a fresh system, do 'make docker-set-multiplatform-builder'
#DOCKER_BUILD := docker buildx build --builder multiplatform --cache-to type=local,dest=$(BUILDX_CACHE),mode=max --cache-from type=local,src=$(BUILDX_CACHE) --platform linux/amd64,linux/arm64
DOCKER_BUILD := docker build

# Comment this to just build locally
DOCKER_PUSH := --push

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

ifndef ignore-not-found
  ignore-not-found = true
endif

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: all ## Build binary an yaml file
all: manager yamlpod yamldev

displayimg:
	@echo $(IMG)

##@ Build

# Run tests
.PHONY: test
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

.PHONY: manifests
manifests: controller-gen ## Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

.PHONY: generate
generate: controller-gen ## Generate code. e.g zz_generated_deepcopy.go
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: manager
build: generate fmt vet ## Build manager binary locally
	go build -o bin/manager main.go


##@ Deployment -common

.PHONY: crd
crd: manifests kustomize ## Build CRD
	$(KUSTOMIZE) build config/crd -o yaml/crd.yaml

.PHONY: crdinstall
crdinstall: crd ## Install CRDs into the configured K8s cluster
	kubectl apply -f yaml/crd.yaml

.PHONY: crduninstall
crduninstall: manifests kustomize ## Uninstall CRDs from the configured K8s cluster
	# Set ignore-not-found=true to ignore resource not found errors during deletion.
	kubectl delete --ignore-not-found=$(ignore-not-found) -f yaml/crd.yaml


.PHONY: useradmin
useradmin: manifests kustomize ## Create user 'admin' with full admin rights
	$(KUSTOMIZE) build config/useradmin -o yaml/useradmin.yaml

.PHONY: useradmininstall
useradmininstall: useradmin  ## Install user 'admin' with full admin rights
	kubectl apply -f yaml/useradmin.yaml

.PHONY: precommit
precommit: generate yaml-basic yaml-goinkube ## To ensure uptodate generated stuff.

# ---------------------------------------------------------------- Deployment basic

##@ Deployment - basic

# NB: These tasks does not handle authentication webhook installation.

.PHONY: yaml-basic
yaml-basic: crd useradmin ## Generate CRD, RBAC and deployment yaml file with basic configuration.
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(DOCKER_IMG):$(DOCKER_TAG)
	$(KUSTOMIZE) build --load_restrictor none config/overlays/basic -o yaml/basic/deployment.yaml


.PHONY: install-basic
install-basic: yaml-basic crdinstall ## Deploy manager to the configured K8s cluster with basic configuration
	kubectl apply -f yaml/basic/deployment.yaml

# NB: This does not uninstall CRD, as may be locked by existing resources.
uninstall-basic: ## UNdeploy manager to the K8s configured cluster with basic configuration
	kubectl delete --ignore-not-found=$(ignore-not-found) -f yaml/useradmin.yaml
	kubectl delete --ignore-not-found=$(ignore-not-found) -f yaml/basic/deployment.yaml

# ---------------------------------------------------------------- Deployment in goinkube (for dev)

##@ Deployment - goinkube

# NB: These tasks does not handle authentication webhook installation.

.PHONY: yaml-goinkube
yaml-goinkube: crd useradmin ## Generate CRD, RBAC and deployment yaml file with goinkube configuration.
	$(KUSTOMIZE) build --load_restrictor none config/overlays/goinkube -o yaml/goinkube/deployment.yaml


.PHONY: install-goinkube
install-goinkube: yaml-goinkube crdinstall ## Deploy manager to the configured K8s cluster with goinkube configuration
	kubectl apply -f yaml/goinkube/deployment.yaml

# NB: This does not uninstall CRD, as may be locked by existing resources.
uninstall-goinkube: ## UNdeploy manager to the K8s configured cluster with goinkube configuration
	kubectl delete --ignore-not-found=$(ignore-not-found) -f yaml/useradmin.yaml
	kubectl delete --ignore-not-found=$(ignore-not-found) -f yaml/goinkube/deployment.yaml


#------------------------------------------------------------------------------------

##@ Deployment - Docker image

.PHONY: docker
docker: ## Build and push koo-manager image
	$(DOCKER_BUILD) $(DOCKER_PUSH) -t $(DOCKER_IMG):$(DOCKER_TAG) -f Dockerfile .

#------------------------------------------------------------------------------------


##@ Build tooling

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CONTROLLER_TOOLS_VERSION ?= v0.9.2

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

