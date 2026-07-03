# Micro-Task 3.04: Create sdk/agent/agent.go

## Info
- **File**: `sdk/agent/agent.go`
- **Package**: `agent`
- **Depends on**: 3.01 (base_plugin.md), 3.02 (manifest_loader.md), 3.03 (prompt_builder.md), 1.12 (provider contract), 1.15 (tool contract), 1.21 (agent contract)
- **Time**: 30 min
- **Verify**: `go build ./sdk/agent/...`

## Purpose
Implements the core agent runner (`BaseAgent`, `StreamCallback`, and constructors) that manages the ReAct (Reasoning + Action) execution loop, handles tool execution, and supports streaming responses via context listeners.

## EXACT code to create

```go
package agent

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
	contractsplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
	"github.com/tiendat1751998/orchestrator/contracts/tool"
	sdkplugin "github.com/tiendat1751998/orchestrator/sdk/plugin"
	sdkprovider "github.com/tiendat1751998/orchestrator/sdk/provider"
)

type contextKey string

const StreamCallbackKey contextKey = "agent_stream_callback"

type StreamCallback func(delta string)

func WithStreamCallback(ctx context.Context, cb StreamCallback) context.Context {
	return context.WithValue(ctx, StreamCallbackKey, cb)
}

type BaseAgent struct {
	*sdkplugin.BasePlugin

	manifest *agent.Manifest
	p        provider.Provider
	logger   *slog.Logger

	mu    sync.RWMutex
	tools map[string]tool.Tool

	maxIterations int
}

func NewBaseAgent(m *agent.Manifest, p provider.Provider, logger *slog.Logger) (*BaseAgent, error) {
	if m == nil {
		return nil, fmt.Errorf("sdk/agent: manifest is nil")
	}
	if p == nil {
		return nil, fmt.Errorf("sdk/agent: provider is nil")
	}

	basePlugin, err := sdkplugin.NewBasePlugin(m.Name, contractsplugin.TypeAgent, m.Version)
	if err != nil {
		return nil, err
	}

	return &BaseAgent{
		BasePlugin:    basePlugin,
		manifest:      m,
		p:             p,
		logger:        logger,
		tools:         make(map[string]tool.Tool),
		maxIterations: 10,
	}
}

func (a *BaseAgent) Role() string {
	return a.manifest.Role
}

func (a *BaseAgent) Capabilities() []agent.Capability {
	return a.manifest.Capabilities
}

func (a *BaseAgent) Manifest() agent.Manifest {
	return *a.manifest
}

func (a *BaseAgent) CanHandle(task *agent.Task) bool {
	if task == nil {
		return false
	}
	for _, cap := range a.manifest.Capabilities {
		if string(cap) == task.Type {
			return true
		}
	}
	return false
}

func (a *BaseAgent) RegisterTool(t tool.Tool) error {
	if t == nil {
		return fmt.Errorf("sdk/agent: cannot register nil tool")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tools[t.Name()] = t
	return nil
}

func (a *BaseAgent) Execute(ctx context.Context, task *agent.Task) (*agent.Result, error) {
	if !a.IsStarted() {
		return nil, fmt.Errorf("sdk/agent: agent %q is not running", a.Name())
	}

	messages := BuildPrompt(a.manifest, task, a.logger)

	a.mu.RLock()
	toolDefs := make([]provider.ToolDefinition, 0, len(a.tools))
	for _, tName := range a.manifest.Tools {
		if t, ok := a.tools[tName]; ok {
			schema := t.Schema()
			if schema != nil {
				toolDefs = append(toolDefs, provider.ToolDefinition{
					Name:        t.Name(),
					Description: t.Description(),
					Parameters:  schema.Raw(),
				})
			}
		}
	}
	a.mu.RUnlock()

	iterations := 0
	startTime := time.Now()

	var totalUsage provider.Usage

	streamCb, hasStream := ctx.Value(StreamCallbackKey).(StreamCallback)

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		iterations++
		if iterations > a.maxIterations {
			return nil, fmt.Errorf("sdk/agent: agent %q exceeded max tool iterations limit (%d)",
				a.Name(), a.maxIterations)
		}

		req := &provider.Request{
			Model:          a.manifest.Model,
			Messages:       messages,
			Tools:          toolDefs,
			Temperature:    &a.manifest.Temperature,
			MaxTokens:      &a.manifest.MaxTokens,
			Stream:         hasStream,
			ResponseFormat: "text",
		}

		var resp *provider.Response
		var err error

		if hasStream {
			if a.logger != nil {
				a.logger.Debug("requesting streaming response from AI provider",
					"task_id", string(task.ID),
					"agent", a.Name(),
					"iteration", iterations,
				)
			}

			streamChan, streamErr := a.p.Stream(ctx, req)
			if streamErr != nil {
				return nil, fmt.Errorf("sdk/agent: provider stream initialization failed: %w", streamErr)
			}

			resp, err = sdkprovider.CollectStream(ctx, streamChan)
			if err == nil && resp != nil {
				if streamCb != nil && resp.Content != "" {
					streamCb(resp.Content)
				}
			}
		} else {
			if a.logger != nil {
				a.logger.Debug("sending blocking request to AI provider",
					"task_id", string(task.ID),
					"agent", a.Name(),
					"iteration", iterations,
				)
			}
			resp, err = a.p.Send(ctx, req)
		}

		if err != nil {
			return nil, fmt.Errorf("sdk/agent: provider communication failed: %w", err)
		}

		totalUsage.Add(resp.Usage)

		assistantMsg := provider.Message{
			Role:      provider.RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		}
		messages = append(messages, assistantMsg)

		if len(resp.ToolCalls) == 0 {
			return &agent.Result{
				TaskID:    task.ID,
				AgentName: a.Name(),
				Status:    contracts.StatusCompleted,
				Output:    resp.Content,
				Duration:  time.Since(startTime),
				Usage:     &totalUsage,
			}, nil
		}

		if a.logger != nil {
			a.logger.Info("AI requested tool execution",
				"task_id", string(task.ID),
				"agent", a.Name(),
				"tool_calls_count", len(resp.ToolCalls),
			)
		}

		// Execute requested tools in parallel while preserving exact order
		toolMsgs := make([]provider.Message, len(resp.ToolCalls))
		var wg sync.WaitGroup

		for i, tc := range resp.ToolCalls {
			wg.Add(1)
			go func(idx int, toolCall provider.ToolCall) {
				defer wg.Done()

				a.mu.RLock()
				t, ok := a.tools[toolCall.Name]
				a.mu.RUnlock()

				if !ok {
					toolMsgs[idx] = provider.NewToolResultMessage(toolCall.ID,
						fmt.Sprintf("error: tool %q is not registered for this agent", toolCall.Name))
					return
				}

				toolResult, err := a.executeToolWithRecovery(ctx, t, toolCall)
				if err != nil {
					toolMsgs[idx] = provider.NewToolResultMessage(toolCall.ID,
						fmt.Sprintf("error: tool execution failed: %v", err))
					return
				}

				var content string
				if toolResult.ExitCode != 0 {
					content = fmt.Sprintf("exit code %d: %s", toolResult.ExitCode, toolResult.Error)
				} else {
					content = toolResult.Output
				}
				toolMsgs[idx] = provider.NewToolResultMessage(toolCall.ID, content)
			}(i, tc)
		}

		wg.Wait()
		messages = append(messages, toolMsgs...)
	}
}

func (a *BaseAgent) executeToolWithRecovery(ctx context.Context, t tool.Tool, tc provider.ToolCall) (res *tool.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			err = fmt.Errorf("tool %q panicked: %v\n%s", t.Name(), r, stack)
			res = nil
		}
	}()
	return t.Execute(ctx, tc.Args)
}
```

