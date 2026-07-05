//go:build integration

package chzzk

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func Test_NewToken1(t *testing.T) {
	chz := New(nil).WithClientAuth(os.Getenv("CHZZK_CLIENT_ID"), os.Getenv("CHZZK_CLIENT_SECRET"))
	resp, err := chz.Token.NewToken(context.TODO(), TokenNewRequest{
		TokenRequest: TokenRequest{
			GrantType:    GrantTypeAuthorizationCode,
			ClientID:     os.Getenv("CHZZK_CLIENT_ID"),
			ClientSecret: os.Getenv("CHZZK_CLIENT_SECRET"),
		},
		Code:  "J5H5VSk8li5x0KWYsyoLDp0jhxc",
		State: "test-state",
	})
	if err != nil {
		t.Fatalf("NewToken failed: %v", err)
	}
	b, _ := json.Marshal(resp)
	fmt.Println("NewToken response:", string(b))
	if resp.AccessToken == "" {
		t.Fatal("NewToken returned empty access token")
	}
	t.Logf("NewToken response: %+v", resp)
}

func Test_RefreshToken(t *testing.T) {
	chz := New(nil).WithClientAuth(os.Getenv("CHZZK_CLIENT_ID"), os.Getenv("CHZZK_CLIENT_SECRET"))
	resp, err := chz.Token.RefreshToken(context.TODO(), TokenRefreshRequest{
		TokenRequest: TokenRequest{
			GrantType:    GrantTypeRefreshToken,
			ClientID:     os.Getenv("CHZZK_CLIENT_ID"),
			ClientSecret: os.Getenv("CHZZK_CLIENT_SECRET"),
		},
		RefreshToken: "-2bXLU25anOL9f0k-y0EcqJf4A1sjvyZ-ifTtW-_NEfF_S09V8dIMgCleJlBGUMp5dqqGc47eLBiz_df_aQfSg",
	})
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("RefreshToken returned empty access token")
	}
	t.Logf("RefreshToken response: %+v", resp)
}
