# Micro-Task 2.03: Tạo kernel/config/env.go

## Thông tin
- **File tạo**: `kernel/config/env.go`
- **Package**: `config`
- **Dependencies trước**: 2.01
- **Thời gian**: 15 phút
- **Verify**: `go build ./kernel/config/...`

## Mục đích
Resolve `${ENV_VAR_NAME}` placeholders trong YAML values.
Ví dụ: `api_key: "${GEMINI_API_KEY}"` → thay bằng giá trị env var thực.

## Nội dung CHÍNH XÁC cần tạo

```go
package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// envPattern matches ${ENV_VAR_NAME} patterns.
//
// Regex breakdown:
//   \$\{       — literal "${" 
//   ([A-Z_][A-Z0-9_]*)  — capture group: env var name
//                          Must start with letter or underscore
//                          Followed by letters, digits, or underscores
//   \}         — literal "}"
//
// Examples that MATCH:
//   ${GEMINI_API_KEY}  → captures "GEMINI_API_KEY"
//   ${HOME}            → captures "HOME"
//   ${MY_VAR_123}      → captures "MY_VAR_123"
//
// Examples that DON'T match:
//   $HOME              → no braces (we require explicit ${} syntax)
//   ${123_VAR}         → starts with digit
//   ${my-var}          → contains hyphen
var envPattern = regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}`)

// ResolveEnvVars replaces all ${ENV_VAR} placeholders in a string with
// their actual values from the environment.
//
// If an env var is not set:
//   - If the placeholder is the ENTIRE value (e.g., api_key: "${MISSING_VAR}"),
//     return an error. This means the user intended to use an env var but forgot to set it.
//   - If the placeholder is PART of a value (e.g., path: "/home/${USER}/data"),
//     replace with empty string and log a warning.
//
// Parameters:
//   - input: the string containing ${...} placeholders
//
// Returns:
//   - resolved: the string with placeholders replaced by env var values
//   - err: if a required env var is missing
func ResolveEnvVars(input string) (string, error) {
	var missingVars []string

	result := envPattern.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name from ${VAR_NAME}
		varName := envPattern.FindStringSubmatch(match)[1]

		value, exists := os.LookupEnv(varName)
		if !exists {
			missingVars = append(missingVars, varName)
			return "" // Replace with empty string for now
		}
		return value
	})

	// If the entire input was a single env var reference and it's missing → error
	if len(missingVars) > 0 {
		trimmed := strings.TrimSpace(input)
		isSingleVar := envPattern.ReplaceAllString(trimmed, "") == ""
		if isSingleVar {
			return "", fmt.Errorf("required environment variable(s) not set: %s",
				strings.Join(missingVars, ", "))
		}
		// For partial env vars, return with empty replacements (best effort)
		// Caller should log a warning about the missing vars
	}

	return result, nil
}

// ResolveEnvInMap recursively resolves env vars in a map[string]any structure.
//
// This is used to process the entire parsed YAML tree before converting
// to typed Config struct.
//
// Handles nested types:
//   - string → ResolveEnvVars
//   - map[string]any → recurse into each value
//   - []any → recurse into each element
//   - other types (int, bool, float64) → keep as-is (no env vars in non-strings)
//
// WHY process the raw YAML map instead of the typed Config struct?
// → Config struct has many fields. Processing each field individually = verbose.
// → Processing the raw map handles ALL string fields automatically.
// → New fields added to Config are automatically handled.
func ResolveEnvInMap(data map[string]any) error {
	for key, value := range data {
		resolved, err := resolveValue(value)
		if err != nil {
			return fmt.Errorf("config key %q: %w", key, err)
		}
		data[key] = resolved
	}
	return nil
}

// resolveValue recursively resolves env vars in any value type.
func resolveValue(value any) (any, error) {
	switch v := value.(type) {
	case string:
		return ResolveEnvVars(v)

	case map[string]any:
		for key, val := range v {
			resolved, err := resolveValue(val)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key, err)
			}
			v[key] = resolved
		}
		return v, nil

	case []any:
		for i, item := range v {
			resolved, err := resolveValue(item)
			if err != nil {
				return nil, fmt.Errorf("[%d]: %w", i, err)
			}
			v[i] = resolved
		}
		return v, nil

	default:
		// int, float64, bool, nil — no env vars to resolve
		return value, nil
	}
}

// ExpandEnvWithDefault resolves ${VAR:-default} syntax.
// If VAR is not set, returns the default value.
//
// This is a future enhancement. For now, only ${VAR} is supported.
// Uncomment and implement when needed.
//
// func ExpandEnvWithDefault(input string) string { ... }
```

## ⚠️ Pitfalls cần tránh

### Pitfall 1: Regex phải strict
```go
// ❌ SAI — quá loose:
var envPattern = regexp.MustCompile(`\$\{(.+?)\}`)
// Matches ${123}, ${a-b}, ${} → invalid var names

// ✅ ĐÚNG — strict:
var envPattern = regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}`)
// Only matches valid env var names: start with letter/underscore, then letters/digits/underscores
```

### Pitfall 2: os.LookupEnv vs os.Getenv
```go
// ❌ SAI — os.Getenv:
value := os.Getenv("MY_VAR")
// Returns "" for both "not set" AND "set to empty string"
// Can't distinguish between them

// ✅ ĐÚNG — os.LookupEnv:
value, exists := os.LookupEnv("MY_VAR")
// exists=false → not set
// exists=true, value="" → set to empty string (user intended this)
```

### Pitfall 3: Error vs warning for missing vars
- `api_key: "${MISSING_KEY}"` → **error** (entire value is env var → user intended it)
- `path: "/home/${USER}/data"` → **warning** (partial → best effort replacement)

### Pitfall 4: YAML parsed types
`yaml.Unmarshal` converts YAML to `map[string]any`:
- YAML strings → Go `string`
- YAML numbers → Go `int` or `float64`
- YAML booleans → Go `bool`
- YAML lists → Go `[]any`
- YAML maps → Go `map[string]any`

`resolveValue()` PHẢI handle tất cả types. Non-string types → return as-is.

### Pitfall 5: Error wrapping cho debugging
```go
return nil, fmt.Errorf("config key %q: %w", key, err)
```
Nested error messages: `"config key "providers": antigravity: api_key: required environment variable(s) not set: GEMINI_API_KEY"`
→ User biết CHÍNH XÁC field nào lỗi.

## Checklist
- [ ] File `kernel/config/env.go` tồn tại
- [ ] Package: `package config`
- [ ] `envPattern` regex compile with strict pattern
- [ ] `ResolveEnvVars()` xử lý `${VAR}` syntax
- [ ] `ResolveEnvInMap()` xử lý recursive map/slice/string
- [ ] `resolveValue()` helper cho recursive descent
- [ ] Dùng `os.LookupEnv()` (KHÔNG `os.Getenv()`)
- [ ] Error cho missing required env vars
- [ ] Handles nested maps, slices, strings, và non-string types
- [ ] Error messages chỉ rõ config key path
- [ ] Godoc comments với examples
- [ ] `go build ./kernel/config/...` không lỗi
