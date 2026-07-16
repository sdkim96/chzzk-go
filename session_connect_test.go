package chzzk

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ws "github.com/coder/websocket"
)

// socketIOEcho is a minimal Socket.IO v2 test server that:
// 1. Sends Open packet (0)
// 2. Waits for Connect packet (40) and responds with Connect (40)
// 3. Sends a ping (2) every pingInterval
// 4. Sends the given events then closes
func socketIOEcho(t *testing.T, events []string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := ws.Accept(w, r, nil)
		if err != nil {
			t.Logf("ws accept: %v", err)
			return
		}
		defer c.CloseNow()

		ctx := r.Context()

		// Phase 1: Send Open (0)
		err = c.Write(ctx, ws.MessageText, []byte(`0{"sid":"test-sid","pingInterval":25000,"pingTimeout":5000}`))
		if err != nil {
			return
		}

		// Phase 2: Wait for Connect (40)
		_, msg, err := c.Read(ctx)
		if err != nil {
			return
		}
		if string(msg) != "40" {
			t.Logf("expected 40, got %q", msg)
			return
		}

		// Phase 3: Respond with Connect (40)
		err = c.Write(ctx, ws.MessageText, []byte("40"))
		if err != nil {
			return
		}

		// Phase 4: Send events
		for _, ev := range events {
			err = c.Write(ctx, ws.MessageText, []byte(ev))
			if err != nil {
				return
			}
		}

		// Phase 5: Send a ping
		err = c.Write(ctx, ws.MessageText, []byte("2"))
		if err != nil {
			return
		}

		// Wait for pong (3)
		_, msg, err = c.Read(ctx)
		if err != nil {
			return
		}
		if string(msg) != "3" {
			t.Logf("expected pong (3), got %q", msg)
		}

		// Close gracefully
		c.Close(ws.StatusNormalClosure, "done")
	})
}

func Test_SessionConnect_EmptyHandlerKey(t *testing.T) {
	s := &SessionService{c: &Client{httpClient: &http.Client{}}}
	err := s.Connect(context.Background(), "wss://example.com", map[string]Handler{
		"": func(data []byte) error { return nil },
	})
	if err != ErrEmptyHandlerKey {
		t.Errorf("expected ErrEmptyHandlerKey, got %v", err)
	}
}

func Test_SessionConnect_NilHandlerSkipped(t *testing.T) {
	srv := httptest.NewServer(socketIOEcho(t, nil))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/socket.io/?EIO=3&transport=websocket"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	s := &SessionService{c: &Client{httpClient: srv.Client()}}
	// nil handlers should be silently skipped, not cause a panic
	err := s.Connect(ctx, wsURL, map[string]Handler{
		"CHAT": nil,
	})
	// Connection will close after server sends close — that's expected
	if err != nil {
		t.Logf("Connect returned (expected after server close): %v", err)
	}
}

func Test_SessionConnect_HandlerKeyUppercased(t *testing.T) {
	received := make(chan string, 1)

	srv := httptest.NewServer(socketIOEcho(t, []string{
		`42["CHAT",{"content":"hello"}]`,
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/socket.io/?EIO=3&transport=websocket"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	s := &SessionService{c: &Client{httpClient: srv.Client()}}
	err := s.Connect(ctx, wsURL, map[string]Handler{
		"chat": func(data []byte) error { // lowercase — should be uppercased
			received <- string(data)
			return nil
		},
	})
	// err is expected after server closes
	if err != nil {
		t.Logf("Connect returned: %v", err)
	}

	select {
	case data := <-received:
		t.Logf("Received: %s", data)
		if !strings.Contains(data, "hello") {
			t.Errorf("expected 'hello' in data, got %q", data)
		}
	default:
		t.Error("handler was not called — lowercase key should have matched CHAT event")
	}
}

func Test_SessionConnect_MultipleEvents(t *testing.T) {
	events := []string{
		`42["SYSTEM",{"type":"connected","data":{"sessionKey":"sk-123"}}]`,
		`42["CHAT",{"content":"test message"}]`,
		`42["DONATION",{"payAmount":"1000"}]`,
	}

	received := make(map[string][]byte)
	done := make(chan struct{})
	count := 0

	srv := httptest.NewServer(socketIOEcho(t, events))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/socket.io/?EIO=3&transport=websocket"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	makeHandler := func(name string) Handler {
		return func(data []byte) error {
			received[name] = data
			count++
			if count == 3 {
				close(done)
			}
			return nil
		}
	}

	s := &SessionService{c: &Client{httpClient: srv.Client()}}
	go func() {
		s.Connect(ctx, wsURL, map[string]Handler{
			"SYSTEM":   makeHandler("SYSTEM"),
			"CHAT":     makeHandler("CHAT"),
			"DONATION": makeHandler("DONATION"),
		})
	}()

	select {
	case <-done:
		// all 3 received
	case <-ctx.Done():
		t.Fatal("timeout waiting for events")
	}

	for _, name := range []string{"SYSTEM", "CHAT", "DONATION"} {
		if _, ok := received[name]; !ok {
			t.Errorf("missing event: %s", name)
		} else {
			t.Logf("%s: %s", name, received[name])
		}
	}
}

func Test_SessionConnect_HandlerError(t *testing.T) {
	srv := httptest.NewServer(socketIOEcho(t, []string{
		`42["CHAT",{"content":"bad"}]`,
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/socket.io/?EIO=3&transport=websocket"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	handlerErr := fmt.Errorf("handler failed intentionally")
	s := &SessionService{c: &Client{httpClient: srv.Client()}}
	err := s.Connect(ctx, wsURL, map[string]Handler{
		"CHAT": func(data []byte) error {
			return handlerErr
		},
	})
	if err == nil {
		t.Fatal("expected error from handler, got nil")
	}
	if !strings.Contains(err.Error(), "handler failed") {
		t.Errorf("expected handler error, got %v", err)
	}
}

func Test_SessionConnect_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := ws.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer c.CloseNow()
		// Send Open
		c.Write(r.Context(), ws.MessageText, []byte(`0{"sid":"s","pingInterval":25000,"pingTimeout":5000}`))
		// Wait for Connect
		c.Read(r.Context())
		// Send Connect
		c.Write(r.Context(), ws.MessageText, []byte("40"))
		// Block indefinitely
		<-r.Context().Done()
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/socket.io/?EIO=3&transport=websocket"

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	s := &SessionService{c: &Client{httpClient: srv.Client()}}
	err := s.Connect(ctx, wsURL, map[string]Handler{
		"CHAT": func(data []byte) error { return nil },
	})
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	t.Logf("Connect returned: %v", err)
}
