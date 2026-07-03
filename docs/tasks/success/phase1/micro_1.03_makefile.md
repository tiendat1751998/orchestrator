# Micro-Task 1.03: Create Makefile

## Info
- **File**: `Makefile`
- **Depends on**: 1.01
- **Time**: 10 min
- **Verify**: `make help`

## Purpose
Establishes a build automation suite (Makefile) providing shortcuts for compilation, linting, formatting, testing, and test coverage outputs.

## EXACT code to create

> [!WARNING]
> Makefile syntax strictly requires using **TAB character indentations** for target command blocks. Do not use spaces.

```makefile
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
```

## ⚠️ Pitfalls

### Pitfall 1: Space Indentation Failure (Makefile Syntax Error)
```makefile
build:
	go build -o bin/orchestrator ./cmd/orchestrator # Indented with 1 raw TAB character.
```
Make requires command lines to start with a tab. Configure your IDE/editor to not translate tabs to spaces for Makefile types.

### Pitfall 2: Shell commands incompatibility on Windows environments
`rm -rf` or other Unix CLI tools fail on native Windows command prompts (`cmd.exe`). In Windows environments, always invoke Make targets inside a POSIX-compliant terminal shell (such as Git Bash, WSL, or MSYS2) rather than standard cmd.

## Verify
```bash
make help
```

## Checklist
- [ ] File `Makefile` exists at root directory
- [ ] Commands are indented using TABs
- [ ] `.PHONY` targets are declared for all commands
- [ ] `build` places compiled outputs in `bin/` directory
- [ ] `test-race` enforces `-race` thread-safety checking
- [ ] `test-cover` output is cleanly configured to export `coverage.html`
- [ ] `make help` executes successfully
