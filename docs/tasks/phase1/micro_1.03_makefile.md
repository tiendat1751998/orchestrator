# Micro-Task 1.03: Tạo Makefile

## Thông tin
- **File tạo**: `Makefile`
- **Dependencies trước**: 1.01
- **Thời gian**: 10 phút
- **Verify**: `make build` chạy không lỗi (sau khi có main.go)

## Nội dung CHÍNH XÁC cần tạo

> ⚠️ QUAN TRỌNG: Makefile dùng TAB (không phải SPACE) để indent. Mỗi dòng command PHẢI bắt đầu bằng 1 TAB character.

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

## Quy tắc
1. TAB indent, KHÔNG PHẢI SPACE — Makefile syntax bắt buộc
2. `.PHONY` liệt kê TẤT CẢ targets — đảm bảo luôn chạy kể cả có file cùng tên
3. `LDFLAGS := -s -w` — strip debug info, giảm binary size
4. `test-race` riêng biệt vì race detector chậm hơn, không chạy mặc định
5. `all` target chạy: fmt → vet → lint → test → build (đúng thứ tự)

## ⚠️ Pitfalls cần tránh
1. **SPACE thay TAB**: Editor có thể auto-convert TAB → SPACE. Kiểm tra bằng `cat -A Makefile` (Linux) hoặc mở bằng editor hiển thị whitespace
2. **Windows compatibility**: `rm -rf` không có trên Windows cmd. Giải pháp: dùng Git Bash, WSL, hoặc thêm Windows-specific clean target
3. **Missing golangci-lint**: `make lint` sẽ fail nếu chưa install. Thêm check hoặc hướng dẫn install

## Checklist
- [ ] File `Makefile` tồn tại ở root
- [ ] Dùng TAB indent (không SPACE)
- [ ] Có `.PHONY` declaration
- [ ] Target `build` tạo binary vào `bin/`
- [ ] Target `test-race` dùng `-race` flag
- [ ] Target `test-cover` tạo coverage report
- [ ] Target `clean` xóa build artifacts
