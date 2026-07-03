package security

import (
	"os"
	"testing"
)

func TestLoadSecret(t *testing.T) {
	key := "TEST_MY_SECRET_KEY"
	val := "  my-secret-value  "
	os.Setenv(key, val)
	defer os.Unsetenv(key)

	got := LoadSecret(key)
	want := "my-secret-value"
	if got != want {
		t.Errorf("LoadSecret() = %q, want %q", got, want)
	}
}

func TestRedactSecrets(t *testing.T) {
	os.Setenv("SECRET_KEY_LONG", "supersecretvalue")
	os.Setenv("SECRET_KEY_SHORT", "123")
	defer os.Unsetenv("SECRET_KEY_LONG")
	defer os.Unsetenv("SECRET_KEY_SHORT")

	tests := []struct {
		name  string
		input string
		keys  []string
		want  string
	}{
		{
			name:  "empty input",
			input: "",
			keys:  []string{"SECRET_KEY_LONG"},
			want:  "",
		},
		{
			name:  "redact long secret",
			input: "This is a supersecretvalue in a log.",
			keys:  []string{"SECRET_KEY_LONG"},
			want:  "This is a [REDACTED] in a log.",
		},
		{
			name:  "do not redact short secret",
			input: "Count to 123 now.",
			keys:  []string{"SECRET_KEY_SHORT"},
			want:  "Count to 123 now.",
		},
		{
			name:  "multiple keys",
			input: "Long: supersecretvalue, Short: 123.",
			keys:  []string{"SECRET_KEY_LONG", "SECRET_KEY_SHORT"},
			want:  "Long: [REDACTED], Short: 123.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactSecrets(tt.input, tt.keys)
			if got != tt.want {
				t.Errorf("RedactSecrets() = %q, want %q", got, tt.want)
			}
		})
	}
}
