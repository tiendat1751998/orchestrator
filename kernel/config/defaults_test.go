package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("expected DefaultConfig() to return non-nil value")
	}
}

func TestMergeWithDefaults(t *testing.T) {
	cfg := &Config{}
	MergeWithDefaults(cfg)

	// Verify Orchestrator defaults are filled
	if cfg.Orchestrator.Name != "orchestrator" {
		t.Errorf("expected Name 'orchestrator', got %q", cfg.Orchestrator.Name)
	}
	if cfg.Orchestrator.MaxConcurrentTasks != 5 {
		t.Errorf("expected MaxConcurrentTasks 5, got %d", cfg.Orchestrator.MaxConcurrentTasks)
	}
	if cfg.Orchestrator.ShutdownTimeout != 30*time.Second {
		t.Errorf("expected ShutdownTimeout 30s, got %v", cfg.Orchestrator.ShutdownTimeout)
	}

	// Verify Security defaults are filled
	if cfg.Security.MaxFileSize != 1*1024*1024 {
		t.Errorf("expected MaxFileSize 1MB, got %d", cfg.Security.MaxFileSize)
	}
	if len(cfg.Security.BlockedCommands) == 0 {
		t.Error("expected default BlockedCommands to be populated")
	}

	// Verify partial config is preserved but other fields are filled
	cfgPartial := &Config{
		Orchestrator: OrchestratorConfig{
			Name: "custom-orchestrator",
		},
		Providers: ProvidersConfig{
			Configs: map[string]ProviderEntry{
				"custom": {
					Type: "api",
				},
			},
		},
	}
	MergeWithDefaults(cfgPartial)

	if cfgPartial.Orchestrator.Name != "custom-orchestrator" {
		t.Errorf("expected custom Name to be preserved, got %q", cfgPartial.Orchestrator.Name)
	}
	if cfgPartial.Orchestrator.LogLevel != "info" {
		t.Errorf("expected default LogLevel to be filled, got %q", cfgPartial.Orchestrator.LogLevel)
	}
	customEntry, exists := cfgPartial.Providers.Configs["custom"]
	if !exists {
		t.Fatal("expected 'custom' provider to exist")
	}
	if customEntry.Type != "api" {
		t.Errorf("expected custom entry type to be preserved, got %q", customEntry.Type)
	}
	if customEntry.Timeout != 120*time.Second {
		t.Errorf("expected default timeout to be filled for custom provider, got %v", customEntry.Timeout)
	}
}
