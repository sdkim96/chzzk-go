package chzzk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

func (s *UserService) Me(ctx context.Context) (*User, error) {
	url := fmt.Sprintf("%s%s%s/me", BaseURL, V1, prefixUser)
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

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
