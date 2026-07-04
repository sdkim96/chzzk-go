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

const prefixToken = "/token"

type TokenRequest struct {
	GrantType    string  `json:"grantType"`
	ClientID     string  `json:"clientId"`
	ClientSecret string  `json:"clientSecret"`
	Code         *string `json:"code,omitempty"`
	RedirectURI  *string `json:"redirectUri,omitempty"`
	RefreshToken *string `json:"refreshToken,omitempty"`
}

type TokenResponse struct {
	AccessToken  string  `json:"accessToken"`
	RefreshToken string  `json:"refreshToken"`
	ExpiresIn    string  `json:"expiresIn"`
	TokenType    string  `json:"tokenType"`
	Scope        *string `json:"scope,omitempty"`
}

type RevokeTokenRequest struct {
	ClientID      string `json:"clientId"`
	ClientSecret  string `json:"clientSecret"`
	Token         string `json:"token"`
	TokenTypeHint string `json:"tokenTypeHint"`
}

func (s *TokenService) GetToken(ctx context.Context, r TokenRequest) (*TokenResponse, error) {
	url, err := url.JoinPath(BaseURL, V1, prefixToken)
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

	var tokenResp TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}
func (s *TokenService) UpdateToken(ctx context.Context, r TokenRequest) (*TokenResponse, error) {
	url, err := url.JoinPath(BaseURL, V1, prefixToken)
	if err != nil {
		return nil, fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	jsonData, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(jsonData))
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

	var tokenResp TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}
func (s *TokenService) RevokeToken(ctx context.Context, r RevokeTokenRequest) error {
	url, err := url.JoinPath(BaseURL, V1, prefixToken, "revoke")
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
