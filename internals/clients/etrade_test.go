package clients

import (
	"testing"
	"time"
)

func TestNewETrade(t *testing.T) {
	workspace := t.TempDir()

	expectPanic(t, "empty keys", func() {
		NewETrade("", "", workspace, true)
	})

	// NewETrade now requires a valid token, so we expect a panic
	// if no token is available.
	expectPanic(t, "no token available", func() {
		NewETrade("key", "secret", workspace, true)
	})

	// Save a valid token and try again.
	expires_at := time.Now().Add(24 * time.Hour)
	SaveETradeToken(workspace, "test_access_token",
		"test_access_secret", true, expires_at)

	client := NewETrade("key", "secret", workspace, true)
	if client == nil {
		t.Fatalf("expected client, got nil")
	}
}

func expectPanic(t *testing.T, name string, fn func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic: %s", name)
		}
	}()

	fn()
}
