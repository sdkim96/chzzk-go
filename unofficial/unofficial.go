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
	chzzkHttp "github.com/sdkim96/chzzk-go/transport/http"
)

const (
	ChzzkBaseURL     = "https://api.chzzk.naver.com"
	NaverGameBaseURL = "https://comm-api.game.naver.com"
	UserAgent        = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

// Client manages communications for using unofficial features of the Chzzk API.
type Client struct {
	httpClient *http.Client  // HTTP client is used for making requests to the unofficial Chzzk API.
	chzzk      *chzzk.Client // chzzk client is used for making requests to the official Chzzk API.

	// user ID hash, empty if anonymous
	uid string

	Chat *ChatService
	Live *LiveService
}

// New creats a new unofficial client for accessing unofficial features of the Chzzk API.
func New(chz *chzzk.Client, httpClient *http.Client) (*Client, error) {
	if chz == nil {
		return nil, errors.New("unofficial: chzzk client cannot be nil")
	}
	hc := &http.Client{}
	if httpClient != nil {
		cp := *httpClient
		hc = &cp
	}

	originalTransport := hc.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}
	hc.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req = req.Clone(req.Context())
		req.Header.Set("User-Agent", UserAgent)
		return originalTransport.RoundTrip(req)
	})

	uc := &Client{httpClient: hc, chzzk: chz}
	uc.initialize()
	return uc, nil
}

// WithCookie returns a new Client with the provided NID cookies.
// This enables authenticated features such as sending chat messages via [ChatService.Connect].
// The user ID hash is fetched automatically from the Chzzk API.
//   - endpoint: nng_main/v1/user/getUserStatus
func (u *Client) WithCookie(ctx context.Context, nidAut, nidSes string) (*Client, error) {
	u2 := u.copy()

	originalTransport := u2.httpClient.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	u2.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req = req.Clone(req.Context())
		req.AddCookie(&http.Cookie{Name: "NID_AUT", Value: nidAut})
		req.AddCookie(&http.Cookie{Name: "NID_SES", Value: nidSes})
		return originalTransport.RoundTrip(req)
	})

	uid, err := u2.userID(ctx)
	if err != nil {
		return nil, fmt.Errorf("unofficial: failed to fetch user ID: %w", err)
	}
	u2.uid = uid

	return u2, nil
}

func (u *Client) initialize() {
	u.Chat = &ChatService{uc: u}
	u.Live = &LiveService{uc: u}
}

func (u *Client) copy() *Client {
	u2 := &Client{
		chzzk:      u.chzzk,
		httpClient: &http.Client{},
		uid:        u.uid,
	}
	if u.httpClient != nil {
		u2.httpClient.Transport = u.httpClient.Transport
		u2.httpClient.CheckRedirect = u.httpClient.CheckRedirect
		u2.httpClient.Jar = u.httpClient.Jar
		u2.httpClient.Timeout = u.httpClient.Timeout
	}
	u2.initialize()
	return u2
}

func (u *Client) userID(ctx context.Context) (string, error) {
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
	resp, err := chzzkHttp.Get[userResp](ctx, u.httpClient, p)
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
