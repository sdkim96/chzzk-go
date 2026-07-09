package unofficial

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/coder/websocket"
	"github.com/sdkim96/chzzk-go"
	"github.com/sdkim96/chzzk-go/internal/rest"
	iconn "github.com/sdkim96/chzzk-go/unofficial/internal"
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
	unofficial *UnofficialChzzk
}

type chatState struct {
	op string

	// channels
	recv  <-chan []byte
	send  chan<- []byte
	errCh <-chan error

	liveID string
	token  *ChatToken
	sid    string // session ID from CmdConnected response
}

// ChatToken contains the access token and extra token returned by the Chzzk chat API.
type ChatToken struct {
	AccessToken string
	ExtraToken  string
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
func (s *ChatService) ReadOnlyConnect(ctx context.Context, liveID string, token *ChatToken) (<-chan []byte, error) {
	recv, _, _, err := s.connect(ctx, "r", liveID, token)
	return recv, err
}

// Connect establishes a bidirectional WebSocket connection to a Chzzk chat channel.
// Returns recv for incoming messages, send for outgoing messages, and the session ID.
// The session ID (sid) must be included in outgoing chat messages (cmd 3101).
// The connection is closed when ctx is cancelled.
//   - protocol: wss://kr-ss{1-9}.chat.naver.com/chat
//   - credential: [UnofficialChzzk.WithCookie]
func (s *ChatService) Connect(ctx context.Context, liveID string, token *ChatToken) (recv <-chan []byte, send chan<- []byte, sid string, err error) {
	if s.unofficial.uid == "" {
		return nil, nil, "", fmt.Errorf("chat: Connect requires authentication. use WithCookie first, or use ReadOnlyConnect")
	}
	return s.connect(ctx, "rw", liveID, token)
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
	resp, err := rest.Get[AccessTokenResp](ctx, s.unofficial.c, pURL.String())
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

func (s *ChatService) connect(ctx context.Context, op string, liveID string, token *ChatToken) (<-chan []byte, chan<- []byte, string, error) {
	if liveID == "" {
		return nil, nil, "", fmt.Errorf("chat: liveID cannot be empty")
	}
	wsURL := fmt.Sprintf("wss://kr-ss%d.chat.naver.com/chat", chatServerID(liveID))

	conn := iconn.NewConn(s.unofficial.c)
	if err := conn.Dial(ctx, wsURL); err != nil {
		return nil, nil, "", fmt.Errorf("chat: dial failed: %w", err)
	}

	internalSend := make(chan []byte, 1)
	rawRecv, errCh, err := conn.Start(ctx, internalSend)
	if err != nil {
		conn.Close(ctx, websocket.StatusNormalClosure, "start failed")
		return nil, nil, "", fmt.Errorf("chat: start failed: %w", err)
	}

	st := chatState{
		op:     op,
		recv:   rawRecv,
		send:   internalSend,
		errCh:  errCh,
		liveID: liveID,
		token:  token,
	}

	if err := s.handshake(ctx, &st); err != nil {
		conn.Close(ctx, websocket.StatusNormalClosure, "handshake failed")
		return nil, nil, "", err
	}

	outRecv := make(chan []byte)
	outSend := make(chan []byte)

	go s.loop(ctx, conn, outRecv, outSend, st)

	return outRecv, outSend, st.sid, nil
}

func (s *ChatService) handshake(ctx context.Context, st *chatState) error {
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
		uid = s.unofficial.uid
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
	connectBytes, err := json.Marshal(connectReq)
	if err != nil {
		return fmt.Errorf("chat: marshal connect: %w", err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case st.send <- connectBytes:
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err, ok := <-st.errCh:
			if !ok {
				return fmt.Errorf("chat: connection closed before handshake")
			}
			return fmt.Errorf("chat: handshake error: %w", err)
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

func (s *ChatService) loop(ctx context.Context, conn *iconn.Conn, recv chan<- []byte, send <-chan []byte, st chatState) {
	type wsResponse struct {
		Bdy json.RawMessage `json:"bdy"`
		Cmd int             `json:"cmd"`
	}
	type wsRequest struct {
		Cmd int    `json:"cmd"`
		Ver string `json:"ver"`
	}

	defer conn.Close(ctx, websocket.StatusNormalClosure, "closing connection")
	defer close(recv)

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-st.errCh:
			if !ok {
				return
			}
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
