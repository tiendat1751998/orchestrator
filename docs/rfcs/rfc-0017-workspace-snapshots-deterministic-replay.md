# RFC-0017: Workspace Snapshots & Deterministic Replay

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0013 (Workspace Engine), RFC-0008 (Event Model)

## Summary

This RFC specifies the design of **Workspace Snapshots & Deterministic Replay** in AEOS. To guarantee that a mission executed today replays identically 5 years later, the system captures environment parameters at mission startup: prompts, toolchain versions, OS parameters, Git commits, dirty file stashes, and policy versions.

## Motivation

AI model behavior, environment toolchains, and file states drift over time.
- If we attempt to replay a mission simply by re-executing the same prompts, differences in model APIs or compiler versions will yield divergent code.
- By snapshotting the exact environment and storing LLM raw outputs as immutable events, we can replay state transitions deterministically without calling external APIs.

## Design

### 1. Architectural Placement

Snapshots are captured by the `Workspace Engine` at mission start, and replay execution is handled by the `Execution Runtime`.

```
  Mission Start ──► [Workspace Engine: Capture Snapshot] ──► Save to Event Store
```

---

### 2. Contracts (`contracts/workspace/snapshot.go`)

```go
package workspace

import "context"

// EnvSnapshot represents captured workspace parameters.
type EnvSnapshot struct {
	MissionID    string            `json:"mission_id"`
	GoVersion    string            `json:"go_version"`
	GitCommit    string            `json:"git_commit"`
	EnvVariables map[string]string `json:"env_variables,omitempty"`
}

// SnapshotManager manages state snapshots.
type SnapshotManager interface {
	// Capture creates an environment snapshot.
	Capture(ctx context.Context, missionID string) (*EnvSnapshot, error)
	
	// Restore reverts the local workspace environment to match the snapshot.
	Restore(ctx context.Context, snapshot EnvSnapshot) error
}
```

## Impact

- **100% Deterministic Auditing**: Missions can be replayed frame-by-frame on any local developer machine to debug scheduling and execution failures.
- **Environment Parity**: Developer workspaces are validated to match the exact compiler parameters before execution.
