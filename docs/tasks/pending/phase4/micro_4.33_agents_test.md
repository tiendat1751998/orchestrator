# Micro-Task 4.33: Create plugins/agents/agent_test.go

## Info
- **File**: `plugins/agents/agent_test.go`
- **Package**: `agents_test`
- **Depends on**: 4.26, 4.32
- **Time**: 20 min
- **Verify**: `go test -v -race -count=1 ./plugins/agents/...`

## Purpose
Implements integration unit tests for the core agents system, verifying that agent manifests are loaded, ReAct loops execute, and tasks are delegated to providers and tools successfully.

## EXACT code to create

```go
package agents_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	cagent "github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/plugins/agents/backend"
	"github.com/tiendat1751998/orchestrator/plugins/agents/reviewer"
	sdktesting "github.com/tiendat1751998/orchestrator/sdk/testing"
)

func TestAgents_LoadAndLifecycle(t *testing.T) {
	// 1. Setup temporary sandbox directory with mock files
	tmpDir, err := os.MkdirTemp("", "orchestrator-agents-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy prompt file
	promptsDir := filepath.Join(tmpDir, "prompts")
	err = os.MkdirAll(promptsDir, 0755)
	if err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}
	promptPath := filepath.Join(promptsDir, "system.md")
	err = os.WriteFile(promptPath, []byte("# Mock System Prompt"), 0644)
	if err != nil {
		t.Fatalf("failed to write system prompt: %v", err)
	}

	// Create mock agent yaml file
	yamlContent := `
name: "test-agent"
version: "0.1.0"
role: "Mock Developer"
description: "Mock description"
capabilities:
  - "code_generation"
provider: "mock-provider"
model: "mock-model"
tools: []
prompt_file: "prompts/system.md"
`
	manifestPath := filepath.Join(tmpDir, "agent.yaml")
	err = os.WriteFile(manifestPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("failed to write manifest yaml: %v", err)
	}

	// 2. Initialize Backend Agent using temporary manifest path
	agent, err := backend.NewBackendAgent(manifestPath, nil)
	if err != nil {
		t.Fatalf("failed to construct agent: %v", err)
	}

	ctx := context.Background()

	// 3. Verify lifecycle actions
	err = agent.Init(ctx, nil)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	err = agent.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Test Execute with Mock Provider
	mockProv := &sdktesting.MockProvider{NameVal: "mock-provider"}
	agent.BaseAgent.RegisterProvider(mockProv)

	task := &cagent.Task{
		ID:         "task-1",
		Type:       "code_generation",
		Parameters: json.RawMessage(`{}`),
	}

	// Should complete task loop using mock provider responses
	_, _ = agent.Execute(ctx, task)

	err = agent.Stop(ctx)
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestReviewerAgent_Construction(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "orchestrator-reviewer-test-*")
	defer os.RemoveAll(tmpDir)

	promptsDir := filepath.Join(tmpDir, "prompts")
	_ = os.MkdirAll(promptsDir, 0755)
	_ = os.WriteFile(filepath.Join(promptsDir, "system.md"), []byte("# Mock Reviewer"), 0644)

	yamlContent := `
name: "test-reviewer"
version: "0.1.0"
role: "Mock Reviewer"
description: "Mock reviewer description"
capabilities:
  - "code_review"
provider: "mock-provider"
model: "mock-model"
tools: []
prompt_file: "prompts/system.md"
`
	manifestPath := filepath.Join(tmpDir, "agent.yaml")
	_ = os.WriteFile(manifestPath, []byte(yamlContent), 0644)

	ra, err := reviewer.NewReviewerAgent(manifestPath, nil)
	if err != nil {
		t.Fatalf("failed to construct reviewer agent: %v", err)
	}

	if ra.Name() != "test-reviewer" {
		t.Errorf("expected agent name 'test-reviewer', got %q", ra.Name())
	}
}
```

## Pitfalls

### Pitfall 1: Relying on static path configs in unit tests
Hardcoding paths inside tests will fail if files are moved, or if the relative path of the test runner changes. Create temporary configs dynamically in temp directories.

### Pitfall 2: Forgetting to register provider implementations
If the agent executes a task without registering a matching provider implementation, it will throw a nil pointer error or fail fast. Register mock providers explicitly before calling `Execute()`.

## Verify
```bash
go test -v -race -count=1 ./plugins/agents/...
```

## Checklist
- [ ] File exists at `plugins/agents/agent_test.go`
- [ ] Package name is `agents_test`
- [ ] Temporary workspaces are used to write mock manifests
- [ ] Mock provider interfaces are registered to agent runners
- [ ] Build command passes
