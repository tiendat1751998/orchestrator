# Micro-Task 2.05: Tạo kernel/config/validator.go

## Thông tin
- **File tạo**: `kernel/config/validator.go`
- **Package**: `config`
- **Dependencies trước**: 2.01 (config.go)
- **Thời gian**: 20 phút
- **Verify**: `go build ./kernel/config/...`

## Mục đích
Validate config SAU khi load. Trả về TẤT CẢ errors cùng lúc.
User fix 1 lần, không phải chạy đi chạy lại từng lỗi.

## Nội dung CHÍNH XÁC cần tạo

```go
package config

import (
	"fmt"
	"strings"
)

// validLogLevels are the accepted log level values.
var validLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

// validLogFormats are the accepted log format values.
var validLogFormats = map[string]bool{
	"json": true,
	"text": true,
}

// validProviderTypes are the accepted provider type values.
var validProviderTypes = map[string]bool{
	"cli":   true,
	"api":   true,
	"local": true,
}

// ValidationErrors collects multiple validation errors.
//
// WHY custom error type instead of returning first error?
// → User runs "orchestrator validate" and gets:
//     ❌ orchestrator.log_level: "verbose" is not valid (use: debug, info, warn, error)
//     ❌ providers.default: "openai" not found in configured providers
//     ❌ providers.configs.antigravity.model: required field is empty
// → User fixes ALL 3 issues in 1 edit → runs again → success
// → If we returned only the first error → user fixes 1, runs again, sees next → frustrating
type ValidationErrors struct {
	Errors []ValidationError
}

// ValidationError is a single validation failure with context.
type ValidationError struct {
	// Field is the dot-path to the config field (e.g., "providers.default")
	Field string

	// Message describes the problem.
	Message string
}

// Error implements the error interface.
// Returns all errors as a formatted multi-line string.
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "config: no validation errors"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("config: %d validation error(s):\n", len(ve.Errors)))
	for i, e := range ve.Errors {
		b.WriteString(fmt.Sprintf("  %d. %s: %s\n", i+1, e.Field, e.Message))
	}
	return b.String()
}

// HasErrors returns true if there are validation errors.
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// Add appends a validation error.
func (ve *ValidationErrors) Add(field, message string) {
	ve.Errors = append(ve.Errors, ValidationError{Field: field, Message: message})
}

// Addf appends a formatted validation error.
func (ve *ValidationErrors) Addf(field, format string, args ...any) {
	ve.Errors = append(ve.Errors, ValidationError{
		Field:   field,
		Message: fmt.Sprintf(format, args...),
	})
}

// Validate checks the entire Config for errors.
//
// Returns nil if config is valid.
// Returns *ValidationErrors (with all errors) if invalid.
//
// Order: orchestrator → providers → agents → security
// This matches the YAML file structure for easy debugging.
func Validate(cfg *Config) error {
	errs := &ValidationErrors{}

	validateOrchestrator(cfg, errs)
	validateProviders(cfg, errs)
	validateAgents(cfg, errs)
	validateSecurity(cfg, errs)

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// validateOrchestrator checks orchestrator-level settings.
func validateOrchestrator(cfg *Config, errs *ValidationErrors) {
	o := cfg.Orchestrator

	// Name: required, non-empty
	if strings.TrimSpace(o.Name) == "" {
		errs.Add("orchestrator.name", "required field is empty")
	}

	// LogLevel: must be one of valid values
	if !validLogLevels[o.LogLevel] {
		errs.Addf("orchestrator.log_level",
			"%q is not valid (use: debug, info, warn, error)", o.LogLevel)
	}

	// LogFormat: must be one of valid values
	if !validLogFormats[o.LogFormat] {
		errs.Addf("orchestrator.log_format",
			"%q is not valid (use: json, text)", o.LogFormat)
	}

	// MaxConcurrentTasks: must be positive
	if o.MaxConcurrentTasks < 1 {
		errs.Addf("orchestrator.max_concurrent_tasks",
			"must be >= 1, got %d", o.MaxConcurrentTasks)
	}
	if o.MaxConcurrentTasks > 50 {
		errs.Addf("orchestrator.max_concurrent_tasks",
			"must be <= 50, got %d (high values may trigger provider rate limits)",
			o.MaxConcurrentTasks)
	}

	// ShutdownTimeout: must be positive
	if o.ShutdownTimeout < 0 {
		errs.Addf("orchestrator.shutdown_timeout",
			"must be positive, got %s", o.ShutdownTimeout)
	}
}

// validateProviders checks provider configurations.
func validateProviders(cfg *Config, errs *ValidationErrors) {
	p := cfg.Providers

	// Default provider must be specified
	if strings.TrimSpace(p.Default) == "" {
		errs.Add("providers.default", "required field is empty")
	}

	// Default provider must exist in configs
	if p.Default != "" && len(p.Configs) > 0 {
		if _, exists := p.Configs[p.Default]; !exists {
			configured := make([]string, 0, len(p.Configs))
			for name := range p.Configs {
				configured = append(configured, name)
			}
			errs.Addf("providers.default",
				"provider %q not found in configured providers: [%s]",
				p.Default, strings.Join(configured, ", "))
		}
	}

	// Validate each provider entry
	for name, entry := range p.Configs {
		prefix := fmt.Sprintf("providers.configs.%s", name)

		// Type: required and valid
		if !validProviderTypes[entry.Type] {
			errs.Addf(prefix+".type",
				"%q is not valid (use: cli, api, local)", entry.Type)
		}

		// Model: required
		if strings.TrimSpace(entry.Model) == "" {
			errs.Add(prefix+".model", "required field is empty")
		}

		// Binary: required for cli type
		if entry.Type == "cli" && strings.TrimSpace(entry.Binary) == "" {
			errs.Add(prefix+".binary",
				"required when type is \"cli\" (path to the CLI executable)")
		}

		// BaseURL: required for api type
		if entry.Type == "api" && strings.TrimSpace(entry.BaseURL) == "" {
			errs.Add(prefix+".base_url",
				"required when type is \"api\" (e.g., \"https://generativelanguage.googleapis.com\")")
		}

		// APIKey: required for api type
		if entry.Type == "api" && strings.TrimSpace(entry.APIKey) == "" {
			errs.Add(prefix+".api_key",
				"required when type is \"api\" (use: ${YOUR_API_KEY_ENV_VAR})")
		}

		// Timeout: must be positive if set
		if entry.Timeout < 0 {
			errs.Addf(prefix+".timeout",
				"must be positive, got %s", entry.Timeout)
		}

		// MaxRetry: must be non-negative
		if entry.MaxRetry < 0 {
			errs.Addf(prefix+".max_retry",
				"must be >= 0, got %d", entry.MaxRetry)
		}
	}
}

// validateAgents checks agent configurations.
func validateAgents(cfg *Config, errs *ValidationErrors) {
	for name, ac := range cfg.Agents {
		prefix := fmt.Sprintf("agents.%s", name)

		// If provider specified, must exist in providers config
		if ac.Provider != "" {
			if _, exists := cfg.Providers.Configs[ac.Provider]; !exists {
				errs.Addf(prefix+".provider",
					"provider %q not found in configured providers", ac.Provider)
			}
		}

		// Temperature: must be 0.0-2.0 if set
		if ac.Temperature != nil {
			if *ac.Temperature < 0.0 || *ac.Temperature > 2.0 {
				errs.Addf(prefix+".temperature",
					"must be between 0.0 and 2.0, got %f", *ac.Temperature)
			}
		}

		// MaxTokens: must be positive if set
		if ac.MaxTokens != nil && *ac.MaxTokens < 1 {
			errs.Addf(prefix+".max_tokens",
				"must be >= 1, got %d", *ac.MaxTokens)
		}
	}
}

// validateSecurity checks security settings.
func validateSecurity(cfg *Config, errs *ValidationErrors) {
	s := cfg.Security

	// MaxFileSize: must be positive
	if s.MaxFileSize < 1 {
		errs.Addf("security.max_file_size",
			"must be >= 1 byte, got %d", s.MaxFileSize)
	}

	// MaxOutputSize: must be positive
	if s.MaxOutputSize < 1 {
		errs.Addf("security.max_output_size",
			"must be >= 1 byte, got %d", s.MaxOutputSize)
	}
}
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Multi-error pattern
```go
// ❌ SAI — dừng ở lỗi đầu tiên:
if cfg.Name == "" {
    return errors.New("name required")
}

