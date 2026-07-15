package chzzk

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/sdkim96/chzzk-go/internal/rest"
)

// ChatService handles APIs prefixed with /chats
type ChatService struct {
	c *Client
}

type NoticeKind string
type ChatAvailableKind string
type AuthorityKind string

const (
	NoticeMessage              NoticeKind        = "message"
	NoticeMessageID            NoticeKind        = "messageId"
	ChatAvailableForAll        ChatAvailableKind = "all"
	ChatAvailableForFollower   ChatAvailableKind = "follower"
	ChatAvailableForManager    ChatAvailableKind = "manager"
	ChatAvailableForSubscriber ChatAvailableKind = "subscriber"
	AuthorityModeAll           AuthorityKind     = "all"
	AuthorityModelRealName     AuthorityKind     = "realName"
)

var ErrInvalidNoticeKind = errors.New("chzzk: invalid notice kind")

// ChatNoticeReq represents a request to Notice API.
// It can be used to set a notice message or a notice message ID.
type ChatNoticeReq struct {
	Kind    NoticeKind
	Content string
}

// ChatSettings represents the chat settings for a channel.
// It is used in the Settings and UpdateSettings APIs.
type ChatSettings struct {
	Kind            ChatAvailableKind
	FollowerSetting *ChatFollowerSetting
	AuthorityMode   *AuthorityKind
	SlowModeSec     *int
	IsEmojiMode     *bool
}

type ChatFollowerSetting struct {
	MinFollowerMinute              int
	AllowSubscriberInFollowerModel bool
}

// ChatBlindMessageReq represents a request to blind (hide) a specific chat message.
type ChatBlindMessageReq struct {
	ChatChannelID   string `json:"chatChannelId"`
	MessageTime     int64  `json:"messageTime"`
	SenderChannelID string `json:"senderChannelId"`
}

// Send sends a chat message and returns the message ID.
//   - pattern: [Create]
//   - credential: [Client.WithAPIKey]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/chat#undefined
//
// [Create]: https://google.aip.dev/133
func (s *ChatService) Send(ctx context.Context, msg string) (string, error) {
	return s.send(ctx, msg)
}

// Notice sets or updates the chat notice.
// Use [NoticeMessage] to set by message text, or [NoticeMessageID] to set by an existing message ID.
//   - pattern: [Create]
//   - credential: [Client.WithAPIKey]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/chat#undefined-1
//
// [Create]: https://google.aip.dev/133
func (s *ChatService) Notice(ctx context.Context, req ChatNoticeReq) error {
	return s.notice(ctx, req)
}

// Settings retrieves the current chat settings for the authenticated user's channel.
//   - pattern: [Get]
//   - credential: [Client.WithAPIKey]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/chat#undefined-2
//
// [Get]: https://google.aip.dev/131
func (s *ChatService) Settings(ctx context.Context) (*ChatSettings, error) {
	return s.settings(ctx)
}

// UpdateSettings updates the chat settings for the authenticated user's channel.
//   - pattern: [Update]
//   - credential: [Client.WithAPIKey]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/chat#undefined-3
//
// [Update]: https://google.aip.dev/134
func (s *ChatService) UpdateSettings(ctx context.Context, req ChatSettings) error {
	return s.updateSettings(ctx, req)
}

