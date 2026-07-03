package config

import (
	"strings"
	"testing"
	"time"
)

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func TestValidate_Success(t *testing.T) {
	// A valid default configuration should pass validation with nil error
	cfg := DefaultConfig()
	cfg.Providers.Default = "openai"
	cfg.Providers.Configs = map[string]ProviderEntry{
		"openai": {
			Type:    "api",
			Model:   "gpt-4o",
			APIKey:  "dummy-key",
			BaseURL: "https://api.openai.com",
		},
	}

	err := Validate(cfg)
	if err != nil {
		t.Fatalf("expected nil error for valid config, got: %v", err)
	}
}

func TestValidate_OrchestratorErrors(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Orchestrator.Name = "   "
	cfg.Orchestrator.LogLevel = "invalid-level"
	cfg.Orchestrator.LogFormat = "invalid-format"
	cfg.Orchestrator.MaxConcurrentTasks = 0
	cfg.Orchestrator.ShutdownTimeout = -1 * time.Second

	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}

	ve, ok := err.(*ValidationErrors)
	if !ok {
		t.Fatalf("expected *ValidationErrors type, got %T", err)
	}

	expectedErrors := map[string]string{
		"orchestrator.name":                 "required field is empty",
		"orchestrator.log_level":            `"invalid-level" is not valid (use: debug, info, warn, error)`,
		"orchestrator.log_format":           `"invalid-format" is not valid (use: json, text)`,
		"orchestrator.max_concurrent_tasks": "must be >= 1, got 0",
		"orchestrator.shutdown_timeout":     "must be positive, got -1s",
	}

	for _, e := range ve.Errors {
		msg, ok := expectedErrors[e.Field]
		if !ok {
			t.Errorf("unexpected error field reported: %s with message %q", e.Field, e.Message)
			continue
		}
		if e.Message != msg {
			t.Errorf("expected field %s to have message %q, got %q", e.Field, msg, e.Message)
		}
		delete(expectedErrors, e.Field)
	}

	if len(expectedErrors) > 0 {
		t.Errorf("missing expected errors: %v", expectedErrors)
	}
}

func TestValidate_OrchestratorMaxTasksUpperLimit(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Orchestrator.MaxConcurrentTasks = 100

	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}

	ve, ok := err.(*ValidationErrors)
	if !ok {
		t.Fatalf("expected *ValidationErrors type, got %T", err)
	}

	found := false
	for _, e := range ve.Errors {
		if e.Field == "orchestrator.max_concurrent_tasks" {
			found = true
			if !strings.Contains(e.Message, "must be <= 50") {
				t.Errorf("expected max task limit error message, got %q", e.Message)
			}
		}
	}
	if !found {
		t.Error("expected orchestrator.max_concurrent_tasks error")
	}
}

func TestValidate_Providers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Providers.Default = "openai"
	cfg.Providers.Configs = map[string]ProviderEntry{
		"openai": {
			Type:     "api",
			Model:    "", // missing model
			BaseURL:  "", // missing base url for api
			APIKey:   "", // missing api key for api
			Timeout:  -1 * time.Second,
			MaxRetry: -1,
		},
		"gemini": {
			Type:   "cli",
			Model:  "gemini-1.5",
			Binary: "", // missing binary for cli
		},
		"local-model": {
			Type:  "invalid-type", // invalid type
			Model: "local",
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}

	ve, ok := err.(*ValidationErrors)
	if !ok {
		t.Fatalf("expected *ValidationErrors type, got %T", err)
	}

	expectedErrors := map[string]string{
		"providers.configs.openai.model":     "required field is empty",
		"providers.configs.openai.base_url":  `required when type is "api" (e.g., "https://generativelanguage.googleapis.com")`,
		"providers.configs.openai.api_key":   `required when type is "api" (use: ${YOUR_API_KEY_ENV_VAR})`,
		"providers.configs.openai.timeout":   "must be positive, got -1s",
		"providers.configs.openai.max_retry": "must be >= 0, got -1",
		"providers.configs.gemini.binary":    `required when type is "cli" (path to the CLI executable)`,
		"providers.configs.local-model.type": `"invalid-type" is not valid (use: cli, api, local)`,
	}

	for _, e := range ve.Errors {
		msg, ok := expectedErrors[e.Field]
		if !ok {
			// ignore other possible errors like Default provider not found (since default is openai and openai configs is defined)
			continue
		}
		if e.Message != msg {
			t.Errorf("expected field %s to have message %q, got %q", e.Field, msg, e.Message)
		}
		delete(expectedErrors, e.Field)
	}

	if len(expectedErrors) > 0 {
		t.Errorf("missing expected errors: %v", expectedErrors)
	}
}

