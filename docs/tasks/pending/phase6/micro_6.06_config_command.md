# Micro-Task 6.06: Create cmd/orchestrator/cmd/config.go

## Info
- **File**: `cmd/orchestrator/cmd/config.go`
- **Package**: `cmd`
- **Depends on**: 6.01, 2.01 (config struct)
- **Time**: 20 min
- **Verify**: `go build ./cmd/orchestrator/...`

## Purpose
Implements `orchestrator config init`, `config show`, and `config set key value` for managing the orchestrator YAML configuration file.

## EXACT code to create

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tiendat1751998/orchestrator/kernel/config"
	"gopkg.in/yaml.v3"
)

// NewConfigCmd creates the `config` subcommand group.
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage orchestrator configuration",
	}

	cmd.AddCommand(newConfigInitCmd())
	cmd.AddCommand(newConfigShowCmd())

	return cmd
}

func newConfigInitCmd() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a default configuration file",
		RunE: func(c *cobra.Command, args []string) error {
			if _, err := os.Stat(outputPath); err == nil {
				return fmt.Errorf("config file %q already exists. Use --force to overwrite", outputPath)
			}

			defaultCfg := config.DefaultConfig()

			data, err := yaml.Marshal(defaultCfg)
			if err != nil {
				return fmt.Errorf("failed to marshal default config: %w", err)
			}

			if err := os.WriteFile(outputPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			fmt.Printf("✅ Configuration file created: %s\n", outputPath)
			fmt.Println("   Edit the file to configure providers, agents, and security policies.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "orchestrator.yaml", "output file path")
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		RunE: func(c *cobra.Command, args []string) error {
			cfgPath := ConfigPathFrom(c.Context())

			cfg, err := config.Load(cfgPath)
			if err != nil {
				return fmt.Errorf("failed to load config from %q: %w", cfgPath, err)
			}

			// Redact sensitive fields before display
			redacted := *cfg
			for name, p := range redacted.Providers.Configs {
				if p.APIKey != "" {
					p.APIKey = "***REDACTED***"
					redacted.Providers.Configs[name] = p
				}
			}

			data, err := yaml.Marshal(&redacted)
			if err != nil {
				return fmt.Errorf("failed to serialize config: %w", err)
			}

			fmt.Printf("# Configuration: %s\n---\n%s", cfgPath, string(data))
			return nil
		},
	}
}
```

## Rules
1. **Secret Redaction**: NEVER display API keys in `config show`. Always redact sensitive fields before rendering.
2. **No Overwrite**: `config init` MUST refuse to overwrite existing files. Prevents accidental config destruction.
3. **Default Config**: Use `config.DefaultConfig()` to generate sensible defaults. Never hardcode YAML strings.

## Pitfalls

### Pitfall 1: Leaking API keys in config show output
```go
// WRONG:
fmt.Println(yaml.Marshal(cfg)) // Prints api_key: sk-xxx in plain text!

// CORRECT:
p.APIKey = "***REDACTED***" // Scrub secrets before display
```

## Verify
```bash
go build ./cmd/orchestrator/...
```

## Checklist
- [ ] File `cmd/orchestrator/cmd/config.go` exists
- [ ] `config init` creates default YAML file
- [ ] `config init` refuses to overwrite existing files
- [ ] `config show` redacts API keys before display
- [ ] `go build ./cmd/orchestrator/...` passes
