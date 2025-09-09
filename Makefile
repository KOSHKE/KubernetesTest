# Makefile for Microservices Order System (dev only)

.PHONY: help proto proto-clean dev-up dev-rebuild dev-down fmt deps-get deps-tidy mod-download update-mod go-mod-all build-service build-all monitoring-up monitoring-down test test-unit test-integration test-coverage generate-mocks install-test-deps

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
	@docker run --rm -v "$(CURDIR)":/workspace -w /workspace/services/inventory-service/tests/testdata/interfaces $(GO_TEST_IMAGE) sh -c "go install go.uber.org/mock/mockgen@latest && go generate"
	@echo "Mocks generated successfully!"

test: ## Run all tests across all services (Docker-based)
	@echo "Running all tests (via $(GO_TEST_IMAGE))..."
	@docker run --rm -v "$(CURDIR)":/workspace -w /workspace/services/inventory-service $(GO_TEST_IMAGE) sh -c "go mod download && go test ./tests/..."
	@echo "All tests completed!"

test-coverage: ## Run tests with coverage report (Docker-based)
	@echo "Running tests with coverage (via $(GO_TEST_IMAGE))..."
	@docker run --rm -v "$(CURDIR)":/workspace -w /workspace/services/inventory-service $(GO_TEST_IMAGE) sh -c "go mod download && go test -coverprofile=coverage.out ./tests/... && go tool cover -html=coverage.out -o coverage.html"
	@echo "Coverage report generated: services/inventory-service/coverage.html"