# Micro-Task 2.06: Tạo kernel/config/config_test.go

## Thông tin
- **File tạo**: `kernel/config/config_test.go`
- **Package**: `config_test`
- **Dependencies trước**: 2.01-2.05
- **Thời gian**: 25 phút
- **Verify**: `go test -v -race ./kernel/config/...`

## Nội dung CHÍNH XÁC cần tạo

```go
package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/kernel/config"
)

// =============================================================================
// Helper: tạo temporary YAML file
// =============================================================================

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write temp YAML: %v", err)
	}
	return path
}

// =============================================================================
// Default Config Tests
// =============================================================================

func TestDefaultConfig_HasAllDefaults(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg.Orchestrator.Name != "orchestrator" {
		t.Errorf("Name: got %q, want %q", cfg.Orchestrator.Name, "orchestrator")
	}
	if cfg.Orchestrator.LogLevel != "info" {
		t.Errorf("LogLevel: got %q, want %q", cfg.Orchestrator.LogLevel, "info")
	}
	if cfg.Orchestrator.LogFormat != "text" {
		t.Errorf("LogFormat: got %q, want %q", cfg.Orchestrator.LogFormat, "text")
	}
	if cfg.Orchestrator.MaxConcurrentTasks != 5 {
		t.Errorf("MaxConcurrentTasks: got %d, want 5", cfg.Orchestrator.MaxConcurrentTasks)
	}
	if cfg.Orchestrator.ShutdownTimeout != 30*time.Second {
		t.Errorf("ShutdownTimeout: got %v, want 30s", cfg.Orchestrator.ShutdownTimeout)
	}
	if cfg.Security.MaxFileSize != 1*1024*1024 {
		t.Errorf("MaxFileSize: got %d, want 1MB", cfg.Security.MaxFileSize)
	}
	if len(cfg.Security.BlockedCommands) < 5 {
		t.Errorf("BlockedCommands: got %d, want >= 5", len(cfg.Security.BlockedCommands))
	}
}

// =============================================================================
// Load from YAML Tests
// =============================================================================

func TestLoad_ValidYAML(t *testing.T) {
	yaml := `
orchestrator:
  name: "test-orchestrator"
  log_level: "debug"
  log_format: "json"
  data_dir: "/tmp/data"
  max_concurrent_tasks: 3
  shutdown_timeout: "10s"
providers:
  default: "test-provider"
  configs:
    test-provider:
      type: "cli"
      model: "test-model"
      binary: "/usr/bin/test"
      timeout: "60s"
      max_retry: 2
security:
  sandbox: true
  max_file_size: 2097152
  max_output_size: 204800
