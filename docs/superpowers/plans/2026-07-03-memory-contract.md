# memory-contract Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the memory store contract (`contracts/memory/memory.go`) defining the storage interface and options.

**Architecture:** Defines the `Store` interface, `Entry` struct, functional save options, and helper `ApplySaveOptions` for setting TTL and tags.

**Tech Stack:** Go (Standard Library only - `context`, `time`).

## Global Constraints

- Must reside in `contracts/memory/memory.go` under package `memory`.
- Zero external dependencies or imports outside the standard library (`context`, `time`).
- Strictly adhere to layer boundaries (no imports from kernel, sdk, etc.).
- Named fields must be used for any struct initializations.

---

### Task 1: Create Memory Contract

**Files:**
- Create: `contracts/memory/memory.go`

**Interfaces:**
- Consumes: Stdlib only (`context`, `time`)
- Produces: `Store` (interface), `Entry` (struct), `SaveOption` (type), `WithTTL` (function), `WithTags` (function), `ApplySaveOptions` (function)

- [ ] **Step 1: Write the minimal implementation**

Create `contracts/memory/memory.go` containing:
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

- [ ] **Step 2: Run verification**

Run: `go build ./contracts/memory/...`
Expected: PASS with no compilation errors.
