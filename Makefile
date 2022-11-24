# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.23

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: build

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

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile cover.out

unit-test:
	go clean -cache; cd ./test; go test -race -cover -coverprofile=coverage.out -coverpkg ../pkg/... ./. 2>&1 | tee test-result

#Generate semver.mk
gen-semver: generate
	(cd core; rm -f core_generated.go; go generate)
	go run core/semver/semver.go -f mk > semver.mk

version: gen-semver
	make -f docker.mk version

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
	kubectl create configmap dell-csi-operator-config --from-file deploy/config.tar.gz -o yaml --dry-run=client | kubectl apply -f -
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
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,v0.7.0)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3,v3.8.7)

ENVTEST = $(shell pwd)/bin/setup-envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)@$(3)" ;\
go get $(2)@$(3) ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
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

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.15.1/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif
