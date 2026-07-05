// Package socketio provides a Go implementation of the Socket.IO v2 (Engine.IO v3) protocol.
// This implementation is a minimalistic version of the protocol, focusing on the Chzzk use case.
//
// Socket.IO consists of three layers:
//   - Application layer:           Socket.IO protocol
//   - Transport Managing layer:    Engine.IO protocol
//   - Transport layer:             WebSocket protocol (canonical protocol for Socket.IO v2)
//
// The actual transport could be any transport such as Websocket, HTTP long-polling, etc.
// However, this implementation only supports WebSocket as the transport layer,
// which is the only required transport for accessing Chzzk session API.
package socketio

import (
	"context"
	"fmt"
	"net/http"

	ws "github.com/coder/websocket"
)

const (
	SocketIOVersion = 2
	EngineIOVersion = 3
)

type EnginePacketType int
type SocketPacketType int

const (
	// This None value is a sentinel value for the zero value of EnginePacketType.
	// It is not a valid EnginePacketType and should not be used in any other context.
	EngineNone EnginePacketType = iota - 1

	Open
	Close

	// from server to client
	Ping

	// from client to server
	Pong
	Message
	Upgrade
	Noop
)

const (
	// This None value is a sentinel value for the zero value of SocketPacketType.
	// It is not a valid SocketPacketType and should not be used in any other context.
	SocketNone SocketPacketType = iota - 1

	Connect
	Disconnect

	Event
	Ack
	Error

	BinaryEvent
	BinaryAck
)

type ConnOption func(*Conn)

type Conn struct {
	conn    *ws.Conn
	c       *http.Client
	url     string
	handler map[string]func([]byte) error
}

func NewConn(url string, opts ...ConnOption) *Conn {
	conn := &Conn{
		url:     url,
		handler: make(map[string]func([]byte) error),
		c:       http.DefaultClient,
	}
	for _, o := range opts {
		o(conn)
	}
	return conn
}

func WithHTTPClient(c *http.Client) ConnOption {
	return func(conn *Conn) {
		conn.c = c
	}
}

func WithOn(pattern string, handler func([]byte) error) ConnOption {
	if pattern == "" {
		panic("socketio: pattern must not be empty")
	}
	return func(conn *Conn) {
		conn.handler[pattern] = handler
	}
}

// Dial establishes a websocket connection and handshake in Socket.IO protocol.
// It embeds a Websocket connection to Conn struct if both connection and handshake are successful.
func (c *Conn) Dial(ctx context.Context) error {

	if c.c == nil {
		c.c = http.DefaultClient
	}
	err := c.dial(ctx)
	if err != nil {
		return err
	}
	if err := c.handshake(ctx); err != nil {
		return err
	}
	return nil

}

// Close closes the underlying websocket connection.
func (c *Conn) Close(status ws.StatusCode, reason string) error {
	// TODO: Add a Socket.IO disconnect packet before closing the connection.
	return c.conn.Close(status, reason)
}

// Loop reads messages from the websocket connection and decodes them into Socket.IO packets.
// It also handles ping/pong messages and invokes the registered event handlers for the decoded packets.
func (c *Conn) Loop(ctx context.Context) error {
	for {
		_, msg, err := c.conn.Read(ctx)
		if err != nil {
			return fmt.Errorf("socketio: read failed: %w", err)
		}

		decoded, err := decode(msg)
		if err != nil {
			return fmt.Errorf("socketio: decode failed: %w", err)
		}
		if decoded.isEmpty() {
			continue
		}

		if decoded.EnginePacketType == Ping {
			c.conn.Write(ctx, ws.MessageText, []byte("3"))
			continue
		}

		pat, err := decoded.event()
		if err != nil {
			continue
		}
		if handler, ok := c.handler[pat]; ok {
			if err := handler(decoded.Body); err != nil {
				return fmt.Errorf("socketio: handler failed: %w", err)
			}
		}

	}
}

func (c *Conn) dial(ctx context.Context) error {

	conn, _, err := ws.Dial(ctx, c.url, &ws.DialOptions{
		HTTPClient: c.c,
	})
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

func (c *Conn) handshake(ctx context.Context) error {

	_, msg, err := c.conn.Read(ctx)
	if err != nil {
		return fmt.Errorf("socketio: handshake read failed: %w", err)
	}

	decoded, err := decode(msg)
	if err != nil {
		return fmt.Errorf("socketio: handshake decode failed: %w", err)
	}

	if decoded.EnginePacketType != Open {
		return fmt.Errorf("socketio: handshake expected EnginePacketType Open, got %v", decoded.EnginePacketType)
	}
	pac := newPacket(Message, Connect, nil)

	b, err := encode(pac)
	if err != nil {
		return fmt.Errorf("socketio: handshake encode failed: %w", err)
	}

	err = c.conn.Write(ctx, ws.MessageText, b)
	if err != nil {
		return fmt.Errorf("socketio: handshake write failed: %w", err)
	}

	// TODO: Implement a Socket.IO handshake response check. The server should respond with a Connect packet.
	return nil

}
