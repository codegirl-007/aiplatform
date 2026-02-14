package clients

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"aiplatform/pkg/assert"
)

// etrade_oauth_token represents the OAuth credentials for ETrade API.
// These tokens are obtained through the OAuth 1.0a flow and must be
// persisted to avoid requiring re-authorization on every application start.
type etrade_oauth_token struct {
	AccessToken       string    `json:"access_token"`
	AccessTokenSecret string    `json:"access_token_secret"`
	CreatedAt         time.Time `json:"created_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	Sandbox           bool      `json:"sandbox"`
}

// is_expired checks if the token has passed its expiration time.
// Returns true if the current time is after ExpiresAt.
func (t *etrade_oauth_token) is_expired() bool {
	assert.Is_true(!t.CreatedAt.IsZero(), "created_at must be set")
	assert.Is_true(!t.ExpiresAt.IsZero(), "expires_at must be set")
	return time.Now().After(t.ExpiresAt)
}

// credentials_path constructs the path to the token storage file.
// Following project pattern: {workspace}/.aiplatform/credentials/etrade_tokens.json
func credentials_path(workspaceRoot string) string {
	assert.Is_true(filepath.IsAbs(workspaceRoot),
		"workspace root must be absolute path")
	return filepath.Join(workspaceRoot, ".aiplatform", "credentials",
		"etrade_tokens.json")
}

// save_etrade_token persists the OAuth token to disk using atomic write.
// Write to a temp file, then rename for atomicity (POSIX rename is atomic).
// Panics if any operation fails - logging into ETrade is not optional.
func save_etrade_token(workspaceRoot string,
	token *etrade_oauth_token) {
	assert.Is_true(filepath.IsAbs(workspaceRoot),
		"workspace root must be absolute path")
	assert.Not_nil(token, "token must not be nil")
	assert.Not_empty(token.AccessToken, "access_token must not be empty")
	assert.Not_empty(token.AccessTokenSecret,
		"access_token_secret must not be empty")
	assert.Is_true(!token.CreatedAt.IsZero(), "created_at must be set")
	assert.Is_true(!token.ExpiresAt.IsZero(), "expires_at must be set")

	credentialsDir := filepath.Join(workspaceRoot, ".aiplatform",
		"credentials")

	err := os.MkdirAll(credentialsDir, 0755)
	assert.No_err(err, fmt.Sprintf("failed to create credentials directory %s",
		credentialsDir))

	data, err := json.MarshalIndent(token, "", "  ")
	assert.No_err(err, "failed to marshal token")

	finalPath := credentials_path(workspaceRoot)

	tempFile, err := os.CreateTemp(credentialsDir, "etrade_tokens.*.tmp")
	assert.No_err(err, "failed to create temp file")
	tempPath := tempFile.Name()

	_, err = tempFile.Write(data)
	if err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		assert.No_err(err, "failed to write token data")
	}

	err = tempFile.Close()
	if err != nil {
		os.Remove(tempPath)
		assert.No_err(err, "failed to close temp file")
	}

	err = os.Chmod(tempPath, 0600)
	if err != nil {
		os.Remove(tempPath)
		assert.No_err(err, "failed to set file permissions")
	}

	err = os.Rename(tempPath, finalPath)
	if err != nil {
		os.Remove(tempPath)
		assert.No_err(err, "failed to rename temp file")
	}
}

// load_etrade_token reads the OAuth token from disk.
// Returns the token if it exists and is valid.
// Returns nil if the file doesn't exist (first-time use).
// Panics if the file exists but is corrupt or unreadable.
func load_etrade_token(workspaceRoot string) *etrade_oauth_token {
	assert.Is_true(filepath.IsAbs(workspaceRoot),
		"workspace root must be absolute path")

	tokenPath := credentials_path(workspaceRoot)

	// Check if file exists.
	stat_info, err := os.Stat(tokenPath)
	if os.IsNotExist(err) {
		return nil
	}
	assert.No_err(err, fmt.Sprintf("failed to stat token file %s", tokenPath))
	assert.Not_nil(stat_info, "stat info should not be nil")

	// Read and parse token file.
	data, err := os.ReadFile(tokenPath)
	assert.No_err(err, fmt.Sprintf("failed to read token file %s", tokenPath))

	var token etrade_oauth_token
	err = json.Unmarshal(data, &token)
	assert.No_err(err, fmt.Sprintf("failed to parse token JSON from %s", tokenPath))

	// Validate token fields are non-empty (zero values are invalid).
	assert.Not_empty(token.AccessToken, "access_token must not be empty")
	assert.Not_empty(token.AccessTokenSecret,
		"access_token_secret must not be empty")
	assert.Is_true(!token.CreatedAt.IsZero(), "created_at must be set")
	assert.Is_true(!token.ExpiresAt.IsZero(), "expires_at must be set")

	return &token
}
