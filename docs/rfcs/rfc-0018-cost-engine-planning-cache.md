# RFC-0018: Cost Engine & Planning Cache

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0002 (Brain Architecture), RFC-0009 (Resource Manager)

## Summary

This RFC specifies the design of the **Cost Engine & Planning Cache** in AEOS. The module tracks real-time token/dollar consumption during execution and implements local cache endpoints for recurring query patterns, reducing LLM calls and execution latency to zero on cached paths.

## Motivation

AI model pricing is a critical execution parameter.
- Repeatedly querying the same plans (e.g. running identical linter fixes) wastes API tokens and introduces latency.
- By caching successful execution sub-graphs locally, the system skips LLM generation for known tasks.

## Design

### 1. Architectural Placement

The Cost Engine wraps LLM provider calls, while the Planning Cache resides in the `Brain Runtime`.

```
  Planner query ──► [Planning Cache] ──(Hit?)──► Return cached Plan DAG
                         │
                      (Miss)
                         ▼
                [LLM Call (Cost Engine)]
```

---

### 2. Contracts (`contracts/brain/cost.go`)

```go
package brain

import "context"

// CostReport represents token and pricing metrics.
type CostReport struct {
	MissionID   string  `json:"mission_id"`
	TokensInput int     `json:"tokens_input"`
	TokensOutput int    `json:"tokens_output"`
	TotalUSD    float64 `json:"total_usd"`
}

// CostEngine tracks and limits API spend.
type CostEngine interface {
	// LogUsage records token consumption.
	LogUsage(ctx context.Context, model string, input int, output int) error
	
	// CurrentSpend returns the compiled budget spent.
	CurrentSpend(ctx context.Context, missionID string) (float64, error)
}
```

## Impact

- **Zero-Latency Hits**: Repeating identical refactoring or compile-fix runs returns plan templates instantly from cache, bypassing LLM delays.
- **Budget Control**: The cost logs feed directly into the Planner's scoring function, ensuring economical paths are selected.
