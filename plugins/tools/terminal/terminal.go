// Package terminal implements shell execution helpers.
package terminal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	contractstool "github.com/tiendat1751998/orchestrator/contracts/tool"
	sdktool "github.com/tiendat1751998/orchestrator/sdk/tool"
)

// RunCommandTool executes shell commands in a sandboxed directory.
type RunCommandTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

// RunCommandArgs maps JSON input parameters.
type RunCommandArgs struct {
	Command string `json:"command"`
}

// NewRunCommandTool constructs a RunCommandTool.
func NewRunCommandTool(workspaceDir string) (*RunCommandTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, err
	}

	schema := contractstool.NewSchema().
		AddProperty("command", contractstool.Property{
			Type:        "string",
			Description: "The exact command string to execute in the shell",
		}).
		AddRequired("command")

	baseTool, err := sdktool.NewBaseTool("run_command", "Executes terminal commands inside the workspace sandbox", schema)
	if err != nil {
		return nil, err
	}

	return &RunCommandTool{
		BaseTool:     baseTool,
		workspaceDir: absWorkspace,
	}, nil
}

// Execute parses command arguments, checks blocklists, and runs commands.
func (t *RunCommandTool) Execute(ctx context.Context, rawArgs json.RawMessage) (*contractstool.Result, error) {
	if err := t.ValidateArguments(rawArgs); err != nil {
		return sdktool.Failure(err.Error()), nil
	}

	var args RunCommandArgs
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return sdktool.Failure(err.Error()), nil
	}

	cmdStr := strings.TrimSpace(args.Command)
	if cmdStr == "" {
		return sdktool.Failure("run_command: empty command query"), nil
	}

	// 1. Verify command security blocklist
	if err := verifySecurityPolicy(cmdStr); err != nil {
		return sdktool.Failure(fmt.Sprintf("run_command: security error: %v", err)), nil
	}

	// 2. Setup shell execution based on target OS
	var shellName string
	var shellArgs []string

	if runtime.GOOS == "windows" {
		shellName = "powershell"
		shellArgs = []string{"-NoProfile", "-NonInteractive", "-Command", cmdStr}
	} else {
		shellName = "/bin/sh"
		shellArgs = []string{"-c", cmdStr}
	}

	// 3. Set standard execution bounds (timeout = 30 seconds)
	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, shellName, shellArgs...)
	cmd.Dir = t.workspaceDir

	// Execute command and fetch combined output
	output, err := cmd.CombinedOutput()

	// 4. Handle results and enforce output size limits (max 100KB)
	outStr := string(output)
	if len(outStr) > 100*1024 {
		outStr = outStr[:100*1024] + "\n...[Command output truncated: exceeded 100KB limit]..."
	}

	if err != nil {
		if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
			return sdktool.Failure(fmt.Sprintf("run_command: execution timeout exceeded (30s)\nOutput:\n%s", outStr)), nil
		}
		return sdktool.Failure(fmt.Sprintf("run_command failed: %v\nOutput:\n%s", err, outStr)), nil
	}

	return sdktool.Success(outStr), nil
}

// verifySecurityPolicy checks the command against a blocklist of dangerous commands.
func verifySecurityPolicy(cmd string) error {
	lower := strings.ToLower(cmd)

	// Block dangerous system-altering commands
	blockedTokens := []string{
		"rm -rf", "del /", "format ", "mkfs", "dd ",
		"taskkill", "kill -9", "shutdown", "reboot",
		"net user", "chmod -r 777",
	}

	for _, token := range blockedTokens {
		if strings.Contains(lower, token) {
			return fmt.Errorf("command contains blocked token %q", token)
		}
	}

	return nil
}
