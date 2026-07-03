# RFC-0044: Economic Engine

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0018 (Cost Engine), RFC-0038 (Resource Planning)

## Summary

This RFC specifies the design of the **Economic Engine** in AEOS. The Economic Engine is responsible for optimizing plan selection based on Expected Business Value (Expected ROI, latency savings, estimated developer hours saved, and bug risk coefficients) rather than raw API token costs ($).

## Motivation

Minimizing API token costs can lead to poor engineering choices (e.g. choosing a cheap, small model that generates insecure, buggy code).
- A professional CTO optimizes for **Business Value**. If spending an extra $2 in premium tokens saves 10 hours of future human debugging, it represents a massive net positive ROI.
- By modeling business value and technical risk coefficients mathematically, the Planner selects plans that maximize overall engineering return.

## Design

### 1. Architectural Placement

The Economic Engine acts as a scoring adapter in the Planner's scoring function, modifying the weights assigned to plan candidates.

```
  Candidate Plans ──► [Economic Engine] ──► Expected Utility Score ──► Choose Winner
```

---

### 2. Contracts (`contracts/brain/economics.go`)

```go
package brain

import "context"

// ProjectROI represents the estimated value return.
type ProjectROI struct {
	DeveloperHoursSaved float64 `json:"developer_hours_saved"`
	LatencyImprovement  float64 `json:"latency_improvement_pct"`
	BugRiskReduction    float64 `json:"bug_risk_reduction"`
}

// EconomicEngine calculates plan business utility.
type EconomicEngine interface {
	// CalculateROI estimates the return metrics of a plan.
	CalculateROI(ctx context.Context, steps []string, techStack []string) (*ProjectROI, error)
	
	// ExpectedUtility computes the final mathematical utility score of a plan.
	ExpectedUtility(ctx context.Context, roi ProjectROI, projectedCost float64) float64
}
```

## Impact

- **Value-Driven Planning**: The Planner naturally favors robust, clean templates (Gin + sqlc) over fragile, hasty configurations, even if they consume slightly more API tokens.
- **CTO Simplicity Alignment**: Overly complex architectures are discounted because they represent negative expected business value (higher future human debugging risk).

## Open Questions

1. **How do we measure "developer hours saved"?**
   - The system uses industry-standard baseline estimates for common tasks (e.g., generating database migrations manually vs. using sqlc) and adjusts them over time.
