# RFC-0046: Execution Graph Manager

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0000 (Everything is State Machine), RFC-0037 (Adaptive Recovery)

## Summary

This RFC specifies the design of the **Execution Graph Manager** in AEOS. The Execution Graph Manager is responsible for pausing, versioning, merging, and migrating plan DAGs at runtime when task failures occur and the Replanning Engine generates corrective sub-graphs.

## Motivation

When a task fails mid-mission, the remaining steps in the execution queue might become invalid.
- Simply appending tasks to the end of the queue does not handle dependencies (e.g., if step 2 failed, we must inject a fix before resuming step 3).
- We need a thread-safe, transaction-aware manager to mutate the execution queue dynamically, dropping orphaned nodes and versioning the DAG.

## Design

### 1. Architectural Placement

The Execution Graph Manager resides in the `Execution Runtime`, controlling the Scheduler's active task queues.

```
  Task Failure ──► [Graph Manager] ──► Lock Queue ──► Merge Corrective Sub-Graph ──► Resume
```

---

### 2. Contracts (`contracts/fsm/graph.go`)

```go
package fsm

import (
	"context"
)

// DAGVersion tracks changes to the plan.
type DAGVersion struct {
	VersionID   string   `json:"version_id"`
	ParentID    string   `json:"parent_id,omitempty"`
	Timestamp   int64    `json:"timestamp"`
	NodesAdded  []string `json:"nodes_added,omitempty"`
	NodesRemoved []string `json:"nodes_removed,omitempty"`
}

// ExecutionGraphManager coordinates plan queue mutations.
type ExecutionGraphManager interface {
	// ActiveGraph returns the current version of the plan DAG.
	ActiveGraph(ctx context.Context, missionID string) (*DAG, error)
	
	// MutateGraph pauses execution, merges corrective nodes, drops deadlocks, and resumes.
	MutateGraph(ctx context.Context, missionID string, subGraph DAG) (*DAGVersion, error)
	
	// GetHistory returns the audit tree of plan mutations.
	GetHistory(ctx context.Context, missionID string) ([]DAGVersion, error)
}
```

## Impact

- **Safe Live Mutations**: The active scheduler queue can be safely paused and updated mid-mission without losing execution progress or violating security boundaries.
- **Clear Audit Trail**: The DAG version tree provides a complete history of how a plan evolved from its initial draft to its final successful completion.

## Open Questions

1. **How do we handle parallel execution conflicts?**
   - When a mutation occurs, the manager halts all running tasks that depend on the changed branches, allowing independent branches to compile or run safely.
