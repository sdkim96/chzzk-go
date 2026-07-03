package chzzk

import (
	"context"
	"os"
	"testing"
)

func Test_AuthClient(t *testing.T) {
	clientID := os.Getenv("CHZZK_CLIENT_ID")
	clientSecret := os.Getenv("CHZZK_SECRET_KEY")
	chzzk := NewChzzk(nil).WithClientAuth(clientID, clientSecret)

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

func Test_AuthUser(t *testing.T) {
	chzzk := NewChzzk(nil).WithAPIKey("your-api-key")

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
