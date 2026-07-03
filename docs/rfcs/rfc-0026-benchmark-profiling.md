# RFC-0026: Benchmark & Runtime Profiling

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0016 (Observation Runtime), RFC-0034 (Advanced Quality Engine)

## Summary

This RFC specifies the design of the **Benchmark & Runtime Profiling** service in AEOS. Located inside the Observation Runtime, it runs CPU, memory, and latency profiling on generated code and benchmarks, feeding results into the quality scorecard to optimize code performance.

## Motivation

AI-generated code often compiles and passes tests, but suffers from performance issues like memory leaks or high latency.
- We must detect these performance regressions automatically.
- Profiling runs help identify CPU-intensive loops or leak indicators, grading the code's real-world efficiency.

## Design

### 1. Architectural Placement

The Profiling service is an analysis adapter inside the `Observation Runtime`, running after code validation.

```
  Execution ──► [Profiling Service] ──► Measure CPU/Memory ──► Quality Scorecard
```

---

### 2. Contracts (`contracts/quality/profiler.go`)

```go
package quality

import (
	"context"
	"time"
)

// ProfileMetrics represents performance results.
type ProfileMetrics struct {
	MemoryAllocationBytes int64         `json:"memory_allocation_bytes"`
	CPUExecutionTime      time.Duration `json:"cpu_execution_time"`
	ThroughputPerSecond   float64       `json:"throughput_per_second"`
}

// RuntimeProfiler profiles generated code.
type RuntimeProfiler interface {
	// Profile runs performance benchmarks on the compiled workspace binary.
	Profile(ctx context.Context, binaryPath string) (*ProfileMetrics, error)
}
```

## Impact

- **Performance Regression Guard**: The final DoD gate rejects code changes that increase latency or memory usage beyond defined thresholds.
- **Accurate Resource Mapping**: The Evolution Engine maps resource costs against stack patterns.
