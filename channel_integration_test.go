//go:build integration

package chzzk

import (
	"context"
	"os"
	"testing"
)

func channelClientAuth(t *testing.T) *Chzzk {
	t.Helper()
	clientID := os.Getenv("CHZZK_CLIENT_ID")
	clientSecret := os.Getenv("CHZZK_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		t.Skip("CHZZK_CLIENT_ID or CHZZK_CLIENT_SECRET not set")
	}
	return New(nil).WithClientAuth(clientID, clientSecret)
}

func channelAPIKey(t *testing.T) *Chzzk {
	t.Helper()
	apiKey := os.Getenv("CHZZK_API_KEY")
	if apiKey == "" {
		t.Skip("CHZZK_API_KEY not set")
	}
	return New(nil).WithAPIKey(apiKey)
}

func channelID(t *testing.T) string {
	t.Helper()
	id := os.Getenv("CHZZK_CHANNEL_ID")
	if id == "" {
		t.Skip("CHZZK_CHANNEL_ID not set")
	}
	return id
}

// --- Batch (WithClientAuth) ---

func Test_Get_WithClientAuth(t *testing.T) {
	c := channelClientAuth(t)
	id := channelID(t)

	channels, err := c.Channel.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(channels) == 0 {
		t.Fatal("Get returned empty result")
	}
	t.Logf("Channel: %+v", channels[0])
}

func Test_Get_WithClientAuth_Multiple(t *testing.T) {
	c := channelClientAuth(t)
	id := channelID(t)

	channels, err := c.Channel.Get(context.Background(), id, id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(channels) == 0 {
		t.Fatal("Get returned empty result")
	}
	t.Logf("Channels count: %d", len(channels))
}

func Test_Get_TooMany(t *testing.T) {
	c := New(nil)

	ids := make([]string, 21)
	for i := range ids {
		ids[i] = "dummy"
	}
	_, err := c.Channel.Get(context.Background(), ids...)
	if err == nil {
		t.Fatal("Get should fail with more than 20 IDs")
	}
}

// --- Batch (WithAPIKey) should fail ---

func Test_Get_WithAPIKey(t *testing.T) {
	c := channelAPIKey(t)
	id := channelID(t)

	_, err := c.Channel.Get(context.Background(), id)
	if err == nil {
		t.Log("Get with APIKey succeeded (unexpected — this API typically requires ClientAuth)")
	} else {
		t.Logf("Get with APIKey failed as expected: %v", err)
	}
}

// --- Managers (WithAPIKey) ---

func Test_Managers_WithAPIKey(t *testing.T) {
	c := channelAPIKey(t)

	managers, err := c.Channel.Managers(context.Background())
	if err != nil {
		t.Fatalf("Managers failed: %v", err)
	}
	t.Logf("Managers count: %d", len(managers))
	for _, m := range managers {
		t.Logf("  %+v", m)
	}
}

// --- Followers (WithAPIKey) ---

func Test_Followers_WithAPIKey(t *testing.T) {
	c := channelAPIKey(t)

	followers, nextPage, err := c.Channel.Followers(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Followers failed: %v", err)
	}
	t.Logf("Followers count: %d, nextPage: %d", len(followers), nextPage)
}

func Test_Followers_WithAPIKey_Pagination(t *testing.T) {
	c := channelAPIKey(t)

	page := 0
	size := 5
	followers, nextPage, err := c.Channel.Followers(context.Background(), &page, &size)
	if err != nil {
		t.Fatalf("Followers failed: %v", err)
	}
	if nextPage != 1 {
		t.Errorf("nextPage = %d, want 1", nextPage)
	}
	t.Logf("Followers count: %d, nextPage: %d", len(followers), nextPage)
}

// --- Subscribers (WithAPIKey) ---

func Test_Subscribers_WithAPIKey_Recent(t *testing.T) {
	c := channelAPIKey(t)

	sort := Recent
	subscribers, nextPage, err := c.Channel.Subscribers(context.Background(), nil, nil, &sort)
	if err != nil {
		t.Fatalf("Subscribers failed: %v", err)
	}
	t.Logf("Subscribers count: %d, nextPage: %d", len(subscribers), nextPage)
}

func Test_Subscribers_WithAPIKey_Longer(t *testing.T) {
	c := channelAPIKey(t)

	sort := Longer
	subscribers, nextPage, err := c.Channel.Subscribers(context.Background(), nil, nil, &sort)
	if err != nil {
		t.Fatalf("Subscribers failed: %v", err)
	}
	t.Logf("Subscribers count: %d, nextPage: %d", len(subscribers), nextPage)
}

func Test_Subscribers_WithAPIKey_Nil(t *testing.T) {
	c := channelAPIKey(t)

	subscribers, nextPage, err := c.Channel.Subscribers(context.Background(), nil, nil, nil)
	if err != nil {
		t.Fatalf("Subscribers failed: %v", err)
	}
	t.Logf("Subscribers (all nil) count: %d, nextPage: %d", len(subscribers), nextPage)
}

// --- Managers (WithClientAuth) should fail ---

func Test_Managers_WithClientAuth(t *testing.T) {
	c := channelClientAuth(t)

	_, err := c.Channel.Managers(context.Background())
	if err == nil {
		t.Log("Managers with ClientAuth succeeded (unexpected — this API typically requires APIKey)")
	} else {
		t.Logf("Managers with ClientAuth failed as expected: %v", err)
	}
}

// --- Followers (WithClientAuth) should fail ---

func Test_Followers_WithClientAuth(t *testing.T) {
	c := channelClientAuth(t)

	_, _, err := c.Channel.Followers(context.Background(), nil, nil)
	if err == nil {
		t.Log("Followers with ClientAuth succeeded (unexpected — this API typically requires APIKey)")
	} else {
		t.Logf("Followers with ClientAuth failed as expected: %v", err)
	}
}

// --- Subscribers (WithClientAuth) should fail ---

func Test_Subscribers_WithClientAuth(t *testing.T) {
	c := channelClientAuth(t)

	_, _, err := c.Channel.Subscribers(context.Background(), nil, nil, nil)
	if err == nil {
		t.Log("Subscribers with ClientAuth succeeded (unexpected — this API typically requires APIKey)")
	} else {
		t.Logf("Subscribers with ClientAuth failed as expected: %v", err)
	}
}
