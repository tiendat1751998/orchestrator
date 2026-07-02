# Micro-Tasks 1.35 & 1.36: Main Skeleton & Verification

---

# Micro-Task 1.35: cmd/orchestrator/main.go

## Thông tin
- **File tạo**: `cmd/orchestrator/main.go`
- **Package**: `main`
- **Thời gian**: 5 phút

## Nội dung CHÍNH XÁC

```go
// Package main is the entry point for the orchestrator CLI.
// Full implementation will be added in Phase 6.
package main

import "fmt"

func main() {
	fmt.Println("orchestrator v0.1.0-dev")
	fmt.Println("Use 'orchestrator --help' for usage information.")
}
```

## Mục đích
- Đảm bảo `go build ./...` hoạt động
- Placeholder cho Phase 6 (CLI implementation)

## Checklist
- [ ] File tồn tại
- [ ] `go build -o bin/orchestrator ./cmd/orchestrator/` thành công
- [ ] `./bin/orchestrator` in ra version

---

# Micro-Task 1.36: Verification — Build & Test toàn bộ Phase 1

## Lệnh verify (PHẢI chạy theo thứ tự)

### Step 1: Build
```bash
go build ./...
# Expected: không output, không error
```

### Step 2: Vet
```bash
go vet ./...
# Expected: không output, không warning
```

### Step 3: Test
```bash
go test -v ./contracts/...
# Expected: ALL PASS, ≥ 25 test functions
```

### Step 4: Test Race
```bash
go test -race ./contracts/...
# Expected: ALL PASS, no race conditions
```

### Step 5: Coverage
```bash
go test -coverprofile=coverage.out ./contracts/...
go tool cover -func=coverage.out
# Expected: ≥ 70% coverage
```

### Step 6: Build binary
```bash
go build -o bin/orchestrator ./cmd/orchestrator/
./bin/orchestrator
# Expected: "orchestrator v0.1.0-dev"
```

### Step 7: Check no import cycles
```bash
go build ./contracts/...
# If import cycles exist, this will fail with:
# "import cycle not allowed"
```

## Checklist tổng Phase 1

### Project Files
- [ ] `go.mod` — module path đúng
- [ ] `.gitignore` — covers all needed patterns
- [ ] `Makefile` — targets: build, test, lint, clean
- [ ] `.golangci.yml` — linter config

### Contracts — Shared (3 files)
- [ ] `contracts/errors.go` — ≥ 15 sentinel errors
- [ ] `contracts/types.go` — 6 named ID types
- [ ] `contracts/status.go` — 8 status constants

### Contracts — Provider (6 files)
- [ ] `contracts/provider/message.go` — Message, ToolCall, Role
- [ ] `contracts/provider/request.go` — Request, ToolDefinition, validation
- [ ] `contracts/provider/response.go` — Response, StreamChunk, Usage
- [ ] `contracts/provider/config.go` — Config struct
- [ ] `contracts/provider/provider.go` — Provider interface (4 methods)
- [ ] `contracts/provider/provider_test.go` — ≥ 15 tests

### Contracts — Tool (3 files)
- [ ] `contracts/tool/schema.go` — Schema, Property, builders
- [ ] `contracts/tool/tool.go` — Tool interface, Result
- [ ] `contracts/tool/tool_test.go` — ≥ 7 tests

### Contracts — Agent (6 files)
- [ ] `contracts/agent/capability.go` — 10 capabilities
- [ ] `contracts/agent/task.go` — Task, ContextItem
- [ ] `contracts/agent/result.go` — Result, Artifact
- [ ] `contracts/agent/manifest.go` — Manifest struct
- [ ] `contracts/agent/agent.go` — Agent interface (4 methods)
- [ ] `contracts/agent/agent_test.go` — ≥ 10 tests

### Contracts — Others (12 files)
- [ ] `contracts/event/event.go`
- [ ] `contracts/plugin/plugin.go`
- [ ] `contracts/memory/memory.go`
- [ ] `contracts/search/search.go`
- [ ] `contracts/workflow/workflow.go`
- [ ] `contracts/context/context.go`
- [ ] `contracts/planner/planner.go`
- [ ] `contracts/orchestrator/orchestrator.go`
- [ ] `contracts/resilience/resilience.go`
- [ ] `contracts/security/security.go`
- [ ] `contracts/gateway/gateway.go`
- [ ] `contracts/feedback/feedback.go`

### Entry Point
- [ ] `cmd/orchestrator/main.go`

### Quality
- [ ] `go build ./...` ✅
- [ ] `go vet ./...` ✅
- [ ] `go test ./contracts/...` ALL PASS
- [ ] `go test -race ./contracts/...` NO RACES
- [ ] Coverage ≥ 70%
- [ ] No import cycles
- [ ] Git commit: `"Phase 1: Complete contracts foundation (36 micro-tasks)"`

### Tổng: 31 Go files + 4 config files = 35 files
