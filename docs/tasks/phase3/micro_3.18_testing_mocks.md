# Micro-Task 3.18: Create sdk/testing/mocks.go

## Info
- **File**: `sdk/testing/mocks.go`
- **Package**: `testing`
- **Depends on**: 1.12 (provider contract), 1.15 (tool contract), 1.21 (agent contract), 1.23 (event contract)
- **Time**: 20 min
- **Verify**: `go build ./sdk/testing/...`

## Purpose
Triển khai bộ giả lập thử nghiệm (`TestingMocks`) chứa các cấu trúc mock dùng chung (`MockAgent`, `MockProvider`, `MockTool`, `MockEventBus`). Đây là công cụ quan quan trọng phục vụ cho việc kiểm thử tích hợp (integration tests) và cô lập logic (unit tests) ở các Phase tiếp theo (Plugins, Modules, và API).

## EXACT code to create

```go
// Package testing provides standard mock structures for core orchestrator components.
package testing

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	contractsagent "github.com/tiendat1751998/orchestrator/contracts/agent"
	contractsevent "github.com/tiendat1751998/orchestrator/contracts/event"
	contractsplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	contractsprovider "github.com/tiendat1751998/orchestrator/contracts/provider"
	contractstool "github.com/tiendat1751998/orchestrator/contracts/tool"
)

// =============================================================================
// Mock Provider
// =============================================================================

// MockProvider simulates an AI provider.
type MockProvider struct {
	NameVal        string
	ModelsVal      []string
	IsAvailableVal bool
	
	SendFn   func(ctx context.Context, req *contractsprovider.Request) (*contractsprovider.Response, error)
	StreamFn func(ctx context.Context, req *contractsprovider.Request) (<-chan contractsprovider.StreamChunk, error)
}

func (m *MockProvider) Name() string                     { return m.NameVal }
func (m *MockProvider) Type() contractsplugin.Type       { return contractsplugin.TypeProvider }
func (m *MockProvider) Version() string                  { return "1.0.0" }
func (m *MockProvider) Init(_ context.Context, _ map[string]any) error { return nil }
func (m *MockProvider) Start(_ context.Context) error    { return nil }
func (m *MockProvider) Stop(_ context.Context) error     { return nil }
func (m *MockProvider) Health(_ context.Context) (contractsplugin.HealthReport, error) {
	return contractsplugin.HealthReport{Status: contractsplugin.HealthOK, Timestamp: time.Now()}, nil
}

func (m *MockProvider) IsAvailable(ctx context.Context) bool {
	return m.IsAvailableVal
}

func (m *MockProvider) Models(ctx context.Context) ([]string, error) {
	return m.ModelsVal, nil
}

func (m *MockProvider) Send(ctx context.Context, req *contractsprovider.Request) (*contractsprovider.Response, error) {
	if m.SendFn != nil {
		return m.SendFn(ctx, req)
	}
	return &contractsprovider.Response{
		ID:        "mock-response-id",
		Content:   "Default mock provider output content",
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockProvider) Stream(ctx context.Context, req *contractsprovider.Request) (<-chan contractsprovider.StreamChunk, error) {
	if m.StreamFn != nil {
		return m.StreamFn(ctx, req)
	}
	ch := make(chan contractsprovider.StreamChunk, 2)
	ch <- contractsprovider.StreamChunk{Delta: "Default stream data"}
	ch <- contractsprovider.StreamChunk{Done: true}
	close(ch)
	return ch, nil
}

// =============================================================================
// Mock Agent
// =============================================================================

// MockAgent simulates an AI agent.
type MockAgent struct {
	NameVal         string
	RoleVal         string
	CapabilitiesVal []contractsagent.Capability
	ManifestVal     contractsagent.Manifest
	
	ExecuteFn   func(ctx context.Context, task *contractsagent.Task) (*contractsagent.Result, error)
	CanHandleFn func(task *contractsagent.Task) bool
}

func (m *MockAgent) Name() string                     { return m.NameVal }
func (m *MockAgent) Role() string                     { return m.RoleVal }
func (m *MockAgent) Type() contractsplugin.Type       { return contractsplugin.TypeAgent }
func (m *MockAgent) Version() string                  { return "1.0.0" }
func (m *MockAgent) Init(_ context.Context, _ map[string]any) error { return nil }
func (m *MockAgent) Start(_ context.Context) error    { return nil }
func (m *MockAgent) Stop(_ context.Context) error     { return nil }
func (m *MockAgent) Health(_ context.Context) (contractsplugin.HealthReport, error) {
	return contractsplugin.HealthReport{Status: contractsplugin.HealthOK, Timestamp: time.Now()}, nil
}

func (m *MockAgent) Capabilities() []contractsagent.Capability {
	return m.CapabilitiesVal
}

func (m *MockAgent) Manifest() contractsagent.Manifest {
	return m.ManifestVal
}

func (m *MockAgent) CanHandle(task *contractsagent.Task) bool {
	if m.CanHandleFn != nil {
		return m.CanHandleFn(task)
	}
	for _, cap := range m.CapabilitiesVal {
		if string(cap) == task.Type {
			return true
		}
	}
	return false
}

func (m *MockAgent) Execute(ctx context.Context, task *contractsagent.Task) (*contractsagent.Result, error) {
	if m.ExecuteFn != nil {
		return m.ExecuteFn(ctx, task)
	}
	return &contractsagent.Result{
		TaskID:   task.ID,
		Status:   "success",
		Output:   "Default mock agent execution success output",
		Duration: 10 * time.Millisecond,
	}, nil
}

// =============================================================================
// Mock Tool
// =============================================================================

// MockTool simulates a tool execution logic.
type MockTool struct {
	NameVal        string
	DescriptionVal string
	SchemaVal      *contractstool.Schema
	
	ExecuteFn func(ctx context.Context, args json.RawMessage) (*contractstool.Result, error)
}

func (m *MockTool) Name() string                     { return m.NameVal }
func (m *MockTool) Description() string              { return m.DescriptionVal }
func (m *MockTool) Type() contractsplugin.Type       { return contractsplugin.TypeTool }
func (m *MockTool) Version() string                  { return "1.0.0" }
func (m *MockTool) Init(_ context.Context, _ map[string]any) error { return nil }
func (m *MockTool) Start(_ context.Context) error    { return nil }
func (m *MockTool) Stop(_ context.Context) error     { return nil }
func (m *MockTool) Health(_ context.Context) (contractsplugin.HealthReport, error) {
	return contractsplugin.HealthReport{Status: contractsplugin.HealthOK, Timestamp: time.Now()}, nil
}

func (m *MockTool) Schema() *contractstool.Schema {
	return m.SchemaVal
}

func (m *MockTool) Execute(ctx context.Context, args json.RawMessage) (*contractstool.Result, error) {
	if m.ExecuteFn != nil {
		return m.ExecuteFn(ctx, args)
	}
	return &contractstool.Result{
		Output:   "Default mock tool execution output success",
		ExitCode: 0,
	}, nil
}

// =============================================================================
// Mock Event Bus
// =============================================================================

// MockEventBus implements contractsevent.Bus. Thread-safe.
type MockEventBus struct {
	mu        sync.RWMutex
	Published []contractsevent.Event
}

// Publish records the event into an in-memory history slice.
func (m *MockEventBus) Publish(ctx context.Context, evt contractsevent.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Published = append(m.Published, evt)
	return nil
}

// Subscribe returns a mock unsubscribe no-op.
func (m *MockEventBus) Subscribe(pattern string, handler func(contractsevent.Event)) (func(), error) {
	// Simple no-op unsubscribe return
	return func() {}, nil
}

// GetPublished returns a copied slice of all captured events.
func (m *MockEventBus) GetPublished() []contractsevent.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copied := make([]contractsevent.Event, len(m.Published))
	copy(copied, m.Published)
	return copied
}

// Clear resets the recorded history.
func (m *MockEventBus) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Published = nil
}
```

