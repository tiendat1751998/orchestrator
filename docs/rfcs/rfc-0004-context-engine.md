# RFC-0004: Context Engine

- **Status**: PROPOSED
- **Priority**: P1 — Core
- **Author**: Orchestrator Architecture Team
- **Created**: 2026-07-03
- **Depends on**: RFC-0001 (Kernel Architecture), RFC-0002 (Brain Architecture)

## Summary

The Context Engine is the core thinking pipeline helper in AEOS. It is responsible for gathering, ranking, compressing, assembling, caching, and managing token windows for all inputs sent to AI providers. AI models are highly sensitive to context quality, token count, and context noise. The Context Engine ensures AI providers receive high-value context that fits precisely within their token window limits.

## Motivation

Without a dedicated Context Engine:
- **Token Overflow**: Sending too much raw text causes AI providers to fail with token limit errors.
- **Context Noise**: Including irrelevant logs or files reduces the reasoning quality of the AI model.
- **High Latency/Cost**: Redundant or uncompressed context increases processing time and cost.
- **Ad-hoc Assembly**: Each agent manual builds prompt context differently, leading to code duplication and inconsistent formats.

## Design

The Context Engine operates as a pipeline inside the Brain Runtime:

```
  Context Request (Task Type, Description, Max Tokens, Sources)
                     │
                     ▼
  ┌────────────────────────────────────────────────────────┐
  │ 1. Builder       : Gather raw items from sources       │
  └──────────────────┬─────────────────────────────────────┘
                     │ Raw ContextItems
                     ▼
  ┌────────────────────────────────────────────────────────┐
  │ 2. Ranker        : Score & sort items by relevance     │
  └──────────────────┬─────────────────────────────────────┘
                     │ Scored ContextItems
                     ▼
  ┌────────────────────────────────────────────────────────┐
  │ 3. Filter/Window : Drop low score, enforce limits      │
  └──────────────────┬─────────────────────────────────────┘
                     │ Filtered ContextItems
                     ▼
  ┌────────────────────────────────────────────────────────┐
  │ 4. Compressor/   : Shorten code/text, summarize logs   │
  │    Summarizer    │                                     │
  └──────────────────┬─────────────────────────────────────┘
                     │ Compressed ContextItems
                     ▼
  ┌────────────────────────────────────────────────────────┐
  │ 5. Assembler     : Merge into final prompt context     │
  └──────────────────┬─────────────────────────────────────┘
                     │ AssembledContext
                     ▼
  ┌────────────────────────────────────────────────────────┐
  │ 6. Cache         : Deduplicate & reuse common context  │
  └────────────────────────────────────────────────────────┘
```

### Contracts

