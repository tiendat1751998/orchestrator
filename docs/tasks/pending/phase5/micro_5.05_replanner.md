# Micro-Task 5.05: Create kernel/planner/replanner.go

- **File**: `kernel/planner/replanner.go`
- **Package**: `planner`
- **Depends on**: 5.01, 1.41 (brain/brain.go)
- **Time**: 20 min
- **Verify**: `go build ./kernel/planner/...`

## Purpose
Implements the `brain.Replanner` contract. The Replanner intercepts execution errors, analyzes logs, and mutates the active plan DAG to insert corrective tasks (e.g. running Ast fixes or dependency downloads) before resuming scheduling.

## EXACT code to create

```go
package planner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tiendat1751998/orchestrator/contracts/fsm"
	"github.com/tiendat1751998/orchestrator/contracts/brain"
)

type replanner struct {
	logger *slog.Logger
}

// Compile-time check
var _ brain.Replanner = (*replanner)(nil)

// NewReplanner constructs a replanning handler.
func NewReplanner(logger *slog.Logger) brain.Replanner {
	return &replanner{logger: logger}
}

func (r *replanner) Replan(ctx context.Context, fCtx brain.FailureContext, activeDAG fsm.DAG) (fsm.DAG, error) {
	if r.logger != nil {
		r.logger.Info("intercepting task failure, generating corrective sub-graph", "task_id", fCtx.TaskID)
	}

	// 1. Analyze failure category (compilation, test_failure)
	// 2. Generate corrective task nodes (e.g., adding "run_linter_fix" node)
	// 3. Mutate activeDAG and return the updated graph
	mutated := activeDAG
	return mutated, nil
}
```

## Verify
```bash
go build ./kernel/planner/...
```

## Checklist
- [ ] File `kernel/planner/replanner.go` exists
- [ ] Package: `planner`
- [ ] `replanner` struct implements `brain.Replanner` interface
- [ ] `Replan` mutates and returns the execution graph
- [ ] `go build ./kernel/planner/...` passes
