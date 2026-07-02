# Micro-Task 2.41: Phase 2 Verification — Build & Test All Kernel Code

## Info
- **File created**: None (verification only)
- **Depends on**: ALL micro-tasks 2.01 → 2.40
- **Time**: 15 min
- **Purpose**: Ensure ALL kernel code compiles, passes tests, and has no race conditions

## Verification Steps (MUST run in this exact order)

### Step 1: Check all files exist
```bash
# Config (6 files)
ls kernel/config/config.go
ls kernel/config/defaults.go
ls kernel/config/env.go
ls kernel/config/loader.go
ls kernel/config/validator.go
ls kernel/config/watcher.go
ls kernel/config/config_test.go

# Logger (5 files)
ls kernel/logger/logger.go
ls kernel/logger/fields.go
ls kernel/logger/formatter.go
ls kernel/logger/redact.go
ls kernel/logger/logger_test.go

# EventBus (7 files)
ls kernel/eventbus/types.go
ls kernel/eventbus/matcher.go
ls kernel/eventbus/subscriber.go
ls kernel/eventbus/bus.go
ls kernel/eventbus/helpers.go
ls kernel/eventbus/dlq.go
ls kernel/eventbus/bus_test.go

# Registry (4 files)
ls kernel/registry/registry.go
ls kernel/registry/finder.go
ls kernel/registry/lifecycle.go
ls kernel/registry/registry_test.go

# Runtime (5 files)
ls kernel/runtime/executor.go
ls kernel/runtime/pool.go
ls kernel/runtime/dispatcher.go
ls kernel/runtime/runtime.go
ls kernel/runtime/runtime_test.go

# Scheduler (4 files)
ls kernel/scheduler/queue.go
ls kernel/scheduler/deps.go
ls kernel/scheduler/scheduler.go
ls kernel/scheduler/scheduler_test.go

# Resilience (2 files)
ls kernel/resilience/retry.go
ls kernel/resilience/circuitbreaker.go

# Metrics (1 file)
ls kernel/metrics/metrics.go

# Kernel (4 files)
ls kernel/state.go
ls kernel/kernel.go
ls kernel/lifecycle/lifecycle.go
ls kernel/kernel_test.go
```

### Step 2: Ensure go.mod has yaml dependency
```bash
go get gopkg.in/yaml.v3
go mod tidy
```

### Step 3: Build all kernel code
```bash
go build ./kernel/...
# Expected: no output, no errors
# If error: check import paths, package names, missing types
```

### Step 4: Go vet
```bash
go vet ./kernel/...
# Expected: no output, no warnings
```

### Step 5: Run all kernel tests
```bash
go test -v ./kernel/...
# Expected: ALL PASS
# Minimum test functions: ~65 across all packages
```

### Step 6: Run with race detector
```bash
go test -race ./kernel/...
# Expected: ALL PASS, no race conditions detected
# Race detector adds ~10x slowdown but catches data races
```

### Step 7: Coverage report
```bash
go test -coverprofile=coverage_kernel.out ./kernel/...
go tool cover -func=coverage_kernel.out
# Expected: ≥ 75% coverage overall
# Individual packages:
#   kernel/config    ≥ 80%
#   kernel/logger    ≥ 70%
#   kernel/eventbus  ≥ 85%
#   kernel/registry  ≥ 80%
#   kernel/runtime   ≥ 70%
#   kernel/scheduler ≥ 80%
#   kernel           ≥ 70%
```

### Step 8: Build entire project (including Phase 1)
```bash
go build ./...
# Expected: no output, no errors
# Verifies: no import cycles between contracts and kernel
```

### Step 9: Run entire project tests
```bash
go test ./...
# Expected: ALL PASS (contracts + kernel)
```

### Step 10: Import cycle check
```bash
go build ./kernel/...
# If import cycle exists, this will fail with:
# "import cycle not allowed"
#
# Valid import graph (NO cycles):
#   contracts/* ← kernel/config (yaml tags)
#   contracts/* ← kernel/logger (none)
#   contracts/event ← kernel/eventbus
#   contracts/* ← kernel/registry
#   kernel/registry, kernel/eventbus ← kernel/runtime
#   kernel/runtime ← kernel/scheduler (via DispatchFunc, NOT direct import)
#   kernel/* ← kernel (kernel.go)
#   kernel ← kernel/lifecycle
#
# FORBIDDEN imports:
#   kernel/* ← contracts/* (contracts must not know about kernel)
#   kernel/* ← plugins/* (kernel must not know about plugins)
#   kernel/scheduler ← kernel/runtime (use DispatchFunc instead)
```

### Step 11: Git commit
```bash
git add -A
git commit -m "Phase 2: Kernel core implementation (40 micro-tasks)

Components:
- Config: YAML loading, env var resolution, validation, watcher (hot-reload)
- Logger: slog-based structured logging with redaction
- EventBus: async pub/sub with wildcard matching and DLQ (dead letter queue)
- Registry: thread-safe plugin management with lifecycle
- Runtime: task execution with worker pool, fallback execution, and leak checks
- Scheduler: priority queue with dependency tracking
- Resilience: retry policy and circuit breaker
- Metrics: in-memory observability collector

All tests pass with race detector. Coverage >= 75%."

git push origin main
```

## File Count Summary

| Package | Source Files | Test Files | Total |
|---|---|---|---|
| kernel/config | 6 | 1 | 7 |
| kernel/logger | 4 | 1 | 5 |
| kernel/eventbus | 6 | 1 | 7 |
| kernel/registry | 3 | 1 | 4 |
| kernel/runtime | 4 | 1 | 5 |
| kernel/scheduler | 3 | 1 | 4 |
| kernel/resilience | 2 | 0 | 2 |
| kernel/metrics | 1 | 0 | 1 |
| kernel | 2 | 1 | 3 |
| kernel/lifecycle | 1 | 0 | 1 |
| **Total** | **32** | **7** | **39** |

## Quality Gates (ALL must pass)

- [ ] All 39 files exist
- [ ] `go build ./kernel/...` ✅
- [ ] `go vet ./kernel/...` ✅
- [ ] `go test ./kernel/...` ALL PASS
- [ ] `go test -race ./kernel/...` NO RACES
- [ ] Coverage ≥ 75%
- [ ] No import cycles
- [ ] `go build ./...` ✅ (entire project)
- [ ] `go test ./...` ALL PASS (entire project)
- [ ] Git commit + push successful

## Architecture Validation

After Phase 2 completion, verify the dependency graph:

```
contracts/        ← kernel/           ← (Phase 3: SDK)
(interfaces)        (implementation)     (developer helpers)
                                       ← (Phase 4: plugins)
                                          (agent/provider implementations)
```

The kernel ONLY imports from contracts. It does NOT import from plugins or SDK.
This is verified by the successful `go build` — if any forbidden import exists, the build fails.
