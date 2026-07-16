//go:build integration

package chzzk

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"
)

func Test_SessionConnect_WithClientAuth_Integration(t *testing.T) {
	clientID := os.Getenv("CHZZK_CLIENT_ID")
	clientSecret := os.Getenv("CHZZK_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		t.Skip("CHZZK_CLIENT_ID or CHZZK_CLIENT_SECRET not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c := New(nil).WithClientAuth(clientID, clientSecret)
	sessionURL, err := c.Session.AuthClient(ctx, nil)
	if err != nil {
		t.Fatalf("AuthClient() error = %v", err)
	}
	t.Logf("Session URL: %s", sessionURL)

	u, err := url.Parse(sessionURL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	u.Scheme = "wss"
	u.Path = "/socket.io/"
	q := u.Query()
	q.Set("EIO", fmt.Sprintf("%d", EngineIOVersion))
	q.Set("transport", "websocket")
	u.RawQuery = q.Encode()
	wsURL := u.String()
	t.Logf("WebSocket URL: %s", wsURL)

	received := make(chan []byte, 1)
	go func() {
		c.Session.Connect(ctx, wsURL, map[string]Handler{
			"SYSTEM": func(data []byte) error {
				received <- data
				return nil
			},
		})
	}()

	select {
	case data := <-received:
		t.Logf("SYSTEM event received: %s", data)
	case <-ctx.Done():
		t.Fatal("timeout waiting for SYSTEM event")
	}
}

func Test_SessionConnect_WithAPIKey_Integration(t *testing.T) {
	apiKey := os.Getenv("CHZZK_API_KEY")
	if apiKey == "" {
		t.Skip("CHZZK_API_KEY not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c := New(nil).WithAPIKey(apiKey)
	sessionURL, err := c.Session.AuthUser(ctx, nil)
	if err != nil {
		t.Fatalf("AuthUser() error = %v", err)
	}
	t.Logf("Session URL: %s", sessionURL)

	u, err := url.Parse(sessionURL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	u.Scheme = "wss"
	u.Path = "/socket.io/"
	q := u.Query()
	q.Set("EIO", fmt.Sprintf("%d", EngineIOVersion))
	q.Set("transport", "websocket")
	u.RawQuery = q.Encode()
	wsURL := u.String()
	t.Logf("WebSocket URL: %s", wsURL)

	received := make(chan []byte, 1)
	go func() {
		c.Session.Connect(ctx, wsURL, map[string]Handler{
			"SYSTEM": func(data []byte) error {
				received <- data
				return nil
			},
		})
	}()

	select {
	case data := <-received:
		t.Logf("SYSTEM event received: %s", data)
	case <-ctx.Done():
		t.Fatal("timeout waiting for SYSTEM event")
	}
}
