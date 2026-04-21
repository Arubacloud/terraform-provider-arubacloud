package provider

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// These tests exercise the real 10s poll interval of WaitForResourceDeleted.
// They run in parallel so total wall time stays close to one tick.

func TestWaitForResourceDeleted_SucceedsWhenCheckerReportsDeleted(t *testing.T) {
	t.Parallel()

	var calls int32
	checker := func(ctx context.Context) (bool, error) {
		atomic.AddInt32(&calls, 1)
		return true, nil
	}

	err := WaitForResourceDeleted(context.Background(), checker, "kaas", "abc", 30*time.Second)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got < 1 {
		t.Fatalf("expected checker to be invoked at least once, got %d", got)
	}
}

func TestWaitForResourceDeleted_TimeoutReturnsErrWaitTimeout(t *testing.T) {
	t.Parallel()

	checker := func(ctx context.Context) (bool, error) {
		return false, nil
	}

	err := WaitForResourceDeleted(context.Background(), checker, "kaas", "abc", time.Second)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !IsWaitTimeout(err) {
		t.Fatalf("expected *ErrWaitTimeout, got %T: %v", err, err)
	}
	var te *ErrWaitTimeout
	if !errors.As(err, &te) {
		t.Fatalf("errors.As failed: %v", err)
	}
	if te.ResourceType != "kaas" || te.ResourceID != "abc" {
		t.Fatalf("unexpected timeout fields: %+v", te)
	}
}

func TestWaitForResourceDeleted_ContextCancelledReturnsError(t *testing.T) {
	t.Parallel()

	checker := func(ctx context.Context) (bool, error) {
		return false, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := WaitForResourceDeleted(ctx, checker, "kaas", "abc", time.Hour)
	if err == nil {
		t.Fatal("expected context-cancelled error, got nil")
	}
	if IsWaitTimeout(err) {
		t.Fatalf("expected cancellation error, got wait-timeout: %v", err)
	}
}

// Verifies the intended caller pattern: deletion checkers use IsNotFound(err)
// on the GET response to return (true, nil).
func TestWaitForResourceDeleted_RecognisesNotFoundPattern(t *testing.T) {
	t.Parallel()

	notFound := newResponseError("get", "kaas", 404, nil, nil)
	if !IsNotFound(notFound) {
		t.Fatalf("test setup: expected IsNotFound to recognise 404 ProviderError")
	}

	checker := func(ctx context.Context) (bool, error) {
		if IsNotFound(notFound) {
			return true, nil
		}
		return false, notFound
	}

	if err := WaitForResourceDeleted(context.Background(), checker, "kaas", "abc", 30*time.Second); err != nil {
		t.Fatalf("expected deletion to be detected via IsNotFound, got %v", err)
	}
}
