package orchestrator

import (
	"testing"
)

func TestPipelineManager_Transitions(t *testing.T) {
	pm := NewPipelineManager()
	if pm.GetState() != StatePlanning {
		t.Fatalf("expected initial state %q, got %q", StatePlanning, pm.GetState())
	}

	// Test a valid transition sequence
	transitions := []PipelineState{
		StateScheduling,
		StateExecuting,
		StateAggregating,
		StateCompleted,
	}

	for _, next := range transitions {
		if err := pm.Transition(next); err != nil {
			t.Fatalf("failed expected transition to %q: %v", next, err)
		}
		if pm.GetState() != next {
			t.Fatalf("expected state %q, got %q", next, pm.GetState())
		}
	}

	// Terminal state transitions should fail
	if err := pm.Transition(StatePlanning); err == nil {
		t.Fatal("expected transition from completed state to fail, but it succeeded")
	}
}

func TestPipelineManager_TransitionFailure(t *testing.T) {
	pm := NewPipelineManager()

	// Planning -> Executing is invalid
	if err := pm.Transition(StateExecuting); err == nil {
		t.Fatal("expected planning to executing to fail")
	}

	// Planning -> Failed is valid
	if err := pm.Transition(StateFailed); err != nil {
		t.Fatalf("failed transition planning to failed: %v", err)
	}

	// Terminal state transitions should fail
	if err := pm.Transition(StateScheduling); err == nil {
		t.Fatal("expected transition from failed state to fail")
	}
}
