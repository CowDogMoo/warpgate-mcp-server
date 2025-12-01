.PHONY: help build test clean docker-build docker-run install

# Version information
VERSION ?= 0.1.0
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags="-X 'github.com/cowdogmoo/warpgate-mcp-server/version.Version=$(VERSION)' \
                      -X 'github.com/cowdogmoo/warpgate-mcp-server/version.GitCommit=$(GIT_COMMIT)' \
                      -X 'github.com/cowdogmoo/warpgate-mcp-server/version.BuildDate=$(BUILD_DATE)'"

# Binary name
BINARY_NAME := warpgate-mcp-server
DOCKER_IMAGE := warpgate-mcp-server:dev

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/warpgate-mcp-server

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	docker rmi -f $(DOCKER_IMAGE) 2>/dev/null || true

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build ## Run the server in Docker
	@echo "Running $(DOCKER_IMAGE)..."
	docker run -i --rm \
		-v $(HOME)/cowdogmoo/warpgate:/warpgate:ro \
		-e WARPGATE_PATH=/warpgate \
		$(DOCKER_IMAGE)

install: build ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./cmd/warpgate-mcp-server

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

lint: ## Run linters
	@echo "Running linters..."
	golangci-lint run ./...

tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	go mod tidy

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download

.DEFAULT_GOAL := help
