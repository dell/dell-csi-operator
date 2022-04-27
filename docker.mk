include semver.mk

# Operator version tagged with build number. For e.g. - v1.2.0.001
VERSION="v$(MAJOR).$(MINOR).$(PATCH).$(BUILD)"

# Bundle Version is the semantic version(required by operator-sdk)
BUNDLE_VERSION="$(MAJOR).$(MINOR).$(PATCH)"

# Operator version tagged with release. For e.g. - v1.2.0
REL_VERSION="v$(MAJOR).$(MINOR).$(PATCH)"
FEATUREVERSION="v$(MAJOR).$(MINOR).$(PATCH).$(SANITIZED_GIT_BRANCH_NAME).$(BUILD_NUMBER)"

# Options for 'bundle-build'
CHANNELS ?= stable
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
DEFAULT_CHANNEL ?= stable
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Registry for all images
REGISTRY ?= "dellemc"

# Default Operator image name
OPERATOR_IMAGE ?= dell-csi-operator
# Default certified bundle image name
BUNDLE_IMAGE ?= csiopbundle_ops_certified
# Default Index Image name
INDEX_IMAGE ?= dellemcregistry_certified
SOURCE_INDEX_IMG ?= dellemc/dell-csi-operator/dellemcregistry_certified:v1.7.0


# Operator Images
# Operator image taffed with patch number. FOr e.g. - v1.2.0.001
IMG ?= "$(REGISTRY)/$(OPERATOR_IMAGE):$(VERSION)"
# Operator image tagged for release. For e.g. - v1.2.0
REL_IMG ?= "$(REGISTRY)/$(OPERATOR_IMAGE):$(REL_VERSION)"
# special image tag just used for PR builds, uses branch name, information from PR
FEATUREIMG = "$(REGISTRY)/$(OPERATOR_IMAGE):$(FEATUREVERSION)"

# OLM Images
# Bundle image tagged with patch number - v1.2.0.001
BUNDLE_IMG ?= "$(REGISTRY)/$(BUNDLE_IMAGE):$(VERSION)"

# Index image tagged with patch number - v1.2.0.001
INDEX_IMG ?= "$(REGISTRY)/$(INDEX_IMAGE):$(VERSION)"
# Index image tagged for release
INDEX_REL_IMG ?= "$(REGISTRY)/$(INDEX_IMAGE):$(REL_VERSION)"

NAMESPACE ?= test-operator

# Deploy controller in a namespace (creates one)
deploy-dev: kustomize
	tar -cf - driverconfig/* | gzip > config/dev/config.tar.gz
	cd config/dev && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	cd config/dev && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/dev | kubectl apply -f -
	rm -f config/dev/config.tar.gz

# Build & deploy controller
build-deploy-dev: docker-push deploy-dev

# Build the docker image
docker-build:
	docker build . -t ${IMG} -t ${REL_IMG}

# Push the docker image
docker-push: docker-build
	docker push ${IMG}
	docker push ${REL_IMG}

docker-feature-build:
	docker build . -t ${FEATUREIMG}

docker-feature-push: docker-feature-build
	docker push ${FEATUREIMG}

version:
	echo ${IMG}

# Generate bundle manifests and metadata, then validate generated files.
# Use this to generate and update bundle manifests
update-bundle: kustomize
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(REL_IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --version $(BUNDLE_VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

# Generate bundle manifests and metadata, then validate generated files.
# Use this to generate and update bundle manifests
update-bundle-keep-base: kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(REL_IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --manifests --version $(BUNDLE_VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

# use this only for building & testing bundles and not for updating bundles in repository
bundle: kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --manifests --version $(BUNDLE_VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

# Build the bundle image.
bundle-build: bundle
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

bundle-push: bundle-build
	docker push $(BUNDLE_IMG)

index: docker-push bundle-push
	opm index add --bundles $(BUNDLE_IMG) --from-index $(SOURCE_INDEX_IMG) --tag $(INDEX_IMG) --container-tool docker
	docker tag $(INDEX_IMG) $(INDEX_REL_IMG)
	docker push $(INDEX_IMG)
	docker push $(INDEX_REL_IMG)

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.8.7 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif
