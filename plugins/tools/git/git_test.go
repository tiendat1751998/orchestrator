package git_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tiendat1751998/orchestrator/plugins/tools/git"
)

func execGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestGitTools_Workflow(t *testing.T) {
	// Create a temp workspace directory
	tmpDir := t.TempDir()

	// Initialize git repo in the temp directory
	ctx := context.Background()
	_, err := execGit(ctx, tmpDir, "init")
	if err != nil {
		t.Skipf("Skipping Git tests because git init failed (git might not be installed or configured): %v", err)
	}

	// Configure dummy git user so commit works
	_, _ = execGit(ctx, tmpDir, "config", "user.name", "Test User")
	_, _ = execGit(ctx, tmpDir, "config", "user.email", "test@example.com")
	// Make sure default branch name doesn't interfere
	_, _ = execGit(ctx, tmpDir, "config", "init.defaultBranch", "main")

	// 1. Test GitStatusTool on empty clean repo
	statusTool, err := git.NewGitStatusTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to create GitStatusTool: %v", err)
	}
	res, err := statusTool.Execute(ctx, json.RawMessage("{}"))
	if err != nil {
		t.Fatalf("unexpected error executing statusTool: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected success status, got exit code %d (error: %q)", res.ExitCode, res.Error)
	}
	if res.Output != "Working tree clean" {
		t.Errorf("expected clean working tree, got %q", res.Output)
	}

	// 2. Create a test file
	testFilePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFilePath, []byte("hello world\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test status showing untracked file
	res, err = statusTool.Execute(ctx, json.RawMessage("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "test.txt") {
		t.Errorf("expected status output to contain test.txt, got %q", res.Output)
	}

	// 3. Test GitAddTool
	addTool, err := git.NewGitAddTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to create GitAddTool: %v", err)
	}
	res, err = addTool.Execute(ctx, json.RawMessage(`{"path": "test.txt"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected success add, got error: %q", res.Error)
	}
	if !strings.Contains(res.Output, "Staged test.txt") {
		t.Errorf("expected add output to confirm staged, got %q", res.Output)
	}

	// 4. Test GitCommitTool
	commitTool, err := git.NewGitCommitTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to create GitCommitTool: %v", err)
	}
	res, err = commitTool.Execute(ctx, json.RawMessage(`{"message": "initial commit"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected success commit, got error: %q", res.Error)
	}

	// 5. Test GitLogTool
	logTool, err := git.NewGitLogTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to create GitLogTool: %v", err)
	}
	res, err = logTool.Execute(ctx, json.RawMessage("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected success log, got error: %q", res.Error)
	}
	if !strings.Contains(res.Output, "initial commit") {
		t.Errorf("expected log output to contain 'initial commit', got %q", res.Output)
	}

	// 6. Test GitDiffTool
	diffTool, err := git.NewGitDiffTool(tmpDir)
	if err != nil {
		t.Fatalf("failed to create GitDiffTool: %v", err)
	}
	// Currently no modifications
	res, err = diffTool.Execute(ctx, json.RawMessage("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Output != "" {
		t.Errorf("expected empty diff, got %q", res.Output)
	}

	// Modify file to produce a diff
	if err := os.WriteFile(testFilePath, []byte("hello world\nmodified line\n"), 0644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}
	res, err = diffTool.Execute(ctx, json.RawMessage("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "+modified line") {
		t.Errorf("expected diff to contain modified line, got %q", res.Output)
	}

	// 7. Test GitDiffTool truncation
	// Create a large change (55KB) to verify truncation
	largeContent := strings.Repeat("a\n", 30000) // 60KB
	if err := os.WriteFile(testFilePath, []byte(largeContent), 0644); err != nil {
		t.Fatalf("failed to write large file: %v", err)
	}
	res, err = diffTool.Execute(ctx, json.RawMessage("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "[Diff truncated due to size limits]") {
		t.Errorf("expected diff to be truncated, got length %d", len(res.Output))
	}
}
