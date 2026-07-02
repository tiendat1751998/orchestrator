# Micro-Task 2.01: Tạo kernel/config/config.go

## Thông tin
- **File tạo**: `kernel/config/config.go`
- **Package**: `config`
- **Dependencies trước**: Phase 1 hoàn thành (contracts package)
- **Thời gian**: 20 phút
- **Verify**: `go build ./kernel/config/...`

## Mục đích
File này định nghĩa Go structs mapping 1:1 với YAML config file.
KHÔNG chứa logic load/validate — chỉ struct definitions.

## External dependency cần thêm vào go.mod
```bash
go get gopkg.in/yaml.v3
```

> ⚠️ SAU khi thêm dependency, chạy `go mod tidy` để clean up.

## Nội dung CHÍNH XÁC cần tạo

```go
// Package config handles loading, parsing, and validating the orchestrator
// configuration from YAML files and environment variables.
//
// Configuration is loaded from (in priority order):
//   1. CLI --config flag path
//   2. .orchestrator/settings.yaml (project directory)
//   3. ~/.orchestrator/settings.yaml (user home)
//   4. Built-in defaults
//
// Environment variables override YAML values.
// Syntax: ${ENV_VAR_NAME} in YAML values.
package config

import "time"

// Config is the root configuration structure.
//
// It maps 1:1 to the YAML config file:
//
//	orchestrator:
//	  name: "my-orchestrator"
//	  log_level: "info"
//	  ...
//	providers:
//	  default: "antigravity"
//	  ...
//	security:
//	  sandbox: true
//	  ...
//
// IMPORTANT: This struct is passed by pointer to kernel components via
// constructor injection. It is NEVER stored as a global variable.
// Global state makes testing impossible and creates hidden coupling.
type Config struct {
	// Orchestrator contains general orchestrator settings.
	Orchestrator OrchestratorConfig `yaml:"orchestrator" json:"orchestrator"`

	// Providers contains AI provider configurations.
	// Key: provider name (e.g., "antigravity", "gemini-api")
	// Value: provider-specific config
	Providers ProvidersConfig `yaml:"providers" json:"providers"`

	// Agents contains agent configurations.
	// Key: agent name (e.g., "backend", "reviewer")
	// Value: agent-specific config
	Agents map[string]AgentConfig `yaml:"agents,omitempty" json:"agents,omitempty"`

	// Security contains security policy settings.
	Security SecurityConfig `yaml:"security" json:"security"`
}

// OrchestratorConfig contains general orchestrator settings.
type OrchestratorConfig struct {
	// Name identifies this orchestrator instance.
	// Used in logs and event sources.
	// Default: "orchestrator"
	Name string `yaml:"name" json:"name"`

	// LogLevel controls log verbosity.
	// Valid values: "debug", "info", "warn", "error"
	// Default: "info"
	//
	// WHY string instead of slog.Level?
	// → YAML deserialization: "info" is human-readable in config files.
	// → Conversion to slog.Level happens in the logger package.
	// → Config package should NOT import log/slog (separation of concerns).
	LogLevel string `yaml:"log_level" json:"log_level"`

	// LogFormat controls log output format.
	// Valid values: "json" (structured), "text" (human-readable)
	// Default: "text"
	//
	// "json" for production (machine-parseable, works with log aggregators).
	// "text" for development (colorized, readable in terminal).
	LogFormat string `yaml:"log_format" json:"log_format"`

	// DataDir is the directory for orchestrator data files.
	// Stores: mission history, audit logs, cached data.
	// Default: ".orchestrator/data"
	//
	// WHY configurable?
	// → Different projects may want different data locations.
	// → CI environments may need /tmp or ephemeral storage.
	DataDir string `yaml:"data_dir" json:"data_dir"`

	// MaxConcurrentTasks limits how many tasks run simultaneously.
	// Set to 1 for sequential execution (debugging).
	// Default: 5
	//
	// WHY default 5?
	// → Most AI providers have rate limits (5-10 requests/minute).
	// → Higher values may trigger rate limiting.
	// → Lower values underutilize available resources.
	MaxConcurrentTasks int `yaml:"max_concurrent_tasks" json:"max_concurrent_tasks"`

	// ShutdownTimeout is the maximum time to wait for graceful shutdown.
	// After this timeout, in-flight tasks are forcefully cancelled.
	// Default: 30s
	//
	// WHY time.Duration and not string?
	// → We use a custom YAML unmarshaller (in loader.go) that converts
	//   "30s", "5m", "1h" strings into time.Duration values.
	// → This gives us type safety in Go code.
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout"`
}

