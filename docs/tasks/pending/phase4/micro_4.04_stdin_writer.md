# Micro-Task 4.04: Create plugins/providers/antigravity/adapter/stdin.go

## Info
- **File**: `plugins/providers/antigravity/adapter/stdin.go`
- **Package**: `adapter`
- **Depends on**: 4.03
- **Time**: 15 min
- **Verify**: `go build ./plugins/providers/antigravity/adapter/...`

## Purpose
Implements the safe standard input pipe writer helper (`WritePrompt`) to handle writing prompts to the CLI adapter process concurrently.

## EXACT code to create

```go
package adapter

import (
	"errors"
	"fmt"
	"strings"
)

// WritePrompt formats and writes the query prompt to the CLI process stdin.
// Thread-safe.
func (a *CLIAdapter) WritePrompt(prompt string) error {
	if prompt == "" {
		return errors.New("adapter: cannot write empty prompt")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cmd == nil || a.cmd.Process == nil {
		return errors.New("adapter: process is not running")
	}

	if a.stdin == nil {
		return errors.New("adapter: stdin pipe is not open")
	}

	// Normalize line endings to standard carriage returns (Windows compat)
	normalized := strings.ReplaceAll(prompt, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\n", "\r\n")

	// Append delimiter token representing end of prompt payload
	if !strings.HasSuffix(normalized, "\r\n") {
		normalized += "\r\n"
	}

	// Write prompt payload
	_, err := fmt.Fprint(a.stdin, normalized)
	if err != nil {
		return fmt.Errorf("adapter: failed to write to stdin pipe: %w", err)
	}

	return nil
}
```

## Pitfalls

### Pitfall 1: Writing concurrently to stdin without locks
```go
// WRONG:
func (a *CLIAdapter) WritePrompt(prompt string) error {
    _, err := a.stdin.Write([]byte(prompt)) // Data race! Multiple threads write corrupt characters.
    return err
}

// CORRECT:
func (a *CLIAdapter) WritePrompt(prompt string) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    _, err := fmt.Fprint(a.stdin, prompt)
    return err
}
```
If two agents write prompts to the same CLI process simultaneously without locking, the characters interleave, producing corrupted inputs. Always wrap writes in mutex locks.

### Pitfall 2: Bypassing line ending normalization
Sending raw Unix-style line breaks (`\n`) to Windows command shells can cause command engines to hang waiting for carriage return endings (`\r\n`). Always normalize line breaks.

## Verify
```bash
go build ./plugins/providers/antigravity/adapter/...
```

## Checklist
- [ ] File exists at `plugins/providers/antigravity/adapter/stdin.go`
- [ ] Package name is `adapter`
- [ ] All exported types have Godoc
- [ ] Writes to stdin are guarded by `a.mu` locks
- [ ] Unix newlines (`\n`) are normalized to Windows-style newlines (`\r\n`)
- [ ] Build command passes
