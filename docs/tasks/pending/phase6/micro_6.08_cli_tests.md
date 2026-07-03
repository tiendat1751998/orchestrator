# Micro-Task 6.08: Create cmd/orchestrator/cmd/cmd_test.go

## Info
- **File**: `cmd/orchestrator/cmd/cmd_test.go`
- **Package**: `cmd_test`
- **Depends on**: 6.01-6.06
- **Time**: 20 min
- **Verify**: `go test ./cmd/orchestrator/cmd/...`

## Purpose
Unit tests for CLI command parsing, flag defaults, context helpers, and error handling.

## EXACT code to create

```go
package cmd_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/tiendat1751998/orchestrator/cmd/orchestrator/cmd"
)

func TestConfigPathContext(t *testing.T) {
	ctx := context.Background()

	// Default value when no config path set
	if got := cmd.ConfigPathFrom(ctx); got != "orchestrator.yaml" {
		t.Errorf("expected default config path, got %q", got)
	}

	// Set custom path
	ctx = cmd.WithConfigPath(ctx, "/custom/config.yaml")
	if got := cmd.ConfigPathFrom(ctx); got != "/custom/config.yaml" {
		t.Errorf("expected custom config path, got %q", got)
	}
}

func TestNewMissionCmdFlagsDefaults(t *testing.T) {
	missionCmd := cmd.NewMissionCmd()

	// Verify flag registrations
	fileFlag := missionCmd.Flags().Lookup("file")
	if fileFlag == nil {
		t.Fatal("expected --file flag to be registered")
	}
	if fileFlag.DefValue != "" {
		t.Errorf("expected empty default for --file, got %q", fileFlag.DefValue)
	}

	dryRunFlag := missionCmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Fatal("expected --dry-run flag to be registered")
	}
	if dryRunFlag.DefValue != "false" {
		t.Errorf("expected false default for --dry-run, got %q", dryRunFlag.DefValue)
	}
}

func TestNewMissionCmdRequiresInput(t *testing.T) {
	missionCmd := cmd.NewMissionCmd()
	missionCmd.SetArgs([]string{}) // No args, no --file
	missionCmd.SetOut(&bytes.Buffer{})
	missionCmd.SetErr(&bytes.Buffer{})

	err := missionCmd.Execute()
	if err == nil {
		t.Error("expected error when no mission description provided")
	}
}

func TestNewStatusCmdAcceptsOptionalArg(t *testing.T) {
	statusCmd := cmd.NewStatusCmd()

	// Should accept 0 args
	if err := cobra.MaximumNArgs(1)(statusCmd, []string{}); err != nil {
		t.Errorf("status should accept 0 args: %v", err)
	}

	// Should accept 1 arg
	if err := cobra.MaximumNArgs(1)(statusCmd, []string{"mission-123"}); err != nil {
		t.Errorf("status should accept 1 arg: %v", err)
	}
}

func TestNewAgentsCmdHasSubcommands(t *testing.T) {
	agentsCmd := cmd.NewAgentsCmd()

	subCmds := agentsCmd.Commands()
	if len(subCmds) < 2 {
		t.Errorf("expected at least 2 subcommands (list, info), got %d", len(subCmds))
	}

	foundList, foundInfo := false, false
	for _, sub := range subCmds {
		switch sub.Name() {
		case "list":
			foundList = true
		case "info":
			foundInfo = true
		}
	}

	if !foundList {
		t.Error("missing 'list' subcommand")
	}
	if !foundInfo {
		t.Error("missing 'info' subcommand")
	}
}
```

## Rules
1. **Test Registration, Not Execution**: CLI tests verify flag parsing, arg validation, and subcommand structure — NOT full kernel bootstrap (that's integration testing).
2. **Isolated Buffers**: Use `SetOut`/`SetErr` with `bytes.Buffer` to capture output without polluting test stdout.

## Verify
```bash
go test ./cmd/orchestrator/cmd/... -v
```

## Checklist
- [ ] File `cmd/orchestrator/cmd/cmd_test.go` exists
- [ ] Tests context helper round-trip
- [ ] Tests flag default values
- [ ] Tests argument validation
- [ ] Tests subcommand registration
- [ ] `go test ./cmd/orchestrator/cmd/...` passes
