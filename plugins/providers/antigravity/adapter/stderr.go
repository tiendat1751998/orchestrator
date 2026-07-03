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
