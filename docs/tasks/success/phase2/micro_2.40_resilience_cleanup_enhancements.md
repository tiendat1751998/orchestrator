# Micro-Task 2.40: Update Executor and Runtime (Graceful Degradation & Leak Detection)

## Info
- **Files updated**:
  - `kernel/runtime/executor.go` (Update `ExecuteTask` to support fallback agents)
  - `kernel/runtime/runtime.go` (Update `Stop` to support resource leak detection)
- **Package**: `runtime`
- [] Depends on: 2.22 (executor.go), 2.25 (runtime.go), 2.36 (resilience)
- **Time**: 20 min
- **Verify**: `go build ./kernel/runtime/...`

## Purpose
Upgrades task execution reliability by implementing graceful degradation (routing transient failures to fallback agents) and resource leak detection (verifying that worker threads exit and channels are drained during shutdowns).

## EXACT code to create

### Part 1: Update `kernel/runtime/executor.go`

In `kernel/runtime/executor.go`, replace `ExecuteTask` and add the `executeWithRecovery` helper:

```go
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
		return nil, fmt.Errorf("executor: no agents registered can handle task type %q", task.Type)
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
		} else {
			if e.eventBus != nil {
				eventbus.PublishTaskCompleted(e.eventBus, taskID, res)
			}
		}
	}()

	return a.Execute(ctx, t)
}
```

---

### Part 2: Update `kernel/runtime/runtime.go`

In `kernel/runtime/runtime.go`, replace `Stop` with the version that includes leak detection:

```go
// Stop gracefully shuts down the runtime and verifies resource release.
func (r *Runtime) Stop(ctx context.Context) error {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return nil // Already stopped, no-op
	}
	r.running = false
	r.mu.Unlock()

	if r.logger != nil {
		r.logger.Info("runtime stopping...")
	}

	// Step 1: Stop accepting new tasks
	r.dispatcher.Stop()

	// Step 2: Wait for in-flight tasks (with timeout)
	done := make(chan struct{})
	go func() {
		r.pool.Wait()
		close(done)
	}()

	select {
	case <-done:
		if r.logger != nil {
			r.logger.Info("all workers completed")
		}
	case <-ctx.Done():
		if r.logger != nil {
			r.logger.Warn("shutdown timeout reached, some tasks may be interrupted")
		}
	}

	// Step 3: Stop result processor
	if r.resultCancel != nil {
		r.resultCancel()
	}
	r.resultWg.Wait()

	// Step 4: Verification of resources (Leak Detection)
	stats := r.pool.Stats()
	if stats.ActiveWorkers > 0 {
		if r.logger != nil {
			r.logger.Error("resource leak: worker goroutines are still active after shutdown",
				"active_workers", stats.ActiveWorkers,
				"submitted", stats.TotalSubmitted,
				"completed", stats.TotalCompleted,
			)
		}
	}

	// Verify the result channel is drained
	undrainedCount := 0
DrainLoop:
	for {
		select {
		case <-r.dispatcher.Results():
			undrainedCount++
		default:
			break DrainLoop
		}
	}
	if undrainedCount > 0 {
		if r.logger != nil {
			r.logger.Warn("resource warning: undrained task results remaining in channel",
				"count", undrainedCount,
			)
		}
	}

	if r.logger != nil {
		r.logger.Info("runtime stopped",
			"total_submitted", stats.TotalSubmitted,
			"total_completed", stats.TotalCompleted,
		)
	}

	return nil
}
```

## Rules
1. **Fatal Failure Escalation**: If an agent returns a fatal (non-retryable) error (e.g. `ValidationError`), fail immediately instead of querying fallback agents.
2. **Context Lifetimes per Attempt**: Create a fresh context with timeout (`context.WithTimeout`) for each individual agent attempt, and call `cancel()` immediately after each run to free resources.
3. **Non-Blocking Channel Drains**: Implement channel leak checks using non-blocking loops (a `select` with a `default` path). Do not block on results channel reads when checking for undrained items.

## ⚠️ Pitfalls

### Pitfall 1: Returning dirty pointer variables on errors
```go
```
Always return explicit `nil` pointers alongside non-nil errors.

### Pitfall 2: Blocking on channel reads during shutdown checks
Reading from channels without a `default` fallback inside the select statement will hang the shutdown routine if the channel is empty and not closed. Always use a `default` case to make it non-blocking.

## Verify
```bash
go build ./kernel/runtime/...
```

## Checklist
- [ ] File `kernel/runtime/executor.go` updated successfully
- [ ] `ExecuteTask` loops over fallback agents for retryable errors
- [ ] Loops break immediately on fatal errors (`!IsRetryable`)
- [ ] Context cancellation functions are called for each attempt context
- [ ] File `kernel/runtime/runtime.go` updated successfully
- [ ] `Stop` validates that no active workers remain, logging errors if any are found
- [ ] `Stop` drains the results channel using non-blocking loops
- [ ] `go build ./kernel/runtime/...` passes
