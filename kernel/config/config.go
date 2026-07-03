// Package config handles loading, parsing, and validating the orchestrator
// configuration from YAML files and environment variables.
package config

import "time"

// Config is the root configuration structure.
type Config struct {
	// Orchestrator contains general orchestrator settings.
	Orchestrator OrchestratorConfig `yaml:"orchestrator" json:"orchestrator"`

	// Providers contains AI provider configurations.
	Providers ProvidersConfig `yaml:"providers" json:"providers"`

	// Agents contains agent configurations.
	Agents map[string]AgentConfig `yaml:"agents,omitempty" json:"agents,omitempty"`

	// Security contains security policy settings.
	Security SecurityConfig `yaml:"security" json:"security"`
}

// OrchestratorConfig contains general orchestrator settings.
type OrchestratorConfig struct {
	// Name identifies this orchestrator instance.
	Name string `yaml:"name" json:"name"`

	// LogLevel controls log verbosity ("debug", "info", "warn", "error").
	LogLevel string `yaml:"log_level" json:"log_level"`

	// LogFormat controls log output format ("json", "text").
	LogFormat string `yaml:"log_format" json:"log_format"`

	// DataDir is the directory for orchestrator data files.
	DataDir string `yaml:"data_dir" json:"data_dir"`

	// MaxConcurrentTasks limits how many tasks run simultaneously.
	MaxConcurrentTasks int `yaml:"max_concurrent_tasks" json:"max_concurrent_tasks"`

	// ShutdownTimeout is the maximum time to wait for graceful shutdown.
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout"`
}

// ProvidersConfig contains all provider configurations.
type ProvidersConfig struct {
	// Default is the name of the default provider.
	Default string `yaml:"default" json:"default"`

	// Configs maps provider name → provider configuration.
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
	APIKey string `yaml:"api_key,omitempty" json:"-"`

	// Timeout for provider requests.
	Timeout time.Duration `yaml:"timeout" json:"timeout"`

	// MaxRetry is retry attempts on transient failures.
	MaxRetry int `yaml:"max_retry" json:"max_retry"`

	// Extra holds provider-specific settings.
	Extra map[string]string `yaml:"extra,omitempty" json:"extra,omitempty"`
}

// AgentConfig is the config for a single agent.
type AgentConfig struct {
	// Provider is the name of the provider this agent uses.
	Provider string `yaml:"provider,omitempty" json:"provider,omitempty"`

	// Model overrides the provider's default model.
	Model string `yaml:"model,omitempty" json:"model,omitempty"`

	// PromptFile is the path to the system prompt file.
	PromptFile string `yaml:"prompt_file,omitempty" json:"prompt_file,omitempty"`

	// Temperature overrides the default temperature.
	Temperature *float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`

	// MaxTokens overrides the default max tokens.
	MaxTokens *int `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty"`

	// Tools lists specific tools this agent can use.
	Tools []string `yaml:"tools,omitempty" json:"tools,omitempty"`
}

// SecurityConfig contains security policy settings.
type SecurityConfig struct {
	// Sandbox enables command sandboxing.
	Sandbox bool `yaml:"sandbox" json:"sandbox"`

	// AllowedPaths are filesystem paths agents can access.
	AllowedPaths []string `yaml:"allowed_paths,omitempty" json:"allowed_paths,omitempty"`

	// BlockedPaths are filesystem paths agents CANNOT access.
	BlockedPaths []string `yaml:"blocked_paths,omitempty" json:"blocked_paths,omitempty"`

	// BlockedCommands are shell commands that are always rejected.
	BlockedCommands []string `yaml:"blocked_commands,omitempty" json:"blocked_commands,omitempty"`

	// MaxFileSize is the maximum file size an agent can read/write (bytes).
	MaxFileSize int64 `yaml:"max_file_size" json:"max_file_size"`

	// MaxOutputSize is the maximum command output size (bytes).
	MaxOutputSize int64 `yaml:"max_output_size" json:"max_output_size"`

	// AuditLog enables audit logging of all agent actions.
	AuditLog bool `yaml:"audit_log" json:"audit_log"`
}
