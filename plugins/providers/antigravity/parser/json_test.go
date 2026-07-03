package parser

import (
	"testing"
)

type testStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		dest        any
		expectError bool
		expected    testStruct
	}{
		{
			name:        "basic json parsing",
			input:       `{"name": "test", "value": 123}`,
			dest:        &testStruct{},
			expectError: false,
			expected:    testStruct{Name: "test", Value: 123},
		},
		{
			name: "strips markdown fences",
			input: "```json\n" +
				`{"name": "fenced", "value": 456}` +
				"\n```",
			dest:        &testStruct{},
			expectError: false,
			expected:    testStruct{Name: "fenced", Value: 456},
		},
		{
			name: "strips untagged markdown fences",
			input: "```\n" +
				`{"name": "untagged", "value": 789}` +
				"\n```",
			dest:        &testStruct{},
			expectError: false,
			expected:    testStruct{Name: "untagged", Value: 789},
		},
		{
			name:        "nil destination pointer error",
			input:       `{"name": "test"}`,
			dest:        nil,
			expectError: true,
		},
		{
			name:        "empty JSON payload",
			input:       "   ",
			dest:        &testStruct{},
			expectError: true,
		},
		{
			name:        "invalid JSON syntax",
			input:       `{"name": "test",`,
			dest:        &testStruct{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.dest != nil {
				err = ParseJSON(tt.input, tt.dest)
				if !tt.expectError {
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					got := *(tt.dest.(*testStruct))
					if got != tt.expected {
						t.Errorf("expected %+v, got %+v", tt.expected, got)
					}
				} else {
					if err == nil {
						t.Error("expected error but got nil")
					}
				}
			} else {
				err = ParseJSON(tt.input, nil)
				if err == nil {
					t.Error("expected error for nil dest, got nil")
				}
			}
		})
	}
}
