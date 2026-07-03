package planner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tiendat1751998/orchestrator/contracts/brain"
	"github.com/tiendat1751998/orchestrator/contracts/fsm"
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

// Replan intercepts a task failure, generates a corrective node to recover from it,
// mutates the active plan DAG to insert this corrective task, and returns the updated graph.
func (r *replanner) Replan(ctx context.Context, fCtx brain.FailureContext, activeDAG fsm.DAG) (fsm.DAG, error) {
	if ctx == nil {
		return fsm.DAG{}, fmt.Errorf("replanner: nil context")
	}
	if err := ctx.Err(); err != nil {
		return fsm.DAG{}, fmt.Errorf("replanner: context error: %w", err)
	}

	if fCtx.TaskID == "" {
		return fsm.DAG{}, fmt.Errorf("replanner: task ID cannot be empty")
	}

	if activeDAG.Nodes == nil {
		return fsm.DAG{}, fmt.Errorf("replanner: active DAG nodes map is nil")
	}

	// Find the failed task node
	failedNode, exists := activeDAG.Nodes[fCtx.TaskID]
	if !exists {
		return fsm.DAG{}, fmt.Errorf("replanner: failed task %q not found in active DAG", fCtx.TaskID)
	}

	// Deep copy DAG to prevent modification side-effects for the caller
	mutated := fsm.DAG{
		Nodes: make(map[string]*fsm.DAGNode, len(activeDAG.Nodes)),
	}
	for id, node := range activeDAG.Nodes {
		if node == nil {
			continue
		}
		mutated.Nodes[id] = &fsm.DAGNode{
			ID:           node.ID,
			Dependencies: append([]string(nil), node.Dependencies...),
			Status:       node.Status,
		}
	}

	// Re-fetch failedNode from mutated DAG
	failedNode = mutated.Nodes[fCtx.TaskID]

	// Determine category suffix for corrective task
	category := fCtx.Category
	if category == "" {
		category = "generic"
	}

	// Build a unique fix node ID, checking for collision and adding suffix if needed
	baseFixID := fmt.Sprintf("%s_fix_%s", fCtx.TaskID, category)
	fixTaskID := baseFixID
	suffix := 1
	for {
		if _, exists := mutated.Nodes[fixTaskID]; !exists {
			break
		}
		suffix++
		fixTaskID = fmt.Sprintf("%s_%d", baseFixID, suffix)
	}

	if r.logger != nil {
		r.logger.Info("intercepting task failure, generating corrective sub-graph",
			"task_id", fCtx.TaskID,
			"category", fCtx.Category,
			"fix_task_id", fixTaskID,
		)
	}

	// Create corrective task node
	// The corrective node depends on the failed task's original dependencies
	fixNode := &fsm.DAGNode{
		ID:           fixTaskID,
		Dependencies: failedNode.Dependencies,
		Status:       fsm.State("pending"),
	}

	// Wire it in: the failed task now depends *only* on the corrective node
	// Reset the failed task's status to pending to queue it for retry
	failedNode.Dependencies = []string{fixTaskID}
	failedNode.Status = fsm.State("pending")

	// Insert the corrective node into the DAG
	mutated.Nodes[fixTaskID] = fixNode

	return mutated, nil
}
