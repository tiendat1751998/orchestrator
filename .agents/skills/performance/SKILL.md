---
name: Performance Engineer
description: Instructions for optimizing task scheduling latency, EventBus throughput, goroutine pool efficiency, and provider response times.
---

# Performance Engineer Playbook

## Session Startup (MANDATORY)
1. Read `.agents/context/performance-budgets.md` — targets and benchmarks.
2. Read `.agents/context/architecture.md` — understand kernel components to optimize.

## Workflow

### 1. Baseline Measurement
- Run existing benchmarks: `go test -bench=. -benchmem ./kernel/...`
- Profile CPU and heap: `go tool pprof`
- Measure scheduling latency, EventBus throughput, registry lookup time.

### 2. Identify Bottlenecks
- Use `pprof` to find CPU and memory hotspots in kernel packages.
- Check for lock contention in `kernel/registry`, `kernel/scheduler`.
- Look for goroutine leaks in `kernel/runtime`.
- Review GC pressure from EventBus event allocation.

### 3. Optimize
- Write targeted optimizations with minimal changes.
- Add benchmarks for specific code paths.
- Run comparative benchmarks (before vs after).
- Verify no race conditions: `go test -race ./...`

### 4. Benchmark Results
```bash
go test -bench=. -benchmem ./kernel/eventbus/...
go test -bench=. -benchmem ./kernel/registry/...
go test -bench=. -benchmem ./kernel/scheduler/...
go test -bench=. -benchmem ./kernel/runtime/...
```

### 5. Report
- Document findings with actual metrics.
- Include before/after comparison tables.
- Update `.agents/context/performance-budgets.md` if targets change.

## Key Performance Targets
- EventBus publish: >100K events/sec
- Registry lookup: <1μs
- Scheduler cycle: <5μs
- Task dispatch: <100μs
- Binary startup: <500ms

## 🚫 ANTI-FAKE RULES
- Every "p99=Xms" claim → MUST run the benchmark and paste output.
- Every "no regressions" claim → MUST show before/after metrics.
- "Code looks optimized" IS NOT proof — run benchmarks.