package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadFromSearchPaths_Fallback(t *testing.T) {
	// Ensure search path files don't exist temporarily if they do (using a custom configSearchPaths or just checking)
	// We can back up and modify configSearchPaths for the test.
	oldPaths := configSearchPaths
	defer func() { configSearchPaths = oldPaths }()

	configSearchPaths = []string{
		"nonexistent_file_12345.yaml",
	}

	cfg, err := LoadFromSearchPaths()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should fall back to default config
	expected := DefaultConfig()
	if cfg.Orchestrator.Name != expected.Orchestrator.Name {
		t.Errorf("expected Orchestrator name %q, got %q", expected.Orchestrator.Name, cfg.Orchestrator.Name)
	}
}

func TestParseBytes_Success(t *testing.T) {
	os.Setenv("TEST_MAX_TASKS", "10")
	os.Setenv("TEST_TIMEOUT", "60s")
	defer func() {
		os.Unsetenv("TEST_MAX_TASKS")
		os.Unsetenv("TEST_TIMEOUT")
	}()

	yamlData := []byte(`
orchestrator:
  name: "custom-orchestrator"
  max_concurrent_tasks: ${TEST_MAX_TASKS}
  shutdown_timeout: "45s"
providers:
  default: "gemini"
  configs:
    gemini:
      type: "api"
      model: "gemini-pro"
      timeout: ${TEST_TIMEOUT}
agents:
  coder:
    provider: "gemini"
    tools:
      - "read_file"
`)

	cfg, err := ParseBytes(yamlData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Orchestrator.Name != "custom-orchestrator" {
		t.Errorf("expected name 'custom-orchestrator', got %q", cfg.Orchestrator.Name)
	}
	if cfg.Orchestrator.MaxConcurrentTasks != 10 {
		t.Errorf("expected max concurrent tasks 10, got %d", cfg.Orchestrator.MaxConcurrentTasks)
	}
	if cfg.Orchestrator.ShutdownTimeout != 45*time.Second {
		t.Errorf("expected shutdown timeout 45s, got %v", cfg.Orchestrator.ShutdownTimeout)
	}
	if cfg.Providers.Default != "gemini" {
		t.Errorf("expected default provider 'gemini', got %q", cfg.Providers.Default)
	}

	geminiEntry, ok := cfg.Providers.Configs["gemini"]
	if !ok {
		t.Fatal("expected 'gemini' provider config to exist")
	}
	if geminiEntry.Timeout != 60*time.Second {
		t.Errorf("expected gemini timeout 60s, got %v", geminiEntry.Timeout)
	}

	coderAgent, ok := cfg.Agents["coder"]
	if !ok {
		t.Fatal("expected 'coder' agent config to exist")
	}
	if coderAgent.Provider != "gemini" {
		t.Errorf("expected agent provider 'gemini', got %q", coderAgent.Provider)
	}
	if len(coderAgent.Tools) != 1 || coderAgent.Tools[0] != "read_file" {
		t.Errorf("expected agent tools ['read_file'], got %v", coderAgent.Tools)
	}

	// Verify defaults merged in (e.g. security block paths/commands)
	if len(cfg.Security.BlockedCommands) == 0 {
		t.Error("expected default security blocked commands to be merged")
	}
}

func TestParseBytes_InvalidDuration(t *testing.T) {
	yamlData := []byte(`
orchestrator:
  shutdown_timeout: "invalid-duration"
`)

	_, err := ParseBytes(yamlData)
	if err == nil {
		t.Fatal("expected error parsing invalid duration, got nil")
	}
}
