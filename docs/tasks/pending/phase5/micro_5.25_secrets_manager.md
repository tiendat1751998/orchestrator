# Micro-Task 5.25: Create kernel/security/secrets.go

## Info
- **File**: `kernel/security/secrets.go`
- **Package**: `security`
- **Depends on**: 5.24
- **Time**: 15 min
- **Verify**: `go build ./kernel/security/...`

## Purpose
Implements the environment secrets loader (`LoadSecret`) and slog redact filter helper (`RedactSecrets`) to redact keys from logs.

## EXACT code to create

```go
package security

import (
	"os"
	"strings"
)

// LoadSecret retrieves a secret from environment variables.
func LoadSecret(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

// RedactSecrets replaces occurrences of registered keys in target strings with redact placeholders.
func RedactSecrets(input string, keys []string) string {
	if input == "" {
		return ""
	}

	redacted := input
	for _, key := range keys {
		secret := LoadSecret(key)
		if secret != "" && len(secret) > 4 {
			redacted = strings.ReplaceAll(redacted, secret, "[REDACTED]")
		}
	}

	return redacted
}
```

## Pitfalls

### Pitfall 1: Hardcoding sensitive keys in configurations
```go
// WRONG:
const ApiKey = "sk-proj-dummysecretvaluethatshouldnotbehere" // Leaks credentials if committed to git!

// CORRECT:
apiKey := LoadSecret("GEMINI_API_KEY")
```
Commiting secrets to source control leads to security breaches. Always load secrets from environment variables.

### Pitfall 2: Redacting short strings
If an environment variable contains a short string (e.g. `"true"` or `"123"`), running `ReplaceAll` will replace all occurrences in the text, corrupting the log. Only redact secrets longer than 4 characters.

## Verify
```bash
go build ./kernel/security/...
# Expected: clean compilation without errors
```

## Checklist
- [ ] File exists at `kernel/security/secrets.go`
- [ ] Package name is `security`
- [ ] All exported types have Godoc
- [ ] Secrets are loaded from environment variables
- [ ] Log redactors prevent replacing short strings (under 5 characters)
- [ ] Build command passes
