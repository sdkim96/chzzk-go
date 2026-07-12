package chzzk

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/sdkim96/chzzk-go/internal/rest"
)

type ChannelService struct {
	c *Client
}

type SubscriptionSort string

const (
	Recent SubscriptionSort = "RECENT"
	Longer SubscriptionSort = "LONGER"
)

type Channel struct {
	ID            string `json:"channelId"`
	Name          string `json:"channelName"`
	ImageURL      string `json:"imageUrl"`
	FollowerCount int    `json:"followerCount"`
	Verified      bool   `json:"verified"`
}

type Manager struct {
	ID          string `json:"managerChannelId"`
	Name        string `json:"managerChannelName"`
	Role        string `json:"userRole"`
	CreatedDate string `json:"createdDate"`
}

type Follower struct {
	ID          string `json:"ChannelId"`
	Name        string `json:"ChannelName"`
	CreatedDate string `json:"createdDate"`
}

type Subscriber struct {
	ID          string `json:"ChannelId"`
	Name        string `json:"ChannelName"`
	Month       int    `json:"month"`
	Tier        int    `json:"tierNo"`
	CreatedDate string `json:"createdDate"`
}

// Get retrieves information for multiple channels by their IDs.
//   - pattern: [BatchGet]
//   - credential: [Chzzk.WithClientAuth]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/channel#undefined
//
// [BatchGet]: https://google.aip.dev/231
func (s *ChannelService) Get(ctx context.Context, ids ...string) ([]Channel, error) {
	if len(ids) > 20 {
		return nil, errors.New("chzzk: cannot request more than 20 channel IDs in a single batch")
	}
	return s.get(ctx, ids...)
}

// Managers retrieves the list of managers for a channel.
// Unlike the original [List] operation, this API does not require pagination, as it returns all managers in a single response.
//   - pattern: [List]
//   - credential: [Chzzk.WithAPIKey]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/channel#undefined-1
//
// [List]: https://google.aip.dev/132
func (s *ChannelService) Managers(ctx context.Context) ([]Manager, error) {
	return s.managers(ctx)
}

// Followers retrieves the list of followers for a channel with pagination support.
//   - pattern: [List]
//   - credential: [Chzzk.WithAPIKey]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/channel#undefined-2
//
// [List]: https://google.aip.dev/132
func (s *ChannelService) Followers(ctx context.Context, page, size *int) ([]Follower, int, error) {
	var (
		defaultPage = 0
		defaultSize = 30
	)
	if page != nil {
		defaultPage = *page
	}
	if size != nil {
		defaultSize = *size
	}
	return s.followers(ctx, defaultPage, defaultSize)
}

// Subscribers retrieves the list of subscribers for a channel with pagination support.
//   - pattern: [List]
//   - credential: [Chzzk.WithAPIKey]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/channel#undefined-3
//
// [List]: https://google.aip.dev/132
func (s *ChannelService) Subscribers(ctx context.Context, page, size *int, sort *SubscriptionSort) ([]Subscriber, int, error) {
	var (
		defaultPage = 0
		defaultSize = 30
		defaultSort = Recent
	)
	if page != nil {
		defaultPage = *page
	}
	if size != nil {
		defaultSize = *size
	}
	if sort != nil {
		defaultSort = *sort
	}
	return s.subscribers(ctx, defaultPage, defaultSize, defaultSort)
}

func (s *ChannelService) get(ctx context.Context, ids ...string) ([]Channel, error) {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixChannel)
	if err != nil {
		return nil, err
	}
	URL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	q := URL.Query()
	for _, id := range ids {
		q.Add("channelIds", id)
	}
	URL.RawQuery = q.Encode()
	type ChannelResp struct {
		Response
		Content struct {
			Data []Channel `json:"data"`
		} `json:"content"`
	}
	resp, err := rest.Get[ChannelResp](ctx, s.c.httpClient, URL.String())
	if err != nil {
		return nil, err
	}
	if err := MightError(resp.Response); err != nil {
		return nil, err
	}
	var channels []Channel
	if resp != nil {
		channels = resp.Content.Data
	}
	return channels, nil
}

func (s *ChannelService) managers(ctx context.Context) ([]Manager, error) {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixChannel, "streaming-roles")
	if err != nil {
		return nil, err
	}
	type ManagerResp struct {
		Response
		Content struct {
			Data []Manager `json:"data"`
		} `json:"content"`
	}
	resp, err := rest.Get[ManagerResp](ctx, s.c.httpClient, u)
	if err != nil {
		return nil, err
	}
	if err := MightError(resp.Response); err != nil {
		return nil, err
	}
	var managers []Manager
	if resp != nil {
		managers = resp.Content.Data
	}
	return managers, nil
}

func (s *ChannelService) followers(ctx context.Context, page, size int) ([]Follower, int, error) {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixChannel, "followers")
	if err != nil {
		return nil, 0, err
	}
	URL, err := url.Parse(u)
	if err != nil {
		return nil, 0, err
	}
	q := URL.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("size", strconv.Itoa(size))
	URL.RawQuery = q.Encode()

	type FollowerResp struct {
		Response
		Content struct {
			Data []Follower `json:"data"`
		} `json:"content"`
	}
	resp, err := rest.Get[FollowerResp](ctx, s.c.httpClient, URL.String())
	if err != nil {
		return nil, 0, err
	}
	if err := MightError(resp.Response); err != nil {
		return nil, 0, err
	}
	var followers []Follower
	if resp != nil {
		followers = resp.Content.Data
	}
	return followers, page + 1, nil
}

func (s *ChannelService) subscribers(ctx context.Context, page, size int, sort SubscriptionSort) ([]Subscriber, int, error) {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixChannel, "subscribers")
	if err != nil {
		return nil, 0, err
	}
	URL, err := url.Parse(u)
	if err != nil {
		return nil, 0, err
	}
	q := URL.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("size", strconv.Itoa(size))
	q.Set("sort", string(sort))
	URL.RawQuery = q.Encode()

	type SubscriberResp struct {
		Response
		Content struct {
			Data []Subscriber `json:"data"`
		} `json:"content"`
	}
	resp, err := rest.Get[SubscriberResp](ctx, s.c.httpClient, URL.String())
	if err != nil {
		return nil, 0, err
	}
	if err := MightError(resp.Response); err != nil {
		return nil, 0, err
	}
	var subscribers []Subscriber
	if resp != nil {
		subscribers = resp.Content.Data
	}
	return subscribers, page + 1, nil
}
