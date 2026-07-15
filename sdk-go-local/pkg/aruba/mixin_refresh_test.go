package aruba

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

// --------------------------------------------------------------------------
// WaitUntilGone
// --------------------------------------------------------------------------

func TestWaitUntilGone_RefreshNil_Error(t *testing.T) {
	var m refreshMixin
	err := m.WaitUntilGone(context.Background(), fastOpts()...)
	if err == nil {
		t.Fatal("expected error when refresh is nil")
	}
	if !strings.Contains(err.Error(), "refresh callback not set") {
		t.Errorf("error message = %q; want 'refresh callback not set'", err.Error())
	}
}

func TestWaitUntilGone_HappyPath(t *testing.T) {
	var m refreshMixin
	calls := 0
	m.setRefresh(func(_ context.Context) error {
		calls++
		if calls >= 3 {
			return &HTTPError{StatusCode: http.StatusNotFound}
		}
		return nil // still exists
	})
	if err := m.WaitUntilGone(context.Background(), fastOpts()...); err != nil {
		t.Fatalf("WaitUntilGone error: %v", err)
	}
	if calls < 3 {
		t.Errorf("refresh called %d times, want >= 3", calls)
	}
}

func TestWaitUntilGone_TransientErrorRetried(t *testing.T) {
	var m refreshMixin
	calls := 0
	m.setRefresh(func(_ context.Context) error {
		calls++
		if calls < 3 {
			return errors.New("temporary network error")
		}
		return &HTTPError{StatusCode: http.StatusNotFound}
	})
	if err := m.WaitUntilGone(context.Background(), fastOpts()...); err != nil {
		t.Fatalf("WaitUntilGone error: %v", err)
	}
	if calls < 3 {
		t.Errorf("refresh called %d times, want >= 3", calls)
	}
}

func TestWaitUntilGone_NonNotFoundHTTPError_Retried(t *testing.T) {
	var m refreshMixin
	m.setRefresh(func(_ context.Context) error {
		return &HTTPError{StatusCode: http.StatusInternalServerError}
	})
	err := m.WaitUntilGone(context.Background(), fastOpts()...)
	if err == nil {
		t.Fatal("expected non-nil error when 500 exhausts retries")
	}
}

func TestWaitUntilGone_StillExists(t *testing.T) {
	var m refreshMixin
	m.setRefresh(func(_ context.Context) error {
		return nil // resource never returns 404
	})
	err := m.WaitUntilGone(context.Background(), fastOpts()...)
	if err == nil {
		t.Fatal("expected non-nil error when resource never disappears")
	}
}

func TestWaitUntilGone_ContextCancellation(t *testing.T) {
	var m refreshMixin
	m.setRefresh(func(_ context.Context) error { return nil })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := m.WaitUntilGone(ctx, fastOpts()...)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
