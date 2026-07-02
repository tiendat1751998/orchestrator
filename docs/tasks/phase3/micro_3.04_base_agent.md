# Micro-Task 3.04: Create sdk/agent/agent.go

## Info
- **File**: `sdk/agent/agent.go`
- **Package**: `agent`
- **Depends on**: 3.01 (base_plugin.md), 3.02 (manifest_loader.md), 3.03 (prompt_builder.md), 1.12 (provider contract), 1.15 (tool contract), 1.21 (agent contract)
- **Time**: 25 min
- **Verify**: `go build ./sdk/agent/...`

## Purpose
Implements the core `BaseAgent` struct. It handles the complete agent reasoning loop (ReAct loop): prompt construction, provider calls, parallel tool execution, error-correction feeding (giving AI feedback when tools fail), maximum iterations protection (to avoid loops), and context cancellation checks.

## EXACT code to create

```go
package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts"
	"github.com/tiendat1751998/orchestrator/contracts/agent"
	contractsplugin "github.com/tiendat1751998/orchestrator/contracts/plugin"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
	"github.com/tiendat1751998/orchestrator/contracts/tool"
	sdkplugin "github.com/tiendat1751998/orchestrator/sdk/plugin"
)

// BaseAgent implements the contracts/agent.Agent interface.
// It integrates the BasePlugin for lifecycle and implements the ReAct loop.
type BaseAgent struct {
	*sdkplugin.BasePlugin

	manifest *agent.Manifest
	p        provider.Provider
	logger   *slog.Logger

	mu    sync.RWMutex
	tools map[string]tool.Tool

	// Configuration
	maxIterations int // Limit to prevent infinite loop of tool calls
}

// NewBaseAgent constructs a BaseAgent.
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
		maxIterations: 10, // Safeguard limit
	}
}

// Role returns the agent's role description.
func (a *BaseAgent) Role() string {
	return a.manifest.Role
}

// Capabilities returns the agent's capabilities list.
func (a *BaseAgent) Capabilities() []agent.Capability {
	return a.manifest.Capabilities
}

// Manifest returns the read-only copy of the manifest.
func (a *BaseAgent) Manifest() agent.Manifest {
	return *a.manifest
}

// CanHandle checks if the agent can handle the task type.
// Matches task.Type with capabilities defined in manifest.
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

// RegisterTool attaches a tool instance that the agent is permitted to execute.
func (a *BaseAgent) RegisterTool(t tool.Tool) error {
	if t == nil {
		return fmt.Errorf("sdk/agent: cannot register nil tool")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tools[t.Name()] = t
	return nil
}

// Execute performs the task by running the ReAct (Reasoning + Action) loop.
func (a *BaseAgent) Execute(ctx context.Context, task *agent.Task) (*agent.Result, error) {
	if !a.IsStarted() {
		return nil, fmt.Errorf("sdk/agent: agent %q is not running", a.Name())
	}

	// 1. Construct prompt messages
	messages := BuildPrompt(a.manifest, task, a.logger)

	// 2. Resolve Tool Definitions
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

	var totalUsage provider.TokenUsage

	for {
		// Stop if context cancellation or timeout happened
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		iterations++
		if iterations > a.maxIterations {
			return nil, fmt.Errorf("sdk/agent: agent %q exceeded max tool iterations limit (%d)", 
				a.Name(), a.maxIterations)
		}

		// 3. Construct Request to AI Provider
		req := &provider.Request{
			Model:          a.manifest.Model,
			Messages:       messages,
			Tools:          toolDefs,
			Temperature:    &a.manifest.Temperature,
			MaxTokens:      &a.manifest.MaxTokens,
			ResponseFormat: "text",
		}

		a.logger.Debug("sending request to AI provider",
			"task_id", string(task.ID),
			"agent", a.Name(),
			"iteration", iterations,
			"messages_count", len(messages),
		)

		resp, err := a.p.Send(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("sdk/agent: provider communication failed: %w", err)
		}

		// Accumulate usage metrics
		totalUsage.PromptTokens += resp.Usage.PromptTokens
		totalUsage.CompletionTokens += resp.Usage.CompletionTokens
		totalUsage.TotalTokens += resp.Usage.TotalTokens

		// Save AI assistant's response to history
		assistantMsg := provider.Message{
			Role:      provider.RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		}
		messages = append(messages, assistantMsg)

		// 4. Check if AI requested Tool Calls
		if len(resp.ToolCalls) == 0 {
			// No tools requested, loop ends. Task is finished.
			return &agent.Result{
				TaskID:   task.ID,
				Status:   contracts.StatusCompleted,
				Output:   resp.Content,
				Duration: time.Since(startTime),
				Usage: &agent.Usage{
					PromptTokens:     totalUsage.PromptTokens,
					CompletionTokens: totalUsage.CompletionTokens,
					TotalTokens:      totalUsage.TotalTokens,
				},
			}, nil
		}

		a.logger.Info("AI requested tool execution",
			"task_id", string(task.ID),
			"agent", a.Name(),
			"tool_calls_count", len(resp.ToolCalls),
		)

		// 5. Execute requested tools in parallel
		toolMsgChan := make(chan provider.Message, len(resp.ToolCalls))
		var wg sync.WaitGroup

		for _, tc := range resp.ToolCalls {
			wg.Add(1)
			go func(toolCall provider.ToolCall) {
				defer wg.Done()
				
				a.mu.RLock()
				t, ok := a.tools[toolCall.Name]
				a.mu.RUnlock()

				if !ok {
					// Tool not registered: report failure to AI so it can self-correct
					toolMsgChan <- provider.NewToolResultMessage(toolCall.ID, 
						fmt.Sprintf("error: tool %q is not registered for this agent", toolCall.Name))
					return
				}

				// Execute tool with recovery wrapper
				toolResult, err := a.executeToolWithRecovery(ctx, t, toolCall)
				if err != nil {
					// System level error
					toolMsgChan <- provider.NewToolResultMessage(toolCall.ID, 
						fmt.Sprintf("error: tool execution failed: %v", err))
					return
				}

				// Tool response message
				var content string
				if toolResult.ExitCode != 0 {
					content = fmt.Sprintf("exit code %d: %s", toolResult.ExitCode, toolResult.Error)
				} else {
					content = toolResult.Output
				}
				toolMsgChan <- provider.NewToolResultMessage(toolCall.ID, content)
			}(tc)
		}

		wg.Wait()
		close(toolMsgChan)

		// Append all tool outputs to message history
		for toolMsg := range toolMsgChan {
			messages = append(messages, toolMsg)
		}
	}
}

// executeToolWithRecovery runs a tool safely inside a defer-recover block.
func (a *BaseAgent) executeToolWithRecovery(ctx context.Context, t tool.Tool, tc provider.ToolCall) (res *tool.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("tool %q panicked: %v", t.Name(), r)
			res = nil
		}
	}()
	return t.Execute(ctx, tc.Args)
}
```

