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
