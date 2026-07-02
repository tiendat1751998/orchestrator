# Micro-Task 3.14: Create sdk/context/builder.go

## Info
- **File**: `sdk/context/builder.go`
- **Package**: `sdkcontext`
- **Depends on**: 1.28 (context.go contract)
- **Time**: 15 min
- **Verify**: `go build ./sdk/context/...`

## Purpose
Triển khai bộ gộp ngữ cảnh (`TruncateItems`) và đo đếm tokens. Tệp helper này giúp lọc và cắt bớt dữ liệu ngữ cảnh (Context Items) dựa trên độ ưu tiên (priority) và giới hạn ngân sách tokens (max tokens budget) của mô hình AI, giúp tránh lỗi tràn cửa sổ ngữ cảnh (Context Window Overflow).

## EXACT code to create

```go
// Package sdkcontext provides helper implementations for managing agent execution context.
package sdkcontext

import (
	"sort"

	"github.com/tiendat1751998/orchestrator/contracts/context"
)

// EstimateTokens calculates a rough approximation of token count based on string length.
// Standard rule: 1 token ≈ 4 characters for English text.
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}
	// Round up to nearest token
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
			// Skip this item as it would overflow token budget
			continue
		}

		// Update actual tokens estimate in item
		item.Tokens = itemTokens
		accepted = append(accepted, item)
		accumulatedTokens += itemTokens
	}

	return accepted
}
```

## ⚠️ Pitfalls

### Pitfall 1: Unstable Sorting of Equal Priority Items
```go
// ❌ WRONG:
sort.Slice(working, func(i, j int) bool {
    return working[i].Priority > working[j].Priority // Unstable sort might shuffle original order of files/logs
})

// ✅ CORRECT:
sort.SliceStable(working, func(i, j int) bool {
    return working[i].Priority > working[j].Priority // Preserves relative order
})
```
If two context items (e.g., consecutive lines of file updates) have the same priority, shuffling them randomly will confuse the AI. Always use `sort.SliceStable`.

### Pitfall 2: Mutating the input slice parameters
Sorting a slice in Go changes the index ordering in place. If the builder modifies the incoming slice, it creates unexpected bugs for other callers who reuse the list. Always allocate and copy the slice using `copy()`.

## Verify
```bash
go build ./sdk/context/...
```

## Checklist
- [ ] File `sdk/context/builder.go` exists
- [ ] Package: `sdkcontext`
- [ ] `EstimateTokens` provides character-based token approximation
- [ ] `TruncateItems` creates a copied slice to avoid mutating inputs
- [ ] Sorting is done via `sort.SliceStable` to preserve relative order of same priority items
- [ ] Iterative token accumulator drops elements exceeding `maxTokens` budget
- [ ] `go build ./sdk/context/...` passes
