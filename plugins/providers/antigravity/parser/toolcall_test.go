package parser

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseToolCalls(t *testing.T) {
	// Mock timeNowUnixNano for predictable tool call IDs
	oldTimeNow := timeNowUnixNano
	defer func() { timeNowUnixNano = oldTimeNow }()
	timeNowUnixNano = func() int64 {
		return 123456789
	}

	tests := []struct {
		name          string
		input         string
		expectedCalls int
		expectedNames []string
		expectedArgs  []string
		expectError   bool
	}{
		{
			name:          "empty input",
			input:         "",
			expectedCalls: 0,
			expectError:   false,
		},
		{
			name: "single tool call in json block",
			input: `Here is the tool call:
` + "```json" + `
{
  "tool": "read_file",
  "args": {"path": "/etc/passwd"}
}
` + "```" + `
Hope this helps!`,
			expectedCalls: 1,
			expectedNames: []string{"read_file"},
			expectedArgs:  []string{`{"path": "/etc/passwd"}`},
			expectError:   false,
		},
		{
			name: "multiple tool calls in separate json blocks",
			input: `First call:
` + "```json" + `
{
  "tool": "read_file",
  "args": {"path": "/etc/passwd"}
}
` + "```" + `
Second call:
` + "```json" + `
[
  {"tool": "write_file", "args": {"path": "/tmp/test", "content": "hello"}}
]
` + "```",
			expectedCalls: 2,
			expectedNames: []string{"read_file", "write_file"},
			expectedArgs:  []string{`{"path": "/etc/passwd"}`, `{"path": "/tmp/test", "content": "hello"}`},
			expectError:   false,
		},
		{
			name:          "raw JSON single tool call fallback",
			input:         `{"tool": "list_dir", "args": {"dir": "."}}`,
			expectedCalls: 1,
			expectedNames: []string{"list_dir"},
			expectedArgs:  []string{`{"dir": "."}`},
			expectError:   false,
		},
		{
			name:          "raw JSON array tool calls fallback",
			input:         `[{"tool": "list_dir", "args": {"dir": "."}}, {"tool": "get_env", "args": {}}]`,
			expectedCalls: 2,
			expectedNames: []string{"list_dir", "get_env"},
			expectedArgs:  []string{`{"dir": "."}`, `{}`},
			expectError:   false,
		},
		{
			name:          "raw text that is not JSON at all",
			input:         `Hello, this is just plain text.`,
			expectedCalls: 0,
			expectError:   false,
		},
		{
			name: "malformed JSON in block fallback to nothing",
			input: `
` + "```json" + `
{
  "tool": "read_file",
  "args": {"path":
` + "```",
			expectedCalls: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseToolCalls(tt.input)
			if (err != nil) != tt.expectError {
				t.Fatalf("ParseToolCalls() error = %v, expectError %v", err, tt.expectError)
			}
			if len(got) != tt.expectedCalls {
				t.Fatalf("expected %d tool calls, got %d", tt.expectedCalls, len(got))
			}
			for i, call := range got {
				expectedID := "call_" + tt.expectedNames[i] + "_123456789"
				if call.ID != expectedID {
					t.Errorf("expected call ID %q, got %q", expectedID, call.ID)
				}
				if call.Name != tt.expectedNames[i] {
					t.Errorf("expected name %q, got %q", tt.expectedNames[i], call.Name)
				}
				// Compare compact JSON args
				var gotMap, expectedMap map[string]interface{}
				if err := json.Unmarshal(call.Args, &gotMap); err != nil {
					t.Fatalf("failed to unmarshal got args: %v", err)
				}
				if err := json.Unmarshal([]byte(tt.expectedArgs[i]), &expectedMap); err != nil {
					t.Fatalf("failed to unmarshal expected args: %v", err)
				}
				gotBytes, _ := json.Marshal(gotMap)
				expectedBytes, _ := json.Marshal(expectedMap)
				if string(gotBytes) != string(expectedBytes) {
					t.Errorf("expected args %s, got %s", string(expectedBytes), string(gotBytes))
				}
			}
		})
	}
}

func TestParseJSONText_Error(t *testing.T) {
	// Directly test parseJSONText error path
	_, err := parseJSONText("invalid json")
	if err == nil {
		t.Fatal("expected error parsing invalid json, got nil")
	}
	if !strings.Contains(err.Error(), "parser: invalid tool call format") {
		t.Errorf("expected 'parser: invalid tool call format' error, got %v", err)
	}
}
