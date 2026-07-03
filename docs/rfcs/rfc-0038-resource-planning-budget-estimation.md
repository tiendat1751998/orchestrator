# RFC-0038: Resource Planning & Budget Estimation

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0009 (Resource Manager), RFC-0030 (Goal Engine)

## Summary

This RFC specifies the design of the **Resource Planning & Budget Estimation** module in AEOS. The module estimates token, dollar, RAM, GPU, and time requirements upfront during planning. If the plan exceeds constraints, the Planner splits the mission or adjusts allocations to prevent budget overruns.

## Motivation

AI operations can become very expensive if not constrained.
- Complex goals can generate huge plans that cost hundreds of dollars in LLM API calls.
- By calculating expected utility and budget costs *before* execution, the system can reject expensive strategies and choose cost-efficient alternatives.

## Design

### 1. Architectural Placement

Resource Planning is an integral step in the Planner's scoring function, reading constraints directly from the Goal struct.

```
  Constraint Bounds ──► [Resource Planner] ──► Validate Candidate cost ──► Filter
```

---

### 2. Contracts (`contracts/resource/planner.go`)

```go
package resource

import "context"

// BudgetLimits represents the mission constraints.
type BudgetLimits struct {
	MaxUSD      float64 `json:"max_usd"`
	MaxDuration int     `json:"max_duration_seconds"`
	MaxMemoryMB int64   `json:"max_memory_mb"`
}

// ProjectedUsage represents the estimated consumption.
type ProjectedUsage struct {
	EstimatedUSD      float64 `json:"estimated_usd"`
	EstimatedDuration int     `json:"estimated_duration_seconds"`
	RequiresGPU       bool    `json:"requires_gpu"`
}

// ResourcePlanner estimates resources for candidates.
type ResourcePlanner interface {
	// Estimate calculates projected usage metrics for a candidate plan.
	Estimate(ctx context.Context, steps []string) (*ProjectedUsage, error)
	
	// Validate verifies if usage is within budget bounds.
	Validate(ctx context.Context, usage ProjectedUsage, limits BudgetLimits) (bool, error)
}
```

## Impact

- **Budget Guards**: Missions are blocked immediately if the projected cost exceeds bounds, protecting users from unexpected API bills.
- **Cost-Optimized Planning**: The Planner chooses cheap local models (Antigravity CLI) for boilerplate tasks, reserving premium models for complex reasoning.

## Open Questions

1. **How accurate are LLM token estimations?**
   - The system tracks historical average token usage per task type and uses a 20% safety margin to prevent underestimation.
