# RFC-0034: Advanced Quality Engine

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0014 (Quality Engine), RFC-0015 (Definition of Done Engine)

## Summary

This RFC specifies the design of the **Advanced Quality Engine** in AEOS. Moving beyond a binary pass/fail validation check, the Advanced Quality Engine compiles a multi-dimensional, real-numbered **Quality Scorecard** based on compile logs, linter warnings, unit test coverage, security scans, benchmark performance, and API contract matching.

## Motivation

Binary pass/fail checks do not provide enough granularity for optimization.
- A Planner cannot perform gradient-based learning or template scoring if outcomes are simply correct/incorrect.
- A multi-dimensional score (e.g. `Compile: 1.0, Linter: 0.8, Coverage: 0.9`) allows the system to measure precise engineering regressions and select templates that maximize code health.

## Design

### 1. Architectural Placement

The Advanced Quality Engine acts as the primary evaluation adapter inside the **Observation Runtime**, executing after test/build tasks complete.

```
  Build Outputs ──► [Advanced Quality Engine] ──► QualityScorecard (float64 metrics)
```

---

### 2. Contracts (`contracts/quality/`)

```go
package quality

import "context"

// QualityMetric represents a single dimensional score.
type QualityMetric struct {
	Name        string  `json:"name"`  // "compile", "lint", "coverage", "security"
	Score       float64 `json:"score"` // [0.0, 1.0]
	Description string  `json:"description,omitempty"`
}

// QualityScorecard represents the multi-dimensional report.
type QualityScorecard struct {
	MissionID   string          `json:"mission_id"`
	Timestamp   int64           `json:"timestamp"`
	Metrics     []QualityMetric `json:"metrics"`
	TotalScore  float64         `json:"total_score"` // Weighted average
}

// QualityEngine compiles the scorecard.
type QualityEngine interface {
	// Evaluate analyzes test reports and logs to compile a scorecard.
	Evaluate(ctx context.Context, logsPath string) (*QualityScorecard, error)
}
```

## Impact

- **Gradient Learning**: The Planner's Pareto Scoring function (RFC-0002) uses the scorecard's `TotalScore` to adjust historical template weights.
- **Architectural Fitness Gates**: The system constitution (docs/adp.md) requires that any new feature must improve a specific metric on the scorecard.

## Open Questions

1. **How are metric weights configured?**
   - Weights are stored in `config.yaml` and tuned using the Benchmark Framework (RFC-0049) to optimize the overall success rates.
