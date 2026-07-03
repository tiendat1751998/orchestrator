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
