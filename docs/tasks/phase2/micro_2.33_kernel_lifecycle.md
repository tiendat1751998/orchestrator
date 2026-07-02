# Micro-Task 2.33: Create kernel/lifecycle/lifecycle.go

## Info
- **File**: `kernel/lifecycle/lifecycle.go`
- **Package**: `lifecycle`
- **Depends on**: 2.32 (kernel.go)
- **Time**: 15 min
- **Verify**: `go build ./kernel/lifecycle/...`

## Purpose
OS signal handling for graceful shutdown. Catches SIGINT (Ctrl+C) and SIGTERM.

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
//
// This function BLOCKS until shutdown is complete.
func WaitForShutdown(ctx context.Context, target Shutdownable, timeout time.Duration, logger *slog.Logger) {
	// Create signal channel with buffer size 1.
	//
	// WHY buffer size 1?
	// → signal.Notify is non-blocking: it sends to the channel but does NOT
	//   wait for a reader. If the channel is unbuffered and nobody is reading
	//   at the exact moment → signal is dropped.
	// → Buffer size 1 ensures the first signal is always captured.
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

## Pitfalls

### Pitfall 1: Signal channel buffer size = 1
```go
sigChan := make(chan os.Signal, 1)  // Buffered
// NOT:
sigChan := make(chan os.Signal)      // Unbuffered — signal may be dropped
```

### Pitfall 2: Second signal = force exit
Users expect Ctrl+C twice = force quit (like most Unix programs).
Without this → stuck shutdown → user kills process → data corruption risk.

### Pitfall 3: signal.Stop(sigChan) in defer
```go
defer signal.Stop(sigChan)
```
After shutdown, stop relaying signals to our channel.
Otherwise, the signal handler stays registered → potential resource leak.

### Pitfall 4: context.Background() for shutdown
```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
```
NOT `ctx` — the original context is already cancelled (that's why we're shutting down).
Using the cancelled ctx → shutdown immediately times out → no graceful cleanup.

### Pitfall 5: Shutdownable interface decouples from kernel
```go
type Shutdownable interface {
    Stop(ctx context.Context) error
}
```
Any component with Stop() works. Not just the kernel.
Testable without importing the kernel package.

## Checklist
- [ ] File `kernel/lifecycle/lifecycle.go` exists
- [ ] Package: `package lifecycle`
- [ ] Shutdownable interface with Stop method
- [ ] `WaitForShutdown(ctx, target, timeout, logger)` — blocks until signal
- [ ] `WaitForShutdownWithDefaults()` convenience wrapper
- [ ] Handles SIGINT and SIGTERM
- [ ] Second signal → force exit (os.Exit(1))
- [ ] Signal channel buffer size 1
- [ ] defer signal.Stop
- [ ] Shutdown uses context.Background() (not cancelled ctx)
- [ ] `go build ./kernel/lifecycle/...` no errors
