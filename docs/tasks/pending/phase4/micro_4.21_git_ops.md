# Micro-Task 4.21: Create plugins/tools/git/git.go

## Info
- **File**: `plugins/tools/git/git.go`
- **Package**: `git`
- **Depends on**: 4.20
- **Time**: 30 min
- **Verify**: `go build ./plugins/tools/git/...`

## Purpose
Implements the Git operations tools (`GitStatusTool`, `GitDiffTool`, `GitAddTool`, `GitCommitTool`, `GitLogTool`, and `GitCloneTool`) to manage source control status and changes within the workspace boundary.

## EXACT code to create

```go
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

type GitStatusTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

func NewGitStatusTool(workspaceDir string) (*GitStatusTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, err
	}

	schema := contractstool.NewSchema()
	baseTool, err := sdktool.NewBaseTool("git_status", "Returns the current git working tree status", schema)
	if err != nil {
		return nil, err
	}

	return &GitStatusTool{BaseTool: baseTool, workspaceDir: absWorkspace}, nil
}

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

type GitDiffTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

func NewGitDiffTool(workspaceDir string) (*GitDiffTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, err
	}

	schema := contractstool.NewSchema()
	baseTool, err := sdktool.NewBaseTool("git_diff", "Shows unstaged diff modifications in the workspace", schema)
	if err != nil {
		return nil, err
	}

	return &GitDiffTool{BaseTool: baseTool, workspaceDir: absWorkspace}, nil
}

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

type GitAddTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

type GitAddArgs struct {
	Path string `json:"path"`
}

func NewGitAddTool(workspaceDir string) (*GitAddTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, err
	}

	schema := contractstool.NewSchema().
		AddProperty("path", contractstool.Property{
			Type:        "string",
			Description: "File path relative to the workspace root to stage",
		}).
		AddRequired("path")

	baseTool, err := sdktool.NewBaseTool("git_add", "Stages files to the index", schema)
	if err != nil {
		return nil, err
	}

	return &GitAddTool{BaseTool: baseTool, workspaceDir: absWorkspace}, nil
}

func (t *GitAddTool) Execute(ctx context.Context, rawArgs json.RawMessage) (*contractstool.Result, error) {
	if err := t.ValidateArguments(rawArgs); err != nil {
		return sdktool.Failure(err.Error()), nil
	}

	var args GitAddArgs
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return sdktool.Failure(err.Error()), nil
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

type GitCommitTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

type GitCommitArgs struct {
	Message string `json:"message"`
}

func NewGitCommitTool(workspaceDir string) (*GitCommitTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, err
	}

	schema := contractstool.NewSchema().
		AddProperty("message", contractstool.Property{
			Type:        "string",
			Description: "Commit description message text",
		}).
		AddRequired("message")

	baseTool, err := sdktool.NewBaseTool("git_commit", "Commits staged changes to history", schema)
	if err != nil {
		return nil, err
	}

	return &GitCommitTool{BaseTool: baseTool, workspaceDir: absWorkspace}, nil
}

func (t *GitCommitTool) Execute(ctx context.Context, rawArgs json.RawMessage) (*contractstool.Result, error) {
	if err := t.ValidateArguments(rawArgs); err != nil {
		return sdktool.Failure(err.Error()), nil
	}

	var args GitCommitArgs
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return sdktool.Failure(err.Error()), nil
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

type GitLogTool struct {
	*sdktool.BaseTool
	workspaceDir string
}

func NewGitLogTool(workspaceDir string) (*GitLogTool, error) {
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, err
	}

	schema := contractstool.NewSchema()
	baseTool, err := sdktool.NewBaseTool("git_log", "Displays recent commits log history", schema)
	if err != nil {
		return nil, err
	}

	return &GitLogTool{BaseTool: baseTool, workspaceDir: absWorkspace}, nil
}

func (t *GitLogTool) Execute(ctx context.Context, rawArgs json.RawMessage) (*contractstool.Result, error) {
	out, err := runGit(ctx, t.workspaceDir, "log", "-n", "10", "--oneline")
	if err != nil {
		return sdktool.Failure(err.Error()), nil
	}
	return sdktool.Success(out), nil
}
```

## Pitfalls

### Pitfall 1: Executing Git command files on parent folders
```go
// WRONG:
cmd := exec.Command("git", "status") // Dir is not set: runs in orchestrator's directory!

// CORRECT:
cmd := exec.Command("git", "status")
cmd.Dir = dir
```
If the adapter process ignores `cmd.Dir` configurations, git executes commands in the system default path or the folder where the orchestrator was launched. Set directory explicitly.

### Pitfall 2: Reaching token limits on large diff modifications
Running `git diff` after making large changes can output hundreds of thousands of lines. If this is piped raw to the prompt, it will exhaust token allocations. Truncate large outputs.

## Verify
```bash
go build ./plugins/tools/git/...
```

## Checklist
- [ ] File exists at `plugins/tools/git/git.go`
- [ ] Package name is `git`
- [ ] All exported types have Godoc
- [ ] Commands execute inside workspace target directories
- [ ] Large git diff outputs are truncated
- [ ] Build command passes
