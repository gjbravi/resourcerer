CLUSTER_NAME ?= resourcerer
KUBE_CONTEXT ?= kind-$(CLUSTER_NAME)
IMG ?= ghcr.io/gjbravi/resourcerer:dev
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

export GOTOOLCHAIN ?= auto

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2} /^##@/ {printf "\n\033[1m%s\033[0m\n", substr($$0, 5)}' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Format Go and UI code
	gofmt -w .
	cd web && npm run format

.PHONY: lint
lint: ## Run linters
	go vet ./...
	go tool golangci-lint run
	cd web && npm run lint && npm run typecheck
	helm lint charts/resourcerer

.PHONY: test
test: ## Run Go and UI unit tests
	go test -race -coverprofile=coverage.out ./...
	cd web && npm test

.PHONY: web
web: ## Build the UI into web/dist
	cd web && npm ci --no-audit --no-fund && npm run build

.PHONY: build
build: web ## Build the binary into bin/
	go build -trimpath -ldflags="-X main.version=$(VERSION)" -o bin/resourcerer ./cmd/resourcerer

.PHONY: run
run: ## Run locally against the current kube context (UI on :8080)
	go run ./cmd/resourcerer --zap-devel

.PHONY: web-dev
web-dev: ## Run the UI dev server with /api proxied to :8080
	cd web && npm run dev

##@ Container image

.PHONY: image
image: ## Build the container image
	docker build --build-arg VERSION=$(VERSION) -t $(IMG) .

##@ Local cluster (kind)

.PHONY: kind-up
kind-up: ## Create a kind cluster with metrics-server, VPA and demo workloads
	CLUSTER_NAME=$(CLUSTER_NAME) ./hack/kind-up.sh

.PHONY: kind-deploy
kind-deploy: image ## Build, load into kind and install the chart
	CLUSTER_NAME=$(CLUSTER_NAME) ./hack/kind-load.sh $(IMG)
	helm upgrade --install resourcerer charts/resourcerer \
	  --kube-context $(KUBE_CONTEXT) --namespace resourcerer --create-namespace \
	  --set image.repository=$(firstword $(subst :, ,$(IMG))) --set image.tag=$(lastword $(subst :, ,$(IMG))) \
	  --set image.pullPolicy=Never --wait
	kubectl --context $(KUBE_CONTEXT) -n resourcerer rollout restart deploy/resourcerer
	kubectl --context $(KUBE_CONTEXT) -n resourcerer rollout status deploy/resourcerer --timeout 120s
	@echo "UI: kubectl --context $(KUBE_CONTEXT) -n resourcerer port-forward svc/resourcerer 8080:80"

.PHONY: kind-down
kind-down: ## Delete the kind cluster
	kind delete cluster --name $(CLUSTER_NAME)

.PHONY: e2e
e2e: ## Run end-to-end tests against the kind cluster
	KUBE_CONTEXT=$(KUBE_CONTEXT) ./test/e2e/run.sh
