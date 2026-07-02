# Micro-Task 3.19: Create sdk/testing/mocks_test.go

## Info
- **File**: `sdk/testing/mocks_test.go`
- **Package**: `testing_test`
- **Depends on**: 3.18 (mocks.go)
- **Time**: 15 min
- **Verify**: `go test -v -race -count=1 ./sdk/testing/...`

## Purpose
Triển khai bộ kiểm thử tự động (Unit Tests) cho chính bộ giả lập `sdk/testing`. Bước này đảm bảo các cấu trúc mock (`MockProvider`, `MockAgent`, `MockTool`, `MockEventBus`) hoạt động ổn định, ghi nhận dữ liệu chính xác và phản hồi đúng đắn theo hành vi tiêm vào (behavior injection) trước khi bàn giao cho các tests của hệ thống.

## EXACT code to create

```go
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
	contractstool "github.com/tiendat1751998/orchestrator/contracts/tool"
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

	// Verify parallel publishes does not cause panic / race condition
	workersCount := 20
	publishesPerWorker := 50
	var wg sync.WaitGroup

	for i := 0; i < workersCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < publishesPerWorker; j++ {
				bus.Publish(context.Background(), contractsevent.Event{
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

	// Verify clear resets state
	bus.Clear()
	if len(bus.GetPublished()) != 0 {
		t.Error("expected published list to be empty after Clear()")
	}
}
```

## Verify
```bash
go test -v -race -count=1 ./sdk/testing/...
```

## Checklist
- [ ] File `sdk/testing/mocks_test.go` tồn tại
- [ ] Package name: `testing_test`
- [ ] Test `TestMockProvider_Send` xác thực hành vi trả về mặc định và hành vi tiêm (SendFn) thành công
- [ ] Test `TestMockAgent_Execute` xác thực `CanHandle` và `Execute` thành công
- [ ] Test `TestMockTool_Execute` kiểm thử thực thi tool mock thành công
- [ ] Test `TestMockEventBus_Concurrency` kiểm tra ghi đồng thời (20 goroutines x 50 publishes) không phát sinh lỗi tranh chấp tài nguyên (data races)
- [ ] Phương thức `Clear` dọn sạch dữ liệu mock thành công
- [ ] `go test -v -race ./sdk/testing/...` trả về ALL PASS
