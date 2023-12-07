.PHONY: all
all: gen build

##@ General

.PHONY: help
help: ## ❓ Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-21s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: build
build: build-rig build-rig-operator ## 🔨 Build all binaries

GOENVS ?= CGO_ENABLED=0
GO ?= $(GOENVS) go
LDFLAGS ?= -s -w
GOBUILD = $(GO) build -ldflags "$(LDFLAGS)"

.PHONY: build-rig
build-rig: ## 🔨 Build rig binary
	(cd cmd/rig/ && $(GO) generate ./... && $(GOBUILD) -o ../../bin/rig ./)

.PHONY: build-rig-operator
build-rig-operator: ## 🔨 Build rig-operator binary
	$(GOBUILD) -o bin/rig-operator ./cmd/rig-operator

.PHONY: gen
gen: proto manifests generate-k8s docs-gen ## 🪄 Run code generation (proto and k8s)

.PHONY: proto
proto: proto-internal proto-public ## 🪄 Generate all protobuf

gen/go/rig/go.mod:
	@mkdir -p gen/go/rig
	@printf "module github.com/rigdev/rig-go-api\n\ngo 1.20\n" > $@

.PHONY: proto-internal
proto-internal: buf protoc-gen-go protoc-gen-connect-go ## 🪄 Generate internal protobuf
	@find . \
		-path './gen/go/*' \
		-not -path './gen/go/rig/*' \
		-type f -name '*.go' -delete
	$(BUF) generate proto/internal --template proto/buf.gen.internal.yaml

.PHONY: proto-public
proto-public: gen/go/rig/go.mod buf protoc-gen-go protoc-gen-connect-go ## 🪄 Generate public protobuf
	@find . \
		-path './gen/go/rig/*' \
		-type f -name '*.go' -delete
	$(BUF) generate proto/rig --template proto/buf.gen.yaml
	@(cd gen/go/rig/; go get -u ./...)

.PHONY: manifests
manifests: controller-gen ## 🪄 Generate k8s manifests
	$(CONTROLLER_GEN) rbac:roleName=rig crd webhook \
		paths="./pkg/api/v1alpha1;./pkg/api/v1alpha2;./pkg/controller" \
		output:rbac:dir=deploy/kustomize/rbac \
		output:webhook:dir=deploy/kustomize/webhook \
		output:crd:dir=deploy/kustomize/crd/bases
	# echo '{{- if .Values.installCRDs }}' > deploy/charts/rig-operator/templates/crd.yaml
	# cat 'deploy/kustomize/crd/bases/rig.dev_capsules.yaml' >> deploy/charts/rig-operator/templates/crd.yaml
	# echo '{{- end }}' >> deploy/charts/rig-operator/templates/crd.yaml

.PHONY: generate-k8s
generate-k8s: controller-gen ## 🪄 Generate runtime.Object implementations.
	$(CONTROLLER_GEN) object paths="./pkg/api/..."

.PHONY: docs
docs: ## 📚 Generate docs
	(cd docs && npm i && npm run start)

.PHONY: docs-gen
docs-gen: crd-ref-docs ## 📚 Generate api references
	$(CRD_REF_DOCS) --renderer markdown \
		--config ./docs/crd-ref-docs/config.yaml \
		--templates-dir ./docs/crd-ref-docs/templates \
		--source-path ./pkg/api/config/v1alpha1 \
		--output-path ./docs/docs/api/config/v1alpha1.md
	$(CRD_REF_DOCS) --renderer markdown \
		--config ./docs/crd-ref-docs/config.yaml \
		--templates-dir ./docs/crd-ref-docs/templates \
		--source-path ./pkg/api/v1alpha1 \
		--output-path ./docs/docs/api/v1alpha1.md
	$(CRD_REF_DOCS) --renderer markdown \
		--config ./docs/crd-ref-docs/v1alpha2-config.yaml \
		--templates-dir ./docs/crd-ref-docs/templates \
		--source-path ./pkg/api/v1alpha2 \
		--output-path ./docs/docs/api/v1alpha2.md

.PHONY: lint
lint: golangci-lint ## 🚨 Run linting
	$(GOLANGCI_LINT) run

.PHONY: test
test: gotestsum ## ✅ Run unit tests
	$(GOTESTSUM) \
		--format-hide-empty-pkg \
		--hide-summary skipped -- \
		-race \
		-short \
		./...

ENVTEST_K8S_VERSION = 1.28.0

.PHONY: test-all
test-all: gotestsum setup-envtest ## ✅ Run all tests
	KUBEBUILDER_ASSETS="$(shell $(SETUP_ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(TOOLSBIN) -p path)" \
	$(GOTESTSUM) \
		--format-hide-empty-pkg \
		--junitfile test-result.xml -- \
		-race \
		-coverprofile cover.out \
		-coverpkg $$(go list ./... | grep rigdev/rig/pkg | tr "\n" ",") \
		-covermode atomic \
		./...

