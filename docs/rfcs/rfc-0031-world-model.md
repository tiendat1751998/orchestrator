# RFC-0031: World Model

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0013 (Workspace Engine), RFC-0003 (Knowledge Engine)

## Summary

This RFC specifies the design of the **World Model** in AEOS. The World Model acts as the **living snapshot / digital twin** of the active workspace and infrastructure (git branches, dirty files, database health, containers, and deployment status). Unlike the long-term, static Knowledge Graph, the World Model resides in active RAM and represents the dynamic "now" state of the project.

## Motivation

To make correct planning decisions, the Planner needs to reason over typed codebase objects rather than raw prompt strings or static files.
- The Knowledge Graph only stores long-term facts (e.g. Gin is a Go framework). It has no awareness of whether the local Postgres container is running or if 12 files are modified.
- By separating long-term knowledge from the dynamic World Model, we prevent database bloat and ensure real-time accuracy during planning loops.

## Design

### 1. Architectural Placement

The World Model is updated in the background by the `Observation Runtime` and queried by the `Planner` during search loops.

```
  [Observation Runtime] ──► Updates ──► [World Model (RAM)] ◄── Queries ── [Planner]
```

---

### 2. Contracts (`contracts/world/`)

```go
package world

import "context"

// ProjectEntity represents a typed object in the workspace.
type ProjectEntity struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // "repository", "container", "database", "secret"
	Attributes map[string]interface{} `json:"attributes"`
}

// WorldState represents the active snapshot of the system.
type WorldState struct {
	Timestamp    int64                    `json:"timestamp"`
	Entities     map[string]ProjectEntity `json:"entities"`
	CurrentBranch string                   `json:"current_branch"`
	DirtyFiles   []string                 `json:"dirty_files"`
}

// WorldModel updates and queries the live state.
type WorldModel interface {
	// Snapshot captures the current live state of the project.
	Snapshot(ctx context.Context) (*WorldState, error)
	
	// Query returns a specific workspace entity by ID.
	Query(ctx context.Context, id string) (*ProjectEntity, error)
	
	// ApplyState records changes to the active memory cache.
	ApplyState(ctx context.Context, state WorldState) error
}
```

## Impact

- **Object-Oriented Planning**: The Planner reasons over typed objects (e.g., checking if `Database.Attributes["port"]` matches the app config) rather than parsing text.
- **Accurate Replays**: Environment parameters, branch names, and container states are saved in the Mission's snapshot record, ensuring deterministic replays.

## Open Questions

1. **How do we handle state out-of-sync bugs?**
   - The Observation Runtime periodically runs validation sweeps. If the cached World Model diverges from physical reality (e.g. a developer manually stopped a Docker container), a state change event is fired to trigger a replan.
