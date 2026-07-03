package filesystem_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tiendat1751998/orchestrator/plugins/tools/filesystem"
)

func TestWriteFileTool_Execute(t *testing.T) {
	// Create a temp workspace directory
	tmpDir := t.TempDir()

	// Create a file outside workspace for absolute path traversal check
	parentDir := filepath.Dir(tmpDir)
	outsidePath := filepath.Join(parentDir, "outside_write_test.txt")

	// Initialize the tool
	tool, err := filesystem.NewWriteFileTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to create WriteFileTool: %v", err)
	}

	tests := []struct {
		name                     string
		args                     string
		wantContent              string
		wantPath                 string
		wantErr                  bool
		expectErrMessageContains string
	}{
		{
			name:        "success relative path",
			args:        `{"path": "test_relative.txt", "content": "hello relative"}`,
			wantPath:    "test_relative.txt",
			wantContent: "hello relative",
		},
		{
			name:        "success absolute path inside workspace",
			args:        `{"path": "` + filepath.ToSlash(filepath.Join(tmpDir, "test_absolute.txt")) + `", "content": "hello absolute"}`,
			wantPath:    "test_absolute.txt",
			wantContent: "hello absolute",
		},
		{
			name:        "success create directories and write",
			args:        `{"path": "subdir/nested/test_nested.txt", "content": "hello nested"}`,
			wantPath:    "subdir/nested/test_nested.txt",
			wantContent: "hello nested",
		},
		{
			name:                     "fail path traversal relative",
			args:                     `{"path": "../outside.txt", "content": "should fail"}`,
			expectErrMessageContains: "path traversal attempt",
		},
		{
			name:                     "fail path traversal absolute",
			args:                     `{"path": "` + filepath.ToSlash(outsidePath) + `", "content": "should fail"}`,
			expectErrMessageContains: "path traversal attempt",
		},
		{
			name:                     "fail path traversal workspace folder itself",
			args:                     `{"path": ".", "content": "should fail"}`,
			expectErrMessageContains: "path traversal attempt",
		},
		{
			name:                     "fail missing required field path",
			args:                     `{"content": "hello"}`,
			expectErrMessageContains: "missing required parameter",
		},
		{
			name:                     "fail missing required field content",
			args:                     `{"path": "test.txt"}`,
			expectErrMessageContains: "missing required parameter",
		},
		{
			name:                     "fail invalid json arguments",
			args:                     `{invalid json}`,
			expectErrMessageContains: "invalid character",
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
				// Verify file content
				filePath := filepath.Join(tmpDir, tt.wantPath)
				contentBytes, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("failed to read expected file: %v", err)
				}
				if string(contentBytes) != tt.wantContent {
					t.Errorf("got file content:\n%q\nwant:\n%q", string(contentBytes), tt.wantContent)
				}
			}
		})
	}
}
