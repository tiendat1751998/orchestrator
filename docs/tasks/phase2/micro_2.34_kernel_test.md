# Micro-Task 2.34: Create kernel/kernel_test.go

## Info
- **File**: `kernel/kernel_test.go`
- **Package**: `kernel_test`
- **Depends on**: 2.31-2.33
- **Time**: 25 min
- **Verify**: `go test -v -race -count=1 ./kernel/...`

## Purpose
Integration tests for kernel: state machine, bootstrap, Start/Stop lifecycle.

## EXACT code to create

```go
package kernel_test

import (
	"context"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/kernel"
	"github.com/tiendat1751998/orchestrator/kernel/config"
)

// =============================================================================
// Test Config Helper
// =============================================================================

func testConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.Orchestrator.Name = "test-kernel"
	cfg.Orchestrator.LogLevel = "error" // Suppress logs in tests
	cfg.Orchestrator.MaxConcurrentTasks = 2
	cfg.Providers.Configs["antigravity"] = config.ProviderEntry{
		Type:   "cli",
		Model:  "test-model",
		Binary: "echo", // Use 'echo' as a dummy binary
	}
	return cfg
}

// =============================================================================
// Mock Agent for Kernel Tests
// =============================================================================

type kernelMockAgent struct {
	name         string
	initCalled   bool
	startCalled  bool
	stopCalled   bool
	healthErr    error
}

func (m *kernelMockAgent) Name() string    { return m.name }
func (m *kernelMockAgent) Type() string    { return plugin.TypeAgent }
func (m *kernelMockAgent) Version() string { return "1.0.0" }

func (m *kernelMockAgent) Init(_ context.Context, _ map[string]any) error {
	m.initCalled = true
	return nil
}
func (m *kernelMockAgent) Start(_ context.Context) error {
	m.startCalled = true
	return nil
}
func (m *kernelMockAgent) Stop(_ context.Context) error {
	m.stopCalled = true
	return nil
}
func (m *kernelMockAgent) Health(_ context.Context) error {
	return m.healthErr
}

func (m *kernelMockAgent) Capabilities() []agent.Capability {
	return []agent.Capability{agent.CapabilityCodeGeneration}
}
func (m *kernelMockAgent) CanHandle(task *agent.Task) bool {
	return task.Type == string(agent.CapabilityCodeGeneration)
}
func (m *kernelMockAgent) Execute(ctx context.Context, task *agent.Task) (*agent.Result, error) {
	return &agent.Result{TaskID: task.ID, Status: "completed", Output: "test output"}, nil
}
func (m *kernelMockAgent) Manifest() agent.Manifest {
	return agent.Manifest{Name: m.name}
}

// =============================================================================
// State Machine Tests
// =============================================================================

func TestStateMachine_InitialState(t *testing.T) {
	sm := kernel.NewStateMachine()
	if sm.Current() != kernel.StateCreated {
		t.Errorf("initial state: got %v, want %v", sm.Current(), kernel.StateCreated)
	}
}

func TestStateMachine_ValidTransition(t *testing.T) {
	sm := kernel.NewStateMachine()

	// Created → Initializing
	if err := sm.Transition(kernel.StateInitializing); err != nil {
		t.Fatalf("Created → Initializing: %v", err)
	}
	if sm.Current() != kernel.StateInitializing {
		t.Errorf("state: got %v, want %v", sm.Current(), kernel.StateInitializing)
	}

	// Initializing → Running
	if err := sm.Transition(kernel.StateRunning); err != nil {
		t.Fatalf("Initializing → Running: %v", err)
	}

	// Running → ShuttingDown
	if err := sm.Transition(kernel.StateShuttingDown); err != nil {
		t.Fatalf("Running → ShuttingDown: %v", err)
	}

	// ShuttingDown → Stopped
	if err := sm.Transition(kernel.StateStopped); err != nil {
		t.Fatalf("ShuttingDown → Stopped: %v", err)
	}
}

func TestStateMachine_InvalidTransition(t *testing.T) {
	sm := kernel.NewStateMachine()

	// Created → Running (skip Initializing)
	err := sm.Transition(kernel.StateRunning)
	if err == nil {
		t.Error("expected error for Created → Running")
	}
}

func TestStateMachine_StoppedIsTerminal(t *testing.T) {
	sm := kernel.NewStateMachine()
	sm.Transition(kernel.StateInitializing)
	sm.Transition(kernel.StateStopped)

	// Cannot transition from Stopped
	err := sm.Transition(kernel.StateCreated)
	if err == nil {
		t.Error("expected error: Stopped is terminal")
	}
}

func TestStateMachine_DoubleTransition(t *testing.T) {
	sm := kernel.NewStateMachine()
	sm.Transition(kernel.StateInitializing)
	sm.Transition(kernel.StateRunning)

	// Cannot transition to Running again
	err := sm.Transition(kernel.StateRunning)
	if err == nil {
		t.Error("expected error: already in Running, cannot transition to Running")
	}
}

func TestStateMachine_IsRunning(t *testing.T) {
	sm := kernel.NewStateMachine()

	if sm.IsRunning() {
		t.Error("should not be running initially")
	}

	sm.Transition(kernel.StateInitializing)
	sm.Transition(kernel.StateRunning)

	if !sm.IsRunning() {
		t.Error("should be running after transition to Running")
	}
}

func TestState_String(t *testing.T) {
	tests := []struct {
		state kernel.State
		want  string
	}{
		{kernel.StateCreated, "created"},
		{kernel.StateInitializing, "initializing"},
		{kernel.StateRunning, "running"},
		{kernel.StateShuttingDown, "shutting_down"},
		{kernel.StateStopped, "stopped"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

// =============================================================================
// Kernel Creation Tests
// =============================================================================

func TestKernel_New_NilConfig(t *testing.T) {
	_, err := kernel.New(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestKernel_New_Success(t *testing.T) {
	cfg := testConfig()
	k, err := kernel.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if k.State() != kernel.StateCreated {
		t.Errorf("state: got %v, want Created", k.State())
	}
}

// =============================================================================
// Plugin Registration Tests
// =============================================================================

func TestKernel_RegisterPlugin(t *testing.T) {
	cfg := testConfig()
	k, _ := kernel.New(cfg)

	agent := &kernelMockAgent{name: "backend"}
	err := k.RegisterPlugin(agent)
	if err != nil {
		t.Fatalf("RegisterPlugin: %v", err)
	}
}

// =============================================================================
// Start / Stop Lifecycle Tests
// =============================================================================

func TestKernel_StartStop(t *testing.T) {
	cfg := testConfig()
	k, _ := kernel.New(cfg)

	agent := &kernelMockAgent{name: "backend"}
	k.RegisterPlugin(agent)

	// Start
	ctx := context.Background()
	err := k.Start(ctx)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	if k.State() != kernel.StateRunning {
		t.Errorf("state after Start: got %v, want Running", k.State())
	}

	if !agent.initCalled {
		t.Error("agent Init was not called")
	}
	if !agent.startCalled {
		t.Error("agent Start was not called")
	}

	// Stop
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = k.Stop(stopCtx)
	if err != nil {
		t.Fatalf("Stop: %v", err)
	}

	if k.State() != kernel.StateStopped {
		t.Errorf("state after Stop: got %v, want Stopped", k.State())
	}

	if !agent.stopCalled {
		t.Error("agent Stop was not called")
	}
}

func TestKernel_DoubleStart(t *testing.T) {
	cfg := testConfig()
	k, _ := kernel.New(cfg)

	k.RegisterPlugin(&kernelMockAgent{name: "agent1"})

	k.Start(context.Background())
	defer k.Stop(context.Background())

	err := k.Start(context.Background())
	if err == nil {
		t.Error("expected error on double start")
	}
}

func TestKernel_StopBeforeStart(t *testing.T) {
	cfg := testConfig()
	k, _ := kernel.New(cfg)

	// Stop before start should not panic (but may error on state transition)
	err := k.Stop(context.Background())
	// Stopping a kernel that was never started — should be a no-op or error
	_ = err // Accept either behavior
}

func TestKernel_StopIdempotent(t *testing.T) {
	cfg := testConfig()
	k, _ := kernel.New(cfg)

	k.RegisterPlugin(&kernelMockAgent{name: "agent1"})
	k.Start(context.Background())

	k.Stop(context.Background())
	k.Stop(context.Background()) // Second stop should be no-op, not panic
}

func TestKernel_Accessors(t *testing.T) {
	cfg := testConfig()
	k, _ := kernel.New(cfg)

	if k.EventBus() == nil {
		t.Error("EventBus() should not be nil")
	}
	if k.Registry() == nil {
		t.Error("Registry() should not be nil")
	}
	if k.Logger() == nil {
		t.Error("Logger() should not be nil")
	}
	if k.Config() == nil {
		t.Error("Config() should not be nil")
	}
}
```

