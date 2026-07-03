package orchestrator

import (
	"fmt"
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
			Status:    fsm.State("failed"), // ponytail: fsm.StateFailed is not defined in contracts/fsm/fsm.go
			Error:     "aggregator: missing DAG data source",
		}
	}

	status := fsm.StateCompleted
	var errStr string

	for id, node := range dag.Nodes {
		if node.Status == fsm.State("failed") { // ponytail: fsm.StateFailed is not defined in contracts/fsm/fsm.go
			status = fsm.State("failed") // ponytail: fsm.StateFailed is not defined in contracts/fsm/fsm.go
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
