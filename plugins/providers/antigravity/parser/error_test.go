package parser

import (
	"errors"
	"testing"

	"github.com/tiendat1751998/orchestrator/contracts"
)

func TestParseError(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError error
	}{
		{
			name:          "empty input",
			input:         "",
			expectedError: nil,
		},
		{
			name:          "rate limit keyword",
			input:         "API rate limit exceeded",
			expectedError: contracts.ErrProviderRateLimited,
		},
		{
			name:          "quota exceeded keyword",
			input:         "daily quota exceeded",
			expectedError: contracts.ErrProviderRateLimited,
		},
		{
			name:          "429 status code",
			input:         "HTTP error 429",
			expectedError: contracts.ErrProviderRateLimited,
		},
		{
			name:          "auth fail api key",
			input:         "invalid API key provided",
			expectedError: contracts.ErrProviderAuthFailed,
		},
		{
			name:          "auth fail invalid credentials",
			input:         "invalid credentials for account",
			expectedError: contracts.ErrProviderAuthFailed,
		},
		{
			name:          "auth fail 403 status code",
			input:         "unauthorized, error 403",
			expectedError: contracts.ErrProviderAuthFailed,
		},
		{
			name:          "timeout deadline exceeded",
			input:         "context deadline exceeded",
			expectedError: contracts.ErrProviderTimeout,
		},
		{
			name:          "timeout 504 status code",
			input:         "gateway timeout status 504",
			expectedError: contracts.ErrProviderTimeout,
		},
		{
			name:          "availability not found",
			input:         "page not found on server",
			expectedError: contracts.ErrProviderUnavailable,
		},
		{
			name:          "availability command not found",
			input:         "bash: antigravity: command not found",
			expectedError: contracts.ErrProviderUnavailable,
		},
		{
			name:          "availability 503 status code",
			input:         "503 Service Temporarily Unavailable",
			expectedError: contracts.ErrProviderUnavailable,
		},
		{
			name:          "fallback raw error",
			input:         "some random system error happened",
			expectedError: errors.New("antigravity CLI error: some random system error happened"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseError(tt.input)
			if got == nil {
				if tt.expectedError != nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				}
				return
			}

			if tt.expectedError == nil {
				t.Errorf("unexpected error: %v", got)
				return
			}

			if errors.Is(got, tt.expectedError) {
				return
			}

			// For fallback errors, check string equivalence
			if got.Error() != tt.expectedError.Error() {
				t.Errorf("expected error %q, got %q", tt.expectedError, got)
			}
		})
	}
}
