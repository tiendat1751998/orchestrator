# Micro-Task 2.02: Tạo kernel/config/defaults.go

## Thông tin
- **File tạo**: `kernel/config/defaults.go`
- **Package**: `config`
- **Dependencies trước**: 2.01 (config.go)
- **Thời gian**: 10 phút
- **Verify**: `go build ./kernel/config/...`

## Mục đích
Cung cấp giá trị mặc định cho TẤT CẢ config fields.
Khi user không set 1 field, hệ thống dùng default thay vì zero value.

## Tại sao tách file riêng?
1. Config struct (config.go) = schema definition = ít thay đổi
2. Defaults (defaults.go) = giá trị cụ thể = thay đổi thường xuyên
3. Tách ra → dễ review defaults mà không sửa nhầm struct

## Nội dung CHÍNH XÁC cần tạo

```go
package config

import "time"

// DefaultConfig returns a Config with all default values set.
//
// This is the baseline configuration. YAML file values override these.
// Environment variables override YAML values.
//
// Priority: Environment > YAML > Defaults
//
// Usage:
//
//	cfg := DefaultConfig()
//	// Then merge YAML on top
//	// Then merge env vars on top
func DefaultConfig() *Config {
	return &Config{
		Orchestrator: OrchestratorConfig{
			Name:               "orchestrator",
			LogLevel:           "info",
			LogFormat:          "text",
			DataDir:            ".orchestrator/data",
			MaxConcurrentTasks: 5,
			ShutdownTimeout:    30 * time.Second,
		},
		Providers: ProvidersConfig{
			Default: "antigravity",
			Configs: map[string]ProviderEntry{},
		},
		Agents: map[string]AgentConfig{},
		Security: SecurityConfig{
			Sandbox: true,
			BlockedCommands: []string{
				"rm -rf /",
				"rm -rf ~",
				"rm -rf .",
				"sudo",
				"chmod 777",
				"mkfs",
				"dd if=",
				":(){ :|:& };:", // Fork bomb
				"format c:",     // Windows format
				"> /dev/sda",    // Direct disk write
			},
			BlockedPaths: []string{
				".env",
				".env.local",
				".git/config",
			},
			MaxFileSize:   1 * 1024 * 1024,  // 1 MB
			MaxOutputSize: 100 * 1024,        // 100 KB
			AuditLog:      true,
		},
	}
}

// DefaultProviderEntry returns defaults for a provider entry.
func DefaultProviderEntry() ProviderEntry {
	return ProviderEntry{
		Type:     "cli",
		Timeout:  120 * time.Second,
		MaxRetry: 3,
	}
}

// MergeWithDefaults fills in zero-value fields with defaults.
//
// IMPORTANT: This is NOT a deep merge. It only fills TOP-LEVEL zero values.
// Nested fields must be handled explicitly.
//
// Why needed?
// → yaml.Unmarshal only sets fields that exist in the YAML file.
// → Fields NOT in YAML remain as Go zero values (0, "", nil).
// → We want those to be meaningful defaults, not zeros.
//
// Example:
//   YAML has: { orchestrator: { name: "my-app" } }
//   After unmarshal: cfg.Orchestrator.Name = "my-app", cfg.Orchestrator.LogLevel = ""
//   After merge:     cfg.Orchestrator.LogLevel = "info" (from default)
func MergeWithDefaults(cfg *Config) {
	defaults := DefaultConfig()

	// Orchestrator defaults
	if cfg.Orchestrator.Name == "" {
		cfg.Orchestrator.Name = defaults.Orchestrator.Name
	}
	if cfg.Orchestrator.LogLevel == "" {
		cfg.Orchestrator.LogLevel = defaults.Orchestrator.LogLevel
	}
	if cfg.Orchestrator.LogFormat == "" {
		cfg.Orchestrator.LogFormat = defaults.Orchestrator.LogFormat
	}
	if cfg.Orchestrator.DataDir == "" {
		cfg.Orchestrator.DataDir = defaults.Orchestrator.DataDir
	}
	if cfg.Orchestrator.MaxConcurrentTasks == 0 {
		cfg.Orchestrator.MaxConcurrentTasks = defaults.Orchestrator.MaxConcurrentTasks
	}
	if cfg.Orchestrator.ShutdownTimeout == 0 {
		cfg.Orchestrator.ShutdownTimeout = defaults.Orchestrator.ShutdownTimeout
	}

	// Providers defaults
	if cfg.Providers.Default == "" {
		cfg.Providers.Default = defaults.Providers.Default
	}
	if cfg.Providers.Configs == nil {
		cfg.Providers.Configs = make(map[string]ProviderEntry)
	}

	// Apply default values to each provider entry
	for name, entry := range cfg.Providers.Configs {
		if entry.Timeout == 0 {
			entry.Timeout = DefaultProviderEntry().Timeout
		}
		if entry.MaxRetry == 0 {
			entry.MaxRetry = DefaultProviderEntry().MaxRetry
		}
		cfg.Providers.Configs[name] = entry // write back (map value is a copy)
	}

	// Agents defaults
	if cfg.Agents == nil {
		cfg.Agents = make(map[string]AgentConfig)
	}

	// Security defaults
	if cfg.Security.MaxFileSize == 0 {
		cfg.Security.MaxFileSize = defaults.Security.MaxFileSize
	}
	if cfg.Security.MaxOutputSize == 0 {
		cfg.Security.MaxOutputSize = defaults.Security.MaxOutputSize
	}
	// Sandbox and AuditLog: bool zero value is false.
	// But our defaults are true. Problem: user sets "sandbox: false" → merge overwrites with true.
	// Solution: DON'T merge booleans here. Handle in loader with explicit "was this field set?" logic.
	// For now: booleans keep their parsed value (false if not set, which means disabled).
	// The loader (2.04) will handle the "default true" behavior.

	// BlockedCommands: merge defaults if user list is empty
	if len(cfg.Security.BlockedCommands) == 0 {
		cfg.Security.BlockedCommands = defaults.Security.BlockedCommands
	}
	if len(cfg.Security.BlockedPaths) == 0 {
		cfg.Security.BlockedPaths = defaults.Security.BlockedPaths
	}
}
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Boolean default = true problem
Go zero value cho `bool` là `false`. Nếu user KHÔNG set `sandbox` trong YAML → `false`.
Nhưng default muốn là `true`. KHÔNG thể phân biệt "user chọn false" vs "user không set".
Giải pháp: 
- Option A: Dùng `*bool` (pointer) → nil = không set, &false = explicitly false
- Option B: Config loader xử lý (ưu tiên: đơn giản hơn cho phase đầu)
- **Chọn Option B** trong micro-task này — loader (2.04) sẽ set defaults trước khi unmarshal.

### Pitfall 2: Map value copy
```go
for name, entry := range cfg.Providers.Configs {
    entry.Timeout = 120 * time.Second
    // entry là COPY, không phải reference
    cfg.Providers.Configs[name] = entry // PHẢI write back
}
```
Nếu quên `cfg.Providers.Configs[name] = entry` → changes bị mất.

### Pitfall 3: BlockedCommands security
Default BlockedCommands bao gồm cả Unix VÀ Windows commands.
Fork bomb `:(){ :|:& };:` — classic Linux denial-of-service.
`format c:` — Windows disk format.
KHÔNG CHO PHÉP user override BlockedCommands bằng empty list (security risk).
Loader nên MERGE user list VỚI defaults, KHÔNG REPLACE.

## Checklist
- [ ] File `kernel/config/defaults.go` tồn tại
- [ ] Package: `package config`
- [ ] `DefaultConfig()` trả về `*Config` với tất cả defaults
- [ ] `DefaultProviderEntry()` trả về defaults cho provider
- [ ] `MergeWithDefaults()` fill zero-value fields
- [ ] DefaultConfig có đủ: Name, LogLevel, LogFormat, DataDir, MaxConcurrentTasks, ShutdownTimeout
- [ ] SecurityConfig defaults: Sandbox=true, 10+ BlockedCommands, MaxFileSize=1MB, MaxOutputSize=100KB
- [ ] Map iteration write-back pattern cho provider entries
- [ ] Boolean merge problem documented (NOT solved here, solved in loader)
- [ ] Godoc comments
- [ ] `go build ./kernel/config/...` không lỗi
