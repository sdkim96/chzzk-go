package unofficial

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/coder/websocket"
	"github.com/sdkim96/chzzk-go"
	chzzkHttp "github.com/sdkim96/chzzk-go/transport/http"
	"github.com/sdkim96/chzzk-go/transport/socket"
)

const (
	cmdPing       = 0     // server ping
	cmdPong       = 10000 // client pong
	cmdConnect    = 100   // connection request
	cmdConnected  = 10100 // connection established
	cmdRecentChat = 15101 // recent chat history
	cmdChat       = 93101 // live chat message
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

	// The operation mode: "r" for read-only, "rw" for read-write.
	op string

	// The channels.
	//
	// TODO: directions are inverted. The library is the sender on recv and the
	// receiver on send, so these must become `recv chan<- []byte` and
	// `send <-chan []byte`, along with the ReadOnlyConnect/Connect signatures.
	// This typechecks today only because make(chan []byte) converts to either
	// direction at the call site.
	recv <-chan []byte
	send chan<- []byte

	// Identities
	liveID string
	token  *ChatToken

	// The session ID returned by the server after a successful handshake.
	sid string
}

// Token retrieves the access token and extra token for a chat channel.
// The returned [ChatToken] is passed to [ChatService.ReadOnlyConnect] or [ChatService.Connect].
//   - endpoint: nng_main/v1/chats/access-token
func (s *ChatService) Token(ctx context.Context, liveID string) (*ChatToken, error) {
	return s.token(ctx, liveID)
}

// ReadOnlyConnect establishes a read-only WebSocket connection to a Chzzk chat channel.
// No authentication is required. The recv channel receives incoming chat messages as raw bytes.
// The connection is closed when ctx is cancelled or the recv channel is closed.
//   - protocol: wss://kr-ss{1-9}.chat.naver.com/chat
//   - credential: none
func (s *ChatService) ReadOnlyConnect(ctx context.Context, recv <-chan []byte, liveID string, token *ChatToken) error {
	state := &chatState{
		op:     "r",
		recv:   recv,
		send:   nil,
		liveID: liveID,
		token:  token,
		sid:    "",
	}
	return s.connect(ctx, state)
}

// Connect establishes a bidirectional WebSocket connection to a Chzzk chat channel.
// Returns recv for incoming messages, send for outgoing messages, and the session ID.
// The session ID (sid) must be included in outgoing chat messages (cmd 3101).
// The connection is closed when ctx is cancelled.
//   - protocol: wss://kr-ss{1-9}.chat.naver.com/chat
//   - credential: [UnofficialChzzk.WithCookie]
func (s *ChatService) Connect(ctx context.Context, recv <-chan []byte, send chan<- []byte, liveID string, token *ChatToken) error {
	if s.uc.uid == "" {
		return fmt.Errorf("chat: Connect requires authentication. use WithCookie first, or use ReadOnlyConnect")
	}
	state := &chatState{
		op:     "rw",
		recv:   recv,
		send:   send,
		liveID: liveID,
		token:  token,
		sid:    "",
	}
	return s.connect(ctx, state)
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

	// TODO: neither pump is wired up, so connect currently dials and returns
	// without doing any I/O. Start conn.ReadLoop and conn.WriteLoop under an
	// errgroup the way SessionService.connect does, with the read path folded
	// in from s.loop below, and block on g.Wait() instead of returning nil.

	// Write path
	if state.op == "rw" {
		go func() {

		}()
	}

	go func() {

	}()

	return nil
}

func (s *ChatService) handshake(ctx context.Context, conn *socket.Conn, st *chatState) error {
	type wsRequest struct {
		Bdy   any    `json:"bdy"`
		Cid   string `json:"cid"`
		Cmd   int    `json:"cmd"`
		Svcid string `json:"svcid"`
		Tid   int    `json:"tid"`
		Ver   string `json:"ver"`
	}
	type wsResponse struct {
		Cmd int `json:"cmd"`
		Bdy struct {
			Sid string `json:"sid"`
		} `json:"bdy"`
	}
	type connectBody struct {
		Uid     any    `json:"uid"`
		DevType int    `json:"devType"`
		AccTkn  string `json:"accTkn"`
		Auth    string `json:"auth"`
	}

	var (
		uid  any    = nil
		auth string = "READ"
	)
	if st.op == "rw" {
		uid = s.uc.uid
		auth = "SEND"
	}

	connectReq := wsRequest{
		Cmd:   cmdConnect,
		Tid:   1,
		Cid:   st.liveID,
		Svcid: "game",
		Ver:   "2",
		Bdy: connectBody{
			Uid:     uid,
			DevType: 2001,
			AccTkn:  st.token.AccessToken,
			Auth:    auth,
		},
	}

	// TODO: the marshaled frame is discarded, so cmd 100 never reaches the
	// server and the loop below blocks until ctx expires. Write it to conn
	// (or hand it to the send path) before waiting for cmd 10100.
	_, err := json.Marshal(connectReq)
	if err != nil {
		return fmt.Errorf("chat: marshal connect: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data, ok := <-st.recv:
			if !ok {
				return fmt.Errorf("chat: connection closed before handshake")
			}
			var resp wsResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("chat: unmarshal handshake: %w", err)
			}
			if resp.Cmd == cmdConnected {
				st.sid = resp.Bdy.Sid
				return nil
			}
		}
	}
}

// TODO: dead code. Nothing calls this, and loopErrCh no longer has a producer
// now that socket.Conn.Loop is split into ReadLoop/WriteLoop. Fold the cmd
// dispatch (ping/pong, 93101/15101 messageList unwrapping) into the read path
// in connect and delete the rest.
func (s *ChatService) loop(ctx context.Context, conn *socket.Conn, recv chan<- []byte, send <-chan []byte, st chatState, loopErrCh <-chan error) {
	type wsResponse struct {
		Bdy json.RawMessage `json:"bdy"`
		Cmd int             `json:"cmd"`
	}
	type wsRequest struct {
		Cmd int    `json:"cmd"`
		Ver string `json:"ver"`
	}

	defer conn.Close(websocket.StatusNormalClosure, "closing connection")
	defer close(recv)

	for {
		select {
		case <-ctx.Done():
			return
		case <-loopErrCh:
			return
		case data, ok := <-st.recv:
			if !ok {
				return
			}
			var resp wsResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return
			}
			switch resp.Cmd {
			case cmdPing:
				pong, _ := json.Marshal(wsRequest{Cmd: cmdPong, Ver: "2"})
				select {
				case <-ctx.Done():
					return
				case st.send <- pong:
				}
			case cmdChat, cmdRecentChat:
				var wrapped struct {
					MessageList json.RawMessage `json:"messageList"`
				}
				if err := json.Unmarshal(resp.Bdy, &wrapped); err != nil {
					select {
					case <-ctx.Done():
						return
					case recv <- resp.Bdy:
					}
					continue
				}
				if len(wrapped.MessageList) > 0 {
					select {
					case <-ctx.Done():
						return
					case recv <- wrapped.MessageList:
					}
				}
			}
		case msg, ok := <-send:
			if !ok {
				return
			}
			select {
			case <-ctx.Done():
				return
			case st.send <- msg:
			}
		}
	}
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
