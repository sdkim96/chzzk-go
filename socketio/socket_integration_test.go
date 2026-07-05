//go:build integration

// socketio_integration_test.go
package socketio_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	chzzk "github.com/sdkim96/chzzk-go"
	"github.com/sdkim96/chzzk-go/socketio"
)

// Test_Dial_socketio_echo tests handshake against a local socket.io echo server.
// Run with: SOCKETIO_URL=ws://localhost:3000 go test -run Integration ./socketio/...
func Test_Dial_socketio_echo_Integration(t *testing.T) {
	url := os.Getenv("SOCKETIO_URL")
	if url == "" {
		t.Skip("SOCKETIO_URL not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn := socketio.New(url + "/socket.io/?EIO=3&transport=websocket")
	if err := conn.Dial(ctx); err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer conn.Close(ctx, 1000, "")
}

// Test_Dial_chzzk_Integration tests against real Chzzk session API.
// Requires CHZZK_CLIENT_ID and CHZZK_CLIENT_SECRET env vars.
func Test_Dial_chzzk_Integration(t *testing.T) {
	clientID := os.Getenv("CHZZK_CLIENT_ID")
	clientSecret := os.Getenv("CHZZK_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		t.Skip("CHZZK_CLIENT_ID or CHZZK_CLIENT_SECRET not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Step 1: Get session URL from Chzzk API
	c := chzzk.New(nil).WithClientAuth(clientID, clientSecret)
	sessionURL, err := c.Session.AuthClient(ctx)
	if err != nil {
		t.Fatalf("AuthClient() error = %v", err)
	}
	t.Logf("Session URL: %s", sessionURL)

	// Step 2: Build Socket.IO WebSocket URL
	u, err := url.Parse(sessionURL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	u.Scheme = "wss"
	u.Path = "/socket.io/"
	q := u.Query()
	q.Set("EIO", fmt.Sprintf("%d", socketio.EngineIOVersion))
	q.Set("transport", "websocket")
	u.RawQuery = q.Encode()
	wsURL := u.String()
	t.Logf("WebSocket URL: %s", wsURL)

	// Step 3: Connect via Socket.IO
	received := make(chan []byte, 1)
	conn := socketio.New(wsURL,
		socketio.WithHandler("SYSTEM", func(data []byte) error {
			received <- data
			return nil
		}),
	)

	if err := conn.Dial(ctx); err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer conn.Close(ctx, 1000, "")

	go conn.Loop(ctx)

	select {
	case data := <-received:
		t.Logf("SYSTEM event received: %s", data)
	case <-ctx.Done():
		t.Fatal("timeout waiting for SYSTEM event")
	}
}
