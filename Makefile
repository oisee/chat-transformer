# Chat Export Transformer Makefile

# Application name
APP_NAME = chat-transformer

# Version (can be overridden)
VERSION ?= v1.0.0

# Build info
BUILD_TIME = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go build flags
GOFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Build directory
BUILD_DIR = build

# Platforms for cross-compilation
PLATFORMS = \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64

.PHONY: all build build-all clean test lint help

# Default target
all: build

# Build for current platform
build:
	@echo "Building $(APP_NAME) for current platform..."
	go build $(GOFLAGS) -o $(APP_NAME) .
	@echo "✓ Built $(APP_NAME)"

# Build for all platforms
build-all: clean-build
	@echo "Building $(APP_NAME) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		output_name=$(BUILD_DIR)/$(APP_NAME)-$$os-$$arch; \
		if [ $$os = "windows" ]; then output_name=$$output_name.exe; fi; \
		echo "Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build $(GOFLAGS) -o $$output_name .; \
		if [ $$? -eq 0 ]; then \
			echo "✓ Built $$output_name"; \
		else \
			echo "✗ Failed to build for $$os/$$arch"; \
		fi; \
	done
	@echo "✓ All builds completed in $(BUILD_DIR)/"

# Quick build targets for specific platforms
build-linux-amd64:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .

build-linux-arm64:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 .

build-darwin-amd64:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 .

build-darwin-arm64:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 .

build-windows-amd64:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe .

build-windows-arm64:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=arm64 go build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe .

# Development targets
dev: build
	@echo "Running $(APP_NAME) in development mode..."
	./$(APP_NAME)

# Test the application
test:
	@echo "Running tests..."
	go test -v ./...

# Run go mod tidy
tidy:
	@echo "Tidying go modules..."
	go mod tidy

# Lint the code (requires golangci-lint)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running linter..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, running go vet instead..."; \
		go vet ./...; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(APP_NAME)
	rm -rf $(BUILD_DIR)
	@echo "✓ Cleaned"

# Clean only build directory
clean-build:
	@rm -rf $(BUILD_DIR)

# Install the application to GOPATH/bin
install: build
	@echo "Installing $(APP_NAME)..."
	go install $(GOFLAGS) .
	@echo "✓ Installed $(APP_NAME) to $(shell go env GOPATH)/bin/"

# Create a release archive
release: build-all
	@echo "Creating release archives..."
	@mkdir -p $(BUILD_DIR)/releases
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		binary_name=$(APP_NAME)-$$os-$$arch; \
		if [ $$os = "windows" ]; then binary_name=$$binary_name.exe; fi; \
		if [ -f $(BUILD_DIR)/$$binary_name ]; then \
			archive_name=$(APP_NAME)-$(VERSION)-$$os-$$arch; \
			if [ $$os = "windows" ]; then \
				cd $(BUILD_DIR) && zip -q $$archive_name.zip $$binary_name README.md; \
				echo "✓ Created $(BUILD_DIR)/$$archive_name.zip"; \
			else \
				cd $(BUILD_DIR) && tar -czf $$archive_name.tar.gz $$binary_name README.md; \
				echo "✓ Created $(BUILD_DIR)/$$archive_name.tar.gz"; \
			fi; \
		fi; \
	done
	@cp README.md $(BUILD_DIR)/

# Show current platform info
info:
	@echo "Platform Information:"
	@echo "  GOOS: $(shell go env GOOS)"
	@echo "  GOARCH: $(shell go env GOARCH)"
	@echo "  Go Version: $(shell go version)"
	@echo ""
	@echo "Build Information:"
	@echo "  App Name: $(APP_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Git Commit: $(GIT_COMMIT)"

# Show help
help:
	@echo "Chat Export Transformer - Makefile Help"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build              Build for current platform (default)"
	@echo "  build-all          Build for all supported platforms"
	@echo "  dev                Build and run in development mode"
	@echo ""
	@echo "Platform-specific builds:"
	@echo "  build-linux-amd64    Build for Linux AMD64"
	@echo "  build-linux-arm64    Build for Linux ARM64"
	@echo "  build-darwin-amd64   Build for macOS AMD64 (Intel)"
	@echo "  build-darwin-arm64   Build for macOS ARM64 (Apple Silicon)"
	@echo "  build-windows-amd64  Build for Windows AMD64"
	@echo "  build-windows-arm64  Build for Windows ARM64"
	@echo ""
	@echo "Development:"
	@echo "  test               Run tests"
	@echo "  lint               Run linter (golangci-lint or go vet)"
	@echo "  tidy               Run go mod tidy"
	@echo ""
	@echo "Release:"
	@echo "  release            Build all platforms and create archives"
	@echo "  install            Install to GOPATH/bin"
	@echo ""
	@echo "Utility:"
	@echo "  clean              Clean all build artifacts"
	@echo "  clean-build        Clean only build directory"
	@echo "  info               Show platform and build information"
	@echo "  help               Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION            Override version (default: v1.0.0)"