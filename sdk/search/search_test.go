package search

import (
	"context"
	"testing"

	contractssearch "github.com/tiendat1751998/orchestrator/contracts/search"
)

type testItem struct {
	id       string
	content  string
	metadata map[string]string
}

func (t testItem) SearchID() string {
	return t.id
}

func (t testItem) SearchContent() string {
	return t.content
}

func (t testItem) SearchMetadata() map[string]string {
	return t.metadata
}

func TestInMemorySearchEngine_NotStarted(t *testing.T) {
	engine, err := NewInMemorySearchEngine("test-engine")
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}

	ctx := context.Background()
	items := []contractssearch.Indexable{
		testItem{id: "1", content: "hello"},
	}

	// Verify Index returns error when not started
	if err := engine.Index(ctx, items); err == nil {
		t.Error("Expected error calling Index on unstarted search engine, got nil")
	}

	// Verify Search returns error when not started
	if _, err := engine.Search(ctx, "hello"); err == nil {
		t.Error("Expected error calling Search on unstarted search engine, got nil")
	}
}

func TestInMemorySearchEngine_IndexingAndBasicSearch(t *testing.T) {
	engine, err := NewInMemorySearchEngine("test-engine")
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}

	ctx := context.Background()

	// Init and Start the plugin
	if err := engine.Init(ctx, nil); err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}

	// Index nil/empty boundary conditions
	if err := engine.Index(ctx, nil); err != nil {
		t.Errorf("Expected nil index list to succeed, got: %v", err)
	}
	if err := engine.Index(ctx, []contractssearch.Indexable{}); err != nil {
		t.Errorf("Expected empty index list to succeed, got: %v", err)
	}

	// Invalid items check
	if err := engine.Index(ctx, []contractssearch.Indexable{nil}); err == nil {
		t.Error("Expected error when indexing nil item, got nil")
	}
	if err := engine.Index(ctx, []contractssearch.Indexable{testItem{id: "", content: "hello"}}); err == nil {
		t.Error("Expected error when indexing item with empty ID, got nil")
	}

	// Index valid items
	items := []contractssearch.Indexable{
		testItem{
			id:      "doc-1",
			content: "The quick brown fox jumps over the lazy dog",
			metadata: map[string]string{
				"author": "fox",
				"type":   "animal",
			},
		},
		testItem{
			id:      "doc-2",
			content: "Heavy rain hits the city center",
			metadata: map[string]string{
				"author": "reporter",
				"type":   "weather",
			},
		},
	}

	if err := engine.Index(ctx, items); err != nil {
		t.Fatalf("Failed to index items: %v", err)
	}

	// Basic query match
	results, err := engine.Search(ctx, "fox")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].ID != "doc-1" {
		t.Errorf("Expected matched ID to be 'doc-1', got %q", results[0].ID)
	}

	if results[0].Metadata["author"] != "fox" {
		t.Errorf("Expected metadata 'author' to be 'fox', got %q", results[0].Metadata["author"])
	}
}

