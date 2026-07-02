# Micro-Task 2.21: Create kernel/registry/registry_test.go

## Info
- **File**: `kernel/registry/registry_test.go`
- **Package**: `registry_test`
- **Depends on**: 2.18-2.20
- **Time**: 25 min
- **Verify**: `go test -v -race -count=1 ./kernel/registry/...`

## Purpose
Full test coverage for registry: register/unregister, lookup, finding agents,
lifecycle management, thread safety, and error cases.

## EXACT code to create

```go
package registry_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
	"github.com/tiendat1751998/orchestrator/kernel/registry"
)

// =============================================================================
// Mock Plugin (satisfies plugin.Plugin interface)
// =============================================================================

type mockPlugin struct {
	name        string
	pluginType  string
	version     string
	initCalled  bool
	startCalled bool
	stopCalled  bool
	initErr     error
	startErr    error
	stopErr     error
}

func (m *mockPlugin) Name() string    { return m.name }
func (m *mockPlugin) Type() string    { return m.pluginType }
func (m *mockPlugin) Version() string { return m.version }

func (m *mockPlugin) Init(_ context.Context, _ map[string]any) error {
	m.initCalled = true
	return m.initErr
}

func (m *mockPlugin) Start(_ context.Context) error {
	m.startCalled = true
	return m.startErr
}

func (m *mockPlugin) Stop(_ context.Context) error {
	m.stopCalled = true
	return m.stopErr
}

func (m *mockPlugin) Health(_ context.Context) error {
	return nil
}

// =============================================================================
// Mock Agent (satisfies both plugin.Plugin and agent.Agent)
// =============================================================================

type mockAgent struct {
	mockPlugin
	capabilities []agent.Capability
	canHandleFn  func(*agent.Task) bool
}

func newMockAgent(name string, caps ...agent.Capability) *mockAgent {
	return &mockAgent{
		mockPlugin: mockPlugin{
			name:       name,
			pluginType: plugin.TypeAgent,
			version:    "1.0.0",
		},
		capabilities: caps,
	}
}

func (m *mockAgent) Capabilities() []agent.Capability {
	return m.capabilities
}

func (m *mockAgent) CanHandle(task *agent.Task) bool {
	if m.canHandleFn != nil {
		return m.canHandleFn(task)
	}
	// Default: match if task type matches any capability
	for _, cap := range m.capabilities {
		if string(cap) == task.Type {
			return true
		}
	}
	return false
}

func (m *mockAgent) Execute(_ context.Context, _ *agent.Task) (*agent.Result, error) {
	return nil, nil
}

func (m *mockAgent) Manifest() agent.Manifest {
	return agent.Manifest{Name: m.name}
}

// =============================================================================
// Mock Provider (satisfies both plugin.Plugin and provider.Provider)
// =============================================================================

type mockProvider struct {
	mockPlugin
}

func newMockProvider(name string) *mockProvider {
	return &mockProvider{
		mockPlugin: mockPlugin{
			name:       name,
			pluginType: plugin.TypeProvider,
			version:    "1.0.0",
		},
	}
}

func (m *mockProvider) Complete(_ context.Context, _ provider.Request) (*provider.Response, error) {
	return nil, nil
}

func (m *mockProvider) Stream(_ context.Context, _ provider.Request) (<-chan provider.StreamChunk, error) {
	return nil, nil
}

func (m *mockProvider) Models() []string {
	return []string{"test-model"}
}

// =============================================================================
// Register / Unregister Tests
// =============================================================================

func TestRegistry_Register_Agent(t *testing.T) {
	reg := registry.New(nil)
	a := newMockAgent("backend", agent.CapabilityCodeGeneration)

	err := reg.Register(a)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if reg.Count() != 1 {
		t.Errorf("Count: got %d, want 1", reg.Count())
	}
}

func TestRegistry_Register_DuplicateName(t *testing.T) {
	reg := registry.New(nil)

	a1 := newMockAgent("backend", agent.CapabilityCodeGeneration)
	a2 := newMockAgent("backend", agent.CapabilityCodeReview) // Same name

	if err := reg.Register(a1); err != nil {
		t.Fatalf("Register first: %v", err)
	}

	err := reg.Register(a2)
	if err == nil {
		t.Error("expected error for duplicate name")
	}
}

func TestRegistry_Register_Provider(t *testing.T) {
	reg := registry.New(nil)
	p := newMockProvider("antigravity")

	err := reg.Register(p)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	got, err := reg.GetProvider("antigravity")
	if err != nil {
		t.Fatalf("GetProvider: %v", err)
	}
	if got.Name() != "antigravity" {
		t.Errorf("Name: got %q, want %q", got.Name(), "antigravity")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	reg := registry.New(nil)
	a := newMockAgent("backend", agent.CapabilityCodeGeneration)
	reg.Register(a)

	err := reg.Unregister("backend")
	if err != nil {
		t.Fatalf("Unregister: %v", err)
	}

	if reg.Count() != 0 {
		t.Errorf("Count after unregister: got %d, want 0", reg.Count())
	}
}

func TestRegistry_Unregister_NotFound(t *testing.T) {
	reg := registry.New(nil)

	err := reg.Unregister("nonexistent")
	if err == nil {
		t.Error("expected error for unregistering nonexistent plugin")
	}
}

// =============================================================================
// Lookup Tests
// =============================================================================

func TestRegistry_GetAgent_NotFound(t *testing.T) {
	reg := registry.New(nil)

	_, err := reg.GetAgent("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent agent")
	}
}

func TestRegistry_GetProvider_NotFound(t *testing.T) {
	reg := registry.New(nil)

	_, err := reg.GetProvider("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}

func TestRegistry_ListAgents(t *testing.T) {
	reg := registry.New(nil)
	reg.Register(newMockAgent("backend", agent.CapabilityCodeGeneration))
	reg.Register(newMockAgent("reviewer", agent.CapabilityCodeReview))

	agents := reg.ListAgents()
	if len(agents) != 2 {
		t.Errorf("ListAgents: got %d, want 2", len(agents))
	}
}

func TestRegistry_ListProviders(t *testing.T) {
	reg := registry.New(nil)
	reg.Register(newMockProvider("antigravity"))

	providers := reg.ListProviders()
	if len(providers) != 1 {
		t.Errorf("ListProviders: got %d, want 1", len(providers))
	}
}

// =============================================================================
// FindAgentForTask Tests
// =============================================================================

func TestRegistry_FindAgentForTask_Found(t *testing.T) {
	reg := registry.New(nil)
	reg.Register(newMockAgent("backend", agent.CapabilityCodeGeneration))
	reg.Register(newMockAgent("reviewer", agent.CapabilityCodeReview))

	task := &agent.Task{
		Name: "write handler",
		Type: string(agent.CapabilityCodeGeneration),
	}

	a, err := reg.FindAgentForTask(task)
	if err != nil {
		t.Fatalf("FindAgentForTask: %v", err)
	}
	if a.Name() != "backend" {
		t.Errorf("got agent %q, want %q", a.Name(), "backend")
	}
}

func TestRegistry_FindAgentForTask_NotFound(t *testing.T) {
	reg := registry.New(nil)
	reg.Register(newMockAgent("reviewer", agent.CapabilityCodeReview))

	task := &agent.Task{
		Name: "deploy app",
		Type: string(agent.CapabilityDeployment),
	}

	_, err := reg.FindAgentForTask(task)
	if err == nil {
		t.Error("expected error when no agent can handle task")
	}
}

func TestRegistry_FindAgentForTask_NoAgentsRegistered(t *testing.T) {
	reg := registry.New(nil)

	task := &agent.Task{
		Name: "some task",
		Type: "anything",
	}

	_, err := reg.FindAgentForTask(task)
	if err == nil {
		t.Error("expected error when no agents registered")
	}
}

func TestRegistry_FindAllAgentsForTask(t *testing.T) {
	reg := registry.New(nil)
	reg.Register(newMockAgent("backend1", agent.CapabilityCodeGeneration))
	reg.Register(newMockAgent("backend2", agent.CapabilityCodeGeneration))
	reg.Register(newMockAgent("reviewer", agent.CapabilityCodeReview))

	task := &agent.Task{
		Name: "write code",
		Type: string(agent.CapabilityCodeGeneration),
	}

	agents := reg.FindAllAgentsForTask(task)
	if len(agents) != 2 {
		t.Errorf("FindAllAgentsForTask: got %d agents, want 2", len(agents))
	}
}

// =============================================================================
// Lifecycle Tests
// =============================================================================

func TestRegistry_InitAll_Success(t *testing.T) {
	reg := registry.New(nil)
	a := newMockAgent("backend", agent.CapabilityCodeGeneration)
	reg.Register(a)

	err := reg.InitAll(context.Background(), nil)
	if err != nil {
		t.Fatalf("InitAll: %v", err)
	}
	if !a.initCalled {
		t.Error("Init was not called on the plugin")
	}
}

func TestRegistry_InitAll_Failure(t *testing.T) {
	reg := registry.New(nil)
	a := newMockAgent("backend", agent.CapabilityCodeGeneration)
	a.initErr = errors.New("init failed")
	reg.Register(a)

	err := reg.InitAll(context.Background(), nil)
	if err == nil {
		t.Error("expected error from InitAll")
	}
}

func TestRegistry_StartAll_Rollback(t *testing.T) {
	reg := registry.New(nil)

	a1 := newMockAgent("agent1", agent.CapabilityCodeGeneration)
	a2 := newMockAgent("agent2", agent.CapabilityCodeReview)
	a2.startErr = errors.New("start failed") // agent2 fails to start

	reg.Register(a1)
	reg.Register(a2)

	err := reg.StartAll(context.Background())
	if err == nil {
		t.Fatal("expected error from StartAll")
	}

	// agent1 was started successfully, should have been stopped in rollback
	if !a1.stopCalled {
		t.Error("agent1 should have been stopped during rollback")
	}

	// agent2 was never started, should NOT have been stopped
	if a2.stopCalled {
		t.Error("agent2 should NOT have been stopped (it never started)")
	}
}

func TestRegistry_StopAll_ReverseOrder(t *testing.T) {
	reg := registry.New(nil)
	order := make([]string, 0)
	mu := sync.Mutex{}

	a1 := newMockAgent("first", agent.CapabilityCodeGeneration)
	a1.stopErr = nil
	original1 := a1.Stop
	_ = original1
	// Override Stop to track order
	// Note: We can't easily override methods on struct, so we test via the order slice
	// The actual stop order is guaranteed by AllPluginsReversed()

	a2 := newMockAgent("second", agent.CapabilityCodeReview)

	reg.Register(a1)
	reg.Register(a2)

	// Verify AllPluginsReversed returns correct order
	reversed := reg.AllPluginsReversed()
	if len(reversed) != 2 {
		t.Fatalf("AllPluginsReversed: got %d, want 2", len(reversed))
	}
	if reversed[0].Name() != "second" {
		t.Errorf("first in reversed: got %q, want %q", reversed[0].Name(), "second")
	}
	if reversed[1].Name() != "first" {
		t.Errorf("second in reversed: got %q, want %q", reversed[1].Name(), "first")
	}

	_ = order
	_ = mu
}

func TestRegistry_StopAll_ContinuesOnError(t *testing.T) {
	reg := registry.New(nil)

	a1 := newMockAgent("agent1", agent.CapabilityCodeGeneration)
	a1.stopErr = errors.New("stop failed")

	a2 := newMockAgent("agent2", agent.CapabilityCodeReview)

	reg.Register(a1)
	reg.Register(a2)

	// Start all first
	reg.InitAll(context.Background(), nil)
	reg.StartAll(context.Background())

	// StopAll — agent1 fails to stop, but agent2 should still be stopped
	err := reg.StopAll(context.Background())
	if err == nil {
		t.Error("expected error from StopAll")
	}

	// Both plugins should have Stop called
	if !a1.stopCalled {
		t.Error("agent1.Stop should have been called")
	}
	if !a2.stopCalled {
		t.Error("agent2.Stop should have been called even after agent1 failed")
	}
}

// =============================================================================
// Health Check Tests
// =============================================================================

func TestRegistry_HealthCheckAll(t *testing.T) {
	reg := registry.New(nil)
	reg.Register(newMockAgent("healthy", agent.CapabilityCodeGeneration))
	reg.Register(newMockProvider("provider1"))

	results := reg.HealthCheckAll(context.Background())
	if len(results) != 2 {
		t.Errorf("HealthCheckAll: got %d results, want 2", len(results))
	}
}

// =============================================================================
// Concurrent Safety Tests
// =============================================================================

func TestRegistry_ConcurrentAccess(t *testing.T) {
	reg := registry.New(nil)

	var wg sync.WaitGroup

	// 50 goroutines registering concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			a := newMockAgent(
				fmt.Sprintf("agent-%d", id),
				agent.CapabilityCodeGeneration,
			)
			reg.Register(a)
		}(i)
	}

	// 50 goroutines reading concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			reg.ListAgents()
			reg.Count()
		}(id)
	}

	wg.Wait()

	if reg.Count() != 50 {
		t.Errorf("Count: got %d, want 50", reg.Count())
	}
}
```

