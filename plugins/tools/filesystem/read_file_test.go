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

func TestReadFileTool_Execute(t *testing.T) {
	// Create a temp workspace directory
	tmpDir := t.TempDir()

	// Write a sample text file
	textFilePath := filepath.Join(tmpDir, "sample.txt")
	textContent := "line 1\nline 2\nline 3\nline 4\nline 5\n"
	if err := os.WriteFile(textFilePath, []byte(textContent), 0644); err != nil {
		t.Fatalf("failed to create sample text file: %v", err)
	}

	// Write a binary file
	binaryFilePath := filepath.Join(tmpDir, "binary.bin")
	binaryContent := []byte{0x01, 0x02, 0x00, 0x03}
	if err := os.WriteFile(binaryFilePath, binaryContent, 0644); err != nil {
		t.Fatalf("failed to create binary file: %v", err)
	}

	// Create a file outside workspace for absolute path traversal check
	parentDir := filepath.Dir(tmpDir)
	outsidePath := filepath.Join(parentDir, "outside_sample_test.txt")
	if err := os.WriteFile(outsidePath, []byte("secret"), 0644); err != nil {
		t.Fatalf("failed to create outside file: %v", err)
	}
	defer os.Remove(outsidePath)

	// Create a subdirectory inside workspace to test directory rejection
	subDir := filepath.Join(tmpDir, "sub_dir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create sub directory: %v", err)
	}

	// Initialize the tool
	tool, err := filesystem.NewReadFileTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to create ReadFileTool: %v", err)
	}

	tests := []struct {
		name                     string
		args                     string
		wantErr                  bool
		wantOutput               string
		expectErrMessageContains string
	}{
		{
			name:       "success whole file relative path",
			args:       `{"path": "sample.txt"}`,
			wantOutput: textContent,
		},
		{
			name:       "success whole file absolute path",
			args:       `{"path": "` + filepath.ToSlash(textFilePath) + `"}`,
			wantOutput: textContent,
		},
		{
			name:       "success line range subset",
			args:       `{"path": "sample.txt", "start_line": 2, "end_line": 4}`,
			wantOutput: "line 2\nline 3\nline 4\n",
		},
		{
			name:       "success start line only",
			args:       `{"path": "sample.txt", "start_line": 4}`,
			wantOutput: "line 4\nline 5\n",
		},
		{
			name:       "success end line only",
			args:       `{"path": "sample.txt", "end_line": 2}`,
			wantOutput: "line 1\nline 2\n",
		},
		{
			name:                     "fail path traversal relative",
			args:                     `{"path": "../passwd"}`,
			expectErrMessageContains: "path traversal attempt",
		},
		{
			name:                     "fail path traversal absolute",
			args:                     `{"path": "` + filepath.ToSlash(outsidePath) + `"}`,
			expectErrMessageContains: "path traversal attempt",
		},
		{
			name:                     "fail is directory",
			args:                     `{"path": "sub_dir"}`,
			expectErrMessageContains: "is a directory",
		},
		{
			name:                     "fail file not found",
			args:                     `{"path": "nonexistent.txt"}`,
			expectErrMessageContains: "not found",
		},
		{
			name:                     "fail binary file",
			args:                     `{"path": "binary.bin"}`,
			expectErrMessageContains: "cannot read binary file content",
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
				if res.Output != tt.wantOutput {
					t.Errorf("got output:\n%q\nwant:\n%q", res.Output, tt.wantOutput)
				}
			}
		})
	}
}
