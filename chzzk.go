package chzzk

import (
	"fmt"
	"net/http"
)

const (
	Version     = "0.5.2"
	BaseURL     = "https://openapi.chzzk.naver.com"
	OpenV1      = "/open/v1"
	AuthV1      = "/auth/v1"
	ContentType = "application/json"
	UserAgent   = "chzzk-go/" + Version

	// NOTE: the token endpoint does not have 's' suffix.
	prefixToken    = "/token"
	prefixUser     = "/users"
	prefixSession  = "/sessions"
	prefixChannel  = "/channels"
	prefixCategory = "/categories"
	prefixLive     = "/lives"
	prefixChat     = "/chats"
)

type Client struct {
	httpClient *http.Client

	// The services
	Token    *TokenService
	User     *UserService
	Session  *SessionService
	Channel  *ChannelService
	Category *CategoryService
	Live     *LiveService
	Chat     *ChatService
}

// New creates a new Chzzk client with the provided http.Client.
// If the provided http.Client is nil, a new http.Client will be created.
func New(httpClient *http.Client) *Client {
	if httpClient == nil {

		// we don't use the default http.Client. (http.DefaultClient)
		// Since the default client could be shared by all codebase,
		// it can be polluted by other code that modifies the default client.
		//
		// Note: this does not mean that we always create a new Transport for each client.
		// The Transport is SHARED between all clients, enabling connection reuse and keep-alive.
		httpClient = &http.Client{}
	} else {

		// This prevents the mutation of the original http.Client
		// from affecting the new Chzzk client.
		cp := *httpClient
		httpClient = &cp
	}

	// We always set the User-Agent header for all requests.
	originalTransport := httpClient.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	httpClient.Transport = roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			req2 := req.Clone(req.Context())
			req2.Header.Set("User-Agent", UserAgent)
			req2.Header.Set("Content-Type", ContentType)
			return originalTransport.RoundTrip(req2)
		},
	)
	cl := &Client{httpClient: httpClient}
	cl.initialize()
	return cl
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
func (c *Client) WithClientAuth(ID, secret string) *Client {
	c2 := c.copy()

	originalTransport := c2.httpClient.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	c2.httpClient.Transport = roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			req2 := req.Clone(req.Context())

			req2.Header.Set("Client-Id", ID)
			req2.Header.Set("Client-Secret", secret)

			return originalTransport.RoundTrip(req2)
		},
	)

	return c2
}

// WithAPIKey returns a new Chzzk client with the provided API key.
func (c *Client) WithAPIKey(apiKey string) *Client {
	c2 := c.copy()

	// captures the original transport
	originalTransport := c2.httpClient.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	// replace the `transport Function` with a new one that adds the Authorization header,
	// not the original transport.
	c2.httpClient.Transport = roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			req2 := req.Clone(req.Context())
			v := fmt.Sprintf("Bearer %s", apiKey)
			req2.Header.Set("Authorization", v)
			return originalTransport.RoundTrip(req2)
		},
	)

	return c2
}

// WithHooks returns a new Chzzk client with the provided hooks for request and response.
// Do not modify (read or write req or resp) in the hooks, as it may cause unexpected behavior.
// The recommended way is to clone the request and response in the hooks, and modify the cloned objects.
// Or just log the request and response in the hooks, without modifying them.
func (c *Client) WithHooks(bef func(req *http.Request), aft func(resp *http.Response)) *Client {
	c2 := c.copy()

	originalTransport := c2.httpClient.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	c2.httpClient.Transport = roundTripperFunc(
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
	return c2
}

// MightError checks the response body for inspecting error from the server.
//
// The Chzzk API does not return HTTP error. Rather, it always
// embeds the [Response] field in the response body, which contains the error code and message.
func MightError(resp Response) error {
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

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}

// initialize initializes the Chzzk client,
// only copy-safe fields should be set in this function.
// It is called by New and copy.
func (c *Client) initialize() {
	c.User = &UserService{c: c}
	c.Token = &TokenService{c: c}
	c.Session = &SessionService{c: c}
	c.Channel = &ChannelService{c: c}
	c.Category = &CategoryService{c: c}
	c.Live = &LiveService{c: c}
	c.Chat = &ChatService{c: c}
}

// copy copies the Chzzk client, returning a new instance with the same configuration.
// But same configuration does not mean the same object / pointer.
// Most of the fields are copied, but the transport is shared.
func (c *Client) copy() *Client {

	c2 := &Client{httpClient: &http.Client{}}
	c2.initialize()

	// Using the same tranport!
	if c.httpClient != nil {
		c2.httpClient.Transport = c.httpClient.Transport
		c2.httpClient.CheckRedirect = c.httpClient.CheckRedirect
		c2.httpClient.Jar = c.httpClient.Jar
		c2.httpClient.Timeout = c.httpClient.Timeout
	}

	// could be nil transport.
	return c2
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
