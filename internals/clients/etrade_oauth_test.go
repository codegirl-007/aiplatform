package clients

import (
	"testing"

	"github.com/dghubble/oauth1"
)

// TestOAuthEndpoints_Sandbox verifies sandbox endpoint URLs.
func TestOAuthEndpoints_Sandbox(t *testing.T) {
	request, access, authorize := oauth_endpoints(true)

	if request != "https://apisb.etrade.com/oauth/request_token" {
		t.Errorf("expected sandbox request token URL, got %s", request)
	}
	if access != "https://apisb.etrade.com/oauth/access_token" {
		t.Errorf("expected sandbox access token URL, got %s", access)
	}
	if authorize != "https://us.etrade.com/e/t/etws/authorize" {
		t.Errorf("expected authorize URL, got %s", authorize)
	}
}

// TestOAuthEndpoints_Production verifies production endpoint URLs.
func TestOAuthEndpoints_Production(t *testing.T) {
	request, access, authorize := oauth_endpoints(false)

	if request != "https://api.etrade.com/oauth/request_token" {
		t.Errorf("expected production request token URL, got %s", request)
	}
	if access != "https://api.etrade.com/oauth/access_token" {
		t.Errorf("expected production access token URL, got %s", access)
	}
	if authorize != "https://us.etrade.com/e/t/etws/authorize" {
		t.Errorf("expected authorize URL, got %s", authorize)
	}
}

// TestAPIBaseURL verifies API base URL selection.
func TestAPIBaseURL(t *testing.T) {
	sandbox_url := APIBaseURL(true)
	if sandbox_url != "https://apisb.etrade.com" {
		t.Errorf("expected sandbox base URL, got %s", sandbox_url)
	}

	prod_url := APIBaseURL(false)
	if prod_url != "https://api.etrade.com" {
		t.Errorf("expected production base URL, got %s", prod_url)
	}
}

// TestRenewTokenURL verifies renew token URL selection.
func TestRenewTokenURL(t *testing.T) {
	sandbox_url := renew_token_url(true)
	if sandbox_url != "https://apisb.etrade.com/oauth/renew_access_token" {
		t.Errorf("expected sandbox renew URL, got %s", sandbox_url)
	}

	prod_url := renew_token_url(false)
	if prod_url != "https://api.etrade.com/oauth/renew_access_token" {
		t.Errorf("expected production renew URL, got %s", prod_url)
	}
}

// TestNewOAuthConfig verifies OAuth config creation.
func TestNewOAuthConfig(t *testing.T) {
	config := NewOAuthConfig("test_key", "test_secret", true)

	if config == nil {
		t.Fatal("expected config, got nil")
	}
	if config.ConsumerKey != "test_key" {
		t.Errorf("expected consumer key 'test_key', got %s",
			config.ConsumerKey)
	}
	if config.ConsumerSecret != "test_secret" {
		t.Errorf("expected consumer secret 'test_secret', got %s",
			config.ConsumerSecret)
	}
	if config.CallbackURL != "oob" {
		t.Errorf("expected callback 'oob', got %s", config.CallbackURL)
	}
	if config.Endpoint.RequestTokenURL !=
		"https://apisb.etrade.com/oauth/request_token" {
		t.Errorf("expected sandbox request token URL, got %s",
			config.Endpoint.RequestTokenURL)
	}
}

// TestNewOAuthConfig_EmptyKey verifies panic on empty consumer key.
func TestNewOAuthConfig_EmptyKey(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for empty consumer key")
		}
	}()

	NewOAuthConfig("", "secret", true)
}

// TestNewOAuthConfig_EmptySecret verifies panic on empty consumer secret.
func TestNewOAuthConfig_EmptySecret(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for empty consumer secret")
		}
	}()

	NewOAuthConfig("key", "", true)
}

// TestAuthorizationURL verifies authorization URL construction.
func TestAuthorizationURL(t *testing.T) {
	config := NewOAuthConfig("test_key", "test_secret", true)
	request_token := "test_request_token"

	auth_url := AuthorizationURL(config, request_token)

	if auth_url == "" {
		t.Fatal("expected non-empty authorization URL")
	}

	// ETrade uses non-standard format: ?key=<consumer_key>&token=<request_token>
	expected := "https://us.etrade.com/e/t/etws/authorize?key=test_key&token=test_request_token"
	if auth_url != expected {
		t.Errorf("expected URL:\n  %s\ngot:\n  %s", expected, auth_url)
	}
}

