package parser

import (
	"testing"
)

func TestParseMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic parsing",
			input:    "hello\nworld",
			expected: "hello\nworld",
		},
		{
			name:     "filters delimiters",
			input:    "hello\n---END---\nworld\n  ---END---  ",
			expected: "hello\nworld",
		},
		{
			name:     "filters fences",
			input:    "hello\n```go\nfmt.Println(\"test\")\n```\nworld",
			expected: "hello\nfmt.Println(\"test\")\nworld",
		},
		{
			name:     "preserves inner space and trims outer",
			input:    "  \nhello\n  world  \n  ",
			expected: "hello\n  world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMarkdown(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("expected:\n%q\ngot:\n%q", tt.expected, got)
			}
		})
	}
}

func TestExtractCodeBlock(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		lang     string
		expected string
	}{
		{
			name: "extract go block",
			input: `Some text
` + "```go" + `
package main

func main() {
	// indentation
}
` + "```" + `
Some other text`,
			lang: "go",
			expected: `package main

func main() {
	// indentation
}`,
		},
		{
			name: "case insensitive lang",
			input: `Some text
` + "```Go" + `
package main
` + "```",
			lang:     "go",
			expected: "package main",
		},
		{
			name: "lang not found",
			input: `Some text
` + "```go" + `
package main
` + "```",
			lang:     "python",
			expected: "",
		},
		{
			name: "empty lang match empty block",
			input: `Some text
` + "```" + `
untagged block
` + "```",
			lang:     "",
			expected: "untagged block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractCodeBlock(tt.input, tt.lang)
			if got != tt.expected {
				t.Errorf("expected:\n%q\ngot:\n%q", tt.expected, got)
			}
		})
	}
}
