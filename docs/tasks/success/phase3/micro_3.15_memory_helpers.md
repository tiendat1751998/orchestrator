# Micro-Task 3.15: Create sdk/memory/memory.go

## Info
- **File**: `sdk/memory/memory.go`
- **Package**: `memory`
- **Depends on**: 3.01 (base_plugin.md), 1.25 (memory.go contract)
- **Time**: 20 min
- **Verify**: `go build ./sdk/memory/...`

## Purpose
Implements the in-memory store (`InMemoryStore` and internal wrappers) satisfying `contracts/memory.Store`, providing thread-safe operations, TTL expirations, tag searches, and mock database persistence boundaries.

## EXACT code to create

```go
// Package memory provides in-memory implementations and helpers for agent memory stores.
package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	contractsplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	contractsmemory "github.com/tiendat1751998/orchestrator/contracts/memory"
	sdkplugin "github.com/tiendat1751998/orchestrator/sdk/plugin"
)

type internalEntry struct {
	value     []byte
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

	s.mu.Lock()
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
			delete(s.storage, k)
			continue
		}
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}

	return keys, nil
}

// Search queries memory entries by matching substring in keys, tags, or JSON string.
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

		for _, t := range entry.tags {
			if strings.ToLower(t) == lowerQuery {
				score = 1.0
				break
			}
		}

		if score == 0.0 && strings.Contains(strings.ToLower(k), lowerQuery) {
			score = 0.8
		}

		valStr := string(entry.value)
		if score == 0.0 && strings.Contains(strings.ToLower(valStr), lowerQuery) {
			score = 0.5
		}

		if score > 0.0 {
			var unmarshalled any
			_ = json.Unmarshal(entry.value, &unmarshalled)

			matched = append(matched, contractsmemory.Entry{
				Key:       k,
				Value:     unmarshalled,
				Score:     score,
				CreatedAt: entry.createdAt,
			})
		}
	}

	s.sortEntries(matched)

	if limit > 0 && len(matched) > limit {
		matched = matched[:limit]
	}

	return matched, nil
}

func (s *InMemoryStore) sortEntries(entries []contractsmemory.Entry) {
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

## Rules
1. **JSON Serialization Boundaries**: Serialize input values to JSON byte arrays on storage `Save()`, and decode on `Load()` back into destination pointers. This breaks in-memory reference sharing between concurrent routines and replicates database boundaries.
2. **Access Synchronization**: Guard the internal map using a read/write mutex (`sync.RWMutex`). Expiration writes during read operations (`Load()`, `List()`, `Search()`) require write locks.
3. **Relevance Sorting**: Sort query outcomes descending by `Score` (1.0 for tag match, 0.8 for key prefix match, 0.5 for JSON substring match) and descending by `CreatedAt` timestamps for ties.

## ⚠️ Pitfalls

### Pitfall 1: Storing raw struct pointers directly in map
Storing struct references directly allows multiple goroutines to read and modify the same memory locations concurrently, causing data races. Always serialize to bytes to decouple memory.

### Pitfall 2: Leaking expired TTL entries in storage
If entries expire but are never queried via `Load`, they accumulate in memory indefinitely. Clean up expired entries dynamically during `List` and `Search` scans.

## Verify
```bash
go build ./sdk/memory/...
```

## Checklist
- [ ] File `sdk/memory/memory.go` exists
- [ ] Package: `memory`
- [ ] `InMemoryStore` implements `contracts/memory.Store`
- [ ] `Save` serializes objects to JSON bytes to ensure isolation
- [ ] `Load` unmarshals bytes back into target pointer references
- [ ] Expired TTL records are deleted dynamically
- [ ] `Search` sorts results by Score and CreatedAt
- [ ] `List` and `Search` clean up expired keys during scans
- [ ] `go build ./sdk/memory/...` passes
