# Micro-Task 4.05: Create plugins/providers/antigravity/adapter/stdout.go

## Info
- **File**: `plugins/providers/antigravity/adapter/stdout.go`
- **Package**: `adapter`
- **Depends on**: 4.04
- **Time**: 25 min
- **Verify**: `go build ./plugins/providers/antigravity/adapter/...`

## Purpose
Implements the standard output reader helper (`ReadResponse` and helper loops) that reads response payloads from the CLI stdout pipe, detects completion delimiters, and avoids pipe deadlock blocks.

## EXACT code to create

```go
package adapter

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ReadResponse blocks and reads characters from the stdout pipe until:
//  1. The specified delimiter token is detected in the stream.
//  2. The process closes the stdout pipe (EOF).
//  3. The context is cancelled or times out.
func (a *CLIAdapter) ReadResponse(ctx context.Context, delimiter string) (string, error) {
	if delimiter == "" {
		delimiter = "---END---" // Default fallback delimiter
	}

	a.mu.Lock()
	stdout := a.stdout
	a.mu.Unlock()

	if stdout == nil {
		return "", errors.New("adapter: stdout pipe is not open")
	}

	type readResult struct {
		output string
		err    error
	}

	resChan := make(chan readResult, 1)

	// Spawn a background goroutine to read from stdout
	go func() {
		var builder strings.Builder
		reader := bufio.NewReader(stdout)
		buf := make([]byte, 1024)

		for {
			n, err := reader.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				builder.WriteString(chunk)

				currentText := builder.String()
				if strings.Contains(currentText, delimiter) {
					// Delimiter detected: strip the delimiter and return
					parts := strings.Split(currentText, delimiter)
					resChan <- readResult{output: parts[0], err: nil}
					return
				}
			}

			if err != nil {
				if errors.Is(err, io.EOF) {
					resChan <- readResult{output: builder.String(), err: nil}
					return
				}
				resChan <- readResult{output: builder.String(), err: err}
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-resChan:
		if res.err != nil {
			return res.output, fmt.Errorf("adapter: failed to read stdout: %w", res.err)
		}
		return res.output, nil
	}
}
```

## Pitfalls

### Pitfall 1: Blocking stdout reads on main threads
```go
// WRONG:
func (a *CLIAdapter) ReadResponse(delimiter string) (string, error) {
    // Read stdout synchronously in main thread loop...
}
```
If the CLI process stops writing or crashes without closing stdout, a synchronous read call will block the caller thread indefinitely, ignoring context cancellations. Always use background reader goroutines and channel triggers.

### Pitfall 2: Buffer deadlocks from unread stdout lines
If the AI output is larger than the OS stdout buffer (typically 64KB) and the reader is not actively draining the pipe, the process blocks on write operations, causing a deadlock. Read the stream incrementally in a loop.

## Verify
```bash
go build ./plugins/providers/antigravity/adapter/...
```

## Checklist
- [ ] File exists at `plugins/providers/antigravity/adapter/stdout.go`
- [ ] Package name is `adapter`
- [ ] All exported types have Godoc
- [ ] Readers run inside background goroutines
- [ ] Delimiter detection stops reads and returns outputs
- [ ] Context cancellation channel interrupts read loops
- [ ] Build command passes
