// Package runtime provides the task execution engine.
//
// Architecture:
//
//	Executor   — runs a single task on an agent (timeout, panic recovery, events)
//	Pool       — limits concurrent goroutines (semaphore pattern)
//	Dispatcher — routes tasks from scheduler to workers
//	Runtime    — ties everything together (Start/Stop lifecycle)
package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/event"
	"github.com/tiendat1751998/orchestrator/kernel/eventbus"
	"github.com/tiendat1751998/orchestrator/kernel/registry"
)

// Executor runs a single task on a suitable agent.
//
// Execution steps:
//  1. Find agent from registry (via FindAgentForTask)
//  2. Create child context with timeout
//  3. Emit "task.started" event
//  4. Call agent.Execute(ctx, task) with panic recovery
//  5. Emit "task.completed" or "task.failed" event
//  6. Return result or error
//
// Thread-safety: Executor is stateless and safe for concurrent use.
// Multiple goroutines can call ExecuteTask simultaneously.
type Executor struct {
	registry *registry.Registry
	eventBus event.Bus
	logger   *slog.Logger

	// defaultTimeout is used when task.Timeout is not set.
	defaultTimeout time.Duration
}

// ExecutorConfig configures the Executor.
type ExecutorConfig struct {
	// DefaultTimeout for tasks without an explicit timeout.
	// Default: 120s (matching the provider timeout).
	DefaultTimeout time.Duration
}

// NewExecutor creates a new Executor.
func NewExecutor(reg *registry.Registry, bus event.Bus, logger *slog.Logger, cfg ExecutorConfig) *Executor {
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = 120 * time.Second
	}

	return &Executor{
		registry:       reg,
		eventBus:       bus,
		logger:         logger,
		defaultTimeout: cfg.DefaultTimeout,
	}
}

// ExecuteTask runs a single task end-to-end with graceful degradation (fallback agents).
//
// If the primary matched agent fails with a retryable error, the executor will
// automatically attempt to execute the task on fallback agents that are registered
// in the registry and capable of handling the task.
func (e *Executor) ExecuteTask(ctx context.Context, task *agent.Task) (result *agent.Result, err error) {
	startTime := time.Now()
	taskID := string(task.ID)

	// Step 1: Find all matching agents
	agents := e.registry.FindAllAgentsForTask(task)
	if len(agents) == 0 {
		return nil, fmt.Errorf("executor: registry: no agent can handle task type %q", task.Type)
	}

	var lastErr error
	timeout := e.defaultTimeout
	if task.Timeout > 0 {
		timeout = task.Timeout
	}

	// Try each matching agent sequentially (graceful degradation)
	for i, selectedAgent := range agents {
		agentName := selectedAgent.Name()

		// Create fresh child context for this specific agent attempt
		attemptCtx, cancel := context.WithTimeout(ctx, timeout)

		// Emit task.started for this attempt
		if e.eventBus != nil {
			eventbus.PublishTaskStarted(e.eventBus, taskID, agentName)
		}

		if e.logger != nil {
			e.logger.Info("task execution attempt started",
				"task_id", taskID,
				"task_name", task.Name,
				"agent", agentName,
				"attempt", i+1,
				"total_attempts", len(agents),
			)
		}

		// Run with panic recovery for this specific agent
		result, err = e.executeWithRecovery(attemptCtx, selectedAgent, task, taskID, startTime)
		cancel() // Release context resources immediately

		if err == nil {
			return result, nil
		}

		lastErr = err

		// If error is NOT retryable, fail immediately (do not try fallback agents)
		if !contracts.IsRetryable(err) {
			break
		}

		if e.logger != nil {
			e.logger.Warn("agent failed with transient error, trying fallback",
				"task_id", taskID,
				"failed_agent", agentName,
				"error", err,
			)
		}
	}

	return nil, fmt.Errorf("executor: all matching agents failed. Last error: %w", lastErr)
}

// executeWithRecovery wraps agent.Execute with panic recovery.
func (e *Executor) executeWithRecovery(
	ctx context.Context,
	a agent.Agent,
	t *agent.Task,
	taskID string,
	startTime time.Time,
) (res *agent.Result, err error) {
	agentName := a.Name()

	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			err = fmt.Errorf("executor: agent %q panicked: %v\n%s", agentName, r, stack)
			res = nil

			if e.logger != nil {
				e.logger.Error("agent panicked during execution",
					"task_id", taskID,
					"agent", agentName,
					"panic", fmt.Sprintf("%v", r),
					"stack", stack,
				)
			}
		}

		duration := time.Since(startTime)
		if err != nil {
			if e.eventBus != nil {
				eventbus.PublishTaskFailed(e.eventBus, taskID, err)
			}
			if e.logger != nil {
				e.logger.Error("task execution failed",
					"task_id", taskID,
					"agent", agentName,
					"duration", duration,
					"error", err,
				)
			}
		} else {
			if e.eventBus != nil {
				eventbus.PublishTaskCompleted(e.eventBus, taskID, res)
			}
			if e.logger != nil {
				e.logger.Info("task execution completed",
					"task_id", taskID,
					"agent", agentName,
					"duration", duration,
					"status", res.Status,
				)
			}
		}
	}()

	return a.Execute(ctx, t)
}
