package socketio

import "testing"

// Test_decode tests the decode function with various Socket.IO v2 wire format inputs.
func Test_decode(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    Packet
		wantErr bool
	}{
		// Engine.IO: Open
		{
			name:  "engine open",
			input: []byte(`0{"sid":"abc123","pingInterval":25000,"pingTimeout":5000}`),
			want: Packet{
				EnginePacketType: Open,
				SocketPacketType: SocketNone,
				Body:             []byte(`{"sid":"abc123","pingInterval":25000,"pingTimeout":5000}`),
			},
		},
		// Engine.IO: Ping — no socket layer, no body
		{
			name:  "engine ping",
			input: []byte(`2`),
			want: Packet{
				EnginePacketType: Ping,
				SocketPacketType: SocketNone,
				Body:             nil,
			},
		},
		// Engine.IO: Pong — no socket layer, no body
		{
			name:  "engine pong",
			input: []byte(`3`),
			want: Packet{
				EnginePacketType: Pong,
				SocketPacketType: SocketNone,
				Body:             nil,
			},
		},
		// Engine.IO: Noop — no socket layer, no body
		{
			name:  "engine noop",
			input: []byte(`6`),
			want: Packet{
				EnginePacketType: Noop,
				SocketPacketType: SocketNone,
				Body:             nil,
			},
		},
		// Socket.IO: Connect — no body
		{
			name:  "socket connect",
			input: []byte(`40`),
			want: Packet{
				EnginePacketType: Message,
				SocketPacketType: Connect,
				Body:             []byte{},
			},
		},
		// Socket.IO: Event - SYSTEM connected
		// sessionKey is delivered via WebSocket, not from the REST API response
		{
			name:  "socket event system connected",
			input: []byte(`42["SYSTEM",{"type":"connected","data":{"sessionKey":"xyz789"}}]`),
			want: Packet{
				EnginePacketType: Message,
				SocketPacketType: Event,
				Body:             []byte(`["SYSTEM",{"type":"connected","data":{"sessionKey":"xyz789"}}]`),
			},
		},
		// Socket.IO: Event - CHAT
		{
			name:  "socket event chat",
			input: []byte(`42["CHAT",{"channelId":"ch123","content":"hello","messageTime":1234567890}]`),
			want: Packet{
				EnginePacketType: Message,
				SocketPacketType: Event,
				Body:             []byte(`["CHAT",{"channelId":"ch123","content":"hello","messageTime":1234567890}]`),
			},
		},
		// Socket.IO: Event - DONATION
		{
			name:  "socket event donation",
			input: []byte(`42["DONATION",{"donationType":"CHAT","payAmount":"1000","donationText":"good luck"}]`),
			want: Packet{
				EnginePacketType: Message,
				SocketPacketType: Event,
				Body:             []byte(`["DONATION",{"donationType":"CHAT","payAmount":"1000","donationText":"good luck"}]`),
			},
		},
		// Socket.IO: Event - SUBSCRIPTION
		{
			name:  "socket event subscription",
			input: []byte(`42["SUBSCRIPTION",{"tierNo":1,"month":3,"subscriberNickname":"subscriber"}]`),
			want: Packet{
				EnginePacketType: Message,
				SocketPacketType: Event,
				Body:             []byte(`["SUBSCRIPTION",{"tierNo":1,"month":3,"subscriberNickname":"subscriber"}]`),
			},
		},

		// Error cases
		{
			name:    "empty data",
			input:   []byte{},
			wantErr: true,
		},
		{
			name:    "unknown engine packet type",
			input:   []byte(`9something`),
			wantErr: true,
		},
		{
			name:    "message without socket type",
			input:   []byte(`4`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.EnginePacketType != tt.want.EnginePacketType {
				t.Errorf("EnginePacketType = %v, want %v", got.EnginePacketType, tt.want.EnginePacketType)
			}
			if got.SocketPacketType != tt.want.SocketPacketType {
				t.Errorf("SocketPacketType = %v, want %v", got.SocketPacketType, tt.want.SocketPacketType)
			}
			if string(got.Body) != string(tt.want.Body) {
				t.Errorf("Body = %q, want %q", got.Body, tt.want.Body)
			}
		})
	}
}

