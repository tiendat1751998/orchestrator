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
