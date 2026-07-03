package parser

import (
	"fmt"
	"strings"

	"github.com/tiendat1751998/orchestrator/contracts"
)

// ParseError scans the CLI stderr or error response text for keywords
// and maps them to standard contracts error variables.
func ParseError(input string) error {
	if input == "" {
		return nil
	}

	lower := strings.ToLower(input)

	// ponytail: helper checks to keep cyclomatic complexity below the budget limit of 10
	containsAny := func(s string, keywords ...string) bool {
		for _, kw := range keywords {
			if strings.Contains(s, kw) {
				return true
			}
		}
		return false
	}

	// 1. Rate Limit mappings
	if containsAny(lower, "rate limit", "quota exceeded", "too many requests", "429") {
		return contracts.ErrProviderRateLimited
	}

	// 2. Authentication/Credentials mappings
	if containsAny(lower, "api key", "invalid credentials", "unauthorized", "auth failed", "401", "403") {
		return contracts.ErrProviderAuthFailed
	}

	// 3. Timeout mappings
	if containsAny(lower, "timeout", "deadline exceeded", "gateway timeout", "504") {
		return contracts.ErrProviderTimeout
	}

	// 4. Availability mappings
	if containsAny(lower, "not found", "unavailable", "no such file", "command not found", "502", "503") {
		return contracts.ErrProviderUnavailable
	}

	// Default fallback wrapped raw error representation
	return fmt.Errorf("antigravity CLI error: %s", strings.TrimSpace(input))
}
