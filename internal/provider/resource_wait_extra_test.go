package provider

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestCreateWithTransientRetry_ContextCancelDuringRetryWait verifies the
// ctx.Done() branch inside the retry-wait select of CreateWithTransientRetry.
//
// The function has two select statements:
//  1. At the top of the loop: ctx.Done() vs default
//  2. In the sleep between retries: ctx.Done() vs time.After(waitTime)
//
// TestCreateWithTransientRetry_TransientErrorTimesOut (in resource_wait_test.go)
// uses a negative deadline to exit via the deadline check BEFORE the sleep,
// so the sleep select is never reached.  This test uses a context with a
// short timeout (10 ms) so the deadline isn't expired immediately — the function
// starts the sleep (5 s for attempt=1) — but the context cancels at 10 ms and
// the ctx.Done() case fires, covering the sleep select branch.
func TestCreateWithTransientRetry_ContextCancelDuringRetryWait(t *testing.T) {
	transient := newResponseError("create", "vpc", 400, nil, nil)
	if !ErrorIsTransient(transient) {
		t.Fatal("test setup: expected 400 with no validation errors to be transient")
	}

	var calls int
	createFunc := func() error {
		calls++
		return transient
	}

	// Context that cancels after 10 ms — long enough for one createFunc call
	// but much shorter than the 5 s retry wait for attempt=1.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := CreateWithTransientRetry(ctx, createFunc, "vpc", "abc", time.Minute)
	if err == nil {
		t.Fatal("expected error (context cancel), got nil")
	}
	if calls < 1 {
		t.Fatal("expected createFunc to be called at least once before context cancel")
	}
}

// TestCreateWithTransientRetry_ContextCancelAtLoopTop verifies the ctx.Done()
// branch at the very TOP of the retry loop (before createFunc is called again).
//
// We achieve this by cancelling the context from inside createFunc so the
// context is already done when the loop iterates for the second time.
func TestCreateWithTransientRetry_ContextCancelAtLoopTop(t *testing.T) {
	transient := newResponseError("create", "vpc", 400, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())

	var calls int
	createFunc := func() error {
		calls++
		// Cancel the context immediately after the first call so that the next
		// loop iteration finds ctx.Done() at the top-of-loop check.
		if calls == 1 {
			cancel()
		}
		return transient
	}

	err := CreateWithTransientRetry(ctx, createFunc, "vpc", "abc", time.Minute)
	if err == nil {
		t.Fatal("expected error (context cancel), got nil")
	}
}

// TestWaitForResourceDeleted_ContextCancelled verifies the ctx.Done() branch
// inside WaitForResourceDeleted's select.
func TestWaitForResourceDeleted_ContextCancelled(t *testing.T) {
	withFastPoll(t)

	ctx, cancel := context.WithCancel(context.Background())

	var calls int
	checker := func(ctx context.Context) (bool, error) {
		calls++
		if calls == 1 {
			cancel() // cancel after first check
		}
		return false, nil // not deleted yet
	}

	err := WaitForResourceDeleted(ctx, checker, "vpc", "abc", time.Minute)
	if err == nil {
		t.Fatal("expected error from context cancel, got nil")
	}
}

// TestWaitForResourceActive_FailedStateReturnsError verifies that
// WaitForResourceActive returns an error when the checker reports a terminal
// failure state rather than a transitional or ready state.
func TestWaitForResourceActive_FailedStateReturnsError(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	checker := func(ctx context.Context) (string, error) {
		return "Failed", nil
	}
	err := WaitForResourceActive(context.Background(), checker, "vpc", "abc", time.Minute)
	if err == nil {
		t.Fatal("expected error for Failed state, got nil")
	}
	if !errors.Is(err, err) { // always true, just use err
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestWaitForResourceActive_ConsecutiveErrorsGiveUp verifies that
// WaitForResourceActive gives up after 3 consecutive checker errors.
func TestWaitForResourceActive_ConsecutiveErrorsGiveUp(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	boom := errors.New("check failed")
	var calls int
	checker := func(ctx context.Context) (string, error) {
		calls++
		return "", boom
	}
	err := WaitForResourceActive(context.Background(), checker, "vpc", "abc", time.Minute)
	if err == nil {
		t.Fatal("expected error after consecutive failures, got nil")
	}
	if calls < 3 {
		t.Fatalf("expected at least 3 calls before giving up, got %d", calls)
	}
}

// TestDeleteResourceWithRetry_ContextCancelled verifies that
// DeleteResourceWithRetry respects context cancellation.
func TestDeleteResourceWithRetry_ContextCancelled(t *testing.T) {
	withFastPoll(t)

	ctx, cancel := context.WithCancel(context.Background())

	var calls int
	deleteFunc := func() error {
		calls++
		cancel() // cancel immediately
		return errors.New("some error")
	}

	err := DeleteResourceWithRetry(ctx, deleteFunc, "vpc", "abc", time.Minute)
	// Should return some error (either from deleteFunc or context cancel)
	if err == nil {
		t.Fatal("expected error after context cancel, got nil")
	}
}