```go
// contracts/brain/context.go — Expanded Context Engine contracts
package brain

import (
	"context"
	"time"
)

// ContextSourceType identifies the origin of a context item.
type ContextSourceType string

const (
	SourceCode      ContextSourceType = "code"      // Source files
	SourceDoc       ContextSourceType = "doc"       // Documentation files
	SourceKnowledge ContextSourceType = "knowledge" // Knowledge graph nodes/patterns
	SourceHistory   ContextSourceType = "history"   // Timeline events & past decisions
	SourceArtifact  ContextSourceType = "artifact"  // Generated file outputs
	SourceMemory    ContextSourceType = "memory"    // Mission working memory
)

// ContextSource defines query details to gather context from a specific source.
type ContextSource struct {
	Type     ContextSourceType `json:"type"`
	Filter   string            `json:"filter,omitempty"` // e.g. glob "*.go", graph query, tag list
	MaxItems int               `json:"max_items,omitempty"`
}

// ContextRequest defines parameters for context assembly.
type ContextRequest struct {
	TaskType        string          `json:"task_type"`
	TaskDescription string          `json:"task_description,omitempty"`
	MaxTokens       int             `json:"max_tokens"`
	Sources         []ContextSource `json:"sources"`
	Priority        []ContextSourceType `json:"priority,omitempty"`
}

// ContextItem represents a single piece of context.
type ContextItem struct {
	ID        string            `json:"id"`
	Source    ContextSourceType `json:"source"`
	Reference string            `json:"reference,omitempty"` // File path, Node ID, etc.
	Content   string            `json:"content"`
	Tokens    int               `json:"tokens"` // Calculated token count
	Score     float64           `json:"score"`  // Relevance score (0.0 to 1.0)
	UpdatedAt time.Time         `json:"updated_at"`
}

// AssembledContext is the output of the context engine pipeline.
type AssembledContext struct {
	Items          []ContextItem             `json:"items"`
	TotalTokens    int                       `json:"total_tokens"`
	Truncated      bool                      `json:"truncated"`
	SourcesSummary map[ContextSourceType]int `json:"sources_summary"`
	Formatted      string                    `json:"formatted"` // Combined prompt text
}

// --- Pipeline Component Interfaces ---

// ContextBuilder is responsible for gathering raw ContextItems from diverse sources.
type ContextBuilder interface {
	// Build gathers raw context items based on sources.
	// Resolves globs, graph traverses, database queries, and file reads.
	Build(ctx context.Context, sources []ContextSource) ([]ContextItem, error)
}

// ContextRanker scores and sorts context items based on task description.
type ContextRanker interface {
	// Rank evaluates relevance of items to the task and assigns scores.
	// Can use heuristic search (keyword overlap, tag match) or semantic scoring.
	Rank(ctx context.Context, taskDescription string, items []ContextItem) ([]ContextItem, error)
}

// ContextCompressor reduces size of context items while preserving information.
type ContextCompressor interface {
	// Compress shortens code or files (e.g. removing comments, extracting headers/AST,
	// or stripping unused code blocks).
	Compress(ctx context.Context, item ContextItem) (ContextItem, error)
}

// ContextSummarizer uses text summarization for long text, logs, or history.
type ContextSummarizer interface {
	// Summarize creates a brief summary of a very long context item.
	Summarize(ctx context.Context, item ContextItem, maxTokens int) (ContextItem, error)
}

// ContextWindow manages the token budget.
type ContextWindow interface {
	// EstimateTokens calculates tokens for an item (e.g., Llama/GPT tokenizers
	// or rough character-based approximation for fallback).
	EstimateTokens(text string) int
	// Fit filters and truncates items to strictly fit within the token budget.
	Fit(items []ContextItem, maxTokens int) ([]ContextItem, bool)
}

// ContextAssembler merges ranked and processed items into a final prompt representation.
type ContextAssembler interface {
	// Assemble formats and joins context items into a single coherent prompt string.
	Assemble(items []ContextItem) (string, map[ContextSourceType]int)
}

// ContextCache stores and retrieves assembled context to avoid repetitive processing.
type ContextCache interface {
	Get(ctx context.Context, requestHash string) (*AssembledContext, bool)
	Set(ctx context.Context, requestHash string, context *AssembledContext, ttl time.Duration) error
}

// ContextEngine orchestrates the entire context pipeline.
type ContextEngine interface {
	// Assemble builds context from multiple sources, executing the entire pipeline.
	Assemble(ctx context.Context, req ContextRequest) (*AssembledContext, error)
}
```

### Pipeline Execution Flow

When `ContextEngine.Assemble()` is called:

1. **Hash Generation**: Hash the `ContextRequest` to check the **ContextCache**.
2. **Build**: Invoke **ContextBuilder** to query the filesystems, knowledge graph, and working memory to pull raw context.
3. **Token Estimation**: Use **ContextWindow** to estimate the raw sizes.
4. **Rank**: Invoke **ContextRanker** to assign scores based on the current `TaskDescription`. Items are sorted in descending order of their relevance score.
5. **Compress & Summarize**:
   - If total tokens exceed `MaxTokens`, apply **ContextCompressor** on code files (e.g., AST/signature extraction) and **ContextSummarizer** on logs and history.
6. **Window Fitting**: Use **ContextWindow.Fit** to filter out low-scoring items and truncate to fit the absolute token budget.
7. **Assemble**: Send the remaining items to **ContextAssembler** to construct the final prompt representation.
8. **Cache Store**: Store the resulting `AssembledContext` in the cache.

## Impact

### New Contract Packages
- Expose all component interfaces within `contracts/brain/` (flat structure, separate files like `context.go`).

### New Kernel Implementation Packages
- `kernel/brain/context/` containing:
  - `builder.go`
  - `ranker.go`
  - `compressor.go`
  - `assembler.go`
  - `cache.go`
  - `window.go`
  - `summarizer.go`
  - `engine.go` (Pipeline coordinator)

## Open Questions

1. **How detailed should the code compressor be?**
   - For Go code, we can parse the AST and strip implementation bodies (leaving only method/struct signatures) when space is tight. This will be implemented incrementally starting with regex-based comment/space stripping.
2. **Token counting accuracy:**
   - Go doesn't have native official tokenizers for Claude/Gemini. We will implement a character-based heuristic token estimator initially (approx. 4 characters = 1 token), and allow plugins to override it with accurate tokenizers if needed.
