# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:crdVersions={v1},trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

#Generate semver.mk
gen-semver: generate
	(cd core; rm -f core_generated.go; go generate)
	go run core/semver/semver.go -f mk > semver.mk

version: gen-semver
	make -f docker.mk version

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

unit-test:
	go clean -cache; cd ./test; go test -race -cover -coverprofile=coverage.out -coverpkg ../pkg/... ./. 2>&1 | tee test-result

# Build manager binary
manager: gen-semver fmt vet
	go build -o bin/manager main.go

static-crd: manifests kustomize
	$(KUSTOMIZE) build config/crd > deploy/crds/storage.dell.com.crds.all.yaml

static-manager: manifests kustomize
	$(KUSTOMIZE) build config/install > deploy/operator.yaml

# generate static manifests
static-manifests: static-crd static-manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate gen-semver fmt vet static-manifests
	go run ./main.go

# Install CRDs into a cluster
install: static-crd
	kubectl apply -f deploy/crds/storage.dell.com.crds.all.yaml

# Uninstall CRDs from a cluster
uninstall: static-crd
	kubectl delete -f deploy/crds/storage.dell.com.crds.all.yaml

config-map:
	tar -cf - driverconfig/* | gzip > deploy/config.tar.gz
	kubectl create configmap dell-csi-operator-config --from-file deploy/config.tar.gz -o yaml --dry-run | kubectl apply -f -
	rm -f deploy/config.tar.gz

remove-config-map:
	kubectl delete configmap dell-csi-operator-config

install-manager: install config-map static-manager
	kubectl apply -f deploy/operator.yaml

uninstall-manager:
	kubectl delete -f deploy/operator.yaml

# Install Operator
deploy: install install-manager

# Remove controller from the configured Kubernetes cluster in ~/.kube/config
undeploy: uninstall-manager remove-config-map

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy-dev: install gen-semver
	make -f docker.mk deploy-dev

# Build & deploy
build-deploy-dev: install gen-semver
	make -f docker.mk build-deploy-dev

# To be used after make dev
remove-dev: kustomize
	tar -cf - driverconfig/* | gzip > config/dev/config.tar.gz
	$(KUSTOMIZE) build config/dev | kubectl delete -f -
	rm -f config/dev/config.tar.gz

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

# Build the docker image
docker-build: gen-semver
	make -f docker.mk docker-build

# Push the docker image
docker-push: gen-semver
	make -f docker.mk docker-push

docker-feature-build: gen-semver
	make -f docker.mk docker-feature-build

docker-feature-push: gen-semver
	make -f docker.mk docker-feature-push

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

# Update bundle manifests
.PHONY: update-bundle
update-bundle: static-manifests gen-semver
	make -f docker.mk update-bundle

# Update bundle manifests without updating the base CSV
update-bundle-keep-base: static-manifests gen-semver
	make -f docker.mk update-bundle-keep-base

# Use for generating bundles for testing
.PHONY: bundle
bundle: manifests gen-semver
	make -f docker.mk bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build: gen-semver
	make -f docker.mk bundle-build

# Push the bundle image.
.PHONY: bundle-push
bundle-push: gen-semver
	make -f docker.mk bundle-push

# Push the docker image
index: gen-semver
	make -f docker.mk index

.PHONY: check
check:
	@./check.sh

# Run integration tests
.PHONY: integ-test
integ-test: check gen-semver
	(CGO_ENABLED=0 go test -v ./test/integration-tests/)
