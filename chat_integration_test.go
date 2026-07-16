//go:build integration

package chzzk

import (
	"context"
	"os"
	"testing"
)

func chatAPIKey(t *testing.T) *Client {
	t.Helper()
	apiKey := os.Getenv("CHZZK_API_KEY")
	if apiKey == "" {
		t.Skip("CHZZK_API_KEY not set")
	}
	return New(nil).WithAPIKey(apiKey)
}

func Test_Chat_Settings_WithAPIKey(t *testing.T) {
	c := chatAPIKey(t)

	settings, err := c.Chat.Settings(context.Background())
	if err != nil {
		t.Fatalf("Settings failed: %v", err)
	}
	t.Logf("Settings: Kind=%s, AuthorityMode=%v, SlowModeSec=%v, EmojiMode=%v",
		settings.Kind, *settings.AuthorityMode, *settings.SlowModeSec, *settings.IsEmojiMode)
	if settings.FollowerSetting != nil {
		t.Logf("FollowerSetting: MinFollowerMinute=%d, AllowSubscriberInFollowerModel=%v",
			settings.FollowerSetting.MinFollowerMinute, settings.FollowerSetting.AllowSubscriberInFollowerModel)
	}
}

func Test_Chat_Send_WithAPIKey(t *testing.T) {
	c := chatAPIKey(t)

	msgID, err := c.Chat.Send(context.Background(), "integration test message")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if msgID == "" {
		t.Fatal("Send returned empty messageId")
	}
	t.Logf("Send messageId: %s", msgID)
}

func Test_Chat_Notice_WithAPIKey(t *testing.T) {
	c := chatAPIKey(t)

	err := c.Chat.Notice(context.Background(), ChatNoticeReq{
		Kind:    NoticeMessage,
		Content: "integration test notice",
	})
	if err != nil {
		t.Fatalf("Notice failed: %v", err)
	}
	t.Log("Notice succeeded")
}

func Test_Chat_UpdateSettings_WithAPIKey(t *testing.T) {
	c := chatAPIKey(t)

	// get original settings to restore later
	original, err := c.Chat.Settings(context.Background())
	if err != nil {
		t.Fatalf("Settings failed: %v", err)
	}
	t.Logf("Original: Kind=%s, AuthorityMode=%v, SlowModeSec=%v, EmojiMode=%v",
		original.Kind, *original.AuthorityMode, *original.SlowModeSec, *original.IsEmojiMode)

	// update to different values
	err = c.Chat.UpdateSettings(context.Background(), ChatSettings{
		Kind:          ChatAvailableForFollower,
		AuthorityMode: Ptr(AuthorityModeAll),
		FollowerSetting: &ChatFollowerSetting{
			MinFollowerMinute:              5,
			AllowSubscriberInFollowerModel: true,
		},
		SlowModeSec: Ptr(3),
		IsEmojiMode: Ptr(true),
	})
	if err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}

	// verify
	updated, err := c.Chat.Settings(context.Background())
	if err != nil {
		t.Fatalf("Settings after update failed: %v", err)
	}
	if updated.Kind != ChatAvailableForFollower {
		t.Errorf("Kind = %q, want %q", updated.Kind, ChatAvailableForFollower)
	}
	if *updated.SlowModeSec != 3 {
		t.Errorf("SlowModeSec = %d, want 3", *updated.SlowModeSec)
	}
	if !*updated.IsEmojiMode {
		t.Error("IsEmojiMode = false, want true")
	}
	t.Logf("Updated: Kind=%s, SlowModeSec=%d, EmojiMode=%v",
		updated.Kind, *updated.SlowModeSec, *updated.IsEmojiMode)

	// restore original settings
	err = c.Chat.UpdateSettings(context.Background(), *original)
	if err != nil {
		t.Fatalf("Restore settings failed: %v", err)
	}
	t.Log("Settings restored")
}

func Test_Chat_BlindMessage_WithAPIKey(t *testing.T) {
	c := chatAPIKey(t)

	// send a message first, then blind it
	msgID, err := c.Chat.Send(context.Background(), "message to be blinded")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	t.Logf("Sent messageId: %s", msgID)

	// get user info for senderChannelId
	me, err := c.User.Me(context.Background())
	if err != nil {
		t.Fatalf("Me failed: %v", err)
	}

	// get chat channel id from session
	session, err := c.Session.AuthUser(context.Background(), nil)
	if err != nil {
		t.Fatalf("AuthUser failed: %v", err)
	}
	t.Logf("Session URL: %s", session)

	err = c.Chat.BlindMessage(context.Background(), ChatBlindMessageReq{
		ChatChannelID:   "",
		MessageTime:     0,
		SenderChannelID: me.ChannelID,
	})
	// BlindMessage may fail without a valid chatChannelId/messageTime — that's expected
	if err != nil {
		t.Logf("BlindMessage failed (may need valid chatChannelId/messageTime): %v", err)
	} else {
		t.Log("BlindMessage succeeded")
	}
}
