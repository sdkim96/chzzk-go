// package unofficial provides useful, but unofficial features for accessing to Chzzk.
// Since these features are not officially supported by the Chzzk API,
// clients should use them with caution and be aware not to be banned or blocked by the Chzzk API.
//
// If you got banned for abusing these features, chzzk-go is not responsible for that.
// Please use these features at your own risk.
package unofficial

import (
	"errors"
	"net/http"

	"github.com/sdkim96/chzzk-go"
)

// UnofficialChzzk is a client for accessing unofficial features of the Chzzk API.
//
// The embedded Chzzk client is used when the official features need to be called during execution of unofficial features.
//
// The underlying http.Clients, which are both the c and chzzk.c don't share the same configurations except for the transport.
// The transport is shared between the two clients, enabling connection reuse and keep-alive.
type UnofficialChzzk struct {
	c     *http.Client
	chzzk *chzzk.Chzzk
}

func New(c *chzzk.Chzzk) (*UnofficialChzzk, error) {
	if c == nil {
		return nil, errors.New("chzzk instance is nil")
	}
	return &UnofficialChzzk{chzzk: c}, nil
}
