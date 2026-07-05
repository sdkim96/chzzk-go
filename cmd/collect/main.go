package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	chzzk "github.com/sdkim96/chzzk-go"
	"github.com/sdkim96/chzzk-go/internal/login"
	"github.com/sdkim96/chzzk-go/socketio"
)

func main() {
	clientID := os.Getenv("CHZZK_CLIENT_ID")
	clientSecret := os.Getenv("CHZZK_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal("CHZZK_CLIENT_ID and CHZZK_CLIENT_SECRET must be set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	c := chzzk.New(nil).WithClientAuth(clientID, clientSecret)

	// Step 1: OAuth login to get access token
	state := fmt.Sprintf("collect-%d", time.Now().Unix())
	authURL := login.URL(clientID, login.RedirectURI, state)
	fmt.Println("Open this URL in your browser to login:")
	fmt.Println(authURL)

	codeCh := make(chan string, 1)
	go func() {
		if err := login.Server(ctx, codeCh); err != nil {
			log.Fatalf("Login server: %v", err)
		}
	}()

	var code string
	select {
	case code = <-codeCh:
		log.Printf("Got authorization code: %s", code)
	case <-ctx.Done():
		log.Fatal("timeout waiting for login")
	}

	// Step 2: Exchange code for access token
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
		log.Fatalf("NewToken: %v", err)
	}
	log.Printf("Got access token (expires in %ds)", tokenResp.ExpiresIn)

	// Step 3: Create user session with access token
	cUser := chzzk.New(nil).WithAPIKey(tokenResp.AccessToken)
	sessionURL, err := cUser.Session.AuthUser(ctx)
	if err != nil {
		log.Fatalf("AuthUser: %v", err)
	}
	log.Printf("Session URL: %s", sessionURL)

	// Step 4: Build WebSocket URL
	u, err := url.Parse(sessionURL)
	if err != nil {
		log.Fatalf("url.Parse: %v", err)
	}
	u.Scheme = "wss"
	u.Path = "/socket.io/"
	q := u.Query()
	q.Set("EIO", fmt.Sprintf("%d", socketio.EngineIOVersion))
	q.Set("transport", "websocket")
	u.RawQuery = q.Encode()
	wsURL := u.String()
	log.Printf("WebSocket URL: %s", wsURL)

	// Collect messages
	var mu sync.Mutex
	var messages []json.RawMessage

	sessionKeyC := make(chan string, 1)

	conn := socketio.New(wsURL,
		socketio.WithHandler("SYSTEM", func(data []byte) error {
			log.Printf("SYSTEM: %s", data)
			var raw string
			if err := json.Unmarshal(data, &raw); err == nil {
				var sys struct {
					Type string `json:"type"`
					Data struct {
						SessionKey string `json:"sessionKey"`
					} `json:"data"`
				}
				if err := json.Unmarshal([]byte(raw), &sys); err == nil && sys.Data.SessionKey != "" {
					select {
					case sessionKeyC <- sys.Data.SessionKey:
					default:
					}
				}
			}
			return nil
		}),
		socketio.WithHandler("CHAT", func(data []byte) error {
			log.Printf("CHAT: %s", data)
			mu.Lock()
			messages = append(messages, json.RawMessage(data))
			mu.Unlock()
			return nil
		}),
		socketio.WithHandler("DONATION", func(data []byte) error {
			log.Printf("DONATION: %s", data)
			mu.Lock()
			messages = append(messages, json.RawMessage(data))
			mu.Unlock()
			return nil
		}),
	)

	// Step 5: Connect via Socket.IO
	if err := conn.Dial(ctx); err != nil {
		log.Fatalf("Dial: %v", err)
	}
	defer conn.Close(ctx, 1000, "done")

	loopErr := make(chan error, 1)
	go func() { loopErr <- conn.Loop(ctx) }()

	// Step 6: Wait for sessionKey
	var sessionKey string
	select {
	case sessionKey = <-sessionKeyC:
		log.Printf("Got sessionKey: %s", sessionKey)
	case err := <-loopErr:
		log.Fatalf("Loop ended before sessionKey: %v", err)
	case <-ctx.Done():
		log.Fatal("timeout waiting for sessionKey")
	}

	// Step 7: Subscribe to chat using user access token
	if err := cUser.Session.SubscribeChat(ctx, sessionKey); err != nil {
		log.Fatalf("SubscribeChat: %v", err)
	}

	// Step 8: Read for 20 seconds
	timer := time.NewTimer(20 * time.Second)
	select {
	case <-timer.C:
		log.Println("20 seconds elapsed, saving messages...")
	case err := <-loopErr:
		log.Printf("Loop ended: %v", err)
	}

	// Save collected messages
	mu.Lock()
	defer mu.Unlock()

	out, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		log.Fatalf("Marshal: %v", err)
	}
	log.Printf("Collected %d messages:\n%s\n", len(messages), out)
}
