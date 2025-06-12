.PHONY: build build-all test clean wire lint help

# Variables
BINARY_NAME=mcp-installer
MAIN_PATH=./cmd
DIST_DIR=dist
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Build flags
LDFLAGS=-ldflags="-s -w -X github.com/kirha-ai/mcp-installer/cmd/cli.version=$(VERSION) -X github.com/kirha-ai/mcp-installer/cmd/cli.commit=$(COMMIT) -X github.com/kirha-ai/mcp-installer/cmd/cli.date=$(DATE)"

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

wire: ## Generate Wire dependency injection code
	@echo "Generating Wire code..."
	go generate ./...

test: wire ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run --timeout=5m

build: wire ## Build binary for current platform
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

build-all: wire ## Build binaries for all platforms
	@echo "Building binaries for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(DIST_DIR)/linux_amd64/$(BINARY_NAME) $(MAIN_PATH)
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(DIST_DIR)/linux_arm64/$(BINARY_NAME) $(MAIN_PATH)
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(DIST_DIR)/darwin_amd64/$(BINARY_NAME) $(MAIN_PATH)
	
	# macOS ARM64
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(DIST_DIR)/darwin_arm64/$(BINARY_NAME) $(MAIN_PATH)
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(DIST_DIR)/windows_amd64/$(BINARY_NAME).exe $(MAIN_PATH)
	
	@echo "Build complete. Binaries available in $(DIST_DIR)/"

npm-package: build-all ## Prepare NPM package
	@echo "Preparing NPM package..."
	@rm -rf npm-package
	@mkdir -p npm-package/binaries
	@cp -r pkg/npm/* npm-package/
	@cp -r $(DIST_DIR)/* npm-package/binaries/
	@echo "NPM package ready in npm-package/"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(DIST_DIR)
	rm -rf npm-package
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	find . -name "wire_gen.go" -delete

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

check: wire lint test ## Run all checks (lint + test)

dev-setup: install-tools deps ## Set up development environment
	@echo "Development environment ready!"

# Docker targets
docker-build: ## Build Docker image
	docker build -t $(BINARY_NAME):$(VERSION) .

docker-run: docker-build ## Run Docker container
	docker run --rm -it $(BINARY_NAME):$(VERSION)

# Release targets
release: clean check build-all npm-package ## Prepare release artifacts
	@echo "Release artifacts ready!"