// BlindMessage blinds (hides) a specific chat message.
//   - pattern: [Create]
//   - credential: [Client.WithAPIKey]
//
// Check the documentation for more details: https://chzzk.gitbook.io/chzzk/chzzk-api/chat#undefined-4
//
// [Create]: https://google.aip.dev/133
func (s *ChatService) BlindMessage(ctx context.Context, req ChatBlindMessageReq) error {
	return s.blindMessage(ctx, req)
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

func (s *ChatService) notice(ctx context.Context, req ChatNoticeReq) error {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixChat, "notice")
	if err != nil {
		return err
	}
	type noticeReq struct {
		Message   *string `json:"message,omitempty"`
		MessageID *string `json:"messageId,omitempty"`
	}
	var innerReq noticeReq
	switch req.Kind {
	case NoticeMessage:
		innerReq = noticeReq{Message: Ptr(req.Content)}
	case NoticeMessageID:
		innerReq = noticeReq{MessageID: Ptr(req.Content)}
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

func (s *ChatService) settings(ctx context.Context) (*ChatSettings, error) {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixChat, "settings")
	if err != nil {
		return nil, err
	}
	type settingsResp struct {
		Response
		Content settings `json:"content"`
	}
	resp, err := rest.Get[settingsResp](ctx, s.c.httpClient, u)
	if err != nil {
		return nil, err
	}
	if err := MightError(resp.Response); err != nil {
		return nil, err
	}
	settings := &ChatSettings{
		Kind: cagToCAK(resp.Content.ChatAvailableGroup),
		FollowerSetting: &ChatFollowerSetting{
			MinFollowerMinute:              resp.Content.MinFollowerMinute,
			AllowSubscriberInFollowerModel: resp.Content.AllowSubscriberInFollowerMode,
		},
		AuthorityMode: Ptr(cacToAK(resp.Content.ChatAvailableCondition)),
		SlowModeSec:   Ptr(resp.Content.ChatSlowModeSec),
		IsEmojiMode:   Ptr(resp.Content.ChatEmojiMode),
	}
	return settings, nil
}

func (s *ChatService) updateSettings(ctx context.Context, req ChatSettings) error {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixChat, "settings")
	if err != nil {
		return err
	}
	type updateSettingsReq struct {
		settings
	}
	type updateSettingsResp struct {
		Response
	}
	settingsReq := updateSettingsReq{
		settings: settings{
			ChatAvailableCondition:        akToCAC(*req.AuthorityMode),
			ChatAvailableGroup:            cakToCAG(req.Kind),
			MinFollowerMinute:             req.FollowerSetting.MinFollowerMinute,
			AllowSubscriberInFollowerMode: req.FollowerSetting.AllowSubscriberInFollowerModel,
			ChatSlowModeSec:               *req.SlowModeSec,
			ChatEmojiMode:                 *req.IsEmojiMode,
		},
	}
	rawReq, err := json.Marshal(settingsReq)
	if err != nil {
		return err
	}
	resp, err := rest.Put[updateSettingsResp](ctx, s.c.httpClient, u, rawReq)
	if err != nil {
		return err
	}
	if err := MightError(resp.Response); err != nil {
		return err
	}
	return nil
}

func (s *ChatService) blindMessage(ctx context.Context, req ChatBlindMessageReq) error {
	u, err := url.JoinPath(BaseURL, OpenV1, prefixChat, "blind-message")
	if err != nil {
		return err
	}
	type blindMessageResp struct {
		Response
	}
	rawReq, err := json.Marshal(req)
	if err != nil {
		return err
	}
	resp, err := rest.Post[blindMessageResp](ctx, s.c.httpClient, u, rawReq)
	if err != nil {
		return err
	}
	if err := MightError(resp.Response); err != nil {
		return err
	}
	return nil
}

type settings struct {
	ChatAvailableCondition        string `json:"chatAvailableCondition"`
	ChatAvailableGroup            string `json:"chatAvailableGroup"`
	MinFollowerMinute             int    `json:"minFollowerMinute"`
	AllowSubscriberInFollowerMode bool   `json:"allowSubscriberInFollowerMode"`
	ChatSlowModeSec               int    `json:"chatSlowModeSec"`
	ChatEmojiMode                 bool   `json:"chatEmojiMode"`
}

// cakToCG means ChatAvailableKind to ChatAvailableGroup.
//
// coverting rules:
// - all -> ALL
// - follower -> FOLLOWER
// - manager -> MANAGER
// - subscriber -> SUBSCRIBER
func cakToCAG(cak ChatAvailableKind) string {
	return strings.ToUpper(string(cak))
}

func cagToCAK(cag string) ChatAvailableKind {
	return ChatAvailableKind(strings.ToLower(cag))
}

// akToCAC means AuthorityKind to chatAvailableCondition
//
// coverting rules:
func akToCAC(am AuthorityKind) string {
	switch am {
	case AuthorityModeAll:
		return "NONE"
	case AuthorityModelRealName:
		return "REAL_NAME"
	}
	return ""
}

func cacToAK(cac string) AuthorityKind {
	switch cac {
	case "NONE":
		return AuthorityModeAll
	case "REAL_NAME":
		return AuthorityModelRealName
	}
	return ""
}
