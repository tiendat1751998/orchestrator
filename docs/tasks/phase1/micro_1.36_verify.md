# Micro-Task 1.36: Verification — Build & Test toàn bộ Phase 1

## Thông tin
- **File tạo**: Không tạo file nào (chỉ verify)
- **Dependencies trước**: TẤT CẢ micro-tasks 1.01 → 1.35
- **Thời gian**: 15 phút
- **Mục đích**: Đảm bảo TẤT CẢ files đã tạo compile và test thành công

## Lệnh verify (PHẢI chạy theo đúng thứ tự)

### Step 1: Kiểm tra files tồn tại
```bash
# Kiểm tra 35 files tồn tại
ls go.mod .gitignore Makefile .golangci.yml
ls contracts/errors.go contracts/types.go contracts/status.go
ls contracts/provider/message.go contracts/provider/request.go
ls contracts/provider/response.go contracts/provider/config.go
ls contracts/provider/provider.go contracts/provider/provider_test.go
ls contracts/tool/schema.go contracts/tool/tool.go contracts/tool/tool_test.go
ls contracts/agent/capability.go contracts/agent/task.go contracts/agent/result.go
ls contracts/agent/manifest.go contracts/agent/agent.go contracts/agent/agent_test.go
ls contracts/event/event.go contracts/plugin/plugin.go
ls contracts/memory/memory.go contracts/search/search.go
ls contracts/workflow/workflow.go contracts/context/context.go
ls contracts/planner/planner.go contracts/orchestrator/orchestrator.go
ls contracts/resilience/resilience.go contracts/security/security.go
ls contracts/gateway/gateway.go contracts/feedback/feedback.go
ls cmd/orchestrator/main.go
```

### Step 2: Go Build (kiểm tra compile)
```bash
go build ./...
# Expected: không output, không error
# Nếu lỗi: check import paths, package names, syntax
```

### Step 3: Go Vet (kiểm tra code quality)
```bash
go vet ./...
# Expected: không output, không warning
# Nếu lỗi: sửa theo warning message
```

### Step 4: Go Test (chạy unit tests)
```bash
go test -v ./contracts/...
# Expected: ALL PASS
# Tối thiểu 30+ test functions
```

### Step 5: Go Test với Race Detector
```bash
go test -race ./contracts/...
# Expected: ALL PASS, no race conditions detected
# Race detector chạy chậm hơn ~10x, nhưng phát hiện data races
```

### Step 6: Coverage Report
```bash
go test -coverprofile=coverage.out ./contracts/...
go tool cover -func=coverage.out
# Expected: ≥ 70% coverage trên các package có test
```

### Step 7: Build Binary
```bash
go build -o bin/orchestrator ./cmd/orchestrator/
./bin/orchestrator
# Expected output:
# orchestrator v0.1.0-dev
# Use 'orchestrator --help' for usage information.
```

### Step 8: Import Cycle Check
```bash
go build ./contracts/...
# Nếu có import cycle, lệnh này sẽ fail với:
# "import cycle not allowed"
# 
# Import graph hợp lệ (KHÔNG có vòng):
#   contracts (errors, types, status) ← provider ← agent ← planner ← orchestrator
#   contracts ← event, plugin, memory, search, workflow, context
#   contracts ← security, gateway, feedback, resilience
```

### Step 9: Git Commit
```bash
git add -A
git commit -m "Phase 1: Complete contracts foundation (36 micro-tasks)"
git push origin main
```

## Checklist tổng Phase 1

### Project Files (4)
- [ ] `go.mod` — Go 1.26, module path đúng
- [ ] `.gitignore` — ignores: .env, bin/, IDE files
- [ ] `Makefile` — targets: build, test, lint, clean (TAB indent)
- [ ] `.golangci.yml` — 10+ linters configured

### Contracts — Shared (3 files)
- [ ] `contracts/errors.go` — ≥ 15 sentinel errors, `Err` prefix
- [ ] `contracts/types.go` — 6 named ID types, `New*ID()` functions
- [ ] `contracts/status.go` — 8 status constants, `IsTerminal()`, `IsValid()`

### Contracts — Provider (6 files)
- [ ] `contracts/provider/message.go` — Message, ToolCall, 4 Role constants, 4 helper constructors
- [ ] `contracts/provider/request.go` — Request, ToolDefinition, `*float64/*int` pointers, `Validate()`
- [ ] `contracts/provider/response.go` — Response, StreamChunk, Usage, `Add()`, `ToMessage()`
- [ ] `contracts/provider/config.go` — Config, `APIKey json:"-"`, `GetExtra()`
- [ ] `contracts/provider/provider.go` — Provider interface (4 methods), `<-chan StreamChunk`
- [ ] `contracts/provider/provider_test.go` — ≥ 15 tests, JSON round-trip, security

### Contracts — Tool (3 files)
- [ ] `contracts/tool/schema.go` — Schema, Property, builder pattern
- [ ] `contracts/tool/tool.go` — Tool interface, Result, `json.RawMessage` args
- [ ] `contracts/tool/tool_test.go` — ≥ 7 tests

### Contracts — Agent (6 files)
- [ ] `contracts/agent/capability.go` — 10 capabilities, `IsValid()`
- [ ] `contracts/agent/task.go` — Task, ContextItem, `[]contracts.TaskID` dependencies
- [ ] `contracts/agent/result.go` — Result, Artifact, `*provider.Usage`, constructors
- [ ] `contracts/agent/manifest.go` — Manifest, 12 fields, YAML+JSON tags
- [ ] `contracts/agent/agent.go` — Agent interface (4 methods), error convention documented
- [ ] `contracts/agent/agent_test.go` — ≥ 10 tests, JSON round-trip

### Contracts — Others (12 files)
- [ ] `contracts/event/event.go` — Event, Bus (Pub/Sub), 10 event constants
- [ ] `contracts/plugin/plugin.go` — Plugin interface (7 methods), lifecycle order
- [ ] `contracts/memory/memory.go` — Store, functional options
- [ ] `contracts/search/search.go` — Engine, Indexable, functional options
- [ ] `contracts/workflow/workflow.go` — Workflow, Step, Result
- [ ] `contracts/context/context.go` — Builder, Item (⚠️ package name: `agentcontext`)
- [ ] `contracts/planner/planner.go` — Planner, Mission
- [ ] `contracts/orchestrator/orchestrator.go` — Orchestrator, MissionResult, MissionStatus
- [ ] `contracts/resilience/resilience.go` — CircuitBreaker, RetryPolicy, Fallback
- [ ] `contracts/security/security.go` — PermissionManager, AuditLogger
- [ ] `contracts/gateway/gateway.go` — Gateway (Start, Stop, Address)
- [ ] `contracts/feedback/feedback.go` — Evaluator, Scorer, AgentScore

### Entry Point (1 file)
- [ ] `cmd/orchestrator/main.go` — prints version

### Quality Gates
- [ ] `go build ./...` ✅ (no errors)
- [ ] `go vet ./...` ✅ (no warnings)
- [ ] `go test ./contracts/...` ALL PASS
- [ ] `go test -race ./contracts/...` NO RACES
- [ ] Coverage ≥ 70%
- [ ] No import cycles
- [ ] Git commit + push thành công

### File Count: 35 Go files + 4 config files = 39 files total
