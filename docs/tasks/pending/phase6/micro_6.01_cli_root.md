# Micro-Task 6.01: Create cmd/orchestrator/main.go

## Info
- **File**: `cmd/orchestrator/main.go`
- **Package**: `main`
- **Depends on**: Phase 2 (kernel/config), Phase 5 (kernel/orchestrator)
- **Time**: 25 min
- **Verify**: `go build ./cmd/orchestrator/...`

## External dependencies to add
```bash
go get github.com/spf13/cobra@latest
```

## Purpose
Implements the CLI entry point with Cobra root command, persistent flags (`--config`, `--verbose`), graceful shutdown via `signal.NotifyContext`, and version injection via `ldflags`.

## EXACT code to create

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/tiendat1751998/orchestrator/cmd/orchestrator/cmd"
)

// Injected at build time via ldflags:
//
//	go build -ldflags "-X main.version=1.0.0 -X main.commitHash=abc123"
var (
	version    = "dev"
	commitHash = "unknown"
	buildDate  = "unknown"
)

func main() {
	// Root context with OS signal interception for graceful shutdown.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	rootCmd := newRootCommand()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	var (
		configPath string
		verbose    bool
	)

	rootCmd := &cobra.Command{
		Use:   "orchestrator",
		Short: "AI-powered multi-agent orchestrator",
		Long: `Orchestrator coordinates multiple AI agents to decompose, plan,
and execute complex software engineering tasks using DAG-based pipelines.`,
		Version:       fmt.Sprintf("%s (commit: %s, built: %s)", version, commitHash, buildDate),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			// Initialize logger based on verbosity flag
			level := slog.LevelInfo
			if verbose {
				level = slog.LevelDebug
			}
			logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
			slog.SetDefault(logger)

			// Store config path in context for subcommands
			c.SetContext(cmd.WithConfigPath(c.Context(), configPath))
			return nil
		},
	}

	// Persistent flags apply to all subcommands
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "orchestrator.yaml", "path to configuration file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging")

	// Register subcommands
	rootCmd.AddCommand(
		cmd.NewMissionCmd(),
		cmd.NewStatusCmd(),
		cmd.NewAgentsCmd(),
		cmd.NewProvidersCmd(),
		cmd.NewConfigCmd(),
	)

	return rootCmd
}
```

## Context helper (same file or cmd/context.go)

```go
// Package cmd contains CLI subcommand implementations.
package cmd

import "context"

type contextKey string

const configPathKey contextKey = "config_path"

// WithConfigPath stores the config file path in context.
func WithConfigPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, configPathKey, path)
}

// ConfigPathFrom retrieves the config file path from context.
func ConfigPathFrom(ctx context.Context) string {
	if v, ok := ctx.Value(configPathKey).(string); ok {
		return v
	}
	return "orchestrator.yaml"
}
```

## Rules
1. **Version Injection**: Version, commit hash, and build date MUST be injected via `ldflags` at compile time. Never hardcode versions.
2. **Signal Interception**: Use `signal.NotifyContext` (Go 1.16+) instead of manual `signal.Notify` + goroutine patterns. Cleaner API, automatic context cancellation.
3. **Silent Errors**: Set `SilenceUsage: true` and `SilenceErrors: true` on root command. Cobra's default error handling prints usage text on every error, cluttering terminal output.

## Pitfalls

### Pitfall 1: Using os.Exit inside command handlers
```go
// WRONG:
RunE: func(cmd *cobra.Command, args []string) error {
    os.Exit(1) // Skips deferred cleanup functions!
}

// CORRECT:
RunE: func(cmd *cobra.Command, args []string) error {
    return fmt.Errorf("failed: %w", err) // Let cobra propagate errors to main()
}
```
`os.Exit` bypasses all `defer` statements, causing resource leaks. Return errors and let `main()` call `os.Exit`.

### Pitfall 2: Registering global logger before PersistentPreRunE
Initializing the logger at package init time means verbosity flags haven't been parsed yet. Always configure the logger inside `PersistentPreRunE`.

## Verify
```bash
go build ./cmd/orchestrator/...
```

## Checklist
- [ ] File `cmd/orchestrator/main.go` exists
- [ ] Package: `main`
- [ ] Cobra root command with `--config` and `--verbose` persistent flags
- [ ] `signal.NotifyContext` for graceful shutdown
- [ ] Version injected via ldflags
- [ ] Subcommands registered: mission, status, agents, providers, config
- [ ] `go build ./cmd/orchestrator/...` passes
