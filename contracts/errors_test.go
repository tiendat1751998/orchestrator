package contracts

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestStructuredErrors(t *testing.T) {
	// Test ValidationError
	vErr := NewValidationError("config", "timeout", "must be positive")
	if vErr == nil {
		t.Fatal("expected non-nil validation error")
	}
	if vErr.Error() != "[config] validation failed on field \"timeout\": must be positive" {
		t.Errorf("unexpected error string: %s", vErr.Error())
	}

	// Test NotFoundError
	nfErr := NewNotFoundError("agent", "reviewer")
	if nfErr.Error() != "resource agent \"reviewer\" not found" {
		t.Errorf("unexpected error string: %s", nfErr.Error())
	}

	// Test TimeoutError
	toErr := NewTimeoutError("execute", 5*time.Second)
	if toErr.Error() != "operation \"execute\" timed out after 5s" {
		t.Errorf("unexpected error string: %s", toErr.Error())
	}

	// Test ConflictError
	cfErr := NewConflictError("agent", "reviewer", "already registered")
	if cfErr.Error() != "conflict on agent \"reviewer\": already registered" {
		t.Errorf("unexpected error string: %s", cfErr.Error())
	}

	// Test PermissionError
	peErr := NewPermissionError("coder", "write", "file")
	if peErr.Error() != "actor \"coder\" denied permission to perform action \"write\" on resource \"file\"" {
		t.Errorf("unexpected error string: %s", peErr.Error())
	}
}

func TestRetryableError(t *testing.T) {
	baseErr := errors.New("transient issue")
	rErr := NewRetryableError(baseErr, 2*time.Second)

	if rErr == nil {
		t.Fatal("expected non-nil retryable error")
	}

	if !IsRetryable(rErr) {
		t.Error("expected rErr to be retryable")
	}

	wrapped := fmt.Errorf("outer: %w", rErr)
	if !IsRetryable(wrapped) {
		t.Error("expected wrapped error to be retryable")
	}

	if IsRetryable(baseErr) {
		t.Error("expected baseErr to not be retryable")
	}

	if IsRetryable(nil) {
		t.Error("expected nil to not be retryable")
	}
}

func TestProxyHelpers(t *testing.T) {
	err1 := NewValidationError("config", "timeout", "must be positive")
	err2 := fmt.Errorf("wrap: %w", err1)

	if !Is(err2, err1) {
		t.Error("expected Is proxy to return true")
	}

	var target *ValidationError
	if !As(err2, &target) || target != err1 {
		t.Error("expected As proxy to extract target error")
	}

	if Unwrap(err2) != err1 {
		t.Error("expected Unwrap proxy to return err1")
	}
}
