//go:build integration

package chzzk

import (
	"context"
	"os"
	"testing"
)

func liveClientAuth(t *testing.T) *Chzzk {
	t.Helper()
	clientID := os.Getenv("CHZZK_CLIENT_ID")
	clientSecret := os.Getenv("CHZZK_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		t.Skip("CHZZK_CLIENT_ID or CHZZK_CLIENT_SECRET not set")
	}
	return New(nil).WithClientAuth(clientID, clientSecret)
}

func liveAPIKey(t *testing.T) *Chzzk {
	t.Helper()
	apiKey := os.Getenv("CHZZK_API_KEY")
	if apiKey == "" {
		t.Skip("CHZZK_API_KEY not set")
	}
	return New(nil).WithAPIKey(apiKey)
}

// --- Get (WithClientAuth) ---

func Test_LiveGet_WithClientAuth(t *testing.T) {
	c := liveClientAuth(t)

	lives, next, err := c.Live.Get(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("Live.Get failed: %v", err)
	}
	t.Logf("Lives count: %d, next: %q", len(lives), next)
	if len(lives) > 0 {
		t.Logf("  First: %+v", lives[0])
	}
}

// --- Get (WithAPIKey) ---

func Test_LiveGet_WithAPIKey(t *testing.T) {
	c := liveAPIKey(t)

	lives, next, err := c.Live.Get(context.Background(), nil, nil)
	if err == nil {
		t.Logf("Live.Get with APIKey succeeded: %d lives, next: %q", len(lives), next)
	} else {
		t.Logf("Live.Get with APIKey failed as expected: %v", err)
	}
}

// --- Get with size ---

func Test_LiveGet_WithSize(t *testing.T) {
	c := liveClientAuth(t)

	size := 5
	lives, next, err := c.Live.Get(context.Background(), &size, nil)
	if err != nil {
		t.Fatalf("Live.Get with size failed: %v", err)
	}
	if len(lives) > 5 {
		t.Errorf("Live.Get returned %d lives, want <= 5", len(lives))
	}
	t.Logf("Lives count: %d, next: %q", len(lives), next)
}

// --- Get with pagination ---

func Test_LiveGet_Pagination(t *testing.T) {
	c := liveClientAuth(t)

	size := 5
	lives1, next1, err := c.Live.Get(context.Background(), &size, nil)
	if err != nil {
		t.Fatalf("Live.Get page 1 failed: %v", err)
	}
	t.Logf("Page 1: %d lives, next: %q", len(lives1), next1)

	if next1 == "" {
		t.Skip("No next page available")
	}

	lives2, next2, err := c.Live.Get(context.Background(), &size, &next1)
	if err != nil {
		t.Fatalf("Live.Get page 2 failed: %v", err)
	}
	t.Logf("Page 2: %d lives, next: %q", len(lives2), next2)
}

// --- Get TooMany ---

func Test_LiveGet_TooMany(t *testing.T) {
	c := New(nil)

	size := 21
	_, _, err := c.Live.Get(context.Background(), &size, nil)
	if err == nil {
		t.Fatal("Live.Get should fail with size > 20")
	}
}

// --- Key (WithAPIKey) ---

func Test_LiveKey_WithAPIKey(t *testing.T) {
	c := liveAPIKey(t)

	key, err := c.Live.Key(context.Background())
	if err != nil {
		t.Logf("Live.Key failed (may require active broadcast): %v", err)
		return
	}
	t.Logf("Stream key: %s", key)
}

// --- Key (WithClientAuth) should fail ---

func Test_LiveKey_WithClientAuth(t *testing.T) {
	c := liveClientAuth(t)

	_, err := c.Live.Key(context.Background())
	if err == nil {
		t.Log("Live.Key with ClientAuth succeeded (unexpected)")
	} else {
		t.Logf("Live.Key with ClientAuth failed as expected: %v", err)
	}
}

// --- Setting (WithAPIKey) ---

func Test_LiveSetting_WithAPIKey(t *testing.T) {
	c := liveAPIKey(t)

	setting, err := c.Live.Setting(context.Background())
	if err != nil {
		t.Fatalf("Live.Setting failed: %v", err)
	}
	t.Logf("Setting: %+v", setting)
}

