//go:build integration

package chzzk

import (
	"context"
	"os"
	"testing"
)

func tokenClientAuth(t *testing.T) *Client {
	t.Helper()
	clientID := os.Getenv("CHZZK_CLIENT_ID")
	clientSecret := os.Getenv("CHZZK_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		t.Skip("CHZZK_CLIENT_ID or CHZZK_CLIENT_SECRET not set")
	}
	return New(nil).WithClientAuth(clientID, clientSecret)
}

func tokenAPIKey(t *testing.T) *Client {
	t.Helper()
	apiKey := os.Getenv("CHZZK_API_KEY")
	if apiKey == "" {
		t.Skip("CHZZK_API_KEY not set")
	}
	return New(nil).WithAPIKey(apiKey)
}

// --- New (WithClientAuth) ---
// Note: This test requires a valid authorization code, which is short-lived.
// It will fail with an expired code. Skip if CHZZK_AUTH_CODE is not set.

func Test_New_WithClientAuth(t *testing.T) {
	c := tokenClientAuth(t)
	code := os.Getenv("CHZZK_AUTH_CODE")
	if code == "" {
		t.Skip("CHZZK_AUTH_CODE not set (requires fresh OAuth code)")
	}

	resp, err := c.Token.New(context.Background(), TokenNewRequest{
		TokenRequest: TokenRequest{
			GrantType:    GrantTypeAuthorizationCode,
			ClientID:     os.Getenv("CHZZK_CLIENT_ID"),
			ClientSecret: os.Getenv("CHZZK_CLIENT_SECRET"),
		},
		Code:  code,
		State: "test-state",
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("New returned empty access token")
	}
	t.Logf("New response: %+v", resp)
}

// --- New (WithAPIKey) should fail ---

func Test_New_WithAPIKey(t *testing.T) {
	c := tokenAPIKey(t)

	_, err := c.Token.New(context.Background(), TokenNewRequest{
		TokenRequest: TokenRequest{
			GrantType:    GrantTypeAuthorizationCode,
			ClientID:     "dummy",
			ClientSecret: "dummy",
		},
		Code:  "dummy",
		State: "test-state",
	})
	if err == nil {
		t.Log("New with APIKey succeeded (unexpected — this API typically requires ClientAuth)")
	} else {
		t.Logf("New with APIKey failed as expected: %v", err)
	}
}

// --- Refresh (WithClientAuth) ---
// Note: Requires a valid refresh token. Skip if CHZZK_REFRESH_TOKEN is not set.

func Test_Refresh_WithClientAuth(t *testing.T) {
	c := tokenClientAuth(t)
	refreshToken := os.Getenv("CHZZK_REFRESH_TOKEN")
	if refreshToken == "" {
		t.Skip("CHZZK_REFRESH_TOKEN not set (requires valid refresh token)")
	}

	resp, err := c.Token.Refresh(context.Background(), TokenRefreshRequest{
		TokenRequest: TokenRequest{
			GrantType:    GrantTypeRefreshToken,
			ClientID:     os.Getenv("CHZZK_CLIENT_ID"),
			ClientSecret: os.Getenv("CHZZK_CLIENT_SECRET"),
		},
		RefreshToken: refreshToken,
	})
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("Refresh returned empty access token")
	}
	t.Logf("Refresh response: %+v", resp)
}

// --- Refresh (WithAPIKey) should fail ---

func Test_Refresh_WithAPIKey(t *testing.T) {
	c := tokenAPIKey(t)

	_, err := c.Token.Refresh(context.Background(), TokenRefreshRequest{
		TokenRequest: TokenRequest{
			GrantType:    GrantTypeRefreshToken,
			ClientID:     "dummy",
			ClientSecret: "dummy",
		},
		RefreshToken: "dummy",
	})
	if err == nil {
		t.Log("Refresh with APIKey succeeded (unexpected — this API typically requires ClientAuth)")
	} else {
		t.Logf("Refresh with APIKey failed as expected: %v", err)
	}
}

// --- Revoke (WithClientAuth) ---
// Note: Requires a valid token to revoke. Skip if CHZZK_REVOKE_TOKEN is not set.

func Test_Revoke_WithClientAuth(t *testing.T) {
	c := tokenClientAuth(t)
	token := os.Getenv("CHZZK_REVOKE_TOKEN")
	if token == "" {
		t.Skip("CHZZK_REVOKE_TOKEN not set (requires valid token to revoke)")
	}

	err := c.Token.Revoke(context.Background(), RevokeTokenRequest{
		ClientID:      os.Getenv("CHZZK_CLIENT_ID"),
		ClientSecret:  os.Getenv("CHZZK_CLIENT_SECRET"),
		Token:         token,
		TokenTypeHint: "access_token",
	})
	if err != nil {
		t.Fatalf("Revoke failed: %v", err)
	}
	t.Log("Revoke succeeded")
}

// --- Revoke (WithAPIKey) should fail ---

func Test_Revoke_WithAPIKey(t *testing.T) {
	c := tokenAPIKey(t)

	err := c.Token.Revoke(context.Background(), RevokeTokenRequest{
		ClientID:      "dummy",
		ClientSecret:  "dummy",
		Token:         "dummy",
		TokenTypeHint: "access_token",
	})
	if err == nil {
		t.Log("Revoke with APIKey succeeded (unexpected — this API typically requires ClientAuth)")
	} else {
		t.Logf("Revoke with APIKey failed as expected: %v", err)
	}
}
