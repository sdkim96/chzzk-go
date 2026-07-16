// package realtime provides a WebSocket client that has a full-duplex communication channel over a single TCP connection.
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
	ErrNotDialed = fmt.Errorf("realtime: not dialed. Call Dial() before Start() or Close()")
	ErrNilCh     = fmt.Errorf("realtime: send channel or receive channel cannot be nil")
)

func (c *Conn) IsDialed() bool {
	return c.Conn != nil
}

func NewConn(c *http.Client) *Conn {
	return &Conn{httpClient: c}
}

func (c *Conn) Dial(ctx context.Context, url string) error {
	return c.dial(ctx, url)
}

func (c *Conn) Loop(ctx context.Context, recv chan<- []byte, send <-chan []byte) error {
	if send == nil || recv == nil {
		return ErrNilCh
	}
	if !c.IsDialed() {
		return ErrNotDialed
	}
	return c.loop(ctx, recv, send)
}

func (c *Conn) Close(ctx context.Context, code ws.StatusCode, reason string) error {
	if !c.IsDialed() {
		return ErrNotDialed
	}
	return c.close(ctx, code, reason)
}

func (c *Conn) loop(ctx context.Context, recv chan<- []byte, send <-chan []byte) error {
	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 2)
	go func() {
		for {
			_, msg, err := c.Conn.Read(ctx2)
			if err != nil {
				errCh <- err
				return
			}
			select {
			case <-ctx2.Done():
				return
			case recv <- msg:
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx2.Done():
				return
			case data, ok := <-send:
				if !ok {
					return
				}
				err := c.Conn.Write(ctx2, ws.MessageText, data)
				if err != nil {
					errCh <- err
					return
				}
			}
		}
	}()
	select {
	case <-ctx2.Done():
		return ctx2.Err()
	case err := <-errCh:
		return err
	}
}

func (c *Conn) dial(ctx context.Context, url string) error {
	cn, _, err := ws.Dial(ctx, url, &ws.DialOptions{
		HTTPClient: c.httpClient,
	})
	if err != nil {
		return err
	}
	c.Conn = cn
	return nil
}

func (c *Conn) close(ctx context.Context, code ws.StatusCode, reason string) error {
	err := c.Conn.Close(code, reason)
	c.Conn = nil
	return err
}
