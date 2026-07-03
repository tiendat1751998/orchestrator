# Micro-Task 6.20: Create modules/execution/logger.go

## Info
- **File**: `modules/execution/logger.go`
- **Package**: `execution`
- **Depends on**: None
- **Time**: 15 min
- **Verify**: `go build ./modules/execution/...`

## Purpose
Per-task structured execution logger with JSON lines output and size-based log rotation. Enables execution replay and debugging of task failures.

## EXACT code to create

```go
// Package execution provides per-task execution logging and replay.
package execution

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry represents a single log entry in the execution log.
type Entry struct {
	Timestamp time.Time `json:"ts"`
	Level     string    `json:"level"`
	TaskID    string    `json:"task_id"`
	AgentName string    `json:"agent"`
	Message   string    `json:"msg"`
	Data      any       `json:"data,omitempty"`
}

// Logger writes structured execution logs for a mission. Thread-safe.
type Logger struct {
	mu         sync.Mutex
	file       *os.File
	encoder    *json.Encoder
	maxSizeBytes int64
	currentSize  int64
	basePath     string
}

// NewLogger creates an execution logger writing to the given file path.
// maxSizeMB sets the rotation threshold (0 = no rotation).
func NewLogger(basePath string, maxSizeMB int) (*Logger, error) {
	dir := filepath.Dir(basePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("execution: failed to create log dir: %w", err)
	}

	f, err := os.OpenFile(basePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("execution: failed to open log file: %w", err)
	}

	info, _ := f.Stat()
	currentSize := int64(0)
	if info != nil {
		currentSize = info.Size()
	}

	return &Logger{
		file:         f,
		encoder:      json.NewEncoder(f),
		maxSizeBytes: int64(maxSizeMB) * 1024 * 1024,
		currentSize:  currentSize,
		basePath:     basePath,
	}, nil
}

// Log writes a structured entry to the execution log.
func (l *Logger) Log(taskID, agentName, level, message string, data any) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := Entry{
		Timestamp: time.Now(),
		Level:     level,
		TaskID:    taskID,
		AgentName: agentName,
		Message:   message,
		Data:      data,
	}

	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Check rotation threshold
	if l.maxSizeBytes > 0 && l.currentSize+int64(len(raw)) > l.maxSizeBytes {
		if err := l.rotate(); err != nil {
			return fmt.Errorf("execution: log rotation failed: %w", err)
		}
	}

	if err := l.encoder.Encode(entry); err != nil {
		return err
	}

	l.currentSize += int64(len(raw)) + 1 // +1 for newline
	return nil
}

func (l *Logger) rotate() error {
	l.file.Close()

	rotatedPath := fmt.Sprintf("%s.%d", l.basePath, time.Now().UnixMilli())
	if err := os.Rename(l.basePath, rotatedPath); err != nil {
		return err
	}

	f, err := os.OpenFile(l.basePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	l.file = f
	l.encoder = json.NewEncoder(f)
	l.currentSize = 0
	return nil
}

// Close flushes and closes the log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

// Replay reads all entries from a log file for debugging/replay.
func Replay(logPath string) ([]Entry, error) {
	f, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []Entry
	decoder := json.NewDecoder(f)
	for {
		var e Entry
		if err := decoder.Decode(&e); err != nil {
			if err == io.EOF {
				break
			}
			return entries, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
```

## Rules
1. **JSON Lines Format**: One JSON object per line. Enables `grep` and streaming analysis.
2. **Size-Based Rotation**: Rotate when file exceeds threshold. Renamed with timestamp suffix.
3. **Append Mode**: Always `O_APPEND` for crash safety.

## Verify
```bash
go build ./modules/execution/...
```

## Checklist
- [ ] JSON Lines structured logging
- [ ] Size-based rotation
- [ ] Replay function for debugging
- [ ] `go build ./modules/execution/...` passes
