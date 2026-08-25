// package socket provides a WebSocket client that has a full-duplex communication channel over a single TCP connection.
package socket

import (
	"context"
	"fmt"
	"net/http"

	ws "github.com/coder/websocket"
)

type Conn struct {
	httpClient *http.Client
	Conn       *ws.Conn
}

var (
	ErrNotDialed = fmt.Errorf("socket: not dialed. Call Dial() before Start() or Close()")
	ErrNilRecvCh = fmt.Errorf("socket: receive channel cannot be nil")
	ErrNilSendCh = fmt.Errorf("socket: send channel cannot be nil")
)

func (c *Conn) IsDialed() bool {
	return c.Conn != nil
}

func NewConn(c *http.Client) *Conn {
	return &Conn{httpClient: c}
}

func (c *Conn) Dial(ctx context.Context, url string) error {
	cn, _, err := ws.Dial(ctx, url, &ws.DialOptions{
		HTTPClient: c.httpClient,
	})
	if err != nil {
		return err
	}
	c.Conn = cn
	return nil
}

func (c *Conn) ReadLoop(ctx context.Context, recv chan<- []byte) error {
	if !c.IsDialed() {
		return ErrNotDialed
	}
	if recv == nil {
		return ErrNilRecvCh
	}

	// close channels who sends data.
	defer close(recv)
	for {
		_, msg, err := c.Conn.Read(ctx)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case recv <- msg:
		}

	}
}

func (c *Conn) WriteLoop(ctx context.Context, send <-chan []byte) error {
	if !c.IsDialed() {
		return ErrNotDialed
	}
	if send == nil {
		return ErrNilSendCh
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-send:
			if !ok {
				return nil // The writer should close it
			}
			if err := c.Conn.Write(ctx, ws.MessageText, msg); err != nil {
				return err
			}
		}
	}
}

func (c *Conn) Close(code ws.StatusCode, reason string) error {
	if !c.IsDialed() {
		return ErrNotDialed
	}
	err := c.Conn.Close(code, reason)
	c.Conn = nil
	return err
}
