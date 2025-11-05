.PHONY: help build install test clean run fmt lint vet tidy release

# Variables
BINARY_NAME=kubetui
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_TIME)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build directory
BUILD_DIR=bin

## help: Display this help message
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  install     - Install the binary to /usr/local/bin"
	@echo "  test        - Run tests"
	@echo "  clean       - Remove build artifacts"
	@echo "  run         - Build and run the application"
	@echo "  fmt         - Format code"
	@echo "  lint        - Run linters"
	@echo "  vet         - Run go vet"
	@echo "  tidy        - Tidy go modules"
	@echo "  release     - Build for multiple platforms"

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/kubetui/
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## install: Install the binary to /usr/local/bin
install:
	@if [ ! -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		echo "Error: Binary not found. Please run 'make build' first (without sudo)"; \
		exit 1; \
	fi
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete"

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## run: Build and run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "Format complete"

## lint: Run linters (requires golangci-lint)
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it from https://golangci-lint.run/usage/install/" && exit 1)
	@golangci-lint run ./...
	@echo "Lint complete"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "Vet complete"

## tidy: Tidy go modules
tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy
	@echo "Tidy complete"

## release: Build for multiple platforms
release:
	@echo "Building releases..."
	@mkdir -p $(BUILD_DIR)/releases

	# Linux amd64
	@echo "Building for Linux amd64..."
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-linux-amd64 ./cmd/kubetui/

	# Linux arm64
	@echo "Building for Linux arm64..."
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-linux-arm64 ./cmd/kubetui/

	# macOS amd64
	@echo "Building for macOS amd64..."
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-darwin-amd64 ./cmd/kubetui/

	# macOS arm64 (Apple Silicon)
	@echo "Building for macOS arm64..."
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-darwin-arm64 ./cmd/kubetui/

	# Windows amd64
	@echo "Building for Windows amd64..."
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-windows-amd64.exe ./cmd/kubetui/

	@echo "Release builds complete in $(BUILD_DIR)/releases/"

## dev: Run in development mode with hot reload (requires air)
dev:
	@which air > /dev/null || (echo "air not found. Install it with: go install github.com/cosmtrek/air@latest" && exit 1)
	@air

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):$(VERSION) .
	@echo "Docker image built: $(BINARY_NAME):$(VERSION)"

## check: Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "All checks passed!"

.DEFAULT_GOAL := help
