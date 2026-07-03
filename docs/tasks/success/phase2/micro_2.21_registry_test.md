# Micro-Task 2.21: Create kernel/registry/registry_test.go

## Info
- **File**: `kernel/registry/registry_test.go`
- **Package**: `registry_test`
- **Depends on**: 2.18-2.20
- **Time**: 25 min
- **Verify**: `go test -v -race -count=1 ./kernel/registry/...`

## Purpose
Implements unit tests verifying the correctness of plugin registration, service lookups, agent capability routing, lifecycle state transitions, rollback procedures, and concurrent safety properties.

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

// Health satisfies the updated Task 1.40 Plugin signature
func (m *mockPlugin) Health(_ context.Context) (plugin.HealthReport, error) {
	return plugin.HealthReport{Status: plugin.HealthOK}, nil
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
	a2 := newMockAgent("backend", agent.CapabilityCodeReview)

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
	a2.startErr = errors.New("start failed")

	reg.Register(a1)
	reg.Register(a2)

	err := reg.StartAll(context.Background())
	if err == nil {
		t.Fatal("expected error from StartAll")
	}

	if !a1.stopCalled {
		t.Error("agent1 should have been stopped during rollback")
	}

	if a2.stopCalled {
		t.Error("agent2 should NOT have been stopped (it never started)")
	}
}

func TestRegistry_StopAll_ReverseOrder(t *testing.T) {
	reg := registry.New(nil)

	a1 := newMockAgent("first", agent.CapabilityCodeGeneration)
	a2 := newMockAgent("second", agent.CapabilityCodeReview)

	reg.Register(a1)
	reg.Register(a2)

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
}

func TestRegistry_StopAll_ContinuesOnError(t *testing.T) {
	reg := registry.New(nil)

	a1 := newMockAgent("agent1", agent.CapabilityCodeGeneration)
	a1.stopErr = errors.New("stop failed")

	a2 := newMockAgent("agent2", agent.CapabilityCodeReview)

	reg.Register(a1)
	reg.Register(a2)

	reg.InitAll(context.Background(), nil)
	reg.StartAll(context.Background())

	err := reg.StopAll(context.Background())
	if err == nil {
		t.Error("expected error from StopAll")
	}

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

## Rules
1. **Mock Signatures Alignment**: Checkers must verify that all mock structures implement the updated `Plugin` interface methods. Specifically, `Health` must declare the `(plugin.HealthReport, error)` signature to compile safely.
2. **Concurrent Safety Assertions**: Parallel tests must verify map structures using distinct component keys (e.g. `fmt.Sprintf("agent-%d", id)`) to prevent duplicate key registration errors.
3. **External Test boundary checks**: Unit tests must utilize the `registry_test` package structure to verify the registry contract operates correctly as an imported dependency.

## ⚠️ Pitfalls

### Pitfall 1: Mismatched mock method signatures breaking compile pipelines
Using a legacy `Health(context.Context) error` method signature on test mocks prevents compilation since `plugin.Plugin` has been upgraded. Always declare `Health(context.Context) (plugin.HealthReport, error)`.

### Pitfall 2: Using static names inside concurrent registration tests
If all concurrent registering goroutines attempt to write the name `"backend"`, registration returns duplicate errors. Enforce unique name indexes inside mock registrations.

## Verify
```bash
go test -v -race ./kernel/registry/...
```

## Checklist
- [ ] File `kernel/registry/registry_test.go` exists
- [ ] Package: `registry_test`
- [ ] Mock structures implement `plugin.Plugin`, `agent.Agent`, and `provider.Provider` contracts
- [ ] `Health` mocks align with the updated signatures
- [ ] Tests verify registry additions, deletions, duplicate name guards, and type assertions
- [ ] Lifecycle tests verify initialization, rollback sequences, and teardown order
- [ ] Concurrent tests verify registry safety under `-race` checks
- [ ] `go test -v -race ./kernel/registry/...` passes
