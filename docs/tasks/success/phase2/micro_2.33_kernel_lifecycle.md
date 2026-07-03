# Micro-Task 2.33: Create kernel/lifecycle/lifecycle.go

## Info
- **File**: `kernel/lifecycle/lifecycle.go`
- **Package**: `lifecycle`
- **Depends on**: 2.32 (kernel.go)
- **Time**: 15 min
- **Verify**: `go build ./kernel/lifecycle/...`

## Purpose
Implements OS signal handling (`Shutdownable`, `WaitForShutdown`, `WaitForShutdownWithDefaults`) to capture SIGINT (Ctrl+C) and SIGTERM events, initiating graceful shutdowns with configurable timeouts and supporting force-exit overrides on double-signals.

## EXACT code to create

```go
// Package lifecycle provides OS signal handling for graceful kernel shutdown.
//
// Usage:
//
//	kernel, _ := kernel.New(cfg)
//	kernel.Start(ctx)
//	lifecycle.WaitForShutdown(ctx, kernel) // Blocks until signal or context cancel
package lifecycle

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Shutdownable is any component that can be gracefully stopped.
// The kernel implements this interface.
type Shutdownable interface {
	Stop(ctx context.Context) error
}

// WaitForShutdown blocks until an OS signal or context cancellation occurs,
// then gracefully shuts down the given component.
//
// Signals handled:
//   - SIGINT (Ctrl+C in terminal)
//   - SIGTERM (sent by process managers like systemd, Docker, Kubernetes)
//
// Shutdown flow:
//   1. First signal → graceful shutdown with timeout
//   2. Second signal → force exit (os.Exit(1))
//
// Parameters:
//   - ctx: parent context (cancelled = shutdown)
//   - target: component to shut down (e.g., kernel)
//   - timeout: maximum time for graceful shutdown
//   - logger: for shutdown progress logging (can be nil)
func WaitForShutdown(ctx context.Context, target Shutdownable, timeout time.Duration, logger *slog.Logger) {
	// Create signal channel with buffer size 1.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan) // Stop relaying signals after we're done

	// Wait for shutdown trigger
	select {
	case sig := <-sigChan:
		if logger != nil {
			logger.Info("received shutdown signal",
				"signal", sig.String(),
			)
		}
	case <-ctx.Done():
		if logger != nil {
			logger.Info("context cancelled, initiating shutdown")
		}
	}

	// Handle second signal (force exit)
	//
	// If user presses Ctrl+C again during graceful shutdown → force exit.
	// This is a safety valve for stuck shutdowns.
	go func() {
		select {
		case sig := <-sigChan:
			if logger != nil {
				logger.Warn("received second signal, forcing exit",
					"signal", sig.String(),
				)
			}
			os.Exit(1) // Force exit
		}
	}()

	// Graceful shutdown with timeout
	//
	// Note: We use context.Background() because the original context is already cancelled.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if logger != nil {
		logger.Info("shutting down gracefully",
			"timeout", timeout.String(),
		)
	}

	if err := target.Stop(shutdownCtx); err != nil {
		if logger != nil {
			logger.Error("shutdown error", "error", err)
		}
	}

	if logger != nil {
		logger.Info("shutdown complete")
	}
}

// WaitForShutdownWithDefaults is a convenience wrapper with 30s timeout.
func WaitForShutdownWithDefaults(ctx context.Context, target Shutdownable, logger *slog.Logger) {
	WaitForShutdown(ctx, target, 30*time.Second, logger)
}
```

## Rules
1. **Buffered Signal Channels**: Enforce buffer size `1` on OS signal notification channels. Unbuffered channels drop signals if no consumer is active at the exact nanosecond the OS raises them.
2. **Double-Signal Force-Exits**: Register a background listener thread to handle a second signal during graceful shutdowns. This allows forcing shutdowns (`os.Exit(1)`) if a component hangs.
3. **Isolated Cleanup Contexts**: Graceful shutdowns must use a fresh parent context (`context.Background()`) when invoking timeouts. Reusing the canceled parent context (`ctx`) causes shutdowns to time out immediately.

## ⚠️ Pitfalls

### Pitfall 1: Reusing canceled parent contexts inside cleanup steps
Reusing the original canceled context `ctx` to configure shutdown timeout bounds causes calls to `target.Stop(ctx)` to abort immediately without running cleanups. Always use `context.Background()`.

### Pitfall 2: Using unbuffered signal channels
```go
```
Use `make(chan os.Signal, 1)` to ensure signals are queued.

## Verify
```bash
go build ./kernel/lifecycle/...
```

## Checklist
- [ ] File `kernel/lifecycle/lifecycle.go` exists
- [ ] Package: `lifecycle`
- [ ] `Shutdownable` interface defines `Stop(context.Context) error`
- [ ] `WaitForShutdown` captures SIGINT and SIGTERM signals
- [ ] Second signal launches force exit handler (`os.Exit(1)`)
- [ ] Channels are buffered to size 1
- [ ] Deferred closures call `signal.Stop`
- [ ] Shutdown timeout uses `context.Background()` contexts
- [ ] `go build ./kernel/lifecycle/...` passes
