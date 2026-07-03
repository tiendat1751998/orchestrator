# Micro-Task 6.04: Create cmd/orchestrator/cmd/agents.go

## Info
- **File**: `cmd/orchestrator/cmd/agents.go`
- **Package**: `cmd`
- **Depends on**: 6.01, 2.20 (plugin registry)
- **Time**: 15 min
- **Verify**: `go build ./cmd/orchestrator/...`

## Purpose
Implements `orchestrator agents list` and `orchestrator agents info <name>` to display registered agent plugins, their capabilities, and health status.

## EXACT code to create

```go
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tiendat1751998/orchestrator/kernel/config"
	"github.com/tiendat1751998/orchestrator/kernel/registry"
)

// NewAgentsCmd creates the `agents` subcommand group.
func NewAgentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Manage registered agents",
	}

	cmd.AddCommand(newAgentsListCmd())
	cmd.AddCommand(newAgentsInfoCmd())

	return cmd
}

func newAgentsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered agents",
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

			agents := reg.ListAgents()
			if len(agents) == 0 {
				fmt.Println("No agents registered.")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "Name\tRole\tProvider\tCapabilities\tStatus")
			fmt.Fprintln(tw, "----\t----\t--------\t------------\t------")
			for _, a := range agents {
				caps := formatCapabilities(a.Capabilities())
				status := "stopped"
				if a.IsAvailable(ctx) {
					status = "ready"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					a.Name(), a.Role(), a.ProviderName(), caps, status)
			}
			tw.Flush()
			return nil
		},
	}
}

func newAgentsInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info [name]",
		Short: "Show details of a specific agent",
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

			a, err := reg.GetAgent(args[0])
			if err != nil {
				return fmt.Errorf("agent %q not found: %w", args[0], err)
			}

			fmt.Printf("\n🤖 Agent: %s\n", a.Name())
			fmt.Printf("   Role:         %s\n", a.Role())
			fmt.Printf("   Capabilities: %s\n", formatCapabilities(a.Capabilities()))
			fmt.Println()
			return nil
		},
	}
}

func formatCapabilities(caps []agent.Capability) string {
	if len(caps) == 0 {
		return "none"
	}
	var names []string
	for _, c := range caps {
		names = append(names, string(c))
	}
	return strings.Join(names, ", ")
}
```

## Rules
1. **Subcommand Groups**: Use Cobra subcommand groups (`agents list`, `agents info`) for logical namespace isolation.
2. **Tabular Formatting**: Always use `tabwriter` for aligned output.

## Pitfalls

### Pitfall 1: Initializing full kernel for read-only queries
For listing agents/providers, avoid bootstrapping the entire orchestrator. Only initialize the registry and config subsystems.

## Verify
```bash
go build ./cmd/orchestrator/...
```

## Checklist
- [ ] File `cmd/orchestrator/cmd/agents.go` exists
- [ ] `agents list` displays tabular agent info
- [ ] `agents info <name>` shows agent details
- [ ] `go build ./cmd/orchestrator/...` passes
