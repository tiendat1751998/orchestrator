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
