package login

import (
	"context"
	"os"
	"testing"
)

// You must test with the verbose option: go test -v
// Copy the Authorization URL printed on the console and paste it into your browser.
func Test_Authorize(t *testing.T) {
	cid := os.Getenv("CHZZK_CLIENT_ID")

	if cid == "" || RedirectURI == "" {
		t.Skip("Skipping test due to missing environment variables")
	}
	url := URL(cid, RedirectURI, "test-state")
	t.Logf("Authorization URL: %s", url)
}

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
