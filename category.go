package chzzk

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	chzzkHttp "github.com/sdkim96/chzzk-go/transport/http"
)

// CategoryService serves an API for searching broadcast categories such as:
//   - games
//   - sports
//   - music
//   - news
//   - just chatting.
type CategoryService struct {
	c *Client
}

type Category struct {
	ID       string `json:"categoryId"`
	Type     string `json:"categoryType"`
	Value    string `json:"categoryValue"`
	ImageURL string `json:"posterImageUrl"`
}

// Search retrieves a list of categories matching the given query string.
//   - pattern: [List]
//   - credential: [Chzzk.WithClientAuth]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/category#undefined
//
// [List]: https://google.aip.dev/132
func (s *CategoryService) Search(ctx context.Context, query string, size *int) ([]Category, error) {
	var defaultSize = 20
	if size != nil {
		if *size > 50 {
			return nil, errors.New("size cannot be greater than 50")
		}
		defaultSize = *size
	}
	return s.search(ctx, query, defaultSize)
}

func (s *CategoryService) search(ctx context.Context, query string, size int) ([]Category, error) {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixCategory, "search")
	if err != nil {
		return nil, err
	}
	URL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	q := URL.Query()
	q.Set("query", query)
	q.Set("size", fmt.Sprintf("%d", size))
	URL.RawQuery = q.Encode()

	type CategoryResp struct {
		Response
		Content struct {
			Data []Category `json:"data"`
		} `json:"content"`
	}
	resp, err := chzzkHttp.Get[CategoryResp](ctx, s.c.httpClient, URL.String())
	if err != nil {
		return nil, err
	}
	if err := MightError(resp.Response); err != nil {
		return nil, err
	}
	var categories []Category
	if resp != nil {
		categories = resp.Content.Data
	}
	return categories, nil
}
