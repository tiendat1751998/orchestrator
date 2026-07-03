// Package search provides in-memory implementations and options for search engines.
package search

import (
	"context"
	"fmt"
	"strings"
	"sync"

	contractsplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	contractssearch "github.com/tiendat1751998/orchestrator/contracts/search"
	sdkplugin "github.com/tiendat1751998/orchestrator/sdk/plugin"
)

// ponytail: in-memory non-persistent storage. SQLite or Bleve if persistence/indexing is needed.
type internalItem struct {
	id       string
	content  string
	metadata map[string]string
}

// InMemorySearchEngine implements contractssearch.Engine. Thread-safe.
type InMemorySearchEngine struct {
	*sdkplugin.BasePlugin

	mu    sync.RWMutex
	items map[string]*internalItem
}

// NewInMemorySearchEngine constructs a new InMemorySearchEngine.
func NewInMemorySearchEngine(name string) (*InMemorySearchEngine, error) {
	basePlugin, err := sdkplugin.NewBasePlugin(name, contractsplugin.TypeSearch, "1.0.0")
	if err != nil {
		return nil, err
	}
	return &InMemorySearchEngine{
		BasePlugin: basePlugin,
		items:      make(map[string]*internalItem),
	}, nil
}

// Index adds or updates items inside the search engine registry.
func (s *InMemorySearchEngine) Index(ctx context.Context, items []contractssearch.Indexable) error {
	if !s.IsStarted() {
		return fmt.Errorf("sdk/search: search engine %q is not running", s.Name())
	}
	if len(items) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, item := range items {
		if item == nil {
			return fmt.Errorf("sdk/search: nil indexable item at index %d", i)
		}
		id := item.SearchID()
		if id == "" {
			return fmt.Errorf("sdk/search: empty search ID at index %d", i)
		}

		copiedMeta := make(map[string]string)
		for k, v := range item.SearchMetadata() {
			copiedMeta[k] = v
		}

		s.items[id] = &internalItem{
			id:       id,
			content:  item.SearchContent(),
			metadata: copiedMeta,
		}
	}

	return nil
}

// Search scans all indexed items and returns matches satisfying options.
func (s *InMemorySearchEngine) Search(
	ctx context.Context,
	query string,
	opts ...contractssearch.SearchOption,
) ([]contractssearch.Result, error) {
	if !s.IsStarted() {
		return nil, fmt.Errorf("sdk/search: search engine %q is not running", s.Name())
	}

	cfg := contractssearch.ApplySearchOptions(opts...)

	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []contractssearch.Result
	lowerQuery := strings.ToLower(query)

	for _, item := range s.items {
		if !matchFilters(item.metadata, cfg.Filters) {
			continue
		}

		score := 0.0
		lowerContent := strings.ToLower(item.content)

		if query != "" {
			if lowerContent == lowerQuery {
				score = 1.0
			} else if strings.HasPrefix(lowerContent, lowerQuery) {
				score = 0.8
			} else if strings.Contains(lowerContent, lowerQuery) {
				score = 0.5
			} else if strings.Contains(strings.ToLower(item.id), lowerQuery) {
				score = 0.3
			}
		} else {
			score = 0.1
		}

		if score >= cfg.MinScore && score > 0.0 {
			copiedMeta := make(map[string]string)
			for k, v := range item.metadata {
				copiedMeta[k] = v
			}

			results = append(results, contractssearch.Result{
				ID:       item.id,
				Content:  item.content,
				Score:    score,
				Metadata: copiedMeta,
			})
		}
	}

	s.sortResults(results)

	if cfg.MaxResults > 0 && len(results) > cfg.MaxResults {
		results = results[:cfg.MaxResults]
	}

	return results, nil
}

func matchFilters(itemMeta, filters map[string]string) bool {
	for k, filterVal := range filters {
		val, ok := itemMeta[k]
		if !ok || val != filterVal {
			return false
		}
	}
	return true
}

// ponytail: simple O(N^2) bubble sort. Use sort.Slice or sort.SliceStable if result size grows large.
func (s *InMemorySearchEngine) sortResults(results []contractssearch.Result) {
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			shouldSwap := false
			if results[i].Score < results[j].Score {
				shouldSwap = true
			} else if results[i].Score == results[j].Score {
				if results[i].ID > results[j].ID {
					shouldSwap = true
				}
			}
			if shouldSwap {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}
