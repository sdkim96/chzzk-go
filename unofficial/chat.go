package unofficial

import (
	"context"
	"net/url"

	"github.com/sdkim96/chzzk-go"
	"github.com/sdkim96/chzzk-go/internal/rest"
)

type ChatService struct {
	unofficial *UnofficialChzzk
}

func (s *ChatService) LiveID(ctx context.Context, channelID string) (string, error) {
	return s.liveID(ctx, channelID)
}

func (s *ChatService) liveID(ctx context.Context, channelID string) (string, error) {
	u, err := url.JoinPath(ChzzkBaseURL, "polling", "v2", "channels", channelID, "live-status")
	if err != nil {
		return "", err
	}
	pURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	type LiveChannelResp struct {
		chzzk.Response
		Content struct {
			ChatChannelID string `json:"chatChannelId"`
		} `json:"content"`
	}
	resp, err := rest.Get[LiveChannelResp](ctx, s.unofficial.c, pURL.String())
	if err != nil {
		return "", err
	}
	if err := chzzk.MightError(resp.Response); err != nil {
		return "", err
	}
	return resp.Content.ChatChannelID, nil

}

func (s *ChatService) AccessToken(ctx context.Context, liveID string) (string, error) {
	return s.accessToken(ctx, liveID)
}

func (s *ChatService) accessToken(ctx context.Context, liveID string) (string, error) {
	u, err := url.JoinPath(NaverGameBaseURL, "nng_main", "v1", "chats", "access-token")
	if err != nil {
		return "", err
	}
	pURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	q := pURL.Query()
	q.Set("channelId", liveID)
	q.Set("chatType", "STREAMING")
	pURL.RawQuery = q.Encode()

	type AccessTokenResp struct {
		chzzk.Response
		Content struct {
			AccessToken string `json:"accessToken"`
		} `json:"content"`
	}
	resp, err := rest.Get[AccessTokenResp](ctx, s.unofficial.c, pURL.String())
	if err != nil {
		return "", err
	}
	if err := chzzk.MightError(resp.Response); err != nil {
		return "", err
	}
	return resp.Content.AccessToken, nil
}
