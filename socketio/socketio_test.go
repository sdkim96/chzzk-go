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
		// Engine.IO: Open — SocketPacketType should be SocketNone
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
		// Socket.IO: Event - SYSTEM subscribed
		{
			name:  "socket event system subscribed",
			input: []byte(`42["SYSTEM",{"type":"subscribed","data":{"eventType":"CHAT","channelId":"ch123"}}]`),
			want: Packet{
				EnginePacketType: Message,
				SocketPacketType: Event,
				Body:             []byte(`["SYSTEM",{"type":"subscribed","data":{"eventType":"CHAT","channelId":"ch123"}}]`),
			},
		},
		// Socket.IO: Event - SYSTEM unsubscribed
		{
			name:  "socket event system unsubscribed",
			input: []byte(`42["SYSTEM",{"type":"unsubscribed","data":{"eventType":"CHAT","channelId":"ch123"}}]`),
			want: Packet{
				EnginePacketType: Message,
				SocketPacketType: Event,
				Body:             []byte(`["SYSTEM",{"type":"unsubscribed","data":{"eventType":"CHAT","channelId":"ch123"}}]`),
			},
		},
		// Socket.IO: Event - SYSTEM revoked
		// revoked is sent when the user revokes consent or scope changes.
		// failing to handle this will cause the client to silently stop receiving events.
		{
			name:  "socket event system revoked",
			input: []byte(`42["SYSTEM",{"type":"revoked","data":{"eventType":"CHAT","channelId":"ch123"}}]`),
			want: Packet{
				EnginePacketType: Message,
				SocketPacketType: Event,
				Body:             []byte(`["SYSTEM",{"type":"revoked","data":{"eventType":"CHAT","channelId":"ch123"}}]`),
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
		// Message packet without socket packet type
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

// Test_isEmpty tests the isEmpty method on Packet.
func Test_isEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input Packet
		want  bool
	}{
		{
			name:  "zero value packet is empty",
			input: newPacket(EngineNone, SocketNone, nil),
			want:  true,
		},
		{
			name:  "packet with engine type is not empty",
			input: newPacket(Open, SocketNone, nil),
			want:  false,
		},
		{
			name:  "packet with body is not empty",
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

// Test_event tests that event() returns error for non-event packets.
func Test_event(t *testing.T) {
	tests := []struct {
		name    string
		input   Packet
		wantErr bool
	}{
		{
			// event() on a non-event packet should return error
			name:    "non-event packet returns error",
			input:   newPacket(Open, SocketNone, nil),
			wantErr: true,
		},
		{
			// event() on an event packet should not return error
			// actual event name parsing is not yet implemented (TODO in event())
			name:    "event packet does not return error",
			input:   newPacket(Message, Event, []byte(`["CHAT",{"content":"hello"}]`)),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.input.event()
			if (err != nil) != tt.wantErr {
				t.Errorf("event() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test_WithOn_emptyPattern tests that WithOn panics on empty pattern.
func Test_WithOn_emptyPattern(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("WithOn with empty pattern should panic")
		}
	}()
	WithOn("", func(p []byte) error { return nil })
}

// Test_NewConn tests that NewConn initializes Conn with correct defaults.
func Test_NewConn(t *testing.T) {
	url := "wss://example.com/socket.io/?EIO=3&transport=websocket"
	conn := NewConn(url)

	if conn == nil {
		t.Fatal("NewConn returned nil")
	}
	if conn.url != url {
		t.Errorf("url = %v, want %v", conn.url, url)
	}
	if conn.handler == nil {
		t.Error("handler map should be initialized")
	}
	if conn.c == nil {
		t.Error("http.Client should be initialized")
	}
}