func TestValidate_ProvidersDefaultNotFound(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Providers.Default = "missing-provider"
	cfg.Providers.Configs = map[string]ProviderEntry{
		"openai": {
			Type:  "local",
			Model: "model",
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}

	ve, ok := err.(*ValidationErrors)
	if !ok {
		t.Fatalf("expected *ValidationErrors type, got %T", err)
	}

	found := false
	for _, e := range ve.Errors {
		if e.Field == "providers.default" {
			found = true
			if !strings.Contains(e.Message, `provider "missing-provider" not found in configured providers: [openai]`) {
				t.Errorf("unexpected error message: %q", e.Message)
			}
		}
	}
	if !found {
		t.Error("expected providers.default not found error")
	}
}

func TestValidate_Agents(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Providers.Default = "openai"
	cfg.Providers.Configs = map[string]ProviderEntry{
		"openai": {
			Type:  "local",
			Model: "model",
		},
	}
	cfg.Agents = map[string]AgentConfig{
		"agent-1": {
			Provider:    "nonexistent",
			Temperature: floatPtr(2.5),
			MaxTokens:   intPtr(0),
		},
		"agent-2": {
			Provider:    "openai",
			Temperature: floatPtr(-0.1),
			MaxTokens:   intPtr(-10),
		},
		"agent-3": {
			Provider:    "openai",
			Temperature: nil, // Valid
			MaxTokens:   nil, // Valid
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}

	ve, ok := err.(*ValidationErrors)
	if !ok {
		t.Fatalf("expected *ValidationErrors type, got %T", err)
	}

	expectedErrors := map[string]string{
		"agents.agent-1.provider":    `provider "nonexistent" not found in configured providers`,
		"agents.agent-1.temperature": "must be between 0.0 and 2.0, got 2.500000",
		"agents.agent-1.max_tokens":  "must be >= 1, got 0",
		"agents.agent-2.temperature": "must be between 0.0 and 2.0, got -0.100000",
		"agents.agent-2.max_tokens":  "must be >= 1, got -10",
	}

	for _, e := range ve.Errors {
		msg, ok := expectedErrors[e.Field]
		if !ok {
			continue
		}
		if e.Message != msg {
			t.Errorf("expected field %s to have message %q, got %q", e.Field, msg, e.Message)
		}
		delete(expectedErrors, e.Field)
	}

	if len(expectedErrors) > 0 {
		t.Errorf("missing expected errors: %v", expectedErrors)
	}
}

func TestValidate_Security(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Security.MaxFileSize = 0
	cfg.Security.MaxOutputSize = -5

	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}

	ve, ok := err.(*ValidationErrors)
	if !ok {
		t.Fatalf("expected *ValidationErrors type, got %T", err)
	}

	expectedErrors := map[string]string{
		"security.max_file_size":   "must be >= 1 byte, got 0",
		"security.max_output_size": "must be >= 1 byte, got -5",
	}

	for _, e := range ve.Errors {
		msg, ok := expectedErrors[e.Field]
		if !ok {
			continue
		}
		if e.Message != msg {
			t.Errorf("expected field %s to have message %q, got %q", e.Field, msg, e.Message)
		}
		delete(expectedErrors, e.Field)
	}

	if len(expectedErrors) > 0 {
		t.Errorf("missing expected errors: %v", expectedErrors)
	}
}

func TestValidationErrors_ErrorFormat(t *testing.T) {
	ve := &ValidationErrors{}
	if ve.Error() != "config: no validation errors" {
		t.Errorf("unexpected error format for empty errors: %q", ve.Error())
	}

	ve.Add("field1", "error1")
	ve.Addf("field2", "error %d", 2)

	expected := "config: 2 validation error(s):\n  1. field1: error1\n  2. field2: error 2\n"
	if ve.Error() != expected {
		t.Errorf("expected format:\n%q\ngot:\n%q", expected, ve.Error())
	}
}
