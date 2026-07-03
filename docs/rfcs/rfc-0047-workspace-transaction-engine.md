# RFC-0047: Workspace Transaction Engine

- **Status**: PROPOSED
- **Priority**: P2 — Extended
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0013 (Workspace Engine), RFC-0012 (Security & Capability Model)

## Summary

This RFC specifies the design of the **Workspace Transaction Engine** in AEOS. To guarantee filesystem safety during generative AI coding tasks, the engine manages Git-backed transactions. Prior to running code mutations, dirty files are stashed, and a transaction branch/checkpoint is created. If validation fails in the Truth Pipeline, the engine executes a hard Git rollback to restore workspace integrity.

## Motivation

AI agents often write code that introduces syntax errors, breaks compile paths, or corrupts files.
- Simply leaving the codebase in a dirty, uncompilable state halts developer productivity.
- Virtual file systems or Docker mounts are over-engineered and break local IDE tools (watches, autocompletion).
- Using local Git staging provides zero-overhead, transaction-like file safety natively.

## Design

### 1. Architectural Placement

The Transaction Engine resides inside the `Workspace Engine` in the Kernel, wrapping all file modification actions.

```
  Start Tx ──► Git Checkpoint ──► Code Gen ──► Verify (DoD) ──► Fail? ──► Git Hard Reset
                                                                └── Commit Tx
```

---

### 2. Contracts (`contracts/workspace/transaction.go`)

```go
package workspace

import "context"

// TransactionID identifies a specific filesystem state change block.
type TransactionID string

// WorkspaceTransactionEngine provides local git-backed transactions.
type WorkspaceTransactionEngine interface {
	// Begin starts a new transaction, stashing dirty files and creating a tx checkpoint.
	Begin(ctx context.Context, missionID string) (TransactionID, error)
	
	// Commit finalizes the changes, merging the tx branch and clearing stashes.
	Commit(ctx context.Context, txID TransactionID) error
	
	// Rollback discards all changes, executing a hard Git reset to restore integrity.
	Rollback(ctx context.Context, txID TransactionID) error
}
```

## Impact

- **Zero-Risk Code Generation**: Agents can safely modify 100 files simultaneously. If compilation fails and recovery is aborted, the Workspace is restored to its exact original state within 100ms.
- **Local Developer Harmony**: No FUSE or container mounts are required, allowing human developers to keep their IDEs open on the active project.

## Open Questions

1. **How do we handle uncommitted human changes when the mission starts?**
   - The engine runs `git stash --include-untracked` at transaction start, preserving the human developer's dirty working directory and restoring it on mission commit/rollback.
