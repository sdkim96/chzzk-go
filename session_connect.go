package chzzk

import (
	"context"
	"fmt"
	"strings"

	ws "github.com/coder/websocket"
	"github.com/sdkim96/chzzk-go/transport/socket"
)

// This file provides a Go implementation of the Socket.IO v2 (Engine.IO v3) protocol.
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
const (
	SocketIOVersion = 2
	EngineIOVersion = 3
)

// Handler is a function type that handles incoming events from the Socket.IO connection.
// By the specification, a packet number 42 is a event packet, which is the only packet type that carries application data.
//
// Following is the example of packet 42.
//
//	42["SYSTEM",{"type":"connected","data":{"sessionKey":"xyz789"}}]
//
// Since this packet is for application level, the optional handlers can be
// registered to handle specific events.
// For example, You can register a handler for the "CHAT" event like this:
//
//	clientID := os.Getenv("CHZZK_CLIENT_ID")
//	clientSecret := os.Getenv("CHZZK_CLIENT_SECRET")
//
//	if clientID == "" || clientSecret == "" {
//		t.Skip("CHZZK_CLIENT_ID or CHZZK_CLIENT_SECRET not set")
//	}
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	c := New(nil).WithClientAuth(clientID, clientSecret)
//	sessionURL, err := c.Session.AuthClient(ctx)
//
//	if err != nil {
//		t.Fatalf("AuthClient() error = %v", err)
//	}
//
//	t.Logf("Session URL: %s", sessionURL)
//
//	wsURL, err := AsWebSocketURL(sessionURL)
//	t.Logf("WebSocket URL: %s", wsURL)
//
//	received := make(chan []byte, 1)
//	go func() {
//		c.Session.Connect(ctx, wsURL, map[string]Handler{
//			"SYSTEM": func(data []byte) error {
//				received <- data
//				return nil
//			},
//		})
//	}()
//
//	select {
//	case data := <-received:
//		t.Logf("SYSTEM event received: %s", data)
//	case <-ctx.Done():
//		t.Fatal("timeout waiting for SYSTEM event")
//	}
type Handler func([]byte) error

var ErrEmptyHandlerKey = fmt.Errorf("handler key cannot be empty. Register a handler with a non-empty key to handle incoming events, example: CHAT, SYSTEM")

func (s *SessionService) Connect(ctx context.Context, u string, h map[string]Handler) error {
	nh := make(map[string]Handler)
	for k, v := range h {
		if v == nil {
			continue
		}
		if k == "" {
			return ErrEmptyHandlerKey
		}
		nk := strings.ToUpper(k)
		nh[nk] = v
	}
	return s.connect(ctx, u, nh)
}

func (s *SessionService) connect(ctx context.Context, u string, h map[string]Handler) error {
	conn := socket.NewConn(s.c.httpClient)
	defer conn.Close(ctx, ws.StatusNormalClosure, "session connect done")
	if err := conn.Dial(ctx, u); err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	if err := handshake(ctx, conn); err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}
	recv := make(chan []byte)
	send := make(chan []byte)
	errCh := make(chan error, 1)
	go func() {
		errCh <- conn.Loop(ctx, recv, send)
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			return fmt.Errorf("loop error: %w", err)
		case msg, ok := <-recv:
			if !ok {
				return fmt.Errorf("receive channel closed")
			}
			p, err := decode(msg)
			if err != nil {
				return fmt.Errorf("failed to decode packet: %w", err)
			}
			if p.isEmpty() {
				continue
			}
			if p.EnginePacketType == socketIOPing {
				b, err := encode(newPacket(socketIOPong, socketIOSocketNone, nil))
				if err != nil {
					return fmt.Errorf("socketio: encode failed: %w", err)
				}
				send <- b
				continue
			}
			ev, data, err := p.event()
			if err != nil {
				continue
			}
			if handler, ok := h[ev]; ok {
				if err := handler(data); err != nil {
					return fmt.Errorf("socketio: handler failed: %w", err)
				}
			}
		}
	}
}

func handshake(ctx context.Context, c *socket.Conn) error {
	// Phase 1: Read Open (0) packet from Server
	_, msg, err := c.Conn.Read(ctx)
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
	b0, err := encode(newPacket(socketIOMessage, socketIOConnect, nil))
	if err != nil {
		return fmt.Errorf("socketio: handshake encode failed: %w", err)
	}

	err = c.Conn.Write(ctx, ws.MessageText, b0)
	if err != nil {
		return fmt.Errorf("socketio: handshake write failed: %w", err)
	}

	// Phase 3: Read Message Connect (40) packet from Server
	_, msg, err = c.Conn.Read(ctx)
	if err != nil {
		return fmt.Errorf("socketio: handshake read failed: %w", err)
	}

	p1, err := decode(msg)
	if err != nil {
		return fmt.Errorf("socketio: handshake decode failed: %w", err)
	}

	if !p1.is40() {
		return fmt.Errorf("socketio: handshake expected EnginePacketType socketIOMessage and SocketPacketType socketIOConnect, got %v and %v", p1.EnginePacketType, p1.SocketPacketType)
	}
	return nil
}
