# Micro-Task 1.25: Create contracts/memory/memory.go

## Info
- **File**: `contracts/memory/memory.go`
- **Package**: `memory`
- **Depends on**: 1.06 (contracts/types.go)
- **Time**: 10 min
- **Verify**: `go build ./contracts/memory/...`

## Purpose
Defines the `Store` interface and supporting types for persistent storage of state, memory logs, and patterns. It leverages Go's functional options pattern for extensible parameters like TTL or tagging.

## EXACT code to create

```go
// Package memory defines the contract for persistent storage.
// Memory stores allow agents and the kernel to save/retrieve data
// across sessions (e.g., past task results, learned patterns).
package memory

import (
	"context"
	"time"
)

// Store provides persistent key-value storage with search.
//
// Implementations:
//   - In-memory (for testing)
//   - File-based (for local development)
//   - SQLite (for single-user)
//   - Redis/PostgreSQL (for production)
type Store interface {
	// Save stores a value with the given key.
	// If the key already exists, the value is overwritten.
	//
	// Options:
	//   - WithTTL(duration): auto-delete after duration
	//   - WithTags(tags...): add searchable tags
	Save(ctx context.Context, key string, value any, opts ...SaveOption) error

	// Load retrieves a value by key.
	// dest must be a pointer to the expected type.
	// Returns an error if the key doesn't exist.
	//
	// Example:
	//   var result agent.Result
	//   err := store.Load(ctx, "task-123-result", &result)
	Load(ctx context.Context, key string, dest any) error

	// Delete removes a value by key.
	// Returns nil if the key doesn't exist (idempotent).
	Delete(ctx context.Context, key string) error

	// Search finds entries matching a text query.
	// Returns entries sorted by relevance (highest score first).
	// limit controls the maximum number of results (0 = no limit).
	Search(ctx context.Context, query string, limit int) ([]Entry, error)

	// List returns all keys matching a prefix.
	// Example: List(ctx, "task-") returns ["task-123", "task-456"]
	List(ctx context.Context, prefix string) ([]string, error)
}

// Entry represents a stored item returned by Search.
type Entry struct {
	// Key is the storage key.
	Key string `json:"key"`

	// Value is the stored data.
	Value any `json:"value"`

	// Score is the relevance score for search results (0.0 to 1.0).
	// Higher = more relevant. Only populated by Search().
	Score float64 `json:"score,omitempty"`

	// CreatedAt is when the entry was first created.
	CreatedAt time.Time `json:"created_at"`
}

// =============================================================================
// Save Options (functional options pattern)
// =============================================================================

// SaveOption configures Save behavior.
type SaveOption func(*saveOptions)

type saveOptions struct {
	TTL  time.Duration
	Tags []string
}

// WithTTL sets a time-to-live for the entry.
// After the TTL expires, the entry is automatically deleted.
func WithTTL(d time.Duration) SaveOption {
	return func(o *saveOptions) { o.TTL = d }
}

// WithTags adds searchable tags to the entry.
// Tags can be used to filter entries without full-text search.
func WithTags(tags ...string) SaveOption {
	return func(o *saveOptions) { o.Tags = tags }
}

// ApplySaveOptions processes functional options into a saveOptions struct.
// Used by Store implementations to read the options.
func ApplySaveOptions(opts ...SaveOption) saveOptions {
	var o saveOptions
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
```

## Rules
1. **Pointer Dest Verification**: `Load` must receive a pointer destination. Values are populated using JSON/YAML deserialization or direct assignment under interface pointers.
2. **Delete Idempotence**: Invoking `Delete` on keys that do not exist must not return an error (return `nil`).
3. **Limit Boundary**: A search parameter limit of `0` denotes unbounded results (all matches), not zero results.

## ⚠️ Pitfalls

### Pitfall 1: Passing non-pointers to the `Load` destination parameter
```go
var res agent.Result
err := store.Load(ctx, "res_key", &res) // Correctly updates the targeted structure memory.
```
Always verify that parameters passed to `dest` are pointer addresses.

### Pitfall 2: Reusing struct configs instead of functional options
If you define `Save` parameters using a plain config struct (`Save(ctx, key, val, cfg SaveConfig)`), adding new options later requires updating all implementations in the codebase. The Functional Options pattern (`WithTTL`, `WithTags`) allows backward-compatible API expansions.

## Verify
```bash
go build ./contracts/memory/...
```

## Checklist
- [ ] File `contracts/memory/memory.go` exists
- [ ] Package: `memory`
- [ ] `Store` interface contains Save, Load, Delete, Search, and List methods
- [ ] `Entry` contains Key, Value, Score, and CreatedAt fields
- [ ] Functional options `WithTTL` and `WithTags` exist
- [ ] `ApplySaveOptions` processes configuration arguments correctly
- [ ] `go build ./contracts/memory/...` passes
