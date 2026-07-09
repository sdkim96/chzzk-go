package chzzk

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sdkim96/chzzk-go/internal/rest"
)

type SessionService struct {
	chzzk *Chzzk
}

// AuthClient returns a URL for connecting to the Chzzk session service via client credentials.
//   - pattern: [Get]
//   - credential: [Chzzk.WithClientAuth]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/session#undefined
//
// [Get]: https://google.aip.dev/131
func (s *SessionService) AuthClient(ctx context.Context) (string, error) {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixSession, "auth", "client")
	if err != nil {
		return "", fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	return s.auth(ctx, u)
}

// AuthUser returns a URL for connecting to the Chzzk session service via user credentials.
//   - pattern: [Get]
//   - credential: [Chzzk.WithAPIKey]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/session#undefined-1
//
// [Get]: https://google.aip.dev/131
func (s *SessionService) AuthUser(ctx context.Context) (string, error) {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixSession, "auth")
	if err != nil {
		return "", fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	return s.auth(ctx, u)
}

// SubscribeChat subscribes to chat events for the given session key.
//   - pattern: [Create]
//   - credential: [Chzzk.WithClientAuth] or [Chzzk.WithAPIKey]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/session#undefined-5
//
// [Create]: https://google.aip.dev/133
func (s *SessionService) SubscribeChat(ctx context.Context, sk string) error {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixSession, "events", "subscribe", "chat")
	if err != nil {
		return fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	return s.sub(ctx, u, sk)
}

// UnSubscribeChat unsubscribes from chat events for the given session key.
//   - pattern: [Delete]
//   - credential: [Chzzk.WithClientAuth] or [Chzzk.WithAPIKey]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/session#undefined-6
//
// [Delete]: https://google.aip.dev/135
func (s *SessionService) UnSubscribeChat(ctx context.Context, sk string) error {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixSession, "events", "unsubscribe", "chat")
	if err != nil {
		return fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	return s.sub(ctx, u, sk)
}

func (s *SessionService) auth(ctx context.Context, u string) (string, error) {
	type AuthResp struct {
		Response
		Content struct {
			URL string `json:"url"`
		} `json:"content"`
	}
	authResp, err := rest.Get[AuthResp](ctx, s.chzzk.c, u)
	if err != nil {
		return "", err
	}
	if err := MightError(authResp.Response); err != nil {
		return "", err
	}
	return authResp.Content.URL, nil
}

func (s *SessionService) sub(ctx context.Context, u, sk string) error {
	URL, err := url.Parse(u)
	if err != nil {
		return err
	}
	q := URL.Query()
	q.Set("sessionKey", sk)
	URL.RawQuery = q.Encode()

	_, err = rest.Post[Response](ctx, s.chzzk.c, URL.String(), nil)
	if err != nil {
		return err
	}
	return nil
}