.PHONY: test-integration
test-integration: gotestsum setup-envtest ## ✅ Run integration tests
	KUBEBUILDER_ASSETS="$(shell $(SETUP_ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(TOOLSBIN) -p path)" && \
	$(GOTESTSUM) \
		--format-hide-empty-pkg -- \
		-race \
		-run "^TestIntegration" \
		./...

TAG ?= dev

.PHONY: docker
docker: ## 🐳 Build docker image
	docker build -t ghcr.io/rigdev/rig-operator:$(TAG) -f ./build/package/Dockerfile .

KUBECTX ?= kind-rig
export KUBECTL ?= kubectl --context $(KUBECTX)
export HELM ?= helm --kube-context $(KUBECTX)

.PHONY: deploy-operator
deploy-operator: ## 🚀 Deploy operator to k8s context defined by $KUBECTX (default: kind-rig)
	$(HELM) upgrade --install rig-operator ./deploy/charts/rig-operator \
			--namespace rig-system \
			--create-namespace \
			--set image.tag=$(TAG) \
			--set config.devModeEnabled=true

PLATFORM_TAG ?= latest

.PHONY: deploy-platform
deploy-platform: deploy-operator ## 🚀 Deploy platform to k8s context defined by $KUBECTX (default: kind-rig)
	$(HELM) upgrade --install rig-platform ./deploy/charts/rig-platform \
			--namespace rig-system \
			--create-namespace \
			--set postgres.enabled=true	\
			--set image.tag=$(PLATFORM_TAG) \
			--set rig.cluster.dev_registry.enabled=true \
			--set rig.cluster.dev_registry.host=localhost:30000 \
			--set rig.cluster.dev_registry.cluster_host=registry:5000

PLATFORM_TAG ?= latest

.PHONY: kind-create
kind-create: kind ## 🐋 Create kind cluster with rig dependencies
	## TODO: simplify this
	./deploy/kind/create.sh

.PHONY: kind-load
kind-load: kind docker ## 🐋 Load docker image into kind cluster
	$(KIND) load docker-image ghcr.io/rigdev/rig-operator:$(TAG) -n rig

.PHONY: kind-load-platform
kind-load-platform: kind docker ## 🐋 Load docker image into kind cluster
	$(KIND) load docker-image ghcr.io/rigdev/rig-platform:$(TAG) -n rig

.PHONY: kind-deploy
kind-deploy: kind kind-load deploy-operator ## 🐋 Deploy rig to kind cluster
	$(KUBECTL) rollout restart deployment -n rig-system rig-operator

.PHONY: clean-kind
clean-kind: ## 🧹 Clean kind cluster
	$(KIND) delete clusters rig || true

.PHONY: clean-gen
clean-gen: ## 🧹 Clean generated files
	rm -r gen || true

.PHONY: clean
clean: clean-kind clean-gen ## 🧹 Clean everything

##@ Release

.PHONY: release
release: goreleaser ## 🔖 Release project
	$(GORELEASER) release -f ./build/package/goreleaser/goreleaser.yml

.PHONY: release-build
release-build: goreleaser ## 📸 Build release snapshot
	$(GORELEASER) build -f ./build/package/goreleaser/goreleaser.yml --snapshot --clean

##@ Release CI

.PHONY: ci
ci: gen setup-envtest gotestsum golangci-lint goreleaser ## Ensure tools are installed, go modules are downloaded and build cache is populated.
	go build ./...

##@ Binaries

TOOLSBIN ?= $(shell pwd)/tools/bin
$(TOOLSBIN):
	mkdir -p $(TOOLSBIN)

BUF ?= $(TOOLSBIN)/buf
BUF_GO_MOD_VERSION ?= $(shell cat tools/go.mod | grep -E "github.com/bufbuild/buf " | cut -d ' ' -f2 | cut -c2-)

.PHONY: buf
buf: ## 📦 Download buf locally if necessary.
	(test -s $(BUF) && $(BUF) --version | grep "$(BUF_GO_MOD_VERSION)") || \
	(cd tools && GOBIN=$(TOOLSBIN) go install github.com/bufbuild/buf/cmd/buf)

PROTOC_GEN_GO ?= $(TOOLSBIN)/protoc-gen-go
PROTOC_GEN_GO_GO_MOD_VERSION ?= $(shell cat tools/go.mod | grep -E "google.golang.org/protobuf " | cut -d ' ' -f2)

