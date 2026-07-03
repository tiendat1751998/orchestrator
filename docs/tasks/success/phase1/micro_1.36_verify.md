# Micro-Task 1.36: Verification — Complete Phase 1 Build & Test

## Info
- **File**: N/A (Verification instructions file)
- **Depends on**: ALL micro-tasks 1.01 → 1.35
- **Time**: 15 min
- **Purpose**: Validates compilation, unit testing coverage, and directory structure consistency across all Phase 1 contracts.

## Verification Steps (Execute in exact order)

### Step 1: File Existence Auditing
Verify that all 35 Go code files and 4 configuration files exist:
```bash
# Check configuration and setup targets
ls go.mod .gitignore Makefile .golangci.yml

# Check base shared contracts
ls contracts/errors.go contracts/types.go contracts/status.go

# Check model provider adapter contracts
ls contracts/provider/message.go contracts/provider/request.go
ls contracts/provider/response.go contracts/provider/config.go
ls contracts/provider/provider.go contracts/provider/provider_test.go

# Check tool contracts
ls contracts/tool/schema.go contracts/tool/tool.go contracts/tool/tool_test.go

# Check agent contracts
ls contracts/agent/capability.go contracts/agent/task.go contracts/agent/result.go
ls contracts/agent/manifest.go contracts/agent/agent.go contracts/agent/agent_test.go

# Check other modular components contracts
ls contracts/event/event.go contracts/plugin/plugin.go
ls contracts/memory/memory.go contracts/search/search.go
ls contracts/workflow/workflow.go contracts/context/context.go
ls contracts/planner/planner.go contracts/orchestrator/orchestrator.go
ls contracts/resilience/resilience.go contracts/security/security.go
ls contracts/gateway/gateway.go contracts/feedback/feedback.go

# Check entry point
ls cmd/orchestrator/main.go
```

### Step 2: Go Build Validation (Compilation Check)
```bash
go build ./...
```
Expected outcome: Empty output, no compilation errors. If compilation fails, double check import path capitalization and interface signatures.

### Step 3: Go Vet Inspection
```bash
go vet ./...
```
Expected outcome: Empty output, no warnings. Correct any issues found.

### Step 4: Unit Test Runs
```bash
go test -v ./contracts/...
```
Expected outcome: All tests pass. Confirms serialization behaviors and builder defaults are correct.

### Step 5: Concurrency Safety Run
```bash
go test -race ./contracts/...
```
Expected outcome: All tests pass, and zero data races are flagged by the runtime race detector.

### Step 6: Coverage Metrics Review
```bash
go test -coverprofile=coverage.out ./contracts/...
go tool cover -func=coverage.out
```
Expected outcome: Checks coverage levels.

### Step 7: Executable Verification
```bash
go build -o bin/orchestrator ./cmd/orchestrator/
./bin/orchestrator
```
Expected output:
```
orchestrator v0.1.0-dev
Use 'orchestrator --help' for usage information.
```

### Step 8: Package Dependency Structure Check
Check for package dependency loops:
```bash
go build ./contracts/...
```
Any dependency cycles will trigger an import cycle compiler error.
Valid dependency flows (acyclic):
- `contracts (errors, types, status) ← provider ← agent ← planner ← orchestrator`
- `contracts ← event, plugin, memory, search, workflow, context`
- `contracts ← security, gateway, feedback, resilience`

### Step 9: Version Control Commit
```bash
git add -A
git commit -m "Phase 1: Complete contracts foundation"
git push origin main
```

## Quality Gates Checklist

### Project Files (4)
- [ ] `go.mod` pins compiler to Go 1.26.3
- [ ] `.gitignore` excludes binary/IDE/secret traces
- [ ] `Makefile` targets build, test, lint, clean using tab indents
- [ ] `.golangci.yml` enables 10 essential safety linters

### Contracts — Shared (3 files)
- [ ] `contracts/errors.go` defines sentinel errors with `Err` prefixes
- [ ] `contracts/types.go` defines named type ID wrappers
- [ ] `contracts/status.go` handles state enums and terminal status helpers

### Contracts — Provider (6 files)
- [ ] `contracts/provider/message.go` defines messages and tool calls
- [ ] `contracts/provider/request.go` defines request configuration parameters
- [ ] `contracts/provider/response.go` handles provider response payloads
- [ ] `contracts/provider/config.go` holds provider config values (api keys hidden in JSON)
- [ ] `contracts/provider/provider.go` declares core Provider adapters interface
- [ ] `contracts/provider/provider_test.go` verifies validation and precision behaviors

### Contracts — Tool (3 files)
- [ ] `contracts/tool/schema.go` builds JSON schemas
- [ ] `contracts/tool/tool.go` defines the Tool execution interface
- [ ] `contracts/tool/tool_test.go` verifies schema builders and output string helpers

### Contracts — Agent (6 files)
- [ ] `contracts/agent/capability.go` handles agent capabilities
- [ ] `contracts/agent/task.go` builds Task structures and context items
- [ ] `contracts/agent/result.go` outputs Result structures and lists artifact items
- [ ] `contracts/agent/manifest.go` parses Agent YAML manifest structures
- [ ] `contracts/agent/agent.go` declares the core Agent persona interface
- [ ] `contracts/agent/agent_test.go` runs serialization validation tests

### Contracts — Other Component Specifications (12 files)
- [ ] `contracts/event/event.go` declares Event and Bus structures
- [ ] `contracts/plugin/plugin.go` specifies lifecycle phases and methods
- [ ] `contracts/memory/memory.go` defines storage options and enums
- [ ] `contracts/search/search.go` handles index engine targets
- [ ] `contracts/workflow/workflow.go` builds workflow steps
- [ ] `contracts/context/context.go` builds context items under package `agentcontext`
- [ ] `contracts/planner/planner.go` defines planning and replanning boundaries
- [ ] `contracts/orchestrator/orchestrator.go` defines the orchestrator coordination engine interface
- [ ] `contracts/resilience/resilience.go` manages breakers, retries, and fallbacks
- [ ] `contracts/security/security.go` checks permissions and audit trails
- [ ] `contracts/gateway/gateway.go` runs listener gateways
- [ ] `contracts/feedback/feedback.go` scores agent executions

### CLI Setup
- [ ] `cmd/orchestrator/main.go` prints compiler verification tags

### Compilation Verification
- [ ] `go build ./...` passes
- [ ] `go vet ./...` reports zero static issues
- [ ] `go test ./contracts/...` runs and passes
- [ ] No circular package import dependencies exist