// ProvidersConfig contains all provider configurations.
type ProvidersConfig struct {
	// Default is the name of the default provider.
	// Used when an agent doesn't specify a provider.
	// Must match a key in the Configs map.
	Default string `yaml:"default" json:"default"`

	// Configs maps provider name → provider configuration.
	// At least one provider must be configured.
	//
	// Example YAML:
	//   providers:
	//     default: "antigravity"
	//     configs:
	//       antigravity:
	//         type: "cli"
	//         binary: "antigravity"
	//         model: "gemini-2.5-pro"
	Configs map[string]ProviderEntry `yaml:"configs" json:"configs"`
}

// ProviderEntry is the config for a single provider.
type ProviderEntry struct {
	// Type indicates communication method: "cli", "api", "local"
	Type string `yaml:"type" json:"type"`

	// Model is the default model name.
	Model string `yaml:"model" json:"model"`

	// Binary is the CLI executable path (for Type="cli").
	Binary string `yaml:"binary,omitempty" json:"binary,omitempty"`

	// BaseURL is the API endpoint (for Type="api").
	BaseURL string `yaml:"base_url,omitempty" json:"base_url,omitempty"`

	// APIKey is the authentication key.
	// SECURITY: Use "${ENV_VAR}" syntax. NEVER hardcode.
	//
	// json:"-" prevents this from appearing in JSON API responses.
	// YAML tag is kept so the config file can reference env vars.
	APIKey string `yaml:"api_key,omitempty" json:"-"`

	// Timeout for provider requests.
	// Default: 120s
	Timeout time.Duration `yaml:"timeout" json:"timeout"`

	// MaxRetry is retry attempts on transient failures.
	// Default: 3
	MaxRetry int `yaml:"max_retry" json:"max_retry"`

	// Extra holds provider-specific settings.
	Extra map[string]string `yaml:"extra,omitempty" json:"extra,omitempty"`
}

// AgentConfig is the config for a single agent.
type AgentConfig struct {
	// Provider is the name of the provider this agent uses.
	// Must match a key in ProvidersConfig.Configs.
	// If empty, uses ProvidersConfig.Default.
	Provider string `yaml:"provider,omitempty" json:"provider,omitempty"`

	// Model overrides the provider's default model.
	Model string `yaml:"model,omitempty" json:"model,omitempty"`

	// PromptFile is the path to the system prompt file.
	// Path is relative to the project root.
	PromptFile string `yaml:"prompt_file,omitempty" json:"prompt_file,omitempty"`

	// Temperature overrides the default temperature.
	Temperature *float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`

	// MaxTokens overrides the default max tokens.
	MaxTokens *int `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty"`

	// Tools lists specific tools this agent can use.
	// If empty, agent can use ALL registered tools.
	// If non-empty, agent can ONLY use listed tools.
	Tools []string `yaml:"tools,omitempty" json:"tools,omitempty"`
}

