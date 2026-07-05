package socketio

import (
	"context"
	"testing"

	ws "github.com/coder/websocket"
)

// Test_New tests that New initializes Conn with correct defaults.
func Test_New(t *testing.T) {
	url := "wss://example.com/socket.io/?EIO=3&transport=websocket"
	conn := New(url)

	if conn == nil {
		t.Fatal("New() returned nil")
	}
	if conn.url != url {
		t.Errorf("url = %q, want %q", conn.url, url)
	}
	if conn.handler == nil {
		t.Error("handler map should be initialized")
	}
	if conn.c == nil {
		t.Error("http.Client should be initialized")
	}
	// not dialed yet
	if conn.isDialed() {
		t.Error("isDialed() should be false before Dial()")
	}
}

// Test_WithHandler tests that WithHandler registers the handler correctly.
func Test_WithHandler(t *testing.T) {
	conn := New("wss://example.com",
		WithHandler("CHAT", func(p []byte) error { return nil }),
		WithHandler("SYSTEM", func(p []byte) error { return nil }),
	)

	if _, ok := conn.handler["CHAT"]; !ok {
		t.Error("CHAT handler should be registered")
	}
	if _, ok := conn.handler["SYSTEM"]; !ok {
		t.Error("SYSTEM handler should be registered")
	}
}

// Test_WithHandler_overwrite tests that registering the same pattern overwrites the previous handler.
func Test_WithHandler_overwrite(t *testing.T) {
	called := 0
	conn := New("wss://example.com",
		WithHandler("CHAT", func(p []byte) error {
			called = 1
			return nil
		}),
		WithHandler("CHAT", func(p []byte) error {
			called = 2
			return nil
		}),
	)

	conn.handler["CHAT"](nil)
	if called != 2 {
		t.Errorf("handler should be overwritten, called = %d, want 2", called)
	}
}

// Test_WithHandler_emptyPattern tests that WithHandler panics on empty pattern.
func Test_WithHandler_emptyPattern(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("WithHandler with empty pattern should panic")
		}
	}()
	WithHandler("", func(p []byte) error { return nil })
}

// Test_Loop_beforeDial tests that Loop returns error if called before Dial.
func Test_Loop_beforeDial(t *testing.T) {
	conn := New("wss://example.com")
	err := conn.Loop(context.Background())
	if err == nil {
		t.Error("Loop() should return error before Dial()")
	}
}

// Test_Close_beforeDial tests that Close returns error if called before Dial.
func Test_Close_beforeDial(t *testing.T) {
	conn := New("wss://example.com")
	err := conn.Close(context.Background(), ws.StatusNormalClosure, "")
	if err == nil {
		t.Error("Close() should return error before Dial()")
	}
}
