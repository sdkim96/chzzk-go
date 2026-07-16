//go:build integration

package chzzk

import (
	"context"
	"os"
	"testing"
)

func sessionClientAuth(t *testing.T) *Client {
	t.Helper()
	clientID := os.Getenv("CHZZK_CLIENT_ID")
	clientSecret := os.Getenv("CHZZK_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		t.Skip("CHZZK_CLIENT_ID or CHZZK_CLIENT_SECRET not set")
	}
	return New(nil).WithClientAuth(clientID, clientSecret)
}

func sessionAPIKey(t *testing.T) *Client {
	t.Helper()
	apiKey := os.Getenv("CHZZK_API_KEY")
	if apiKey == "" {
		t.Skip("CHZZK_API_KEY not set")
	}
	return New(nil).WithAPIKey(apiKey)
}

// --- AuthClient (WithClientAuth) ---

func Test_AuthClient_WithClientAuth(t *testing.T) {
	c := sessionClientAuth(t)

	url, err := c.Session.AuthClient(context.Background(), nil)
	if err != nil {
		t.Fatalf("AuthClient failed: %v", err)
	}
	if url == "" {
		t.Fatal("AuthClient returned empty URL")
	}
	t.Logf("AuthClient URL: %s", url)
}

// --- AuthClient (WithAPIKey) should fail ---

func Test_AuthClient_WithAPIKey(t *testing.T) {
	c := sessionAPIKey(t)

	_, err := c.Session.AuthClient(context.Background(), nil)
	if err == nil {
		t.Log("AuthClient with APIKey succeeded (unexpected — this API typically requires ClientAuth)")
	} else {
		t.Logf("AuthClient with APIKey failed as expected: %v", err)
	}
}

// --- AuthUser (WithAPIKey) ---

func Test_AuthUser_WithAPIKey(t *testing.T) {
	c := sessionAPIKey(t)

	url, err := c.Session.AuthUser(context.Background(), nil)
	if err != nil {
		t.Fatalf("AuthUser failed: %v", err)
	}
	if url == "" {
		t.Fatal("AuthUser returned empty URL")
	}
	t.Logf("AuthUser URL: %s", url)
}

// --- AuthUser (WithClientAuth) should fail ---

func Test_AuthUser_WithClientAuth(t *testing.T) {
	c := sessionClientAuth(t)

	_, err := c.Session.AuthUser(context.Background(), nil)
	if err == nil {
		t.Log("AuthUser with ClientAuth succeeded (unexpected — this API typically requires APIKey)")
	} else {
		t.Logf("AuthUser with ClientAuth failed as expected: %v", err)
	}
}