// ✅ ĐÚNG — thu thập TẤT CẢ lỗi:
errs := &ValidationErrors{}
errs.Add("orchestrator.name", "required")
errs.Add("providers.default", "required")
if errs.HasErrors() {
    return errs
}
```

### Pitfall 2: Cross-field validation
Provider trong agent config PHẢI tồn tại trong providers config.
```yaml
providers:
  configs:
    antigravity: { ... }
agents:
  backend:
    provider: "openai"  # ← LỖI: "openai" không có trong providers
```

### Pitfall 3: Pointer field validation
```go
if ac.Temperature != nil {
    if *ac.Temperature < 0.0 { ... }
}
```
PHẢI check nil TRƯỚC khi dereference. Nil = "not set" = skip validation.

### Pitfall 4: Validate returns nil, NOT empty ValidationErrors
```go
// ❌ SAI:
return errs // errs.Errors is empty slice → non-nil interface

// ✅ ĐÚNG:
if errs.HasErrors() {
    return errs
}
return nil
```
`return errs` khi errors empty → `err != nil` vẫn true (non-nil pointer trong interface).

## Checklist
- [ ] File `kernel/config/validator.go` tồn tại
- [ ] Package: `package config`
- [ ] `ValidationErrors` struct với `Error()`, `HasErrors()`, `Add()`, `Addf()`
- [ ] `Validate()` function trả về nil hoặc `*ValidationErrors`
- [ ] `validateOrchestrator()` — check Name, LogLevel, LogFormat, MaxConcurrentTasks
- [ ] `validateProviders()` — check Default exists, each entry valid, type-specific requirements
- [ ] `validateAgents()` — check provider reference, Temperature range, MaxTokens positive
- [ ] `validateSecurity()` — check MaxFileSize, MaxOutputSize
- [ ] Cross-field validation (agent.provider → providers.configs)
- [ ] Error messages chỉ rõ field path + giá trị sai + gợi ý giá trị đúng
- [ ] Trả về nil khi no errors (KHÔNG return empty ValidationErrors)
- [ ] Godoc comments
- [ ] `go build ./kernel/config/...` không lỗi
