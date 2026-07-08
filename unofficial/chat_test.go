package unofficial

import (
	"context"
	"testing"

	"github.com/sdkim96/chzzk-go"
)

func client() *UnofficialChzzk {
	chz := chzzk.New(nil)
	uc, err := New(chz, nil)
	if err != nil {
		panic(err)
	}
	return uc
}

func Test_LiveID(t *testing.T) {
	liveID, err := client().Chat.LiveID(context.Background(), "3497a9a7221cc3ee5d3f95991d9f95e9")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(liveID)
}

func Test_AccessToken(t *testing.T) {
	accessToken, err := client().Chat.AccessToken(context.Background(), "N2beLf")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(accessToken)
}
