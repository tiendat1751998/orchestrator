# Micro-Task 3.05: Create sdk/agent/agent_test.go

## Info
- **File**: `sdk/agent/agent_test.go`
- **Package**: `agent_test`
- **Depends on**: 3.04 (agent.go), 1.40 (health/plugin contracts), 1.12 (provider contract)
- **Time**: 20 min
- **Verify**: `go test -v -race -count=1 ./sdk/agent/...`

## Purpose
Implements integration unit tests for the Agent SDK. It verifies manifest loading, relative path prompt resolutions, prompt parsing, token Warnings checks, and ReAct loop controls (including iteration protections, cancellations, and mock tool executions).

## EXACT code to create

```go
package agent_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
	cplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
	"github.com/tiendat1751998/orchestrator/contracts/tool"
	sdkagent "github.com/tiendat1751998/orchestrator/sdk/agent"
)

// =============================================================================
// Lightweight Local Mocks for Testing
// =============================================================================

type mockProvider struct {
	sendFn   func(ctx context.Context, req *provider.Request) (*provider.Response, error)
	streamFn func(ctx context.Context, req *provider.Request) (<-chan provider.StreamChunk, error)
}

func (m *mockProvider) Name() string        { return "mock-provider" }
func (m *mockProvider) Type() cplugin.Type  { return cplugin.TypeProvider }
func (m *mockProvider) Version() string     { return "1.0.0" }

func (m *mockProvider) Init(_ context.Context, _ map[string]any) error { return nil }
func (m *mockProvider) Start(_ context.Context) error                  { return nil }
func (m *mockProvider) Stop(_ context.Context) error                   { return nil }

func (m *mockProvider) Health(_ context.Context) (cplugin.HealthReport, error) {
	return cplugin.HealthReport{Status: cplugin.HealthOK}, nil
}

func (m *mockProvider) Send(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	if m.sendFn != nil {
		return m.sendFn(ctx, req)
	}
	return &provider.Response{Content: "mock output"}, nil
}

func (m *mockProvider) Stream(ctx context.Context, req *provider.Request) (<-chan provider.StreamChunk, error) {
	if m.streamFn != nil {
		return m.streamFn(ctx, req)
	}
	ch := make(chan provider.StreamChunk, 1)
	ch <- provider.StreamChunk{Delta: "mock output", Done: true}
	close(ch)
	return ch, nil
}

func (m *mockProvider) IsAvailable(ctx context.Context) bool {
	return true
}

func (m *mockProvider) Models(ctx context.Context) ([]string, error) {
	return []string{"mock-model"}, nil
}

type mockTool struct {
	name      string
	executeFn func(ctx context.Context, args json.RawMessage) (*tool.Result, error)
}

func (m *mockTool) Name() string         { return m.name }
func (m *mockTool) Description() string  { return "Mock Tool Description" }
func (m *mockTool) Schema() *tool.Schema {
	return tool.NewSchema("object", "Args schema")
}
func (m *mockTool) Execute(ctx context.Context, args json.RawMessage) (*tool.Result, error) {
	if m.executeFn != nil {
		return m.executeFn(ctx, args)
	}
	return &tool.Result{Output: "tool success", ExitCode: 0}, nil
}

// =============================================================================
// Manifest Tests
// =============================================================================

func TestLoadManifest_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manifestPath := filepath.Join(tmpDir, "agent.yaml")
	promptPath := filepath.Join(tmpDir, "prompts", "system.md")

	if err := os.MkdirAll(filepath.Dir(promptPath), 0755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}

	promptContent := "You are a specialized code generation agent."
	if err := os.WriteFile(promptPath, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	manifestContent := `
name: coder
version: 1.0.0
role: Coder
description: Code generation agent
capabilities:
  - code_generation
