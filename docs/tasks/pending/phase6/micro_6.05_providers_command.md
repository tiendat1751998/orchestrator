# Micro-Task 6.05: Create cmd/orchestrator/cmd/providers.go

## Info
- **File**: `cmd/orchestrator/cmd/providers.go`
- **Package**: `cmd`
- **Depends on**: 6.01, 2.20 (plugin registry)
- **Time**: 15 min
- **Verify**: `go build ./cmd/orchestrator/...`

## Purpose
Implements `orchestrator providers list` and `orchestrator providers test <name>` for listing registered AI providers and testing their connectivity.

## EXACT code to create

```go
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/tiendat1751998/orchestrator/kernel/config"
	"github.com/tiendat1751998/orchestrator/kernel/registry"
)

// NewProvidersCmd creates the `providers` subcommand group.
func NewProvidersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "providers",
		Short: "Manage registered AI providers",
	}

	cmd.AddCommand(newProvidersListCmd())
	cmd.AddCommand(newProvidersTestCmd())
	return cmd
}

func newProvidersListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered providers",
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			cfg, err := config.Load(ConfigPathFrom(ctx))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			reg, err := registry.NewPluginRegistry(cfg, slog.Default())
			if err != nil {
				return fmt.Errorf("failed to initialize registry: %w", err)
			}

			providers := reg.ListProviders()
			if len(providers) == 0 {
				fmt.Println("No providers registered.")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "Name\tType\tModel\tStatus")
			fmt.Fprintln(tw, "----\t----\t-----\t------")
			for _, p := range providers {
				status := "unavailable"
				if p.IsAvailable(ctx) {
					status = "available"
				}
				models, _ := p.Models(ctx)
				modelStr := "none"
				if len(models) > 0 {
					modelStr = models[0]
					if len(models) > 1 {
						modelStr += fmt.Sprintf(" (+%d)", len(models)-1)
					}
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", p.Name(), "cli", modelStr, status)
			}
			tw.Flush()
			return nil
		},
	}
}

func newProvidersTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test [name]",
		Short: "Test provider connectivity",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			cfg, err := config.Load(ConfigPathFrom(ctx))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			reg, err := registry.NewPluginRegistry(cfg, slog.Default())
			if err != nil {
				return fmt.Errorf("failed to initialize registry: %w", err)
			}

			p, err := reg.GetProvider(args[0])
			if err != nil {
				return fmt.Errorf("provider %q not found: %w", args[0], err)
			}

			fmt.Printf("🔍 Testing provider %q...\n", args[0])

			testCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			if p.IsAvailable(testCtx) {
				fmt.Printf("✅ Provider %q is available\n", args[0])
				models, _ := p.Models(testCtx)
				if len(models) > 0 {
					fmt.Printf("   Models: %v\n", models)
				}
			} else {
				fmt.Printf("❌ Provider %q is not available\n", args[0])
			}
			return nil
		},
	}
}
```

## Rules
1. **Connectivity Test Timeout**: Always use a bounded context (30s) for provider test calls. Never block indefinitely.
2. **Model Summary**: Show first model + count of others, not the full list (which can be very long for API providers).

## Verify
```bash
go build ./cmd/orchestrator/...
```

## Checklist
- [ ] File `cmd/orchestrator/cmd/providers.go` exists
- [ ] `providers list` shows tabular provider info
- [ ] `providers test <name>` performs bounded connectivity check
- [ ] `go build ./cmd/orchestrator/...` passes
