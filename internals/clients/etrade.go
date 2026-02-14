package clients

import (
	"aiplatform/pkg/assert"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// ETrade is the interface for the etrade API.
type ETrade interface {
	GetOrders(symbol string) ([]Order, error)
	GetTrades(symbol string) ([]Trade, error)
}

// Order is a single order from etrade.
type Order struct {
	Symbol string
	ID     string
	Price  float64
	Qty    float64
	Side   string
}

// Trade is a single trade from etrade.
type Trade struct {
	Symbol string
	ID     string
	Price  float64
	Qty    float64
	Side   string
}

// etrade is the implementation of the ETrade interface.
type etrade struct {
	consumer_key    string
	consumer_secret string
	workspace_root  string
	sandbox         bool
	http_client     *http.Client
}

// NewETrade creates a new etrade API client.
// Loads OAuth token from storage and creates authenticated HTTP client.
// Panics if token is missing, expired, or invalid (fail-fast).
func NewETrade(consumer_key, consumer_secret, workspace_root string,
	sandbox bool) ETrade {
	assert.Not_empty(workspace_root, "workspace_root must not be empty")
	assert.Not_empty(consumer_key, "consumer_key must not be empty")
	assert.Not_empty(consumer_secret, "consumer_secret must not be empty")

	// Load OAuth token from storage.
	access_token, access_secret, _, _, err := LoadETradeToken(
		workspace_root, sandbox)

	assert.No_err(err, "failed to load token")
	assert.Not_empty(access_token,
		"authentication required: no token found - run etrade-oauth-test to authenticate")
	assert.Not_empty(access_secret,
		"authentication required: no token secret found")

	config := NewOAuthConfig(consumer_key, consumer_secret, sandbox)
	http_client := NewOAuthClient(config, access_token, access_secret)

	assert.Not_nil(http_client, "http_client must not be nil")

	return &etrade{
		consumer_key:    consumer_key,
		consumer_secret: consumer_secret,
		workspace_root:  workspace_root,
		sandbox:         sandbox,
		http_client:     http_client,
	}
}

// GetOrders returns the orders for the given symbol.
// TODO(COD-17): Implement via E*TRADE accounts/orders endpoint. https://linear.app/codegirl/issue/COD-17
func (e *etrade) GetOrders(symbol string) ([]Order, error) {
	assert.Not_empty(symbol, "symbol must not be empty")
	return nil, nil
}

// GetTrades returns the trades for the given symbol.
// TODO(COD-17): Implement via E*TRADE accounts/transactions endpoint. https://linear.app/codegirl/issue/COD-17
func (e *etrade) GetTrades(symbol string) ([]Trade, error) {
	assert.Not_empty(symbol, "symbol must not be empty")
	return nil, nil
}

// get makes an OAuth-signed GET request to the ETrade API.
func (e *etrade) get(path string) ([]byte, error) {
	assert.Not_empty(path, "path must not be empty")
	assert.Not_nil(e.http_client, "http_client must not be nil")

	base_url := APIBaseURL(e.sandbox)
	url := fmt.Sprintf("%s%s", base_url, path)

	resp, err := e.http_client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode,
			string(body))
	}

	return body, nil
}

// post makes an OAuth-signed POST request to the ETrade API.
func (e *etrade) post(path string, content_type string,
	body io.Reader) ([]byte, error) {
	assert.Not_empty(path, "path must not be empty")
	assert.Not_empty(content_type, "content_type must not be empty")
	assert.Not_nil(body, "body must not be nil")
	assert.Not_nil(e.http_client, "http_client must not be nil")

	base_url := APIBaseURL(e.sandbox)
	url := fmt.Sprintf("%s%s", base_url, path)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}
	req.Header.Set("Content-Type", content_type)

	resp, err := e.http_client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	resp_body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode,
			string(resp_body))
	}

	return resp_body, nil
}

// ParseSandboxEnv parses the ETRADE_SANDBOX environment variable.
// Returns true if set to "true", false otherwise (defaults to false).
func ParseSandboxEnv() bool {
	val := strings.ToLower(strings.TrimSpace(os.Getenv("ETRADE_SANDBOX")))
	return val == "true" || val == "1"
}
