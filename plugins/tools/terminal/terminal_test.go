package terminal_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tiendat1751998/orchestrator/plugins/tools/terminal"
)

func TestRunCommandTool_Execute(t *testing.T) {
	tmpDir := t.TempDir()

	tool, err := terminal.NewRunCommandTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to construct RunCommandTool: %v", err)
	}

	tests := []struct {
		name                     string
		args                     string
		expectErrMessageContains string
		expectOutputContains     string
	}{
		{
			name:                 "echo hello",
			args:                 `{"command": "echo hello_world"}`,
			expectOutputContains: "hello_world",
		},
		{
			name:                     "empty command",
			args:                     `{"command": ""}`,
			expectErrMessageContains: "empty command query",
		},
		{
			name:                     "blocked command rm -rf",
			args:                     `{"command": "rm -rf /some/path"}`,
			expectErrMessageContains: "security error: command contains blocked token \"rm -rf\"",
		},
		{
			name:                     "blocked command del /",
			args:                     `{"command": "del /s /q c:\\"}`,
			expectErrMessageContains: "security error: command contains blocked token \"del /\"",
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
				if !strings.Contains(res.Output, tt.expectOutputContains) {
					t.Errorf("expected output %q to contain %q", res.Output, tt.expectOutputContains)
				}
			}
		})
	}
}
