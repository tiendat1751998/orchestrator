# Micro-Task 5.27: Verification — Build & Test All Phase 5

## Info
- **File**: None (verification task definition only)
- **Depends on**: ALL micro-tasks 5.01 → 5.26
- **Time**: 15 min
- **Verify**: `go test -race ./kernel/...`

## Purpose
Verifies that all kernel planner packages, orchestrator handlers, resilience middleware, and security checkers compile cleanly, and pass all verification tests with the race detector.

## Verification Commands (Execute in exact order)

### Step 1: Verify Kernel Files Exist
```bash
# Planner Engine
ls kernel/planner/mission.go
ls kernel/planner/decomposer.go
ls kernel/planner/dag.go
ls kernel/planner/strategy.go
ls kernel/planner/replanner.go
ls kernel/planner/optimizer.go
ls kernel/planner/planner_test.go

# Orchestrator Pipeline
ls kernel/orchestrator/orchestrator.go
ls kernel/orchestrator/coordinator.go
ls kernel/orchestrator/pipeline.go
ls kernel/orchestrator/supervisor.go
ls kernel/orchestrator/aggregator.go
ls kernel/orchestrator/feedback.go
ls kernel/orchestrator/orchestrator_test.go

# Resilience Middleware
ls kernel/resilience/circuit_breaker.go
ls kernel/resilience/retry.go
ls kernel/resilience/fallback.go
ls kernel/resilience/timeout.go
ls kernel/resilience/health.go
ls kernel/resilience/recovery.go
ls kernel/resilience/resilience_test.go

# Security Checks
ls kernel/security/permission.go
ls kernel/security/sandbox.go
ls kernel/security/audit.go
ls kernel/security/secrets.go
ls kernel/security/security_test.go
```

### Step 2: Go Build (Compiler Check)
```bash
go build ./kernel/...
```

### Step 3: Go Vet (Linter Check)
```bash
go vet ./kernel/...
```

### Step 4: Go Test (Unit Tests)
```bash
go test -v ./kernel/...
```

### Step 5: Go Test with Race Detector
```bash
go test -race ./kernel/...
```

### Step 6: Full Workspace Check
```bash
go build ./...
```

### Step 7: Git Commit
```bash
git add -A
git commit -m "Phase 5: Orchestration engine implementation (27 micro-tasks)"
git push origin main
```

## Phase 5 Quality Checklist

### Planner Engine
- [ ] `kernel/planner/mission.go` — Mission struct and validations.
- [ ] `kernel/planner/decomposer.go` — Decomposes mission using AI prompt structures.
- [ ] `kernel/planner/dag.go` — Graph checks detect cycle loops using Kahn's topological sort.
- [ ] `kernel/planner/replanner.go` — Recovery planner generates sub-tasks with max attempt limits.

### Orchestrator Pipeline
- [ ] `kernel/orchestrator/orchestrator.go` — Executes mission loop in topological order, with recovery routes.
- [ ] `kernel/orchestrator/coordinator.go` — Dependency injection maps parameters without deleting key properties.
- [ ] `kernel/orchestrator/pipeline.go` — Validates state machine pathways.
- [ ] `kernel/orchestrator/supervisor.go` — Scanner checks timeouts in background goroutines.

### Resilience Middleware
- [ ] `kernel/resilience/circuit_breaker.go` — Closed, Open, and HalfOpen states are protected under mutexes.
- [ ] `kernel/resilience/retry.go` — Jitters transient errors and checks context cancel triggers.
- [ ] `kernel/resilience/timeout.go` — Caps child timeout allocations to remaining parent durations.
- [ ] `kernel/resilience/recovery.go` — Writes checkpoint files atomically.

### Security Checks
- [ ] `kernel/security/permission.go` — Default Deny is applied, checking absolute target paths.
- [ ] `kernel/security/audit.go` — JSON lines logs write events to disk under lock guards.
- [ ] `kernel/security/secrets.go` — Secrets redact matching strings over 4 characters.

### Quality Gates
- [ ] `go build ./...` ✅ (clean compilation of entire workspace)
- [ ] `go test ./kernel/...` ALL PASS
- [ ] `go test -race ./kernel/...` NO RACES
- [ ] No circular package imports are found
- [ ] Git commit and push succeeds
