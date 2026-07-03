# RFC-0039: Evolution Engine

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0002 (Brain Architecture), RFC-0010 (Cognitive Layer)

## Summary

This RFC specifies the design of the **Evolution Engine** in AEOS. The Evolution Engine is responsible for updating the Planner's rules, scoring coefficients, and template success weights over time based on historical execution metrics (Planner self-learning), rather than just updating agent weights.

## Motivation

Even if individual agents learn to write code better, the Planner itself can remain sub-optimal if it continues choosing outdated plan templates.
- We need a feedback loop that evaluates the quality of the plans generated.
- By tuning planner scoring weights ($w_i$) based on historical replanning frequency and compile rates, the Planner evolves to choose safer, faster paths.

## Design

### 1. Architectural Placement

The Evolution Engine executes in the background as part of the **Observation Runtime**, subscribing to mission completion events.

```
  Mission Completed ──► [Evolution Engine] ──► Update Template Weights in Knowledge Graph
```

---

### 2. Contracts (`contracts/brain/evolution.go`)

```go
package brain

import (
	"context"
	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

// EvolutionReport outlines planner weight modifications.
type EvolutionReport struct {
	TemplateID string             `json:"template_id"`
	OldWeight  float64            `json:"old_weight"`
	NewWeight  float64            `json:"new_weight"`
	Metrics    map[string]float64 `json:"metrics"`
}

// EvolutionEngine updates planner success parameters.
type EvolutionEngine interface {
	// Evolve computes and applies weight adjustments based on execution history.
	Evolve(ctx context.Context, record fsm.TransitionRecord) (*EvolutionReport, error)
}
```

## Impact

- **Self-Improving Planner**: Plan templates that frequently result in compilation failure are automatically demoted.
- **Auto-Deprecation**: Obsolete technologies (e.g. Go 1.18 templates) naturally decline in rating, forcing the Planner to choose modern stack variants.

## Open Questions

1. **How do we prevent weight starvation?**
   - The UCB-1 algorithm (RFC-0039) adds an exploration bonus to low-usage nodes, ensuring newly updated templates have a chance to be evaluated.
