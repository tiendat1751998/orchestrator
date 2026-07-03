package agents_test

import (
	"context"
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

	// 2. Initialize Backend Agent using temporary manifest path and a mock provider
	mockProv := &sdktesting.MockProvider{NameVal: "mock-provider"}
	agent, err := backend.NewBackendAgent(manifestPath, mockProv, nil)
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

	task := &cagent.Task{
		ID:    "task-1",
		Type:  "code_generation",
		Input: map[string]any{},
	}

	// Should complete task loop using mock provider responses
	_, err = agent.Execute(ctx, task)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

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

	mockProv := &sdktesting.MockProvider{NameVal: "mock-provider"}
	ra, err := reviewer.NewReviewerAgent(manifestPath, mockProv, nil)
	if err != nil {
		t.Fatalf("failed to construct reviewer agent: %v", err)
	}

	if ra.Name() != "test-reviewer" {
		t.Errorf("expected agent name 'test-reviewer', got %q", ra.Name())
	}
}

func TestAgents_NilProvider_ErrorHandling(t *testing.T) {
	// This test specifically validates the user requested behavior of passing nil provider,
	// asserting that both constructors return an appropriate validation error.
	tmpDir, _ := os.MkdirTemp("", "orchestrator-nil-provider-test-*")
	defer os.RemoveAll(tmpDir)

	promptsDir := filepath.Join(tmpDir, "prompts")
	_ = os.MkdirAll(promptsDir, 0755)
	_ = os.WriteFile(filepath.Join(promptsDir, "system.md"), []byte("# Mock"), 0644)

	yamlContent := `
name: "test-nil"
version: "0.1.0"
role: "Mock"
description: "Mock"
capabilities:
  - "code_generation"
provider: "mock-provider"
model: "mock-model"
tools: []
prompt_file: "prompts/system.md"
`
	manifestPath := filepath.Join(tmpDir, "agent.yaml")
	_ = os.WriteFile(manifestPath, []byte(yamlContent), 0644)

	_, err := backend.NewBackendAgent(manifestPath, nil, nil)
	if err == nil {
		t.Error("expected error constructing BackendAgent with nil provider, got nil")
	}

	_, err = reviewer.NewReviewerAgent(manifestPath, nil, nil)
	if err == nil {
		t.Error("expected error constructing ReviewerAgent with nil provider, got nil")
	}
}
