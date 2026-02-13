package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jerryryle/etrade-cli/pkg/etradelib/client"
	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

func main() {
	fmt.Println("ETrade OAuth Test using etradelib")
	fmt.Println("==================================")
	fmt.Println()

	if err := godotenv.Load(); err != nil {
		fmt.Println("Note: .env not found, relying on environment variables")
	}

	consumerKey, err := requireEnv("ETRADE_CONSUMER_KEY")
	if err != nil {
		fmt.Printf("✗ %v\n", err)
		os.Exit(1)
	}

	consumerSecret, err := requireEnv("ETRADE_CONSUMER_SECRET")
	if err != nil {
		fmt.Printf("✗ %v\n", err)
		os.Exit(1)
	}

	// Create a logger (discard output for simplicity)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create ETrade client for sandbox (production=false)
	fmt.Println("Step 1: Creating ETrade client...")
	etradeClient, err := client.CreateETradeClient(logger, false, consumerKey, consumerSecret, "", "")
	if err != nil {
		fmt.Printf("✗ Failed to create client: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Client created")
	fmt.Println()

	// Start authentication
	fmt.Println("Step 2: Starting authentication...")
	authResponse, err := etradeClient.Authenticate()
	if err != nil {
		fmt.Printf("✗ Failed to authenticate: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Got authorization URL")
	fmt.Println()

	// Parse the authorization URL from response
	var authData map[string]interface{}
	if err := json.Unmarshal(authResponse, &authData); err != nil {
		fmt.Printf("✗ Failed to parse auth response: %v\n", err)
		os.Exit(1)
	}

	authURL, ok := authData["authorizationUrl"].(string)
	if !ok {
		fmt.Println("✗ No authorization URL in response")
		fmt.Printf("Response: %s\n", string(authResponse))
		os.Exit(1)
	}

	// Show URL to user
	fmt.Println("Step 3: Open this URL in your browser:")
	fmt.Println(authURL)
	fmt.Println()
	fmt.Println("Instructions:")
	fmt.Println("1. Log in with your ETrade sandbox credentials")
	fmt.Println("2. Click 'Accept' to authorize this app")
	fmt.Println("3. Copy the verification code shown")
	fmt.Println()

	// Get verification code from user
	fmt.Print("Enter the verification code: ")
	reader := bufio.NewReader(os.Stdin)
	verifyCode, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("✗ Failed to read verification code: %v\n", err)
		os.Exit(1)
	}
	verifyCode = strings.TrimSpace(verifyCode)
	if verifyCode == "" {
		fmt.Println("✗ Verification code cannot be empty")
		os.Exit(1)
	}
	fmt.Println()

	// Complete authentication
	fmt.Println("Step 4: Completing authentication...")
	verifyResponse, err := etradeClient.Verify(verifyCode)
	if err != nil {
		fmt.Printf("✗ Failed to verify: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Authentication complete")
	fmt.Printf("Response: %s\n", string(verifyResponse))
	fmt.Println()

	// Test API call - list accounts
	fmt.Println("Step 5: Testing API call - listing accounts...")
	accountsResponse, err := etradeClient.ListAccounts()
	if err != nil {
		fmt.Printf("✗ Failed to list accounts: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ API call successful!")
	fmt.Println()

	// Parse and display accounts
	var accountsData map[string]interface{}
	if err := json.Unmarshal(accountsResponse, &accountsData); err != nil {
		fmt.Printf("Response (raw): %s\n", string(accountsResponse))
	} else {
		fmt.Println("Accounts response:")
		prettyJSON, _ := json.MarshalIndent(accountsData, "", "  ")
		fmt.Println(string(prettyJSON))
	}
	fmt.Println()
	fmt.Println("==================================")
	fmt.Println("ETrade OAuth test completed successfully!")
}

func requireEnv(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("missing required environment variable: %s", key)
	}
	return value, nil
}
