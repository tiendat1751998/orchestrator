package runtime

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/kernel/registry"
)

type logRecord struct {
	Level slog.Level
	Msg   string
}

type testLogHandler struct {
	mu      sync.Mutex
	records []logRecord
}

func (h *testLogHandler) Enabled(ctx context.Context, level slog.Level) bool { return true }
func (h *testLogHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, logRecord{Level: r.Level, Msg: r.Message})
	return nil
}
func (h *testLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *testLogHandler) WithGroup(name string) slog.Handler       { return h }

func TestExecutor_GracefulDegradation(t *testing.T) {
	reg := registry.New(nil)
	bus := &mockEventBus{}

	var primaryCalls int32
	primaryAgent := &mockAgent{
		mockPlugin: mockPlugin{name: "primary-agent"},
		caps:       []agent.Capability{"test-task-type"},
		executeFn: func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
			atomic.AddInt32(&primaryCalls, 1)
			return nil, contracts.NewRetryableError(errors.New("transient error"), 1*time.Millisecond)
		},
	}

	var fallbackCalls int32
	fallbackAgent := &mockAgent{
		mockPlugin: mockPlugin{name: "fallback-agent"},
		caps:       []agent.Capability{"test-task-type"},
		executeFn: func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
			atomic.AddInt32(&fallbackCalls, 1)
			return &agent.Result{
				TaskID: task.ID,
				Status: contracts.StatusSuccess,
				Output: "fallback success",
			}, nil
		},
	}

	if err := reg.Register(primaryAgent); err != nil {
		t.Fatalf("failed to register primary: %v", err)
	}
	if err := reg.Register(fallbackAgent); err != nil {
		t.Fatalf("failed to register fallback: %v", err)
	}

	executor := NewExecutor(reg, bus, nil, ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
	})

	task := &agent.Task{
		ID:   contracts.TaskID("tsk-fallback"),
		Name: "fallback task",
		Type: "test-task-type",
	}

	res, err := executor.ExecuteTask(context.Background(), task)
	if err != nil {
		t.Fatalf("expected success with fallback, got: %v", err)
	}

	if res.Output != "fallback success" {
		t.Errorf("got %q, want %q", res.Output, "fallback success")
	}

	if atomic.LoadInt32(&primaryCalls) != 1 {
		t.Errorf("expected 1 call to primary, got %d", primaryCalls)
	}
	if atomic.LoadInt32(&fallbackCalls) != 1 {
		t.Errorf("expected 1 call to fallback, got %d", fallbackCalls)
	}
}

func TestExecutor_NonRetryableHalts(t *testing.T) {
	reg := registry.New(nil)
	bus := &mockEventBus{}

	// Primary fails with a non-retryable error
	primaryAgent := &mockAgent{
		mockPlugin: mockPlugin{name: "primary-agent"},
		caps:       []agent.Capability{"test-task-type"},
		executeFn: func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
			return nil, errors.New("non-retryable error")
		},
	}

	var fallbackCalls int32
	fallbackAgent := &mockAgent{
		mockPlugin: mockPlugin{name: "fallback-agent"},
		caps:       []agent.Capability{"test-task-type"},
		executeFn: func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
			atomic.AddInt32(&fallbackCalls, 1)
			return &agent.Result{
				TaskID: task.ID,
				Status: contracts.StatusSuccess,
				Output: "fallback success",
			}, nil
		},
	}

	if err := reg.Register(primaryAgent); err != nil {
		t.Fatalf("failed to register primary: %v", err)
	}
	if err := reg.Register(fallbackAgent); err != nil {
		t.Fatalf("failed to register fallback: %v", err)
	}

	executor := NewExecutor(reg, bus, nil, ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
	})

	task := &agent.Task{
		ID:   contracts.TaskID("tsk-nonretry"),
		Name: "non-retry task",
		Type: "test-task-type",
	}

	_, err := executor.ExecuteTask(context.Background(), task)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if atomic.LoadInt32(&fallbackCalls) != 0 {
		t.Errorf("fallback should not be called on fatal error, but got %d calls", fallbackCalls)
	}
}

