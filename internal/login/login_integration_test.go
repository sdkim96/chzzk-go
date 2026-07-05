//go:build integration

package login

import (
	"context"
	"testing"
)

func Test_Server(t *testing.T) {
	ch := make(chan string, 1)
	go func() {
		err := Server(context.Background(), ch)
		if err != nil {
			t.Fatalf("Server failed to start: %v", err)
		}
	}()
	code := <-ch
	if code == "" {
		t.Fatalf("Expected authorization code, got empty string")
	}
	t.Logf("Received authorization code: %s", code)
}
