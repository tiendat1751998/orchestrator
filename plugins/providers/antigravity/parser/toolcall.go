package parser

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
	return time.Now().UnixNano()
}
