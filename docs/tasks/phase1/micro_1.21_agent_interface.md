# Micro-Task 1.21: Tạo contracts/agent/agent.go

## Thông tin
- **File tạo**: `contracts/agent/agent.go`
- **Package**: `agent`
- **Dependencies trước**: 1.17 (capability.go), 1.18 (task.go), 1.19 (result.go)
- **Thời gian**: 10 phút
- **Verify**: `go build ./contracts/agent/...`

## Nội dung CHÍNH XÁC cần tạo

```go
package agent

import "context"

// Agent is the core interface that all AI agents must implement.
//
// An agent is a specialized AI persona with specific capabilities.
// Examples: "Backend Developer", "Code Reviewer", "DevOps Engineer".
//
// Each agent:
//   - Has a role and capabilities (defined in Manifest)
//   - Uses a provider to communicate with an AI model
//   - Can use tools to interact with the outside world
//   - Receives Tasks and returns Results
//
// Implementation guidelines:
//   - Execute() must be safe for concurrent use (thread-safe).
//   - Execute() must respect context cancellation (ctx.Done()).
//   - CanHandle() must be fast (no I/O, no provider calls).
//   - Name() and Role() must return constant values.
//
// Lifecycle:
//   Agent lifecycle (Init, Start, Stop) is managed by the Plugin interface
//   in contracts/plugin. This interface only defines runtime behavior.
//   DO NOT add Init/Start/Stop methods here.
type Agent interface {
	// Name returns the unique identifier (e.g., "backend", "reviewer", "devops").
	//
	// Rules:
	//   - Lowercase, alphanumeric + hyphens only
	//   - Must match the name in the agent's Manifest
	//   - Must not change after initialization
	Name() string

	// Role returns the human-readable role description.
	// Example: "Backend Developer", "Code Reviewer", "DevOps Engineer"
	Role() string

	// Capabilities returns the list of capabilities this agent has.
	// Used by the orchestrator to match tasks to agents.
	Capabilities() []Capability

	// Execute performs a task and returns the result.
	//
	// Typical implementation flow:
	//   1. Build prompt from system prompt + task description + context
	//   2. Send prompt to provider via provider.Send()
	//   3. If AI response contains tool calls:
	//      a. Execute each tool
	//      b. Send tool results back to AI
	//      c. Go to step 2 (repeat until no more tool calls)
	//   4. Build Result from the AI's final response
	//
	// Error handling:
	//   - System errors (network down, panic) → return (nil, error)
	//   - Task failures (AI says "I can't") → return (Result{Status: Failed}, nil)
	//   - NEVER return (non-nil Result, non-nil error) simultaneously
	//
	// Context cancellation:
	//   - When ctx.Done() fires, stop immediately
	//   - Return (nil, context.Canceled) or (nil, context.DeadlineExceeded)
	//
	// Timeout:
	//   - The caller wraps ctx with context.WithTimeout(ctx, task.Timeout)
	//   - Agent does NOT need to manage timeout itself
	Execute(ctx context.Context, task *Task) (*Result, error)

	// CanHandle checks if this agent can handle the given task.
	//
	// The orchestrator calls CanHandle() on all registered agents
	// to find the right one for a task. First agent that returns true wins.
	//
	// Default implementation (in SDK BaseAgent):
	//   Check if task.Type matches any of this agent's Capabilities.
	//
	// Override for custom matching:
	//   - "backend" agent only handles tasks where Input["language"] == "go"
	//   - "reviewer" agent handles any "code_review" task
	//
	// Performance requirement:
	//   - Must NOT perform I/O (no file reads, no API calls)
	//   - Must return in < 1ms
	//   - Called frequently by the orchestrator
	CanHandle(task *Task) bool
}
```

## ⚠️ Pitfalls cần tránh
1. **Execute return convention**: NGHIÊM NGẶT:
   - Thành công: `return (result, nil)` với `result.Status = StatusSuccess`
   - Task thất bại: `return (result, nil)` với `result.Status = StatusFailed`
   - System error: `return (nil, err)`
   - KHÔNG BAO GIỜ: `return (result, err)` cả hai non-nil → consumer không biết dùng cái nào
2. **KHÔNG có Init/Start/Stop**: Lifecycle = Plugin interface. Mixing concerns = spaghetti code.
3. **CanHandle PHẢI nhanh**: Orchestrator gọi CanHandle() trên N agents × M tasks. Nếu mỗi call tốn 100ms → performance disaster.
4. **Thread-safety**: Nhiều goroutines có thể gọi Execute() cùng lúc (parallel tasks). Agent implementation PHẢI thread-safe.

## Checklist
- [ ] File `contracts/agent/agent.go` tồn tại
- [ ] Package: `package agent`
- [ ] Agent interface với ĐÚNG 4 methods: Name, Role, Capabilities, Execute
- [ ] Execute nhận `context.Context` là param đầu tiên
- [ ] Execute trả về `(*Result, error)` — PHẢI pointer Result
- [ ] CanHandle nhận `*Task` — PHẢI pointer Task
- [ ] KHÔNG có Init, Start, Stop, Close methods
- [ ] Godoc comments chi tiết với implementation flow
- [ ] Error handling convention documented
- [ ] `go build ./contracts/agent/...` không lỗi
