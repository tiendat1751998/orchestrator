package filesystem_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tiendat1751998/orchestrator/plugins/tools/filesystem"
)

func TestSearchTool_Execute(t *testing.T) {
	// Create a temp workspace directory
	tmpDir := t.TempDir()

	// Write a sample text file
	textFilePath := filepath.Join(tmpDir, "sample.txt")
	textContent := "hello world\nline two\nhello there\n"
	if err := os.WriteFile(textFilePath, []byte(textContent), 0644); err != nil {
		t.Fatalf("failed to create sample text file: %v", err)
	}

	// Write a file in an excluded directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}
	excludedFilePath := filepath.Join(gitDir, "exclude.txt")
	if err := os.WriteFile(excludedFilePath, []byte("hello hidden"), 0644); err != nil {
		t.Fatalf("failed to write excluded file: %v", err)
	}

	// Write a binary file
	binaryFilePath := filepath.Join(tmpDir, "binary.bin")
	binaryContent := []byte{0x01, 0x02, 0x00, 0x03, 'h', 'e', 'l', 'l', 'o'}
	if err := os.WriteFile(binaryFilePath, binaryContent, 0644); err != nil {
		t.Fatalf("failed to create binary file: %v", err)
	}

	// Write a large text file (> 2MB)
	largeFilePath := filepath.Join(tmpDir, "large.txt")
	largeFileContent := strings.Repeat("hello\n", 500000) // approx 3MB
	if err := os.WriteFile(largeFilePath, []byte(largeFileContent), 0644); err != nil {
		t.Fatalf("failed to create large file: %v", err)
	}

	// Create 100 small files to test results limit (max 50 matches)
	limitDir := filepath.Join(tmpDir, "limit")
	if err := os.Mkdir(limitDir, 0755); err != nil {
		t.Fatalf("failed to create limit directory: %v", err)
	}
	for i := 0; i < 60; i++ {
		filePath := filepath.Join(limitDir, fmt.Sprintf("file_%d.txt", i))
		if err := os.WriteFile(filePath, []byte("hello limit"), 0644); err != nil {
			t.Fatalf("failed to create limit file %d: %v", i, err)
		}
	}

	// Initialize the tool
	tool, err := filesystem.NewSearchTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to create SearchTool: %v", err)
	}

	tests := []struct {
		name                     string
		args                     string
		wantErr                  bool
		wantMatchCount           int
		expectErrMessageContains string
	}{
		{
			name:           "success search query matches",
			args:           `{"query": "hello"}`,
			wantMatchCount: 50, // Capped at maxMatches (50)
		},
		{
			name:           "success search target specific query",
			args:           `{"query": "there"}`,
			wantMatchCount: 1,
		},
		{
			name:                     "fail query is empty",
			args:                     `{"query": ""}`,
			expectErrMessageContains: "query string cannot be empty",
		},
		{
			name:                     "fail invalid JSON arguments",
			args:                     `{invalid`,
			expectErrMessageContains: "invalid JSON arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tool.Execute(context.Background(), json.RawMessage(tt.args))
			if err != nil {
				t.Fatalf("unexpected error executing tool: %v", err)
			}
			if tt.expectErrMessageContains != "" {
				if res.ExitCode != 1 {
					t.Errorf("expected failure (exit code 1), got %d", res.ExitCode)
				}
				if res.Error == "" {
					t.Errorf("expected error message containing %q, got empty error", tt.expectErrMessageContains)
				} else if !strings.Contains(res.Error, tt.expectErrMessageContains) {
					t.Errorf("expected error message containing %q, got %q", tt.expectErrMessageContains, res.Error)
				}
			} else {
				if res.ExitCode != 0 {
					t.Errorf("expected success (exit code 0), got %d (error: %q)", res.ExitCode, res.Error)
				}

				// Unmarshal output matches
				var matches []filesystem.MatchInfo
				if err := json.Unmarshal([]byte(res.Output), &matches); err != nil {
					t.Fatalf("failed to unmarshal output matches: %v", err)
				}

				if len(matches) != tt.wantMatchCount {
					t.Errorf("got %d matches, want %d", len(matches), tt.wantMatchCount)
				}

				// Ensure no matches from .git directory
				for _, match := range matches {
					if strings.Contains(match.File, ".git") {
						t.Errorf("found unexpected match in excluded directory: %s", match.File)
					}
					if strings.Contains(match.File, "large.txt") {
						t.Errorf("found unexpected match in large file: %s", match.File)
					}
					if strings.Contains(match.File, "binary.bin") {
						t.Errorf("found unexpected match in binary file: %s", match.File)
					}
				}
			}
		})
	}

	// Test context cancellation
	t.Run("context cancelled mid-search", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		res, err := tool.Execute(ctx, json.RawMessage(`{"query": "hello"}`))
		if err != nil {
			t.Fatalf("unexpected error executing tool with cancelled context: %v", err)
		}
		if res.ExitCode != 1 {
			t.Errorf("expected failure, got exit code %d", res.ExitCode)
		}
		if !strings.Contains(res.Error, "context canceled") {
			t.Errorf("expected error to contain 'context canceled', got %q", res.Error)
		}
	})
}