// Test_encode tests the encode function produces correct Socket.IO v2 wire format.
func Test_encode(t *testing.T) {
	tests := []struct {
		name    string
		input   Packet
		want    string
		wantErr bool
	}{
		{
			name:  "pong",
			input: newPacket(Pong, SocketNone, nil),
			want:  "3",
		},
		{
			name:  "close",
			input: newPacket(Close, SocketNone, nil),
			want:  "1",
		},
		{
			name:  "socket connect",
			input: newPacket(Message, Connect, nil),
			want:  "40",
		},
		{
			name:  "socket event with body",
			input: newPacket(Message, Event, []byte(`["CHAT",{"content":"hello"}]`)),
			want:  `42["CHAT",{"content":"hello"}]`,
		},
		{
			// message without SocketPacketType must fail
			name:    "message without socket type",
			input:   newPacket(Message, SocketNone, nil),
			wantErr: true,
		},
		{
			// EngineNone is not a valid packet type
			name:    "unknown engine packet type",
			input:   newPacket(EngineNone, SocketNone, nil),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if string(got) != tt.want {
				t.Errorf("encode() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Test_encode_decode_symmetry tests that encode → decode round-trip produces the same packet.
func Test_encode_decode_symmetry(t *testing.T) {
	tests := []struct {
		name  string
		input Packet
		wire  string
	}{
		{
			name:  "pong",
			input: newPacket(Pong, SocketNone, nil),
			wire:  "3",
		},
		{
			name:  "socket connect",
			input: newPacket(Message, Connect, nil),
			wire:  "40",
		},
		{
			name:  "socket event",
			input: newPacket(Message, Event, []byte(`["CHAT",{"content":"hello"}]`)),
			wire:  `42["CHAT",{"content":"hello"}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := encode(tt.input)
			if err != nil {
				t.Fatalf("encode() error = %v", err)
			}
			if string(encoded) != tt.wire {
				t.Errorf("encode() = %q, want %q", encoded, tt.wire)
			}
			decoded, err := decode(encoded)
			if err != nil {
				t.Fatalf("decode() error = %v", err)
			}
			if decoded.EnginePacketType != tt.input.EnginePacketType {
				t.Errorf("EnginePacketType = %v, want %v", decoded.EnginePacketType, tt.input.EnginePacketType)
			}
			if decoded.SocketPacketType != tt.input.SocketPacketType {
				t.Errorf("SocketPacketType = %v, want %v", decoded.SocketPacketType, tt.input.SocketPacketType)
			}
			if string(decoded.Body) != string(tt.input.Body) {
				t.Errorf("Body = %q, want %q", decoded.Body, tt.input.Body)
			}
		})
	}
}

// Test_isEmpty tests the isEmpty method.
func Test_isEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input Packet
		want  bool
	}{
		{
			name:  "sentinel values — empty",
			input: newPacket(EngineNone, SocketNone, nil),
			want:  true,
		},
		{
			name:  "has engine type — not empty",
			input: newPacket(Open, SocketNone, nil),
			want:  false,
		},
		{
			name:  "has body — not empty",
			input: newPacket(EngineNone, SocketNone, []byte("data")),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.isEmpty(); got != tt.want {
				t.Errorf("isEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_event tests event() extracts event name and data from a 42 packet.
func Test_event(t *testing.T) {
	tests := []struct {
		name     string
		input    Packet
		wantName string
		wantData string
		wantErr  bool
	}{
		{
			// non-event packet should return error
			name:    "non-event packet",
			input:   newPacket(Open, SocketNone, nil),
			wantErr: true,
		},
		{
			// SYSTEM connected — extracts name and data
			name:     "system connected",
			input:    newPacket(Message, Event, []byte(`["SYSTEM",{"type":"connected","data":{"sessionKey":"xyz"}}]`)),
			wantName: "SYSTEM",
			wantData: `{"type":"connected","data":{"sessionKey":"xyz"}}`,
		},
		{
			// CHAT — extracts name and data
			name:     "chat event",
			input:    newPacket(Message, Event, []byte(`["CHAT",{"channelId":"ch1","content":"hello"}]`)),
			wantName: "CHAT",
			wantData: `{"channelId":"ch1","content":"hello"}`,
		},
		{
			// DONATION
			name:     "donation event",
			input:    newPacket(Message, Event, []byte(`["DONATION",{"payAmount":"1000"}]`)),
			wantName: "DONATION",
			wantData: `{"payAmount":"1000"}`,
		},
		{
			// SUBSCRIPTION
			name:     "subscription event",
			input:    newPacket(Message, Event, []byte(`["SUBSCRIPTION",{"tierNo":1}]`)),
			wantName: "SUBSCRIPTION",
			wantData: `{"tierNo":1}`,
		},
		{
			// revoked — must not be silently dropped
			// failing to handle revoked causes the client to stop receiving events silently
			name:     "system revoked",
			input:    newPacket(Message, Event, []byte(`["SYSTEM",{"type":"revoked","data":{"eventType":"CHAT","channelId":"ch1"}}]`)),
			wantName: "SYSTEM",
			wantData: `{"type":"revoked","data":{"eventType":"CHAT","channelId":"ch1"}}`,
		},
		{
			// empty payload array must fail
			name:    "empty payload",
			input:   newPacket(Message, Event, []byte(`[]`)),
			wantErr: true,
		},
		{
			// malformed JSON must fail
			name:    "malformed json",
			input:   newPacket(Message, Event, []byte(`not json`)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, data, err := tt.input.event()
			if (err != nil) != tt.wantErr {
				t.Errorf("event() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if name != tt.wantName {
				t.Errorf("event name = %q, want %q", name, tt.wantName)
			}
			if string(data) != tt.wantData {
				t.Errorf("event data = %q, want %q", data, tt.wantData)
			}
		})
	}
}
