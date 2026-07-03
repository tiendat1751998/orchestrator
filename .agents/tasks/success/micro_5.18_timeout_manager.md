# Micro-Task 5.18: Create kernel/resilience/timeout.go Success

Completed successfully.

## Verification
- Implemented `CascadingTimeoutContext` to derive a sub-context bounded by parent deadline.
- Capped child timeouts to parent remaining times.
- Verified resources are cleaned up immediately if the parent context is already expired.
- Formatted, vetted, and compiled cleanly. All unit tests passed.