## Rules
1. **Tool Execution Recovery**: Wrap all tool executions in deferred recovery blocks (`executeToolWithRecovery`) to prevent third-party tool crashes from crashing the agent loop.
2. **Context-Key safety**: Use private types for context keys (`contextKey`) to prevent key collisions with other modules.
3. **Execution Limit Guards**: Enforce iteration limits (`maxIterations`) on ReAct loops to prevent infinite tool loops.
4. **Deterministic Tool Result Ordering**: Execute tool calls in parallel using a `sync.WaitGroup`, but store outcomes in a pre-allocated slice using each tool's request index. This ensures the output message slice maintains a 1:1 order alignment with the requested tool call sequence, avoiding provider validation schema failures.

## Pitfalls

### Pitfall 1: Type collisions from string context keys
Always define private key types when passing callbacks in contexts to avoid namespace clashes.

### Pitfall 2: Non-deterministic tool result indexing
Running parallel tool executions writing to a raw shared message channel causes the output messages to append in arbitrary completion-time order. If tool result IDs mismatch their original request sequence, AI APIs will fail validation. Always pre-allocate slices and write to indices.

## Verify
```bash
go build ./sdk/agent/...
```

## Checklist
- [ ] File `sdk/agent/agent.go` exists
- [ ] Package: `agent`
- [ ] Context values are retrieved using private key types
- [ ] `Execute` supports both streaming and blocking modes
- [ ] Max iteration limit guards protect ReAct loops from deadlocks
- [ ] Usage records accumulate token counts across all calls
- [ ] Tool calls are executed in parallel using WaitGroups
- [ ] Tool execution order is deterministic (corresponds to request sequence)
- [ ] Tool panics are caught and converted to error result feedbacks
- [ ] `go build ./sdk/agent/...` passes
