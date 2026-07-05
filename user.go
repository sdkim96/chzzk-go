package chzzk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// UserService handles APIs prefixed with /users
type UserService struct {
	chzzk *Chzzk
}

const prefixUser = "/users"

type User struct {
	ChannelID   string `json:"channelId"`
	ChannelName string `json:"channelName"`
}

// Me retreives the current user's information from the Chzzk API.
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/user
func (s *UserService) Me(ctx context.Context) (*User, error) {
	url, err := url.JoinPath(BaseURL, OpenV1, prefixUser, "me")
	if err != nil {
		return nil, fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	var userResp struct {
		OnSuccess
		Content User `json:"content"`
	}
	err = json.NewDecoder(resp.Body).Decode(&userResp)
	if err != nil {
		return nil, err
	}
	return &userResp.Content, nil
}
