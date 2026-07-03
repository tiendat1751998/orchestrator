# Micro-Task 5.10: Create kernel/orchestrator/pipeline.go

## Info
- **File**: `kernel/orchestrator/pipeline.go`
- **Package**: `orchestrator`
- **Depends on**: 5.09
- **Time**: 15 min
- **Verify**: `go build ./kernel/orchestrator/...`

## Purpose
Declares the execution pipeline state machine (`State` and lifecycle hooks) to track mission progress within orchestration registries.

## EXACT code to create

```go
package orchestrator

import (
	"fmt"
	"sync"
)

// PipelineState represents the current state of a running mission.
type PipelineState string

const (
	StatePlanning    PipelineState = "planning"
	StateScheduling  PipelineState = "scheduling"
	StateExecuting   PipelineState = "executing"
	StateAggregating PipelineState = "aggregating"
	StateCompleted   PipelineState = "completed"
	StateFailed      PipelineState = "failed"
)

// PipelineManager tracks transitions between execution phases.
// Thread-safe.
type PipelineManager struct {
	mu    sync.RWMutex
	state PipelineState
}

// NewPipelineManager constructs a new PipelineManager.
func NewPipelineManager() *PipelineManager {
	return &PipelineManager{
		state: StatePlanning,
	}
}

// GetState returns the current state.
func (pm *PipelineManager) GetState() PipelineState {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.state
}

// Transition moves to the target state, validating the state change path.
func (pm *PipelineManager) Transition(to PipelineState) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Validate permitted state paths
	permitted := false
	switch pm.state {
	case StatePlanning:
		permitted = (to == StateScheduling || to == StateFailed)
	case StateScheduling:
		permitted = (to == StateExecuting || to == StateFailed)
	case StateExecuting:
		permitted = (to == StateAggregating || to == StateFailed)
	case StateAggregating:
		permitted = (to == StateCompleted || to == StateFailed)
	case StateCompleted, StateFailed:
		permitted = false // Terminal states
	}

	if !permitted {
		return fmt.Errorf("pipeline: invalid state transition from %q to %q", pm.state, to)
	}

	pm.state = to
	return nil
}
```

## Pitfalls

### Pitfall 1: Race conditions from unsynchronized status checks
```go
// WRONG:
if pm.state == StatePlanning { ... } // Reads mutable state field without acquiring lock.
```
Reading state variables directly inside worker threads without using locks returns stale values or triggers panics. Always wrap reads under read locks.

### Pitfall 2: Bypassing transition logic rules
Allowing direct mutations to state properties makes it easy to bypass validation pathways (like jumping from `Planning` directly to `Completed`), corrupting monitoring dashboards.

## Verify
```bash
go build ./kernel/orchestrator/...
# Expected: clean compilation without errors
```

## Checklist
- [ ] File exists at `kernel/orchestrator/pipeline.go`
- [ ] Package name is `orchestrator`
- [ ] All exported types have Godoc
- [ ] Status variables are protected under mutex locks
- [ ] State paths are validated during transitions
- [ ] Build command passes
