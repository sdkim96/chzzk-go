//go:build integration

package chzzk

import (
	"context"
	"os"
	"testing"
)

func Test_AuthClient1(t *testing.T) {
	clientID := os.Getenv("CHZZK_CLIENT_ID")
	clientSecret := os.Getenv("CHZZK_CLIENT_SECRET")
	chzzk := New(nil).WithClientAuth(clientID, clientSecret)

	ctx := context.TODO()
	url, err := chzzk.Session.AuthClient(ctx)
	if err != nil {
		t.Fatalf("AuthClient failed: %v", err)
	}
	if url == "" {
		t.Fatal("AuthClient returned empty URL")
	}
	t.Logf("AuthClient URL: %s", url)
}

func Test_AuthUser1(t *testing.T) {
	chzzk := New(nil).WithAPIKey("_")

	ctx := context.Background()
	url, err := chzzk.Session.AuthUser(ctx)
	if err != nil {
		t.Fatalf("AuthUser failed: %v", err)
	}
	if url == "" {
		t.Fatal("AuthUser returned empty URL")
	}
	t.Logf("AuthUser URL: %s", url)
}
