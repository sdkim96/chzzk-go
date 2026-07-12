//go:build integration

package chzzk

import (
	"context"
	"os"
	"testing"
)

func categoryClientAuth(t *testing.T) *Client {
	t.Helper()
	clientID := os.Getenv("CHZZK_CLIENT_ID")
	clientSecret := os.Getenv("CHZZK_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		t.Skip("CHZZK_CLIENT_ID or CHZZK_CLIENT_SECRET not set")
	}
	return New(nil).WithClientAuth(clientID, clientSecret)
}

func categoryAPIKey(t *testing.T) *Client {
	t.Helper()
	apiKey := os.Getenv("CHZZK_API_KEY")
	if apiKey == "" {
		t.Skip("CHZZK_API_KEY not set")
	}
	return New(nil).WithAPIKey(apiKey)
}

// --- Search (WithClientAuth) ---

func Test_CategorySearch_WithClientAuth(t *testing.T) {
	c := categoryClientAuth(t)

	categories, err := c.Category.Search(context.Background(), "리그", nil)
	if err != nil {
		t.Fatalf("Category.Search failed: %v", err)
	}
	t.Logf("Categories count: %d", len(categories))
	for _, cat := range categories {
		t.Logf("  %+v", cat)
	}
}

// --- Search (WithAPIKey) ---

func Test_CategorySearch_WithAPIKey(t *testing.T) {
	c := categoryAPIKey(t)

	categories, err := c.Category.Search(context.Background(), "리그", nil)
	if err == nil {
		t.Logf("Category.Search with APIKey succeeded: %d categories", len(categories))
	} else {
		t.Logf("Category.Search with APIKey failed as expected: %v", err)
	}
}

// --- Search with size ---

func Test_CategorySearch_WithSize(t *testing.T) {
	c := categoryClientAuth(t)

	size := 3
	categories, err := c.Category.Search(context.Background(), "게임", &size)
	if err != nil {
		t.Fatalf("Category.Search with size failed: %v", err)
	}
	if len(categories) > 3 {
		t.Errorf("Category.Search returned %d categories, want <= 3", len(categories))
	}
	t.Logf("Categories count: %d", len(categories))
}

// --- Search TooMany ---

func Test_CategorySearch_TooMany(t *testing.T) {
	c := New(nil)

	size := 51
	_, err := c.Category.Search(context.Background(), "test", &size)
	if err == nil {
		t.Fatal("Category.Search should fail with size > 50")
	}
}

// --- Search empty query ---

func Test_CategorySearch_EmptyQuery(t *testing.T) {
	c := categoryClientAuth(t)

	categories, err := c.Category.Search(context.Background(), "", nil)
	if err != nil {
		t.Logf("Category.Search with empty query failed: %v", err)
	} else {
		t.Logf("Category.Search with empty query: %d categories", len(categories))
	}
}
