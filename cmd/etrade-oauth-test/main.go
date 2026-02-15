package main

import (
	"aiplatform/internals/clients"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("ETrade OAuth Demo (OOB Flow)")
	fmt.Println("=============================")
	fmt.Println()

	// Load .env (best-effort; not required if env vars already set).
	_ = godotenv.Load()

	// Get consumer key/secret from environment.
	consumer_key := strings.TrimSpace(os.Getenv("ETRADE_CONSUMER_KEY"))
	consumer_secret := strings.TrimSpace(os.Getenv("ETRADE_CONSUMER_SECRET"))

	if consumer_key == "" || consumer_secret == "" {
		fmt.Println("Error: ETRADE_CONSUMER_KEY and ETRADE_CONSUMER_SECRET must be set")
		fmt.Println()
		fmt.Println("Set these in your .env file or environment:")
		fmt.Println("  ETRADE_CONSUMER_KEY=your_key")
		fmt.Println("  ETRADE_CONSUMER_SECRET=your_secret")
		fmt.Println("  ETRADE_SANDBOX=true   # optional, defaults to true")
		os.Exit(1)
	}

	// Determine sandbox vs production from env (default: sandbox).
	sandbox := clients.ParseSandboxEnv()
	if sandbox {
		fmt.Println("Environment: sandbox")
	} else {
		fmt.Println("Environment: production")
	}
	fmt.Println()

	// Determine workspace root (absolute path to repo root).
	workspace_root, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error: failed to get current directory: %v\n", err)
		os.Exit(1)
	}
	workspace_root, err = filepath.Abs(workspace_root)
	if err != nil {
		fmt.Printf("Error: failed to resolve absolute path: %v\n", err)
		os.Exit(1)
	}
	workspace_root, err = filepath.EvalSymlinks(workspace_root)
	if err != nil {
		fmt.Printf("Error: failed to resolve symlinks: %v\n", err)
		os.Exit(1)
	}

	token_path := filepath.Join(workspace_root, ".aiplatform",
		"credentials", "etrade_tokens.json")
	fmt.Printf("Workspace: %s\n", workspace_root)
	fmt.Printf("Token storage: %s\n", token_path)
	fmt.Println()

	// Check if we have a saved token.
	access_token, access_secret, _, expires_at, err :=
		clients.LoadETradeToken(workspace_root, sandbox)

	if err != nil || access_token == "" {
		// No token or expired/invalid; run OOB flow.
		if err != nil {
			fmt.Printf("Token load issue: %v\n", err)
		} else {
			fmt.Println("No saved token found")
		}
		fmt.Println("Starting OAuth authentication flow...")
		fmt.Println()

		access_token, access_secret, err = run_oauth_flow(
			consumer_key, consumer_secret, sandbox)
		if err != nil {
			fmt.Printf("Error: OAuth flow failed: %v\n", err)
			os.Exit(1)
		}

		// Save token (ETrade tokens expire at midnight US Eastern).
		expires_at = calculate_token_expiry()
		clients.SaveETradeToken(workspace_root, access_token,
			access_secret, sandbox, expires_at)

		fmt.Println("Token saved successfully")
		fmt.Printf("Token expires at: %s\n",
			expires_at.Format("2006-01-02 15:04:05 MST"))
		fmt.Println()
	} else {
		fmt.Println("Using saved token")
		fmt.Printf("Token expires at: %s\n",
			expires_at.Format("2006-01-02 15:04:05 MST"))
		fmt.Println()
	}

	// Test API call: list accounts.
	fmt.Println("Testing API call: GET /v1/accounts/list")
	fmt.Println()

	config := clients.NewOAuthConfig(consumer_key, consumer_secret, sandbox)
	http_client := clients.NewOAuthClient(config, access_token, access_secret)

	base_url := clients.APIBaseURL(sandbox)
	accounts_url := fmt.Sprintf("%s/v1/accounts/list", base_url)

	resp, err := http_client.Get(accounts_url)
	if err != nil {
		fmt.Printf("Error: API call failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("Response status: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Println()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Response body: %s\n", string(body))
		os.Exit(1)
	}

	// Read and print response body (likely XML).
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Response body:")
	fmt.Println(string(body))
	fmt.Println()

	fmt.Println("=============================")
	fmt.Println("ETrade OAuth demo completed successfully!")
}

// run_oauth_flow executes the OOB OAuth flow.
// Returns (access_token, access_secret, error).
func run_oauth_flow(consumer_key, consumer_secret string,
	sandbox bool) (string, string, error) {

	config := clients.NewOAuthConfig(consumer_key, consumer_secret, sandbox)

	// Step 1: Get request token.
	fmt.Println("Step 1: Requesting OAuth token from ETrade...")
	request_token, request_secret, err := clients.RequestToken(config)
	if err != nil {
		return "", "", fmt.Errorf("failed to get request token: %w", err)
	}
	fmt.Println("✓ Request token obtained")
	fmt.Println()

	// Step 2: Build authorization URL and show to user.
	auth_url := clients.AuthorizationURL(config, request_token)
	fmt.Println("Step 2: Authorize this application")
	fmt.Println()
	fmt.Println(clients.OAuthHelperMessage())
	fmt.Println("Authorization URL:")
	fmt.Println(auth_url)
	fmt.Println()

	// Step 3: Read verifier from user (OOB).
	fmt.Print("Enter the verification code: ")
	reader := bufio.NewReader(os.Stdin)
	verifier, err := reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("failed to read verifier: %w", err)
	}
	verifier = strings.TrimSpace(verifier)
	if verifier == "" {
		return "", "", fmt.Errorf("verifier cannot be empty")
	}
	fmt.Println()

	// Step 4: Exchange verifier for access token.
	fmt.Println("Step 3: Exchanging verifier for access token...")
	access_token, access_secret, err := clients.ExchangeToken(config,
		request_token, request_secret, verifier)
	if err != nil {
		return "", "", fmt.Errorf("failed to exchange token: %w", err)
	}
	fmt.Println("✓ Access token obtained")
	fmt.Println()

	return access_token, access_secret, nil
}

// calculate_token_expiry returns the token expiry time.
// ETrade tokens expire at midnight US Eastern time.
// We compute next midnight US/Eastern minus a safety margin.
func calculate_token_expiry() time.Time {
	// Load US/Eastern timezone.
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		// Fall back to conservative 1-hour TTL if timezone unavailable.
		return time.Now().Add(1 * time.Hour)
	}

	now_eastern := time.Now().In(location)

	// Calculate next calendar day's midnight in US/Eastern.
	tomorrow := now_eastern.AddDate(0, 0, 1)
	next_midnight_eastern := time.Date(
		tomorrow.Year(),
		tomorrow.Month(),
		tomorrow.Day(),
		0, 0, 0, 0,
		location,
	)

	// Apply 5-minute safety margin to avoid using token after real expiry.
	const safety_margin = 5 * time.Minute
	expiry_eastern := next_midnight_eastern.Add(-safety_margin)

	// Ensure computed expiry is in the future; otherwise fall back.
	if !expiry_eastern.After(now_eastern) {
		return time.Now().Add(1 * time.Hour)
	}

	// Return in UTC to avoid timezone surprises.
	return expiry_eastern.UTC()
}
