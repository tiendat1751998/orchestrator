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
