package resilience

import (
	"errors"
)

// WithFallback executes the primary function call, falling back to a backup function if the primary fails.
func WithFallback(primary func() error, fallback func() error) error {
	if primary == nil {
		return errors.New("fallback: primary call function cannot be nil")
	}

	err := primary()
	if err == nil {
		return nil // Success on primary call
	}

	if fallback == nil {
		return err // No fallback provided, return primary failure
	}

	// Try backup fallback call
	return fallback()
}
