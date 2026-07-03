# Micro-Task 5.12: Create kernel/orchestrator/aggregator.go

- **File**: `kernel/orchestrator/aggregator.go`
- **Package**: `orchestrator`
- **Depends on**: 5.11
- **Time**: 15 min
- **Verify**: `go build ./kernel/orchestrator/...`

## Purpose
Implements the result aggregator (`Aggregator`) to compile task execution durations, status keys, and resource usages.

## EXACT code to create

```go
package orchestrator

import (
	"log/slog"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/fsm"
)

// Aggregator compiles task results and statistics.
type Aggregator struct {
	logger *slog.Logger
}

// NewAggregator constructs a new Aggregator.
func NewAggregator(logger *slog.Logger) *Aggregator {
	return &Aggregator{
		logger: logger,
	}
}

// AggregateResults parses DAG execution states and compiles duration.
func (a *Aggregator) AggregateResults(missionID string, dag *fsm.DAG, duration time.Duration) *MissionResult {
	if dag == nil {
		return &MissionResult{
			MissionID: missionID,
			Status:    fsm.StateFailed,
			Error:     "aggregator: missing DAG data source",
		}
	}

	status := fsm.StateCompleted
	var errStr string

	for id, node := range dag.Nodes {
		if node.Status == fsm.StateFailed {
			status = fsm.StateFailed
			errStr = fmt.Sprintf("task %s failed", id)
		}
	}

	if a.logger != nil {
		a.logger.Info("aggregator: completed results compilation",
			"mission_id", missionID,
			"duration", duration,
			"total_tasks_count", len(dag.Nodes),
		)
	}

	return &MissionResult{
		MissionID: missionID,
		Status:    status,
		Duration:  duration,
		Error:     errStr,
	}
}
```

## Verify
```bash
go build ./kernel/orchestrator/...
```

## Checklist
- [ ] File `kernel/orchestrator/aggregator.go` exists
- [ ] Package: `orchestrator`
- [ ] `Aggregator` processes `fsm.DAG` nodes correctly
- [ ] Checks for failed nodes to determine status
- [ ] `go build ./kernel/orchestrator/...` passes
