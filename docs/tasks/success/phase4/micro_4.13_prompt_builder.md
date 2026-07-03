# Micro-Task 4.13: Create plugins/providers/antigravity/prompt/builder.go

## Info
- **File**: `plugins/providers/antigravity/prompt/builder.go`
- **Package**: `prompt`
- **Depends on**: 4.12
- **Time**: 15 min
- **Verify**: `go build ./plugins/providers/antigravity/prompt/...`

## Purpose
Implements the query formatter (`BuildCLIPrompt` and serialization helpers) that converts standard `provider.Request` payloads into plain-text strings accepted by the Antigravity CLI.

## EXACT code to create

```go
// Package prompt handles prompt serialization for CLI process execution.
package prompt

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

// BuildCLIPrompt converts a structured Request into a single text prompt
// compatible with the Antigravity CLI input format.
//
// Format:
//   - System instructions are prepended at the top.
//   - Messages are formatted sequentially with role markers (User, Assistant, Tool).
//   - Available tools are described inside a JSON block to let the model choose them.
func BuildCLIPrompt(req *provider.Request) (string, error) {
	if req == nil {
		return "", fmt.Errorf("prompt: request cannot be nil")
	}

	var sb strings.Builder

	// 1. Append System Instructions & Tools
	if len(req.Tools) > 0 {
		sb.WriteString("System Instructions: You have access to the following tools.\n")
		sb.WriteString("To call a tool, output a JSON block matching this structure:\n")
		sb.WriteString("```json\n{\"tool\": \"tool_name\", \"args\": {\"arg_name\": \"value\"}}\n```\n\n")
		sb.WriteString("Available Tools:\n")

		toolsJSON, err := json.MarshalIndent(req.Tools, "", "  ")
		if err == nil {
			sb.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", string(toolsJSON)))
		} else {
			for _, t := range req.Tools {
				sb.WriteString(fmt.Sprintf("- %s: %s\n", t.Name, t.Description))
			}
		}
	}

	// 2. Format Conversation History
	for _, msg := range req.Messages {
		switch msg.Role {
		case provider.RoleSystem:
			sb.WriteString(fmt.Sprintf("System Instruction: %s\n\n", msg.Content))
		case provider.RoleUser:
			sb.WriteString(fmt.Sprintf("User: %s\n\n", msg.Content))
		case provider.RoleAssistant:
			if msg.Content != "" {
				sb.WriteString(fmt.Sprintf("Assistant: %s\n\n", msg.Content))
			}
			if len(msg.ToolCalls) > 0 {
				sb.WriteString("Assistant (Requested Tools):\n")
				for _, tc := range msg.ToolCalls {
					sb.WriteString(fmt.Sprintf("- call tool %q with args: %s\n", tc.Name, string(tc.Args)))
				}
				sb.WriteString("\n")
			}
		case provider.RoleTool:
			sb.WriteString(fmt.Sprintf("Tool (ID: %s) Output: %s\n\n", msg.ToolCallID, msg.Content))
		}
	}

	// 3. Append Sentinel Delimiter to mark end of input block
	sb.WriteString("\n---END-OF-PROMPT---\n")

	return sb.String(), nil
}
```

## Pitfalls

### Pitfall 1: Omitting tool schemas when formatting requests
If you only append user and assistant messages but omit `req.Tools` schemas, the AI model won't know which tools are available or how to structure call arguments. Always include tool definitions in the formatted prompt.

### Pitfall 2: Merging role prompts without delimiters
Concatenating message contents without role markers makes it difficult for CLI models to distinguish user queries from system instructions, causing confusing outputs. Use clear role headers.

## Verify
```bash
go build ./plugins/providers/antigravity/prompt/...
```

## Checklist
- [ ] File exists at `plugins/providers/antigravity/prompt/builder.go`
- [ ] Package name is `prompt`
- [ ] All exported types have Godoc
- [ ] Tool schemas are serialized to JSON blocks in instructions
- [ ] Role prefixes distinguish message sources
- [ ] Build command passes
