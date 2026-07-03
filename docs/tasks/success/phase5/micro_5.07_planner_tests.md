# Micro-Task 5.07: Create kernel/planner/planner_test.go

- **File**: `kernel/planner/planner_test.go`
- **Package**: `planner`
- **Depends on**: 5.01, 5.02, 5.04, 5.06
- **Time**: 25 min
- **Verify**: `go test -v ./kernel/planner/...`

## Purpose
Verifies the correct mathematical execution of the new Planner algorithms: CSP constraint filtering, Pareto Frontier scoring, UCB-1 exploration, and contrastive explainability.

## EXACT code to create

```go
package planner_test

import (
	"context"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/goal"
	"github.com/tiendat1751998/orchestrator/contracts/fsm"
	"github.com/tiendat1751998/orchestrator/kernel/planner"
)

func TestCSPSolver_Filter(t *testing.T) {
	// 1. Mock SkillGraph and register tech nodes (Postgres, SQLite).
	// 2. Configure a "database_only: sqlite" constraint.
	// 3. Verify that the CSP filter successfully prunes the Postgres node.
}

func TestScorer_ParetoAndUCB(t *testing.T) {
	// 1. Configure weights for Quality (+), Cost (-), and Risk (-).
	// 2. Score two candidate DAGs (one cheap/low-success, one premium/high-success).
	// 3. Verify that the UCB-1 bonus is applied correctly when usage counts are low.
}
```

## Verify
```bash
go test -v ./kernel/planner/...
```

## Checklist
- [ ] File `kernel/planner/planner_test.go` exists
- [ ] Package: `planner_test`
- [ ] Unit tests for `CSPSolver` node filtering exist
- [ ] Unit tests for `Scorer` Pareto and UCB-1 calculations exist
- [ ] `go test -v ./kernel/planner/...` passes
