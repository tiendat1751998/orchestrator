# Micro-Task 5.20: Create kernel/resilience/recovery.go

## Info
- **File**: `kernel/resilience/recovery.go`
- **Package**: `resilience`
- **Depends on**: 5.19
- **Time**: 20 min
- **Verify**: `go build ./kernel/resilience/...`

## Purpose
Implements the plan checkpoints persistency manager (`CheckpointStore` and serialization helpers) to restore and resume execution state after system crashes.

## EXACT code to create

```go
package resilience

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Checkpoint represents saved status of a mission execution.
type Checkpoint struct {
	MissionID string            `json:"mission_id"`
	State     string            `json:"state"`
	Results   map[string]string `json:"results"`
}

// CheckpointStore saves checkpoints to disk.
// Thread-safe.
type CheckpointStore struct {
	mu      sync.Mutex
	dataDir string
}

// NewCheckpointStore constructs a CheckpointStore.
func NewCheckpointStore(dataDir string) (*CheckpointStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("recovery: failed to create checkpoint directory: %w", err)
	}
	return &CheckpointStore{dataDir: dataDir}, nil
}

// Save serializes and writes checkpoints atomically.
func (cs *CheckpointStore) Save(c *Checkpoint) error {
	if c == nil || c.MissionID == "" {
		return fmt.Errorf("recovery: invalid checkpoint parameters")
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	targetPath := filepath.Join(cs.dataDir, fmt.Sprintf("%s.json", c.MissionID))
	tempPath := targetPath + ".tmp"

	bytes, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("recovery: failed to marshal checkpoint: %w", err)
	}

	// Write to temp file first to prevent corruption
	if err := os.WriteFile(tempPath, bytes, 0644); err != nil {
		return fmt.Errorf("recovery: failed to write temp checkpoint: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, targetPath); err != nil {
		_ = os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("recovery: failed to rename checkpoint: %w", err)
	}

	return nil
}

// Load reads and deserializes target checkpoints.
func (cs *CheckpointStore) Load(missionID string) (*Checkpoint, error) {
	if missionID == "" {
		return nil, fmt.Errorf("recovery: invalid mission ID")
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	targetPath := filepath.Join(cs.dataDir, fmt.Sprintf("%s.json", missionID))

	bytes, err := os.ReadFile(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No checkpoint exists
		}
		return nil, fmt.Errorf("recovery: failed to read checkpoint file: %w", err)
	}

	var c Checkpoint
	if err := json.Unmarshal(bytes, &c); err != nil {
		return nil, fmt.Errorf("recovery: failed to unmarshal checkpoint: %w", err)
	}

	return &c, nil
}
```

## Pitfalls

### Pitfall 1: Corrupting checkpoint files on crashes during write
```go
// WRONG:
os.WriteFile(targetPath, bytes, 0644) // If writing gets interrupted mid-way, the checkpoint file is corrupted!
```
Writing directly to the target checkpoint file can leave it in a corrupted state if the program crashes. Write to a temporary file first and rename it.

### Pitfall 2: Locking stores during slow disk writes
Acquiring global locks during slow disk writes blocks concurrent read queries on other missions. Limit locks to atomic write windows.

## Verify
```bash
go build ./kernel/resilience/...
# Expected: clean compilation without errors
```

## Checklist
- [ ] File exists at `kernel/resilience/recovery.go`
- [ ] Package name is `resilience`
- [ ] All exported types have Godoc
- [ ] Checkpoint files are written atomically via temp files and renames
- [ ] Missing checkpoint files return nil result without throwing errors
- [ ] Build command passes
