package runtime

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/event"
	"github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/kernel/registry"
)

// =============================================================================
// Mocks
// =============================================================================

type mockPlugin struct {
	name string
}

func (m *mockPlugin) Name() string                                          { return m.name }
func (m *mockPlugin) Type() plugin.Type                                     { return plugin.TypeAgent }
func (m *mockPlugin) Version() string                                       { return "1.0.0" }
func (m *mockPlugin) Init(ctx context.Context, config map[string]any) error { return nil }
func (m *mockPlugin) Start(ctx context.Context) error                       { return nil }
func (m *mockPlugin) Stop(ctx context.Context) error                        { return nil }
func (m *mockPlugin) Health(ctx context.Context) (plugin.HealthReport, error) {
	return plugin.HealthReport{Status: plugin.HealthOK}, nil
}

type mockAgent struct {
	mockPlugin
	caps      []agent.Capability
	executeFn func(ctx context.Context, task *agent.Task) (*agent.Result, error)
}

func (m *mockAgent) Capabilities() []agent.Capability { return m.caps }
func (m *mockAgent) Role() string                     { return "Mock Agent" }
func (m *mockAgent) CanHandle(task *agent.Task) bool {
	for _, c := range m.caps {
		if string(c) == task.Type {
			return true
		}
	}
	return false
}

func (m *mockAgent) Execute(ctx context.Context, task *agent.Task) (*agent.Result, error) {
	if m.executeFn != nil {
		return m.executeFn(ctx, task)
	}
	return agent.SuccessResult(task.ID, m.name, "default success output"), nil
}

type mockEventBus struct {
	mu     sync.Mutex
	events []event.Event
}

func (m *mockEventBus) Publish(ctx context.Context, evt event.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, evt)
	return nil
}

func (m *mockEventBus) Subscribe(pattern string, handler func(event.Event)) (func(), error) {
	return func() {}, nil
}

func (m *mockEventBus) GetEvents() []event.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]event.Event, len(m.events))
	copy(copied, m.events)
	return copied
}

// =============================================================================
// Tests
// =============================================================================

func TestExecutor_Success(t *testing.T) {
	reg := registry.New(nil)
	a := &mockAgent{
		mockPlugin: mockPlugin{name: "test-agent"},
		caps:       []agent.Capability{"test-task-type"},
	}
	err := reg.Register(a)
	if err != nil {
		t.Fatalf("failed to register agent: %v", err)
	}

	bus := &mockEventBus{}
	executor := NewExecutor(reg, bus, slog.Default(), ExecutorConfig{
		DefaultTimeout: 2 * time.Second,
	})

	task := &agent.Task{
		ID:       contracts.TaskID("task-123"),
		Name:     "Test Task",
		Type:     "test-task-type",
		Priority: 5,
	}

	res, err := executor.ExecuteTask(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res == nil {
		t.Fatal("expected result to be non-nil")
	}

	if res.Output != "default success output" {
		t.Errorf("expected output 'default success output', got %q", res.Output)
	}

	// Verify events
	events := bus.GetEvents()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if events[0].Type != event.EventTaskStarted {
		t.Errorf("expected first event to be %q, got %q", event.EventTaskStarted, events[0].Type)
	}

	if events[1].Type != event.EventTaskCompleted {
		t.Errorf("expected second event to be %q, got %q", event.EventTaskCompleted, events[1].Type)
	}
}

func TestExecutor_NoAgentMatched(t *testing.T) {
	reg := registry.New(nil)
	bus := &mockEventBus{}
	executor := NewExecutor(reg, bus, slog.Default(), ExecutorConfig{})

	task := &agent.Task{
		ID:   contracts.TaskID("task-123"),
		Name: "Test Task",
		Type: "unknown-task-type",
	}

	res, err := executor.ExecuteTask(context.Background(), task)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if res != nil {
		t.Fatalf("expected nil result, got %v", res)
	}

	if !strings.Contains(err.Error(), "registry: no agent can handle task") {
		t.Errorf("expected error message to complain about agent matching, got: %v", err)
	}

	// Verify no events published
	if len(bus.GetEvents()) != 0 {
		t.Errorf("expected 0 events, got %d", len(bus.GetEvents()))
	}
}

func TestExecutor_AgentFailure(t *testing.T) {
	reg := registry.New(nil)
	expectedErr := errors.New("agent custom execution error")
	a := &mockAgent{
		mockPlugin: mockPlugin{name: "test-agent"},
		caps:       []agent.Capability{"test-task-type"},
		executeFn: func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
			return nil, expectedErr
		},
	}
	_ = reg.Register(a)

	bus := &mockEventBus{}
	executor := NewExecutor(reg, bus, nil, ExecutorConfig{})

	task := &agent.Task{
		ID:   contracts.TaskID("task-123"),
		Name: "Test Task",
		Type: "test-task-type",
	}

	res, err := executor.ExecuteTask(context.Background(), task)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped agent error, got: %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result, got %v", res)
	}

	// Verify events
	events := bus.GetEvents()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != event.EventTaskStarted {
		t.Errorf("expected first event to be %q, got %q", event.EventTaskStarted, events[0].Type)
	}
	if events[1].Type != event.EventTaskFailed {
		t.Errorf("expected second event to be %q, got %q", event.EventTaskFailed, events[1].Type)
	}
}

