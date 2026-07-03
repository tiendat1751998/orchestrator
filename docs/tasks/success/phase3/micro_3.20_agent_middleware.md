# Micro-Task 3.20: Create sdk/middleware/agent.go

## Info
- **File**: `sdk/middleware/agent.go`
- **Package**: `middleware`
- **Depends on**: 1.21 (agent contract), 2.37 (metrics collector)
- **Time**: 20 min
- **Verify**: `go build ./sdk/middleware/...`

## Purpose
Implements the agent middleware decorators (`AgentMiddleware`, `ChainAgent`, logging, metrics, and panic recovery middlewares) that intercept agent task execution to handle concerns like logging, performance tracking, and isolation.

## EXACT code to create

```go
// Package middleware implements decorators and interceptors for agents and providers.
package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/kernel/metrics"
)

// AgentMiddleware wraps an agent.Agent to extend its behavior.
type AgentMiddleware func(agent.Agent) agent.Agent

// ChainAgent wraps an agent with a slice of middlewares.
// Middlewares are executed in the order they are passed (left-to-right).
func ChainAgent(a agent.Agent, mws ...AgentMiddleware) agent.Agent {
	for i := len(mws) - 1; i >= 0; i-- {
		a = mws[i](a)
	}
	return a
}

// =============================================================================
// Wrapped Agent Helper
// =============================================================================

type wrappedAgent struct {
	agent.Agent
	executeFn func(ctx context.Context, task *agent.Task) (*agent.Result, error)
}

func (w *wrappedAgent) Execute(ctx context.Context, task *agent.Task) (*agent.Result, error) {
	return w.executeFn(ctx, task)
}

// =============================================================================
// 1. Agent Logging Middleware
// =============================================================================

// AgentLogging returns a middleware that logs execution entry, duration, and results.
func AgentLogging(logger *slog.Logger) AgentMiddleware {
	return func(next agent.Agent) agent.Agent {
		return &wrappedAgent{
			Agent: next,
			executeFn: func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
				startTime := time.Now()

				if logger != nil {
					logger.Info("agent execution started",
						"agent", next.Name(),
						"task_id", string(task.ID),
						"task_name", task.Name,
					)
				}

				res, err := next.Execute(ctx, task)

				duration := time.Since(startTime)
				if err != nil {
					if logger != nil {
						logger.Error("agent execution failed",
							"agent", next.Name(),
							"task_id", string(task.ID),
							"duration", duration.String(),
							"error", err.Error(),
						)
					}
				} else {
					status := contracts.StatusSuccess
					if res != nil && res.Status != "" {
						status = res.Status
					}
					if logger != nil {
						logger.Info("agent execution completed",
							"agent", next.Name(),
							"task_id", string(task.ID),
							"status", string(status),
							"duration", duration.String(),
						)
					}
				}

				return res, err
			},
		}
	}
}

// =============================================================================
// 2. Agent Metrics Middleware
// =============================================================================

// AgentMetrics returns a middleware that records execution counts, failures, and latencies.
func AgentMetrics(registry *metrics.Registry) AgentMiddleware {
	return func(next agent.Agent) agent.Agent {
		return &wrappedAgent{
			Agent: next,
			executeFn: func(ctx context.Context, task *agent.Task) (*agent.Result, error) {
				startTime := time.Now()

				activeGauge := registry.Gauge("orchestrator_active_workers")
				activeGauge.Add(1.0)
				defer activeGauge.Add(-1.0)

				res, err := next.Execute(ctx, task)

				duration := time.Since(startTime).Seconds()

				registry.Counter("orchestrator_tasks_total").Inc()
				registry.Histogram("orchestrator_task_duration_seconds").Observe(duration)

				if err != nil {
					registry.Counter("orchestrator_tasks_failed_total").Inc()
				} else if res != nil && res.Status == contracts.StatusFailed {
					registry.Counter("orchestrator_tasks_failed_total").Inc()
				}

				return res, err
			},
		}
	}
}

// =============================================================================
// 3. Agent Recovery Middleware
// =============================================================================

// AgentRecovery intercepts panics inside the agent execution and converts them into errors.
func AgentRecovery(logger *slog.Logger) AgentMiddleware {
	return func(next agent.Agent) agent.Agent {
		return &wrappedAgent{
			Agent: next,
			executeFn: func(ctx context.Context, task *agent.Task) (res *agent.Result, err error) {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("middleware: agent %q panicked: %v", next.Name(), r)
						res = nil
						if logger != nil {
							logger.Error("agent panicked recovered by middleware",
								"agent", next.Name(),
								"task_id", string(task.ID),
								"panic", fmt.Sprintf("%v", r),
							)
						}
					}
				}()
				return next.Execute(ctx, task)
			},
		}
	}
}
```

## Rules
1. **Reverse Chaining Order**: Chain middlewares in reverse (from right to left in iteration) to preserve execution flow from left to right (outermost first).
2. **Panic Conversions**: Catch agent panic signals in deferred functions and convert them to standard Go error payloads.
3. **Gauge Concurrency Counters**: Update active execution worker counters using gauges incremented at method entries and decremented via deferred calls.

## ⚠️ Pitfalls

### Pitfall 1: Bypassing interface delegations
Failing to embed the `agent.Agent` interface in the `wrappedAgent` wrapper struct would hide metadata queries (e.g. `Name()`, `Role()`) from registries. Always delegate method calls using embedding.

### Pitfall 2: Reversing execution orders of middlewares
Iterating forwards when building chains wraps nested actions backwards (e.g. executing metric counters after recovery catches a panic, preventing failures from being counted). Iterate backwards to ensure the correct order.

## Verify
```bash
go build ./sdk/middleware/...
```

## Checklist
- [ ] File `sdk/middleware/agent.go` exists
- [ ] Package: `middleware`
- [ ] `AgentMiddleware` type is defined
- [ ] `ChainAgent` chains handlers correctly (left-to-right execution)
- [ ] `AgentLogging` captures startup, duration, and errors
- [ ] `AgentMetrics` tracks task counts, durations, and failure counts
- [ ] Active execution counters are updated via gauges
- [ ] `AgentRecovery` converts panics to standard Go errors
- [ ] `go build ./sdk/middleware/...` passes