## ⚠️ Pitfalls

### Pitfall 1: No Max Iterations Guard
```go
// ❌ WRONG:
for {
    resp, _ := provider.Send(...)
    if len(resp.ToolCalls) == 0 { break }
    // execute tools
} // Infinite reasoning loop if the AI keeps requesting the same failing tool.
```
LLMs can get stuck in repetitive loops calling failing tools. A hard iteration limit (`maxIterations = 10`) protects billing and prevents process hangs.

### Pitfall 2: Locking mutex during I/O provider.Send calls
```go
// ❌ WRONG:
a.mu.Lock()
resp, err := a.p.Send(ctx, req) // Holds lock during slow HTTP network call -> blocks all other threads!
a.mu.Unlock()
```
Never hold locks during slow context operations or external I/O boundaries. Access tools map under a brief read lock (`RLock`), copy the references, and release it before running calls.

## Verify
```bash
go build ./sdk/agent/...
```

## Checklist
- [ ] File `sdk/agent/agent.go` exists
- [ ] Package: `agent`
- [ ] `BaseAgent` embeds `*sdkplugin.BasePlugin`
- [ ] ReAct reasoning loop handles parallel tool execution using `sync.WaitGroup`
- [ ] Safeguard `maxIterations` prevents infinite agent loops
- [ ] Tool execution panics are caught using `executeToolWithRecovery`
- [ ] Transient tool failures (ExitCode != 0) are formatted back to RoleTool messages for self-correction
- [ ] Token usage metrics are aggregated accurately
- [ ] `go build ./sdk/agent/...` passes
