package login

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const AuthorizeURL = "https://chzzk.naver.com/account-interlock"

type AuthorizationRequest struct {
	ClientID    string `json:"clientId"`
	RedirectURI string `json:"redirectUri"`
	State       string `json:"state"`
}

type AuthorizationResponse struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

func Authorize(ctx context.Context, c *http.Client, r AuthorizationRequest) (*AuthorizationResponse, error) {
	fullURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&state=%s", AuthorizeURL, r.ClientID, r.RedirectURI, r.State)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var authResp AuthorizationResponse

	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return nil, err
	}

	return &authResp, nil
}
