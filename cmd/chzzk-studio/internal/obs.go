package internal

import "github.com/andreykaipov/goobs"

func NewOBSClient(port, password string) (*goobs.Client, error) {
	return goobs.New("127.0.0.1:"+port, goobs.WithPassword(password))
}
