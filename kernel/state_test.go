package kernel

import (
	"fmt"
	"sync"
	"testing"
)

func TestStateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{state: StateCreated, expected: "created"},
		{state: StateInitializing, expected: "initializing"},
		{state: StateRunning, expected: "running"},
		{state: StateShuttingDown, expected: "shutting_down"},
		{state: StateStopped, expected: "stopped"},
		{state: State(-1), expected: "unknown(-1)"},
		{state: State(100), expected: "unknown(100)"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("state_%d", tc.state), func(t *testing.T) {
			if actual := tc.state.String(); actual != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, actual)
			}
		})
	}
}

func TestStateMachine_InitialState(t *testing.T) {
	sm := NewStateMachine()
	if !sm.Is(StateCreated) {
		t.Errorf("expected initial state to be StateCreated, got %v", sm.Current())
	}
	if sm.IsRunning() {
		t.Error("expected IsRunning() to be false initially")
	}
	if sm.IsStopped() {
		t.Error("expected IsStopped() to be false initially")
	}
}

func TestStateMachine_ValidTransitions(t *testing.T) {
	t.Run("normal_lifecycle", func(t *testing.T) {
		sm := NewStateMachine()

		// Created -> Initializing
		if err := sm.Transition(StateInitializing); err != nil {
			t.Fatalf("unexpected transition error: %v", err)
		}
		if sm.Current() != StateInitializing {
			t.Errorf("expected state to be Initializing, got %v", sm.Current())
		}

		// Initializing -> Running
		if err := sm.Transition(StateRunning); err != nil {
			t.Fatalf("unexpected transition error: %v", err)
		}
		if !sm.IsRunning() {
			t.Errorf("expected state to be Running, got %v", sm.Current())
		}

		// Running -> ShuttingDown
		if err := sm.Transition(StateShuttingDown); err != nil {
			t.Fatalf("unexpected transition error: %v", err)
		}
		if sm.Current() != StateShuttingDown {
			t.Errorf("expected state to be ShuttingDown, got %v", sm.Current())
		}

		// ShuttingDown -> Stopped
		if err := sm.Transition(StateStopped); err != nil {
			t.Fatalf("unexpected transition error: %v", err)
		}
		if !sm.IsStopped() {
			t.Errorf("expected state to be Stopped, got %v", sm.Current())
		}
	})

	t.Run("init_failure_lifecycle", func(t *testing.T) {
		sm := NewStateMachine()

		// Created -> Initializing
		if err := sm.Transition(StateInitializing); err != nil {
			t.Fatalf("unexpected transition error: %v", err)
		}

		// Initializing -> Stopped (failure during initialization)
		if err := sm.Transition(StateStopped); err != nil {
			t.Fatalf("unexpected transition error: %v", err)
		}
		if !sm.IsStopped() {
			t.Errorf("expected state to be Stopped, got %v", sm.Current())
		}
	})
}

func TestStateMachine_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name string
		from []State
		to   State
	}{
		{
			name: "Created cannot transition to Running directly",
			from: []State{},
			to:   StateRunning,
		},
		{
			name: "Created cannot transition to Stopped directly",
			from: []State{},
			to:   StateStopped,
		},
		{
			name: "Initializing cannot transition to ShuttingDown",
			from: []State{StateInitializing},
			to:   StateShuttingDown,
		},
		{
			name: "Running cannot transition to Initializing",
			from: []State{StateInitializing, StateRunning},
			to:   StateInitializing,
		},
		{
			name: "Running cannot transition to Stopped directly",
			from: []State{StateInitializing, StateRunning},
			to:   StateStopped,
		},
		{
			name: "Stopped cannot transition to anything",
			from: []State{StateInitializing, StateStopped},
			to:   StateRunning,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sm := NewStateMachine()
			for _, step := range tc.from {
				if err := sm.Transition(step); err != nil {
					t.Fatalf("setup transition to %v failed: %v", step, err)
				}
			}

			err := sm.Transition(tc.to)
			if err == nil {
				t.Errorf("expected transition to fail from %v to %v, but it succeeded", sm.Current(), tc.to)
			}
		})
	}
}

func TestStateMachine_ConcurrentUse(t *testing.T) {
	sm := NewStateMachine()

	var wg sync.WaitGroup
	const workers = 10
	const iterations = 100

	// Concurrent readers of Current, Is, IsRunning, IsStopped
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = sm.Current()
				_ = sm.Is(StateRunning)
				_ = sm.IsRunning()
				_ = sm.IsStopped()
			}
		}()
	}

	// Single writer attempting transition at the same time
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Created -> Initializing
		_ = sm.Transition(StateInitializing)
		// Initializing -> Running
		_ = sm.Transition(StateRunning)
		// Running -> ShuttingDown
		_ = sm.Transition(StateShuttingDown)
		// ShuttingDown -> Stopped
		_ = sm.Transition(StateStopped)
	}()

	wg.Wait()
}
