# Micro-Task 6.15: Create modules/mission/manager.go

## Info
- **File**: `modules/mission/manager.go`
- **Package**: `mission`
- **Depends on**: 5.01 (mission struct)
- **Time**: 20 min
- **Verify**: `go build ./modules/mission/...`

## Purpose
Implements the mission CRUD manager with interface-backed storage for swappable backends (SQLite, PostgreSQL, in-memory).

## EXACT code to create

```go
// Package mission provides mission lifecycle management and persistence.
package mission

import (
	"fmt"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
)

// MissionRecord represents a persisted mission with execution metadata.
type MissionRecord struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Status      string        `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Duration    time.Duration `json:"duration"`
	TotalTokens int           `json:"total_tokens"`
	Tasks       []TaskRecord  `json:"tasks,omitempty"`
}

// TaskRecord represents a persisted task within a mission.
type TaskRecord struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Agent    string        `json:"agent"`
	Status   string        `json:"status"`
	Duration time.Duration `json:"duration"`
	Output   string        `json:"output,omitempty"`
	Error    string        `json:"error,omitempty"`
}

// Store defines the persistence interface for missions.
type Store interface {
	Create(record *MissionRecord) error
	Get(id string) (*MissionRecord, error)
	List() ([]*MissionRecord, error)
	UpdateStatus(id string, status string) error
	UpdateTokens(id string, tokens int) error
	AddTask(missionID string, task TaskRecord) error
	Delete(id string) error
	Close() error
}

// Manager provides high-level mission lifecycle operations.
type Manager struct {
	store Store
}

// NewManager constructs a new mission Manager.
func NewManager(store Store) *Manager {
	return &Manager{store: store}
}

// Create initializes and persists a new mission record.
func (m *Manager) Create(title, description string) (*MissionRecord, error) {
	record := &MissionRecord{
		ID:          string(contracts.NewTaskID()),
		Title:       title,
		Description: description,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := m.store.Create(record); err != nil {
		return nil, fmt.Errorf("mission/manager: failed to create: %w", err)
	}
	return record, nil
}

// Get retrieves a mission by ID.
func (m *Manager) Get(id string) (*MissionRecord, error) {
	record, err := m.store.Get(id)
	if err != nil {
		return nil, fmt.Errorf("mission/manager: failed to get %q: %w", id, err)
	}
	return record, nil
}

// List returns all missions ordered by creation time (newest first).
func (m *Manager) List() ([]*MissionRecord, error) {
	return m.store.List()
}

// UpdateStatus changes mission status and updates the timestamp.
func (m *Manager) UpdateStatus(id, status string) error {
	return m.store.UpdateStatus(id, status)
}

// RecordTask adds a completed task record to the mission.
func (m *Manager) RecordTask(missionID string, task TaskRecord) error {
	return m.store.AddTask(missionID, task)
}
```

## Rules
1. **Interface-Backed Store**: ALL persistence goes through the `Store` interface. Manager NEVER accesses database directly.
2. **ID Generation**: Use `contracts.NewTaskID()` for consistent ID format across the system.
3. **Timestamps**: Set `CreatedAt` and `UpdatedAt` at creation time. `UpdatedAt` refreshed on every mutation.

## Pitfalls

### Pitfall 1: Tight-coupling Manager to SQLite
```go
// WRONG:
type Manager struct { db *sql.DB } // Can't switch to PostgreSQL or in-memory for testing

// CORRECT:
type Manager struct { store Store } // Swap implementations freely
```

## Verify
```bash
go build ./modules/mission/...
```

## Checklist
- [ ] File `modules/mission/manager.go` exists
- [ ] `Store` interface with CRUD + Close
- [ ] `Manager` wraps Store with business logic
- [ ] Mission IDs generated via contracts
- [ ] `go build ./modules/mission/...` passes
