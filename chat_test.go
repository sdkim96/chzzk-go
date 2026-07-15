package chzzk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestChatService(handler http.HandlerFunc) *ChatService {
	srv := httptest.NewServer(handler)
	c := &Client{httpClient: srv.Client()}
	c.httpClient.Transport = &rewriteTransport{
		base:    c.httpClient.Transport,
		baseURL: srv.URL,
	}
	return &ChatService{c: c}
}

// rewriteTransport redirects all requests to the test server.
type rewriteTransport struct {
	base    http.RoundTripper
	baseURL string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.URL.Scheme = "http"
	req2.URL.Host = t.baseURL[len("http://"):]
	return t.base.RoundTrip(req2)
}

func Test_ChatService_Send(t *testing.T) {
	svc := newTestChatService(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body struct {
			Message string `json:"message"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		if body.Message != "hello" {
			t.Errorf("expected message 'hello', got %q", body.Message)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"code":    200,
			"message": "OK",
			"content": map[string]any{
				"messageId": "msg-123",
			},
		})
	})

	msgID, err := svc.Send(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if msgID != "msg-123" {
		t.Errorf("expected messageId 'msg-123', got %q", msgID)
	}
}

func Test_ChatService_Notice_Message(t *testing.T) {
	svc := newTestChatService(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["message"] != "notice text" {
			t.Errorf("expected message 'notice text', got %v", body["message"])
		}
		if _, ok := body["messageId"]; ok {
			t.Error("expected messageId to be absent")
		}
		json.NewEncoder(w).Encode(map[string]any{"code": 200, "message": "OK"})
	})

	err := svc.Notice(context.Background(), ChatNoticeReq{
		Kind:    NoticeMessage,
		Content: "notice text",
	})
	if err != nil {
		t.Fatalf("Notice failed: %v", err)
	}
}

func Test_ChatService_Notice_MessageID(t *testing.T) {
	svc := newTestChatService(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["messageId"] != "id-456" {
			t.Errorf("expected messageId 'id-456', got %v", body["messageId"])
		}
		json.NewEncoder(w).Encode(map[string]any{"code": 200, "message": "OK"})
	})

	err := svc.Notice(context.Background(), ChatNoticeReq{
		Kind:    NoticeMessageID,
		Content: "id-456",
	})
	if err != nil {
		t.Fatalf("Notice failed: %v", err)
	}
}

func Test_ChatService_Notice_InvalidKind(t *testing.T) {
	svc := &ChatService{c: &Client{httpClient: &http.Client{}}}
	err := svc.Notice(context.Background(), ChatNoticeReq{Kind: "invalid"})
	if err != ErrInvalidNoticeKind {
		t.Errorf("expected ErrInvalidNoticeKind, got %v", err)
	}
}

func Test_ChatService_Settings(t *testing.T) {
	svc := newTestChatService(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"code":    200,
			"message": "OK",
			"content": map[string]any{
				"chatAvailableCondition":        "REAL_NAME",
				"chatAvailableGroup":            "FOLLOWER",
				"minFollowerMinute":             10,
				"allowSubscriberInFollowerMode": true,
				"chatSlowModeSec":               5,
				"chatEmojiMode":                 true,
			},
		})
	})

	s, err := svc.Settings(context.Background())
	if err != nil {
		t.Fatalf("Settings failed: %v", err)
	}
	if s.Kind != ChatAvailableForFollower {
		t.Errorf("expected Kind 'follower', got %q", s.Kind)
	}
	if *s.AuthorityMode != AuthorityModelRealName {
		t.Errorf("expected AuthorityMode 'realName', got %q", *s.AuthorityMode)
	}
	if s.FollowerSetting.MinFollowerMinute != 10 {
		t.Errorf("expected MinFollowerMinute 10, got %d", s.FollowerSetting.MinFollowerMinute)
	}
	if !s.FollowerSetting.AllowSubscriberInFollowerModel {
		t.Error("expected AllowSubscriberInFollowerModel to be true")
	}
	if *s.SlowModeSec != 5 {
		t.Errorf("expected SlowModeSec 5, got %d", *s.SlowModeSec)
	}
	if !*s.IsEmojiMode {
		t.Error("expected IsEmojiMode to be true")
	}
}

func Test_ChatService_UpdateSettings(t *testing.T) {
	svc := newTestChatService(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		var body settings
		json.NewDecoder(r.Body).Decode(&body)
		if body.ChatAvailableGroup != "ALL" {
			t.Errorf("expected ChatAvailableGroup 'ALL', got %q", body.ChatAvailableGroup)
		}
		if body.ChatAvailableCondition != "NONE" {
			t.Errorf("expected ChatAvailableCondition 'NONE', got %q", body.ChatAvailableCondition)
		}
		if body.MinFollowerMinute != 0 {
			t.Errorf("expected MinFollowerMinute 0, got %d", body.MinFollowerMinute)
		}
		if body.ChatSlowModeSec != 3 {
			t.Errorf("expected ChatSlowModeSec 3, got %d", body.ChatSlowModeSec)
		}
		json.NewEncoder(w).Encode(map[string]any{"code": 200, "message": "OK"})
	})

	err := svc.UpdateSettings(context.Background(), ChatSettings{
		Kind:          ChatAvailableForAll,
		AuthorityMode: Ptr(AuthorityModeAll),
		FollowerSetting: &ChatFollowerSetting{
			MinFollowerMinute:              0,
			AllowSubscriberInFollowerModel: false,
		},
		SlowModeSec: Ptr(3),
		IsEmojiMode: Ptr(false),
	})
	if err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}
}

func Test_ChatService_BlindMessage(t *testing.T) {
	svc := newTestChatService(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body ChatBlindMessageReq
		json.NewDecoder(r.Body).Decode(&body)
		if body.ChatChannelID != "ch-1" {
			t.Errorf("expected chatChannelId 'ch-1', got %q", body.ChatChannelID)
		}
		if body.MessageTime != 1234567890 {
			t.Errorf("expected messageTime 1234567890, got %d", body.MessageTime)
		}
		if body.SenderChannelID != "sender-1" {
			t.Errorf("expected senderChannelId 'sender-1', got %q", body.SenderChannelID)
		}
		json.NewEncoder(w).Encode(map[string]any{"code": 200, "message": "OK"})
	})

	err := svc.BlindMessage(context.Background(), ChatBlindMessageReq{
		ChatChannelID:   "ch-1",
		MessageTime:     1234567890,
		SenderChannelID: "sender-1",
	})
	if err != nil {
		t.Fatalf("BlindMessage failed: %v", err)
	}
}

func Test_cakToCAG(t *testing.T) {
	tests := []struct {
		in  ChatAvailableKind
		out string
	}{
		{ChatAvailableForAll, "ALL"},
		{ChatAvailableForFollower, "FOLLOWER"},
		{ChatAvailableForManager, "MANAGER"},
		{ChatAvailableForSubscriber, "SUBSCRIBER"},
	}
	for _, tt := range tests {
		if got := cakToCAG(tt.in); got != tt.out {
			t.Errorf("cakToCAG(%q) = %q, want %q", tt.in, got, tt.out)
		}
	}
}

func Test_cagToCAK(t *testing.T) {
	tests := []struct {
		in  string
		out ChatAvailableKind
	}{
		{"ALL", ChatAvailableForAll},
		{"FOLLOWER", ChatAvailableForFollower},
		{"MANAGER", ChatAvailableForManager},
		{"SUBSCRIBER", ChatAvailableForSubscriber},
	}
	for _, tt := range tests {
		if got := cagToCAK(tt.in); got != tt.out {
			t.Errorf("cagToCAK(%q) = %q, want %q", tt.in, got, tt.out)
		}
	}
}

func Test_akToCAC(t *testing.T) {
	tests := []struct {
		in  AuthorityKind
		out string
	}{
		{AuthorityModeAll, "NONE"},
		{AuthorityModelRealName, "REAL_NAME"},
	}
	for _, tt := range tests {
		if got := akToCAC(tt.in); got != tt.out {
			t.Errorf("akToCAC(%q) = %q, want %q", tt.in, got, tt.out)
		}
	}
}

func Test_cacToAK(t *testing.T) {
	tests := []struct {
		in  string
		out AuthorityKind
	}{
		{"NONE", AuthorityModeAll},
		{"REAL_NAME", AuthorityModelRealName},
		{"UNKNOWN", ""},
	}
	for _, tt := range tests {
		if got := cacToAK(tt.in); got != tt.out {
			t.Errorf("cacToAK(%q) = %q, want %q", tt.in, got, tt.out)
		}
	}
}
