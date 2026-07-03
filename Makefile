# Orchestrator Makefile
# =====================

# Variables
BINARY_NAME := orchestrator
BUILD_DIR := bin
MAIN_PACKAGE := ./cmd/orchestrator
COVERAGE_FILE := coverage.out

# Go commands
GOTEST := go test
GOBUILD := go build
GOVET := go vet
GOFMT := gofmt

# Build flags
LDFLAGS := -s -w

.PHONY: all build test test-race test-cover lint clean run fmt vet help

# Default target
all: fmt vet lint test build

# Build binary
build:
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with race detector
test-race:
	$(GOTEST) -v -race ./...

# Run tests with coverage
test-cover:
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report: coverage.html"

# Run linter
lint:
	golangci-lint run ./...

# Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE) coverage.html

# Run the application
run:
	go run $(MAIN_PACKAGE)

# Format code
fmt:
	$(GOFMT) -w -s .

# Run go vet
vet:
	$(GOVET) ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build binary"
	@echo "  test       - Run tests"
	@echo "  test-race  - Run tests with race detector"
	@echo "  test-cover - Run tests with coverage report"
	@echo "  lint       - Run linter"
	@echo "  clean      - Remove build artifacts"
	@echo "  run        - Run the application"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  help       - Show this help"
