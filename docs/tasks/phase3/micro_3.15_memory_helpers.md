# Micro-Task 3.15: Create sdk/memory/memory.go

## Info
- **File**: `sdk/memory/memory.go`
- **Package**: `memory`
- **Depends on**: 3.01 (base_plugin.md), 1.25 (memory.go contract)
- **Time**: 20 min
- **Verify**: `go build ./sdk/memory/...`

## Purpose
Triển khai bộ lưu trữ in-memory (`InMemoryStore`) hoàn chỉnh hiện thực hóa giao diện `contracts/memory.Store`. Bộ helper này cung cấp khả năng đọc/ghi an toàn song song (thread-safe), hỗ trợ lọc thẻ (tags), tự động dọn dẹp theo thời gian sống (TTL expiration), và sử dụng kỹ thuật JSON marshal/unmarshal để sao chép dữ liệu vào con trỏ đích nhằm mô phỏng chính xác hành vi lưu trữ tuần tự hóa của cơ sở dữ liệu.

## EXACT code to create

```go
// Package memory provides in-memory implementations and helpers for agent memory stores.
package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	contractsplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	contractsmemory "github.com/tiendat1751998/orchestrator/contracts/memory"
	sdkplugin "github.com/tiendat1751998/orchestrator/sdk/plugin"
)

type internalEntry struct {
	value     []byte // Persisted as JSON bytes to simulate serialization bounds
	createdAt time.Time
	expiresAt time.Time
	tags      []string
}

func (e *internalEntry) isExpired() bool {
	if e.expiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.expiresAt)
}

// InMemoryStore implements contractsmemory.Store. Thread-safe.
type InMemoryStore struct {
	*sdkplugin.BasePlugin

	mu      sync.RWMutex
	storage map[string]*internalEntry
}

// NewInMemoryStore constructs a new InMemoryStore.
func NewInMemoryStore(name string) (*InMemoryStore, error) {
	basePlugin, err := sdkplugin.NewBasePlugin(name, contractsplugin.TypeMemory, "1.0.0")
	if err != nil {
		return nil, err
	}
	return &InMemoryStore{
		BasePlugin: basePlugin,
		storage:    make(map[string]*internalEntry),
	}, nil
}

// Save inserts or updates a key-value pair in memory.
func (s *InMemoryStore) Save(ctx context.Context, key string, value any, opts ...contractsmemory.SaveOption) error {
	if !s.IsStarted() {
		return fmt.Errorf("sdk/memory: store %q is not running", s.Name())
	}
	if key == "" {
		return errors.New("sdk/memory: key cannot be empty")
	}

	cfg := contractsmemory.ApplySaveOptions(opts...)

	// JSON Marshalling ensures the data can be serialized and breaks reference sharing.
	serialized, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("sdk/memory: failed to marshal value: %w", err)
	}

	var expires time.Time
	if cfg.TTL > 0 {
		expires = time.Now().Add(cfg.TTL)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.storage[key] = &internalEntry{
		value:     serialized,
		createdAt: time.Now(),
		expiresAt: expires,
		tags:      cfg.Tags,
	}

	return nil
}

// Load retrieves an item, deserializing it into dest (which must be a pointer).
func (s *InMemoryStore) Load(ctx context.Context, key string, dest any) error {
	if !s.IsStarted() {
		return fmt.Errorf("sdk/memory: store %q is not running", s.Name())
	}
	if key == "" {
		return errors.New("sdk/memory: key cannot be empty")
	}

	s.mu.Lock() // Write lock needed to delete expired entries on the fly
	entry, ok := s.storage[key]
	if ok && entry.isExpired() {
		delete(s.storage, key)
		entry = nil
		ok = false
	}
	s.mu.Unlock()

	if !ok || entry == nil {
		return fmt.Errorf("sdk/memory: key %q not found", key)
	}

	// Unmarshal back to target struct pointer
	if err := json.Unmarshal(entry.value, dest); err != nil {
		return fmt.Errorf("sdk/memory: failed to unmarshal into dest pointer: %w", err)
	}

	return nil
}

// Delete removes the item from the map. Idempotent.
func (s *InMemoryStore) Delete(ctx context.Context, key string) error {
	if !s.IsStarted() {
		return fmt.Errorf("sdk/memory: store %q is not running", s.Name())
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.storage, key)
	return nil
}

// List returns all active, non-expired keys matching the given prefix.
func (s *InMemoryStore) List(ctx context.Context, prefix string) ([]string, error) {
	if !s.IsStarted() {
		return nil, fmt.Errorf("sdk/memory: store %q is not running", s.Name())
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var keys []string

	for k, entry := range s.storage {
		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			delete(s.storage, k) // Clean up expired keys on scan
			continue
		}
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}

	return keys, nil
}

// Search queries memory entries by matching substring in keys, tags, or JSON string.
// Returns matched items sorted by score (relevance: 1.0 for tag match, 0.5 for content match).
func (s *InMemoryStore) Search(ctx context.Context, query string, limit int) ([]contractsmemory.Entry, error) {
	if !s.IsStarted() {
		return nil, fmt.Errorf("sdk/memory: store %q is not running", s.Name())
	}
	if query == "" {
		return nil, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var matched []contractsmemory.Entry
	lowerQuery := strings.ToLower(query)

	for k, entry := range s.storage {
		if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
			delete(s.storage, k)
			continue
		}

		score := 0.0

		// Check tag matches (Highest priority)
		for _, t := range entry.tags {
			if strings.ToLower(t) == lowerQuery {
				score = 1.0
				break
			}
		}

		// Check key matches
		if score == 0.0 && strings.Contains(strings.ToLower(k), lowerQuery) {
			score = 0.8
		}

		// Check raw value content matches
		valStr := string(entry.value)
		if score == 0.0 && strings.Contains(strings.ToLower(valStr), lowerQuery) {
			score = 0.5
		}

		if score > 0.0 {
			var unmarshalled any
			json.Unmarshal(entry.value, &unmarshalled) // Ignore error on query deserialization

			matched = append(matched, contractsmemory.Entry{
				Key:       k,
				Value:     unmarshalled,
				Score:     score,
				CreatedAt: entry.createdAt,
			})
		}
	}

	// Sort by score descending (highest score first)
	// If scores are equal, sort by CreatedAt descending (newest first)
	s.sortEntries(matched)

	if limit > 0 && len(matched) > limit {
		matched = matched[:limit]
	}

	return matched, nil
}

func (s *InMemoryStore) sortEntries(entries []contractsmemory.Entry) {
	// Stable sort to maintain consistency
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			shouldSwap := false
			if entries[i].Score < entries[j].Score {
				shouldSwap = true
			} else if entries[i].Score == entries[j].Score {
				if entries[i].CreatedAt.Before(entries[j].CreatedAt) {
					shouldSwap = true
				}
			}
			if shouldSwap {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}
```

