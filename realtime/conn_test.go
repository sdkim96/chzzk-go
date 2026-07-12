package realtime

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
	defer conn.Close(ctx, ws.StatusNormalClosure, "test done")
}

func Test_ChatConn_Loop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	conn := NewConn(http.DefaultClient)
	defer conn.Close(ctx, ws.StatusNormalClosure, "test done")
	if err := conn.Dial(ctx, "wss://echo.websocket.org"); err != nil {
		t.Fatalf("failed to dial: %v", err)
	}

	recvCh := make(chan []byte)
	sendCh := make(chan []byte)
	errCh := make(chan error, 1)

	go func() {
		err := conn.Loop(ctx, recvCh, sendCh)
		if err != nil {
			errCh <- err
		}
	}()

	for {
		time.Sleep(time.Second)
		t.Logf("Sending message...")
		select {
		case <-ctx.Done():
			close(recvCh)
			close(sendCh)
			close(errCh)
		case sendCh <- []byte("Hello, WebSocket!"):
			t.Logf("Message sent.")
		case err := <-errCh:
			t.Fatalf("Loop error: %v", err)
		}
		select {
		case <-ctx.Done():
			close(recvCh)
			close(sendCh)
			close(errCh)
		case msg := <-recvCh:
			t.Logf("Received message: %s", string(msg))
		case err := <-errCh:
			t.Fatalf("Loop error: %v", err)
		}
	}
}
