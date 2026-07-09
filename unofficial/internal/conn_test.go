package internal

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
	if err := conn.Dial(ctx, "wss://echo.websocket.org"); err != nil {
		t.Fatalf("failed to dial: %v", err)
	}

	sendCh := make(chan []byte)

	recvCh, errCh, err := conn.Start(ctx, sendCh)
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	defer conn.Close(ctx, ws.StatusNormalClosure, "test done")

	for {
		time.Sleep(time.Second)
		t.Logf("Sending message...")
		select {
		case <-ctx.Done():
			close(sendCh)
			return
		case sendCh <- []byte("Hello, WebSocket!"):
			t.Log("message sent.")
		case err := <-errCh:
			close(sendCh)
			t.Fatalf("error from errCh: %v", err)
			return
		}
		select {
		case <-ctx.Done():
			close(sendCh)
			return
		case msg := <-recvCh:
			t.Logf("Received message: %s", string(msg))
		case err := <-errCh:
			close(sendCh)
			t.Fatalf("error from errCh: %v", err)
			return
		}
	}
}
