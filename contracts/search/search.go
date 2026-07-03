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
