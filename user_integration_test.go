//go:build integration

package chzzk

import (
	"context"
	"os"
	"testing"
)

func userClientAuth(t *testing.T) *Client {
	t.Helper()
	clientID := os.Getenv("CHZZK_CLIENT_ID")
	clientSecret := os.Getenv("CHZZK_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		t.Skip("CHZZK_CLIENT_ID or CHZZK_CLIENT_SECRET not set")
	}
	return New(nil).WithClientAuth(clientID, clientSecret)
}

func userAPIKey(t *testing.T) *Client {
	t.Helper()
	apiKey := os.Getenv("CHZZK_API_KEY")
	if apiKey == "" {
		t.Skip("CHZZK_API_KEY not set")
	}
	return New(nil).WithAPIKey(apiKey)
}

// --- Me (no auth) ---

func Test_Me_NoAuth(t *testing.T) {
	c := New(nil)
	_, err := c.User.Me(context.Background())
	if err == nil {
		t.Fatal("Me without auth should fail")
	}
	t.Logf("Me without auth failed as expected: %v", err)
}

// --- Me (WithAPIKey) ---

func Test_Me_WithAPIKey(t *testing.T) {
	c := userAPIKey(t)

	user, err := c.User.Me(context.Background())
	if err != nil {
		t.Fatalf("Me failed: %v", err)
	}
	if user.ChannelID == "" || user.ChannelName == "" {
		t.Fatalf("Me returned incomplete user: %+v", user)
	}
	t.Logf("User: %+v", user)
}

// --- Me (WithClientAuth) ---

func Test_Me_WithClientAuth(t *testing.T) {
	c := userClientAuth(t)

	user, err := c.User.Me(context.Background())
	if err == nil {
		t.Logf("Me with ClientAuth succeeded: %+v", user)
	} else {
		t.Logf("Me with ClientAuth failed as expected: %v", err)
	}
}
