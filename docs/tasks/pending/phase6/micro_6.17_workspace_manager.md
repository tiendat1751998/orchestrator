# Micro-Task 6.17: Create modules/workspace/workspace.go

## Info
- **File**: `modules/workspace/workspace.go`
- **Package**: `workspace`
- **Depends on**: None
- **Time**: 15 min
- **Verify**: `go build ./modules/workspace/...`

## Purpose
Manages isolated working directories per mission. Creates, locks, and cleans up workspace directories to prevent conflicts when multiple missions run concurrently.

## EXACT code to create

```go
// Package workspace provides mission workspace isolation and lifecycle management.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Manager handles workspace directory lifecycle. Thread-safe.
type Manager struct {
	mu      sync.Mutex
	baseDir string
	active  map[string]string // missionID → workspace path
}

// NewManager constructs a workspace manager rooted at baseDir.
func NewManager(baseDir string) (*Manager, error) {
	absDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("workspace: invalid base dir: %w", err)
	}

	if err := os.MkdirAll(absDir, 0755); err != nil {
		return nil, fmt.Errorf("workspace: failed to create base dir: %w", err)
	}

	return &Manager{
		baseDir: absDir,
		active:  make(map[string]string),
	}, nil
}

// Acquire creates and locks a workspace directory for the given mission.
func (m *Manager) Acquire(missionID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if path, exists := m.active[missionID]; exists {
		return path, nil // Already acquired
	}

	workDir := filepath.Join(m.baseDir, missionID)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("workspace: failed to create dir for %q: %w", missionID, err)
	}

	m.active[missionID] = workDir
	return workDir, nil
}

// Release marks a workspace as no longer active.
func (m *Manager) Release(missionID string) {
	m.mu.Lock()
	delete(m.active, missionID)
	m.mu.Unlock()
}

// Cleanup removes the workspace directory for a completed mission.
func (m *Manager) Cleanup(missionID string) error {
	m.mu.Lock()
	path, exists := m.active[missionID]
	if exists {
		delete(m.active, missionID)
	} else {
		path = filepath.Join(m.baseDir, missionID)
	}
	m.mu.Unlock()

	return os.RemoveAll(path)
}

// Path returns the workspace path for a mission, or empty string if not acquired.
func (m *Manager) Path(missionID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active[missionID]
}
```

## Rules
1. **Directory Isolation**: Each mission gets its own directory under `baseDir/<missionID>/`. No shared filesystem state.
2. **Acquire/Release Pattern**: Callers must `Acquire` before use and `Release` after completion. Prevents stale locks.

## Verify
```bash
go build ./modules/workspace/...
```

## Checklist
- [ ] File exists, creates isolated directories per mission
- [ ] Thread-safe acquire/release
- [ ] Cleanup removes directory
- [ ] `go build ./modules/workspace/...` passes
