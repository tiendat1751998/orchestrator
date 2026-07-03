# Performance Budgets & Benchmarks

## 1. Core Kernel Latencies

| Component | Target (p50) | Limit (p99) |
|-----------|--------------|-------------|
| EventBus publish throughput | > 100,000 events/sec | — |
| EventBus publish latency | < 0.1ms | < 1ms |
| Registry capability lookup | < 1μs | < 5μs |
| Scheduler queue push/pop cycle | < 5μs | < 20μs |
| Task dispatcher routing | < 100μs | < 500μs |
| CLI startup bootstrap | < 200ms | < 500ms |

---

## 2. Benchmark Requirements

All optimizations must be validated using Go benchmarks. The following modules must have benchmark suites:

### EventBus (`kernel/eventbus/bus_test.go`)
Ensures subscription matching and dispatch scales with high subscriber counts.
```bash
go test -bench=BenchmarkEventBus -benchmem ./kernel/eventbus/...
```

### Registry (`kernel/registry/registry_test.go`)
Ensures capability resolution is O(1) or O(log N) rather than O(N).
```bash
go test -bench=BenchmarkRegistry -benchmem ./kernel/registry/...
```

### Scheduler (`kernel/scheduler/scheduler_test.go`)
Ensures topological sorting and priority queueing does not degrade.
```bash
go test -bench=BenchmarkScheduler -benchmem ./kernel/scheduler/...
```
