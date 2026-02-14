package clients

import (
	"aiplatform/pkg/assert"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dghubble/oauth1"
)

// oauth_endpoints returns the request token, access token, and authorize URLs
// for the given sandbox flag.
func oauth_endpoints(sandbox bool) (string, string, string) {
	if sandbox {
		return "https://apisb.etrade.com/oauth/request_token",
			"https://apisb.etrade.com/oauth/access_token",
			"https://us.etrade.com/e/t/etws/authorize"
	}
	return "https://api.etrade.com/oauth/request_token",
		"https://api.etrade.com/oauth/access_token",
		"https://us.etrade.com/e/t/etws/authorize"
}

// APIBaseURL returns the API base URL for the given sandbox flag.
// Exported for use by cmd utilities and Wails backend.
func APIBaseURL(sandbox bool) string {
	if sandbox {
		return "https://apisb.etrade.com"
	}
	return "https://api.etrade.com"
}

// renew_token_url returns the renew access token URL for the given sandbox flag.
func renew_token_url(sandbox bool) string {
	if sandbox {
		return "https://apisb.etrade.com/oauth/renew_access_token"
	}
	return "https://api.etrade.com/oauth/renew_access_token"
}

// NewOAuthConfig creates an OAuth 1.0a config for ETrade.
// ETrade requires oauth_callback to be set to "oob" for out-of-band verification.
// Exported for use by cmd utilities and Wails backend.
func NewOAuthConfig(consumer_key, consumer_secret string,
	sandbox bool) *oauth1.Config {
	assert.Not_empty(consumer_key, "consumer_key must not be empty")
	assert.Not_empty(consumer_secret, "consumer_secret must not be empty")

	request_url, access_url, authorize_url := oauth_endpoints(sandbox)

	return &oauth1.Config{
		ConsumerKey:    consumer_key,
		ConsumerSecret: consumer_secret,
		CallbackURL:    "oob",
		Endpoint: oauth1.Endpoint{
			RequestTokenURL: request_url,
			AuthorizeURL:    authorize_url,
			AccessTokenURL:  access_url,
		},
	}
}

// RequestToken fetches a request token from ETrade.
// Returns the request token and request secret, or an error.
// Exported for use by cmd utilities and Wails backend.
func RequestToken(config *oauth1.Config) (string, string, error) {
	assert.Not_nil(config, "config must not be nil")

	request_token, request_secret, err := config.RequestToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to get request token: %w", err)
	}

	assert.Not_empty(request_token, "request_token must not be empty")
	assert.Not_empty(request_secret, "request_secret must not be empty")

	return request_token, request_secret, nil
}

// AuthorizationURL builds the ETrade authorization URL that the user
// must visit to approve the application and receive a verification code.
// ETrade uses non-standard OAuth parameters: key=<consumer_key>&token=<request_token>
// instead of the standard oauth_token parameter.
// Exported for use by cmd utilities and Wails backend.
func AuthorizationURL(config *oauth1.Config,
	request_token string) string {
	assert.Not_nil(config, "config must not be nil")
	assert.Not_empty(request_token, "request_token must not be empty")

	// ETrade requires: ?key=<consumer_key>&token=<request_token>
	// This is non-standard OAuth 1.0a; most providers use ?oauth_token=<token>
	auth_url := fmt.Sprintf("%s?key=%s&token=%s",
		config.Endpoint.AuthorizeURL,
		url.QueryEscape(config.ConsumerKey),
		url.QueryEscape(request_token))

	assert.Not_empty(auth_url, "authorization URL must not be empty")

	return auth_url
}

// ExchangeToken exchanges the request token and verifier for an access token.
// Returns the access token and access secret, or an error.
// Exported for use by cmd utilities and Wails backend.
func ExchangeToken(config *oauth1.Config, request_token,
	request_secret, verifier string) (string, string, error) {
	assert.Not_nil(config, "config must not be nil")
	assert.Not_empty(request_token, "request_token must not be empty")
	assert.Not_empty(request_secret, "request_secret must not be empty")
	assert.Not_empty(verifier, "verifier must not be empty")

	access_token, access_secret, err := config.AccessToken(
		request_token, request_secret, verifier)
	if err != nil {
		return "", "", fmt.Errorf("failed to exchange for access token: %w",
			err)
	}

	assert.Not_empty(access_token, "access_token must not be empty")
	assert.Not_empty(access_secret, "access_secret must not be empty")

	return access_token, access_secret, nil
}

// NewOAuthClient creates an HTTP client that signs requests with OAuth 1.0a.
// Exported for use by cmd utilities and Wails backend.
func NewOAuthClient(config *oauth1.Config,
	access_token, access_secret string) *http.Client {
	assert.Not_nil(config, "config must not be nil")
	assert.Not_empty(access_token, "access_token must not be empty")
	assert.Not_empty(access_secret, "access_secret must not be empty")

	token := oauth1.NewToken(access_token, access_secret)
	return config.Client(oauth1.NoContext, token)
}

// renew_access_token attempts to renew an ETrade access token.
// Returns nil on success, error otherwise.
func renew_access_token(client *http.Client, sandbox bool) error {
	assert.Not_nil(client, "client must not be nil")

	renew_url := renew_token_url(sandbox)
	resp, err := client.Get(renew_url)
	if err != nil {
		return fmt.Errorf("failed to renew access token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("renew access token failed with status %d",
			resp.StatusCode)
	}

	return nil
}

// OAuthHelperMessage returns a user-friendly message explaining the OAuth
// OOB flow for first-time authentication.
// Exported for use by cmd utilities and Wails backend.
func OAuthHelperMessage() string {
	return `ETrade OAuth Setup Required

To authenticate with ETrade, you need to:
1. Visit the authorization URL below in your browser
2. Log in with your ETrade credentials
3. Accept the authorization request
4. Copy the verification code displayed
5. Paste the code when prompted

The access token will be saved and reused for future requests.
Tokens expire at midnight US Eastern time or after 2 hours of inactivity.
`
}

// parse_callback_verifier extracts the oauth_verifier from a callback URL.
// This is a helper for future loopback/callback implementations.
func parse_callback_verifier(callback_url string) (string, error) {
	assert.Not_empty(callback_url, "callback_url must not be empty")

	parsed, err := url.Parse(callback_url)
	if err != nil {
		return "", fmt.Errorf("failed to parse callback URL: %w", err)
	}

	verifier := parsed.Query().Get("oauth_verifier")
	if verifier == "" {
		return "", fmt.Errorf("oauth_verifier not found in callback URL")
	}

	return verifier, nil
}