## ⚠️ Pitfalls

### Pitfall 1: Non-thread-safe records on MockEventBus
```go
// ❌ WRONG:
func (m *MockEventBus) Publish(ctx, evt) error {
    m.Published = append(m.Published, evt) // Concurrent writes during parallel task runs causes data corruption crashes!
    return nil
}
```
Mocks are often used in stress tests or concurrent task execution suites. Shared slices inside test mocks must be protected by a `sync.Mutex` or `sync.RWMutex` to avoid map/slice corruption panics.

### Pitfall 2: Returning direct slice reference in helper
The helper `GetPublished()` must return a copied slice instead of a direct reference to prevent unit tests from mutating the internal history of the mock.

## Verify
```bash
go build ./sdk/testing/...
```

## Checklist
- [ ] File `sdk/testing/mocks.go` exists
- [ ] Package: `testing`
- [ ] `MockProvider` implements all `provider.Provider` and `plugin.Plugin` methods
- [ ] `MockAgent` implements all `agent.Agent` and `plugin.Plugin` methods
- [ ] `MockTool` implements all `tool.Tool` and `plugin.Plugin` methods
- [ ] `MockEventBus` uses a `sync.RWMutex` to guard the slice of published events
- [ ] `MockEventBus.GetPublished` returns a clean cloned slice copy
- [ ] Mocks allow injecting behaviors using function pointers (`SendFn`, `ExecuteFn`)
- [ ] `go build ./sdk/testing/...` passes
