# Micro-Task 5.24: Create kernel/security/audit.go

## Info
- **File**: `kernel/security/audit.go`
- **Package**: `security`
- **Depends on**: 5.23
- **Time**: 15 min
- **Verify**: `go build ./kernel/security/...`

## Purpose
Implements the structured audit log writer (`AuditLogger` and formatting formats) to record agent actions to an append-only file.

## EXACT code to create

```go
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
	mu   sync.Mutex
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
```

## Pitfalls

### Pitfall 1: Corrupting logs on concurrent writes
```go
// WRONG:
func (al *AuditLogger) Log(...) {
    al.file.Write(bytes) // Data race! Multiple agents writing concurrently will corrupt the log file.
}

// CORRECT:
al.mu.Lock()
defer al.mu.Unlock()
```
Multiple agents execute tools in parallel. Writing to the audit log file without synchronization will interleave log entries and corrupt the file. Protect writes under locks.

### Pitfall 2: Silent failures when logs fail to write
Discarding errors when writes fail allows security events to go unrecorded. Always check write results and bubble up errors.

## Verify
```bash
go build ./kernel/security/...
# Expected: clean compilation without errors
```

## Checklist
- [ ] File exists at `kernel/security/audit.go`
- [ ] Package name is `security`
- [ ] All exported types have Godoc
- [ ] Log writes are synchronized under locks
- [ ] Entries are written in JSON lines format with newline endings
- [ ] Build command passes
