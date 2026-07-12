package unofficial

import (
	"context"
	"net/url"

	"github.com/sdkim96/chzzk-go"
	"github.com/sdkim96/chzzk-go/internal/rest"
)

// LiveService provides methods for accessing unofficial live streaming features of the Chzzk API.
type LiveService struct {
	uc *Client
}

// ID retrieves the chat channel ID for a given channel.
// The returned ID is used to connect to the chat WebSocket via [ChatService.ReadOnlyConnect] or [ChatService.Connect].
//   - endpoint: polling/v2/channels/{channelID}/live-status
func (s *LiveService) ID(ctx context.Context, channelID string) (string, error) {
	return s.id(ctx, channelID)
}

func (s *LiveService) id(ctx context.Context, channelID string) (string, error) {
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
	resp, err := rest.Get[LiveChannelResp](ctx, s.uc.httpClient, pURL.String())
	if err != nil {
		return "", err
	}
	if err := chzzk.MightError(resp.Response); err != nil {
		return "", err
	}
	return resp.Content.ChatChannelID, nil
}
