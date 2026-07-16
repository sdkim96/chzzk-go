package chzzk

import (
	"encoding/json"
	"fmt"
)

type EnginePacketType int
type SocketPacketType int

const (
	// This None value is a sentinel value for the zero value of EnginePacketType.
	// It is not a valid EnginePacketType and should not be used in any other context.
	socketIOEngineNone EnginePacketType = iota - 1

	socketIOOpen
	socketIOClose

	// from server to client
	socketIOPing

	// from client to server
	socketIOPong
	socketIOMessage
	socketIOUpgrade
	socketIONoop
)

const (
	// This None value is a sentinel value for the zero value of SocketPacketType.
	// It is not a valid SocketPacketType and should not be used in any other context.
	socketIOSocketNone SocketPacketType = iota - 1

	socketIOConnect
	socketIODisconnect

	socketIOEvent
	socketIOAck
	socketIOError

	socketIOBinaryEvent
	socketIOBinaryAck
)

type Packet struct {
	EnginePacketType EnginePacketType
	SocketPacketType SocketPacketType
	Body             []byte
}

func newPacket(ept EnginePacketType, spt SocketPacketType, msg []byte) Packet {
	return Packet{
		EnginePacketType: ept,
		SocketPacketType: spt,
		Body:             msg,
	}
}

func (p Packet) isEmpty() bool {
	return (p.EnginePacketType == socketIOEngineNone &&
		p.SocketPacketType == socketIOSocketNone &&
		p.Body == nil)
}

func (p Packet) is40() bool {
	return p.EnginePacketType == socketIOMessage && p.SocketPacketType == socketIOConnect
}

func (p Packet) is42() bool {
	return p.EnginePacketType == socketIOMessage && p.SocketPacketType == socketIOEvent
}

func (p Packet) is0() bool {
	return p.EnginePacketType == socketIOOpen
}

func (p Packet) event() (string, []byte, error) {
	if p.is42() {
		var data []json.RawMessage
		if err := json.Unmarshal(p.Body, &data); err != nil {
			return "", nil, fmt.Errorf("socketio: failed to unmarshal event packet body: %w", err)
		}
		if len(data) < 1 {
			return "", nil, fmt.Errorf("socketio: event packet body is empty")
		}
		var eventName string
		if err := json.Unmarshal(data[0], &eventName); err != nil {
			return "", nil, fmt.Errorf("socketio: failed to unmarshal event name: %w", err)
		}
		return eventName, data[1], nil
	}
	return "", nil, fmt.Errorf("socketio: not an event packet. Currently: EnginePacketType=%v, SocketPacketType=%v", p.EnginePacketType, p.SocketPacketType)
}

func encode(p Packet) ([]byte, error) {
	var b []byte
	b = append(b, byte(p.EnginePacketType)+'0')
	switch p.EnginePacketType {
	case socketIOClose, socketIOPing, socketIOPong, socketIOUpgrade, socketIONoop:
		// No SocketPacketType or Body for these types.
		return b, nil
	case socketIOOpen:
		// Body Only. No SocketPacketType for Open packets.
		b = append(b, p.Body...)
		return b, nil
	case socketIOMessage:
		if p.SocketPacketType == socketIOSocketNone {
			return nil, fmt.Errorf("socketio: message packet must have a SocketPacketType")
		}
		b = append(b, byte(p.SocketPacketType)+'0')
		b = append(b, p.Body...)
		return b, nil
	default:
		return nil, fmt.Errorf("socketio: unknown engine packet type: %d", p.EnginePacketType)
	}
}

// Decode decodes a Socket.IO packet from the given byte slice.
// It returns a Packet struct that contains the Engine.IO packet type, Socket.IO packet type, and message body.
func decode(p []byte) (Packet, error) {

	dec := newPacket(socketIOEngineNone, socketIOSocketNone, nil)

	if len(p) == 0 {
		return dec, fmt.Errorf("socketio: empty data")
	}

	ept := EnginePacketType(p[0] - '0')
	dec.EnginePacketType = ept

	switch ept {
	case socketIOOpen:
		dec.Body = p[1:]
		return dec, nil
	case socketIOPing, socketIOPong, socketIONoop:
		return dec, nil
	case socketIOMessage:
		if len(p) < 2 {
			return dec, fmt.Errorf("socketio: message packet too short")
		}
		spt := SocketPacketType(p[1] - '0')
		dec.SocketPacketType = spt

		dec.Body = p[2:]
		return dec, nil

	default:
		return dec, fmt.Errorf("socketio: unknown engine packet type: %d", ept)
	}

}
