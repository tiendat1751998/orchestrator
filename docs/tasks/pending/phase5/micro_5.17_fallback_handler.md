# Micro-Task 5.17: Create kernel/resilience/fallback.go

## Info
- **File**: `kernel/resilience/fallback.go`
- **Package**: `resilience`
- **Depends on**: 5.16
- **Time**: 15 min
- **Verify**: `go build ./kernel/resilience/...`

## Purpose
Implements the backup driver fallback orchestrator (`WithFallback` and selectors) to redirect calls to alternative providers if the primary option fails.

## EXACT code to create

```go
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
```

## Pitfalls

### Pitfall 1: Returning primary errors when fallback succeeds
If the primary call fails but the backup call succeeds, returning the primary error to the caller is incorrect. Return the result of the fallback function.

### Pitfall 2: Silent failures when calling nil functions
Executing unvalidated function pointer arguments will cause runtime panics. Validate pointers before calling them.

## Verify
```bash
go build ./kernel/resilience/...
# Expected: clean compilation without errors
```

## Checklist
- [ ] File exists at `kernel/resilience/fallback.go`
- [ ] Package name is `resilience`
- [ ] All exported types have Godoc
- [ ] Primary errors are handled and trigger backup routines
- [ ] Backup function results are returned correctly
- [ ] Build command passes
