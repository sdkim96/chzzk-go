// package unofficial provides useful, but unofficial features for accessing to Chzzk.
// Since these features are not officially supported by the Chzzk API,
// clients should use them with caution and be aware not to be banned or blocked by the Chzzk API.
//
// If you got banned for abusing these features, chzzk-go is not responsible for that.
// Please use these features at your own risk.
package unofficial

import (
	"errors"
	"net/http"

	"github.com/sdkim96/chzzk-go"
)

const ChzzkBaseURL = "https://api.chzzk.naver.com"
const NaverGameBaseURL = "https://comm-api.game.naver.com"
const ChatWebSocketURL = "wss://kr-ss1.chat.naver.com/chat"

// UnofficialChzzk is a client for accessing unofficial features of the Chzzk API.
//
// The embedded Chzzk client is used when the official features need to be called during execution of unofficial features.
//
// The underlying http.Clients, which are both the c and chzzk.c don't share the same configurations except for the transport.
// The transport is shared between the two clients, enabling connection reuse and keep-alive.
type UnofficialChzzk struct {
	c     *http.Client
	chzzk *chzzk.Chzzk

	Chat *ChatService
}

func New(chz *chzzk.Chzzk, c *http.Client) (*UnofficialChzzk, error) {
	if chz == nil {
		return nil, errors.New("unofficial: chzzk client cannot be nil")
	}
	hc := &http.Client{}
	if c != nil {
		cp := *c
		hc = &cp
	}

	originalTransport := hc.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}
	hc.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
		return originalTransport.RoundTrip(req)
	})

	uc := &UnofficialChzzk{chzzk: chz, c: hc}
	uc.initialize()
	return uc, nil
}

func (u *UnofficialChzzk) initialize() {
	u.Chat = &ChatService{unofficial: u}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