// --- Setting (WithClientAuth) should fail ---

func Test_LiveSetting_WithClientAuth(t *testing.T) {
	c := liveClientAuth(t)

	_, err := c.Live.Setting(context.Background())
	if err == nil {
		t.Log("Live.Setting with ClientAuth succeeded (unexpected)")
	} else {
		t.Logf("Live.Setting with ClientAuth failed as expected: %v", err)
	}
}

// --- PatchSetting (WithAPIKey) — set category ---

func Test_LivePatchSetting_SetCategory(t *testing.T) {
	c := liveAPIKey(t)

	err := c.Live.PatchSetting(context.Background(), &PatchLiveSettingRequest{
		Category: &Category{
			Type: "GAME",
			ID:   "League_of_Legends",
		},
	})
	if err != nil {
		t.Fatalf("PatchSetting set category failed: %v", err)
	}

	// verify
	setting, err := c.Live.Setting(context.Background())
	if err != nil {
		t.Fatalf("Setting failed: %v", err)
	}
	if setting.CategoryID != "League_of_Legends" {
		t.Errorf("CategoryID = %q, want %q", setting.CategoryID, "League_of_Legends")
	}
	t.Logf("Setting after set: %+v", setting)
}

// --- PatchSetting (WithAPIKey) — wrong category type ---

func Test_LivePatchSetting_WrongCategoryType(t *testing.T) {
	c := liveAPIKey(t)

	err := c.Live.PatchSetting(context.Background(), &PatchLiveSettingRequest{
		Category: &Category{
			Type: "ETC",
			ID:   "League_of_Legends",
		},
	})
	if err == nil {
		t.Fatal("PatchSetting with wrong category type should fail")
	}
	t.Logf("PatchSetting wrong type failed as expected: %v", err)
}

// --- PatchSetting (WithAPIKey) — clear category ---

func Test_LivePatchSetting_ClearCategory(t *testing.T) {
	c := liveAPIKey(t)

	err := c.Live.PatchSetting(context.Background(), &PatchLiveSettingRequest{
		Category: &Category{
			ID: "",
		},
	})
	if err != nil {
		t.Fatalf("PatchSetting clear category failed: %v", err)
	}

	// verify
	setting, err := c.Live.Setting(context.Background())
	if err != nil {
		t.Fatalf("Setting failed: %v", err)
	}
	if setting.CategoryID != "" {
		t.Errorf("CategoryID = %q, want empty", setting.CategoryID)
	}
	t.Logf("Setting after clear: %+v", setting)
}

// --- PatchSetting (WithAPIKey) — set title ---

func Test_LivePatchSetting_SetTitle(t *testing.T) {
	c := liveAPIKey(t)

	// get original
	original, err := c.Live.Setting(context.Background())
	if err != nil {
		t.Fatalf("Setting failed: %v", err)
	}

	newTitle := "chzzk-go integration test"
	err = c.Live.PatchSetting(context.Background(), &PatchLiveSettingRequest{
		Title: &newTitle,
	})
	if err != nil {
		t.Fatalf("PatchSetting set title failed: %v", err)
	}

	// verify
	setting, err := c.Live.Setting(context.Background())
	if err != nil {
		t.Fatalf("Setting failed: %v", err)
	}
	if setting.Title != newTitle {
		t.Errorf("Title = %q, want %q", setting.Title, newTitle)
	}

	// restore
	err = c.Live.PatchSetting(context.Background(), &PatchLiveSettingRequest{
		Title: &original.Title,
	})
	if err != nil {
		t.Fatalf("PatchSetting restore title failed: %v", err)
	}
	t.Logf("Title set and restored successfully")
}

// --- PatchSetting (WithClientAuth) should fail ---

func Test_LivePatchSetting_WithClientAuth(t *testing.T) {
	c := liveClientAuth(t)

	title := "test"
	err := c.Live.PatchSetting(context.Background(), &PatchLiveSettingRequest{
		Title: &title,
	})
	if err == nil {
		t.Log("PatchSetting with ClientAuth succeeded (unexpected)")
	} else {
		t.Logf("PatchSetting with ClientAuth failed as expected: %v", err)
	}
}
