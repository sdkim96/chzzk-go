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

// ConnOption is a function type that modifies the Conn struct.
type ConnOption func(*Conn)

// Conn represents a Socket.IO connection over WebSocket.
// By the specification, a packet number 42 is a event packet,
// which is the only packet type that carries application data.
//
// Following is the example of packet 42.
//
//	42["SYSTEM",{"type":"connected","data":{"sessionKey":"xyz789"}}]
//
// Since this packet is for application level, the optional handlers can be
// registered to handle specific events.
// For example, You can register a handler for the "CHAT" event like this:
//
//	conn := socketio.New(url)
//	conn.WithHandler("SYSTEM", func(p []byte) error {
//	    fmt.Println("Received SYSTEM event:", string(p))
//
//	.     var data struct {
//		        Type string `json:"type"`
//		        Data struct {
//		            SessionKey string `json:"sessionKey"`
//		        } `json:"data"`
//		    }
//		    if err := json.Unmarshal(p, &data); err != nil {
//		        return fmt.Errorf("failed to unmarshal SYSTEM event: %w", err)
//		    }
//		    fmt.Println("Session Key:", data.Data.SessionKey)
//		    return nil
//		})
//
// >> Note: The event name is case-sensitive. "CHAT" and "chat" are different events.
type Conn struct {

	// The base connection to the Websocket server.
	conn *ws.Conn
	c    *http.Client
	url  string

	// The event handler map, where the key is the event name and the value is the handler function.
	handler map[string]func([]byte) error
}

// New creates a new Socket.IO connection with the given URL and optional connection options.
// This does not establish the connection yet.
// You need to call Dial() to establish the connection and handshake.
func New(url string, opts ...ConnOption) *Conn {
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

func WithHandler(pattern string, handler func([]byte) error) ConnOption {
	if pattern == "" {
		panic("socketio: pattern must not be empty")
	}
	return func(conn *Conn) {
		conn.handler[pattern] = handler
	}
}

// Dial establishes a websocket connection and handshake in Socket.IO protocol.
// It embeds a Websocket connection to Conn object if both connection and handshake are successful.
func (c *Conn) Dial(ctx context.Context) error {
	if err := c.dial(ctx); err != nil {
		return err
	}
	if err := c.handshake(ctx); err != nil {
		return err
	}
	return nil

}

// Close closes the underlying websocket connection.
// It sends a Close packet to the server before closing the connection.
func (c *Conn) Close(ctx context.Context, status ws.StatusCode, reason string) error {
	if !c.isDialed() {
		return fmt.Errorf("socketio: not dialed. Call Dial() before Close()")
	}
	defer c.conn.Close(status, reason)
	return c.close(ctx)
}

// Loop reads messages from the websocket connection and decodes them into Socket.IO packets.
// It also handles ping/pong messages and invokes the registered event handlers for the decoded packets.
func (c *Conn) Loop(ctx context.Context) error {
	if !c.isDialed() {
		return fmt.Errorf("socketio: not dialed. Call Dial() before Loop()")
	}
	return c.loop(ctx)
}

func (c *Conn) isDialed() bool {
	return c.conn != nil
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
	// Phase 1: Read Open (0) packet from Server
	_, msg, err := c.conn.Read(ctx)
	if err != nil {
		return fmt.Errorf("socketio: handshake read failed: %w", err)
	}

	p0, err := decode(msg)
	if err != nil {
		return fmt.Errorf("socketio: handshake decode failed: %w", err)
	}

	if !p0.is0() {
		return fmt.Errorf("socketio: handshake expected EnginePacketType Open, got %v", p0.EnginePacketType)
	}

	// Phase 2: Send Message Connect (40) packet to Server
	b0, err := encode(newPacket(Message, Connect, nil))
	if err != nil {
		return fmt.Errorf("socketio: handshake encode failed: %w", err)
	}

	err = c.conn.Write(ctx, ws.MessageText, b0)
	if err != nil {
		return fmt.Errorf("socketio: handshake write failed: %w", err)
	}

	// Phase 3: Read Message Connect (40) packet from Server
	_, msg, err = c.conn.Read(ctx)
	if err != nil {
		return fmt.Errorf("socketio: handshake read failed: %w", err)
	}

	p1, err := decode(msg)
	if err != nil {
		return fmt.Errorf("socketio: handshake decode failed: %w", err)
	}

	if !p1.is40() {
		return fmt.Errorf("socketio: handshake expected EnginePacketType Message and SocketPacketType Connect, got %v and %v", p1.EnginePacketType, p1.SocketPacketType)
	}
	return nil
}

func (c *Conn) close(ctx context.Context) error {
	b, err := encode(newPacket(Close, SocketNone, nil))
	if err != nil {
		return fmt.Errorf("socketio: encode failed: %w", err)
	}
	return c.conn.Write(ctx, ws.MessageText, b)
}

func (c *Conn) loop(ctx context.Context) error {
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
			b, err := encode(newPacket(Pong, SocketNone, nil))
			if err != nil {
				return fmt.Errorf("socketio: encode failed: %w", err)
			}
			err = c.conn.Write(ctx, ws.MessageText, b)
			if err != nil {
				return fmt.Errorf("socketio: pong failed: %w", err)
			}
			continue
		}

		ev, data, err := decoded.event()
		if err != nil {
			continue
		}
		if handler, ok := c.handler[ev]; ok {
			if err := handler(data); err != nil {
				return fmt.Errorf("socketio: handler failed: %w", err)
			}
		}
	}
}
