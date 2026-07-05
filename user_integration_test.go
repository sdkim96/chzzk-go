// go:build integration

package chzzk

import (
	"context"
	"os"
	"testing"
)

func Test_User401(t *testing.T) {
	chzzk := New(nil)
	_, err := chzzk.User.Me(context.Background())
	if err == nil {
		t.Errorf("Expected error for unauthorized request, got nil")
	}
}

func Test_User(t *testing.T) {
	chz := New(nil).WithAPIKey(os.Getenv("CHZZK_API_KEY"))
	user, err := chz.User.Me(context.Background())
	if err != nil {
		t.Errorf("Expected no error for authorized request, got %v", err)
	}
	if user.ChannelID == "" || user.ChannelName == "" {
		t.Errorf("Expected valid user data, got %+v", user)
	}
	t.Logf("User: %+v", user)
}
