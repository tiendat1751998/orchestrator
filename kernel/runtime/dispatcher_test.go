package runtime

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/kernel/registry"
)

func TestDispatcher_Success(t *testing.T) {
	reg := registry.New(nil)
	a := &mockAgent{
		mockPlugin: mockPlugin{name: "test-agent"},
		caps:       []agent.Capability{"test-task"},
		executeFn: func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
			return agent.SuccessResult(task.ID, "test-agent", "hello world"), nil
		},
	}
	if err := reg.Register(a); err != nil {
		t.Fatalf("failed to register agent: %v", err)
	}

	bus := &mockEventBus{}
	executor := NewExecutor(reg, bus, nil, ExecutorConfig{})
	pool := NewPool(2, nil)
	dispatcher := NewDispatcher(executor, pool, slog.Default(), DispatcherConfig{ResultBufferSize: 10})

	task := &agent.Task{
		ID:   contracts.TaskID("task-1"),
		Name: "test-task-1",
		Type: "test-task",
	}

	ctx := context.Background()
	err := dispatcher.Dispatch(ctx, task)
	if err != nil {
		t.Fatalf("failed to dispatch task: %v", err)
	}

	// Read result
	select {
	case res := <-dispatcher.Results():
		if res.TaskID != task.ID {
			t.Errorf("expected task ID %s, got %s", task.ID, res.TaskID)
		}
		if res.Error != nil {
			t.Errorf("unexpected execution error: %v", res.Error)
		}
		if res.Result.Output != "hello world" {
			t.Errorf("expected output 'hello world', got %q", res.Result.Output)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for task execution result")
	}

	pool.Wait()
}

func TestDispatcher_StopAndReject(t *testing.T) {
	reg := registry.New(nil)
	a := &mockAgent{
		mockPlugin: mockPlugin{name: "test-agent"},
		caps:       []agent.Capability{"test-task"},
	}
	_ = reg.Register(a)

	bus := &mockEventBus{}
	executor := NewExecutor(reg, bus, nil, ExecutorConfig{})
	pool := NewPool(2, nil)
	dispatcher := NewDispatcher(executor, pool, nil, DispatcherConfig{})

	dispatcher.Stop()

	task := &agent.Task{
		ID:   contracts.TaskID("task-2"),
		Name: "test-task-2",
		Type: "test-task",
	}

	err := dispatcher.Dispatch(context.Background(), task)
	if err == nil {
		t.Fatal("expected error dispatching to stopped dispatcher, got nil")
	}
}

func TestDispatcher_ResultContextCancel(t *testing.T) {
	reg := registry.New(nil)
	// We want to simulate the dispatcher trying to write to the results channel
	// but the context gets cancelled. We can set up a dispatcher with buffer size 1.
	// We will fill the channel with 1 result, then dispatch a second task, and cancel its context.
	// Since the results channel is full, the second task's goroutine will block on writing to d.results.
	// When we cancel the context, the send select block will exit via case <-ctx.Done(), avoiding a hang.

	a := &mockAgent{
		mockPlugin: mockPlugin{name: "test-agent"},
		caps:       []agent.Capability{"test-task"},
		executeFn: func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
			return agent.SuccessResult(task.ID, "test-agent", "success"), nil
		},
	}
	_ = reg.Register(a)

	bus := &mockEventBus{}
	executor := NewExecutor(reg, bus, nil, ExecutorConfig{})
	pool := NewPool(2, nil)
	// buffer size 1
	dispatcher := NewDispatcher(executor, pool, slog.Default(), DispatcherConfig{ResultBufferSize: 1})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	task1 := &agent.Task{
		ID:   contracts.TaskID("task-1"),
		Name: "task-1",
		Type: "test-task",
	}
	task2 := &agent.Task{
		ID:   contracts.TaskID("task-2"),
		Name: "task-2",
		Type: "test-task",
	}

	// Dispatch task1. It will execute and write its result to the channel of size 1.
	err := dispatcher.Dispatch(ctx, task1)
	if err != nil {
		t.Fatalf("failed to dispatch task1: %v", err)
	}

	// Give a bit of time for task1 to finish and fill the results channel.
	time.Sleep(50 * time.Millisecond)

	// Now results channel is full.
	// Dispatch task2. It will finish execution and block when trying to send to dispatcher.Results().
	err = dispatcher.Dispatch(ctx, task2)
	if err != nil {
		t.Fatalf("failed to dispatch task2: %v", err)
	}

	// Let's cancel the context. This should unblock the task2 worker select write.
	cancel()

	// Wait for the pool to ensure the worker finished (it shouldn't hang).
	done := make(chan struct{})
	go func() {
		pool.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success: worker didn't hang
	case <-time.After(2 * time.Second):
		t.Fatal("pool.Wait hung: worker likely blocked on full results channel instead of aborting via cancelled context")
	}
}
