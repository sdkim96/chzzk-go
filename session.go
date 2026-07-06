package chzzk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type SessionService struct {
	chzzk *Chzzk
}

const prefixSession = "/sessions"

// AuthClient returns a URL for connecting to the Chzzk session service
// via Authorization of client credentials.
// You must use this API by [Chzzk.WithClientAuth] only.
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/session#undefined
func (s *SessionService) AuthClient(ctx context.Context) (string, error) {
	url, err := url.JoinPath(BaseURL, OpenV1, prefixSession, "auth", "client")
	if err != nil {
		return "", fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	return s.auth(ctx, url)
}

// AuthUser returns a URL for connecting to the Chzzk session service
// via Authorization of user credentials.
// You must use this API by [Chzzk.WithAPIKey] only.
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/session#undefined-1
func (s *SessionService) AuthUser(ctx context.Context) (string, error) {
	url, err := url.JoinPath(BaseURL, OpenV1, prefixSession, "auth")
	if err != nil {
		return "", fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	return s.auth(ctx, url)
}

// SubscribeChat subscribes to chat events for the given session key.
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/session#undefined-5
func (s *SessionService) SubscribeChat(ctx context.Context, sk string) error {
	urlStr, err := url.JoinPath(BaseURL, OpenV1, prefixSession, "events", "subscribe", "chat")
	if err != nil {
		return fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	return s.sub(ctx, urlStr, sk)
}

// UnSubscribeChat unsubscribes from chat events for the given session key.
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/session#undefined-6
func (s *SessionService) UnSubscribeChat(ctx context.Context, sk string) error {
	urlStr, err := url.JoinPath(BaseURL, OpenV1, prefixSession, "events", "unsubscribe", "chat")
	if err != nil {
		return fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	return s.sub(ctx, urlStr, sk)
}

func (s *SessionService) auth(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := s.chzzk.c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	err = mightError(resp)
	if err != nil {
		return "", err
	}

	var authResp struct {
		OnSuccess
		Content struct {
			URL string `json:"url"`
		} `json:"content"`
	}
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return "", err
	}

	return authResp.Content.URL, nil
}

func (s *SessionService) sub(ctx context.Context, urlStr, sk string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("sessionKey", sk)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return err
	}
	resp, err := s.chzzk.c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return mightError(resp)
}
