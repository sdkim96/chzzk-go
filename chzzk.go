package chzzk

import (
	"fmt"
	"net/http"
)

const (
	Version     = "0.3.0"
	BaseURL     = "https://openapi.chzzk.naver.com"
	OpenV1      = "/open/v1"
	AuthV1      = "/auth/v1"
	ContentType = "application/json"
	UserAgent   = "chzzk-go/" + Version

	// The service prefixes
	prefixToken    = "/tokens"
	prefixUser     = "/users"
	prefixSession  = "/sessions"
	prefixChannel  = "/channels"
	prefixCategory = "/categories"
	prefixLive     = "/lives"
)

type Chzzk struct {
	c *http.Client

	// The services
	Token    *TokenService
	User     *UserService
	Session  *SessionService
	Channel  *ChannelService
	Category *CategoryService
	Live     *LiveService
}

// New creates a new Chzzk client with the provided http.Client.
// If the provided http.Client is nil, a new http.Client will be created.
func New(c *http.Client) *Chzzk {
	if c == nil {

		// we don't use the default http.Client. (http.DefaultClient)
		// Since the default client could be shared by all codebase,
		// it can be polluted by other code that modifies the default client.
		//
		// Note: this does not mean that we always create a new Transport for each client.
		// The Transport is SHARED between all clients, enabling connection reuse and keep-alive.
		c = &http.Client{}
	} else {

		// This prevents the mutation of the original http.Client
		// from affecting the new Chzzk client.
		cp := *c
		c = &cp
	}

	// We always set the User-Agent header for all requests.
	originalTransport := c.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	c.Transport = roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			req2 := req.Clone(req.Context())
			req2.Header.Set("User-Agent", UserAgent)
			req2.Header.Set("Content-Type", ContentType)
			return originalTransport.RoundTrip(req2)
		},
	)
	chz := &Chzzk{c: c}
	chz.initialize()
	return chz
}

// Response is a struct that Chzzk always returns in the response body,
// with the actual response type in the Content field.
//
// Each Service should embed Response to its own response struct,
// including the Content field with the actual response type.
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	// Content any
}

// WithClientAuth returns a new Chzzk client with the provided client ID and secret.
// You must either use WithClientAuth or WithAPIKey, not both. Using both will cause unexpected behavior.
//
// Check the Chzzk API documentation to see further details: https://chzzk.gitbook.io/chzzk/chzzk-api/tips#access-token-api
func (chz *Chzzk) WithClientAuth(ID, secret string) *Chzzk {
	chz2 := chz.copy()

	originalTransport := chz2.c.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	chz2.c.Transport = roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			req2 := req.Clone(req.Context())

			req2.Header.Set("Client-Id", ID)
			req2.Header.Set("Client-Secret", secret)

			return originalTransport.RoundTrip(req2)
		},
	)

	return chz2
}

// WithAPIKey returns a new Chzzk client with the provided API key.
func (chz *Chzzk) WithAPIKey(apiKey string) *Chzzk {
	chz2 := chz.copy()

	// captures the original transport
	originalTransport := chz2.c.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	// replace the `transport Function` with a new one that adds the Authorization header,
	// not the original transport.
	chz2.c.Transport = roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			req2 := req.Clone(req.Context())
			v := fmt.Sprintf("Bearer %s", apiKey)
			req2.Header.Set("Authorization", v)
			return originalTransport.RoundTrip(req2)
		},
	)

	return chz2
}

// WithHooks returns a new Chzzk client with the provided hooks for request and response.
// Do not modify (read or write req or resp) in the hooks, as it may cause unexpected behavior.
// The recommended way is to clone the request and response in the hooks, and modify the cloned objects.
// Or just log the request and response in the hooks, without modifying them.
func (chz *Chzzk) WithHooks(bef func(req *http.Request), aft func(resp *http.Response)) *Chzzk {
	chz2 := chz.copy()

	originalTransport := chz2.c.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	chz2.c.Transport = roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			req2 := req.Clone(req.Context())
			if bef != nil {
				bef(req2)
			}
			resp, err := originalTransport.RoundTrip(req2)
			if err != nil {
				return resp, err
			}
			if aft != nil {
				aft(resp)
			}
			return resp, err
		},
	)
	return chz2
}

// initialize initializes the Chzzk client,
// only copy-safe fields should be set in this function.
// It is called by New and copy.
func (chz *Chzzk) initialize() {
	chz.User = &UserService{chzzk: chz}
	chz.Token = &TokenService{chzzk: chz}
	chz.Session = &SessionService{chzzk: chz}
	chz.Channel = &ChannelService{chzzk: chz}
	chz.Category = &CategoryService{chzzk: chz}
	chz.Live = &LiveService{chzzk: chz}
}

// copy copies the Chzzk client, returning a new instance with the same configuration.
// But same configuration does not mean the same object / pointer.
// Most of the fields are copied, but the transport is shared.
func (chz *Chzzk) copy() *Chzzk {

	chz2 := &Chzzk{c: &http.Client{}}
	chz2.initialize()

	// Using the same tranport!
	if chz.c != nil {
		chz2.c.Transport = chz.c.Transport
		chz2.c.CheckRedirect = chz.c.CheckRedirect
		chz2.c.Jar = chz.c.Jar
		chz2.c.Timeout = chz.c.Timeout
	}

	// could be nil transport.
	return chz2
}

func mightError(resp Response) error {
	if resp.Code >= 200 && resp.Code < 300 {
		return nil
	}
	switch resp.Code {
	case 400:
		return fmt.Errorf("400: parameter error, %s", resp.Message)
	case 401:
		return fmt.Errorf("401: unauthorized error, %s", resp.Message)
	case 403:
		return fmt.Errorf("403: forbidden error, %s", resp.Message)
	case 404:
		return fmt.Errorf("404: not found error, %s", resp.Message)
	case 429:
		return fmt.Errorf("429: exceeded quota error, %s", resp.Message)
	case 500:
		return fmt.Errorf("500: internal error, %s", resp.Message)
	default:
		return fmt.Errorf("%d: %s", resp.Code, resp.Message)
	}
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
