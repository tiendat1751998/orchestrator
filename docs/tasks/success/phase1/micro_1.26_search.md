# Micro-Task 1.26: Create contracts/search/search.go

## Info
- **File**: `contracts/search/search.go`
- **Package**: `search`
- **Depends on**: 1.06 (contracts/types.go)
- **Time**: 10 min
- **Verify**: `go build ./contracts/search/...`

## Purpose
Defines the `Engine` and `Indexable` interfaces, supporting structs, and functional configuration options for full-text code, documentation, and data indexing.

## EXACT code to create

```go
// Package search defines the contract for search and indexing.
// The search engine allows agents to find relevant code, docs, and data.
package search

import "context"

// Engine provides search and indexing capabilities.
//
// Implementations:
//   - BleveSearch (local full-text search)
//   - ElasticSearch (distributed)
//   - Simple grep-based (Phase 1 default)
type Engine interface {
	// Index adds items to the search index.
	// Existing items with the same ID are updated.
	Index(ctx context.Context, items []Indexable) error

	// Search queries the index and returns matching results.
	// Results are sorted by relevance (highest score first).
	Search(ctx context.Context, query string, opts ...SearchOption) ([]Result, error)
}

// Indexable is anything that can be indexed for search.
// Types that want to be searchable must implement this interface.
type Indexable interface {
	// SearchID returns a unique identifier for this item.
	SearchID() string

	// SearchContent returns the text content to index.
	SearchContent() string

	// SearchMetadata returns key-value metadata for filtering.
	SearchMetadata() map[string]string
}

// Result represents a search match.
type Result struct {
	// ID is the unique identifier of the matched item.
	ID string `json:"id"`

	// Content is the matched text content (may be truncated).
	Content string `json:"content"`

	// Score is the relevance score (0.0 to 1.0, higher = more relevant).
	Score float64 `json:"score"`

	// Metadata contains the item's metadata (from SearchMetadata()).
	Metadata map[string]string `json:"metadata"`
}

// =============================================================================
// Search Options
// =============================================================================

// SearchOption configures search behavior.
type SearchOption func(*searchOptions)

type searchOptions struct {
	MaxResults int
	MinScore   float64
	Filters    map[string]string
}

// WithMaxResults limits the number of results returned.
// Default: 50
func WithMaxResults(n int) SearchOption {
	return func(o *searchOptions) { o.MaxResults = n }
}

// WithMinScore filters out results below a minimum relevance score.
// Default: 0.0 (include all results)
func WithMinScore(score float64) SearchOption {
	return func(o *searchOptions) { o.MinScore = score }
}

// WithFilter adds a metadata filter.
// Only results where Metadata[key] == value are included.
func WithFilter(key, value string) SearchOption {
	return func(o *searchOptions) {
		if o.Filters == nil {
			o.Filters = make(map[string]string)
		}
		o.Filters[key] = value
	}
}

// ApplySearchOptions processes functional options.
func ApplySearchOptions(opts ...SearchOption) searchOptions {
	o := searchOptions{
		MaxResults: 50, // Default
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
```

## Rules
1. **Indexable Flexibility**: Any model, text file, or task log can implement `Indexable` to participate in indexing pipelines.
2. **Filters Map Safety**: Functional options that assign to maps (like `WithFilter`) must initialize the maps if they are `nil`.
3. **Default Max Limit**: `ApplySearchOptions` must define a default result boundary limit (e.g. `50`) to avoid large search payload fetches.

## ⚠️ Pitfalls

### Pitfall 1: Panic on uninitialized filters map in `WithFilter` option
```go
func WithFilter(key, value string) SearchOption {
    return func(o *searchOptions) {
        if o.Filters == nil {
            o.Filters = make(map[string]string)
        }
        o.Filters[key] = value
    }
}
```
Always check and instantiate dynamic configurations maps before writing values inside options callbacks.

### Pitfall 2: Re-generating large indexes on single document updates
If the indexer treats document additions as complete index rebuilds, latency increases exponentially with catalog size. Implementations must treat duplicate `SearchID` insertions as targeted replacements/updates.

## Verify
```bash
go build ./contracts/search/...
```

## Checklist
- [ ] File `contracts/search/search.go` exists
- [ ] Package: `search`
- [ ] `Engine` interface contains `Index` and `Search` methods
- [ ] `Indexable` interface contains SearchID, SearchContent, and SearchMetadata methods
- [ ] `Result` contains ID, Content, Score, and Metadata fields
- [ ] Functional options `WithMaxResults`, `WithMinScore`, and `WithFilter` exist
- [ ] `WithFilter` handles nil map initialization safely
- [ ] `go build ./contracts/search/...` passes
