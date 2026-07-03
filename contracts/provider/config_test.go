package provider_test

import (
	"testing"
	"time"

	"github.com/tiendat1751998/orchestrator/contracts/provider"
)

func TestConfigGetExtra(t *testing.T) {
	c := provider.Config{
		Extra: map[string]string{
			"key1": "value1",
		},
	}

	if val := c.GetExtra("key1", "default"); val != "value1" {
		t.Errorf("expected value1, got %q", val)
	}

	if val := c.GetExtra("key2", "default"); val != "default" {
		t.Errorf("expected default, got %q", val)
	}
}

func TestConfigTimeoutOrDefault(t *testing.T) {
	c1 := provider.Config{
		Timeout: 5 * time.Second,
	}
	if timeout := c1.TimeoutOrDefault(); timeout != 5*time.Second {
		t.Errorf("expected 5s, got %v", timeout)
	}

	c2 := provider.Config{}
	if timeout := c2.TimeoutOrDefault(); timeout != 120*time.Second {
		t.Errorf("expected 120s, got %v", timeout)
	}
}

func TestConfigMaxRetryOrDefault(t *testing.T) {
	c1 := provider.Config{
		MaxRetry: 5,
	}
	if maxRetry := c1.MaxRetryOrDefault(); maxRetry != 5 {
		t.Errorf("expected 5, got %d", maxRetry)
	}

	c2 := provider.Config{}
	if maxRetry := c2.MaxRetryOrDefault(); maxRetry != 3 {
		t.Errorf("expected 3, got %d", maxRetry)
	}
}
