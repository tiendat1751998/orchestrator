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
