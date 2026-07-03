# Micro-Task 2.10: Create kernel/logger/redact.go

## Info
- **File to create**: `kernel/logger/redact.go`
- **File to update**: `kernel/logger/logger.go` (Update replaceAttr implementation)
- **Package**: `logger`
- **Depends on**: 2.07 (logger.go)
- **Time**: 10 min
- **Verify**: `go build ./kernel/logger/...`

## Purpose
Implements log redaction helpers (`IsSensitiveField`, `Redact`, `RedactString`, `RedactMap`) and updates the `slog` output hook to block sensitive keys (like API keys, passwords, or session tokens) from leaking into log files.

## EXACT code to create

### Part 1: Create `kernel/logger/redact.go`

```go
package logger

import "strings"

// RedactedValue is a placeholder shown instead of sensitive data.
const RedactedValue = "[REDACTED]"

// sensitiveFields are field names whose values should be redacted.
//
// WHY lowercase map keys?
// → Field names come from various sources (config, user code).
// → "API_KEY", "api_key", "ApiKey" should all be redacted.
// → Normalize to lowercase before checking.
var sensitiveFields = map[string]bool{
	"api_key":       true,
	"apikey":        true,
	"api-key":       true,
	"secret":        true,
	"password":      true,
	"token":         true,
	"access_token":  true,
	"refresh_token": true,
	"authorization": true,
	"private_key":   true,
	"secret_key":    true,
	"credentials":   true,
}

// IsSensitiveField checks if a field name should be redacted.
//
// Case-insensitive matching. Checks both exact match and suffix match.
//
// Examples:
//	IsSensitiveField("api_key")         → true (exact match)
//	IsSensitiveField("GEMINI_API_KEY")  → true (suffix match: "api_key")
//	IsSensitiveField("provider")        → false
func IsSensitiveField(fieldName string) bool {
	lower := strings.ToLower(fieldName)

	// Exact match
	if sensitiveFields[lower] {
		return true
	}

	// Suffix match: "gemini_api_key" contains "api_key"
	for sensitive := range sensitiveFields {
		if strings.HasSuffix(lower, sensitive) {
			return true
		}
		// Also check with underscore prefix: "_api_key"
		if strings.HasSuffix(lower, "_"+sensitive) {
			return true
		}
	}

	return false
}

// Redact replaces a value with [REDACTED] if the field name is sensitive.
func Redact(fieldName string, value any) any {
	if IsSensitiveField(fieldName) {
		return RedactedValue
	}
	return value
}

// RedactString redacts a string value.
//
// Shows first 4 and last 4 characters for debugging, rest is masked.
// If string is too short (< 12 chars), fully redacted.
//
// Examples:
//	RedactString("sk-1234567890abcdef")  → "sk-1****cdef"
//	RedactString("short")               → "[REDACTED]"
//	RedactString("")                     → "[REDACTED]"
func RedactString(value string) string {
	if len(value) < 12 {
		return RedactedValue
	}
	return value[:4] + "****" + value[len(value)-4:]
}

// RedactMap creates a copy of a map with sensitive values redacted.
func RedactMap(m map[string]string) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		if IsSensitiveField(k) {
			result[k] = RedactedValue
		} else {
			result[k] = v
		}
	}
	return result
}
```

---

### Part 2: Update `kernel/logger/logger.go`

In [kernel/logger/logger.go](file:///d:/project/orchestrator/kernel/logger/logger.go), replace the placeholder `replaceAttr` helper function with the active redaction hook:

```go
// replaceAttr is the slog attributes filter callback used to redact secrets.
func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if IsSensitiveField(a.Key) {
		return slog.String(a.Key, RedactedValue)
	}
	return a
}
```

## Rules
1. **Case-Insensitive Match**: Check target keys in lowercase to ensure variants like `ApiKey` and `api_key` are matched.
2. **Suffix Match**: Support suffix checks to catch keys with env prefixes (such as `GEMINI_API_KEY`).
3. **No In-Place Mutation**: When redacting maps (`RedactMap`), copy properties into a new map to prevent mutating the caller's memory variables.

## ⚠️ Pitfalls

### Pitfall 1: Mutating maps in-place inside log routines
```go
result := make(map[string]string, len(m))
for k, v := range m {
    if IsSensitiveField(k) {
        result[k] = RedactedValue
    } else {
        result[k] = v
    }
}
```
Always allocate fresh target maps when redacting to avoid side effects on application memory.

### Pitfall 2: Revealing short keys in partial redactions
If an API key is very short (e.g. less than 12 characters), attempting to show the first 4 and last 4 characters reveals the entire key. If the key length is short, redact the value fully.

## Verify
```bash
go build ./kernel/logger/...
```

## Checklist
- [ ] File `kernel/logger/redact.go` exists
- [ ] Package: `logger`
- [ ] `RedactedValue` constant defined as `[REDACTED]`
- [ ] Suffix matching matches compound keys (e.g. `GEMINI_API_KEY`)
- [ ] `RedactString` masks short values (< 12 chars) completely
- [ ] `RedactMap` returns a new copied map structure
- [ ] `replaceAttr` hook implemented inside `logger.go`
- [ ] `go build ./kernel/logger/...` passes
