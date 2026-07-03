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
