# Micro-Task 1.29: Create contracts/planner/planner.go

- **File**: `contracts/planner/planner.go`
- **Package**: `planner`
- **Depends on**: 1.28 (context/context.go), 1.23 (event/event.go)
- **Time**: 15 min
- **Verify**: `go build ./contracts/planner/...`

## Purpose
Declares the frozen, locked `Planner` contract. In accordance with the system constitution (docs/adp.md) and standard specifications (docs/specification.md), the Planner is a **deterministic-cognitive search engine** that compiles, grades, and explains candidate plan DAGs. It does not call AI directly; it uses structural constraints, Pareto scoring, and EMA failure logs.

## EXACT code to create

```go
// Package planner defines the locked contract for plan generation, scoring, and learning.
package planner

import (
	"context"
	
	"github.com/tiendat1751998/orchestrator/contracts/goal"
	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

// Planner defines the locked contract for plan generation and optimization.
type Planner interface {
	// Plan generates candidate plan DAGs satisfying the goals.
	// It applies CSP constraint propagation to prune the search space.
	Plan(ctx context.Context, g goal.Goal) ([]fsm.DAG, error)
	
	// Score evaluates candidate plans mathematically using multi-objective Pareto Frontier calculations.
	Score(ctx context.Context, candidates []fsm.DAG) (fsm.DAG, error)
	
	// Explain generates the contrastive mathematical reasoning report detailing why the plan was chosen.
	Explain(ctx context.Context, chosen fsm.DAG, candidates []fsm.DAG) (string, error)
	
	// Learn updates template weights and failure association edges based on transition results.
	Learn(ctx context.Context, history fsm.TransitionRecord) error
}
```

## Rules
1. **Interface Stability**: This interface must remain 100% stable to ensure plugin and engine backward compatibility.
2. **Context Propagation**: Every I/O-bound or cognitive operation must propagate `context.Context` as its first parameter.
3. **No AI imports**: The package must remain clean of external provider dependencies.

## Verify
```bash
go build ./contracts/planner/...
```

## Checklist
- [ ] File `contracts/planner/planner.go` exists
- [ ] Package: `planner`
- [ ] Imports `contracts/goal` and `contracts/fsm` cleanly
- [ ] `Planner` interface declares `Plan`, `Score`, `Explain`, and `Learn` methods
- [ ] `go build ./contracts/planner/...` passes
