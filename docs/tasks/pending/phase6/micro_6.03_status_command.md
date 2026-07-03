# Micro-Task 6.03: Create cmd/orchestrator/cmd/status.go

## Info
- **File**: `cmd/orchestrator/cmd/status.go`
- **Package**: `cmd`
- **Depends on**: 6.01, 6.15 (mission manager)
- **Time**: 15 min
- **Verify**: `go build ./cmd/orchestrator/...`

## Purpose
Implements `orchestrator status [mission-id]` to display the current or historical mission execution status with task DAG and completion indicators.

## EXACT code to create

```go
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tiendat1751998/orchestrator/kernel/config"
	"github.com/tiendat1751998/orchestrator/modules/mission"
)

// NewStatusCmd creates the `status` subcommand.
func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [mission-id]",
		Short: "Show mission execution status",
		Long: `Display the current or historical mission execution status.
Without arguments, shows the most recent mission.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			cfgPath := ConfigPathFrom(ctx)

			cfg, err := config.Load(cfgPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			store, err := mission.NewSQLiteStore(cfg.Orchestrator.DataDir)
			if err != nil {
				return fmt.Errorf("failed to open mission store: %w", err)
			}
			defer store.Close()

			mgr := mission.NewManager(store)

			if len(args) > 0 {
				return showMissionStatus(mgr, args[0])
			}
			return showLatestMission(mgr)
		},
	}
}

func showMissionStatus(mgr *mission.Manager, id string) error {
	m, err := mgr.Get(id)
	if err != nil {
		return fmt.Errorf("mission %q not found: %w", id, err)
	}

	printMissionDetails(m)
	return nil
}

func showLatestMission(mgr *mission.Manager) error {
	missions, err := mgr.List()
	if err != nil {
		return fmt.Errorf("failed to list missions: %w", err)
	}

	if len(missions) == 0 {
		fmt.Println("No missions found. Run `orchestrator mission \"...\"` to start one.")
		return nil
	}

	printMissionDetails(missions[0])
	return nil
}

func printMissionDetails(m *mission.MissionRecord) {
	fmt.Printf("\n🎯 Mission: %s\n", m.Title)
	fmt.Printf("   ID:      %s\n", m.ID)
	fmt.Printf("   Status:  %s\n", m.Status)
	fmt.Printf("   Created: %s\n", m.CreatedAt.Format("2006-01-02 15:04:05"))

	if len(m.Tasks) > 0 {
		fmt.Printf("\n📋 Tasks (%d):\n", len(m.Tasks))
		tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  #\tName\tAgent\tStatus\tDuration")
		fmt.Fprintln(tw, "  -\t----\t-----\t------\t--------")
		for i, t := range m.Tasks {
			icon := statusIcon(t.Status)
			fmt.Fprintf(tw, "  %d\t%s %s\t%s\t%s\t%s\n",
				i+1, icon, t.Name, t.Agent, t.Status, t.Duration)
		}
		tw.Flush()
	}

	if m.TotalTokens > 0 {
		fmt.Printf("\n💰 Tokens: %d | ⏱️ Duration: %s\n", m.TotalTokens, m.Duration)
	}
	fmt.Println()
}

func statusIcon(status string) string {
	switch status {
	case "completed":
		return "✅"
	case "running":
		return "🔄"
	case "failed":
		return "❌"
	default:
		return "⏳"
	}
}
```

## Rules
1. **Default Behavior**: No arguments → show most recent mission. With ID → show specific mission.
2. **Tabular Output**: Use `tabwriter` for aligned columns. Never raw `Printf` for tabular data.
3. **Store Cleanup**: Always `defer store.Close()` after opening the SQLite store.

## Pitfalls

### Pitfall 1: Leaving SQLite connections open
Forgetting `defer store.Close()` causes file lock contention on Windows where SQLite locks are mandatory.

## Verify
```bash
go build ./cmd/orchestrator/...
```

## Checklist
- [ ] File `cmd/orchestrator/cmd/status.go` exists
- [ ] Shows most recent mission when no ID provided
- [ ] Shows specific mission when ID provided
- [ ] Renders task table with status icons
- [ ] `go build ./cmd/orchestrator/...` passes