func TestExecutor_PanicRecovery(t *testing.T) {
	reg := registry.New(nil)
	a := &mockAgent{
		mockPlugin: mockPlugin{name: "test-agent"},
		caps:       []agent.Capability{"test-task-type"},
		executeFn: func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
			panic("something exploded!")
		},
	}
	_ = reg.Register(a)

	bus := &mockEventBus{}
	executor := NewExecutor(reg, bus, nil, ExecutorConfig{})

	task := &agent.Task{
		ID:   contracts.TaskID("task-123"),
		Name: "Test Task",
		Type: "test-task-type",
	}

	res, err := executor.ExecuteTask(context.Background(), task)
	if err == nil {
		t.Fatal("expected error from recovered panic, got nil")
	}
	if res != nil {
		t.Fatalf("expected nil result, got %v", res)
	}

	if !strings.Contains(err.Error(), "agent \"test-agent\" panicked: something exploded!") {
		t.Errorf("expected panic error message, got: %v", err)
	}
	if !strings.Contains(err.Error(), "executor_test.go") {
		t.Errorf("expected stack trace in panic error, got: %v", err)
	}

	// Verify events
	events := bus.GetEvents()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != event.EventTaskStarted {
		t.Errorf("expected first event to be %q, got %q", event.EventTaskStarted, events[0].Type)
	}
	if events[1].Type != event.EventTaskFailed {
		t.Errorf("expected second event to be %q, got %q", event.EventTaskFailed, events[1].Type)
	}
}

func TestExecutor_TimeoutEnforced(t *testing.T) {
	reg := registry.New(nil)
	a := &mockAgent{
		mockPlugin: mockPlugin{name: "test-agent"},
		caps:       []agent.Capability{"test-task-type"},
		executeFn: func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(1 * time.Second):
				return agent.SuccessResult(task.ID, "test-agent", "too late"), nil
			}
		},
	}
	_ = reg.Register(a)

	bus := &mockEventBus{}
	// Set default timeout very short (10ms)
	executor := NewExecutor(reg, bus, nil, ExecutorConfig{
		DefaultTimeout: 10 * time.Millisecond,
	})

	task := &agent.Task{
		ID:   contracts.TaskID("task-123"),
		Name: "Test Task",
		Type: "test-task-type",
	}

	res, err := executor.ExecuteTask(context.Background(), task)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if res != nil {
		t.Fatalf("expected nil result, got %v", res)
	}

	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("expected context deadline exceeded error, got: %v", err)
	}
}
