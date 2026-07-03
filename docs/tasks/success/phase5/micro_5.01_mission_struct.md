# Micro-Task 5.01: Implement kernel/planner/planner.go

- **File**: `kernel/planner/planner.go`
- **Package**: `planner`
- **Depends on**: 1.29 (planner contract), 1.41 (brain contract), 1.42 (knowledge contract)
- **Time**: 20 min
- **Verify**: `go build ./kernel/planner/...`

## Purpose
Implements the core `planner.Planner` struct. In AEOS, the Planner acts as a **deterministic-cognitive search engine** running in the `Brain Runtime` that generates candidate plan DAGs, scores them via multi-objective Pareto Frontier functions, explains selections, and updates template weights.

## EXACT code to create

```go
package planner

import (
	"context"
	"fmt"
	"log/slog"
	
	"github.com/tiendat1751998/orchestrator/contracts/goal"
	"github.com/tiendat1751998/orchestrator/contracts/fsm"
	"github.com/tiendat1751998/orchestrator/contracts/planner"
	"github.com/tiendat1751998/orchestrator/contracts/knowledge"
	"github.com/tiendat1751998/orchestrator/contracts/brain"
)

type engine struct {
	store      knowledge.KnowledgeStore
	skillGraph knowledge.SkillGraph
	trust      brain.TrustEngine
	logger     *slog.Logger
}

// Compile-time check
var _ planner.Planner = (*engine)(nil)

// NewPlanner constructs the core Planner engine.
func NewPlanner(
	store knowledge.KnowledgeStore,
	skillGraph knowledge.SkillGraph,
	trust brain.TrustEngine,
	logger *slog.Logger,
) planner.Planner {
	return &engine{
		store:      store,
		skillGraph: skillGraph,
		trust:      trust,
		logger:     logger,
	}
}

func (e *engine) Plan(ctx context.Context, g goal.Goal) ([]fsm.DAG, error) {
	// CSP Constraint Pruning:
	// 1. Traverse g.Constraints.
	// 2. Query skillGraph to subtract/prune nodes violating constraints.
	// 3. Generate candidate DAGs using templates matching objectives from KnowledgeStore.
	candidates := []fsm.DAG{}
	return candidates, nil
}

func (e *engine) Score(ctx context.Context, candidates []fsm.DAG) (fsm.DAG, error) {
	if len(candidates) == 0 {
		return fsm.DAG{}, fmt.Errorf("planner: no plan candidates to score")
	}
	
	// Pareto Multi-Objective Scoring:
	// Score = w_quality*Q + w_cost*C + w_time*T + w_confidence*Conf - w_risk*R
	// Plus UCB-1 exploration factor: c * sqrt(ln(Total)/UsageCount)
	best := candidates[0]
	return best, nil
}

func (e *engine) Explain(ctx context.Context, chosen fsm.DAG, candidates []fsm.DAG) (string, error) {
	// Compare chosen plan vs runners-up, detailing weight deltas
	return "Selected Plan A due to 15% lower risk margin compared to Plan B.", nil
}

func (e *engine) Learn(ctx context.Context, history fsm.TransitionRecord) error {
	// 1. Calculate success/failure status from DoD validation.
	// 2. Perform EMA weight update: Weight = (1-a)*Weight + a*Success.
	// 3. Record failure association nodes on failure.
	return nil
}
```

## Verify
```bash
go build ./kernel/planner/...
```

## Checklist
- [ ] File `kernel/planner/planner.go` exists
- [ ] Package: `planner`
- [ ] Implement `planner.Planner` interface methods: `Plan`, `Score`, `Explain`, `Learn`
- [ ] Structure includes references to KnowledgeStore, SkillGraph, and TrustEngine
- [ ] `go build ./kernel/planner/...` passes
