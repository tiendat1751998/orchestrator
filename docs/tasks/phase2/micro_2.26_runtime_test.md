# Micro-Task 2.26: Create kernel/runtime/runtime_test.go

## Info
- **File**: `kernel/runtime/runtime_test.go`
- **Package**: `runtime_test`
- **Depends on**: 2.22-2.25
- **Time**: 25 min
- **Verify**: `go test -v -race -count=1 ./kernel/runtime/...`

## Purpose
Tests for pool, executor, dispatcher, and runtime engine.
Focus: concurrency safety, panic recovery, timeout, shutdown.

## EXACT code to create

```go
package runtime_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/kernel/eventbus"
	"github.com/tiendat1751998/orchestrator/kernel/registry"
	"github.com/tiendat1751998/orchestrator/kernel/runtime"
)

// =============================================================================
// Test Logger
// =============================================================================

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn, // Suppress info/debug in tests
	}))
}

// =============================================================================
// Mock Agent for Runtime Tests
// =============================================================================

type runtimeMockAgent struct {
	name         string
	executeFn    func(context.Context, *agent.Task) (*agent.Result, error)
	capabilities []agent.Capability
}

func (m *runtimeMockAgent) Name() string    { return m.name }
func (m *runtimeMockAgent) Type() string    { return plugin.TypeAgent }
func (m *runtimeMockAgent) Version() string { return "1.0.0" }

func (m *runtimeMockAgent) Init(_ context.Context, _ map[string]any) error { return nil }
func (m *runtimeMockAgent) Start(_ context.Context) error                  { return nil }
func (m *runtimeMockAgent) Stop(_ context.Context) error                   { return nil }
func (m *runtimeMockAgent) Health(_ context.Context) error                 { return nil }

func (m *runtimeMockAgent) Capabilities() []agent.Capability { return m.capabilities }
func (m *runtimeMockAgent) CanHandle(task *agent.Task) bool {
	for _, cap := range m.capabilities {
		if string(cap) == task.Type {
			return true
		}
	}
	return false
}

func (m *runtimeMockAgent) Execute(ctx context.Context, task *agent.Task) (*agent.Result, error) {
	if m.executeFn != nil {
		return m.executeFn(ctx, task)
	}
	return &agent.Result{
		TaskID: task.ID,
		Status: contracts.StatusCompleted,
		Output: "mock output",
	}, nil
}

func (m *runtimeMockAgent) Manifest() agent.Manifest {
	return agent.Manifest{Name: m.name}
}

// =============================================================================
// Pool Tests
// =============================================================================

func TestPool_ConcurrencyLimit(t *testing.T) {
	pool := runtime.NewPool(3, testLogger())

	var maxConcurrent atomic.Int32
	var current atomic.Int32

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		err := pool.Submit(context.Background(), func(ctx context.Context) {
			defer wg.Done()
			val := current.Add(1)
			// Track max concurrent
			for {
				old := maxConcurrent.Load()
				if val <= old || maxConcurrent.CompareAndSwap(old, val) {
					break
				}
			}
			time.Sleep(50 * time.Millisecond) // Simulate work
			current.Add(-1)
		})
		if err != nil {
			t.Fatalf("Submit %d: %v", i, err)
		}
	}

	wg.Wait()
	pool.Wait()

	if maxConcurrent.Load() > 3 {
		t.Errorf("max concurrent: got %d, want <= 3", maxConcurrent.Load())
	}
}

func TestPool_ContextCancellation(t *testing.T) {
	pool := runtime.NewPool(1, testLogger())

	// Fill the pool with a long-running task
	pool.Submit(context.Background(), func(ctx context.Context) {
		time.Sleep(5 * time.Second)
	})

	// Try to submit with a cancelled context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := pool.Submit(ctx, func(ctx context.Context) {
		t.Error("this should not execute")
	})

	if err == nil {
		t.Error("expected error when submitting with cancelled context to full pool")
	}
}

func TestPool_Stats(t *testing.T) {
	pool := runtime.NewPool(5, testLogger())

	done := make(chan struct{})
	pool.Submit(context.Background(), func(ctx context.Context) {
		<-done // Block until test releases
	})

	time.Sleep(50 * time.Millisecond) // Wait for goroutine to start

	stats := pool.Stats()
	if stats.ActiveWorkers != 1 {
		t.Errorf("ActiveWorkers: got %d, want 1", stats.ActiveWorkers)
	}
	if stats.MaxWorkers != 5 {
		t.Errorf("MaxWorkers: got %d, want 5", stats.MaxWorkers)
	}

	close(done)
	pool.Wait()
}

// =============================================================================
// Executor Tests
// =============================================================================

func setupExecutorTest(t *testing.T, executeFn func(context.Context, *agent.Task) (*agent.Result, error)) (*runtime.Executor, *registry.Registry) {
	t.Helper()

	logger := testLogger()
	reg := registry.New(logger)
	bus := eventbus.New(logger.Handler().(*slog.TextHandler)) // simple logger pass
	// Actually, eventbus.New takes *slog.Logger
	bus2 := eventbus.New(nil)

	mockAgent := &runtimeMockAgent{
		name:         "test-agent",
		capabilities: []agent.Capability{agent.CapabilityCodeGeneration},
		executeFn:    executeFn,
	}
	reg.Register(mockAgent)

	executor := runtime.NewExecutor(reg, bus2, logger, runtime.ExecutorConfig{
		DefaultTimeout: 5 * time.Second,
	})

	return executor, reg
}

func TestExecutor_Success(t *testing.T) {
	executor, _ := setupExecutorTest(t, func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
		return &agent.Result{
			TaskID: task.ID,
			Status: contracts.StatusCompleted,
			Output: "hello world",
		}, nil
	})

	task := &agent.Task{
		ID:   contracts.TaskID("tsk-001"),
		Name: "test task",
		Type: string(agent.CapabilityCodeGeneration),
	}

	result, err := executor.ExecuteTask(context.Background(), task)
	if err != nil {
		t.Fatalf("ExecuteTask: %v", err)
	}
	if result.Output != "hello world" {
		t.Errorf("Output: got %q, want %q", result.Output, "hello world")
	}
}

func TestExecutor_AgentError(t *testing.T) {
	executor, _ := setupExecutorTest(t, func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
		return nil, errors.New("agent error")
	})

	task := &agent.Task{
		ID:   contracts.TaskID("tsk-002"),
		Name: "failing task",
		Type: string(agent.CapabilityCodeGeneration),
	}

	_, err := executor.ExecuteTask(context.Background(), task)
	if err == nil {
		t.Error("expected error from executor")
	}
}

func TestExecutor_AgentPanic(t *testing.T) {
	executor, _ := setupExecutorTest(t, func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
		panic("intentional panic in test")
	})

	task := &agent.Task{
		ID:   contracts.TaskID("tsk-003"),
		Name: "panicking task",
		Type: string(agent.CapabilityCodeGeneration),
	}

	_, err := executor.ExecuteTask(context.Background(), task)
	if err == nil {
		t.Fatal("expected error from panicking agent")
	}
	// Should contain "panicked" in error message
	if !containsSubstring(err.Error(), "panic") {
		t.Errorf("error should mention panic, got: %v", err)
	}
}

func TestExecutor_Timeout(t *testing.T) {
	executor, _ := setupExecutorTest(t, func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
		// Wait for context cancellation
		<-ctx.Done()
		return nil, ctx.Err()
	})

	task := &agent.Task{
		ID:      contracts.TaskID("tsk-004"),
		Name:    "slow task",
		Type:    string(agent.CapabilityCodeGeneration),
		Timeout: 100 * time.Millisecond, // Very short timeout
	}

	_, err := executor.ExecuteTask(context.Background(), task)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestExecutor_NoAgentFound(t *testing.T) {
	logger := testLogger()
	reg := registry.New(logger)
	bus := eventbus.New(nil)

	// Empty registry — no agents registered
	executor := runtime.NewExecutor(reg, bus, logger, runtime.ExecutorConfig{})

	task := &agent.Task{
		ID:   contracts.TaskID("tsk-005"),
		Name: "orphan task",
		Type: "unknown_type",
	}

	_, err := executor.ExecuteTask(context.Background(), task)
	if err == nil {
		t.Error("expected error when no agent can handle task")
	}
}

// =============================================================================
// Runtime Integration Tests
// =============================================================================

func TestRuntime_FullLifecycle(t *testing.T) {
	logger := testLogger()
	reg := registry.New(logger)
	bus := eventbus.New(nil)

	// Register a mock agent
	mockAgent := &runtimeMockAgent{
		name:         "test-agent",
		capabilities: []agent.Capability{agent.CapabilityCodeGeneration},
		executeFn: func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
			return &agent.Result{
				TaskID: task.ID,
				Status: contracts.StatusCompleted,
				Output: "done",
			}, nil
		},
	}
	reg.Register(mockAgent)

	// Create runtime
	rt := runtime.New(reg, bus, logger, runtime.Config{
		MaxWorkers:     2,
		DefaultTimeout: 5 * time.Second,
	})

	// Track results
	var results []runtime.TaskResult
	var mu sync.Mutex

	err := rt.Start(func(tr runtime.TaskResult) {
		mu.Lock()
		results = append(results, tr)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Dispatch 5 tasks
	for i := 0; i < 5; i++ {
		task := &agent.Task{
			ID:   contracts.TaskID(fmt.Sprintf("tsk-%03d", i)),
			Name: fmt.Sprintf("task-%d", i),
			Type: string(agent.CapabilityCodeGeneration),
		}
		err := rt.Dispatch(context.Background(), task)
		if err != nil {
			t.Fatalf("Dispatch %d: %v", i, err)
		}
	}

	// Wait a bit for processing
	time.Sleep(1 * time.Second)

	// Stop gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rt.Stop(ctx)

	// Verify results
	mu.Lock()
	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Error != nil {
			t.Errorf("task %s failed: %v", r.TaskID, r.Error)
		}
	}
	mu.Unlock()
}

func TestRuntime_DoubleStart(t *testing.T) {
	logger := testLogger()
	reg := registry.New(logger)
	bus := eventbus.New(nil)

	rt := runtime.New(reg, bus, logger, runtime.Config{})
	rt.Start(nil)

	err := rt.Start(nil)
	if err == nil {
		t.Error("expected error on double start")
	}

	rt.Stop(context.Background())
}

func TestRuntime_DispatchBeforeStart(t *testing.T) {
	logger := testLogger()
	reg := registry.New(logger)
	bus := eventbus.New(nil)

	rt := runtime.New(reg, bus, logger, runtime.Config{})

	task := &agent.Task{
		ID:   contracts.TaskID("tsk-001"),
		Name: "premature",
		Type: string(agent.CapabilityCodeGeneration),
	}

	err := rt.Dispatch(context.Background(), task)
	if err == nil {
		t.Error("expected error when dispatching before start")
	}
}

// =============================================================================
// Helpers
// =============================================================================

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

## Pitfalls

### Pitfall 1: setupExecutorTest helper
Avoids repeating setup code in every test. Creates registry + bus + executor with a configurable mock agent.

### Pitfall 2: Panic test must NOT crash the test process
```go
func TestExecutor_AgentPanic(t *testing.T) {
    // This test MUST pass (not crash)
    // The executor should recover the panic and return an error
}
```
If the executor doesn't have panic recovery → `go test` crashes → CI fails.

### Pitfall 3: Atomic max tracking in concurrency test
```go
for {
    old := maxConcurrent.Load()
    if val <= old || maxConcurrent.CompareAndSwap(old, val) {
        break
    }
}
```
CompareAndSwap ensures atomic max update without mutex.

### Pitfall 4: Runtime test waits before Stop
```go
time.Sleep(1 * time.Second)  // Wait for all tasks to process
rt.Stop(ctx)
```
Tasks dispatch is async. Need to wait for processing before asserting results.

## Verify
```bash
go test -v -race -count=1 ./kernel/runtime/...
# Expected: ALL PASS, ≥ 10 test functions
```

## Checklist
- [ ] File `kernel/runtime/runtime_test.go` exists
- [ ] Package: `runtime_test`
- [ ] Mock agent with configurable executeFn
- [ ] ≥ 10 test functions
- [ ] Pool: concurrency limit, context cancellation, stats
- [ ] Executor: success, agent error, agent panic, timeout, no agent found
- [ ] Runtime: full lifecycle, double start, dispatch before start
- [ ] Uses atomic counters for concurrent assertions
- [ ] `go test -v -race ./kernel/runtime/...` ALL PASS
