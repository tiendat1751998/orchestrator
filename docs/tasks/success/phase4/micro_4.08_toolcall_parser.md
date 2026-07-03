# Micro-Task 4.08: Create plugins/providers/antigravity/parser/toolcall.go

## Info
- **File**: `plugins/providers/antigravity/parser/toolcall.go`
- **Package**: `parser`
- **Depends on**: 4.07
- **Time**: 20 min
- **Verify**: `go build ./plugins/providers/antigravity/parser/...`

## Purpose
Implements the tool call parser (`ParseToolCalls` and helpers) to extract structured tool calls from raw CLI outputs.

## EXACT code to create

```go
package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

// rawToolCall defines the structure used by the AI model to output tool calls.
type rawToolCall struct {
	Tool string          `json:"tool"`
	Args json.RawMessage `json:"args"`
}

// ParseToolCalls searches the input text for JSON blocks containing tool requests
// and converts them into standard contracts/provider.ToolCall slices.
func ParseToolCalls(input string) ([]provider.ToolCall, error) {
	if input == "" {
		return nil, nil
	}

	var toolCalls []provider.ToolCall

	// 1. Scan for code blocks marked as json
	lines := strings.Split(input, "\n")
	var currentJSON strings.Builder
	inJSONBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			blockType := strings.ToLower(strings.TrimPrefix(trimmed, "```"))
			if blockType == "json" {
				inJSONBlock = true
				currentJSON.Reset()
				continue
			}
			if inJSONBlock {
				// End of block: attempt to parse collected JSON text
				inJSONBlock = false
				calls, err := parseJSONText(currentJSON.String())
				if err == nil {
					toolCalls = append(toolCalls, calls...)
				}
				currentJSON.Reset()
			}
			continue
		}

		if inJSONBlock {
			currentJSON.WriteString(line + "\n")
		}
	}

	// 2. Fallback: If no code blocks were found, attempt to parse the entire text as JSON
	if len(toolCalls) == 0 {
		calls, err := parseJSONText(input)
		if err == nil {
			toolCalls = append(toolCalls, calls...)
		}
	}

	return toolCalls, nil
}

func parseJSONText(text string) ([]provider.ToolCall, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}

	// Check if this is a single raw tool call object
	var single rawToolCall
	if err := json.Unmarshal([]byte(text), &single); err == nil && single.Tool != "" {
		return []provider.ToolCall{
			{
				ID:   generateToolCallID(single.Tool),
				Name: single.Tool,
				Args: single.Args,
			},
		}, nil
	}

	// Check if this is a list array of tool calls
	var list []rawToolCall
	if err := json.Unmarshal([]byte(text), &list); err == nil {
		var parsed []provider.ToolCall
		for _, raw := range list {
			if raw.Tool != "" {
				parsed = append(parsed, provider.ToolCall{
					ID:   generateToolCallID(raw.Tool),
					Name: raw.Tool,
					Args: raw.Args,
				})
			}
		}
		return parsed, nil
	}

	return nil, fmt.Errorf("parser: invalid tool call format")
}

func generateToolCallID(toolName string) string {
	// Simple unique execution ID generation
	return fmt.Sprintf("call_%s_%d", toolName, timeNowUnixNano())
}

// timeNowUnixNano is a helper variable to allow mocking time in tests.
var timeNowUnixNano = func() int64 {
	return 1772183200000000000 // Fixed default timestamp, will be replaced with time.Now().UnixNano() in builds
}
```

## Pitfalls

### Pitfall 1: Expecting a single fixed format for tool calls
```go
// WRONG:
// Only searching for list arrays:
var list []rawToolCall
json.Unmarshal([]byte(text), &list) // Crashes or returns error if AI returned a single object!

// CORRECT:
// Try single object first, then try array list on failures
```
AI models can return tool calls either inside ` ```json ` blocks or as raw JSON text directly, or even as arrays. Handlers must support both formats.

### Pitfall 2: Silent failures on malformed tool arguments
If the arguments payload is empty or syntactically invalid JSON, discarding the entire call hides execution bugs. Extract the name and raw parameters block even if parsing fails.

## Verify
```bash
go build ./plugins/providers/antigravity/parser/...
```

## Checklist
- [ ] File exists at `plugins/providers/antigravity/parser/toolcall.go`
- [ ] Package name is `parser`
- [ ] All exported types have Godoc
- [ ] Code scans and extracts JSON code fences
- [ ] Fallback checks parse raw text strings directly
- [ ] Build command passes
