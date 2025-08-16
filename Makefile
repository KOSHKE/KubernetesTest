# Makefile for Microservices Order System (dev only)

.PHONY: help proto proto-clean dev-up dev-rebuild dev-down fmt deps-get deps-tidy mod-download update-mod

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

BUF_IMAGE ?= bufbuild/buf:latest

proto: ## Generate protobuf stubs for Go using Buf (Docker-based)
	@echo "Generating protobuf stubs with Buf (via $(BUF_IMAGE))..."
	@docker run --rm -v "$(CURDIR)":/workspace -w /workspace/proto $(BUF_IMAGE) generate

proto-clean: ## Remove generated protobuf stubs
	@echo "Cleaning generated protobuf stubs..."
	@rm -rf proto-go/*

dev-up: ## Start local development environment with Air hot-reload (no rebuild)
	@echo "Generating protobuf stubs (buf generate) and starting dev environment with Air..."
	$(MAKE) proto
	docker compose up -d

dev-rebuild: ## Rebuild images and start dev environment with Air (one-off)
	@echo "Rebuilding images and starting dev environment with Air..."
	$(MAKE) proto
	docker compose up -d --build

dev-down: ## Stop local development environment
	@echo "Stopping development environment..."
	docker compose down

mod-download: ## Run 'go mod download' using disposable golang:1.25-alpine (no need for running services)
	@echo "Downloading Go modules for all services (using golang:1.25-alpine)..."
	docker run --rm -v "$(CURDIR)":/workspace -w /workspace/services/api-gateway golang:1.25-alpine sh -lc "apk add --no-cache git && export PATH=/usr/local/go/bin:\$$PATH && go mod download && go mod tidy"
	docker run --rm -v "$(CURDIR)":/workspace -w /workspace/services/user-service golang:1.25-alpine sh -lc "apk add --no-cache git && export PATH=/usr/local/go/bin:\$$PATH && go mod download && go mod tidy"
	docker run --rm -v "$(CURDIR)":/workspace -w /workspace/services/order-service golang:1.25-alpine sh -lc "apk add --no-cache git && export PATH=/usr/local/go/bin:\$$PATH && go mod download && go mod tidy"
	docker run --rm -v "$(CURDIR)":/workspace -w /workspace/services/inventory-service golang:1.25-alpine sh -lc "apk add --no-cache git && export PATH=/usr/local/go/bin:\$$PATH && go mod download && go mod tidy"
	docker run --rm -v "$(CURDIR)":/workspace -w /workspace/services/payment-service golang:1.25-alpine sh -lc "apk add --no-cache git && export PATH=/usr/local/go/bin:\$$PATH && go mod download && go mod tidy"

fmt: ## Run gofmt locally across all Go services (requires local Go toolchain)
	@echo "Formatting Go code locally with go fmt..."
	@echo "(Ensure Go is installed and available in PATH)"
	cd services/api-gateway && go fmt ./...
	cd services/user-service && go fmt ./...
	cd services/order-service && go fmt ./...
	cd services/inventory-service && go fmt ./...
	cd services/payment-service && go fmt ./...

# Dependency management via disposable golang container (keeps reproducibility)
DEPS_IMAGE := golang:1.25-alpine

deps-get: ## Update/add a Go module in a service: make deps-get SERVICE=inventory-service MOD=go.uber.org/zap@v1.27.0
	@if [ -z "$(SERVICE)" ] || [ -z "$(MOD)" ]; then \
		echo "Usage: make deps-get SERVICE=<service-dir> MOD=<module@version>"; \
		exit 2; \
	fi
	@echo "Adding/updating module $(MOD) in services/$(SERVICE)"
	docker run --rm -v "$(CURDIR)":/workspace -w /workspace/services/$(SERVICE) $(DEPS_IMAGE) sh -lc "apk add --no-cache git && export PATH=/usr/local/go/bin:\$$PATH && go get $(MOD) && go mod tidy"

deps-tidy: ## Run go mod tidy in all services (no version changes, cleans unused)
	@echo "Running go mod tidy in all services using $(DEPS_IMAGE)..."
	docker run --rm -v "$(CURDIR)":/workspace -w /workspace/services/api-gateway $(DEPS_IMAGE) sh -lc "apk add --no-cache git && export PATH=/usr/local/go/bin:\$$PATH && go mod tidy"
	docker run --rm -v "$(CURDIR)":/workspace -w /workspace/services/user-service $(DEPS_IMAGE) sh -lc "apk add --no-cache git && export PATH=/usr/local/go/bin:\$$PATH && go mod tidy"
	docker run --rm -v "$(CURDIR)":/workspace -w /workspace/services/order-service $(DEPS_IMAGE) sh -lc "apk add --no-cache git && export PATH=/usr/local/go/bin:\$$PATH && go mod tidy"
	docker run --rm -v "$(CURDIR)":/workspace -w /workspace/services/inventory-service $(DEPS_IMAGE) sh -lc "apk add --no-cache git && export PATH=/usr/local/go/bin:\$$PATH && go mod tidy"
	docker run --rm -v "$(CURDIR)":/workspace -w /workspace/services/payment-service $(DEPS_IMAGE) sh -lc "apk add --no-cache git && export PATH=/usr/local/go/bin:\$$PATH && go mod tidy"


update-mod: ## Update modules for all services (download + tidy) and rebuild dev stack
	@echo "Updating Go modules across all services, then rebuilding dev stack..."
	$(MAKE) mod-download
	$(MAKE) dev-rebuild