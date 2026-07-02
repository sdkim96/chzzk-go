package login

import (
	"context"
	"fmt"
	"net/http"
)

const (

	// By RFC8252, The OAuth 2.0 server should allow any port to be specified at request time.
	// However, Chzzk does not follow this rule.
	// We need to use a custom hard-coded port number to receive the authorization code from the redirect URI.
	//
	// You can check futher details about RFC8252 here:
	// https://datatracker.ietf.org/doc/html/rfc8252?utm_source=chatgpt.com#section-7.3
	RedirectURI  = "http://localhost:57777/callback"
	AuthorizeURL = "https://chzzk.naver.com/account-interlock"
)

func URL(clientID, redirectURI, state string) string {
	return fmt.Sprintf("%s?clientId=%s&redirectUri=%s&state=%s", AuthorizeURL, clientID, redirectURI, state)
}

func Server(ctx context.Context, codeCh chan<- string) error {
	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:    ":57777",
		Handler: mux,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if code == "" || state == "" {
			http.Error(w, "Missing code or state", http.StatusBadRequest)
			return
		}

		fmt.Fprintln(w, "Login successful. You may close this window.")

		codeCh <- code

		go srv.Shutdown(ctx)
	})

	go func() {
		<-ctx.Done()
		go srv.Shutdown(ctx)
	}()

	srv.Handler = mux

	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
