// package unofficial provides useful, but unofficial features for accessing to Chzzk.
// Since these features are not officially supported by the Chzzk API,
// clients should use them with caution and be aware not to be banned or blocked by the Chzzk API.
//
// If you got banned for abusing these features, chzzk-go is not responsible for that.
// Please use these features at your own risk.
package unofficial

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sdkim96/chzzk-go"
	"github.com/sdkim96/chzzk-go/internal/rest"
)

const ChzzkBaseURL = "https://api.chzzk.naver.com"
const NaverGameBaseURL = "https://comm-api.game.naver.com"

// UnofficialChzzk is a client for accessing unofficial features of the Chzzk API.
//
// The embedded Chzzk client is used when the official features need to be called during execution of unofficial features.
//
// The underlying http.Clients, which are both the c and chzzk.c don't share the same configurations except for the transport.
// The transport is shared between the two clients, enabling connection reuse and keep-alive.
type UnofficialChzzk struct {
	c     *http.Client
	chzzk *chzzk.Chzzk

	// user ID hash, empty if anonymous
	uid string

	Chat *ChatService
	Live *LiveService
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
		req = req.Clone(req.Context())
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
		return originalTransport.RoundTrip(req)
	})

	uc := &UnofficialChzzk{chzzk: chz, c: hc}
	uc.initialize()
	return uc, nil
}

// WithCookie returns a new UnofficialChzzk client with the provided NID cookies.
// This enables authenticated features such as sending chat messages via [ChatService.Connect].
// The user ID hash is fetched automatically from the Chzzk API.
//   - endpoint: nng_main/v1/user/getUserStatus
func (u *UnofficialChzzk) WithCookie(ctx context.Context, nidAut, nidSes string) (*UnofficialChzzk, error) {
	u2 := u.copy()

	originalTransport := u2.c.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	u2.c.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req = req.Clone(req.Context())
		req.AddCookie(&http.Cookie{Name: "NID_AUT", Value: nidAut})
		req.AddCookie(&http.Cookie{Name: "NID_SES", Value: nidSes})
		return originalTransport.RoundTrip(req)
	})

	uid, err := u2.userID(ctx)
	if err != nil {
		return nil, fmt.Errorf("unofficial: failed to fetch uid: %w", err)
	}
	u2.uid = uid

	return u2, nil
}

func (u *UnofficialChzzk) initialize() {
	u.Chat = &ChatService{unofficial: u}
	u.Live = &LiveService{unofficial: u}
}

func (u *UnofficialChzzk) copy() *UnofficialChzzk {
	u2 := &UnofficialChzzk{
		chzzk: u.chzzk,
		c:     &http.Client{},
		uid:   u.uid,
	}
	if u.c != nil {
		u2.c.Transport = u.c.Transport
		u2.c.CheckRedirect = u.c.CheckRedirect
		u2.c.Jar = u.c.Jar
		u2.c.Timeout = u.c.Timeout
	}
	u2.initialize()
	return u2
}

func (u *UnofficialChzzk) userID(ctx context.Context) (string, error) {
	p, err := url.JoinPath(NaverGameBaseURL, "nng_main", "v1", "user", "getUserStatus")
	if err != nil {
		return "", err
	}
	type userResp struct {
		chzzk.Response
		Content struct {
			UserIDHash string `json:"userIdHash"`
		} `json:"content"`
	}
	resp, err := rest.Get[userResp](ctx, u.c, p)
	if err != nil {
		return "", err
	}
	if err := chzzk.MightError(resp.Response); err != nil {
		return "", err
	}
	return resp.Content.UserIDHash, nil
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
