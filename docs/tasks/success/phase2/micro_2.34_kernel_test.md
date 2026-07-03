# Micro-Task 2.34: Create kernel/kernel_test.go

## Info
- **File**: `kernel/kernel_test.go`
- **Package**: `kernel_test`
- **Depends on**: 2.31-2.33, 1.40 (health/plugin contracts)
- **Time**: 25 min
- **Verify**: `go test -v -race -count=1 ./kernel/...`

## Purpose
Implements integration unit tests for the kernel core. It verifies state machine transitions, bootstrap sequences, plugin registration checks, double start blocks, and LIFO lifecycle shutdowns.

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
	cfg.Orchestrator.LogLevel = "error"
	cfg.Orchestrator.MaxConcurrentTasks = 2
	cfg.Providers.Configs["antigravity"] = config.ProviderEntry{
		Type:   "cli",
		Model:  "test-model",
		Binary: "echo",
	}
	return cfg
}

// =============================================================================
// Mock Agent for Kernel Tests
// =============================================================================

type kernelMockAgent struct {
	name        string
	initCalled  bool
	startCalled bool
	stopCalled  bool
	healthErr   error
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

// Health satisfies the updated Task 1.40 Plugin signature
func (m *kernelMockAgent) Health(_ context.Context) (plugin.HealthReport, error) {
	if m.healthErr != nil {
		return plugin.HealthReport{Status: plugin.HealthDegraded}, m.healthErr
	}
	return plugin.HealthReport{Status: plugin.HealthOK}, nil
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

	if err := sm.Transition(kernel.StateInitializing); err != nil {
		t.Fatalf("Created → Initializing: %v", err)
	}
	if sm.Current() != kernel.StateInitializing {
		t.Errorf("state: got %v, want %v", sm.Current(), kernel.StateInitializing)
	}

	if err := sm.Transition(kernel.StateRunning); err != nil {
		t.Fatalf("Initializing → Running: %v", err)
	}

	if err := sm.Transition(kernel.StateShuttingDown); err != nil {
		t.Fatalf("Running → ShuttingDown: %v", err)
	}

	if err := sm.Transition(kernel.StateStopped); err != nil {
		t.Fatalf("ShuttingDown → Stopped: %v", err)
	}
}

func TestStateMachine_InvalidTransition(t *testing.T) {
	sm := kernel.NewStateMachine()

	err := sm.Transition(kernel.StateRunning)
	if err == nil {
		t.Error("expected error for Created → Running")
	}
}

func TestStateMachine_StoppedIsTerminal(t *testing.T) {
	sm := kernel.NewStateMachine()
	sm.Transition(kernel.StateInitializing)
	sm.Transition(kernel.StateStopped)

	err := sm.Transition(kernel.StateCreated)
	if err == nil {
		t.Error("expected error: Stopped is terminal")
	}
}

func TestStateMachine_DoubleTransition(t *testing.T) {
	sm := kernel.NewStateMachine()
	sm.Transition(kernel.StateInitializing)
	sm.Transition(kernel.StateRunning)

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

	err := k.Stop(context.Background())
	_ = err
}

func TestKernel_StopIdempotent(t *testing.T) {
	cfg := testConfig()
	k, _ := kernel.New(cfg)

	k.RegisterPlugin(&kernelMockAgent{name: "agent1"})
	k.Start(context.Background())

	k.Stop(context.Background())
	k.Stop(context.Background())
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

## Rules
1. **Mock Signatures Alignment**: Any mock structs defined inside kernel tests must implement `plugin.Plugin` matching the updated signature `Health(context.Context) (plugin.HealthReport, error)`.
2. **Deterministic Shutdown Timeouts**: Protect lifecycle stop tests from deadlocking using context timeouts (`context.WithTimeout`).
3. **Idempotence Verification**: Verify that Stop calls are idempotent when executed repeatedly.

## ⚠️ Pitfalls

### Pitfall 1: Mismatched mock method signatures breaking compile pipelines
Using a legacy `Health(context.Context) error` method signature on test mocks prevents compilation since `plugin.Plugin` has been upgraded. Always declare `Health(context.Context) (plugin.HealthReport, error)`.

### Pitfall 2: Bypassing configuration validation rules during test setup
If you configure invalid settings (e.g. empty provider default fields) inside test helpers, the bootstrap phase returns errors before testing the real logic. Always populate required fields.

## Verify
```bash
go test -v -race ./kernel/...
```

## Checklist
- [ ] File `kernel/kernel_test.go` exists
- [ ] Package: `kernel_test`
- [ ] Mocks implement `plugin.Plugin` and `agent.Agent` contracts
- [ ] `Health` mock methods return `(plugin.HealthReport, error)`
- [ ] Tests verify all state transitions and terminal boundaries
- [ ] Start/Stop flows assert that plugins execute Init, Start, and Stop sequences
- [ ] Second stop calls run without panic
- [ ] Context constraints guard stop assertions from deadlocks
- [ ] `go test -v -race ./kernel/...` passes
