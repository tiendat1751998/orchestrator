package security

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// AuditEntry records a single security event.
type AuditEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Agent     string    `json:"agent"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Allowed   bool      `json:"allowed"`
	Details   string    `json:"details,omitempty"`
}

// AuditLogger writes events to an append-only log file.
// Thread-safe.
type AuditLogger struct {
	mu   sync.Mutex // ponytail: global lock for file append, upgrade to channel-based writer if high throughput is needed
	file *os.File
}

// NewAuditLogger constructs an AuditLogger.
func NewAuditLogger(logPath string) (*AuditLogger, error) {
	// Open file in append-only mode, create if missing
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("audit: failed to open audit log file: %w", err)
	}

	return &AuditLogger{
		file: f,
	}, nil
}

// Log records a security event to the log file.
func (al *AuditLogger) Log(agent, action, target string, allowed bool, details string) error {
	al.mu.Lock()
	defer al.mu.Unlock()

	entry := AuditEntry{
		Timestamp: time.Now(),
		Agent:     agent,
		Action:    action,
		Target:    target,
		Allowed:   allowed,
		Details:   details,
	}

	bytes, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("audit: failed to serialize entry: %w", err)
	}

	// Append entry with newline ending
	_, err = al.file.Write(append(bytes, '\n'))
	if err != nil {
		return fmt.Errorf("audit: failed to write entry to disk: %w", err)
	}

	return nil
}

// Close closes the underlying log file.
func (al *AuditLogger) Close() error {
	al.mu.Lock()
	defer al.mu.Unlock()
	if al.file != nil {
		return al.file.Close()
	}
	return nil
}
