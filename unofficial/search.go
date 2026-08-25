package unofficial

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sdkim96/chzzk-go"
	chzzkHttp "github.com/sdkim96/chzzk-go/transport/http"
)

// SearchService provides methods for searching channels on Chzzk.
type SearchService struct {
	uc *Client
}

// Channels searches for channels matching the given keyword.
// Returns matching channels along with the next offset for pagination.
//   - endpoint: service/v1/search/channels?keyword={keyword}&size={size}&offset={offset}
func (s *SearchService) Channels(ctx context.Context, keyword string, size, offset *int) ([]SearchedChannel, int, error) {
	var (
		defaultSize   = 13
		defaultOffset = 0
	)
	if size != nil {
		defaultSize = *size
	}
	if offset != nil {
		defaultOffset = *offset
	}
	return s.channels(ctx, keyword, defaultSize, defaultOffset)
}

func (s *SearchService) channels(ctx context.Context, keyword string, size, offset int) ([]SearchedChannel, int, error) {
	u, err := url.JoinPath(ChzzkBaseURL, "service", "v1", "search", "channels")
	if err != nil {
		return nil, 0, err
	}
	pURL, err := url.Parse(u)
	if err != nil {
		return nil, 0, err
	}
	q := pURL.Query()
	q.Set("keyword", keyword)
	q.Set("size", fmt.Sprintf("%d", size))
	q.Set("offset", fmt.Sprintf("%d", offset))
	pURL.RawQuery = q.Encode()

	type channelData struct {
		Channel SearchedChannel `json:"channel"`
	}
	type searchResp struct {
		chzzk.Response
		Content struct {
			Size       int           `json:"size"`
			NextOffset int           `json:"nextOffset"`
			Data       []channelData `json:"data"`
		} `json:"content"`
	}

	resp, err := chzzkHttp.Get[searchResp](ctx, s.uc.httpClient, pURL.String())
	if err != nil {
		return nil, 0, err
	}
	if err := chzzk.MightError(resp.Response); err != nil {
		return nil, 0, err
	}

	channels := make([]SearchedChannel, len(resp.Content.Data))
	for i, d := range resp.Content.Data {
		channels[i] = d.Channel
	}
	return channels, resp.Content.NextOffset, nil
}
