# Makefile for Microservices Order System (dev only)

.PHONY: help proto proto-clean dev-up dev-rebuild dev-down fmt deps-get deps-tidy mod-download update-mod go-mod-all

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
	
fmt: ## Run gofmt locally across all Go services (requires local Go toolchain)
	@echo "Formatting Go code locally with go fmt..."
	@echo "(Ensure Go is installed and available in PATH)"
	cd services/api-gateway && go fmt ./...
	cd services/user-service && go fmt ./...
	cd services/order-service && go fmt ./...
	cd services/inventory-service && go fmt ./...
	cd services/payment-service && go fmt ./...

go-mod-all: ## Run go mod download && go mod tidy locally for all Go modules
	@echo "Running go mod download && go mod tidy for all modules (locally)..."
	@echo "(Ensure Go is installed and available in PATH)"
	cd libs/kafka && go mod download && go mod tidy
	cd services/api-gateway && go mod download && go mod tidy
	cd services/user-service && go mod download && go mod tidy
	cd services/order-service && go mod download && go mod tidy
	cd services/inventory-service && go mod download && go mod tidy
	cd services/payment-service && go mod download && go mod tidy