func TestRuntime_InFlightCancellation(t *testing.T) {
	reg := registry.New(nil)
	bus := &mockEventBus{}

	execStarted := make(chan struct{})

	mockAgent := &mockAgent{
		mockPlugin: mockPlugin{name: "test-agent"},
		caps:       []agent.Capability{"test-task-type"},
		executeFn: func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
			close(execStarted)
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}

	if err := reg.Register(mockAgent); err != nil {
		t.Fatalf("register agent: %v", err)
	}

	rt := New(reg, bus, nil, Config{
		MaxWorkers: 1,
	})

	if err := rt.Start(nil); err != nil {
		t.Fatalf("start runtime: %v", err)
	}

	taskCtx, cancelTask := context.WithCancel(context.Background())
	task := &agent.Task{
		ID:   contracts.TaskID("tsk-cancel"),
		Name: "cancellable task",
		Type: "test-task-type",
	}

	if err := rt.Dispatch(taskCtx, task); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	<-execStarted
	cancelTask() // Cancel context of the in-flight task execution

	stopCtx, cancelStop := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelStop()
	rt.Stop(stopCtx)
}

func TestRuntime_StopLeakDetection(t *testing.T) {
	handler := &testLogHandler{}
	logger := slog.New(handler)

	reg := registry.New(logger)
	bus := &mockEventBus{}

	mockAgent := &mockAgent{
		mockPlugin: mockPlugin{name: "hang-agent"},
		caps:       []agent.Capability{"test-task-type"},
		executeFn: func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
			// Do nothing, just simulate task executing
			time.Sleep(100 * time.Millisecond)
			return &agent.Result{TaskID: task.ID, Status: contracts.StatusSuccess}, nil
		},
	}

	if err := reg.Register(mockAgent); err != nil {
		t.Fatalf("register agent: %v", err)
	}

	rt := New(reg, bus, logger, Config{
		MaxWorkers: 1,
	})

	if err := rt.Start(nil); err != nil {
		t.Fatalf("start runtime: %v", err)
	}

	task := &agent.Task{
		ID:   contracts.TaskID("tsk-hang"),
		Name: "hanging task",
		Type: "test-task-type",
	}

	if err := rt.Dispatch(context.Background(), task); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	// Wait a tiny bit to ensure worker started running
	time.Sleep(10 * time.Millisecond)

	// Stop with zero/immediate deadline context so it times out, leaving workers active.
	stopCtx, cancelStop := context.WithCancel(context.Background())
	cancelStop() // immediately cancelled context

	err := rt.Stop(stopCtx)
	if err != nil {
		t.Fatalf("expected nil from stop, got: %v", err)
	}

	handler.mu.Lock()
	var foundWorkerLeak bool
	for _, rec := range handler.records {
		if rec.Level == slog.LevelError && strings.Contains(rec.Msg, "resource leak: worker goroutines are still active") {
			foundWorkerLeak = true
		}
	}
	handler.mu.Unlock()

	if !foundWorkerLeak {
		t.Error("expected error log for worker goroutine leak, but none was found")
	}
}

func TestRuntime_StopUndrainedResults(t *testing.T) {
	handler := &testLogHandler{}
	logger := slog.New(handler)

	reg := registry.New(logger)
	bus := &mockEventBus{}

	rt := New(reg, bus, logger, Config{
		MaxWorkers: 1,
	})

	// Manually set as running so we can use it, or just use Start/Stop normally.
	// We can write a result directly to dispatcher results channel.
	rt.dispatcher.results <- TaskResult{
		TaskID: contracts.TaskID("tsk-undrained"),
		Error:  errors.New("some error"),
	}

	// We set running to true so Stop doesn't no-op.
	rt.running = true

	stopCtx := context.Background()
	err := rt.Stop(stopCtx)
	if err != nil {
		t.Fatalf("Stop: %v", err)
	}

	handler.mu.Lock()
	var foundUndrained bool
	for _, rec := range handler.records {
		if rec.Level == slog.LevelWarn && strings.Contains(rec.Msg, "resource warning: undrained task results remaining in channel") {
			foundUndrained = true
		}
	}
	handler.mu.Unlock()

	if !foundUndrained {
		t.Error("expected warn log for undrained task results, but none was found")
	}
}
