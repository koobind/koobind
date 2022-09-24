
include ../Makefile.conf

HELM_TARGET_NAMESPACE := koobind
HELM_KOOMGR_OPTIONS := --values ./values.kspray5.yaml
HELM_KOOMGR_OPTIONS := $(HELM_KOOMGR_OPTIONS) --values ./values.kspray5-gik.yaml
HELM_DIRECTORY_OPTIONS :=

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

# Run tests
.PHONY: test
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...


.PHONY: crd
crd: controller-gen ## Generate CRD inside helm chart
	$(CONTROLLER_GEN) crd paths="./..."  output:crd:artifacts:config=../helm/koomgr/crds


.PHONY: generate
generate: controller-gen ## Generate code. e.g zz_generated_deepcopy.go
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: manager
build: generate fmt vet ## Build manager binary locally
	go build -o bin/manager main.go


.PHONY: crduninstall
crduninstall: crd ## Uninstall CRDs from the configured K8s cluster
	# Set ignore-not-found=true to ignore resource not found errors during deletion.
	kubectl delete --ignore-not-found=$(ignore-not-found) -f ../helm/koomgr/crds


.PHONY: precommit
precommit: generate crd ## To ensure uptodate generated stuff.

#------------------------------------------------------------------------------------

##@ Deployment

deploy: deploy-mgr deploy-dir ## Deploy all in configured cluster (deploy-dir may fail for first deployment)

.PHONY: deploy-mgr
deploy-mgr: crd ## Deploy koobind in configured cluster:
	cd ../helm/koomgr && helm -n $(HELM_TARGET_NAMESPACE) upgrade -i --create-namespace $(HELM_KOOMGR_OPTIONS) koobind .

.PHONY: deploy-dir
deploy-dir: ## Deploy a basic users set
	cd ../helm/koodirectory && helm -n $(HELM_TARGET_NAMESPACE) upgrade -i --create-namespace $(HELM_DIRECTORY_OPTIONS) koodir .


.PHONY: undeploy
undeploy:  ## UNdeploy in configured cluster
	-cd ../helm/koodirectory && helm -n $(HELM_TARGET_NAMESPACE) uninstall koodir
	-cd ../helm/koomgr && helm -n $(HELM_TARGET_NAMESPACE) uninstall koobind
	make crduninstall
	kubectl delete ns $(HELM_TARGET_NAMESPACE)


#------------------------------------------------------------------------------------

##@ Docker image

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

CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_TOOLS_VERSION ?= v0.9.2

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