// TestAuthorizationURL_EmptyToken verifies panic on empty request token.
func TestAuthorizationURL_EmptyToken(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for empty request token")
		}
	}()

	config := NewOAuthConfig("key", "secret", true)
	AuthorizationURL(config, "")
}

// TestNewOAuthClient verifies OAuth client creation.
func TestNewOAuthClient(t *testing.T) {
	config := NewOAuthConfig("test_key", "test_secret", true)
	client := NewOAuthClient(config, "access_token", "access_secret")

	if client == nil {
		t.Fatal("expected HTTP client, got nil")
	}
}

// TestNewOAuthClient_EmptyToken verifies panic on empty access token.
func TestNewOAuthClient_EmptyToken(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for empty access token")
		}
	}()

	config := NewOAuthConfig("key", "secret", true)
	NewOAuthClient(config, "", "secret")
}

// TestNewOAuthClient_EmptySecret verifies panic on empty access secret.
func TestNewOAuthClient_EmptySecret(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for empty access secret")
		}
	}()

	config := NewOAuthConfig("key", "secret", true)
	NewOAuthClient(config, "token", "")
}

// TestParseCallbackVerifier verifies verifier extraction from callback URL.
func TestParseCallbackVerifier(t *testing.T) {
	callback_url := "https://myapp.com/callback?oauth_verifier=test_verifier&other=param"

	verifier, err := parse_callback_verifier(callback_url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verifier != "test_verifier" {
		t.Errorf("expected verifier 'test_verifier', got %s", verifier)
	}
}

// TestParseCallbackVerifier_Missing verifies error when verifier missing.
func TestParseCallbackVerifier_Missing(t *testing.T) {
	callback_url := "https://myapp.com/callback?other=param"

	_, err := parse_callback_verifier(callback_url)
	if err == nil {
		t.Errorf("expected error for missing verifier")
	}
}

// TestParseCallbackVerifier_InvalidURL verifies error on invalid URL.
func TestParseCallbackVerifier_InvalidURL(t *testing.T) {
	callback_url := "not a valid url ://"

	_, err := parse_callback_verifier(callback_url)
	if err == nil {
		t.Errorf("expected error for invalid URL")
	}
}

// TestParseCallbackVerifier_Empty verifies panic on empty callback URL.
func TestParseCallbackVerifier_Empty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for empty callback URL")
		}
	}()

	parse_callback_verifier("")
}

// TestOAuthHelperMessage verifies helper message is non-empty.
func TestOAuthHelperMessage(t *testing.T) {
	message := OAuthHelperMessage()
	if message == "" {
		t.Errorf("expected non-empty OAuth helper message")
	}
	if len(message) < 50 {
		t.Errorf("expected longer OAuth helper message, got %d chars",
			len(message))
	}
}

// TestCreateOAuthToken verifies oauth1.Token creation.
func TestCreateOAuthToken(t *testing.T) {
	token := CreateOAuthToken("access", "secret")

	if token == nil {
		t.Fatal("expected token, got nil")
	}
	if token.Token != "access" {
		t.Errorf("expected token 'access', got %s", token.Token)
	}
	if token.TokenSecret != "secret" {
		t.Errorf("expected token secret 'secret', got %s",
			token.TokenSecret)
	}
}

// TestCreateOAuthToken_EmptyToken verifies panic on empty access token.
func TestCreateOAuthToken_EmptyToken(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for empty access token")
		}
	}()

	CreateOAuthToken("", "secret")
}

// TestCreateOAuthToken_EmptySecret verifies panic on empty access secret.
func TestCreateOAuthToken_EmptySecret(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for empty access secret")
		}
	}()

	CreateOAuthToken("token", "")
}

// TestParseSandboxEnv verifies ETRADE_SANDBOX parsing.
func TestParseSandboxEnv(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"true", "true", true},
		{"True", "True", true},
		{"TRUE", "TRUE", true},
		{"1", "1", true},
		{"false", "false", false},
		{"False", "False", false},
		{"0", "0", false},
		{"empty", "", false},
		{"other", "yes", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save current env value.
			old := oauth1.NoContext
			defer func() { _ = old }()

			t.Setenv("ETRADE_SANDBOX", tt.value)
			got := ParseSandboxEnv()
			if got != tt.want {
				t.Errorf("ParseSandboxEnv(%q) = %v, want %v",
					tt.value, got, tt.want)
			}
		})
	}
}
