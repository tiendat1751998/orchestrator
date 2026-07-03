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
//  1. Items are sorted by Priority in descending order (highest priority first).
//  2. If priority is equal, items are kept in their original sequence.
//  3. Items are accumulated until adding the next item would exceed maxTokens.
//  4. The remaining items are discarded.
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
