package config

import (
	"os"
	"reflect"
	"testing"
)

func TestResolveEnvVars(t *testing.T) {
	// Set test environment variables
	os.Setenv("TEST_VAR_ONE", "value_one")
	os.Setenv("TEST_VAR_EMPTY", "")
	defer func() {
		os.Unsetenv("TEST_VAR_ONE")
		os.Unsetenv("TEST_VAR_EMPTY")
	}()

	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		{
			name:      "Simple substitution",
			input:     "prefix-${TEST_VAR_ONE}-suffix",
			expected:  "prefix-value_one-suffix",
			expectErr: false,
		},
		{
			name:      "Empty environment variable value",
			input:     "prefix-${TEST_VAR_EMPTY}-suffix",
			expected:  "prefix--suffix",
			expectErr: false,
		},
		{
			name:      "Single env var that is missing",
			input:     "${TEST_VAR_MISSING}",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Single env var that is missing with spaces",
			input:     "  ${TEST_VAR_MISSING}  ",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "Mixed text and missing env var",
			input:     "prefix-${TEST_VAR_MISSING}",
			expected:  "prefix-",
			expectErr: false, // implementation returns result, nil when not a single var
		},
		{
			name:      "Invalid env var pattern",
			input:     "${invalid-name}",
			expected:  "${invalid-name}",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ResolveEnvVars(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error presence %v, got error: %v", tt.expectErr, err)
			}
			if res != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, res)
			}
		})
	}
}

func TestResolveEnvInMap(t *testing.T) {
	os.Setenv("PORT", "8080")
	os.Setenv("DB_USER", "postgres")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DB_USER")
	}()

	data := map[string]any{
		"server": map[string]any{
			"port": "${PORT}",
			"host": "localhost",
		},
		"database": []any{
			"${DB_USER}",
			"password",
		},
		"version": 1.0,
	}

	expected := map[string]any{
		"server": map[string]any{
			"port": "8080",
			"host": "localhost",
		},
		"database": []any{
			"postgres",
			"password",
		},
		"version": 1.0,
	}

	err := ResolveEnvInMap(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(data, expected) {
		t.Errorf("expected resolved map %+v, got %+v", expected, data)
	}
}
