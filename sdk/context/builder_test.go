package sdkcontext

import (
	"reflect"
	"testing"

	agentcontext "github.com/tiendat1751998/orchestrator/contracts/context"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "one character",
			input:    "a",
			expected: 1,
		},
		{
			name:     "two characters",
			input:    "ab",
			expected: 1,
		},
		{
			name:     "three characters",
			input:    "abc",
			expected: 1,
		},
		{
			name:     "four characters",
			input:    "abcd",
			expected: 1,
		},
		{
			name:     "five characters",
			input:    "abcde",
			expected: 2,
		},
		{
			name:     "eight characters",
			input:    "abcdefgh",
			expected: 2,
		},
		{
			name:     "english sentence example",
			input:    "hello world!", // 12 chars -> 12 + 3 / 4 = 3 tokens
			expected: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := EstimateTokens(tc.input)
			if got != tc.expected {
				t.Errorf("EstimateTokens(%q) = %d; expected %d", tc.input, got, tc.expected)
			}
		})
	}
}

func TestTruncateItems(t *testing.T) {
	t.Run("empty or invalid maxTokens inputs", func(t *testing.T) {
		items := []agentcontext.Item{
			{Content: "test", Priority: 1},
		}

		if got := TruncateItems(nil, 100); got != nil {
			t.Errorf("expected nil for nil items slice, got %v", got)
		}
		if got := TruncateItems([]agentcontext.Item{}, 100); got != nil {
			t.Errorf("expected nil for empty items slice, got %v", got)
		}
		if got := TruncateItems(items, 0); got != nil {
			t.Errorf("expected nil for maxTokens <= 0, got %v", got)
		}
		if got := TruncateItems(items, -5); got != nil {
			t.Errorf("expected nil for negative maxTokens, got %v", got)
		}
	})

	t.Run("priority ordering and stable sort for equal priority", func(t *testing.T) {
		items := []agentcontext.Item{
			{Content: "A", Priority: 2, Source: "orig1"},
			{Content: "B", Priority: 5, Source: "orig2"},
			{Content: "C", Priority: 2, Source: "orig3"},
			{Content: "D", Priority: 1, Source: "orig4"},
			{Content: "E", Priority: 5, Source: "orig5"},
		}

		// Let's use large enough maxTokens to fit all items.
		// Content lengths: A(1 token), B(1 token), C(1 token), D(1 token), E(1 token) -> total 5 tokens.
		got := TruncateItems(items, 100)

		// Expected order:
		// 1. Highest Priority (5): B ("orig2") and E ("orig5"). Original sequence: B first, E second.
		// 2. Next Priority (2): A ("orig1") and C ("orig3"). Original sequence: A first, C second.
		// 3. Lowest Priority (1): D ("orig4").
		expected := []agentcontext.Item{
			{Content: "B", Priority: 5, Source: "orig2", Tokens: 1},
			{Content: "E", Priority: 5, Source: "orig5", Tokens: 1},
			{Content: "A", Priority: 2, Source: "orig1", Tokens: 1},
			{Content: "C", Priority: 2, Source: "orig3", Tokens: 1},
			{Content: "D", Priority: 1, Source: "orig4", Tokens: 1},
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("TruncateItems mismatch\ngot:  %+v\nwant: %+v", got, expected)
		}
	})

	t.Run("budget truncation", func(t *testing.T) {
		// Tokens specified directly:
		// item1: priority 10, tokens 10
		// item2: priority 8, tokens 20
		// item3: priority 6, tokens 5
		// item4: priority 4, tokens 15
		items := []agentcontext.Item{
			{Content: "item3", Priority: 6, Tokens: 5},
			{Content: "item1", Priority: 10, Tokens: 10},
			{Content: "item4", Priority: 4, Tokens: 15},
			{Content: "item2", Priority: 8, Tokens: 20},
		}

		// Sort order by priority:
		// 1. item1: Priority 10, Tokens 10
		// 2. item2: Priority 8, Tokens 20
		// 3. item3: Priority 6, Tokens 5
		// 4. item4: Priority 4, Tokens 15

		// If maxTokens = 15:
		// item1 (10 tokens) fits. Remaining budget = 5.
		// item2 (20 tokens) does not fit. Remaining budget = 5.
		// item3 (5 tokens) fits. Remaining budget = 0.
		// item4 (15 tokens) does not fit.
		// Expected: item1 and item3.
		got := TruncateItems(items, 15)
		expected := []agentcontext.Item{
			{Content: "item1", Priority: 10, Tokens: 10},
			{Content: "item3", Priority: 6, Tokens: 5},
		}

		if !reflect.DeepEqual(got, expected) {
			t.Errorf("TruncateItems budget truncation mismatch\ngot:  %+v\nwant: %+v", got, expected)
		}
	})

	t.Run("no mutation of input slice", func(t *testing.T) {
		original := []agentcontext.Item{
			{Content: "A", Priority: 1, Tokens: 0},
			{Content: "B", Priority: 5, Tokens: 0},
		}

		// Make a copy of the original slice to verify nothing in the underlying slice is mutated
		originalCopy := make([]agentcontext.Item, len(original))
		copy(originalCopy, original)

		_ = TruncateItems(original, 100)

		if !reflect.DeepEqual(original, originalCopy) {
			t.Errorf("TruncateItems mutated the input slice!\ngot:  %+v\nwant: %+v", original, originalCopy)
		}
	})
}
