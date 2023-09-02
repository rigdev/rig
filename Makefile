.PHONY: all
all: gen build

##@ General

.PHONY: help
help: ## ❓ Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-21s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: build
build: build-rig build-rig-server build-rig-proxy build-rig-admin ## 🔨 Build all binaries

GOENVS ?= CGO_ENABLED=0
GO ?= $(GOENVS) go
LDFLAGS ?= -s -w
GOBUILD = $(GO) build -ldflags "$(LDFLAGS)"
DEVENVS = RIG_MANAGEMENT_TELEMETRY_ENABLED=false

.PHONY: build-rig
build-rig: ## 🔨 Build rig binary
	(cd cmd/rig/ && $(GOBUILD) -o ../../bin/rig ./)

.PHONY: build-rig-server
build-rig-server: ## 🔨 Build rig-server binary
	$(GOBUILD) -o bin/rig-server ./cmd/rig-server

.PHONY: build-rig-proxy
build-rig-proxy: ## 🔨 Build rig-proxy binary
	$(GOBUILD) -o bin/rig-proxy ./cmd/rig-proxy

.PHONY: build-rig-admin
build-rig-admin: ## 🔨 Build rig-admin binary
	$(GOBUILD) -o bin/rig-admin ./cmd/rig-admin

.PHONY: gen
gen: proto mocks ## 🪄 Run code generation (proto and mocks)

.PHONY: proto
proto: proto-internal proto-public ## 🪄 Generate all protobuf

gen/go/rig/go.mod:
	@mkdir -p gen/go/rig
	@printf "module github.com/rigdev/rig-go-api\n\ngo 1.20\n" > $@

.PHONY: proto-internal
proto-internal: buf protoc-gen-go protoc-gen-connect-go ## 🪄 Generate internal protobuf
	@find . -path './gen/go/*' -not -path './gen/go/rig/*' -type f -name '*.go' -delete
	$(BUF) generate proto/internal --template proto/buf.gen.internal.yaml

.PHONY: proto-public
proto-public: gen/go/rig/go.mod buf protoc-gen-go protoc-gen-connect-go ## 🪄 Generate public protobuf
	@find . -path './gen/go/rig/*' -type f -name '*.go' -delete
	$(BUF) generate proto/rig --template proto/buf.gen.yaml
	@(cd gen/go/rig/; go get -u ./...)

.PHONY: mocks
mocks: mockery mocks-clean ## 🪄 Generate mocks
	$(MOCKERY) --config ./build/.mockery.yaml

.PHONY: mocks-clean
mocks-clean: ## 🧹 Clean mocks
	@find . -type f -name 'mock_*.go' -delete

.PHONY: test
test: gotestsum ## ✅ Run tests
	$(GOTESTSUM) --format-hide-empty-pkg --junitfile test-result.xml

.PHONY: run
run: build-rig-server ## 🏃 Run rig-server
	$(DEVENVS) ./bin/rig-server -c ./configs/server-config.yaml

.PHONY: watch
watch: modd ## 👀 Run modd
	$(MODD) -f ./build/modd.conf

.PHONY: download-dashboard
download-dashboard: ## ⬇️ Download latest dashboard to /pkg/service/web
	gh release download --repo rigdev/dashboard -p public.tar.gz -O - | \
		tar -xvz && \
	rsync -a ./public/ ./pkg/service/web/ && \
	rm -rf public

TAG ?= dev

.PHONY: docker
docker: ## 🐳 Build docker image
	docker build -t ghcr.io/rigdev/rig:$(TAG) -f ./build/package/Dockerfile .

.PHONY: docker-compose-up
docker-compose-up: ## 🐳 Run docker-compose
	@echo "$(DEVENVS)" | sed -e "s/ /\n/" > ./deploy/docker-compose/.env.dev
	docker-compose \
		-f ./deploy/docker-compose/docker-compose.yaml \
		--env-file ./deploy/docker-compose/.env.dev \
		up --build -d

.PHONY: docker-compose-down
docker-compose-down: ## 🐳 Stop docker-compose
	docker-compose -f ./deploy/docker-compose/docker-compose.yaml down

KUBECTX ?= kind-rig
KUBECTL ?= kubectl --context $(KUBECTX)
HELM ?= helm --kube-context $(KUBECTX)

