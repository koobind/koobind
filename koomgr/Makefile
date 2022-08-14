
include ../Makefile.conf

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

rbac: manifests kustomize ## Build RBAC
	$(KUSTOMIZE) build config/rbac -o yaml/rbac.yaml

useradmin: manifests kustomize ## Build RBAC
	$(KUSTOMIZE) build config/useradmin -o yaml/useradmin.yaml


# --------------------------------------- crd

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: crdinstall
crdinstall: crd ## Install CRDs into the configured K8s cluster
	kubectl apply -f yaml/crd.yaml

.PHONY: crduninstall
crduninstall: manifests kustomize ## Uninstall CRDs from the configured K8s cluster
	# Set ignore-not-found=true to ignore resource not found errors during deletion.
	kubectl delete --ignore-not-found=$(ignore-not-found) -f yaml/crd.yaml


# --------------------------------------- basic (Standard)

##@ Deployment - basic

# ---------------------------------------------------------------- Deployment in pod (std)
# NB: These tasks does not handle authentication webhook installation.

.PHONY: yaml-basic
yaml-basic: crd rbac useradmin ## Generate CRD, RBAC and deployment yaml file with basic configuration.
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(DOCKER_IMG):$(DOCKER_TAG)
	$(KUSTOMIZE) build --load_restrictor none config/overlays/basic -o yaml/basic/deployment.yaml


.PHONY: install-basic
install-basic: yaml-basic crdinstall ## Deploy manager to the configured K8s cluster with basic configuration
	kubectl apply -f yaml/basic/deployment.yaml
	kubectl apply -f yaml/rbac.yaml
	kubectl apply -f yaml/useradmin.yaml

# NB: This does not uninstall CRD, as may be locked by existing resources.
uninstall-basic: ## UNdeploy manager to the K8s configured cluster with basic configuration
	kubectl delete --ignore-not-found=$(ignore-not-found) -f yaml/useradmin.yaml
	kubectl delete --ignore-not-found=$(ignore-not-found) -f yaml/rbac.yaml
	kubectl delete --ignore-not-found=$(ignore-not-found) -f yaml/basic/deployment.yaml

# ---------------------------------------------------------------- Deployment for dev (local)
# NB: These tasks does not handle authentication webhook installation.

config/overlays/dev/cert/tls.key:
	openssl req -x509 -sha256 -nodes -days 3650 -newkey rsa:2048 -keyout config/overlays/dev/cert/tls.key -out config/overlays/dev/cert/tls.crt -subj "/C=FR/ST=Paris/L=Paris/O=BROADSoftware/OU=IT Department/CN=koomgrdev"

certificate: config/overlays/dev/cert/tls.key

yamldev: certificate crd rbac
	bash config/overlays/dev/generate_webhooks_patches.sh
	kustomize build config/overlays/dev -o yaml/dev/deploy.yaml

deploydev: yamldev
	kubectl apply -f yaml/crd.yaml
	kubectl apply -f yaml/rbac.yaml
	kubectl apply -f yaml/dev/deploy.yaml

undeploydev: yamldev
	-kubectl delete -f yaml/dev/deploy.yaml
	kubectl delete -f yaml/rbac.yaml
	kubectl delete -f yaml/crd.yaml

#------------------------------------------------------------------------------------


##@ Build Dependencies

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

