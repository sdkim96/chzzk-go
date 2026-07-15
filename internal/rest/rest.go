package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func Get[T any](ctx context.Context, c *http.Client, url string) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return do[T](c, req)
}

func Post[T any](ctx context.Context, c *http.Client, url string, body []byte) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return do[T](c, req)
}

func Patch[T any](ctx context.Context, c *http.Client, url string, body []byte) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return do[T](c, req)
}

func Put[T any](ctx context.Context, c *http.Client, url string, body []byte) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return do[T](c, req)
}

func do[T any](c *http.Client, req *http.Request) (*T, error) {
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
