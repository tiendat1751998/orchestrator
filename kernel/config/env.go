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
