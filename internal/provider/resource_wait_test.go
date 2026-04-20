package provider

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// withFastPoll reduces the WaitForResourceDeleted poll interval so tests
// don't have to wait 10s per tick. It restores the original value on cleanup.
func withFastPoll(t *testing.T, d time.Duration) {
	t.Helper()
	orig := waitForDeletedPollInterval
	waitForDeletedPollInterval = d
	t.Cleanup(func() { waitForDeletedPollInterval = orig })
}

func TestWaitForResourceDeleted_SucceedsWhenCheckerReportsDeleted(t *testing.T) {
	withFastPoll(t, 5*time.Millisecond)

	var calls int32
	checker := func(ctx context.Context) (bool, error) {
		atomic.AddInt32(&calls, 1)
		return true, nil
	}

	err := WaitForResourceDeleted(context.Background(), checker, "kaas", "abc", time.Second)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got < 1 {
		t.Fatalf("expected checker to be invoked at least once, got %d", got)
	}
}

func TestWaitForResourceDeleted_PollsUntilDeleted(t *testing.T) {
	withFastPoll(t, 5*time.Millisecond)

	var calls int32
	checker := func(ctx context.Context) (bool, error) {
		n := atomic.AddInt32(&calls, 1)
		return n >= 3, nil
	}

	err := WaitForResourceDeleted(context.Background(), checker, "kaas", "abc", time.Second)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got < 3 {
		t.Fatalf("expected >=3 checker calls, got %d", got)
	}
}

func TestWaitForResourceDeleted_TimeoutReturnsErrWaitTimeout(t *testing.T) {
	withFastPoll(t, 5*time.Millisecond)

	checker := func(ctx context.Context) (bool, error) {
		return false, nil
	}

	start := time.Now()
	err := WaitForResourceDeleted(context.Background(), checker, "kaas", "abc", 50*time.Millisecond)
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
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("timeout took too long: %v", elapsed)
	}
}

func TestWaitForResourceDeleted_GivesUpAfterThreeConsecutiveErrors(t *testing.T) {
	withFastPoll(t, 5*time.Millisecond)

	boom := errors.New("transport exploded")
	var calls int32
	checker := func(ctx context.Context) (bool, error) {
		atomic.AddInt32(&calls, 1)
		return false, boom
	}

	err := WaitForResourceDeleted(context.Background(), checker, "kaas", "abc", 5*time.Second)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("expected wrapped boom error, got %v", err)
	}
	if IsWaitTimeout(err) {
		t.Fatalf("did not expect a wait-timeout error, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected exactly 3 checker calls before giving up, got %d", got)
	}
}

func TestWaitForResourceDeleted_ResetsErrorStreakAfterSuccess(t *testing.T) {
	withFastPoll(t, 5*time.Millisecond)

	var calls int32
	checker := func(ctx context.Context) (bool, error) {
		n := atomic.AddInt32(&calls, 1)
		// errors on calls 1, 2, 4, 5 — success (not deleted) on 3 — deleted on 6
		switch n {
		case 1, 2, 4, 5:
			return false, fmt.Errorf("flaky: %d", n)
		case 3:
			return false, nil
		default:
			return true, nil
		}
	}

	err := WaitForResourceDeleted(context.Background(), checker, "kaas", "abc", 5*time.Second)
	if err != nil {
		t.Fatalf("expected nil error — streak should reset after call 3, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got < 6 {
		t.Fatalf("expected >=6 checker calls, got %d", got)
	}
}

func TestWaitForResourceDeleted_ContextCancelledReturnsError(t *testing.T) {
	withFastPoll(t, 5*time.Millisecond)

	checker := func(ctx context.Context) (bool, error) {
		return false, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
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

// TestWaitForResourceDeleted_RecognisesNotFoundPattern documents the intended
// caller pattern from the PR: deletion checkers use IsNotFound(err) on the
// GET response to return (true, nil). This keeps the helper's contract in
// sync with the way resource Delete() implementations invoke it.
func TestWaitForResourceDeleted_RecognisesNotFoundPattern(t *testing.T) {
	withFastPoll(t, 5*time.Millisecond)

	notFound := newResponseError("get", "kaas", 404, nil, nil)
	if !IsNotFound(notFound) {
		t.Fatalf("test setup: expected IsNotFound to recognise 404 ProviderError")
	}

	checker := func(ctx context.Context) (bool, error) {
		// Simulates: _, err := client.Get(...); if IsNotFound(err) { return true, nil }
		if IsNotFound(notFound) {
			return true, nil
		}
		return false, notFound
	}

	if err := WaitForResourceDeleted(context.Background(), checker, "kaas", "abc", time.Second); err != nil {
		t.Fatalf("expected deletion to be detected via IsNotFound, got %v", err)
	}
}
