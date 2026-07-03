# Micro-Task 6.18: Create modules/session/session.go

## Info
- **File**: `modules/session/session.go`
- **Package**: `session`
- **Depends on**: None
- **Time**: 15 min
- **Verify**: `go build ./modules/session/...`

## Purpose
Session state serialization for crash recovery. Periodically checkpoints mission execution state to JSON files so missions can resume after unexpected restarts.

## EXACT code to create

```go
// Package session provides mission execution state checkpoint and recovery.
package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// State holds the serializable execution state for a mission.
type State struct {
	MissionID      string            `json:"mission_id"`
	CurrentTaskIdx int               `json:"current_task_idx"`
	CompletedTasks []string          `json:"completed_tasks"`
	FailedTasks    []string          `json:"failed_tasks"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	CheckpointAt   time.Time         `json:"checkpoint_at"`
}

// Manager handles session checkpoint persistence. Thread-safe.
type Manager struct {
	mu      sync.Mutex
	dataDir string
}

// NewManager constructs a session manager storing checkpoints in dataDir.
func NewManager(dataDir string) (*Manager, error) {
	checkpointDir := filepath.Join(dataDir, "sessions")
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		return nil, fmt.Errorf("session: failed to create checkpoint dir: %w", err)
	}
	return &Manager{dataDir: checkpointDir}, nil
}

// Save persists the current session state to disk.
func (m *Manager) Save(state *State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state.CheckpointAt = time.Now()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("session: failed to marshal state: %w", err)
	}

	path := m.checkpointPath(state.MissionID)

	// Atomic write: write to temp file then rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("session: failed to write checkpoint: %w", err)
	}

	return os.Rename(tmpPath, path)
}

// Restore loads a session state from disk. Returns nil if no checkpoint exists.
func (m *Manager) Restore(missionID string) (*State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.checkpointPath(missionID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No checkpoint found
		}
		return nil, fmt.Errorf("session: failed to read checkpoint: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("session: corrupted checkpoint: %w", err)
	}

	return &state, nil
}

// Clear removes the checkpoint file for a completed mission.
func (m *Manager) Clear(missionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.checkpointPath(missionID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (m *Manager) checkpointPath(missionID string) string {
	return filepath.Join(m.dataDir, missionID+".json")
}
```

## Rules
1. **Atomic Write**: Write to `.tmp` file then `os.Rename`. Prevents corrupted checkpoints on crash during write.
2. **Serializable State**: Only store primitive types and slices/maps. NEVER store function pointers, channels, or mutexes.
3. **Nil on Missing**: `Restore` returns `nil, nil` when no checkpoint exists (fresh mission). Callers check for nil.

## Verify
```bash
go build ./modules/session/...
```

## Checklist
- [ ] Atomic write via temp file + rename
- [ ] Restore returns nil for fresh missions
- [ ] Clear removes checkpoint after completion
- [ ] `go build ./modules/session/...` passes
