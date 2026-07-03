package memory

import (
	"context"
	"sort"
	"testing"
	"time"

	contractsmemory "github.com/tiendat1751998/orchestrator/contracts/memory"
)

func TestInMemoryStore_Basic(t *testing.T) {
	ctx := context.Background()
	store, err := NewInMemoryStore("test-store")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// 1. Should return error when not started
	err = store.Save(ctx, "key1", "value1")
	if err == nil {
		t.Error("expected error when saving to unstarted store, got nil")
	}

	// Initialize and start the store
	if err := store.Init(ctx, nil); err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	if err := store.Start(ctx); err != nil {
		t.Fatalf("failed to start store: %v", err)
	}

	// 2. Validate empty key errors
	err = store.Save(ctx, "", "value")
	if err == nil {
		t.Error("expected error for empty key in Save, got nil")
	}
	var val string
	err = store.Load(ctx, "", &val)
	if err == nil {
		t.Error("expected error for empty key in Load, got nil")
	}

	// 3. Save, Load, and Delete
	err = store.Save(ctx, "key1", "value1")
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	var dest string
	err = store.Load(ctx, "key1", &dest)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	if dest != "value1" {
		t.Errorf("expected 'value1', got '%s'", dest)
	}

	// Delete
	err = store.Delete(ctx, "key1")
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	// Load deleted
	err = store.Load(ctx, "key1", &dest)
	if err == nil {
		t.Error("expected error loading deleted key, got nil")
	}
}

func TestInMemoryStore_TTL(t *testing.T) {
	ctx := context.Background()
	store, err := NewInMemoryStore("ttl-store")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	_ = store.Init(ctx, nil)
	_ = store.Start(ctx)

	// Save with TTL
	err = store.Save(ctx, "temp", "val", contractsmemory.WithTTL(30*time.Millisecond))
	if err != nil {
		t.Fatalf("failed to save with TTL: %v", err)
	}

	// Load immediately should succeed
	var dest string
	err = store.Load(ctx, "temp", &dest)
	if err != nil {
		t.Fatalf("failed to load before TTL expiration: %v", err)
	}

	// Wait for expiration
	time.Sleep(50 * time.Millisecond)

	// Load should now fail and clean up the key
	err = store.Load(ctx, "temp", &dest)
	if err == nil {
		t.Error("expected error after TTL expiration, got nil")
	}

	// Verify it's actually removed from the internal map
	store.mu.RLock()
	_, exists := store.storage["temp"]
	store.mu.RUnlock()
	if exists {
		t.Error("expected key to be deleted from internal map")
	}
}

func TestInMemoryStore_List(t *testing.T) {
	ctx := context.Background()
	store, err := NewInMemoryStore("list-store")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	_ = store.Init(ctx, nil)
	_ = store.Start(ctx)

	keys := []string{"task-1", "task-2", "user-1", "task-3"}
	for _, k := range keys {
		if err := store.Save(ctx, k, "data"); err != nil {
			t.Fatalf("failed to save: %v", err)
		}
	}

	res, err := store.List(ctx, "task-")
	if err != nil {
		t.Fatalf("failed to list keys: %v", err)
	}

	sort.Strings(res)
	expected := []string{"task-1", "task-2", "task-3"}
	if len(res) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(res))
	}
	for i := range expected {
		if res[i] != expected[i] {
			t.Errorf("expected key at index %d to be %s, got %s", i, expected[i], res[i])
		}
	}
}

