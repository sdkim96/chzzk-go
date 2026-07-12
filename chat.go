package chzzk

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/sdkim96/chzzk-go/internal/rest"
)

// ChatService provides capabilities for controlling and interacting with the chat systems,
// including sending messages, receiving messages, and managing chat sessions.
type ChatService struct {
	c *Client
}

type NoticeKind string

const (
	Message   NoticeKind = "message"
	MessageID NoticeKind = "messageId"
)

var ErrInvalidNoticeKind = errors.New("chzzk: invalid notice kind")

type NoticeReq struct {
	Kind    NoticeKind `json:"kind"`
	Content string     `json:"content"`
}

func (s *ChatService) Send(ctx context.Context, msg string) (string, error) {
	return s.send(ctx, msg)
}

func (s *ChatService) Notice(ctx context.Context, req NoticeReq) error {
	return s.notice(ctx, req)
}

func (s *ChatService) send(ctx context.Context, msg string) (string, error) {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixChat, "send")
	if err != nil {
		return "", err
	}
	type SendReq struct {
		Message string `json:"message"`
	}
	type SendResp struct {
		Response
		Content struct {
			MessageID string `json:"messageId"`
		} `json:"content"`
	}
	req := SendReq{Message: msg}
	rawReq, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := rest.Post[SendResp](ctx, s.c.httpClient, u, rawReq)
	if err != nil {
		return "", err
	}
	if err := MightError(resp.Response); err != nil {
		return "", err
	}
	return resp.Content.MessageID, nil
}

func (s *ChatService) notice(ctx context.Context, req NoticeReq) error {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixChat, "notice")
	if err != nil {
		return err
	}
	type noticeReq struct {
		Message   string `json:"message,omitempty"`
		MessageID string `json:"messageId,omitempty"`
	}
	var innerReq noticeReq
	switch req.Kind {
	case Message:
		innerReq = noticeReq{Message: req.Content}
	case MessageID:
		innerReq = noticeReq{MessageID: req.Content}
	default:
		return ErrInvalidNoticeKind
	}
	rawReq, err := json.Marshal(innerReq)
	if err != nil {
		return err
	}
	type noticeResp struct {
		Response
	}
	resp, err := rest.Post[noticeResp](ctx, s.c.httpClient, u, rawReq)
	if err != nil {
		return err
	}
	if err := MightError(resp.Response); err != nil {
		return err
	}
	return nil
}
