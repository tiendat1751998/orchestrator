# Micro-Task 1.28: Create contracts/context/context.go

## Info
- **File**: `contracts/context/context.go`
- **Package**: `agentcontext`
- **Depends on**: 1.06
- **Time**: 10 min
- **Verify**: `go build ./contracts/context/...`

## Purpose
Defines the `Builder` and `Item` interfaces for managing agent prompt context window allocations without colliding with the Go standard `context` library.

## EXACT code to create

```go
// Package agentcontext defines the contract for context window management.
// The context builder decides what information to include in the AI prompt.
//
// WHY the package name is "agentcontext" and NOT "context"?
// → Go standard library already has "context" package.
// → If this package were named "context", importing both would require aliasing:
//     import (
//         stdctx "context"
//         "github.com/tiendat1751998/orchestrator/contracts/context"
//     )
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
```

## Rules
1. **Package Namespace Independence**: The package name must be declared as `agentcontext` to prevent conflicts with standard library imports, while residing in the `contracts/context/` directory.
2. **Prioritization Truncation**: When selecting items under token constraints, lower priority context items must be dropped first.
3. **Estimation Standard**: Approximate token sizes using the standard ratio (e.g. 1 token ≈ 4 bytes/characters for English text).

## ⚠️ Pitfalls

### Pitfall 1: Naming the package `context` inside `contracts/context/`
```go
package agentcontext
```
Keep the directory name as `context` but use `package agentcontext` at the top of Go source files.

### Pitfall 2: Overlooking the MaxTokens budget when building prompt files
If the context builder appends files blindly without checking token counts, the final prompt can exceed the model's maximum allowed context window limit, causing model providers to reject the API call with HTTP 400 error.

## Verify
```bash
go build ./contracts/context/...
```

## Checklist
- [ ] File `contracts/context/context.go` exists
- [ ] Package: `agentcontext`
- [ ] `Builder` interface contains a `Build` method
- [ ] `Item` struct contains Type, Content, Source, Priority, and Tokens fields
- [ ] Functional options `WithMaxTokens`, `WithSources`, and `WithQuery` exist
- [ ] `ApplyBuildOptions` defaults `MaxTokens` to 8192
- [ ] `go build ./contracts/context/...` passes