`
	path := writeTempYAML(t, yaml)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Orchestrator.Name != "test-orchestrator" {
		t.Errorf("Name: got %q", cfg.Orchestrator.Name)
	}
	if cfg.Orchestrator.LogLevel != "debug" {
		t.Errorf("LogLevel: got %q", cfg.Orchestrator.LogLevel)
	}
	if cfg.Orchestrator.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout: got %v, want 10s", cfg.Orchestrator.ShutdownTimeout)
	}

	provider, ok := cfg.Providers.Configs["test-provider"]
	if !ok {
		t.Fatal("test-provider not found")
	}
	if provider.Timeout != 60*time.Second {
		t.Errorf("Provider Timeout: got %v, want 60s", provider.Timeout)
	}
	if provider.Binary != "/usr/bin/test" {
		t.Errorf("Provider Binary: got %q", provider.Binary)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/settings.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeTempYAML(t, `invalid: yaml: [broken`)
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoad_InvalidDuration(t *testing.T) {
	yaml := `
orchestrator:
  name: "test"
  log_level: "info"
  log_format: "text"
  shutdown_timeout: "not-a-duration"
providers:
  default: "x"
  configs: {}
`
	path := writeTempYAML(t, yaml)
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected error for invalid duration")
	}
}

// =============================================================================
// Env Var Resolution Tests
// =============================================================================

func TestResolveEnvVars_SimpleVar(t *testing.T) {
	t.Setenv("TEST_ORCHESTRATOR_VAR", "hello-world")

	result, err := config.ResolveEnvVars("${TEST_ORCHESTRATOR_VAR}")
	if err != nil {
		t.Fatalf("ResolveEnvVars: %v", err)
	}
	if result != "hello-world" {
		t.Errorf("got %q, want %q", result, "hello-world")
	}
}

func TestResolveEnvVars_PartialVar(t *testing.T) {
	t.Setenv("TEST_USER", "alice")

	result, err := config.ResolveEnvVars("/home/${TEST_USER}/data")
	if err != nil {
		t.Fatalf("ResolveEnvVars: %v", err)
	}
	if result != "/home/alice/data" {
		t.Errorf("got %q, want %q", result, "/home/alice/data")
	}
}

func TestResolveEnvVars_MissingSingleVar_Error(t *testing.T) {
	_, err := config.ResolveEnvVars("${DEFINITELY_NOT_SET_12345}")
	if err == nil {
		t.Error("expected error for missing required env var")
	}
}

func TestResolveEnvVars_NoVars(t *testing.T) {
	result, err := config.ResolveEnvVars("plain string without vars")
	if err != nil {
		t.Fatalf("ResolveEnvVars: %v", err)
	}
	if result != "plain string without vars" {
		t.Errorf("got %q", result)
	}
}

func TestLoad_WithEnvVar(t *testing.T) {
	t.Setenv("TEST_API_KEY_2", "secret-key-123")

	yaml := `
orchestrator:
  name: "env-test"
  log_level: "info"
  log_format: "text"
providers:
  default: "api-provider"
  configs:
    api-provider:
      type: "api"
      model: "test-model"
      base_url: "https://api.example.com"
      api_key: "${TEST_API_KEY_2}"
      timeout: "30s"
security:
  sandbox: true
  max_file_size: 1048576
  max_output_size: 102400
`
	path := writeTempYAML(t, yaml)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	provider := cfg.Providers.Configs["api-provider"]
	if provider.APIKey != "secret-key-123" {
		t.Errorf("APIKey: got %q, want %q", provider.APIKey, "secret-key-123")
	}
}

// =============================================================================
// Merge Defaults Tests
// =============================================================================

func TestMergeWithDefaults_FillsMissing(t *testing.T) {
	yaml := `
orchestrator:
  name: "custom-name"
providers:
  default: "antigravity"
  configs: {}
`
	path := writeTempYAML(t, yaml)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Name should be custom, others should be defaults
	if cfg.Orchestrator.Name != "custom-name" {
		t.Errorf("Name: got %q, want %q", cfg.Orchestrator.Name, "custom-name")
	}
	if cfg.Orchestrator.LogLevel != "info" {
		t.Errorf("LogLevel: got %q, want default %q", cfg.Orchestrator.LogLevel, "info")
	}
	if cfg.Orchestrator.MaxConcurrentTasks != 5 {
		t.Errorf("MaxConcurrentTasks: got %d, want default 5", cfg.Orchestrator.MaxConcurrentTasks)
	}
}

// =============================================================================
// Validation Tests
// =============================================================================

func TestValidate_ValidConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers.Configs["antigravity"] = config.ProviderEntry{
		Type:  "cli",
		Model: "gemini-2.5-pro",
		Binary: "antigravity",
	}

	err := config.Validate(cfg)
	if err != nil {
		t.Errorf("Validate: unexpected error: %v", err)
	}
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Orchestrator.LogLevel = "verbose" // invalid
	cfg.Providers.Configs["antigravity"] = config.ProviderEntry{
		Type:  "cli",
		Model: "test",
		Binary: "test",
	}

	err := config.Validate(cfg)
	if err == nil {
		t.Error("expected validation error for invalid log level")
	}

	verrs, ok := err.(*config.ValidationErrors)
	if !ok {
		t.Fatalf("expected *ValidationErrors, got %T", err)
	}
	if len(verrs.Errors) == 0 {
		t.Error("expected at least 1 validation error")
	}
}

func TestValidate_MissingDefaultProvider(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers.Configs["openai"] = config.ProviderEntry{
		Type:  "api",
		Model: "gpt-4",
		BaseURL: "https://api.openai.com",
		APIKey: "sk-test",
	}
	// Default is "antigravity" but only "openai" is configured

	err := config.Validate(cfg)
	if err == nil {
		t.Error("expected validation error: default provider not in configs")
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := &config.Config{} // Empty config

	err := config.Validate(cfg)
	if err == nil {
		t.Fatal("expected validation errors")
	}

	verrs, ok := err.(*config.ValidationErrors)
	if !ok {
		t.Fatalf("expected *ValidationErrors, got %T", err)
	}

	// Should have multiple errors (name, log_level, log_format, etc.)
	if len(verrs.Errors) < 3 {
		t.Errorf("expected >= 3 errors, got %d: %s", len(verrs.Errors), err)
	}
}

func TestValidate_CLIProviderRequiresBinary(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers.Configs["antigravity"] = config.ProviderEntry{
		Type:  "cli",
		Model: "test",
		// Binary is empty → should fail
	}

	err := config.Validate(cfg)
	if err == nil {
		t.Error("expected error: CLI provider needs binary path")
	}
}

func TestValidate_AgentReferencesInvalidProvider(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers.Configs["antigravity"] = config.ProviderEntry{
		Type:   "cli",
		Model:  "test",
		Binary: "test",
	}
	cfg.Agents["backend"] = config.AgentConfig{
		Provider: "nonexistent", // This provider doesn't exist
	}

	err := config.Validate(cfg)
	if err == nil {
		t.Error("expected error: agent references nonexistent provider")
	}
}

// =============================================================================
// ParseBytes Tests (for testing without filesystem)
// =============================================================================

func TestParseBytes_MinimalConfig(t *testing.T) {
	yaml := []byte(`
orchestrator:
  name: "test"
providers:
  default: "x"
  configs: {}
`)
	cfg, err := config.ParseBytes(yaml)
	if err != nil {
		t.Fatalf("ParseBytes: %v", err)
	}
	if cfg.Orchestrator.Name != "test" {
		t.Errorf("Name: got %q", cfg.Orchestrator.Name)
	}
}
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: t.Setenv() cho env var tests
```go
t.Setenv("MY_VAR", "value")  // Go 1.17+
// Automatically cleaned up after test — NO manual os.Unsetenv needed.
// DO NOT use os.Setenv in tests — parallel tests will interfere.
```

### Pitfall 2: t.TempDir() cho temp files
```go
dir := t.TempDir()  // Auto-cleanup, unique per test, parallel-safe
// DO NOT use os.MkdirTemp — must manually cleanup.
```

### Pitfall 3: Type assertion cho ValidationErrors
```go
verrs, ok := err.(*config.ValidationErrors)
if !ok {
    t.Fatalf("expected *ValidationErrors, got %T", err)
}
```
PHẢI check `ok` — nếu err là khác type (ví dụ file not found) → panic nếu assert trực tiếp.

### Pitfall 4: Test cả success VÀ failure paths
- Valid config → no error
- Invalid log level → validation error
- Missing file → file error  
- Invalid YAML → parse error
- Missing env var → env error
Nếu chỉ test success → bugs hide in error paths.

## Lệnh verify
```bash
go test -v -race -count=1 ./kernel/config/...
# Expected: ALL PASS, ≥ 15 test functions
# -count=1: disable test caching
# -race: detect data races
```

## Checklist
- [ ] File `kernel/config/config_test.go` tồn tại
- [ ] Package: `config_test` (external test package)
- [ ] `writeTempYAML()` helper function
- [ ] ≥ 15 test functions
- [ ] Tests for: DefaultConfig, Load, invalid YAML, invalid duration
- [ ] Tests for: ResolveEnvVars (simple, partial, missing, none)
- [ ] Tests for: Load with env vars
- [ ] Tests for: MergeWithDefaults
- [ ] Tests for: Validate (valid, invalid log level, missing provider, multiple errors)
- [ ] Tests for: cross-field validation (agent → provider)
- [ ] Tests for: ParseBytes (no filesystem)
- [ ] Dùng `t.Setenv()` (KHÔNG `os.Setenv()`)
- [ ] Dùng `t.TempDir()` (KHÔNG `os.MkdirTemp()`)
- [ ] `go test -v -race ./kernel/config/...` ALL PASS
