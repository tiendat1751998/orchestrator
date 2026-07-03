# Micro-Task 6.02: Create cmd/orchestrator/cmd/mission.go

- **File**: `cmd/orchestrator/cmd/mission.go`
- **Package**: `cmd`
- **Depends on**: 6.01, 6.07 (progress UI), 5.08 (orchestrator)
- **Time**: 30 min
- **Verify**: `go build ./cmd/orchestrator/...`

## Purpose
Implements the `orchestrator mission` command that accepts a mission description, parses it into a structured `goal.Goal` contract, and executes the FSM pipeline through the Orchestrator.

## EXACT code to create

```go
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tiendat1751998/orchestrator/cmd/orchestrator/ui"
	"github.com/tiendat1751998/orchestrator/kernel/config"
	"github.com/tiendat1751998/orchestrator/kernel/orchestrator"
	"github.com/tiendat1751998/orchestrator/contracts/goal"
)

// NewMissionCmd creates the `mission` subcommand.
func NewMissionCmd() *cobra.Command {
	var (
		missionFile string
		dryRun      bool
	)

	cmd := &cobra.Command{
		Use:   "mission [description]",
		Short: "Execute an AI-powered mission",
		Long: `Submit a mission for the orchestrator to plan, score, and execute.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			logger := slog.Default()

			// 1. Build goal from input args
			g, err := buildGoal(args, missionFile)
			if err != nil {
				return err
			}

			// 2. Load configuration
			cfgPath := ConfigPathFrom(ctx)
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// 3. Initialize orchestrator
			orch, err := orchestrator.NewOrchestrator(nil, nil, logger) // parameters passed here
			if err != nil {
				return fmt.Errorf("failed to initialize orchestrator: %w", err)
			}

			// 4. Start progress UI
			progress := ui.NewProgressRenderer(os.Stdout)
			progress.Start()
			defer progress.Stop()

			// 5. Execute mission
			startTime := time.Now()
			result, err := orch.Execute(ctx, "mission-uuid", *g)
			if err != nil {
				return fmt.Errorf("mission failed: %w", err)
			}

			elapsed := time.Since(startTime)
			progress.RenderFinalResult(result, elapsed)

			return nil
		},
	}

	cmd.Flags().StringVarP(&missionFile, "file", "f", "", "path to mission YAML file")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "plan and score without executing")

	return cmd
}

func buildGoal(args []string, filePath string) (*goal.Goal, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("mission description required")
	}
	return &goal.Goal{
		RawInput: args[0],
		Objectives: []goal.Objective{
			{ID: "obj_main", Description: args[0]},
		},
	}, nil
}
```

## Verify
```bash
go build ./cmd/orchestrator/...
```

## Checklist
- [ ] File `cmd/orchestrator/cmd/mission.go` exists
- [ ] Package: `cmd`
- [ ] Maps command-line input to `goal.Goal` struct
- [ ] Invokes `orchestrator.Execute` with FSM lifecycle
- [ ] `go build ./cmd/cmd/orchestrator/...` passes
