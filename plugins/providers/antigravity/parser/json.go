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