## ⚠️ Pitfalls

### Pitfall 1: Shared Pointers References
```go
// ❌ WRONG:
s.storage[key] = value // Storing directly allows multiple threads to mutate the same struct reference!

// ✅ CORRECT:
bytes, _ := json.Marshal(value)
s.storage[key] = bytes // Decoupled: serializing copies the data.
```
In Go, saving a pointer to a struct in an in-memory map allows concurrent callers to read and write to the same memory space directly. Marshalling to JSON bytes isolates the database copy, exactly replicating SQL/Redis persistence behavior.

### Pitfall 2: Garbage accumulation of expired items
If TTL items are written but never loaded again, they will linger in memory forever. Our implementation cleans up expired records dynamically during `Load()`, `List()`, and `Search()`.

## Verify
```bash
go build ./sdk/memory/...
```

## Checklist
- [ ] File `sdk/memory/memory.go` exists
- [ ] Package: `memory`
- [ ] `InMemoryStore` embeds `*sdkplugin.BasePlugin`
- [ ] Internally serializes structs using `json.Marshal` to break memory sharing
- [ ] `Load` unmarshals bytes back into target pointer `dest`
- [ ] TTL verification handles automated record deletion correctly
- [ ] Search sorts elements by Score (relevance) descending
- [ ] Expired keys are deleted on scans (`List`, `Search`) to avoid memory leaks
- [ ] `go build ./sdk/memory/...` passes
