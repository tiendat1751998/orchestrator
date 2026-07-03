package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
)

func TestLoadManifest_Success(t *testing.T) {
	tmpDir := t.TempDir()

	manifestContent := `
name: test-agent
version: 1.0.0
role: Developer
description: A test agent
capabilities:
  - code_generation
  - debugging
provider: antigravity
model: custom-model
tools:
  - git
system_prompt: "Write clean code."
temperature: 0.5
max_tokens: 1024
`
	manifestPath := filepath.Join(tmpDir, "agent.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("failed to write test manifest: %v", err)
	}

	m, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	if m.Name != "test-agent" {
		t.Errorf("expected Name %q, got %q", "test-agent", m.Name)
	}
	if m.Version != "1.0.0" {
		t.Errorf("expected Version %q, got %q", "1.0.0", m.Version)
	}
	if m.Role != "Developer" {
		t.Errorf("expected Role %q, got %q", "Developer", m.Role)
	}
	if len(m.Capabilities) != 2 || m.Capabilities[0] != agent.CapabilityCodeGeneration || m.Capabilities[1] != agent.CapabilityDebugging {
		t.Errorf("unexpected capabilities: %v", m.Capabilities)
	}
	if m.Provider != "antigravity" {
		t.Errorf("expected Provider %q, got %q", "antigravity", m.Provider)
	}
	if m.Model != "custom-model" {
		t.Errorf("expected Model %q, got %q", "custom-model", m.Model)
	}
	if len(m.Tools) != 1 || m.Tools[0] != "git" {
		t.Errorf("unexpected tools: %v", m.Tools)
	}
	if m.SystemPrompt != "Write clean code." {
		t.Errorf("expected SystemPrompt %q, got %q", "Write clean code.", m.SystemPrompt)
	}
	if m.Temperature != 0.5 {
		t.Errorf("expected Temperature %f, got %f", 0.5, m.Temperature)
	}
	if m.MaxTokens != 1024 {
		t.Errorf("expected MaxTokens %d, got %d", 1024, m.MaxTokens)
	}
}

func TestLoadManifest_PromptFileResolution(t *testing.T) {
	tmpDir := t.TempDir()

	// Create prompt file in a subdirectory relative to manifest
	promptsDir := filepath.Join(tmpDir, "prompts")
	if err := os.Mkdir(promptsDir, 0755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}

	promptContent := "Hello from external prompt file!"
	promptFilePath := filepath.Join(promptsDir, "system.txt")
	if err := os.WriteFile(promptFilePath, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	manifestContent := `
name: test-agent
version: 1.0.0
role: Developer
capabilities:
  - code_generation
provider: antigravity
prompt_file: prompts/system.txt
system_prompt: "This should be overwritten"
`
	manifestPath := filepath.Join(tmpDir, "agent.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatalf("failed to write test manifest: %v", err)
	}

	m, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	if m.SystemPrompt != promptContent {
		t.Errorf("expected SystemPrompt %q, got %q", promptContent, m.SystemPrompt)
	}
}

func TestLoadManifest_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "missing name",
			yaml: `
version: 1.0.0
role: Developer
capabilities: [code_generation]
provider: antigravity
`,
			wantErr: "name is required",
		},
		{
			name: "missing version",
			yaml: `
name: test-agent
role: Developer
capabilities: [code_generation]
provider: antigravity
`,
			wantErr: "version is required",
		},
		{
			name: "missing role",
			yaml: `
name: test-agent
version: 1.0.0
capabilities: [code_generation]
provider: antigravity
`,
			wantErr: "role is required",
		},
		{
			name: "missing capabilities",
			yaml: `
name: test-agent
version: 1.0.0
role: Developer
provider: antigravity
`,
			wantErr: "at least one capability is required",
		},
		{
			name: "invalid capability",
			yaml: `
name: test-agent
version: 1.0.0
role: Developer
capabilities: [invalid_capability_name]
provider: antigravity
`,
			wantErr: "invalid capability at index 0",
		},
		{
			name: "missing provider",
			yaml: `
name: test-agent
version: 1.0.0
role: Developer
capabilities: [code_generation]
`,
			wantErr: "provider is required",
		},
		{
			name: "negative temperature",
			yaml: `
name: test-agent
version: 1.0.0
role: Developer
capabilities: [code_generation]
provider: antigravity
temperature: -0.1
`,
			wantErr: "temperature must be between 0.0 and 2.0",
		},
		{
			name: "temperature too high",
			yaml: `
name: test-agent
version: 1.0.0
role: Developer
capabilities: [code_generation]
provider: antigravity
temperature: 2.1
`,
			wantErr: "temperature must be between 0.0 and 2.0",
		},
		{
			name: "negative max tokens",
			yaml: `
name: test-agent
version: 1.0.0
role: Developer
capabilities: [code_generation]
provider: antigravity
max_tokens: -5
`,
			wantErr: "max_tokens cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			manifestPath := filepath.Join(tmpDir, "agent.yaml")
			if err := os.WriteFile(manifestPath, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("failed to write test manifest: %v", err)
			}

			_, err := LoadManifest(manifestPath)
			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.wantErr)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestLoadManifest_FileErrors(t *testing.T) {
	// Empty manifest path
	_, err := LoadManifest("")
	if err == nil {
		t.Error("expected error for empty manifest path, got nil")
	}

	// Missing manifest file
	_, err = LoadManifest("nonexistent_agent.yaml")
	if err == nil {
		t.Error("expected error for nonexistent manifest path, got nil")
	}

	// Malformed YAML
	tmpDir := t.TempDir()
	malformedPath := filepath.Join(tmpDir, "malformed.yaml")
	if err := os.WriteFile(malformedPath, []byte("name: : :"), 0644); err != nil {
		t.Fatalf("failed to write test manifest: %v", err)
	}
	_, err = LoadManifest(malformedPath)
	if err == nil {
		t.Error("expected error for malformed YAML, got nil")
	}

	// Missing prompt file
	manifestWithMissingPrompt := `
name: test-agent
version: 1.0.0
role: Developer
capabilities: [code_generation]
provider: antigravity
prompt_file: nonexistent_prompt.txt
`
	missingPromptManifestPath := filepath.Join(tmpDir, "agent_missing_prompt.yaml")
	if err := os.WriteFile(missingPromptManifestPath, []byte(manifestWithMissingPrompt), 0644); err != nil {
		t.Fatalf("failed to write test manifest: %v", err)
	}
	_, err = LoadManifest(missingPromptManifestPath)
	if err == nil {
		t.Error("expected error for missing prompt file, got nil")
	}
}
