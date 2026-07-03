# RFC-0049: Benchmark Framework

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0034 (Advanced Quality Engine), RFC-0013 (Workspace Engine)

## Summary

This RFC specifies the design of the **Benchmark Framework** in AEOS. It is responsible for running offline evaluations of Planner versions (A vs. B) across standard task profiles, compiling comparative metrics (latency, token costs, compile rates, test pass rates), and outputting standard Benchmark Manifests.

## Motivation

We cannot rely on subjective design reviews to validate architectural improvements.
- We must prove empirically that a new Planner algorithm outperforms the baseline.
- The Benchmark Framework runs repeatable conformance test suites in a local environment, generating standard YAML manifests.

## Design

### 1. Architectural Placement

The Benchmark Framework is an offline runner package located in `sdk/testing/` that runs separate from the active execution kernel.

```
  Benchmark Corpus ──► [Benchmark Framework] ──► Compare Planners ──► Benchmark Manifest
```

---

### 2. Contracts (`sdk/testing/benchmark.go`)

```go
package testing

import (
	"context"
	"time"
)

// BenchmarkMetrics represents performance results.
type BenchmarkMetrics struct {
	CompileSuccessRate float64       `json:"compile_success_rate"`
	TestPassRate       float64       `json:"test_pass_rate"`
	AverageDuration    time.Duration `json:"average_duration"`
	TotalUSDConsumed   float64       `json:"total_usd_consumed"`
}

// PlannerEvaluator compares planner models.
type PlannerEvaluator interface {
	// EvaluateRuns executes a suite of 100 baseline goals against the target planner.
	EvaluateRuns(ctx context.Context, plannerID string) (*BenchmarkMetrics, error)
}
```

## Impact

- **Empirical Validation**: Axiom 16 of the constitution requires that any architectural change must improve a specific fitness score compiled by this framework.
- **Overfitting Prevention**: Benchmark suites are divided into Train, Validation, and Hidden test sets to prevent planner models from "learning" the tests.

## Open Questions

1. **How do we isolate benchmark execution?**
   - The framework spawns isolated docker containers or local sandbox processes to execute test compilation, preventing system contamination.
