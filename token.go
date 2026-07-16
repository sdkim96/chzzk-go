package chzzk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	chzzkHttp "github.com/sdkim96/chzzk-go/transport/http"
)

type TokenService struct {
	c *Client
}

type GrantType string

const (
	GrantTypeAuthorizationCode GrantType = "authorization_code"
	GrantTypeRefreshToken      GrantType = "refresh_token"
)

type TokenRequest struct {
	GrantType    GrantType `json:"grantType"`
	ClientID     string    `json:"clientId"`
	ClientSecret string    `json:"clientSecret"`
}
type TokenNewRequest struct {
	TokenRequest
	Code  string `json:"code"`
	State string `json:"state"`
}
type TokenRefreshRequest struct {
	TokenRequest
	RefreshToken string `json:"refreshToken"`
}

type TokenResponse struct {
	AccessToken  string  `json:"accessToken"`
	RefreshToken string  `json:"refreshToken"`
	ExpiresIn    int     `json:"expiresIn"`
	TokenType    string  `json:"tokenType"`
	Scope        *string `json:"scope,omitempty"`
}

type RevokeTokenRequest struct {
	ClientID      string `json:"clientId"`
	ClientSecret  string `json:"clientSecret"`
	Token         string `json:"token"`
	TokenTypeHint string `json:"tokenTypeHint"`
}

// New requests a new access token using the provided authorization code and state.
//   - pattern: [Create]
//   - credential: [Chzzk.WithClientAuth]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/authorization#access-token
//
// [Create]: https://google.aip.dev/133
func (s *TokenService) New(ctx context.Context, r TokenNewRequest) (*TokenResponse, error) {
	return s.token(ctx, r)
}

// Refresh requests a new access token using the provided refresh token.
//   - pattern: [Create]
//   - credential: [Chzzk.WithClientAuth]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/authorization#access-token-1
//
// [Create]: https://google.aip.dev/133
func (s *TokenService) Refresh(ctx context.Context, r TokenRefreshRequest) (*TokenResponse, error) {
	return s.token(ctx, r)
}

// Revoke revokes the provided access or refresh token.
//   - pattern: [Delete]
//   - credential: [Chzzk.WithClientAuth]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/authorization#access-token-2
//
// [Delete]: https://google.aip.dev/135
func (s *TokenService) Revoke(ctx context.Context, r RevokeTokenRequest) error {
	return s.revoke(ctx, r)
}

func (s *TokenService) token(ctx context.Context, r any) (*TokenResponse, error) {
	u, err := url.JoinPath(BaseURL, AuthV1, prefixToken)
	if err != nil {
		return nil, fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	jsonData, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}
	type TokenResp struct {
		Response
		Content TokenResponse `json:"content"`
	}
	resp, err := chzzkHttp.Post[TokenResp](ctx, s.c.httpClient, u, jsonData)
	if err != nil {
		return nil, err
	}
	if err := MightError(resp.Response); err != nil {
		return nil, err
	}
	return &resp.Content, nil
}

func (s *TokenService) revoke(ctx context.Context, r RevokeTokenRequest) error {
	u, err := url.JoinPath(BaseURL, AuthV1, prefixToken, "revoke")
	if err != nil {
		return fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	jsonData, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("failed to marshal revoke token request: %w", err)
	}
	resp, err := chzzkHttp.Post[Response](ctx, s.c.httpClient, u, jsonData)
	if err != nil {
		return err
	}
	if err := MightError(*resp); err != nil {
		return err
	}

	return nil
}