.PHONY: deploy
deploy: ## 🚀 Deploy to k8s context defined by $KUBECTX (default: kind-rig)
	$(HELM) upgrade --install rig ./deploy/charts/rig \
  		--namespace rig-system \
		--set image.tag=$(TAG) \
		--set mongodb.enabled=true \
		--set rig.management.telemetry.enabled=false \
  		--create-namespace
	$(KUBECTL) rollout restart deployment -n rig-system rig

.PHONY: kind-create
kind-create: kind ## 🐋 Create kind cluster with rig dependencies
	$(KIND) get clusters | grep '^rig$$' || \
		$(KIND) create cluster --name rig

.PHONY: kind-load
kind-load: kind docker ## 🐋 Load docker image into kind cluster
	$(KIND) load docker-image ghcr.io/rigdev/rig:$(TAG) -n rig

.PHONY: kind-deploy
kind-deploy: kind kind-load deploy ## 🐋 Deploy rig to kind cluster

.PHONY: kind-clean
kind-clean: ## 🧹 Clean kind cluster
	$(KIND) delete clusters rig

##@ Release

.PHONY: release
release: goreleaser ## 🔖 Release project
	$(GORELEASER) release -f ./build/package/goreleaser/goreleaser.yml

.PHONY: release-build
release-build: goreleaser ## 📸 Build release snapshot
	$(GORELEASER) build -f ./build/package/goreleaser/goreleaser.yml --snapshot --clean

##@ Binaries

TOOLSBIN ?= $(shell pwd)/tools/bin
$(TOOLSBIN):
	mkdir -p $(TOOLSBIN)

.PHONY: tools
tools: buf mockery protoc-gen-go protoc-gen-connect-go modd goreleaser kind gotestsum ## 📦 Download all tools

.PHONY: tools-ci
tools-ci: buf mockery protoc-gen-go protoc-gen-connect-go goreleaser gotestsum ## 📦 Download tools used in CI

BUF ?= $(TOOLSBIN)/buf
BUF_GO_MOD_VERSION ?= $(shell cat tools/go.mod | grep -E "github.com/bufbuild/buf " | cut -d ' ' -f2 | cut -c2-)

.PHONY: buf
buf: ## 📦 Download buf locally if necessary.
	(test -s $(BUF) && $(BUF) --version | grep "$(BUF_GO_MOD_VERSION)") || \
	(cd tools && GOBIN=$(TOOLSBIN) go install github.com/bufbuild/buf/cmd/buf)

MOCKERY ?= $(TOOLSBIN)/mockery
MOCKERY_GO_MOD_VERSION ?= $(shell cat tools/go.mod | grep -E "github.com/vektra/mockery/v2 " | cut -d ' ' -f2 | cut -c2-)

.PHONY: mockery
mockery: ## 📦 Download mockery locally if necessary.
	(test -s $(MOCKERY) && $(MOCKERY) --version | grep "$(MOCKERY_GO_MOD_VERSION)") || \
	(cd tools && GOBIN=$(TOOLSBIN) go install github.com/vektra/mockery/v2)

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

MODD ?= $(TOOLSBIN)/modd
MODD_GO_MOD_VERSION ?= $(shell cat tools/go.mod | grep -E "github.com/cortesi/modd " | cut -d ' ' -f2 | cut -c2- | cut -d. -f-2)

.PHONY: modd
modd: ## 📦 Download modd locally if necessary.
	(test -s $(MODD) && $(MODD) --version 2>&1 | grep "$(MODD_GO_MOD_VERSION)") || \
	(cd tools && GOBIN=$(TOOLSBIN) go install github.com/cortesi/modd/cmd/modd)

GORELEASER ?= $(TOOLSBIN)/goreleaser
GORELEASER_GO_MOD_VERSION ?= $(shell cat tools/go.mod | grep -E "github.com/goreleaser/goreleaser " | cut -d ' ' -f2)

.PHONY: goreleaser
goreleaser: ## 📦 Download goreleaser locally if necessary.
	(test -s $(GORELEASER) && $(GORELEASER) --version | grep "$(GORELEASER_GO_MOD_VERSION)") || \
	(cd tools && GOBIN=$(TOOLSBIN) go install github.com/goreleaser/goreleaser)

KIND ?= $(TOOLSBIN)/kind
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