.PHONY: protoc-gen-go
protoc-gen-go: ## 📦 Download protoc-gen-go locally if necessary.
	(test -s $(PROTOC_GEN_GO) && $(PROTOC_GEN_GO) --version | grep "$(PROTOC_GEN_GO_GO_MOD_VERSION)") || \
	(cd tools && GOBIN=$(TOOLSBIN) go install google.golang.org/protobuf/cmd/protoc-gen-go)

PROTOC_GEN_CONNECT_GO ?= $(TOOLSBIN)/protoc-gen-connect-go
PROTOC_GEN_CONNECT_GO_GO_MOD_VERSION ?= $(shell cat tools/go.mod | grep -E "github.com/bufbuild " | cut -d ' ' -f2)

.PHONY: protoc-gen-connect-go
protoc-gen-connect-go: ## 📦 Download protoc-gen-connect-go locally if necessary.
	(test -s $(PROTOC_GEN_CONNECT_GO) && $(PROTOC_GEN_CONNECT_GO) --version | grep "$(PROTOC_GEN_CONNECT_GO_GO_MOD_VERSION)") || \
	(cd tools && GOBIN=$(TOOLSBIN) go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go)

GORELEASER ?= $(TOOLSBIN)/goreleaser
GORELEASER_GO_MOD_VERSION ?= $(shell cat tools/go.mod | grep -E "github.com/goreleaser/goreleaser " | cut -d ' ' -f2)

.PHONY: goreleaser
goreleaser: ## 📦 Download goreleaser locally if necessary.
	(test -s $(GORELEASER) && $(GORELEASER) --version | grep "$(GORELEASER_GO_MOD_VERSION)") || \
	(cd tools && GOBIN=$(TOOLSBIN) go install github.com/goreleaser/goreleaser)

export KIND ?= $(TOOLSBIN)/kind
KIND_GO_MOD_VERSION ?= $(shell cat tools/go.mod | grep -E "sigs.k8s.io/kind" | cut -d ' ' -f2)

.PHONY: kind
kind: ## 📦 Download kind locally if necessary.
	(test -s $(KIND) && \
	$(KIND) version | grep $(KIND_GO_MOD_VERSION)) || \
	(cd tools && GOBIN=$(TOOLSBIN) go install sigs.k8s.io/kind)

GOTESTSUM ?= $(TOOLSBIN)/gotestsum
GOTESTSUM_GO_MOD_VERSION ?= $(shell cat tools/go.mod | grep -E "gotest.tools/gotestsum" | cut -d ' ' -f2)

.PHONY: gotestsum
gotestsum: ## 📦 Download kind locally if necessary.
	(test -s $(GOTESTSUM) && \
	$(GOTESTSUM) --version | grep $(GOTESTSUM_GO_MOD_VERSION)) || \
	(cd tools && GOBIN=$(TOOLSBIN) go install -ldflags="-X gotest.tools/gotestsum/cmd.version=$(GOTESTSUM_GO_MOD_VERSION)" gotest.tools/gotestsum)

CONTROLLER_GEN ?= $(TOOLSBIN)/controller-gen
CONTROLLER_GEN_GO_MOD_VERSION ?= $(shell cat tools/go.mod | grep -E "sigs.k8s.io/controller-tools" | cut -d ' ' -f2)

.PHONY: controller-gen
controller-gen: ## 📦 Download controller-gen locally if necessary.
	(test -s $(CONTROLLER_GEN) && \
	$(CONTROLLER_GEN) --version | grep $(CONTROLLER_GEN_GO_MOD_VERSION)) || \
	(cd tools && GOBIN=$(TOOLSBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen)

SETUP_ENVTEST ?= $(TOOLSBIN)/setup-envtest
.PHONY: setup-envtest
setup-envtest: ## 📦 Download setup-envtest locally if necessary.
	(test -s $(SETUP_ENVTEST)) || \
	(cd tools && GOBIN=$(TOOLSBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest)

GOLANGCI_LINT ?= $(TOOLSBIN)/golangci-lint
GOLANGCI_LINT_GO_MOD_VERSION ?= $(shell cat tools/go.mod | grep -E "github.com/golangci/golangci-lint" | cut -d ' ' -f2)

.PHONY: golangci-lint
golangci-lint: ## 📦 Download golangci-lint locally if necessary.
	(test -s $(GOLANGCI_LINT) && \
	$(GOLANGCI_LINT) --version | grep $(GOLANGCI_LINT_GO_MOD_VERSION)) || \
	(cd tools && GOBIN=$(TOOLSBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint)

CRD_REF_DOCS ?= $(TOOLSBIN)/crd-ref-docs

.PHONY: crd-ref-docs
crd-ref-docs: ## 📦 Download crd-ref-docs locally if necessary.
	(test -s $(CRD_REF_DOCS)) || \
	(cd tools && GOBIN=$(TOOLSBIN) go install github.com/elastic/crd-ref-docs)
