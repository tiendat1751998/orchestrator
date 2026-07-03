package resilience_test

import (
	"errors"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts/resilience"
)

func TestErrCircuitOpen(t *testing.T) {
	err := resilience.ErrCircuitOpen
	if err.Error() != "circuit breaker is open" {
		t.Errorf("expected %q, got %q", "circuit breaker is open", err.Error())
	}

	// Verify standard error checking behavior
	if !errors.Is(err, resilience.ErrCircuitOpen) {
		t.Error("errors.Is failed for ErrCircuitOpen")
	}
}
