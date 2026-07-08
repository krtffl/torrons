package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecretEnv(t *testing.T) {
	t.Run("falls back to the plain env var when no _FILE is set", func(t *testing.T) {
		t.Setenv("TEST_SECRET", "plain-value")

		if got := secretEnv("TEST_SECRET"); got != "plain-value" {
			t.Errorf("secretEnv() = %q, want %q", got, "plain-value")
		}
	})

	t.Run("reads from the file when _FILE is set, trimming whitespace", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "secret")
		if err := os.WriteFile(path, []byte("file-value\n"), 0o600); err != nil {
			t.Fatalf("failed to write test secret file: %v", err)
		}
		t.Setenv("TEST_SECRET_FILE", path)
		t.Setenv("TEST_SECRET", "plain-value")

		if got := secretEnv("TEST_SECRET"); got != "file-value" {
			t.Errorf("secretEnv() = %q, want %q (file should take precedence)", got, "file-value")
		}
	})

	t.Run("returns empty when the _FILE path is unreadable", func(t *testing.T) {
		t.Setenv("TEST_SECRET_FILE", filepath.Join(t.TempDir(), "does-not-exist"))

		if got := secretEnv("TEST_SECRET"); got != "" {
			t.Errorf("secretEnv() = %q, want empty string on read error", got)
		}
	})

	t.Run("returns empty when neither is set", func(t *testing.T) {
		if got := secretEnv("TEST_SECRET_UNSET"); got != "" {
			t.Errorf("secretEnv() = %q, want empty string", got)
		}
	})
}
