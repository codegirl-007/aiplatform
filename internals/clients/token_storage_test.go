package clients

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestInvariant_TokenStoragePath verifies that credentials_path
// constructs the correct workspace-relative path.
func TestInvariant_TokenStoragePath(t *testing.T) {
	tempDir := t.TempDir()
	workspace := filepath.Join(tempDir, "workspace")

	expected := filepath.Join(workspace, ".aiplatform", "credentials",
		"etrade_tokens.json")
	actual := credentials_path(workspace)

	if actual != expected {
		t.Fatalf("expected path %s, got %s", expected, actual)
	}
}

// TestInvariant_TokenStoragePathRequiresAbsolute verifies that
// credentials_path panics if given a relative path (Invariant 4a).
func TestInvariant_TokenStoragePathRequiresAbsolute(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for relative path, got none")
		}
	}()

	// This should panic because "workspace" is relative.
	credentials_path("workspace")
}

// TestTokenStorage_SaveAndLoad verifies the complete round-trip:
// save a token, then load it back and verify all fields match.
func TestTokenStorage_SaveAndLoad(t *testing.T) {
	workspace := t.TempDir()

	// Create a valid token.
	created := time.Date(2026, 2, 13, 12, 0, 0, 0, time.UTC)
	expires := created.Add(24 * time.Hour)
	token := &etrade_oauth_token{
		AccessToken:       "test_access_token",
		AccessTokenSecret: "test_access_secret",
		CreatedAt:         created,
		ExpiresAt:         expires,
		Sandbox:           true,
	}

	// Save the token (panics on error, so no error return).
	save_etrade_token(workspace, token)

	// Load the token back.
	loaded := load_etrade_token(workspace)
	if loaded == nil {
		t.Fatalf("expected loaded token, got nil")
	}

	// Verify all fields match.
	if loaded.AccessToken != token.AccessToken {
		t.Errorf("access_token mismatch: expected %s, got %s",
			token.AccessToken, loaded.AccessToken)
	}
	if loaded.AccessTokenSecret != token.AccessTokenSecret {
		t.Errorf("access_token_secret mismatch: expected %s, got %s",
			token.AccessTokenSecret, loaded.AccessTokenSecret)
	}
	if !loaded.CreatedAt.Equal(token.CreatedAt) {
		t.Errorf("created_at mismatch: expected %v, got %v",
			token.CreatedAt, loaded.CreatedAt)
	}
	if !loaded.ExpiresAt.Equal(token.ExpiresAt) {
		t.Errorf("expires_at mismatch: expected %v, got %v",
			token.ExpiresAt, loaded.ExpiresAt)
	}
	if loaded.Sandbox != token.Sandbox {
		t.Errorf("sandbox mismatch: expected %v, got %v",
			token.Sandbox, loaded.Sandbox)
	}
}

// TestTokenStorage_LoadNonExistent verifies that loading from
// a workspace with no token file returns nil.
func TestTokenStorage_LoadNonExistent(t *testing.T) {
	workspace := t.TempDir()

	loaded := load_etrade_token(workspace)
	if loaded != nil {
		t.Fatalf("expected nil token for missing file, got: %+v", loaded)
	}
}

// TestTokenStorage_LoadCorruptJSON verifies that loading a file
// with invalid JSON panics (assertion failure).
func TestTokenStorage_LoadCorruptJSON(t *testing.T) {
	workspace := t.TempDir()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for corrupt JSON, got none")
		}
	}()

	// Create the credentials directory.
	credDir := filepath.Join(workspace, ".aiplatform", "credentials")
	if err := os.MkdirAll(credDir, 0755); err != nil {
		t.Fatalf("failed to create credentials dir: %v", err)
	}

	// Write invalid JSON to the token file.
	tokenPath := credentials_path(workspace)
	if err := os.WriteFile(tokenPath, []byte("not valid json"), 0600); err != nil {
		t.Fatalf("failed to write corrupt file: %v", err)
	}

	// Attempt to load should panic.
	load_etrade_token(workspace)
}

// TestTokenStorage_LoadEmptyAccessToken verifies that a token
// with empty access_token field panics (assertion failure).
func TestTokenStorage_LoadEmptyAccessToken(t *testing.T) {
	workspace := t.TempDir()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for empty access_token, got none")
		}
	}()

	// Create credentials directory.
	credDir := filepath.Join(workspace, ".aiplatform", "credentials")
	if err := os.MkdirAll(credDir, 0755); err != nil {
		t.Fatalf("failed to create credentials dir: %v", err)
	}

	// Create token with empty access_token.
	token := etrade_oauth_token{
		AccessToken:       "", // Empty (invalid)
		AccessTokenSecret: "secret",
		CreatedAt:         time.Now(),
		ExpiresAt:         time.Now().Add(24 * time.Hour),
		Sandbox:           true,
	}
	data, _ := json.MarshalIndent(token, "", "  ")

	tokenPath := credentials_path(workspace)
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}

	// Load should panic on validation.
	load_etrade_token(workspace)
}

