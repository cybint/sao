GO ?= go

.PHONY: help build plugins run test lint

help: ## Show available make targets
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_-]+:.*## / {printf "%-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build only the sao binary
	mkdir -p bin
	$(GO) build -o bin/sao ./cmd/sao

plugins: ## Build optional plugin binaries
	mkdir -p bin
	$(GO) build -o bin/example ./cmd/example
	$(GO) build -o bin/ui ./cmd/ui

run: ## Run the SAO server locally
	$(GO) run ./cmd/sao

test: ## Run all unit tests
	$(GO) test ./...

lint: ## Run static analysis checks
	$(GO) vet ./...
