# Micro-Task 2.02: Create kernel/config/defaults.go

## Info
- **File**: `kernel/config/defaults.go`
- **Package**: `config`
- **Depends on**: 2.01 (config.go)
- **Time**: 10 min
- **Verify**: `go build ./kernel/config/...`

## Purpose
Establishes baseline configurations (`DefaultConfig`, `DefaultProviderEntry`) and implements the default settings merger logic (`MergeWithDefaults`) to resolve empty settings properties with system fallbacks.

## EXACT code to create

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

	// BlockedCommands: merge defaults if user list is empty
	if len(cfg.Security.BlockedCommands) == 0 {
		cfg.Security.BlockedCommands = defaults.Security.BlockedCommands
	}
	if len(cfg.Security.BlockedPaths) == 0 {
		cfg.Security.BlockedPaths = defaults.Security.BlockedPaths
	}
}
```

## Rules
1. **Separation of Structure and Configuration**: Struct formats are declared in `config.go`, while baseline default configurations live in `defaults.go` to ease modifications.
2. **Boolean Defaults Problem**: In Go, boolean fields default to `false`. If a user does not configure `Sandbox` or `AuditLog` in their YAML, they read as `false` rather than their default target state (`true`). This is addressed in the custom loader logic in task 2.04.
3. **Map Copy Write-Back**: When updating properties inside map loops (such as iterating `Configs`), updates must be explicitly written back to the map (e.g. `cfg.Providers.Configs[name] = entry`).

## ⚠️ Pitfalls

### Pitfall 1: Modifying loop copies directly without writing back to maps
```go
for name, entry := range cfg.Providers.Configs {
    entry.Timeout = 120 * time.Second
    cfg.Providers.Configs[name] = entry // Explicit write back to target map.
```
Go map range iterators return value copies rather than address pointers. Always write back changes.

### Pitfall 2: Permitting command sandboxing bypasses via empty command lists
If users override `BlockedCommands` with empty lists, safety policies are deactivated. The configuration loader should merge custom rules with defaults rather than replacing security rules entirely.

## Verify
```bash
go build ./kernel/config/...
```

## Checklist
- [ ] File `kernel/config/defaults.go` exists
- [ ] Package: `config`
- [ ] `DefaultConfig()` returns configured `*Config` containing base settings defaults
- [ ] `MergeWithDefaults()` replaces empty settings with default fallbacks
- [ ] `DefaultProviderEntry` defines default timeout and retry values
- [ ] Iteration of configs uses correct map write-back updates
- [ ] `go build ./kernel/config/...` passes