## Pitfalls

### Pitfall 1: testConfig() creates valid config
The test config must pass validation. If validation fails, Start() fails before testing the real logic.
Set all required fields: providers.default, provider entry with type+model+binary.

### Pitfall 2: Mock must implement BOTH plugin.Plugin AND agent.Agent
The registry checks `p.Type() == plugin.TypeAgent` and then type-asserts to `agent.Agent`.
If the mock doesn't implement `agent.Agent` → registration fails.

### Pitfall 3: Suppress logs in tests
```go
cfg.Orchestrator.LogLevel = "error" // Suppress info/debug during tests
```
Without this, test output is flooded with log messages.

### Pitfall 4: Stop timeout in tests
```go
stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
```
If Stop hangs due to a bug, the test times out after 5s instead of hanging forever.

## Verify
```bash
go test -v -race -count=1 ./kernel/...
# Expected: ALL PASS
# This runs tests for ALL kernel sub-packages
```

## Checklist
- [ ] File `kernel/kernel_test.go` exists
- [ ] Package: `kernel_test`
- [ ] testConfig() helper with valid config
- [ ] kernelMockAgent implementing plugin.Plugin + agent.Agent
- [ ] ≥ 15 test functions
- [ ] State machine: initial state, valid transitions, invalid transition, terminal state
- [ ] State machine: double transition, IsRunning, State.String
- [ ] Kernel: New with nil config, New success
- [ ] Kernel: RegisterPlugin
- [ ] Kernel: Start/Stop lifecycle (init/start/stop called on agent)
- [ ] Kernel: double start, stop before start, stop idempotent
- [ ] Kernel: accessor methods (EventBus, Registry, Logger, Config)
- [ ] `go test -v -race ./kernel/...` ALL PASS
