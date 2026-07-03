# Micro-Task 1.11: Create contracts/provider/config.go

## Info
- **File**: `contracts/provider/config.go`
- **Package**: `provider`
- **Depends on**: 1.06
- **Time**: 10 min
- **Verify**: `go build ./contracts/...`

## Purpose
Declares the `Config` models used to store instance properties (e.g. endpoint URLs, timeouts, retry limits, and API keys) for specific providers.

## EXACT code to create

```go
package provider

import "time"

// Config holds the configuration for a provider instance.
type Config struct {
	Name     string            `yaml:"name" json:"name"`
	Type     string            `yaml:"type" json:"type"`
	Model    string            `yaml:"model" json:"model"`
	BaseURL  string            `yaml:"base_url,omitempty" json:"base_url,omitempty"`
	APIKey   string            `yaml:"api_key,omitempty" json:"-"`
	Binary   string            `yaml:"binary,omitempty" json:"binary,omitempty"`
	Timeout  time.Duration     `yaml:"timeout" json:"timeout"`
	MaxRetry int               `yaml:"max_retry" json:"max_retry"`
	Extra    map[string]string `yaml:"extra,omitempty" json:"extra,omitempty"`
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

## Pitfalls

### Pitfall 1: Leaking sensitive credentials in logs
If you omit the `json:"-"` tag on API Key properties, printing configs using structured logs will leak keys to console outputs. Always redact secrets.

### Pitfall 2: Overriding duration formats during YAML parsing
Go's `time.Duration` requires custom string parsers in some libraries. Keep unmarshallers robust and fall back to default durations if parsing fails.

## Verify
```bash
go build ./contracts/...
```

## Checklist
- [ ] File exists at `contracts/provider/config.go`
- [ ] Package name is `provider`
- [ ] Config structs define both JSON and YAML tags
- [ ] Secrets omit serialization tags to protect credentials
- [ ] Default values are returned if properties are empty
- [ ] Build command passes
