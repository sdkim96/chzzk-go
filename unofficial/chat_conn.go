package unofficial

import (
	"context"
	"fmt"
	"net/http"

	ws "github.com/coder/websocket"
)

type Conn struct {
	c    *http.Client
	conn *ws.Conn

	closeCh chan struct {
		code   ws.StatusCode
		reason string
	}
}

func (c *Conn) IsDialed() bool {
	return c.conn != nil
}

func NewConn(c *http.Client) *Conn {
	return &Conn{
		c: c,
		closeCh: make(chan struct {
			code   ws.StatusCode
			reason string
		}, 1),
	}
}

func (c *Conn) Dial(ctx context.Context, url string) error {
	cn, _, err := ws.Dial(ctx, url, &ws.DialOptions{
		HTTPClient: c.c,
	})
	if err != nil {
		return err
	}
	c.conn = cn
	return nil
}

func (c *Conn) Loop(ctx context.Context, recv chan []byte, send chan []byte, errCh chan<- error) {
	if !c.IsDialed() {
		errCh <- fmt.Errorf("socketio: not dialed. Call Dial() before Loop()")
		return
	}
	go func() {
		for {
			_, msg, err := c.conn.Read(ctx)
			if err != nil {
				errCh <- err
				return
			}
			select {
			case <-ctx.Done():
				return
			case recv <- msg:
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case data, ok := <-send:
				if !ok {
					return
				}
				err := c.conn.Write(ctx, ws.MessageText, data)
				if err != nil {
					errCh <- err
					return
				}
			}
		}
	}()
}

func (c *Conn) Close(ctx context.Context) error {
	if !c.IsDialed() {
		return nil
	}
	closeInfo := <-c.closeCh
	return c.conn.Close(closeInfo.code, closeInfo.reason)
}
