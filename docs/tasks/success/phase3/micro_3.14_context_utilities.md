# Micro-Task 3.14: Create sdk/context/builder.go

## Info
- **File**: `sdk/context/builder.go`
- **Package**: `sdkcontext`
- **Depends on**: 1.28 (context.go contract)
- **Time**: 15 min
- **Verify**: `go build ./sdk/context/...`

## Purpose
Implements context building utilities (`EstimateTokens` and `TruncateItems`) that filter and prioritize context files to fit payload sizes within token limits, avoiding context window overflows.

## EXACT code to create

```go
// Package sdkcontext provides helper implementations for managing agent execution context.
package sdkcontext

import (
	"sort"

	agentcontext "github.com/tiendat1751998/orchestrator/contracts/context"
)

// EstimateTokens calculates a rough approximation of token count based on string length.
// Standard rule: 1 token ≈ 4 characters for English text.
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return (len(text) + 3) / 4
}

// TruncateItems filters and drops lower priority context items
// to guarantee the total token count fits within the allowed maxTokens limit.
//
// Rules:
//   1. Items are sorted by Priority in descending order (highest priority first).
//   2. If priority is equal, items are kept in their original sequence.
//   3. Items are accumulated until adding the next item would exceed maxTokens.
//   4. The remaining items are discarded.
func TruncateItems(items []agentcontext.Item, maxTokens int) []agentcontext.Item {
	if len(items) == 0 || maxTokens <= 0 {
		return nil
	}

	// Create a copy of slice to prevent modifying caller's data
	working := make([]agentcontext.Item, len(items))
	copy(working, items)

	// Sort stable keeps original order for equal priority elements
	sort.SliceStable(working, func(i, j int) bool {
		return working[i].Priority > working[j].Priority
	})

	var accepted []agentcontext.Item
	accumulatedTokens := 0

	for _, item := range working {
		itemTokens := item.Tokens
		if itemTokens <= 0 {
			itemTokens = EstimateTokens(item.Content)
		}

		if accumulatedTokens+itemTokens > maxTokens {
			continue
		}

		item.Tokens = itemTokens
		accepted = append(accepted, item)
		accumulatedTokens += itemTokens
	}

	return accepted
}
```

## Rules
1. **Import Aliasing**: Alias the `contracts/context` package import as `agentcontext` to prevent conflicts with standard context packages.
2. **Stable Sorting**: Use `sort.SliceStable` when sorting context items by priority. This preserves chronological order when priorities match, preventing context shuffling.
3. **No Input Mutations**: Copy incoming slices before sorting to avoid side effects for callers.

## ⚠️ Pitfalls

### Pitfall 1: Unstable sorting of equal priority items
```go
```
If two file chunks are registered with identical priority, unstable sorting can shuffle their order, confusing the reasoning engine. Always use `sort.SliceStable`.

### Pitfall 2: Mutating input parameters
Sorting slices modifies indices in-place. If builders mutate the incoming slices directly, it affects other modules using the same list. Always create deep copies of slices.

## Verify
```bash
go build ./sdk/context/...
```

## Checklist
- [ ] File `sdk/context/builder.go` exists
- [ ] Package: `sdkcontext`
- [ ] Import `contracts/context` aliased to `agentcontext`
- [ ] `EstimateTokens` implements token length approximations
- [ ] `TruncateItems` creates copies of slice buffers to prevent mutations
- [ ] Stable sorting maintains original item sequences for equal priorities
- [ ] Accumulated token budgets drop overflowing items cleanly
- [ ] `go build ./sdk/context/...` passes
