package socketio

import (
	"encoding/json"
	"fmt"
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
	return (p.EnginePacketType == EngineNone &&
		p.SocketPacketType == SocketNone &&
		p.Body == nil)
}

func (p Packet) is40() bool {
	return p.EnginePacketType == Message && p.SocketPacketType == Connect
}

func (p Packet) is42() bool {
	return p.EnginePacketType == Message && p.SocketPacketType == Event
}

func (p Packet) is0() bool {
	return p.EnginePacketType == Open
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
	case Close, Ping, Pong, Upgrade, Noop:
		// No SocketPacketType or Body for these types.
		return b, nil
	case Open:
		// Body Only. No SocketPacketType for Open packets.
		b = append(b, p.Body...)
		return b, nil
	case Message:
		if p.SocketPacketType == SocketNone {
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

	dec := newPacket(EngineNone, SocketNone, nil)

	if len(p) == 0 {
		return dec, fmt.Errorf("socketio: empty data")
	}

	ept := EnginePacketType(p[0] - '0')
	dec.EnginePacketType = ept

	switch ept {
	case Open:
		dec.Body = p[1:]
		return dec, nil
	case Ping, Pong, Noop:
		return dec, nil
	case Message:
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
