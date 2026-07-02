# Micro-Task 3.03: Create sdk/agent/prompt.go

## Info
- **File**: `sdk/agent/prompt.go`
- **Package**: `agent`
- **Depends on**: 1.08 (message.go contract), 1.18 (task.go contract), 1.20 (manifest.go contract)
- **Time**: 20 min
- **Verify**: `go build ./sdk/agent/...`

## Purpose
Implements the prompt builder (`BuildPrompt`) for agents. It converts an `agent.Task` into a slice of `provider.Message` structs suitable for sending to LLM providers. It structures task descriptions, inputs, and context items safely, while estimating tokens to log warnings if the payload size risks exceeding context limits.

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
		// Encode inputs to pretty JSON
		inputsJSON, err := json.MarshalIndent(task.Input, "", "  ")
		if err == nil {
			header += fmt.Sprintf("```json\n%s\n```\n", string(inputsJSON))
		} else {
			// Fallback plain key-value format
			for k, v := range task.Input {
				header += fmt.Sprintf("- %s: %v\n", k, v)
			}
		}
	}

	return header
}
```

## ⚠️ Pitfalls

### Pitfall 1: Packing all context into a single user message
```go
// ❌ WRONG:
// Ghép toàn bộ File content + Task description + System Prompt thành một khối User Message duy nhất.
// Làm LLM khó phân biệt đâu là chỉ dẫn làm việc (Instruction) và đâu là dữ liệu đầu vào (Data).

// ✅ CORRECT:
// Tách biệt rõ ràng: System Prompt gửi qua RoleSystem. 
// Task instruction và input gửi qua một User Message đầu tiên.
// Mỗi Context Item gửi thành một User Message riêng biệt kèm tiêu đề đánh dấu loại context rõ ràng.
```
LLMs chú ý tốt hơn khi dữ liệu đầu vào được tổ chức thành cấu trúc rõ ràng với nhãn phân biệt.

### Pitfall 2: Bỏ qua kiểm tra tràn giới hạn Tokens (Context Window Overflow)
Hệ thống không cảnh báo khi payload quá lớn sẽ làm nhà cung cấp (provider) trả về lỗi API. Bước ước lượng thô (`estimatedTokens := total_chars / 4`) cung cấp cảnh báo chủ động để logger ghi nhận hành vi quá tải trước khi trigger API call.

## Verify
```bash
go build ./sdk/agent/...
```

## Checklist
- [ ] File `sdk/agent/prompt.go` tồn tại
- [ ] Package: `agent`
- [ ] `BuildPrompt` tạo ra RoleSystem message chứa SystemPrompt
- [ ] `BuildPrompt` tạo ra RoleUser message chứa Task Details
- [ ] Thêm các Context items thành các messages phân tách rõ ràng
- [ ] `formatTaskHeader` tự động format dữ liệu dạng map sang JSON định dạng đẹp
- [ ] Tích hợp cảnh báo log khi ước tính dung lượng token vượt cấu hình `MaxTokens` của manifest
- [ ] `go build ./sdk/agent/...` chạy thành công
