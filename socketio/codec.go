package socketio

import "fmt"

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

func (p Packet) event() (string, error) {
	if p.SocketPacketType == Event {
		//TODO: The event packet looks like this: 42["event_name",{"key":"value"}]
		// We need to parse the event name and the data from the body.
		// For now, we just return the body as a string.

		return string(p.Body), nil
	}
	return "", fmt.Errorf("socketio: not an event packet")
}

func encode(dp Packet) ([]byte, error) {

	// TODO: Implement the encoding logic for Socket.IO packets.
	var p []byte
	return p, nil

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
