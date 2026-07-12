package unofficial

import (
	"context"
	"testing"
	"time"

	"github.com/sdkim96/chzzk-go"
)

func testClient(t *testing.T) *Client {
	t.Helper()
	chz := chzzk.New(nil)
	uc, err := New(chz, nil)
	if err != nil {
		t.Fatal(err)
	}
	return uc
}

func Test_ChatServerID(t *testing.T) {
	tests := []struct {
		channelID string
		wantMin   int
		wantMax   int
	}{
		{"AAAAAA", 1, 9},
		{"", 1, 9},
		{"abc123", 1, 9},
		{"zzzzzzzzzzzzzzz", 1, 9},
	}
	for _, tt := range tests {
		got := chatServerID(tt.channelID)
		if got < tt.wantMin || got > tt.wantMax {
			t.Errorf("chatServerID(%q) = %d, want [%d, %d]", tt.channelID, got, tt.wantMin, tt.wantMax)
		}
	}
}

func Test_ChatConnect_ContextCancelled(t *testing.T) {
	uc := testClient(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := uc.Chat.ReadOnlyConnect(ctx, "AAAAAA", &ChatToken{AccessToken: "token"})
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
	t.Logf("expected error: %v", err)
}

func Test_ChatConnect_EmptyLiveID(t *testing.T) {
	uc := testClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := uc.Chat.ReadOnlyConnect(ctx, "", &ChatToken{AccessToken: "some_token"})
	if err == nil {
		t.Fatal("expected error for empty liveID, got nil")
	}
	t.Logf("expected error: %v", err)
}
