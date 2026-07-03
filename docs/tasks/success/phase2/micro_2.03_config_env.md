# Micro-Task 2.03: Create kernel/config/env.go

## Info
- **File**: `kernel/config/env.go`
- **Package**: `config`
- **Depends on**: 2.01
- **Time**: 15 min
- **Verify**: `go build ./kernel/config/...`

## Purpose
Resolves `${ENV_VAR_NAME}` placeholders embedded within configuration values. For example, replacing `api_key: "${GEMINI_API_KEY}"` with the actual value retrieved from the operating system environment.

## EXACT code to create

```go
package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var envPattern = regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}`)

// ResolveEnvVars replaces all ${ENV_VAR} placeholders in a string with
// their actual values from the environment.
func ResolveEnvVars(input string) (string, error) {
	var missingVars []string

	result := envPattern.ReplaceAllStringFunc(input, func(match string) string {
		varName := envPattern.FindStringSubmatch(match)[1]

		value, exists := os.LookupEnv(varName)
		if !exists {
			missingVars = append(missingVars, varName)
			return ""
		}
		return value
	})

	if len(missingVars) > 0 {
		trimmed := strings.TrimSpace(input)
		isSingleVar := envPattern.ReplaceAllString(trimmed, "") == ""
		if isSingleVar {
			return "", fmt.Errorf("required environment variable(s) not set: %s",
				strings.Join(missingVars, ", "))
		}
	}

	return result, nil
}

// ResolveEnvInMap recursively resolves env vars in a map[string]any structure.
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
		return value, nil
	}
}
```

## Rules
1. **Strict Regex Filtering**: The pattern must match valid environment variable name structures (starting with a letter or underscore, followed by digits/letters/underscores). Block invalid formats (like `${my-var}`).
2. **Lookup Safety Checks**: Use `os.LookupEnv()` rather than `os.Getenv()` to correctly identify when variables are set to empty strings.
3. **Recursive Resolution**: Process raw configuration map structures recursively (`ResolveEnvInMap`) to parse nested string fields.

## Pitfalls

### Pitfall 1: Using loose matching regex patterns
```go
// WRONG:
var envPattern = regexp.MustCompile(`\$\{(.+?)\}`) // Matches bad identifiers like ${123_VAR} or ${}.

// CORRECT:
var envPattern = regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}`)
```
Strict regex patterns protect from trying to parse malformed YAML fields as system envar references.

### Pitfall 2: Treating empty env values as unset variables
```go
// WRONG:
value := os.Getenv("MY_VAR") // Returns "" for both missing and empty variables, making it impossible to check if it's set.

// CORRECT:
value, exists := os.LookupEnv("MY_VAR") // exists identifies when variables are set to empty values.
```
Use `os.LookupEnv` to support empty env strings as valid values.

## Verify
```bash
go build ./kernel/config/...
```

## Checklist
- [ ] File `kernel/config/env.go` exists
- [ ] Package: `config`
- [ ] Strict regex `envPattern` defined
- [ ] `ResolveEnvVars` replaces variable matches or errors on missing values
- [ ] `ResolveEnvInMap` walks nested structures recursively
- [ ] Resolver uses `os.LookupEnv` instead of `os.Getenv`
- [ ] `go build ./kernel/config/...` passes
