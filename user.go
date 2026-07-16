package chzzk

import (
	"context"
	"fmt"
	"net/url"

	chzzkHttp "github.com/sdkim96/chzzk-go/transport/http"
)

// UserService handles APIs prefixed with /users
type UserService struct {
	c *Client
}

type User struct {
	ChannelID   string `json:"channelId"`
	ChannelName string `json:"channelName"`
}

// Me retrieves the current user's information from the Chzzk API.
//   - pattern: [Get]
//   - credential: [Chzzk.WithAPIKey]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/user
//
// [Get]: https://google.aip.dev/131
func (s *UserService) Me(ctx context.Context) (*User, error) {
	return s.me(ctx)
}

func (s *UserService) me(ctx context.Context) (*User, error) {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixUser, "me")
	if err != nil {
		return nil, fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	type UserResp struct {
		Response
		Content User `json:"content"`
	}
	resp, err := chzzkHttp.Get[UserResp](ctx, s.c.httpClient, u)
	if err != nil {
		return nil, err
	}
	if err := MightError(resp.Response); err != nil {
		return nil, err
	}
	return &resp.Content, nil
}
