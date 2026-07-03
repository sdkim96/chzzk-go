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

// AuthClient returns a URL for connecting to the Chzzk session service.
// The user could use this URL to subscribe to various events from Chzzk,
// such as:
//   - Chat messages
//   - Subscription events
func (s *SessionService) AuthClient(ctx context.Context) (string, error) {
	urlStr, err := url.JoinPath(BaseURL, V1, prefixSession, "auth", "client")
	if err != nil {
		return "", fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	return s.auth(ctx, urlStr)
}

func (s *SessionService) AuthUser(ctx context.Context) (string, error) {
	urlStr, err := url.JoinPath(BaseURL, V1, prefixSession, "auth")
	if err != nil {
		return "", fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	return s.auth(ctx, urlStr)
}

func (s *SessionService) auth(ctx context.Context, urlStr string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
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
