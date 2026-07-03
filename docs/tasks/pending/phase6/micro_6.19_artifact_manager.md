# Micro-Task 6.19: Create modules/artifact/artifact.go

## Info
- **File**: `modules/artifact/artifact.go`
- **Package**: `artifact`
- **Depends on**: None
- **Time**: 15 min
- **Verify**: `go build ./modules/artifact/...`

## Purpose
Manages output artifacts (generated code, diffs, reports) produced by mission tasks. Each mission gets an isolated artifact directory with a metadata index for listing and retrieval.

## EXACT code to create

```go
// Package artifact manages mission output file storage and metadata indexing.
package artifact

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Record describes a single artifact file.
type Record struct {
	ID        string    `json:"id"`
	MissionID string    `json:"mission_id"`
	TaskID    string    `json:"task_id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Type      string    `json:"type"` // "code", "diff", "report", "log"
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

// Manager handles artifact storage and indexing. Thread-safe.
type Manager struct {
	mu      sync.Mutex
	baseDir string
	index   map[string][]Record // missionID → artifacts
}

// NewManager constructs an artifact manager rooted at baseDir.
func NewManager(baseDir string) (*Manager, error) {
	absDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("artifact: invalid base dir: %w", err)
	}
	if err := os.MkdirAll(absDir, 0755); err != nil {
		return nil, fmt.Errorf("artifact: failed to create base dir: %w", err)
	}
	return &Manager{
		baseDir: absDir,
		index:   make(map[string][]Record),
	}, nil
}

// Store writes artifact content to disk and indexes it.
func (m *Manager) Store(missionID, taskID, name, artifactType string, content []byte) (*Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	dir := filepath.Join(m.baseDir, missionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("artifact: failed to create mission dir: %w", err)
	}

	filePath := filepath.Join(dir, name)
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return nil, fmt.Errorf("artifact: failed to write file: %w", err)
	}

	record := Record{
		ID:        fmt.Sprintf("%s_%s_%s", missionID, taskID, name),
		MissionID: missionID,
		TaskID:    taskID,
		Name:      name,
		Path:      filePath,
		Type:      artifactType,
		Size:      int64(len(content)),
		CreatedAt: time.Now(),
	}

	m.index[missionID] = append(m.index[missionID], record)

	// Persist index
	m.saveIndex(missionID)

	return &record, nil
}

// List returns all artifacts for a mission.
func (m *Manager) List(missionID string) []Record {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.index[missionID]
}

// Read returns the content of an artifact file.
func (m *Manager) Read(missionID, name string) ([]byte, error) {
	path := filepath.Join(m.baseDir, missionID, name)
	return os.ReadFile(path)
}

func (m *Manager) saveIndex(missionID string) {
	indexPath := filepath.Join(m.baseDir, missionID, "_index.json")
	data, _ := json.MarshalIndent(m.index[missionID], "", "  ")
	os.WriteFile(indexPath, data, 0644)
}
```

## Rules
1. **Mission Isolation**: Artifacts stored under `baseDir/<missionID>/<filename>`. No cross-mission file access.
2. **Metadata Index**: `_index.json` persisted alongside artifacts for fast listing without directory scanning.

## Verify
```bash
go build ./modules/artifact/...
```

## Checklist
- [ ] File exists, stores artifacts with metadata indexing
- [ ] Mission-isolated directories
- [ ] `go build ./modules/artifact/...` passes
