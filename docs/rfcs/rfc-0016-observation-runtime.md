# RFC-0016: Observation Runtime & Telemetry Collector

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0001 (Kernel Architecture), RFC-0008 (Event Model)

## Summary

This RFC specifies the design of the **Observation Runtime & Telemetry Collector** in AEOS. The Observation Runtime runs as an independent, background process to collect AST modifications, Git changes, compilation output metrics, and system resources, isolating learning loops from the execution runtime to prevent bottlenecks.

## Motivation

Gathering metrics, indexing files, and analyzing code changes is computationally expensive.
- If these checks run synchronously inside the active execution loop, they introduce significant latency.
- By isolating these tasks in a background Observation Runtime, we gather rich telemetry and compile knowledge without affecting task scheduling speed.

## Design

### 1. Architectural Placement

The Observation Runtime is one of the 4 core runtimes of AEOS, running asynchronously in the background.

```
  [Execution Runtime] ──(Emits Events)──► [EventBus] ──► [Observation Runtime (Telemetry)]
                                                                    │
                                                                    ▼
                                                            Knowledge Database
```

---

### 2. Contracts (`contracts/brain/observation.go`)

```go
package brain

import (
	"context"
)

// MetricSnapshot represents collected telemetry.
type MetricSnapshot struct {
	Timestamp int64   `json:"timestamp"`
	CPUUsage  float64 `json:"cpu_usage"`
	MemoryMB  int64   `json:"memory_mb"`
}

// ObservationCollector gathers background telemetry.
type ObservationCollector interface {
	// Start begins background telemetry collection.
	Start(ctx context.Context) error
	
	// Stop shuts down the collector.
	Stop(ctx context.Context) error
}
```

## Impact

- **Decoupled Learning**: The system records AST changes and build details in the background, allowing the active Planner to remain fast.
- **Rich Telemetry**: Captures execution metrics (latency, resource utilization) linked to the `MissionID` aggregate.
