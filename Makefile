# Makefile for Microservices Order System (dev only)

.PHONY: help proto proto-clean dev-up dev-rebuild dev-down test generate-mocks

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

BUF_IMAGE ?= bufbuild/buf:latest
GO_TEST_IMAGE ?= golang:1.25-bookworm

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

generate-mocks: ## Generate mocks for all services using gomock (Docker-based)
	@echo "Generating mocks with gomock (via $(GO_TEST_IMAGE))..."
	@docker run --rm \
		-v "$(CURDIR)":/workspace \
		-v go-cache:/go/pkg/mod \
		-v go-build-cache:/root/.cache/go-build \
		$(GO_TEST_IMAGE) sh -c " \
			go install go.uber.org/mock/mockgen@latest && \
			cd /workspace/services/inventory-service && \
			go generate ./tests/mocks/generate.go && \
			cd /workspace/services/user-service && \
			go generate ./tests/mocks/generate.go && \
			cd /workspace/services/order-service && \
			go generate ./tests/mocks/generate.go \
		"
	@echo "Mocks generated successfully!"

test: ## Run all tests locally
	@echo "Running all tests locally..."
	@cd services/inventory-service && go test ./tests/...
	@cd services/inventory-service && go test -tags=integration -v ./tests/integration/...
	@cd services/user-service && go test ./tests/...
	@cd services/user-service && go test -tags=integration -v ./tests/integration/...
	@cd services/order-service && go test ./tests/unit/...
	@echo "All tests completed!"

test-user-integration: ## Run user-service integration tests
	@echo "Running user-service integration tests..."
	@cd services/user-service && go test -tags=integration -v ./tests/integration/...
	@echo "User-service integration tests completed!"