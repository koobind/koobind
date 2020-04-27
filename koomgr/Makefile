
# Image URL to use all building/pushing image targets
IMG ?= koobind/manager:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager yamlpod yamldev

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go


# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

crd: manifests
	kustomize build config/crd -o yaml/crd.yaml

rbac: manifests
	kustomize build config/rbac -o yaml/rbac.yaml

# ---------------------------------------------------------------- Deployment in pod (std)
# NB: These tasks does not handle authentication webhook installation.

# Generate ready to install manifests
yamlpod: crd rbac
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/overlays/pod -o yaml/pod/deploy.yaml

deploypod: yamlpod
	kubectl apply -f yaml/crd.yaml
	kubectl apply -f yaml/pod/deploy.yaml

undeploypod: yamlpod
	-kubectl delete -f yaml/pod/deploy.yaml
	kubectl delete -f yaml/crd.yaml



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
	kubectl apply -f yaml/dev/deploy.yaml

undeploydev: yamldev
	-kubectl delete -f yaml/dev/deploy.yaml
	kubectl delete -f yaml/crd.yaml

#------------------------------------------------------------------------------------

# Build the docker image
docker-build: test
	cd ..; docker build -f koomgr/Dockerfile . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
