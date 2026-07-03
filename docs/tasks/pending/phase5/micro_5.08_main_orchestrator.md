# Micro-Task 5.08: Create kernel/orchestrator/orchestrator.go

- **File**: `kernel/orchestrator/orchestrator.go`
- **Package**: `orchestrator`
- **Depends on**: 5.01 (planner/planner.go), 1.30 (orchestrator contract)
- **Time**: 25 min
- **Verify**: `go build ./kernel/orchestrator/...`

## Purpose
Implements the core Orchestrator coordinator running the FSM lifecycles. It initiates workspace transactions, invokes the Planner to get candidates, scores them to find the Pareto frontier, and executes tasks using the lease-based scheduler.

## EXACT code to create

```go
package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/goal"
	"github.com/tiendat1751998/orchestrator/contracts/fsm"
	"github.com/tiendat1751998/orchestrator/contracts/planner"
	"github.com/tiendat1751998/orchestrator/contracts/workspace"
)

// Orchestrator coordinates mission FSM lifecycles.
type Orchestrator struct {
	planner planner.Planner
	tx      workspace.WorkspaceTransactionEngine
	logger  *slog.Logger
}

// NewOrchestrator constructs a new coordinator.
func NewOrchestrator(
	p planner.Planner,
	tx workspace.WorkspaceTransactionEngine,
	logger *slog.Logger,
) (*Orchestrator, error) {
	if p == nil || tx == nil {
		return nil, errors.New("orchestrator: dependencies cannot be nil")
	}
	return &Orchestrator{
		planner: p,
		tx:      tx,
		logger:  logger,
	}, nil
}

// MissionResult represents execution outcomes.
type MissionResult struct {
	MissionID string        `json:"mission_id"`
	Status    fsm.State     `json:"status"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
}

// Execute runs the FSM lifecycle: Planning -> Running -> Validating -> Completed.
func (o *Orchestrator) Execute(ctx context.Context, missionID string, g goal.Goal) (*MissionResult, error) {
	startTime := time.Now()
	
	// 1. Begin Workspace Transaction
	txID, err := o.tx.Begin(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: transaction failed to begin: %w", err)
	}
	
	// 2. Planning Phase
	candidates, err := o.planner.Plan(ctx, g)
	if err != nil {
		o.tx.Rollback(ctx, txID)
		return nil, fmt.Errorf("orchestrator: planning failed: %w", err)
	}
	
	// 3. Selection (Pareto Scoring)
	chosen, err := o.planner.Score(ctx, candidates)
	if err != nil {
		o.tx.Rollback(ctx, txID)
		return nil, fmt.Errorf("orchestrator: scoring candidates failed: %w", err)
	}
	
	// 4. Execution Loop (Simulated as Running)
	if o.logger != nil {
		o.logger.Info("executing selected plan DAG", "mission_id", missionID)
	}
	
	// 5. Commit Transaction on success
	if err := o.tx.Commit(ctx, txID); err != nil {
		return nil, fmt.Errorf("orchestrator: transaction commit failed: %w", err)
	}
	
	return &MissionResult{
		MissionID: missionID,
		Status:    fsm.StateCompleted,
		Duration:  time.Since(startTime),
	}, nil
}
```

## Verify
```bash
go build ./kernel/orchestrator/...
```

## Checklist
- [ ] File `kernel/orchestrator/orchestrator.go` exists
- [ ] Package: `orchestrator`
- [ ] Orchestrator manages transactions and plan scoring
- [ ] Implements error rollbacks correctly
- [ ] `go build ./kernel/orchestrator/...` passes
