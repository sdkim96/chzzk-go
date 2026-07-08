package unofficial

import (
	"context"
	"net/http"
	"testing"
	"time"

	ws "github.com/coder/websocket"
)

func Test_ChatConn(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	conn := NewConn(http.DefaultClient)
	if err := conn.Dial(ctx, "wss://echo.websocket.org"); err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
}

func Test_ChatConn_Loop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	conn := NewConn(http.DefaultClient)
	if err := conn.Dial(ctx, "wss://echo.websocket.org"); err != nil {
		t.Fatalf("failed to dial: %v", err)
	}

	recvCh := make(chan []byte)
	sendCh := make(chan []byte)
	errCh := make(chan error)

	conn.Loop(ctx, recvCh, sendCh, errCh)

	for {
		time.Sleep(time.Second)
		t.Logf("Sending message...")
		select {
		case <-ctx.Done():
			t.Log("context done")
			conn.closeCh <- struct {
				code   ws.StatusCode
				reason string
			}{code: ws.StatusNormalClosure, reason: "test done"}
			conn.Close(context.Background())
			return
		case sendCh <- []byte("Hello, WebSocket!"):
			t.Log("MEssage sent")
		case err := <-errCh:
			t.Logf("loop error: %v", err)
		}

		select {
		case <-ctx.Done():
			t.Log("context done")
			conn.closeCh <- struct {
				code   ws.StatusCode
				reason string
			}{code: ws.StatusNormalClosure, reason: "test done"}
			conn.Close(context.Background())
			return
		case err := <-errCh:
			t.Fatalf("loop error: %v", err)
		case msg := <-recvCh:
			t.Logf("received message: %s", string(msg))
		}
	}
}