provider: mock-provider
model: mock-model
prompt_file: prompts/system.md
temperature: 0.5
max_tokens: 2048
`
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	manifest, err := sdkagent.LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	if manifest.Name != "coder" {
		t.Errorf("manifest name: got %q, want %q", manifest.Name, "coder")
	}
	if manifest.SystemPrompt != promptContent {
		t.Errorf("manifest prompt: got %q, want %q", manifest.SystemPrompt, promptContent)
	}
}

// =============================================================================
// Execution Loop Tests
// =============================================================================

func TestBaseAgent_Execute_HappyPath(t *testing.T) {
	manifest := &agent.Manifest{
		Name:         "coder",
		Version:      "1.0.0",
		Role:         "Coder",
		Capabilities: []agent.Capability{agent.CapabilityCodeGeneration},
		Provider:     "mock-provider",
	}

	prov := &mockProvider{
		sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
			return &provider.Response{
				Content: "final output message",
				Usage:   provider.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			}, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	ba, err := sdkagent.NewBaseAgent(manifest, prov, logger)
	if err != nil {
		t.Fatalf("NewBaseAgent: %v", err)
	}

	ba.Init(context.Background(), nil)
	ba.Start(context.Background())

	task := &agent.Task{ID: "tsk-01", Name: "hello", Type: "code_generation"}
	res, err := ba.Execute(context.Background(), task)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if res.Output != "final output message" {
		t.Errorf("output: got %q, want %q", res.Output, "final output message")
	}
	if res.Usage.TotalTokens != 15 {
		t.Errorf("usage: got %d, want 15", res.Usage.TotalTokens)
	}
}

func TestBaseAgent_Execute_WithToolCalls(t *testing.T) {
	manifest := &agent.Manifest{
		Name:         "coder",
		Version:      "1.0.0",
		Role:         "Coder",
		Capabilities: []agent.Capability{agent.CapabilityCodeGeneration},
		Provider:     "mock-provider",
		Tools:        []string{"read_file"},
	}

	var callCount int32

	prov := &mockProvider{
		sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
			count := atomic.AddInt32(&callCount, 1)
			if count == 1 {
				return &provider.Response{
					ToolCalls: []provider.ToolCall{
						{ID: "tc-01", Name: "read_file", Args: json.RawMessage(`{"path":"main.go"}`)},
					},
				}, nil
			}
			return &provider.Response{
				Content: "final code",
			}, nil
		},
	}

	var toolExecuted int32
	mockToolInst := &mockTool{
		name: "read_file",
		executeFn: func(ctx context.Context, args json.RawMessage) (*tool.Result, error) {
			atomic.StoreInt32(&toolExecuted, 1)
			return &tool.Result{Output: "file content lines", ExitCode: 0}, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	ba, _ := sdkagent.NewBaseAgent(manifest, prov, logger)
	ba.RegisterTool(mockToolInst)

	ba.Init(context.Background(), nil)
	ba.Start(context.Background())

	task := &agent.Task{ID: "tsk-02", Name: "hello", Type: "code_generation"}
	res, err := ba.Execute(context.Background(), task)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if atomic.LoadInt32(&toolExecuted) != 1 {
		t.Error("expected tool to be executed by agent loop")
	}
	if res.Output != "final code" {
		t.Errorf("output: got %q, want %q", res.Output, "final code")
	}
}

func TestBaseAgent_Execute_MaxIterationsProtection(t *testing.T) {
	manifest := &agent.Manifest{
		Name:         "coder",
		Version:      "1.0.0",
		Role:         "Coder",
		Capabilities: []agent.Capability{agent.CapabilityCodeGeneration},
		Provider:     "mock-provider",
		Tools:        []string{"read_file"},
	}

	prov := &mockProvider{
		sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
			return &provider.Response{
				ToolCalls: []provider.ToolCall{
					{ID: "tc-loop", Name: "read_file", Args: json.RawMessage(`{}`)},
				},
			}, nil
		},
	}

	mockToolInst := &mockTool{name: "read_file"}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	ba, _ := sdkagent.NewBaseAgent(manifest, prov, logger)
	ba.RegisterTool(mockToolInst)

	ba.Init(context.Background(), nil)
	ba.Start(context.Background())

	task := &agent.Task{ID: "tsk-03", Name: "loop", Type: "code_generation"}
	_, err := ba.Execute(context.Background(), task)

	if err == nil {
		t.Fatal("expected error due to loop protection, got nil")
	}
}

func TestBaseAgent_Execute_ContextCancellation(t *testing.T) {
	manifest := &agent.Manifest{
		Name:         "coder",
		Version:      "1.0.0",
		Role:         "Coder",
		Capabilities: []agent.Capability{agent.CapabilityCodeGeneration},
		Provider:     "mock-provider",
	}

	prov := &mockProvider{
		sendFn: func(ctx context.Context, req *provider.Request) (*provider.Response, error) {
			time.Sleep(500 * time.Millisecond)
			return &provider.Response{Content: "done"}, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	ba, _ := sdkagent.NewBaseAgent(manifest, prov, logger)

	ba.Init(context.Background(), nil)
	ba.Start(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	task := &agent.Task{ID: "tsk-04", Name: "cancel", Type: "code_generation"}
	_, err := ba.Execute(ctx, task)

	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error type: got %v, want context.Canceled", err)
	}
}
```

## Rules
1. **Mock Signature Compliance**: Test mocks must match the upgraded `Provider` and `Plugin` interface definitions, using `cplugin.HealthReport` return signatures and `cplugin.TypeProvider` types.
2. **Deterministic Parallel Assertions**: Synchronize concurrent test loops using safe atomic types (`atomic.AddInt32` / `atomic.StoreInt32`) to satisfy race checkers.
3. **Task Lifecycle Verification**: Ensure tests verify initialization transitions and starting states before executing tasks.

## ⚠️ Pitfalls

### Pitfall 1: Mismatched mock method signatures breaking compile pipelines
Declaring obsolete signatures like `Health(context.Context) (any, error)` or omitting standard methods like `IsAvailable(context.Context) bool` on mock structures prevents compilation. Ensure mocks align with updated contracts.

### Pitfall 2: Race conditions in concurrent mock tracking variables
Updating mock execution counters using simple increment operations inside parallel provider calls triggers data races. Always use thread-safe atomic variables.

## Verify
```bash
go test -v -race -count=1 ./sdk/agent/...
```

## Checklist
- [ ] File `sdk/agent/agent_test.go` exists
- [ ] Package: `agent_test` (external testing package)
- [ ] Mock structures implement `plugin.Plugin` and `provider.Provider` contracts
- [ ] Mock signatures utilize updated types (`cplugin.HealthReport`, `cplugin.TypeProvider`)
- [ ] Tests verify manifest loading and relative path resolutions
- [ ] Execution tests verify normal successful runs
- [ ] Parallel tool execution is tested
- [ ] Loop iteration limit guards are verified to prevent infinite runs
- [ ] Context cancellation checks abort executions cleanly
- [ ] `go test -v -race ./sdk/agent/...` passes