// TestTokenStorage_SaveEmptyAccessToken verifies that saving
// a token with empty access_token panics (assertion failure).
func TestTokenStorage_SaveEmptyAccessToken(t *testing.T) {
	workspace := t.TempDir()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for empty access_token, got none")
		}
	}()

	token := &etrade_oauth_token{
		AccessToken:       "", // Empty (invalid)
		AccessTokenSecret: "secret",
		CreatedAt:         time.Now(),
		ExpiresAt:         time.Now().Add(24 * time.Hour),
		Sandbox:           true,
	}

	// Should panic due to assertion.
	save_etrade_token(workspace, token)
}

// TestTokenStorage_SaveRelativeWorkspace verifies that saving
// with a relative workspace path panics (assertion failure).
func TestTokenStorage_SaveRelativeWorkspace(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for relative workspace path, got none")
		}
	}()

	token := &etrade_oauth_token{
		AccessToken:       "test_token",
		AccessTokenSecret: "test_secret",
		CreatedAt:         time.Now(),
		ExpiresAt:         time.Now().Add(24 * time.Hour),
		Sandbox:           true,
	}

	// Should panic due to assertion on relative path.
	save_etrade_token("relative/path", token)
}

// TestTokenStorage_AtomicWrite verifies that the save operation
// is atomic by checking that only the final file exists after save.
func TestTokenStorage_AtomicWrite(t *testing.T) {
	workspace := t.TempDir()

	token := &etrade_oauth_token{
		AccessToken:       "test_token",
		AccessTokenSecret: "test_secret",
		CreatedAt:         time.Now(),
		ExpiresAt:         time.Now().Add(24 * time.Hour),
		Sandbox:           false,
	}

	save_etrade_token(workspace, token)

	// Check that the final file exists.
	tokenPath := credentials_path(workspace)
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		t.Fatalf("expected token file to exist at %s", tokenPath)
	}

	// Check that no temp files remain.
	credDir := filepath.Join(workspace, ".aiplatform", "credentials")
	entries, err := os.ReadDir(credDir)
	if err != nil {
		t.Fatalf("failed to read credentials dir: %v", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".tmp" {
			t.Errorf("found temp file after save: %s", entry.Name())
		}
	}
}

// TestTokenStorage_FilePermissions verifies that the token file
// is created with restrictive permissions (0600).
func TestTokenStorage_FilePermissions(t *testing.T) {
	workspace := t.TempDir()

	token := &etrade_oauth_token{
		AccessToken:       "test_token",
		AccessTokenSecret: "test_secret",
		CreatedAt:         time.Now(),
		ExpiresAt:         time.Now().Add(24 * time.Hour),
		Sandbox:           true,
	}

	save_etrade_token(workspace, token)

	tokenPath := credentials_path(workspace)
	info, err := os.Stat(tokenPath)
	if err != nil {
		t.Fatalf("failed to stat token file: %v", err)
	}

	// Check permissions (0600 = owner read/write only).
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected permissions 0600, got %04o", perm)
	}
}

// TestToken_IsExpired verifies the token expiration check.
func TestToken_IsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired_yesterday",
			expiresAt: now.Add(-24 * time.Hour),
			want:      true,
		},
		{
			name:      "expires_tomorrow",
			expiresAt: now.Add(24 * time.Hour),
			want:      false,
		},
		{
			name:      "expires_in_one_second",
			expiresAt: now.Add(1 * time.Second),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &etrade_oauth_token{
				AccessToken:       "test",
				AccessTokenSecret: "secret",
				CreatedAt:         now.Add(-1 * time.Hour),
				ExpiresAt:         tt.expiresAt,
				Sandbox:           true,
			}

			if got := token.is_expired(); got != tt.want {
				t.Errorf("is_expired() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTokenStorage_Overwrite verifies that saving a new token
// overwrites the previous one.
func TestTokenStorage_Overwrite(t *testing.T) {
	workspace := t.TempDir()

	// Save first token.
	token1 := &etrade_oauth_token{
		AccessToken:       "first_token",
		AccessTokenSecret: "first_secret",
		CreatedAt:         time.Now(),
		ExpiresAt:         time.Now().Add(24 * time.Hour),
		Sandbox:           true,
	}
	save_etrade_token(workspace, token1)

	// Save second token (should overwrite).
	token2 := &etrade_oauth_token{
		AccessToken:       "second_token",
		AccessTokenSecret: "second_secret",
		CreatedAt:         time.Now(),
		ExpiresAt:         time.Now().Add(48 * time.Hour),
		Sandbox:           false,
	}
	save_etrade_token(workspace, token2)

	// Load and verify it's the second token.
	loaded := load_etrade_token(workspace)
	if loaded.AccessToken != token2.AccessToken {
		t.Errorf("expected second token, got first token")
	}
}
