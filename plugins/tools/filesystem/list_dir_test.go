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

func TestListDirTool_Execute(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup directories and files
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	testFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	subFile := filepath.Join(subDir, "subfile.txt")
	if err := os.WriteFile(subFile, []byte("world"), 0644); err != nil {
		t.Fatalf("failed to create subfile: %v", err)
	}

	parentDir := filepath.Dir(tmpDir)
	outsideDir := filepath.Join(parentDir, "outside_dir")
	if err := os.Mkdir(outsideDir, 0755); err != nil {
		t.Fatalf("failed to create outside dir: %v", err)
	}
	defer os.RemoveAll(outsideDir)

	tool, err := filesystem.NewListDirTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to construct ListDirTool: %v", err)
	}

	tests := []struct {
		name                     string
		args                     string
		expectErrMessageContains string
		expectCount              int
	}{
		{
			name:        "list workspace root empty json",
			args:        `{}`,
			expectCount: 2, // subdir and file.txt
		},
		{
			name:        "list workspace root null path",
			args:        `{"path": ""}`,
			expectCount: 2,
		},
		{
			name:        "list subdir relative",
			args:        `{"path": "subdir"}`,
			expectCount: 1, // subfile.txt
		},
		{
			name:                     "fail path traversal",
			args:                     `{"path": "../outside_dir"}`,
			expectErrMessageContains: "path traversal attempt",
		},
		{
			name:                     "fail target is file",
			args:                     `{"path": "file.txt"}`,
			expectErrMessageContains: "is a file, not a directory",
		},
		{
			name:                     "fail not found",
			args:                     `{"path": "nonexistent"}`,
			expectErrMessageContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tool.Execute(context.Background(), json.RawMessage(tt.args))
			if err != nil {
				t.Fatalf("unexpected execution error: %v", err)
			}
			if tt.expectErrMessageContains != "" {
				if res.ExitCode != 1 {
					t.Errorf("expected failure (exit code 1), got %d", res.ExitCode)
				}
				if !strings.Contains(res.Error, tt.expectErrMessageContains) {
					t.Errorf("expected error %q to contain %q", res.Error, tt.expectErrMessageContains)
				}
			} else {
				if res.ExitCode != 0 {
					t.Fatalf("expected success, got exit code %d, error: %s", res.ExitCode, res.Error)
				}
				var list []struct {
					Name  string `json:"name"`
					IsDir bool   `json:"is_dir"`
					Size  int64  `json:"size_bytes"`
				}
				if err := json.Unmarshal([]byte(res.Output), &list); err != nil {
					t.Fatalf("failed to unmarshal output: %v", err)
				}
				if len(list) != tt.expectCount {
					t.Errorf("expected %d entries, got %d", tt.expectCount, len(list))
				}
			}
		})
	}
}
