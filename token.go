package chzzk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type TokenService struct {
	chzzk *Chzzk
}

type GrantType string

const (
	prefixToken                          = "/token"
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

// NewToken requests a new access token using the provided authorization code and state.
// You must use this API by [WithClientAuth] only.
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/authorization#access-token
func (s *TokenService) NewToken(ctx context.Context, r TokenNewRequest) (*TokenResponse, error) {
	return s.token(ctx, r)
}

// RefreshToken requests a new access token using the provided refresh token.
// You must use this API by [WithClientAuth] only.
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/authorization#access-token-1
func (s *TokenService) RefreshToken(ctx context.Context, r TokenRefreshRequest) (*TokenResponse, error) {
	return s.token(ctx, r)
}

// RevokeToken revokes the provided access or refresh token.
// You must use this API by [WithClientAuth] only.
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/authorization#access-token-2
func (s *TokenService) RevokeToken(ctx context.Context, r RevokeTokenRequest) error {
	url, err := url.JoinPath(BaseURL, AuthV1, prefixToken, "revoke")
	if err != nil {
		return fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	jsonData, err := json.Marshal(r)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	resp, err := s.chzzk.c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = mightError(resp)
	if err != nil {
		return err
	}

	return nil
}

func (s *TokenService) token(ctx context.Context, r any) (*TokenResponse, error) {
	url, err := url.JoinPath(BaseURL, AuthV1, prefixToken)
	if err != nil {
		return nil, fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	jsonData, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	resp, err := s.chzzk.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = mightError(resp)
	if err != nil {
		return nil, err
	}

	var tokenResp struct {
		OnSuccess
		Content TokenResponse `json:"content"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp.Content, nil
}
