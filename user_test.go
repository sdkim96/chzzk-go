package chzzk

import (
	"context"
	"testing"
)

func Test_User401(t *testing.T) {

	chzzk := NewChzzk(nil)
	_, err := chzzk.User.Me(context.Background())
	if err == nil {
		t.Errorf("Expected error for unauthorized request, got nil")
	}

}
