# Micro-Task 1.25: Tạo contracts/memory/memory.go

## Thông tin
- **File tạo**: `contracts/memory/memory.go`
- **Package**: `memory`
- **Dependencies trước**: 1.06 (contracts/types.go)
- **Thời gian**: 10 phút
- **Verify**: `go build ./contracts/memory/...`

## Nội dung CHÍNH XÁC cần tạo

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

## ⚠️ Pitfalls cần tránh
1. **Functional options pattern**: `SaveOption func(*saveOptions)` cho phép thêm options mà không break existing code. Nếu dùng `SaveConfig struct` → thêm field = break all callers.
2. **Load dest phải là pointer**: `store.Load(ctx, key, &result)` — pass non-pointer → runtime error. Document rõ.
3. **Delete idempotent**: Xóa key không tồn tại → nil (không lỗi). Giúp retry logic đơn giản.
4. **Search limit = 0**: Nghĩa là "no limit", KHÔNG phải "0 results". Document rõ.

## Checklist
- [ ] File `contracts/memory/memory.go` tồn tại
- [ ] Package: `package memory`
- [ ] Store interface với 5 methods
- [ ] Entry struct với 4 fields
- [ ] SaveOption functional options
- [ ] `WithTTL()` và `WithTags()` functions
- [ ] `ApplySaveOptions()` helper
- [ ] Godoc comments với examples
- [ ] `go build ./contracts/memory/...` không lỗi
