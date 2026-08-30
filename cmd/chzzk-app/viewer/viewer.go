package viewer

import (
	"net/http"

	"github.com/sdkim96/chzzk-go"
	"github.com/sdkim96/chzzk-go/unofficial"
)

type Viewer struct {
	c *unofficial.Client
}

func NewViewer(ufHttpClient *http.Client, chz *chzzk.Client) (*Viewer, error) {
	if chz == nil {
		chz = chzzk.New(nil)
	}
	uc, err := unofficial.New(chz, ufHttpClient)
	if err != nil {
		return nil, err
	}
	return &Viewer{
		c: uc,
	}, nil
}
