# Makefile for Microservices Order System (dev only)

.PHONY: help proto proto-clean dev-up dev-rebuild dev-down

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