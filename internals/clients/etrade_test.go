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

func TestGetOrders(t *testing.T) {
	workspace := t.TempDir()

	// Create valid token for tests.
	expires_at := time.Now().Add(24 * time.Hour)
	SaveETradeToken(workspace, "test_access_token",
		"test_access_secret", true, expires_at)

	client := NewETrade("key", "secret", workspace, true)

	expectPanic(t, "empty symbol", func() {
		_, _ = client.GetOrders("")
	})

	if _, err := client.GetOrders("BTC-USD"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetTrades(t *testing.T) {
	workspace := t.TempDir()

	// Create valid token for tests.
	expires_at := time.Now().Add(24 * time.Hour)
	SaveETradeToken(workspace, "test_access_token",
		"test_access_secret", true, expires_at)

	client := NewETrade("key", "secret", workspace, true)

	expectPanic(t, "empty symbol", func() {
		_, _ = client.GetTrades("")
	})

	tests := []struct {
		name   string
		symbol string
	}{
		{"BTC-USD", "BTC-USD"},
		{"ETH-USD", "ETH-USD"},
		{"LTC-USD", "LTC-USD"},
		{"XRP-USD", "XRP-USD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := client.GetTrades(tt.symbol); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
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
