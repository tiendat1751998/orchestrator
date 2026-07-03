# Micro-Task 4.03: Create plugins/providers/antigravity/adapter/cli.go

## Info
- **File**: `plugins/providers/antigravity/adapter/cli.go`
- **Package**: `adapter`
- **Depends on**: 4.02
- **Time**: 25 min
- **Verify**: `go build ./plugins/providers/antigravity/adapter/...`

## Purpose
Implements the background CLI process lifecycle manager (`CLIAdapter` and constructors) that spawns the Antigravity command-line process, exposes stdin/stdout pipes, and manages thread-safe process cleanup.

## EXACT code to create

```go
// Package adapter implements the low-level CLI pipe adapter for the Antigravity model.
package adapter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// CLIAdapter manages the execution lifecycle of a single Antigravity CLI process.
// Thread-safe.
type CLIAdapter struct {
	binary string
	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

// NewCLIAdapter constructs a new CLIAdapter.
func NewCLIAdapter(binary string) *CLIAdapter {
	return &CLIAdapter{
		binary: binary,
	}
}

// Start spawns the CLI process and hooks standard input/output/error pipes.
func (a *CLIAdapter) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cmd != nil && a.cmd.Process != nil && a.cmd.ProcessState == nil {
		return errors.New("adapter: CLI process is already running")
	}

	cmd := exec.CommandContext(ctx, a.binary)

	// Set platform-specific process group attributes.
	// setProcAttr is defined in build-tagged files:
	//   - procattr_unix.go    (//go:build !windows)  → sets Setpgid: true
	//   - procattr_windows.go (//go:build windows)   → sets CreationFlags: CREATE_NEW_PROCESS_GROUP
	cmd.SysProcAttr = newProcAttr()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("adapter: failed to open stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return fmt.Errorf("adapter: failed to open stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return fmt.Errorf("adapter: failed to open stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		_ = stderr.Close()
		return fmt.Errorf("adapter: failed to start CLI command: %w", err)
	}

	a.cmd = cmd
	a.stdin = stdin
	a.stdout = stdout
	a.stderr = stderr

	return nil
}

// Stop gracefully halts the CLI process and closes all active pipes.
// Safe to call multiple times (idempotent).
func (a *CLIAdapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cmd == nil || a.cmd.Process == nil {
		return nil
	}

	// Close pipes to unblock any reading goroutines
	if a.stdin != nil {
		_ = a.stdin.Close()
	}
	if a.stdout != nil {
		_ = a.stdout.Close()
	}
	if a.stderr != nil {
		_ = a.stderr.Close()
	}

	// Terminate the process group
	pid := a.cmd.Process.Pid
	err := killProcessGroup(pid)

	_ = a.cmd.Wait() // Clean up process descriptors to prevent zombie processes

	a.cmd = nil
	a.stdin = nil
	a.stdout = nil
	a.stderr = nil

	return err
}

// killProcessGroup kills the target process and all children in its process group.
// Defined in build-tagged files:
//   - procattr_unix.go    (//go:build !windows) → syscall.Kill(-pid, SIGKILL)
//   - procattr_windows.go (//go:build windows)  → taskkill /F /T /PID

// Pipes returns the active readers/writers. Safe for adapter helpers.
func (a *CLIAdapter) Pipes() (io.WriteCloser, io.ReadCloser, io.ReadCloser, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cmd == nil || a.cmd.Process == nil {
		return nil, nil, nil, errors.New("adapter: process not running")
	}

	return a.stdin, a.stdout, a.stderr, nil
}
```

## Pitfalls

### Pitfall 1: Leaking child processes on Windows
```go
// WRONG:
a.cmd.Process.Kill() // Only kills CMD.exe, leaving compilation/AI processes orphaned and running.

// CORRECT:
if runtime.GOOS == "windows" {
    exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", pid)).Run()
}
```
Using standard process kill commands leaves orphaned child processes running on Windows. Always utilize process groups and taskkill tools to clear the tree.

### Pitfall 2: Cross-platform compilation crashes on platform-specific constants
Using Windows-only syscall constants directly on POSIX environments fails compilation. Access them using hex numbers or build tags to preserve portability.

## Verify
```bash
go build ./plugins/providers/antigravity/adapter/...
# Expected: clean compilation without errors
```

## Checklist
- [ ] File exists at `plugins/providers/antigravity/adapter/cli.go`
- [ ] Package name is `adapter`
- [ ] All exported types have Godoc
- [ ] `CLIAdapter` synchronization protects process commands
- [ ] Process creation uses cross-platform process group flags
- [ ] `Stop` handles child process cleanups using taskkill or PGID signals
- [ ] Build command passes
