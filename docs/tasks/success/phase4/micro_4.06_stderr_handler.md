# Micro-Task 4.06: Create plugins/providers/antigravity/adapter/stderr.go

## Info
- **File**: `plugins/providers/antigravity/adapter/stderr.go`
- **Package**: `adapter`
- **Depends on**: 4.05
- **Time**: 15 min
- **Verify**: `go build ./plugins/providers/antigravity/adapter/...`

## Purpose
Implements the standard error pipe reader helper (`MonitorStderr`) to drain and log CLI process stderr messages in the background, preventing pipe buffer deadlocks.

## EXACT code to create

```go
package adapter

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
)

// MonitorStderr spawns a background goroutine to drain the stderr pipe
// and logs any error output to the provided slog.Logger.
//
// Goroutine Safe:
//   - Immediately returns to allow concurrent execution of stdout readers.
//   - Drains stderr completely until EOF to prevent OS pipe buffer exhaustion.
func (a *CLIAdapter) MonitorStderr(logger *slog.Logger) error {
	a.mu.Lock()
	stderr := a.stderr
	a.mu.Unlock()

	if stderr == nil {
		return errors.New("adapter: stderr pipe is not open")
	}

	go func() {
		reader := bufio.NewReader(stderr)
		buf := make([]byte, 1024)

		for {
			n, err := reader.Read(buf)
			if n > 0 {
				errMsg := string(buf[:n])
				if logger != nil {
					logger.Error("antigravity CLI stderr output",
						"content", errMsg,
					)
				}
			}

			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				if logger != nil {
					logger.Error("antigravity CLI stderr read error",
						"error", err.Error(),
					)
				}
				return
			}
		}
	}()

	return nil
}
```

## Pitfalls

### Pitfall 1: Leaving the stderr pipe unread
If you read only stdout but ignore the stderr pipe, and the CLI process prints logs or warnings to stderr that exceed the OS pipe buffer limit (usually 64KB), the process will block on stderr writes, causing a deadlock. Read both stdout and stderr concurrently.

### Pitfall 2: Blocking stdout readers on stderr logs
Reading stdout and stderr sequentially on a single thread will cause hangs. Stderr reading must run in the background.

## Verify
```bash
go build ./plugins/providers/antigravity/adapter/...
```

## Checklist
- [ ] File exists at `plugins/providers/antigravity/adapter/stderr.go`
- [ ] Package name is `adapter`
- [ ] All exported types have Godoc
- [ ] `MonitorStderr` drains the pipe asynchronously in a background goroutine
- [ ] Stderr lines are captured and logged to `slog.Logger`
- [ ] Build command passes
