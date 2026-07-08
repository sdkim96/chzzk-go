package chzzk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/sdkim96/chzzk-go/internal/rest"
)

// LiveService provides methods for accessing live streaming features of the Chzzk API.
type LiveService struct {
	chzzk *Chzzk
}

type Live struct {
	ID                int      `json:"liveId"`
	Title             string   `json:"liveTitle"`
	ThumbnailImageURL string   `json:"liveThumbnailImageUrl"`
	ViewerCount       int      `json:"concurrentUserCount"`
	StartDate         string   `json:"openDate"`
	IsAdult           bool     `json:"adult"`
	Tags              []string `json:"tags"`
	CategoryID        string   `json:"liveCategory"`
	CategoryType      string   `json:"categoryType"`
	CategoryValue     string   `json:"liveCategoryValue"`
	ChannelID         string   `json:"channelId"`
	ChannelName       string   `json:"channelName"`
	ChannelImageURL   string   `json:"channelImageUrl"`
}

type LiveSetting struct {
	Title            string   `json:"defaultLiveTitle"`
	Tags             []string `json:"tags"`
	CategoryID       string   `json:"categoryId"`
	CategoryType     string   `json:"categoryType"`
	CategoryValue    string   `json:"categoryValue"`
	CategoryImageURL string   `json:"posterImageUrl"`
}

type PatchLiveSettingRequest struct {
	Title    *string   `json:"defaultLiveTitle,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
	Category *Category `json:"category,omitempty"`
}

// Get retrieves a list of live streams sorted by viewer count, with optional pagination.
//   - pattern: [List]
//   - credential: [Chzzk.WithClientAuth]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/live#undefined
//
// [List]: https://google.aip.dev/132
func (s *LiveService) Get(ctx context.Context, size *int, next *string) ([]Live, string, error) {
	var defaultSize = 20
	if size != nil {
		if *size > 20 {
			return nil, "", errors.New("size cannot be greater than 20")
		}
		defaultSize = *size
	}
	return s.get(ctx, defaultSize, next)
}

// Key retrieves the live streaming key for the authenticated user.
// Returns empty string if the user is not currently broadcasting.
//   - pattern: [Get]
//   - credential: [Chzzk.WithAPIKey]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/live#undefined-1
//
// [Get]: https://google.aip.dev/131
func (s *LiveService) Key(ctx context.Context) (string, error) {
	return s.key(ctx)
}

// Setting retrieves the default live streaming settings for the authenticated user.
//   - pattern: [Get]
//   - credential: [Chzzk.WithAPIKey]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/live#undefined-2
//
// [Get]: https://google.aip.dev/131
func (s *LiveService) Setting(ctx context.Context) (*LiveSetting, error) {
	return s.setting(ctx)
}

// PatchSetting updates the live streaming settings for the authenticated user.
//   - pattern: [Update]
//   - credential: [Chzzk.WithAPIKey]
//
// Example — set category (categoryType is required when categoryId is non-empty):
//
//	c.Live.PatchSetting(ctx, &chzzk.PatchLiveSettingRequest{
//	    Category: &chzzk.Category{Type: "GAME", ID: "League_of_Legends"},
//	})
//
// Example — clear category:
//
//	c.Live.PatchSetting(ctx, &chzzk.PatchLiveSettingRequest{
//	    Category: &chzzk.Category{ID: ""},
//	})
//
// Example — set title and tags:
//
//	title := "My Stream"
//	c.Live.PatchSetting(ctx, &chzzk.PatchLiveSettingRequest{
//	    Title: &title,
//	    Tags:  []string{"game", "lol"},
//	})
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/live#undefined-3
//
// [Update]: https://google.aip.dev/134
func (s *LiveService) PatchSetting(ctx context.Context, p *PatchLiveSettingRequest) error {
	return s.patchSetting(ctx, p)
}

func (s *LiveService) get(ctx context.Context, size int, next *string) ([]Live, string, error) {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixLive)
	if err != nil {
		return nil, "", err
	}
	URL, err := url.Parse(u)
	if err != nil {
		return nil, "", err
	}
	q := URL.Query()
	q.Set("size", fmt.Sprintf("%d", size))
	if next != nil {
		q.Set("next", *next)
	}
	URL.RawQuery = q.Encode()

	type livePage struct {
		Next string `json:"next"`
	}

	type LiveResp struct {
		Response
		Content struct {
			Data []Live   `json:"data"`
			Page livePage `json:"page"`
		} `json:"content"`
	}
	resp, err := rest.Get[LiveResp](ctx, s.chzzk.c, URL.String())
	if err != nil {
		return nil, "", err
	}
	if err := MightError(resp.Response); err != nil {
		return nil, "", err
	}

	return resp.Content.Data, resp.Content.Page.Next, nil
}

func (s *LiveService) key(ctx context.Context) (string, error) {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixLive, "key")
	if err != nil {
		return "", fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	type KeyResp struct {
		Response
		Content struct {
			Key string `json:"key"`
		} `json:"content"`
	}
	resp, err := rest.Get[KeyResp](ctx, s.chzzk.c, u)
	if err != nil {
		return "", err
	}
	if err := MightError(resp.Response); err != nil {
		return "", err
	}
	return resp.Content.Key, nil
}

func (s *LiveService) setting(ctx context.Context) (*LiveSetting, error) {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixLive, "setting")
	if err != nil {
		return nil, fmt.Errorf("chzzk: failed to build URL: %w", err)
	}
	type SettingResp struct {
		Response
		Content struct {
			Title    string   `json:"defaultLiveTitle"`
			Tags     []string `json:"tags"`
			Category struct {
				CategoryID       string `json:"categoryId"`
				CategoryType     string `json:"categoryType"`
				CategoryValue    string `json:"categoryValue"`
				CategoryImageURL string `json:"posterImageUrl"`
			} `json:"category"`
		} `json:"content"`
	}
	resp, err := rest.Get[SettingResp](ctx, s.chzzk.c, u)
	if err != nil {
		return nil, err
	}
	if err := MightError(resp.Response); err != nil {
		return nil, err
	}
	setting := &LiveSetting{
		Title:            resp.Content.Title,
		Tags:             resp.Content.Tags,
		CategoryID:       resp.Content.Category.CategoryID,
		CategoryType:     resp.Content.Category.CategoryType,
		CategoryValue:    resp.Content.Category.CategoryValue,
		CategoryImageURL: resp.Content.Category.CategoryImageURL,
	}
	return setting, nil
}

func (s *LiveService) patchSetting(ctx context.Context, p *PatchLiveSettingRequest) error {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixLive, "setting")
	if err != nil {
		return fmt.Errorf("chzzk: failed to build URL: %w", err)
	}

	type patchSettingResp struct {
		Response
	}

	type patchSettingReq struct {
		Title        *string  `json:"defaultLiveTitle,omitempty"`
		Tags         []string `json:"tags,omitempty"`
		CategoryType *string  `json:"categoryType,omitempty"`
		CategoryID   *string  `json:"categoryId,omitempty"`
	}

	req := &patchSettingReq{
		Title: p.Title,
		Tags:  p.Tags,
	}
	if p.Category != nil {
		req.CategoryID = &p.Category.ID
		if p.Category.ID != "" {
			req.CategoryType = &p.Category.Type
		}
	}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("chzzk: failed to marshal request body: %w", err)
	}

	resp, err := rest.Patch[patchSettingResp](ctx, s.chzzk.c, u, body)
	if err != nil {
		return fmt.Errorf("chzzk: failed to patch live setting: %w", err)
	}
	if err := MightError(resp.Response); err != nil {
		return err
	}
	return nil
}
