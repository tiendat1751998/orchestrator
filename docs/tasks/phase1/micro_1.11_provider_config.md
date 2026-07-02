# Micro-Task 1.11: Tạo contracts/provider/config.go

## Thông tin
- **File tạo**: `contracts/provider/config.go`
- **Package**: `provider`
- **Dependencies trước**: 1.06
- **Thời gian**: 10 phút

## Nội dung CHÍNH XÁC cần tạo

```go
package provider

import "time"

// Config holds the configuration for a provider instance.
//
// Example YAML:
//
//	providers:
//	  antigravity:
//	    type: cli
//	    binary: antigravity
//	    model: gemini-2.5-pro
//	    timeout: 120s
//	    max_retry: 3
type Config struct {
	// Name is the unique identifier for this provider.
	// Example: "antigravity", "gemini", "claude"
	Name string `yaml:"name" json:"name"`

	// Type indicates how to communicate with this provider.
	//   "cli" — Spawn a CLI process (e.g., Antigravity CLI)
	//   "api" — Call an HTTP/gRPC API (e.g., Gemini API, Claude API)
	//   "local" — Run a local model (e.g., Ollama, llama.cpp)
	Type string `yaml:"type" json:"type"`

	// Model is the default model to use.
	// Can be overridden per-request in Request.Model.
	Model string `yaml:"model" json:"model"`

	// BaseURL is the API endpoint URL.
	// Only used for Type="api".
	// Example: "https://generativelanguage.googleapis.com/v1beta"
	BaseURL string `yaml:"base_url,omitempty" json:"base_url,omitempty"`

	// APIKey is the authentication key for the provider.
	//
	// SECURITY: This value should NOT be hardcoded in YAML files.
	// Use environment variable reference: "${GEMINI_API_KEY}"
	// The config loader will resolve env vars at load time.
	APIKey string `yaml:"api_key,omitempty" json:"-"` // json:"-" prevents accidental JSON serialization

	// Binary is the path to the CLI executable.
	// Only used for Type="cli".
	// Example: "antigravity", "/usr/local/bin/antigravity"
	Binary string `yaml:"binary,omitempty" json:"binary,omitempty"`

	// Timeout is the maximum time to wait for a response.
	// Applies to both Send() and individual Stream() chunks.
	// Default: 120s
	Timeout time.Duration `yaml:"timeout" json:"timeout"`

	// MaxRetry is the maximum number of retry attempts on transient errors.
	// Set to 0 to disable retries.
	// Default: 3
	MaxRetry int `yaml:"max_retry" json:"max_retry"`

	// Extra contains provider-specific configuration that doesn't fit
	// into the standard fields above.
	//
	// WHY?
	// → Each provider may have unique settings (safety filters, API version, etc.)
	// → Rather than adding fields for every provider, use a generic map.
	// → Provider implementation reads from Extra what it needs.
	//
	// Example:
	//   extra:
	//     api_version: "v1beta"
	//     safety_level: "block_only_high"
	Extra map[string]string `yaml:"extra,omitempty" json:"extra,omitempty"`
}

// GetExtra returns the value for a key in Extra, or defaultValue if not found.
func (c *Config) GetExtra(key, defaultValue string) string {
	if v, ok := c.Extra[key]; ok {
		return v
	}
	return defaultValue
}

// TimeoutOrDefault returns the configured timeout, or the default if not set.
func (c *Config) TimeoutOrDefault() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return 120 * time.Second // Default: 2 minutes
}

// MaxRetryOrDefault returns the configured max retry, or the default if not set.
func (c *Config) MaxRetryOrDefault() int {
	if c.MaxRetry > 0 {
		return c.MaxRetry
	}
	return 3 // Default: 3 retries
}
```

## ⚠️ Pitfalls
1. **APIKey JSON tag `-`**: `json:"-"` ngăn APIKey bị leak khi serialize config thành JSON (ví dụ: trong API response hoặc debug log). YAML tag vẫn hoạt động bình thường để load config.
2. **time.Duration YAML**: Go `time.Duration` không tự parse từ YAML string. YAML library sẽ parse "120s" → cần custom unmarshaller hoặc dùng `yaml:"timeout"` với library hỗ trợ (như `gopkg.in/yaml.v3`).
3. **Env var resolution**: Config loader (Task 2.1) sẽ resolve `"${ENV_VAR}"` thành giá trị thực. File này chỉ định nghĩa struct, KHÔNG resolve env vars.

## Checklist
- [ ] Config struct với 9 fields
- [ ] YAML tags cho tất cả fields
- [ ] APIKey có `json:"-"` tag
- [ ] Extra map cho provider-specific config
- [ ] `GetExtra()`, `TimeoutOrDefault()`, `MaxRetryOrDefault()` helpers
- [ ] Godoc comments
- [ ] `go build ./contracts/...` không lỗi