func TestInMemorySearchEngine_RelevanceScoring(t *testing.T) {
	engine, err := NewInMemorySearchEngine("test-engine")
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}

	ctx := context.Background()
	_ = engine.Init(ctx, nil)
	_ = engine.Start(ctx)

	items := []contractssearch.Indexable{
		testItem{id: "exact-match", content: "apple"},
		testItem{id: "prefix-match", content: "apple pie"},
		testItem{id: "substring-match", content: "crabapple"},
		testItem{id: "apple-id-match", content: "some random fruit"},
		testItem{id: "no-match", content: "banana"},
	}

	if err := engine.Index(ctx, items); err != nil {
		t.Fatalf("Failed to index: %v", err)
	}

	// 1. Check matching weights for query "apple"
	results, err := engine.Search(ctx, "apple")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// We expect 4 matching items: exact, prefix, substring, ID-match.
	if len(results) != 4 {
		t.Fatalf("Expected 4 results, got %d", len(results))
	}

	for _, res := range results {
		switch res.ID {
		case "exact-match":
			if res.Score != 1.0 {
				t.Errorf("Expected exact-match score to be 1.0, got %f", res.Score)
			}
		case "prefix-match":
			if res.Score != 0.8 {
				t.Errorf("Expected prefix-match score to be 0.8, got %f", res.Score)
			}
		case "substring-match":
			if res.Score != 0.5 {
				t.Errorf("Expected substring-match score to be 0.5, got %f", res.Score)
			}
		case "apple-id-match":
			if res.Score != 0.3 {
				t.Errorf("Expected apple-id-match score to be 0.3, got %f", res.Score)
			}
		default:
			t.Errorf("Unexpected result item: %s", res.ID)
		}
	}

	// 2. Check score 0.1 for empty query
	emptyQueryResults, err := engine.Search(ctx, "")
	if err != nil {
		t.Fatalf("Empty search failed: %v", err)
	}

	// All 5 items should match since empty query assigns score 0.1
	if len(emptyQueryResults) != 5 {
		t.Fatalf("Expected 5 results for empty query, got %d", len(emptyQueryResults))
	}
	for _, res := range emptyQueryResults {
		if res.Score != 0.1 {
			t.Errorf("Expected empty query score to be 0.1 for item %s, got %f", res.ID, res.Score)
		}
	}
}

func TestInMemorySearchEngine_MetadataFilters(t *testing.T) {
	engine, err := NewInMemorySearchEngine("test-engine")
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}

	ctx := context.Background()
	_ = engine.Init(ctx, nil)
	_ = engine.Start(ctx)

	items := []contractssearch.Indexable{
		testItem{
			id:      "apple-1",
			content: "apple juice",
			metadata: map[string]string{
				"category": "fruit",
				"origin":   "US",
			},
		},
		testItem{
			id:      "apple-2",
			content: "apple sauce",
			metadata: map[string]string{
				"category": "fruit",
				"origin":   "FR",
			},
		},
		testItem{
			id:      "potato-1",
			content: "potato fries",
			metadata: map[string]string{
				"category": "vegetable",
				"origin":   "US",
			},
		},
	}

	if err := engine.Index(ctx, items); err != nil {
		t.Fatalf("Failed to index: %v", err)
	}

	// Filter: category = fruit
	res, err := engine.Search(ctx, "apple", contractssearch.WithFilter("category", "fruit"))
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(res) != 2 {
		t.Errorf("Expected 2 fruit results, got %d", len(res))
	}

	// Filter: origin = US
	res, err = engine.Search(ctx, "apple", contractssearch.WithFilter("origin", "US"))
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(res) != 1 || res[0].ID != "apple-1" {
		t.Errorf("Expected apple-1 from US, got results: %+v", res)
	}

	// Filter: origin = US, category = fruit
	res, err = engine.Search(ctx, "",
		contractssearch.WithFilter("origin", "US"),
		contractssearch.WithFilter("category", "fruit"),
	)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(res) != 1 || res[0].ID != "apple-1" {
		t.Errorf("Expected apple-1, got %+v", res)
	}

	// Non-matching filter
	res, err = engine.Search(ctx, "apple", contractssearch.WithFilter("category", "meat"))
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("Expected 0 results for non-matching filter, got %d", len(res))
	}
}

