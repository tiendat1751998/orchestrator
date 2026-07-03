package testing_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	contractsagent "github.com/tiendat1751998/orchestrator/contracts/agent"
	contractsevent "github.com/tiendat1751998/orchestrator/contracts/event"
	contractsprovider "github.com/tiendat1751998/orchestrator/contracts/provider"
	sdktesting "github.com/tiendat1751998/orchestrator/sdk/testing"
)

// =============================================================================
// Provider Mock Verification
// =============================================================================

func TestMockProvider_Send(t *testing.T) {
	// Case 1: Default response
	mp := &sdktesting.MockProvider{NameVal: "default-mock"}
	resp, err := mp.Send(context.Background(), &contractsprovider.Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Default mock provider output content" {
		t.Errorf("got content: %q", resp.Content)
	}

	// Case 2: Injected failure behavior
	mpFailed := &sdktesting.MockProvider{
		SendFn: func(ctx context.Context, req *contractsprovider.Request) (*contractsprovider.Response, error) {
			return nil, errors.New("injected network error")
		},
	}
	_, err = mpFailed.Send(context.Background(), &contractsprovider.Request{})
	if err == nil || err.Error() != "injected network error" {
		t.Errorf("expected error 'injected network error', got %v", err)
	}
}

// =============================================================================
// Agent Mock Verification
// =============================================================================

func TestMockAgent_Execute(t *testing.T) {
	ma := &sdktesting.MockAgent{
		NameVal: "mock-coder",
		CapabilitiesVal: []contractsagent.Capability{
			contractsagent.CapabilityCodeGeneration,
		},
	}

	task := &contractsagent.Task{ID: "tsk-x", Type: "code_generation"}
	if !ma.CanHandle(task) {
		t.Error("expected mock agent to handle 'code_generation' capability")
	}

	res, err := ma.Execute(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if res.Output != "Default mock agent execution success output" {
		t.Errorf("got output: %q", res.Output)
	}
}

// =============================================================================
// Tool Mock Verification
// =============================================================================

func TestMockTool_Execute(t *testing.T) {
	mt := &sdktesting.MockTool{
		NameVal:        "write_file",
		DescriptionVal: "Write text content to file",
	}

	res, err := mt.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Output != "Default mock tool execution output success" || res.ExitCode != 0 {
		t.Errorf("got result: %v", res)
	}
}

// =============================================================================
// EventBus Mock Concurrency Verification
// =============================================================================

func TestMockEventBus_Concurrency(t *testing.T) {
	bus := &sdktesting.MockEventBus{}

	workersCount := 20
	publishesPerWorker := 50
	var wg sync.WaitGroup

	for i := 0; i < workersCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < publishesPerWorker; j++ {
				_ = bus.Publish(context.Background(), contractsevent.Event{
					ID:     "evt-123",
					Type:   "test.event",
					Source: "worker",
				})
			}
		}(i)
	}

	wg.Wait()

	events := bus.GetPublished()
	expectedCount := workersCount * publishesPerWorker
	if len(events) != expectedCount {
		t.Errorf("published events count: got %d, want %d", len(events), expectedCount)
	}

	bus.Clear()
	if len(bus.GetPublished()) != 0 {
		t.Error("expected published list to be empty after Clear()")
	}
}
