# Micro-Task 5.11: Create kernel/orchestrator/supervisor.go

## Info
- **File**: `kernel/orchestrator/supervisor.go`
- **Package**: `orchestrator`
- **Depends on**: 5.10
- **Time**: 20 min
- **Verify**: `go build ./kernel/orchestrator/...`

## Purpose
Implements the background monitoring supervisor daemon (`Supervisor` and health checkers) to detect stale tasks and check plugin states.

## EXACT code to create

```go
package orchestrator

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// ActiveTask represents a running task under observation.
type ActiveTask struct {
	ID        string
	StartedAt time.Time
	Timeout   time.Duration
}

// Supervisor monitors running tasks and detects timeouts or crashes.
// Thread-safe.
type Supervisor struct {
	mu          sync.RWMutex
	activeTasks map[string]*ActiveTask
	logger      *slog.Logger
}

// NewSupervisor constructs a new Supervisor.
func NewSupervisor(logger *slog.Logger) *Supervisor {
	return &Supervisor{
		activeTasks: make(map[string]*ActiveTask),
		logger:      logger,
	}
}

// Register adds a running task to the supervision list.
func (s *Supervisor) Register(taskID string, timeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.activeTasks[taskID] = &ActiveTask{
		ID:        taskID,
		StartedAt: time.Now(),
		Timeout:   timeout,
	}
}

// Deregister removes a completed task from supervision.
func (s *Supervisor) Deregister(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.activeTasks, taskID)
}

// Start runs the periodic scanner loop to check for timed-out tasks.
func (s *Supervisor) Start(ctx context.Context, checkInterval time.Duration) {
	if checkInterval <= 0 {
		checkInterval = 5 * time.Second
	}

	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.checkTimeouts()
			}
		}
	}()
}

func (s *Supervisor) checkTimeouts() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, t := range s.activeTasks {
		if t.Timeout > 0 && now.Sub(t.StartedAt) > t.Timeout {
			if s.logger != nil {
				s.logger.Error("supervisor: task exceeded timeout limit", "task_id", id, "duration", now.Sub(t.StartedAt))
			}
			// Deregister timed-out task
			delete(s.activeTasks, id)
		}
	}
}
```

## Pitfalls

### Pitfall 1: Unsynchronized checks of task states
```go
// WRONG:
func (s *Supervisor) Register(taskID string) {
    s.activeTasks[taskID] = ... // Modifies map without lock! Crashes when checkTimeouts runs concurrently.
}

// CORRECT:
s.mu.Lock()
defer s.mu.Unlock()
```
Modifying maps from concurrent threads without using mutex locks triggers race conditions. Protect all map actions with locks.

### Pitfall 2: High CPU usage from short check intervals
Setting scan loops to extremely short durations (e.g. 50ms) wastes CPU resources. Default to 5-second intervals.

## Verify
```bash
go build ./kernel/orchestrator/...
# Expected: clean compilation without errors
```

## Checklist
- [ ] File exists at `kernel/orchestrator/supervisor.go`
- [ ] Package name is `orchestrator`
- [ ] All exported types have Godoc
- [ ] Task maps mutations are guarded under mutex locks
- [ ] Scanning loop runs in a background goroutine
- [ ] Build command passes
