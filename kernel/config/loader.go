package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// configSearchPaths defines where to look for config files, in order.
// The first file found wins.
var configSearchPaths = []string{
	".orchestrator/settings.yaml",
	".orchestrator/settings.yml",
}

// Load loads configuration from the given file path.
//
// Processing pipeline:
//  1. Read YAML file bytes
//  2. Parse into raw map[string]any
//  3. Resolve ${ENV_VAR} placeholders
//  4. Re-serialize to YAML bytes (with resolved values)
//  5. Unmarshal into typed Config struct
//  6. Parse duration strings ("120s" → time.Duration)
//  7. Merge with defaults (fill zero-value fields)
//
// WHY this 7-step pipeline?
// → Step 2-4: Env var resolution must happen on raw strings BEFORE
//
//	type conversion, because "${VAR}" is a string but the target
//	field might be int, bool, or time.Duration.
//
// → Step 5: yaml.Unmarshal into typed struct handles type conversion.
// → Step 6: time.Duration custom parsing (yaml.v3 doesn't handle "120s").
// → Step 7: Fill remaining zero values with defaults.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file %q: %w", path, err)
	}

	return parseConfig(data)
}

// LoadFromSearchPaths finds and loads the config file from standard locations.
//
// Search order:
//  1. .orchestrator/settings.yaml (project root)
//  2. .orchestrator/settings.yml (project root)
//
// If no config file is found, returns DefaultConfig() (NOT an error).
// A project without config file should still work with defaults.
func LoadFromSearchPaths() (*Config, error) {
	for _, searchPath := range configSearchPaths {
		absPath, err := filepath.Abs(searchPath)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return Load(absPath)
		}
	}

	// No config file found → use defaults
	return DefaultConfig(), nil
}

// parseConfig processes raw YAML bytes into a Config.
func parseConfig(data []byte) (*Config, error) {
	// Step 1: Parse into raw map for env var resolution
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("config: invalid YAML: %w", err)
	}

	// Step 2: Resolve ${ENV_VAR} placeholders in all string values
	if err := ResolveEnvInMap(raw); err != nil {
		return nil, fmt.Errorf("config: env var resolution: %w", err)
	}

	// Step 2.5: Coerce resolved string values to appropriate types (int, bool, float)
	// so they marshal back to YAML as typed literals instead of quoted strings.
	raw = coerceTypes(raw).(map[string]any)

	// Step 3: Re-serialize resolved map back to YAML
	resolvedData, err := yaml.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("config: re-serialize: %w", err)
	}

	// Step 4: Unmarshal into typed Config struct
	// We use an intermediate struct with string durations for Step 5.
	var intermediate intermediateConfig
	if err := yaml.Unmarshal(resolvedData, &intermediate); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}

	// Step 5: Convert intermediate → final Config (parse durations)
	cfg, err := intermediate.toConfig()
	if err != nil {
		return nil, err
	}

	// Step 6: Merge with defaults
	MergeWithDefaults(cfg)

	return cfg, nil
}

// coerceString tries to convert a string to bool, int64, or float64.
func coerceString(s string) any {
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

// coerceTypes recursively walks a map/slice and coerces string values.
func coerceTypes(val any) any {
	switch v := val.(type) {
	case string:
		return coerceString(v)
	case map[string]any:
		for k, item := range v {
			v[k] = coerceTypes(item)
		}
		return v
	case []any:
		for i, item := range v {
			v[i] = coerceTypes(item)
		}
		return v
	default:
		return val
	}
}

// intermediateConfig mirrors Config but uses string for duration fields.
//
// WHY?
// yaml.v3 CANNOT unmarshal "120s" into time.Duration directly.
// It tries to parse it as an integer (nanoseconds) and fails.
//
// Solution: unmarshal into string first, then time.ParseDuration().
type intermediateConfig struct {
	Orchestrator intermediateOrchestratorConfig `yaml:"orchestrator"`
	Providers    intermediateProvidersConfig    `yaml:"providers"`
	Agents       map[string]AgentConfig         `yaml:"agents,omitempty"`
	Security     SecurityConfig                 `yaml:"security"`
}

type intermediateOrchestratorConfig struct {
	Name               string `yaml:"name"`
	LogLevel           string `yaml:"log_level"`
	LogFormat          string `yaml:"log_format"`
	DataDir            string `yaml:"data_dir"`
	MaxConcurrentTasks int    `yaml:"max_concurrent_tasks"`
	ShutdownTimeout    string `yaml:"shutdown_timeout"` // "30s", "5m" → parsed below
}

type intermediateProvidersConfig struct {
	Default string                          `yaml:"default"`
	Configs map[string]intermediateProvider `yaml:"configs"`
}

type intermediateProvider struct {
	Type     string            `yaml:"type"`
	Model    string            `yaml:"model"`
	Binary   string            `yaml:"binary,omitempty"`
	BaseURL  string            `yaml:"base_url,omitempty"`
	APIKey   string            `yaml:"api_key,omitempty"`
	Timeout  string            `yaml:"timeout"` // "120s" → parsed below
	MaxRetry int               `yaml:"max_retry"`
	Extra    map[string]string `yaml:"extra,omitempty"`
}

// toConfig converts intermediateConfig → Config, parsing duration strings.
func (ic *intermediateConfig) toConfig() (*Config, error) {
	cfg := &Config{
		Orchestrator: OrchestratorConfig{
			Name:               ic.Orchestrator.Name,
			LogLevel:           ic.Orchestrator.LogLevel,
			LogFormat:          ic.Orchestrator.LogFormat,
			DataDir:            ic.Orchestrator.DataDir,
			MaxConcurrentTasks: ic.Orchestrator.MaxConcurrentTasks,
		},
		Providers: ProvidersConfig{
			Default: ic.Providers.Default,
			Configs: make(map[string]ProviderEntry),
		},
		Agents:   ic.Agents,
		Security: ic.Security,
	}

	// Parse orchestrator shutdown timeout
	if ic.Orchestrator.ShutdownTimeout != "" {
		d, err := time.ParseDuration(ic.Orchestrator.ShutdownTimeout)
		if err != nil {
			return nil, fmt.Errorf("config: orchestrator.shutdown_timeout %q: %w",
				ic.Orchestrator.ShutdownTimeout, err)
		}
		cfg.Orchestrator.ShutdownTimeout = d
	}

	// Parse provider timeouts
	for name, ip := range ic.Providers.Configs {
		entry := ProviderEntry{
			Type:     ip.Type,
			Model:    ip.Model,
			Binary:   ip.Binary,
			BaseURL:  ip.BaseURL,
			APIKey:   ip.APIKey,
			MaxRetry: ip.MaxRetry,
			Extra:    ip.Extra,
		}
		if ip.Timeout != "" {
			d, err := time.ParseDuration(ip.Timeout)
			if err != nil {
				return nil, fmt.Errorf("config: providers.%s.timeout %q: %w",
					name, ip.Timeout, err)
			}
			entry.Timeout = d
		}
		cfg.Providers.Configs[name] = entry
	}

	return cfg, nil
}

// ParseBytes parses config from raw YAML bytes (for testing).
func ParseBytes(data []byte) (*Config, error) {
	return parseConfig(data)
}
