package internal

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	ws "github.com/coder/websocket"
)

type Conn struct {
	c    *http.Client
	conn *ws.Conn
}

func (c *Conn) IsDialed() bool {
	return c.conn != nil
}

func NewConn(c *http.Client) *Conn {
	return &Conn{c: c}
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

func (c *Conn) Start(ctx context.Context, send <-chan []byte) (<-chan []byte, <-chan error, error) {
	var (
		wg           = sync.WaitGroup{}
		ctx2, cancel = context.WithCancel(ctx)
		recv         = make(chan []byte)
		errCh        = make(chan error, 2)
	)
	if !c.IsDialed() {
		cancel()
		return nil, nil, fmt.Errorf("conn: not dialed. Call Dial() before Start()")
	}
	wg.Add(2)
	go func() {
		wg.Wait()
		close(recv)
		close(errCh)
	}()
	go func() {
		defer cancel()
		defer wg.Done()
		for {
			_, msg, err := c.conn.Read(ctx2)
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
		defer cancel()
		defer wg.Done()
		for {
			select {
			case <-ctx2.Done():
				return
			case data, ok := <-send:
				if !ok {
					return
				}
				err := c.conn.Write(ctx2, ws.MessageText, data)
				if err != nil {
					errCh <- err
					return
				}
			}
		}
	}()
	return recv, errCh, nil
}

func (c *Conn) Close(ctx context.Context, code ws.StatusCode, reason string) error {
	if !c.IsDialed() {
		return nil
	}
	return c.conn.Close(code, reason)
}
