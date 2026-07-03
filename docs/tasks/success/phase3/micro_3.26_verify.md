# Micro-Task 3.26: Verification — Build & Test All Phase 3

## Info
- **File**: None (verification task definition only)
- **Depends on**: ALL micro-tasks 3.01 → 3.25
- **Time**: 15 min
- **Purpose**: Ensures that all SDK components (Base classes, Middlewares, and Helpers) compile successfully, contain no circular package import loops, and pass all verification tests with the race detector.

## Verification Commands (Execute in exact order)

### Step 1: Verify SDK Files Exist
```bash
# Plugin SDK
ls sdk/plugin/plugin.go

# Agent SDK
ls sdk/agent/manifest.go
ls sdk/agent/prompt.go
ls sdk/agent/agent.go
ls sdk/agent/agent_test.go

# Provider SDK
ls sdk/provider/provider.go
ls sdk/provider/request.go
ls sdk/provider/stream.go
ls sdk/provider/provider_test.go

# Tool SDK
ls sdk/tool/tool.go
ls sdk/tool/result.go
ls sdk/tool/tool_test.go

# Support Skeletons & Workflows
ls sdk/workflow/workflow.go
ls sdk/workflow/state.go
ls sdk/workflow/workflow_test.go
ls sdk/context/builder.go
ls sdk/memory/memory.go
ls sdk/search/search.go
ls sdk/task/task.go

# Middlewares & Helpers
ls sdk/middleware/agent.go
ls sdk/middleware/provider.go
ls sdk/middleware/middleware_test.go
ls sdk/helpers/ratelimit.go

# Testing SDK Mocks
ls sdk/testing/mocks.go
ls sdk/testing/mocks_test.go
```

### Step 2: Go Build (Compiler Check)
```bash
go build ./sdk/...
# Expected: no output, exit code 0
```

### Step 3: Go Vet (Linter Check)
```bash
go vet ./sdk/...
# Expected: no output, exit code 0
```

### Step 4: Go Test (Unit Tests)
```bash
go test -v ./sdk/...
# Expected: ALL PASS
```

### Step 5: Go Test with Race Detector
```bash
go test -race ./sdk/...
# Expected: ALL PASS, no race conditions detected
```

### Step 6: Import Cycle Check
```bash
go build ./...
# Ensure no import cycles exist between contracts, kernel, and sdk.
# Valid import graph:
#   contracts/ <- kernel/ <- sdk/
#   sdk/ may only import from contracts/ or kernel/
#   sdk/ must never import from plugins/, modules/, or api/
```

### Step 7: Git Commit
```bash
git add -A
git commit -m "Phase 3: SDK Developer Helpers implementation (26 micro-tasks)"
git push origin main
```

## Phase 3 Quality Checklist

### SDK Core Packages
- [ ] `sdk/plugin/plugin.go` — `BasePlugin` tracks initialized and started states and implements health reports.
- [ ] `sdk/agent/manifest.go` — `LoadManifest` resolves system prompt files relative to manifest directories.
- [ ] `sdk/agent/prompt.go` — `BuildPrompt` separates instruction prompts and context items into separate messages.
- [ ] `sdk/agent/agent.go` — `BaseAgent` handles ReAct loops, streaming callbacks, and parallel tool calls.
- [ ] `sdk/provider/provider.go` — `BaseProvider` copies model slices before returning.
- [ ] `sdk/provider/request.go` — `RequestBuilder` implements immutable fluent API methods.
- [ ] `sdk/provider/stream.go` — `CollectStream` drains channels in the background on cancellations to prevent leaks.
- [ ] `sdk/tool/tool.go` — `BaseTool` validates arguments against schema, checking float64 integer bounds.
- [ ] `sdk/tool/result.go` — Result builder `JSON` helper returns system errors on marshalling failures.
- [ ] `sdk/workflow/state.go` — `WorkflowState` resolves template parameters (`{{ inputs.Key }}`) and step outputs dynamically.
- [ ] `sdk/middleware/agent.go` — Implements logging, metrics, and panic recovery decorators.
- [ ] `sdk/middleware/provider.go` — Implements logging, retry, circuit breaker, and token metrics decorators.
- [ ] `sdk/helpers/ratelimit.go` — Token Bucket rate limiter implements mutex release before waiting.
- [ ] `sdk/testing/mocks.go` — Test mocks protect shared state arrays using RWMutex.

### Quality Gates
- [ ] `go build ./...` ✅ (clean compilation of entire workspace)
- [ ] `go test ./sdk/...` ALL PASS
- [ ] `go test -race ./sdk/...` NO RACES
- [ ] No circular package imports are found
- [ ] Git commit and push succeeds
