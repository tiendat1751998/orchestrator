# Micro-Task 2.40: Cập nhật Executor và Runtime (Graceful Degradation & Leak Detection)

## Thông tin
- **File cập nhật**: 
  - `kernel/runtime/executor.go` (Cập nhật ExecuteTask hỗ trợ Fallback Agents)
  - `kernel/runtime/runtime.go` (Cập nhật Stop hỗ trợ phát hiện rò rỉ Worker/Channel)
- **Package**: `runtime`
- **Dependencies trước**: 2.22 (executor.go), 2.25 (runtime.go), 2.36 (resilience)
- **Thời gian**: 20 phút
- **Verify**: `go build ./kernel/runtime/...`

## Purpose
Nâng cấp hai thành phần trung tâm của Task Execution:
1. `Executor.ExecuteTask` hỗ trợ suy giảm chất lượng dịch vụ có kiểm soát (Graceful Degradation). Nếu agent được chỉ định bị lỗi tạm thời (retryable error), hệ thống tự động tìm và thử thực thi trên các fallback agents khác có cùng khả năng trước khi báo hỏng hoàn toàn.
2. `Runtime.Stop` thêm cơ chế phát hiện rò rỉ tài nguyên (Resource Leak Detection) khi đóng máy: kiểm tra số lượng active workers còn dư và cảnh báo nếu kết quả trên kết quả kênh (results channel) chưa được xử lý hết.

## EXACT code to create

### Phần 1: Cập nhật `kernel/runtime/executor.go`

Thay thế phương thức `ExecuteTask` trong `kernel/runtime/executor.go` bằng phiên bản hỗ trợ fallback agents dưới đây:

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

		e.logger.Info("task execution attempt started",
			"task_id", taskID,
			"task_name", task.Name,
			"agent", agentName,
			"attempt", i+1,
			"total_attempts", len(agents),
		)

		// Run with panic recovery for this specific agent
		result, err = e.executeWithRecovery(attemptCtx, selectedAgent, task, taskID, startTime)
		cancel() // Release context resources immediately

		if err == nil {
			// Success! Return result
			return result, nil
		}

		lastErr = err

		// If error is NOT retryable, fail immediately (do not try fallback agents)
		// ValidationError, AuthFailed, PermissionDenied are fatal.
		// Timeout, ProviderUnavailable are retryable.
		if !contracts.IsRetryable(err) {
			break
		}

		e.logger.Warn("agent failed with transient error, trying fallback",
			"task_id", taskID,
			"failed_agent", agentName,
			"error", err,
		)
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

			e.logger.Error("agent panicked during execution",
				"task_id", taskID,
				"agent", agentName,
				"panic", fmt.Sprintf("%v", r),
				"stack", stack,
			)
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

### Phần 2: Cập nhật `kernel/runtime/runtime.go`

Thay thế phương thức `Stop` trong `kernel/runtime/runtime.go` bằng phiên bản tăng cường kiểm tra rò rỉ tài nguyên:

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

	r.logger.Info("runtime stopping...")

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
		r.logger.Info("all workers completed")
	case <-ctx.Done():
		r.logger.Warn("shutdown timeout reached, some tasks may be interrupted")
	}

	// Step 3: Stop result processor
	if r.resultCancel != nil {
		r.resultCancel()
	}
	r.resultWg.Wait()

	// Step 4: Verification of resources (Leak Detection)
	stats := r.pool.Stats()
	if stats.ActiveWorkers > 0 {
		r.logger.Error("resource leak: worker goroutines are still active after shutdown",
			"active_workers", stats.ActiveWorkers,
			"submitted", stats.TotalSubmitted,
			"completed", stats.TotalCompleted,
		)
	}

	// Verify the result channel is drained
	undrainedCount := 0
	// Non-blocking drain verification
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
		r.logger.Warn("resource warning: undrained task results remaining in channel",
			"count", undrainedCount,
		)
	}

	r.logger.Info("runtime stopped",
		"total_submitted", stats.TotalSubmitted,
		"total_completed", stats.TotalCompleted,
	)

	return nil
}
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Trả về nil pointer của Result cùng với non-nil error
```go
// ❌ SAI:
res, err := agent.Execute(...)
if err != nil {
    return res, err // Nếu res khác nil (chứa con trỏ rác) sẽ làm caller confused.
}
```
Khi phát sinh lỗi, luôn đảm bảo trả về `nil` cho giá trị con trỏ Result để làm sạch luồng dữ liệu lỗi.

### Pitfall 2: Chặn vô tận (infinite block) khi kiểm tra rò rỉ kênh Results
```go
// ❌ SAI:
for res := range r.dispatcher.Results() { ... } // Vòng lặp này sẽ chặn vô hạn nếu kênh chưa được đóng.

// ✅ ĐÚNG:
for {
    select {
    case <-r.dispatcher.Results():
        undrainedCount++
    default:
        break DrainLoop // default case biến việc đọc thành non-blocking, lập tức thoát khi kênh trống.
    }
}
```
Kênh Results có thể không được đóng hoàn toàn tại thời điểm dừng. Đọc bằng `select` với `default` đảm bảo tính chất không chặn (non-blocking) để việc shutdown không bị treo.

## Checklist
- [ ] File `kernel/runtime/executor.go` cập nhật thành công
- [ ] `ExecuteTask` lặp qua danh sách fallback agents khi gặp lỗi retryable
- [ ] `ExecuteTask` ngắt vòng lặp lập tức nếu lỗi fatal (`!IsRetryable`)
- [ ] Con trỏ context của từng attempt được giải phóng bằng `cancel()`
- [ ] File `kernel/runtime/runtime.go` cập nhật thành công
- [ ] `Stop` kiểm tra số lượng active workers và log lỗi nếu > 0
- [ ] `Stop` dọn sạch kênh Results bằng non-blocking loop và log cảnh báo số lượng kết quả bị bỏ sót
- [ ] `go build ./kernel/runtime/...` không lỗi
