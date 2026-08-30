package unofficial

import (
	"context"
	"fmt"
	"net/url"

	"github.com/coder/websocket"
	"github.com/sdkim96/chzzk-go"
	chzzkHttp "github.com/sdkim96/chzzk-go/transport/http"
	"github.com/sdkim96/chzzk-go/transport/socket"
	"golang.org/x/sync/errgroup"
)

const (
	cmdPing       = 0     // server ping
	cmdPong       = 10000 // client pong
	cmdConnect    = 100   // connection request
	cmdConnected  = 10100 // connection established
	cmdRecentChat = 15101 // recent chat history
	cmdChat       = 93101 // live chat message
	cmdSend       = 3101  // outgoing chat message
	cmdSendAck    = 13101 // acknowledgement of an outgoing chat message
)

// ChatService serves an API for connecting to the unofficial Chat WebSocket API of Chzzk.
type ChatService struct {
	uc *Client
}

// ChatToken contains the access token and extra token returned by the Chzzk chat API.
type ChatToken struct {
	AccessToken string
	ExtraToken  string
}

type chatState struct {

	// The channels.
	recv chan<- []byte
	send <-chan []byte

	// Identities
	liveID string
	token  *ChatToken

	// The session ID returned by the server after a successful handshake.
	sid string
	// The counter for outgoing messages.
	tid int
}

func (st *chatState) Sid(sid string) {
	st.sid = sid
}
func (st *chatState) IncrementTid() {
	st.tid++
}

func (st *chatState) Op() string {
	if st.send == nil {
		return "r"
	}
	return "rw"
}

// Token retrieves the access token and extra token for a chat channel.
// The returned [ChatToken] is passed to [ChatService.ReadOnlyConnect] or [ChatService.Connect].
//   - endpoint: nng_main/v1/chats/access-token
func (s *ChatService) Token(ctx context.Context, liveID string) (*ChatToken, error) {
	return s.token(ctx, liveID)
}

// ReadOnlyConnect establishes a read-only WebSocket connection to a Chzzk chat
// channel. No authentication is required.
//
// It blocks until the connection ends, delivering each batch of incoming
// messages to recv as the raw JSON array the server sent. Ping/pong and every
// other protocol frame are handled internally and never reach recv.
//
// recv is closed when ReadOnlyConnect returns, so the caller must not close it.
// Cancel ctx to shut the connection down; the returned error is ctx.Err() in that
// case, or the underlying failure otherwise.
//   - protocol: wss://kr-ss{1-9}.chat.naver.com/chat
//   - credential: none
func (s *ChatService) ReadOnlyConnect(ctx context.Context, recv chan<- []byte, liveID string, token *ChatToken) error {
	state := &chatState{
		recv:   recv,
		send:   nil,
		liveID: liveID,
		token:  token,
		sid:    "",
	}
	return s.connect(ctx, state)
}

// Connect establishes a bidirectional WebSocket connection to a Chzzk chat
// channel. It blocks until the connection ends.
//
// recv behaves as in [ChatService.ReadOnlyConnect] and is closed on return.
//
// Each value sent on send becomes the bdy of one outgoing cmd 3101 frame; the
// envelope, including the session ID and the tid counter, is built internally.
// The server expects that bdy to hold msg, msgTypeCode, msgTime and extras, where
// extras is a JSON string:
//
//	extras, _ := json.Marshal(map[string]any{
//		"chatType":           "STREAMING",
//		"emojis":             map[string]string{},
//		"osType":             "PC",
//		"streamingChannelId": channelID,   // the channel ID, not the live ID
//		"extraToken":         token.ExtraToken,
//	})
//	body, _ := json.Marshal(map[string]any{
//		"msg":         "hello",
//		"msgTypeCode": 1,
//		"msgTime":     time.Now().UnixMilli(),
//		"extras":      string(extras),
//	})
//	send <- body
//
// extraToken is required. Without it the server discards the frame and reports
// nothing on the wire at all — no rejection, no ack, no error — so the message
// simply never appears. streamingChannelId is likewise the channel ID passed to
// [LiveService.ID], not the live ID passed to this method.
//
// A send the server does reject surfaces as an error from this method, decoded
// from its cmd 13101 acknowledgement, which ends the connection.
//
// Closing send shuts the connection down cleanly. Cancelling ctx does the same.
//   - protocol: wss://kr-ss{1-9}.chat.naver.com/chat
//   - credential: [Client.WithCookie]
func (s *ChatService) Connect(ctx context.Context, recv chan<- []byte, send <-chan []byte, liveID string, token *ChatToken) error {
	if s.uc.uid == "" {
		return fmt.Errorf("chat: Connect requires authentication. use WithCookie first, or use ReadOnlyConnect")
	}
	state := &chatState{
		recv:   recv,
		send:   send,
		liveID: liveID,
		token:  token,
		sid:    "",
	}
	return s.connect(ctx, state)
}