func TestInMemorySearchEngine_SortingAndTies(t *testing.T) {
	engine, err := NewInMemorySearchEngine("test-engine")
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}

	ctx := context.Background()
	_ = engine.Init(ctx, nil)
	_ = engine.Start(ctx)

	// Items designed to test score sort first, and then ID tie-breaking
	items := []contractssearch.Indexable{
		testItem{id: "doc-z", content: "apple pie recipe"}, // Prefix match (0.8)
		testItem{id: "doc-a", content: "apple"},            // Exact match (1.0)
		testItem{id: "doc-m", content: "apple"},            // Exact match (1.0)
		testItem{id: "doc-x", content: "crabapple cake"},   // Substring match (0.5)
	}

	if err := engine.Index(ctx, items); err != nil {
		t.Fatalf("Failed to index: %v", err)
	}

	res, err := engine.Search(ctx, "apple")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(res) != 4 {
		t.Fatalf("Expected 4 results, got %d", len(res))
	}

	// Expected order:
	// 1. doc-a (1.0, tie breaker first lexicographically)
	// 2. doc-m (1.0)
	// 3. doc-z (0.8)
	// 4. doc-x (0.5)

	expectedOrder := []string{"doc-a", "doc-m", "doc-z", "doc-x"}
	for i, expectedID := range expectedOrder {
		if res[i].ID != expectedID {
			t.Errorf("Expected index %d to be %q, got %q (score: %f)", i, expectedID, res[i].ID, res[i].Score)
		}
	}
}

func TestInMemorySearchEngine_LimitAndMinScore(t *testing.T) {
	engine, err := NewInMemorySearchEngine("test-engine")
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}

	ctx := context.Background()
	_ = engine.Init(ctx, nil)
	_ = engine.Start(ctx)

	items := []contractssearch.Indexable{
		testItem{id: "doc-1", content: "apple"},         // Exact match (1.0)
		testItem{id: "doc-2", content: "apple pie"},     // Prefix match (0.8)
		testItem{id: "doc-3", content: "crabapple"},     // Substring match (0.5)
		testItem{id: "apple-4", content: "random text"}, // ID match (0.3)
	}

	if err := engine.Index(ctx, items); err != nil {
		t.Fatalf("Failed to index: %v", err)
	}

	// Test MinScore configuration
	res, err := engine.Search(ctx, "apple", contractssearch.WithMinScore(0.6))
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	// Only 1.0 (doc-1) and 0.8 (doc-2) should match.
	if len(res) != 2 {
		t.Errorf("Expected 2 results with MinScore 0.6, got %d", len(res))
	}
	for _, r := range res {
		if r.Score < 0.6 {
			t.Errorf("Returned result %q has score %f which is less than 0.6", r.ID, r.Score)
		}
	}

	// Test MaxResults configuration
	res, err = engine.Search(ctx, "apple", contractssearch.WithMaxResults(2))
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(res) != 2 {
		t.Errorf("Expected 2 results due to MaxResults, got %d", len(res))
	}
	// The results should be the top scoring ones (doc-1, doc-2)
	if res[0].ID != "doc-1" || res[1].ID != "doc-2" {
		t.Errorf("Expected results doc-1 and doc-2, got: %+v", res)
	}
}

func TestInMemorySearchEngine_MetadataDeepCopy(t *testing.T) {
	engine, err := NewInMemorySearchEngine("test-engine")
	if err != nil {
		t.Fatalf("Failed to create search engine: %v", err)
	}

	ctx := context.Background()
	_ = engine.Init(ctx, nil)
	_ = engine.Start(ctx)

	sharedMeta := map[string]string{
		"category": "fruit",
	}
	item := testItem{
		id:       "copy-test",
		content:  "apple",
		metadata: sharedMeta,
	}

	if err := engine.Index(ctx, []contractssearch.Indexable{item}); err != nil {
		t.Fatalf("Failed to index: %v", err)
	}

	// Modify the original metadata
	sharedMeta["category"] = "vegetable"

	// Fetch result
	res, err := engine.Search(ctx, "apple")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(res) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(res))
	}

	// Verify that the stored value is unaffected by the post-index modification
	if res[0].Metadata["category"] != "fruit" {
		t.Errorf("Expected metadata 'category' to remain 'fruit', got %q", res[0].Metadata["category"])
	}

	// Modify metadata in fetched result
	res[0].Metadata["category"] = "grain"

	// Fetch result again
	res2, err := engine.Search(ctx, "apple")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Verify that the stored value is unaffected by the post-search modification
	if res2[0].Metadata["category"] != "fruit" {
		t.Errorf("Expected second search metadata 'category' to remain 'fruit', got %q", res2[0].Metadata["category"])
	}
}
