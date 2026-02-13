package clients

import (
	"testing"
)

func TestNewETrade(t *testing.T) {
	expectPanic(t, "empty keys", func() {
		NewETrade("", "")
	})

	if client := NewETrade("key", "secret"); client == nil {
		t.Fatalf("expected client, got nil")
	}
}

func TestGetOrders(t *testing.T) {
	client := NewETrade("key", "secret")

	expectPanic(t, "empty symbol", func() {
		_, _ = client.GetOrders("")
	})

	if _, err := client.GetOrders("BTC-USD"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetTrades(t *testing.T) {
	client := NewETrade("key", "secret")

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
