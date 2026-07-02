# Micro-Task 2.10: Tạo kernel/logger/redact.go

## Thông tin
- **File tạo**: `kernel/logger/redact.go`
- **Package**: `logger`
- **Dependencies trước**: 2.07
- **Thời gian**: 10 phút
- **Verify**: `go build ./kernel/logger/...`

## Mục đích
Redact (ẩn) sensitive data trong logs. API keys, passwords, tokens
KHÔNG BAO GIỜ xuất hiện trong log files.

## Nội dung CHÍNH XÁC cần tạo

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
//
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
//
// Usage:
//
//	log.Info("provider config",
//	    "provider", name,
//	    "api_key", logger.Redact("api_key", actualKey),
//	)
//	// Output: ... provider=antigravity api_key=[REDACTED]
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
//
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
//
// Used when logging entire config maps.
//
// Example:
//
//	log.Info("config loaded", "settings", logger.RedactMap(configMap))
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

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Log leak via error messages
```go
// ❌ NGUY HIỂM:
log.Error("auth failed", "error", fmt.Sprintf("invalid key: %s", apiKey))
// Log chứa API key trong error message!

// ✅ AN TOÀN:
log.Error("auth failed", "error", "invalid api key", "key_preview", logger.RedactString(apiKey))
// Log: error="invalid api key" key_preview="sk-1****cdef"
```

### Pitfall 2: Suffix matching cho compound names
`GEMINI_API_KEY` phải match vì suffix là `api_key`.
`MY_SECRET_TOKEN` phải match vì suffix là `token`.

### Pitfall 3: RedactMap creates COPY
```go
// ❌ SAI — modifies original:
for k, v := range m {
    if sensitive {
        m[k] = "[REDACTED]"  // Mutates original!
    }
}

// ✅ ĐÚNG — returns new map:
result := make(map[string]string, len(m))
```

## Checklist
- [ ] File `kernel/logger/redact.go` tồn tại
- [ ] `RedactedValue` constant = "[REDACTED]"
- [ ] `sensitiveFields` map với ≥ 12 entries
- [ ] `IsSensitiveField()` — case-insensitive, exact + suffix match
- [ ] `Redact()` — returns [REDACTED] or original value
- [ ] `RedactString()` — partial reveal (first 4 + last 4 chars)
- [ ] `RedactMap()` — returns NEW map (not mutated)
- [ ] `go build ./kernel/logger/...` không lỗi
