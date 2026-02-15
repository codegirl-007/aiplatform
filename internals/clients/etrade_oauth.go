package clients

import (
	"aiplatform/pkg/assert"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dghubble/oauth1"
)

const (
	sandbox_request_token_url = "https://apisb.etrade.com/oauth/request_token"
	sandbox_access_token_url  = "https://apisb.etrade.com/oauth/access_token"
	sandbox_authorize_url     = "https://us.etrade.com/e/t/etws/authorize"

	prod_request_token_url  = "https://api.etrade.com/oauth/request_token"
	prod_access_token_url   = "https://api.etrade.com/oauth/access_token"
	prod_authorize_url      = "https://us.etrade.com/e/t/etws/authorize"
	sandbox_api_base_url    = "https://apisb.etrade.com"
	prod_api_base_url       = "https://api.etrade.com"
	sandbox_renew_token_url = "https://apisb.etrade.com/oauth/renew_access_token"
	prod_renew_token_url    = "https://api.etrade.com/oauth/renew_access_token"
)

// oauth_endpoints returns the request token, access token, and authorize URLs
// for the given sandbox flag.
func oauth_endpoints(sandbox bool) (string, string, string) {
	var request_url, access_url, authorize_url string

	if sandbox {
		request_url = sandbox_request_token_url
		access_url = sandbox_access_token_url
		authorize_url = sandbox_authorize_url
	} else {
		request_url = prod_request_token_url
		access_url = prod_access_token_url
		authorize_url = prod_authorize_url
	}

	assert.Not_empty(request_url, "request token URL must not be empty")
	assert.Not_empty(access_url, "access token URL must not be empty")
	assert.Not_empty(authorize_url, "authorize URL must not be empty")

	return request_url, access_url, authorize_url
}

// APIBaseURL returns the API base URL for the given sandbox flag.
func APIBaseURL(sandbox bool) string {
	var base_url string
	if sandbox {
		base_url = sandbox_api_base_url
	} else {
		base_url = prod_api_base_url
	}

	assert.Not_empty(base_url, "API base URL must not be empty")

	return base_url
}

const ()

// renew_token_url returns the renew access token URL for the given sandbox flag.
func renew_token_url(sandbox bool) string {
	var renew_url string
	if sandbox {
		renew_url = sandbox_renew_token_url
	} else {
		renew_url = prod_renew_token_url
	}

	assert.Not_empty(renew_url, "renew token URL must not be empty")

	return renew_url
}

// NewOAuthConfig creates an OAuth 1.0a config for ETrade.
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
func AuthorizationURL(config *oauth1.Config, request_token string) string {
	assert.Not_nil(config, "config must not be nil")
	assert.Not_empty(request_token, "request_token must not be empty")

	auth_url := fmt.Sprintf("%s?key=%s&token=%s",
		config.Endpoint.AuthorizeURL,
		url.QueryEscape(config.ConsumerKey),
		url.QueryEscape(request_token))

	assert.Not_empty(auth_url, "authorization URL must not be empty")

	return auth_url
}

// ExchangeToken exchanges the request token and verifier for an access token.
func ExchangeToken(config *oauth1.Config,
	request_token, request_secret, verifier string) (string, string, error) {
	assert.Not_nil(config, "config must not be nil")
	assert.Not_empty(request_token, "request_token must not be empty")
	assert.Not_empty(request_secret, "request_secret must not be empty")
	assert.Not_empty(verifier, "verifier must not be empty")

	access_token, access_secret, err := config.AccessToken(request_token, request_secret, verifier)
	assert.No_err(err, "failed to exchange for access token")
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
