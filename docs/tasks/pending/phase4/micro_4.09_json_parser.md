# Micro-Task 4.09: Create plugins/providers/antigravity/parser/json.go

## Info
- **File**: `plugins/providers/antigravity/parser/json.go`
- **Package**: `parser`
- **Depends on**: 4.07
- **Time**: 15 min
- **Verify**: `go build ./plugins/providers/antigravity/parser/...`

## Purpose
Implements the JSON response parser (`ParseJSON` and formatting helpers) to strip markdown wrappers from structured JSON outputs and decode them into target interface variables.

## EXACT code to create

```go
package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ParseJSON cleans raw input by stripping markdown code fences and decodes the JSON payload.
//
// Parameters:
//   - input: raw response string containing JSON.
//   - dest: pointer to target struct/map where output will be decoded.
func ParseJSON(input string, dest any) error {
	if dest == nil {
		return errors.New("parser: destination pointer cannot be nil")
	}

	cleaned := strings.TrimSpace(input)

	// 1. Strip markdown fences if present
	if strings.HasPrefix(cleaned, "```") {
		// Remove opening line (e.g. ```json or ```)
		lines := strings.Split(cleaned, "\n")
		if len(lines) >= 2 {
			// Find closing fence index
			closingIdx := -1
			for i := len(lines) - 1; i > 0; i-- {
				if strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
					closingIdx = i
					break
				}
			}

			if closingIdx != -1 {
				cleaned = strings.Join(lines[1:closingIdx], "\n")
			}
		}
	}

	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return errors.New("parser: empty JSON payload")
	}

	// 2. Decode into destination
	if err := json.Unmarshal([]byte(cleaned), dest); err != nil {
		return fmt.Errorf("parser: failed to decode JSON: %w\nPayload:\n%s", err, cleaned)
	}

	return nil
}
```

## Pitfalls

### Pitfall 1: Unmarshalling raw markdown wrapper headers
```go
// WRONG:
json.Unmarshal([]byte(input), dest) // Fails when raw string contains ```json wrappers.

// CORRECT:
// Clean markdown blocks first before passing data to JSON parser
```
AI models frequently wrap JSON outputs in markdown code blocks. Attempting to parse these wrappers directly with Go's JSON parser will return syntax errors.

### Pitfall 2: Silently ignoring empty JSON payloads
If the AI returns an empty code block or fails to output JSON, returning a nil error hides the issue. Validate payloads and report errors if they are empty.

## Verify
```bash
go build ./plugins/providers/antigravity/parser/...
```

## Checklist
- [ ] File exists at `plugins/providers/antigravity/parser/json.go`
- [ ] Package name is `parser`
- [ ] All exported types have Godoc
- [ ] `ParseJSON` strips ` ```json ` and ` ``` ` markdown code block headers
- [ ] Destination pointer parameters are validated for nil values
- [ ] Build command passes
