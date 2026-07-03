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
//  1. System Message (RoleSystem) containing the agent's configured SystemPrompt.
//  2. User Message (RoleUser) describing the task itself, parameters, and input details.
//  3. Subsequent User Messages (RoleUser) formatting the attached task ContextItems.
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
