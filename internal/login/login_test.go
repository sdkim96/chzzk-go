package login

import "testing"

func Test_URL(t *testing.T) {
	got := URL("my-client-id", RedirectURI, "test-state")
	want := "https://chzzk.naver.com/account-interlock?clientId=my-client-id&redirectUri=http://localhost:57777/callback&state=test-state"
	if got != want {
		t.Errorf("URL() = %q, want %q", got, want)
	}
}
