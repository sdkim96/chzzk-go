package chzzk

import (
	"net/http"
	"testing"
)

func Test_NewChzzk(t *testing.T) {

	chzzk := NewChzzk(nil)
	if chzzk.c == nil {
		t.Errorf("NewChzzk(nil) returned a nil http.Client")
	}
	if chzzk.User == nil {
		t.Errorf("NewChzzk(nil) returned a nil UserService")
	}

	chzzk2 := NewChzzk(http.DefaultClient)
	if chzzk2.c == nil {
		t.Errorf("NewChzzk(http.DefaultClient) returned a nil http.Client")
	}
	if chzzk2.User == nil {
		t.Errorf("NewChzzk(http.DefaultClient) returned a nil UserService")
	}

}

func Test_initialize(t *testing.T) {
	chzzk := &Chzzk{}
	chzzk.initialize()
	if chzzk.User == nil {
		t.Errorf("initialize() did not initialize UserService")
	}
}

func Test_copy(t *testing.T) {
	chzzk := NewChzzk(nil)
	chzzk2 := chzzk.copy()
	if chzzk2 == nil {
		t.Errorf("copy() returned a nil Chzzk")
	}
	if chzzk2.User == nil {
		t.Errorf("copy() did not initialize UserService")
	}

	if chzzk2.c.Transport != chzzk.c.Transport {
		t.Errorf("copy() did not copy the Transport correctly")
	}
}

func Test_WithAPIKey(t *testing.T) {
	chzzk := NewChzzk(nil)
	apiKey := "test-api-key"
	chzzk2 := chzzk.WithAPIKey(apiKey)

	if chzzk2 == nil {
		t.Errorf("WithAPIKey() returned a nil Chzzk")
	}
	if chzzk2.User == nil {
		t.Errorf("WithAPIKey() did not initialize UserService")
	}

	// Check if the Transport is set correctly
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, err := chzzk2.c.Transport.RoundTrip(req)
	if err != nil {
		t.Errorf("WithAPIKey() Transport RoundTrip error: %v", err)
	}
	if resp.Request.Header.Get("Authorization") != "Bearer "+apiKey {
		t.Errorf("WithAPIKey() did not set the Authorization header correctly")
	}
}

func Test_WithHooks(t *testing.T) {
	chzzk := NewChzzk(nil)
	beforeHookCalled := false
	afterHookCalled := false

	beforeHook := func(req *http.Request) {
		beforeHookCalled = true
	}

	afterHook := func(resp *http.Response) {
		afterHookCalled = true
	}

	chzzk2 := chzzk.WithHooks(beforeHook, afterHook)

	if chzzk2 == nil {
		t.Errorf("WithHooks() returned a nil Chzzk")
	}
	if chzzk2.User == nil {
		t.Errorf("WithHooks() did not initialize UserService")
	}

	// Check if the hooks are called correctly
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	chzzk2.c.Transport.RoundTrip(req)

	if !beforeHookCalled {
		t.Errorf("WithHooks() did not call the before hook")
	}
	if !afterHookCalled {
		t.Errorf("WithHooks() did not call the after hook")
	}
}

func Test_WithHooks_NilHooks(t *testing.T) {
	chzzk := NewChzzk(nil)

	chzzk2 := chzzk.WithHooks(nil, nil)

	if chzzk2 == nil {
		t.Errorf("WithHooks() returned a nil Chzzk")
	}
	if chzzk2.User == nil {
		t.Errorf("WithHooks() did not initialize UserService")
	}

	// Check if the Transport is still functional
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	_, err := chzzk2.c.Transport.RoundTrip(req)
	if err != nil {
		t.Errorf("WithHooks() Transport RoundTrip error with nil hooks: %v", err)
	}
}

func Test_WithAPIKey_And_WithHooks(t *testing.T) {
	chzzk := NewChzzk(nil)
	apiKey := "test-api-key"
	beforeHookCalled := false
	beforeHook := func(req *http.Request) {
		beforeHookCalled = true
	}

	chzzk2 := chzzk.WithAPIKey(apiKey).WithHooks(beforeHook, nil)

	if chzzk2 == nil {
		t.Errorf("WithAPIKey().WithHooks() returned a nil Chzzk")
	}
	if chzzk2.User == nil {
		t.Errorf("WithAPIKey().WithHooks() did not initialize UserService")
	}

	// Check if the Transport is set correctly and hooks are called
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, err := chzzk2.c.Transport.RoundTrip(req)
	if err != nil {
		t.Errorf("WithAPIKey().WithHooks() Transport RoundTrip error: %v", err)
	}
	if resp.Request.Header.Get("Authorization") != "Bearer "+apiKey {
		t.Errorf("WithAPIKey().WithHooks() did not set the Authorization header correctly")
	}
	if !beforeHookCalled {
		t.Errorf("WithAPIKey().WithHooks() did not call the before hook")
	}
}
