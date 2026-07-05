package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sdkim96/chzzk-go"
	"github.com/sdkim96/chzzk-go/internal/login"
)

func main() {
	clientID, clientSecret := readCredentials()

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	state := fmt.Sprintf("chzzk-go-login-%d", time.Now().Unix())
	code := authorize(ctx, clientID, state)

	tokenResp := exchangeToken(ctx, clientID, clientSecret, code, state)

	printResult(tokenResp)
}

func readCredentials() (clientID, clientSecret string) {
	fmt.Println("=== chzzk-go login ===")
	fmt.Println()
	fmt.Println("Before you start, make sure you have:")
	fmt.Println("  1. Created an app at https://developers.chzzk.naver.com")
	fmt.Println("  2. Registered the redirect URL: http://localhost:57777/callback")
	fmt.Println()

	fmt.Print("Continue? (y/n): ")
	var answer string
	fmt.Scanln(&answer)
	if answer != "y" {
		fmt.Println("Aborted.")
		os.Exit(0)
	}
	fmt.Println()

	clientID = os.Getenv("CHZZK_CLIENT_ID")
	if clientID == "" {
		fmt.Print("Client ID: ")
		fmt.Scanln(&clientID)
	} else {
		fmt.Printf("Client ID: %s (from env)\n", clientID)
	}

	clientSecret = os.Getenv("CHZZK_CLIENT_SECRET")
	if clientSecret == "" {
		fmt.Print("Client Secret: ")
		fmt.Scanln(&clientSecret)
	} else {
		fmt.Printf("Client Secret: %s (from env)\n", clientSecret)
	}

	if clientID == "" || clientSecret == "" {
		log.Fatal("Client ID and Client Secret are required.")
	}
	fmt.Println()
	return clientID, clientSecret
}

func authorize(ctx context.Context, clientID, state string) string {

	authURL := login.URL(clientID, login.RedirectURI, state)

	fmt.Println("--- Step 1: Authorize ---")
	fmt.Println("Open this URL in your browser:")
	fmt.Println()
	fmt.Printf("  %s\n", authURL)
	fmt.Println()
	fmt.Println("Waiting for callback...")

	codeCh := make(chan string, 1)
	go func() {
		if err := login.Server(ctx, codeCh); err != nil {
			log.Fatalf("Login server error: %v", err)
		}
	}()

	select {
	case code := <-codeCh:
		fmt.Printf("Authorization code received.\n\n")
		return code
	case <-ctx.Done():
		log.Fatal("Timeout waiting for login.")
		return ""
	}
}

func exchangeToken(ctx context.Context, clientID, clientSecret, code, state string) *chzzk.TokenResponse {
	fmt.Println("--- Step 2: Exchange Token ---")

	c := chzzk.New(nil).WithClientAuth(clientID, clientSecret)
	tokenResp, err := c.Token.NewToken(ctx, chzzk.TokenNewRequest{
		TokenRequest: chzzk.TokenRequest{
			GrantType:    chzzk.GrantTypeAuthorizationCode,
			ClientID:     clientID,
			ClientSecret: clientSecret,
		},
		Code:  code,
		State: state,
	})
	if err != nil {
		log.Fatalf("Token exchange failed: %v", err)
	}
	fmt.Printf("Token received.\n\n")
	return tokenResp
}

func printResult(resp *chzzk.TokenResponse) {
	fmt.Println("=== Login Complete ===")
	fmt.Println()
	fmt.Printf("  Access Token:  %s\n", resp.AccessToken)
	fmt.Printf("  Refresh Token: %s\n", resp.RefreshToken)
	fmt.Printf("  Expires In:    %ds\n", resp.ExpiresIn)
	fmt.Println()
	fmt.Println("Example usage:")
	fmt.Println()
	fmt.Printf("  c := chzzk.New(nil).WithAPIKey(\"%s\")\n", resp.AccessToken)
}
