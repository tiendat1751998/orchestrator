// Package git implements source control management wrappers.
package git

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	contractstool "github.com/tiendat1751998/orchestrator/contracts/tool"
	sdktool "github.com/tiendat1751998/orchestrator/sdk/tool"
)

// =============================================================================
// Helper: Run Git Command
// =============================================================================

// runGit runs a git command with context inside the specified directory.
func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("git command failed: %w\nOutput:\n%s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// =============================================================================
// Git Status Tool
// =============================================================================

// GitStatusTool returns the current git working tree status.
type GitStatusTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

// NewGitStatusTool creates a new instance of GitStatusTool.
func NewGitStatusTool(workspaceDir string) (*GitStatusTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("git_status: failed to resolve workspace path: %w", err)
	}

	schema := contractstool.NewSchema()
	baseTool, err := sdktool.NewBaseTool("git_status", "Returns the current git working tree status", schema)
	if err != nil {
		return nil, fmt.Errorf("git_status: failed to create base tool: %w", err)
	}

	return &GitStatusTool{BaseTool: baseTool, workspaceDir: absWorkspace}, nil
}

// Execute runs the git status --porcelain command and returns the status output.
func (t *GitStatusTool) Execute(ctx context.Context, rawArgs json.RawMessage) (*contractstool.Result, error) {
	out, err := runGit(ctx, t.workspaceDir, "status", "--porcelain")
	if err != nil {
		return sdktool.Failure(err.Error()), nil
	}
	if out == "" {
		return sdktool.Success("Working tree clean"), nil
	}
	return sdktool.Success(out), nil
}

// =============================================================================
// Git Diff Tool
// =============================================================================

// GitDiffTool shows unstaged diff modifications in the workspace.
type GitDiffTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

// NewGitDiffTool creates a new instance of GitDiffTool.
func NewGitDiffTool(workspaceDir string) (*GitDiffTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("git_diff: failed to resolve workspace path: %w", err)
	}

	schema := contractstool.NewSchema()
	baseTool, err := sdktool.NewBaseTool("git_diff", "Shows unstaged diff modifications in the workspace", schema)
	if err != nil {
		return nil, fmt.Errorf("git_diff: failed to create base tool: %w", err)
	}

	return &GitDiffTool{BaseTool: baseTool, workspaceDir: absWorkspace}, nil
}

// Execute runs the git diff command and truncates output if it exceeds limits.
func (t *GitDiffTool) Execute(ctx context.Context, rawArgs json.RawMessage) (*contractstool.Result, error) {
	out, err := runGit(ctx, t.workspaceDir, "diff")
	if err != nil {
		return sdktool.Failure(err.Error()), nil
	}
	// Limit diff size to prevent token limits overflow
	if len(out) > 50000 {
		out = out[:50000] + "\n...[Diff truncated due to size limits]..."
	}
	return sdktool.Success(out), nil
}

// =============================================================================
// Git Add Tool
// =============================================================================

// GitAddTool stages files to the index.
type GitAddTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

// GitAddArgs holds arguments for the git add tool.
type GitAddArgs struct {
	Path string `json:"path"`
}

// NewGitAddTool creates a new instance of GitAddTool.
func NewGitAddTool(workspaceDir string) (*GitAddTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("git_add: failed to resolve workspace path: %w", err)
	}

	schema := contractstool.NewSchema().
		AddProperty("path", contractstool.Property{
			Type:        "string",
			Description: "File path relative to the workspace root to stage",
		}).
		AddRequired("path")

	baseTool, err := sdktool.NewBaseTool("git_add", "Stages files to the index", schema)
	if err != nil {
		return nil, fmt.Errorf("git_add: failed to create base tool: %w", err)
	}

	return &GitAddTool{BaseTool: baseTool, workspaceDir: absWorkspace}, nil
}

// Execute runs the git add command to stage specified paths.
func (t *GitAddTool) Execute(ctx context.Context, rawArgs json.RawMessage) (*contractstool.Result, error) {
	if err := t.ValidateArguments(rawArgs); err != nil {
		return sdktool.Failure(err.Error()), nil
	}

	var args GitAddArgs
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return sdktool.Failure(fmt.Sprintf("git_add: invalid arguments: %v", err)), nil
	}

	_, err := runGit(ctx, t.workspaceDir, "add", args.Path)
	if err != nil {
		return sdktool.Failure(err.Error()), nil
	}

	return sdktool.Success(fmt.Sprintf("Staged %s", args.Path)), nil
}

// =============================================================================
// Git Commit Tool
// =============================================================================

// GitCommitTool commits staged changes to history.
type GitCommitTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

// GitCommitArgs holds arguments for the git commit tool.
type GitCommitArgs struct {
	Message string `json:"message"`
}

// NewGitCommitTool creates a new instance of GitCommitTool.
func NewGitCommitTool(workspaceDir string) (*GitCommitTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("git_commit: failed to resolve workspace path: %w", err)
	}

	schema := contractstool.NewSchema().
		AddProperty("message", contractstool.Property{
			Type:        "string",
			Description: "Commit description message text",
		}).
		AddRequired("message")

	baseTool, err := sdktool.NewBaseTool("git_commit", "Commits staged changes to history", schema)
	if err != nil {
		return nil, fmt.Errorf("git_commit: failed to create base tool: %w", err)
	}

	return &GitCommitTool{BaseTool: baseTool, workspaceDir: absWorkspace}, nil
}

// Execute runs the git commit command with the specified message.
func (t *GitCommitTool) Execute(ctx context.Context, rawArgs json.RawMessage) (*contractstool.Result, error) {
	if err := t.ValidateArguments(rawArgs); err != nil {
		return sdktool.Failure(err.Error()), nil
	}

	var args GitCommitArgs
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return sdktool.Failure(fmt.Sprintf("git_commit: invalid arguments: %v", err)), nil
	}

	out, err := runGit(ctx, t.workspaceDir, "commit", "-m", args.Message)
	if err != nil {
		return sdktool.Failure(err.Error()), nil
	}

	return sdktool.Success(out), nil
}

// =============================================================================
// Git Log Tool
// =============================================================================

// GitLogTool displays recent commits log history.
type GitLogTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

// NewGitLogTool creates a new instance of GitLogTool.
func NewGitLogTool(workspaceDir string) (*GitLogTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("git_log: failed to resolve workspace path: %w", err)
	}

	schema := contractstool.NewSchema()
	baseTool, err := sdktool.NewBaseTool("git_log", "Displays recent commits log history", schema)
	if err != nil {
		return nil, fmt.Errorf("git_log: failed to create base tool: %w", err)
	}

	return &GitLogTool{BaseTool: baseTool, workspaceDir: absWorkspace}, nil
}

// Execute runs the git log command.
func (t *GitLogTool) Execute(ctx context.Context, rawArgs json.RawMessage) (*contractstool.Result, error) {
	out, err := runGit(ctx, t.workspaceDir, "log", "-n", "10", "--oneline")
	if err != nil {
		return sdktool.Failure(err.Error()), nil
	}
	return sdktool.Success(out), nil
}
