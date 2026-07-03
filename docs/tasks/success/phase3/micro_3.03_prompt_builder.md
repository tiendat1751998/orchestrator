# Micro-Task 3.03: Create sdk/agent/prompt.go

## Info
- **File**: `sdk/agent/prompt.go`
- **Package**: `agent`
- **Depends on**: 1.08 (message.go contract), 1.18 (task.go contract), 1.20 (manifest.go contract)
- **Time**: 20 min
- **Verify**: `go build ./sdk/agent/...`

## Purpose
Implements the prompt builder (`BuildPrompt` and formatting helpers) for agents. It converts an `agent.Task` into a slice of `provider.Message` structs suitable for sending to LLM providers. It structures task descriptions, inputs, and context items safely, while estimating tokens to log warnings if the payload size risks exceeding context limits.

## EXACT code to create

```go
package agent

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/tiendat1751998/orchestrator/contracts/agent"
	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

// BuildPrompt constructs the message history for the AI provider request.
//
// Prompt Structure:
//   1. System Message (RoleSystem) containing the agent's configured SystemPrompt.
//   2. User Message (RoleUser) describing the task itself, parameters, and input details.
//   3. Subsequent User Messages (RoleUser) formatting the attached task ContextItems.
func BuildPrompt(manifest *agent.Manifest, task *agent.Task, logger *slog.Logger) []provider.Message {
	messages := make([]provider.Message, 0)

	// 1. System Prompt (if defined)
	if manifest.SystemPrompt != "" {
		messages = append(messages, provider.NewSystemMessage(manifest.SystemPrompt))
	}

	// 2. Format Task Details
	taskDetails := formatTaskHeader(task)
	messages = append(messages, provider.NewUserMessage(taskDetails))

	// 3. Format Context Items (e.g. workspace files, previous outputs)
	totalContextChars := 0
	for _, ctxItem := range task.Context {
		if ctxItem.Content == "" {
			continue
		}

		ctxPrompt := fmt.Sprintf("### Context Item (%s) from %q:\n%s\n---",
			ctxItem.Type, ctxItem.Source, ctxItem.Content)

		totalContextChars += len(ctxPrompt)
		messages = append(messages, provider.NewUserMessage(ctxPrompt))
	}

	// 4. Observability Token Check (1 token ≈ 4 characters estimation)
	estimatedTokens := (len(manifest.SystemPrompt) + len(taskDetails) + totalContextChars) / 4
	if manifest.MaxTokens > 0 && estimatedTokens > manifest.MaxTokens {
		if logger != nil {
			logger.Warn("prompt size estimation exceeds agent max_tokens limit",
				"task_id", string(task.ID),
				"estimated_tokens", estimatedTokens,
				"max_tokens_limit", manifest.MaxTokens,
			)
		}
	}

	return messages
}

// formatTaskHeader formats task name, description, and inputs into a single user message.
func formatTaskHeader(task *agent.Task) string {
	header := fmt.Sprintf("You have been assigned the following task:\n"+
		"Task ID: %s\n"+
		"Task Name: %s\n"+
		"Task Type: %s\n\n"+
		"Description:\n%s\n\n",
		string(task.ID), task.Name, task.Type, task.Description)

	if len(task.Input) > 0 {
		header += "Input Parameters:\n"
		inputsJSON, err := json.MarshalIndent(task.Input, "", "  ")
		if err == nil {
			header += fmt.Sprintf("```json\n%s\n```\n", string(inputsJSON))
		} else {
			for k, v := range task.Input {
				header += fmt.Sprintf("- %s: %v\n", k, v)
			}
		}
	}

	return header
}
```

## Rules
1. **Isolated Message Sections**: Separate context prompts and system directives into distinct system/user messages. Do not concatenate all details into a single prompt block.
2. **Context Window Overload Guards**: Estimate token usage (using a 1-to-4 character ratio) and log warning messages when inputs exceed manifest token limits.
3. **Pretty JSON Formatting**: Always format input maps to indent-spaced JSON within markdown blocks inside formatted headers.

## ⚠️ Pitfalls

### Pitfall 1: Merging instructions and context details into a single User Message
Concatenating instruction prompts, parameter configurations, and target context files into a single text block makes it difficult for LLMs to distinguish instructions from input data. Separate them into different roles.

### Pitfall 2: Neglecting token estimation warnings
Failing to inspect payload sizes before triggering provider APIs causes silent failures or API context overflow errors. Log warnings early.

## Verify
```bash
go build ./sdk/agent/...
```

## Checklist
- [ ] File `sdk/agent/prompt.go` exists
- [ ] Package: `agent`
- [ ] `BuildPrompt` maps `SystemPrompt` to `RoleSystem` messages
- [ ] `BuildPrompt` formats task headers as `RoleUser` messages
- [ ] Context items are appended as separate messages with clear source identifiers
- [ ] Task headers pretty-print JSON parameters
- [ ] Token estimates are verified against `MaxTokens` limits, logging warnings on overflows
- [ ] `go build ./sdk/agent/...` passes