func (s *ChatService) token(ctx context.Context, liveID string) (*ChatToken, error) {
	u, err := url.JoinPath(NaverGameBaseURL, "nng_main", "v1", "chats", "access-token")
	if err != nil {
		return nil, err
	}
	pURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	q := pURL.Query()
	q.Set("channelId", liveID)
	q.Set("chatType", "STREAMING")
	pURL.RawQuery = q.Encode()

	type AccessTokenResp struct {
		chzzk.Response
		Content struct {
			AccessToken string `json:"accessToken"`
			ExtraToken  string `json:"extraToken"`
		} `json:"content"`
	}
	resp, err := chzzkHttp.Get[AccessTokenResp](ctx, s.uc.httpClient, pURL.String())
	if err != nil {
		return nil, err
	}
	if err := chzzk.MightError(resp.Response); err != nil {
		return nil, err
	}
	return &ChatToken{
		AccessToken: resp.Content.AccessToken,
		ExtraToken:  resp.Content.ExtraToken,
	}, nil
}

func (s *ChatService) connect(ctx context.Context, state *chatState) error {
	if state.liveID == "" {
		return fmt.Errorf("chat: liveID cannot be empty")
	}

	conn := socket.NewConn(s.uc.httpClient)
	if err := conn.Dial(ctx, fmt.Sprintf("wss://kr-ss%d.chat.naver.com/chat", chatServerID(state.liveID))); err != nil {
		return fmt.Errorf("chat: dial failed: %w", err)
	}

	if err := s.handshake(ctx, conn, state); err != nil {
		conn.Close(websocket.StatusNormalClosure, "handshake failed")
		return err
	}
	defer conn.Close(websocket.StatusNormalClosure, "Connection closed")

	// Start tossing around messages!
	g, gctx := errgroup.WithContext(ctx)

	// 1. reads.
	g.Go(func() error {
		defer close(state.recv)
		for {
			data, err := conn.Read(gctx)
			if err != nil {
				return err
			}
			cmd, bdy, err := decodeFrame(data)
			if err != nil {
				return err
			}
			switch cmd {
			case cmdPing:
				pong, err := encodePong()
				if err != nil {
					return fmt.Errorf("chat: encode pong: %w", err)
				}
				if err := conn.Write(gctx, pong); err != nil {
					return fmt.Errorf("chat: write pong failed: %w", err)
				}
			case cmdSendAck:
				// The only report of a rejected send: the server drops a bad
				// frame silently otherwise.
				if err := decodeSendAck(data); err != nil {
					return err
				}
			case cmdChat, cmdRecentChat:
				messages, err := decodeMessageList(bdy)
				if err != nil {
					return err
				}
				if messages == nil {
					continue
				}
				select {
				case <-gctx.Done():
					return gctx.Err()
				case state.recv <- messages:
				}
			}
		}
	})

	// 2. If the connection is bidirectional, start a goroutine to read from the send channel and write to the WebSocket.
	if state.Op() == "rw" && state.sid != "" {
		g.Go(func() error {
			for {
				select {
				case <-gctx.Done():
					return gctx.Err()
				case body, ok := <-state.send:
					if !ok {
						// The caller closed send: a clean shutdown, not a failure.
						return nil
					}
					frame, err := state.encodeSend(body)
					if err != nil {
						return err
					}
					if err := conn.Write(gctx, frame); err != nil {
						return fmt.Errorf("chat: write failed: %w", err)
					}
				}
			}
		})
	}
	return g.Wait()
}

func (s *ChatService) handshake(ctx context.Context, conn *socket.Conn, st *chatState) error {
	var (
		uid  any    = nil
		auth string = "READ"
	)
	if st.Op() == "rw" {
		uid = s.uc.uid
		auth = "SEND"
	}

	req, err := st.encodeConnect(uid, auth)
	if err != nil {
		return fmt.Errorf("chat: encode connect: %w", err)
	}
	if err := conn.Write(ctx, req); err != nil {
		return fmt.Errorf("chat: write connect: %w", err)
	}

	// Frames arriving before the ack are dropped: the server sends nothing the
	// caller needs until the session exists.
	for {
		data, err := conn.Read(ctx)
		if err != nil {
			return fmt.Errorf("chat: read handshake response: %w", err)
		}
		cmd, bdy, err := decodeFrame(data)
		if err != nil {
			return err
		}
		if cmd != cmdConnected {
			continue
		}
		sid, err := decodeConnected(bdy)
		if err != nil {
			return err
		}
		st.Sid(sid)
		return nil
	}
}

func chatServerID(chatChannelID string) int {
	var sum int
	for _, r := range chatChannelID {
		sum += int(r)
	}
	if sum < 0 {
		sum = -sum
	}
	return sum%9 + 1
}