// SecurityConfig contains security policy settings.
type SecurityConfig struct {
	// Sandbox enables command sandboxing.
	// When true, all commands are checked against BlockedCommands before execution.
	// Default: true
	Sandbox bool `yaml:"sandbox" json:"sandbox"`

	// AllowedPaths are filesystem paths agents can access.
	// If empty, agents can access the entire project directory.
	// Paths are relative to the project root or absolute.
	// Example: ["src/", "tests/", "docs/"]
	AllowedPaths []string `yaml:"allowed_paths,omitempty" json:"allowed_paths,omitempty"`

	// BlockedPaths are filesystem paths agents CANNOT access.
	// Takes precedence over AllowedPaths.
	// Example: [".env", ".git/config", "secrets/"]
	BlockedPaths []string `yaml:"blocked_paths,omitempty" json:"blocked_paths,omitempty"`

	// BlockedCommands are shell commands that are always rejected.
	// Matched by prefix: "rm -rf" blocks "rm -rf /", "rm -rf ~", etc.
	// Default: ["rm -rf /", "sudo", "chmod 777", "mkfs", "dd if="]
	BlockedCommands []string `yaml:"blocked_commands,omitempty" json:"blocked_commands,omitempty"`

	// MaxFileSize is the maximum file size an agent can read/write (bytes).
	// Default: 1MB (1048576 bytes)
	// Files larger than this are rejected to prevent memory issues.
	MaxFileSize int64 `yaml:"max_file_size" json:"max_file_size"`

	// MaxOutputSize is the maximum command output size (bytes).
	// Default: 100KB (102400 bytes)
	// Command outputs larger than this are truncated.
	MaxOutputSize int64 `yaml:"max_output_size" json:"max_output_size"`

	// AuditLog enables audit logging of all agent actions.
	// Default: true
	AuditLog bool `yaml:"audit_log" json:"audit_log"`
}
```

## ⚠️ Pitfalls cần tránh (QUAN TRỌNG — đọc kỹ)

### Pitfall 1: Global variable
```go
// ❌ SAI — TUYỆT ĐỐI KHÔNG LÀM:
var GlobalConfig *Config

// ✅ ĐÚNG — Truyền qua constructor:
func NewKernel(cfg *Config) *Kernel { ... }
```
Global state → tests chạy song song sẽ fail (shared state). Dependency injection via constructor = testable code.

### Pitfall 2: time.Duration trong YAML
Go standard `yaml.v3` KHÔNG tự parse `"120s"` thành `time.Duration`. Giá trị YAML `120s` sẽ gây lỗi unmarshal.
Giải pháp: custom YAML unmarshal function (sẽ implement trong micro-task 2.04 loader.go).
File config.go chỉ ĐỊNH NGHĨA struct, KHÔNG giải quyết vấn đề parsing.

### Pitfall 3: APIKey json:"-"
`json:"-"` tag ngăn APIKey leak qua JSON serialization (ví dụ: API /config endpoint trả về config).
YAML tag `yaml:"api_key"` vẫn hoạt động bình thường cho config loading.

### Pitfall 4: Pointer types cho optional overrides
`Temperature *float64` và `MaxTokens *int` trong AgentConfig.
- `nil` = "không override, dùng default của provider"  
- `&0.0` = "override thành 0.0 (deterministic output)"
Nếu dùng value type `float64` → không phân biệt được "không set" vs "set = 0".

### Pitfall 5: Config package KHÔNG import log/slog
LogLevel là `string` ("info"), KHÔNG phải `slog.Level`. 
Conversion `string → slog.Level` xảy ra trong logger package.
Config package nên có minimal dependencies.

### Pitfall 6: ProvidersConfig.Configs naming
YAML structure là nested map:
```yaml
providers:
  default: "antigravity"
  configs:
    antigravity:
      type: "cli"
```
KHÔNG PHẢI:
```yaml
providers:
  antigravity:
    type: "cli"
```
Lý do: field `default` nằm cùng level với provider entries → cần tách `configs` sub-map để tránh conflict.

## Lệnh verify
```bash
go get gopkg.in/yaml.v3
go mod tidy
go build ./kernel/config/...
go vet ./kernel/config/...
```

## Checklist
- [ ] File `kernel/config/config.go` tồn tại
- [ ] Package: `package config`
- [ ] Config struct với 4 top-level fields (Orchestrator, Providers, Agents, Security)
- [ ] OrchestratorConfig với 6 fields
- [ ] ProvidersConfig với 2 fields (Default, Configs)
- [ ] ProviderEntry với 8 fields
- [ ] AgentConfig với 6 fields
- [ ] SecurityConfig với 7 fields
- [ ] Temperature và MaxTokens dùng pointer types (`*float64`, `*int`)
- [ ] APIKey có `json:"-"` tag
- [ ] YAML tags (`yaml:"..."`) trên TẤT CẢ fields
- [ ] JSON tags (`json:"..."`) trên TẤT CẢ fields
- [ ] `omitempty` trên optional fields
- [ ] KHÔNG có global variable
- [ ] KHÔNG import `log/slog`
- [ ] Godoc comments giải thích WHY cho mỗi design decision
- [ ] `go build ./kernel/config/...` không lỗi