func TestInMemoryStore_SearchScoringAndSorting(t *testing.T) {
	ctx := context.Background()
	store, err := NewInMemoryStore("search-store")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	_ = store.Init(ctx, nil)
	_ = store.Start(ctx)

	// We want to verify sorting by:
	// 1. Score descending
	// 2. CreatedAt descending (newer first) for ties

	// Save C: Score 0.5 (val match), older
	err = store.Save(ctx, "key-c", "query-val")
	if err != nil {
		t.Fatalf("save c: %v", err)
	}

	// Sleep to ensure distinct timestamp
	time.Sleep(5 * time.Millisecond)

	// Save D: Score 0.5 (val match), newer
	err = store.Save(ctx, "key-d", "query-val")
	if err != nil {
		t.Fatalf("save d: %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	// Save B: Score 0.8 (key match)
	err = store.Save(ctx, "key-query-b", "other")
	if err != nil {
		t.Fatalf("save b: %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	// Save A: Score 1.0 (tag match)
	err = store.Save(ctx, "key-a", "other", contractsmemory.WithTags("query"))
	if err != nil {
		t.Fatalf("save a: %v", err)
	}

	results, err := store.Search(ctx, "query", 10)
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}

	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}

	// Expect order: A (Score 1.0), B (Score 0.8), D (Score 0.5, newer), C (Score 0.5, older)
	if results[0].Key != "key-a" || results[0].Score != 1.0 {
		t.Errorf("expected index 0 to be key-a (Score 1.0), got %s (Score %f)", results[0].Key, results[0].Score)
	}
	if results[1].Key != "key-query-b" || results[1].Score != 0.8 {
		t.Errorf("expected index 1 to be key-query-b (Score 0.8), got %s (Score %f)", results[1].Key, results[1].Score)
	}
	if results[2].Key != "key-d" || results[2].Score != 0.5 {
		t.Errorf("expected index 2 to be key-d (Score 0.5, newer), got %s (Score %f)", results[2].Key, results[2].Score)
	}
	if results[3].Key != "key-c" || results[3].Score != 0.5 {
		t.Errorf("expected index 3 to be key-c (Score 0.5, older), got %s (Score %f)", results[3].Key, results[3].Score)
	}

	// Verify limit works
	limitedResults, err := store.Search(ctx, "query", 2)
	if err != nil {
		t.Fatalf("failed to search with limit: %v", err)
	}
	if len(limitedResults) != 2 {
		t.Errorf("expected 2 results with limit=2, got %d", len(limitedResults))
	}
	if limitedResults[0].Key != "key-a" || limitedResults[1].Key != "key-query-b" {
		t.Errorf("unexpected limited results: %+v", limitedResults)
	}
}

func TestInMemoryStore_DynamicExpirationCleanup(t *testing.T) {
	ctx := context.Background()
	store, err := NewInMemoryStore("cleanup-store")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	_ = store.Init(ctx, nil)
	_ = store.Start(ctx)

	// Save active and soon-to-expire keys
	if err := store.Save(ctx, "active-1", "data"); err != nil {
		t.Fatalf("failed to save active-1: %v", err)
	}
	if err := store.Save(ctx, "expired-1", "data", contractsmemory.WithTTL(10*time.Millisecond)); err != nil {
		t.Fatalf("failed to save expired-1: %v", err)
	}
	if err := store.Save(ctx, "expired-2", "data", contractsmemory.WithTTL(10*time.Millisecond)); err != nil {
		t.Fatalf("failed to save expired-2: %v", err)
	}

	time.Sleep(30 * time.Millisecond)

	// 1. Verify List cleans up expired keys
	keys, err := store.List(ctx, "")
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(keys) != 1 || keys[0] != "active-1" {
		t.Errorf("expected only 'active-1', got: %v", keys)
	}

	// Re-add expired keys to test Search cleanup
	if err := store.Save(ctx, "expired-3", "query-val", contractsmemory.WithTTL(10*time.Millisecond)); err != nil {
		t.Fatalf("failed to save expired-3: %v", err)
	}

	time.Sleep(30 * time.Millisecond)

	// 2. Verify Search cleans up expired keys
	results, err := store.Search(ctx, "val", 10)
	if err != nil {
		t.Fatalf("search error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 search results (all expired), got: %+v", results)
	}

	// Verify they are removed from the internal map
	store.mu.RLock()
	_, exists3 := store.storage["expired-3"]
	_, exists1 := store.storage["expired-1"]
	store.mu.RUnlock()

	if exists3 {
		t.Error("expected 'expired-3' to be dynamically cleaned up from storage")
	}
	if exists1 {
		t.Error("expected 'expired-1' to be dynamically cleaned up from storage")
	}
}
