package clients

import (
	"aiplatform/pkg/assert"
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
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
	apiKey    string
	apiSecret string
}

// NewETrade creates a new etrade API client.
func NewETrade(apiKey, apiSecret string) ETrade {
	if apiKey == "" {
		err := godotenv.Load()
		if err == nil {
			apiKey = os.Getenv("ETRADE_API_KEY")
		}
	}
	if apiSecret == "" {
		err := godotenv.Load()
		if err == nil {
			apiSecret = os.Getenv("ETRADE_API_SECRET")
		}
	}
	assert.Not_empty(apiKey, "apiKey must not be empty")
	assert.Not_empty(apiSecret, "apiSecret must not be empty")
	return &etrade{
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

// GetOrders returns the orders for the given symbol.
func (e *etrade) GetOrders(symbol string) ([]Order, error) {
	assert.Not_empty(symbol, "symbol must not be empty")
	return nil, nil
}

// GetTrades returns the trades for the given symbol.
func (e *etrade) GetTrades(symbol string) ([]Trade, error) {
	assert.Not_empty(symbol, "symbol must not be empty")
	return nil, nil
}

// getAuthHeader returns an http.Header with the authorization header.
func (e *etrade) getAuthHeader() http.Header {
	assert.Not_empty(e.apiKey, "apiKey must not be empty")
	assert.Not_empty(e.apiSecret, "apiSecret must not be empty")
	auth := fmt.Sprintf("%s:%s", e.apiKey, e.apiSecret)
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	return http.Header{
		"Authorization": []string{fmt.Sprintf("Basic %s", encoded)},
	}
}

// get makes a GET request to the etrade API.
func (e *etrade) get(path string) ([]byte, error) {
	assert.Not_empty(path, "path must not be empty")

	url := fmt.Sprintf("https://etrade.com/api/v4/%s", path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = e.getAuthHeader()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("etrade API error: %s", body)
	}

	return body, nil
}

// post makes a POST request to the etrade API.
func (e *etrade) post(path string, body []byte) ([]byte, error) {
	assert.Not_empty(path, "path must not be empty")
	assert.Not_nil(body, "body must not be nil")

	url := fmt.Sprintf("https://etrade.com/api/v4/%s", path)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header = e.getAuthHeader()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("etrade API error: %s", respBody)
	}

	return respBody, nil
}
