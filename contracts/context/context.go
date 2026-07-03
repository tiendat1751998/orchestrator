// Package agentcontext defines the contract for context window management.
// The context builder decides what information to include in the AI prompt.
//
// WHY the package name is "agentcontext" and NOT "context"?
// → Go standard library already has "context" package.
// → If this package were named "context", importing both would require aliasing:
//
//	import (
//	    stdctx "context"
//	    "github.com/tiendat1751998/orchestrator/contracts/context"
//	)
//
// → Using "agentcontext" avoids this confusion entirely.
//
// NOTE: The directory is still "contracts/context/" for clean path structure.
// Go allows package name to differ from directory name.
package agentcontext

import "context"

// Builder constructs the context window for an AI agent.
//
// The context window is the set of information included in the AI prompt.
// It typically includes: relevant source files, previous task outputs,
// project documentation, search results, and memory entries.
//
// The builder must balance:
//   - Relevance: include the most useful information
//   - Token budget: don't exceed the model's context window size
type Builder interface {
	// Build creates a list of context items from various sources.
	//
	// The builder gathers information from:
	//   - Filesystem (relevant source files)
	//   - Search results (matching code/docs)
	//   - Memory (past task results, learned patterns)
	//   - Previous task outputs (in multi-step workflows)
	//
	// Items are sorted by Priority (highest first) and truncated
	// to fit within the token budget set by WithMaxTokens().
	Build(ctx context.Context, opts ...BuildOption) ([]Item, error)
}

// Item is a piece of context to include in the AI prompt.
type Item struct {
	// Type identifies the kind of context.
	// Values: "file", "snippet", "search_result", "memory", "task_output"
	Type string `json:"type"`

	// Content is the actual text to include in the prompt.
	Content string `json:"content"`

	// Source identifies where this content came from.
	// For "file": file path. For "memory": memory key. For "search_result": query.
	Source string `json:"source"`

	// Priority determines inclusion order.
	// Higher = more important = include first.
	// When token budget is limited, low-priority items are dropped.
	Priority int `json:"priority"`

	// Tokens is the estimated token count for this item.
	// Used by the builder to calculate remaining token budget.
	// Approximate: 1 token ≈ 4 characters (for English text).
	Tokens int `json:"tokens"`
}

// =============================================================================
// Build Options
// =============================================================================

// BuildOption configures context building.
type BuildOption func(*buildOptions)

type buildOptions struct {
	MaxTokens int
	Sources   []string
	Query     string
}

// WithMaxTokens sets the maximum total tokens for all context items.
// Items exceeding the budget are dropped (lowest priority first).
func WithMaxTokens(n int) BuildOption {
	return func(o *buildOptions) { o.MaxTokens = n }
}

// WithSources limits context to specific source types.
// Example: WithSources("file", "memory") excludes search results.
func WithSources(sources ...string) BuildOption {
	return func(o *buildOptions) { o.Sources = sources }
}

// WithQuery provides a search query to find relevant context.
func WithQuery(query string) BuildOption {
	return func(o *buildOptions) { o.Query = query }
}

// ApplyBuildOptions processes functional options.
func ApplyBuildOptions(opts ...BuildOption) buildOptions {
	o := buildOptions{
		MaxTokens: 8192, // Default: 8K tokens
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
