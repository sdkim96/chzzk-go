// package realtime provides a WebSocket client that has a full-duplex communication channel over a single TCP connection.
package realtime

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	ws "github.com/coder/websocket"
)

type Conn struct {
	httpClient *http.Client
	conn       *ws.Conn
}

var (
	ErrNotDialed = fmt.Errorf("realtime: not dialed. Call Dial() before Start() or Close()")
	ErrNilCh     = fmt.Errorf("realtime: send channel or receive channel cannot be nil")
)

func (c *Conn) IsDialed() bool {
	return c.conn != nil
}

func NewConn(c *http.Client) *Conn {
	return &Conn{httpClient: c}
}

func (c *Conn) Dial(ctx context.Context, url string) error {
	return c.dial(ctx, url)
}

// Start provides a full-duplex communication channel over a single TCP connection.
// It starts three goroutines:
//  1. A goroutine that reads messages from the WebSocket connection and sends them to the recv channel.
//  2. A goroutine that reads messages from the send channel and writes them to the WebSocket connection.
//  3. A goroutine that waits for the other two goroutines to finish and closes the recv and errCh channels.
//
// The caller must close the send channel. The recv channel is closed when the connection is closed or an error occurs.
func (c *Conn) Start(ctx context.Context, send <-chan []byte) (<-chan []byte, <-chan error, error) {
	if send == nil {
		return nil, nil, ErrNilCh
	}
	if !c.IsDialed() {
		return nil, nil, ErrNotDialed
	}
	return c.start(ctx, send)
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

func (c *Conn) loop(ctx context.Context, recv chan<- []byte, send <-chan []byte) error {
	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 2)
	go func() {
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
	select {
	case <-ctx2.Done():
		return ctx2.Err()
	case err := <-errCh:
		return err
	}
}

func (c *Conn) Close(ctx context.Context, code ws.StatusCode, reason string) error {
	if !c.IsDialed() {
		return ErrNotDialed
	}
	return c.close(ctx, code, reason)
}

func (c *Conn) dial(ctx context.Context, url string) error {
	cn, _, err := ws.Dial(ctx, url, &ws.DialOptions{
		HTTPClient: c.httpClient,
	})
	if err != nil {
		return err
	}
	c.conn = cn
	return nil
}

func (c *Conn) start(ctx context.Context, send <-chan []byte) (<-chan []byte, <-chan error, error) {
	var (
		wg           = sync.WaitGroup{}
		ctx2, cancel = context.WithCancel(ctx)
		recv         = make(chan []byte)
		errCh        = make(chan error, 2)
	)
	wg.Add(2)

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
	go func() {
		wg.Wait()
		close(recv)
		close(errCh)
	}()
	return recv, errCh, nil
}

func (c *Conn) close(ctx context.Context, code ws.StatusCode, reason string) error {
	err := c.conn.Close(code, reason)
	c.conn = nil
	return err
}
