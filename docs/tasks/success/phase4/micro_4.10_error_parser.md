# Micro-Task 4.10: Create plugins/providers/antigravity/parser/error.go

## Info
- **File**: `plugins/providers/antigravity/parser/error.go`
- **Package**: `parser`
- **Depends on**: 4.09
- **Time**: 15 min
- **Verify**: `go build ./plugins/providers/antigravity/parser/...`

## Purpose
Implements the CLI error normalizer (`ParseError`) to map raw error output logs and CLI stream diagnostic warnings to standard orchestrator error sentinels.

## EXACT code to create

```go
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

	// 1. Rate Limit mappings
	if strings.Contains(lower, "rate limit") ||
		strings.Contains(lower, "quota exceeded") ||
		strings.Contains(lower, "too many requests") ||
		strings.Contains(lower, "429") {
		return contracts.ErrProviderRateLimited
	}

	// 2. Authentication/Credentials mappings
	if strings.Contains(lower, "api key") ||
		strings.Contains(lower, "invalid credentials") ||
		strings.Contains(lower, "unauthorized") ||
		strings.Contains(lower, "auth failed") ||
		strings.Contains(lower, "401") ||
		strings.Contains(lower, "403") {
		return contracts.ErrProviderAuthFailed
	}

	// 3. Timeout mappings
	if strings.Contains(lower, "timeout") ||
		strings.Contains(lower, "deadline exceeded") ||
		strings.Contains(lower, "gateway timeout") ||
		strings.Contains(lower, "504") {
		return contracts.ErrProviderTimeout
	}

	// 4. Availability mappings
	if strings.Contains(lower, "not found") ||
		strings.Contains(lower, "unavailable") ||
		strings.Contains(lower, "no such file") ||
		strings.Contains(lower, "command not found") ||
		strings.Contains(lower, "502") ||
		strings.Contains(lower, "503") {
		return contracts.ErrProviderUnavailable
	}

	// Default fallback wrapped raw error representation
	return fmt.Errorf("antigravity CLI error: %s", strings.TrimSpace(input))
}
```

## Pitfalls

### Pitfall 1: Returning raw errors instead of standard sentinels
```go
// WRONG:
return errors.New("rate limit reached") // Retries logic will fail because it looks for ErrProviderRateLimited.

// CORRECT:
return contracts.ErrProviderRateLimited
```
Always map errors to the standard sentinels so the orchestrator registry and scheduler engines can handle them correctly.

### Pitfall 2: Overly specific string matching
AI models or CLI wrappers can alter error strings between versions. Relying on exact string matches will cause matches to fail. Search for smaller keywords instead.

## Verify
```bash
go build ./plugins/providers/antigravity/parser/...
```

## Checklist
- [ ] File exists at `plugins/providers/antigravity/parser/error.go`
- [ ] Package name is `parser`
- [ ] All exported types have Godoc
- [ ] Maps 429 status keys to `contracts.ErrProviderRateLimited`
- [ ] Maps authorization credentials failures to `contracts.ErrProviderAuthFailed`
- [ ] Build command passes
