# Micro-Task 3.20: Create sdk/middleware/agent.go

## Info
- **File**: `sdk/middleware/agent.go`
- **Package**: `middleware`
- **Depends on**: 1.21 (agent contract), 2.37 (metrics collector)
- **Time**: 20 min
- **Verify**: `go build ./sdk/middleware/...`

## Purpose
Triển khai cơ chế Middleware cho Agents (`AgentMiddleware`). Điều này cho phép bọc (wrap) thực thi của Agent để tự động chèn các tính năng chéo (cross-cutting concerns) như ghi nhật ký chi tiết (Logging), thu thập chỉ số vận hành (Metrics), và tự phục hồi khi crash (Recovery) mà không làm ô nhiễm mã nguồn chính của Agent.

## EXACT code to create

```go
// Package middleware implements decorators and interceptors for agents and providers.
package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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
				
				logger.Info("agent execution started",
					"agent", next.Name(),
					"task_id", string(task.ID),
					"task_name", task.Name,
				)

				res, err := next.Execute(ctx, task)

				duration := time.Since(startTime)
				if err != nil {
					logger.Error("agent execution failed",
						"agent", next.Name(),
						"task_id", string(task.ID),
						"duration", duration.String(),
						"error", err.Error(),
					)
				} else {
					status := "success"
					if res.Status != "success" && res.Status != "" {
						status = string(res.Status)
					}
					logger.Info("agent execution completed",
						"agent", next.Name(),
						"task_id", string(task.ID),
						"status", status,
						"duration", duration.String(),
					)
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

				// Record active workers gauge
				activeGauge := registry.Gauge("orchestrator_active_workers")
				activeGauge.Add(1.0)
				defer activeGauge.Add(-1.0)

				res, err := next.Execute(ctx, task)

				duration := time.Since(startTime).Seconds()
				
				// Metrics indicators
				registry.Counter("orchestrator_tasks_total").Inc()
				registry.Histogram("orchestrator_task_duration_seconds").Observe(duration)

				if err != nil {
					registry.Counter("orchestrator_tasks_failed_total").Inc()
				} else if res != nil && res.Status == "failed" {
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

## ⚠️ Pitfalls

### Pitfall 1: Incorrect Middleware Chaining Order (Stack order)
```go
// ❌ WRONG:
// Nếu duyệt xuôi (i = 0; i < len; i++):
// Middlewares sẽ bị bọc lộn ngược thứ tự truyền vào.
// Ví dụ: Chain(a, Log, Recovery) -> Recovery sẽ chạy trước, Log chạy sau -> nếu Recovery panic, Log sẽ không bắt được duration.

// ✅ CORRECT:
// Duyệt ngược (i = len-1; i >= 0; i--):
// Đảm bảo thứ tự thực thi đúng từ trái qua phải (Log bọc ngoài, Recovery bọc trong).
```
Khi thiết kế bộ gộp Chain, bắt buộc phải duyệt ngược mảng middleware để bọc hành vi chính xác.

### Pitfall 2: Forgetting to forward metadata accessors
`wrappedAgent` phải embed interface `agent.Agent` trực tiếp để các phương thức metadata (`Name()`, `Role()`, `Capabilities()`, `CanHandle()`) được ủy quyền (delegated) thẳng tới agent cụ thể bên trong.

## Verify
```bash
go build ./sdk/middleware/...
```

## Checklist
- [ ] File `sdk/middleware/agent.go` tồn tại
- [ ] Package: `middleware`
- [ ] Định nghĩa kiểu hàm `AgentMiddleware`
- [ ] Hàm `ChainAgent` duyệt ngược mảng bọc middleware chính xác
- [ ] `AgentLogging` ghi nhận log bắt đầu, thành công và thất bại
- [ ] `AgentMetrics` tăng giảm số lượng worker hoạt động qua Gauge
- [ ] `AgentRecovery` bắt panic và chuyển thành Go error trả về an toàn
- [ ] `go build ./sdk/middleware/...` chạy thành công