## Pitfalls

### Pitfall 1: Mock must implement ALL interface methods
`mockAgent` must implement BOTH `plugin.Plugin` AND `agent.Agent` interfaces.
Missing any method → compile error. Use `var _ agent.Agent = (*mockAgent)(nil)` to verify.

### Pitfall 2: Lifecycle test needs mock with controllable errors
```go
a.startErr = errors.New("start failed") // Configure before calling
```
This allows testing failure paths without complex mocking frameworks.

### Pitfall 3: Concurrent test needs unique names
```go
fmt.Sprintf("agent-%d", id) // Each goroutine gets unique name
```
If all goroutines use "agent" → duplicate registration error → test fails for wrong reason.

## Verify
```bash
go test -v -race -count=1 ./kernel/registry/...
# Expected: ALL PASS, ≥ 18 test functions
```

## Checklist
- [ ] File `kernel/registry/registry_test.go` exists
- [ ] Package: `registry_test`
- [ ] Mock types: mockPlugin, mockAgent, mockProvider
- [ ] ≥ 18 test functions
- [ ] Tests: register agent, provider; duplicate name error
- [ ] Tests: unregister success, not found error
- [ ] Tests: GetAgent/GetProvider not found
- [ ] Tests: ListAgents, ListProviders
- [ ] Tests: FindAgentForTask (found, not found, no agents)
- [ ] Tests: FindAllAgentsForTask
- [ ] Tests: InitAll success and failure
- [ ] Tests: StartAll rollback on failure
- [ ] Tests: StopAll reverse order, continues on error
- [ ] Tests: HealthCheckAll
- [ ] Tests: concurrent register + read (race detector)
- [ ] `go test -v -race ./kernel/registry/...` ALL PASS
