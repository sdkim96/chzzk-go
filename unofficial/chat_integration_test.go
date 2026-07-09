//go:build integration

package unofficial

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/sdkim96/chzzk-go"
)

func integrationClient(t *testing.T) *UnofficialChzzk {
	t.Helper()
	chz := chzzk.New(nil)
	uc, err := New(chz, nil)
	if err != nil {
		t.Fatal(err)
	}

	nidAut := os.Getenv("NID_AUT")
	nidSes := os.Getenv("NID_SES")
	if nidAut != "" && nidSes != "" {
		uc, err = uc.WithCookie(context.Background(), nidAut, nidSes)
		if err != nil {
			t.Fatalf("WithCookie failed: %v", err)
		}
	}
	return uc
}

func integrationChannelID(t *testing.T) string {
	t.Helper()
	ch := os.Getenv("CHZZK_CHANNEL_ID")
	if ch == "" {
		t.Skip("CHZZK_CHANNEL_ID not set")
	}
	return ch
}

func TestIntegration_LiveID(t *testing.T) {
	uc := integrationClient(t)
	channelID := integrationChannelID(t)

	liveID, err := uc.Live.ID(context.Background(), channelID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("liveID: %s", liveID)
}

func TestIntegration_LiveID_InvalidChannel(t *testing.T) {
	uc := integrationClient(t)

	_, err := uc.Live.ID(context.Background(), "invalid_channel_id_that_does_not_exist")
	if err == nil {
		t.Fatal("expected error for invalid channel ID, got nil")
	}
	t.Logf("expected error: %v", err)
}

func TestIntegration_ChatToken(t *testing.T) {
	uc := integrationClient(t)
	channelID := integrationChannelID(t)

	liveID, err := uc.Live.ID(context.Background(), channelID)
	if err != nil {
		t.Fatal(err)
	}

	token, err := uc.Chat.Token(context.Background(), liveID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("access: %s, extra: %s", token.AccessToken, token.ExtraToken)
}

func TestIntegration_ChatReadOnlyConnect(t *testing.T) {
	uc := integrationClient(t)
	channelID := integrationChannelID(t)

	liveID, err := uc.Live.ID(context.Background(), channelID)
	if err != nil {
		t.Fatal(err)
	}

	token, err := uc.Chat.Token(context.Background(), liveID)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	recv, err := uc.Chat.ReadOnlyConnect(ctx, liveID, token)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case msg, ok := <-recv:
		if !ok {
			t.Log("recv closed")
			return
		}
		t.Logf("received: %s", string(msg))
	case <-ctx.Done():
		t.Log("timeout (no messages received, may be normal for quiet channels)")
	}
}

func TestIntegration_ChatReadOnlyConnect_InvalidToken(t *testing.T) {
	uc := integrationClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	recv, err := uc.Chat.ReadOnlyConnect(ctx, "AAAAAA", &ChatToken{AccessToken: "invalid"})
	if err != nil {
		t.Logf("expected error at connect: %v", err)
		return
	}

	select {
	case _, ok := <-recv:
		if !ok {
			t.Log("recv closed immediately as expected")
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for connection to fail")
	}
}

func TestIntegration_ChatSendMessage(t *testing.T) {
	nidAut := os.Getenv("NID_AUT")
	nidSes := os.Getenv("NID_SES")
	if nidAut == "" || nidSes == "" {
		t.Skip("NID_AUT or NID_SES not set")
	}

	uc := integrationClient(t)
	channelID := integrationChannelID(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	liveID, err := uc.Live.ID(ctx, channelID)
	if err != nil {
		t.Fatalf("Live.ID failed: %v", err)
	}
	t.Logf("liveID: %s", liveID)

	token, err := uc.Chat.Token(ctx, liveID)
	if err != nil {
		t.Fatalf("Token failed: %v", err)
	}

	recv, send, sid, err := uc.Chat.Connect(ctx, liveID, token)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	t.Logf("sid: %s", sid)

	type sendChatBody struct {
		Msg         string `json:"msg"`
		MsgTypeCode int    `json:"msgTypeCode"`
		MsgTime     int64  `json:"msgTime"`
		Extras      string `json:"extras"`
	}
	type sendChatRequest struct {
		Bdy   sendChatBody `json:"bdy"`
		Cmd   int          `json:"cmd"`
		Sid   string       `json:"sid"`
		Cid   string       `json:"cid"`
		Svcid string       `json:"svcid"`
		Tid   int          `json:"tid"`
		Retry bool         `json:"retry"`
		Ver   string       `json:"ver"`
	}

	extras, _ := json.Marshal(map[string]any{
		"chatType":           "STREAMING",
		"emojis":             map[string]string{},
		"osType":             "PC",
		"streamingChannelId": channelID,
	})

	msg := sendChatRequest{
		Cmd:   3101,
		Sid:   sid,
		Cid:   liveID,
		Svcid: "game",
		Tid:   3,
		Retry: false,
		Ver:   "2",
		Bdy: sendChatBody{
			Msg:         "ㅋㅋㅋㅋㅋㅋㅋㅋ!",
			MsgTypeCode: 1,
			MsgTime:     time.Now().UnixMilli(),
			Extras:      string(extras),
		},
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	select {
	case send <- msgBytes:
		t.Log("message sent: ㅋㅋㅋㅋㅋㅋㅋㅋ!")
	case <-ctx.Done():
		t.Fatal("timed out sending message")
	}

	select {
	case data, ok := <-recv:
		if !ok {
			t.Fatal("recv closed unexpectedly")
		}
		t.Logf("received: %s", string(data))
	case <-time.After(5 * time.Second):
		t.Log("no echo received within 5s (may be normal)")
	}
}
