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
func withFastPoll(t *testing.T) {
	t.Helper()
	orig := waitForDeletedPollInterval
	waitForDeletedPollInterval = 5 * time.Millisecond
	t.Cleanup(func() { waitForDeletedPollInterval = orig })
}

func TestWaitForResourceDeleted_SucceedsWhenCheckerReportsDeleted(t *testing.T) {
	withFastPoll(t)

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
	withFastPoll(t)

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
	withFastPoll(t)

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
	withFastPoll(t)

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
	withFastPoll(t)

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
	withFastPoll(t)

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

// ── WaitForResourceActive ────────────────────────────────────────────────────

// withFastActivePoll sets the active-poll interval to a very short duration for
// testing and restores the original value on cleanup.
func withFastActivePoll(t *testing.T) {
	t.Helper()
	orig := waitForActivePollInterval
	waitForActivePollInterval = 5 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = orig })
}

func TestWaitForResourceActive_SucceedsWhenCheckerReturnsReadyState(t *testing.T) {
	withFastActivePoll(t)

	var calls int32
	checker := func(ctx context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "Active", nil
	}

	if err := WaitForResourceActive(context.Background(), checker, "kaas", "abc", time.Second); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got < 1 {
		t.Fatalf("expected checker called at least once, got %d", got)
	}
}

func TestWaitForResourceActive_TimeoutReturnsErrWaitTimeout(t *testing.T) {
	withFastActivePoll(t)

	checker := func(ctx context.Context) (string, error) {
		return "InCreation", nil
	}

	err := WaitForResourceActive(context.Background(), checker, "kaas", "abc", 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsWaitTimeout(err) {
		t.Fatalf("expected ErrWaitTimeout, got %T: %v", err, err)
	}
}

func TestWaitForResourceActive_ContextCancelledReturnsError(t *testing.T) {
	withFastActivePoll(t)

	checker := func(ctx context.Context) (string, error) {
		return "InCreation", nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(20 * time.Millisecond); cancel() }()

	err := WaitForResourceActive(ctx, checker, "kaas", "abc", time.Hour)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if IsWaitTimeout(err) {
		t.Fatalf("expected cancellation error, got wait-timeout: %v", err)
	}
}

func TestWaitForResourceActive_GivesUpAfterThreeConsecutiveErrors(t *testing.T) {
	withFastActivePoll(t)

	boom := errors.New("checker exploded")
	var calls int32
	checker := func(ctx context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", boom
	}

	err := WaitForResourceActive(context.Background(), checker, "kaas", "abc", 5*time.Second)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("expected wrapped boom, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected exactly 3 checker calls, got %d", got)
	}
}

func TestWaitForResourceActive_ErrorStreakResetsAfterSuccess(t *testing.T) {
	withFastActivePoll(t)

	var calls int32
	checker := func(ctx context.Context) (string, error) {
		n := atomic.AddInt32(&calls, 1)
		// errors on calls 1+2, transitional on 3, errors on 4+5, active on 6
		switch n {
		case 1, 2, 4, 5:
			return "", fmt.Errorf("flaky: %d", n)
		case 3:
			return "InCreation", nil
		default:
			return "Active", nil
		}
	}

	if err := WaitForResourceActive(context.Background(), checker, "kaas", "abc", 5*time.Second); err != nil {
		t.Fatalf("expected nil — streak should reset after call 3, got %v", err)
	}
}

func TestWaitForResourceActive_FailedStateReturnsImmediately(t *testing.T) {
	withFastActivePoll(t)

	checker := func(ctx context.Context) (string, error) {
		return "Failed", nil
	}

	err := WaitForResourceActive(context.Background(), checker, "kaas", "abc", time.Second)
	if err == nil {
		t.Fatal("expected error for failed state, got nil")
	}
	if IsWaitTimeout(err) {
		t.Fatalf("expected non-timeout error, got wait-timeout: %v", err)
	}
}

// ── DeleteResourceWithRetry ──────────────────────────────────────────────────

// withFastDeleteRetry sets the per-attempt retry base wait to zero so tests
// don't sleep between retries.
func withFastDeleteRetry(t *testing.T) {
	t.Helper()
	orig := deleteRetryBaseWait
	deleteRetryBaseWait = time.Millisecond
	t.Cleanup(func() { deleteRetryBaseWait = orig })
}

func TestDeleteResourceWithRetry_SucceedsOnFirstAttempt(t *testing.T) {
	withFastDeleteRetry(t)

	var calls int32
	deleteFunc := func() error {
		atomic.AddInt32(&calls, 1)
		return nil
	}

	if err := DeleteResourceWithRetry(context.Background(), deleteFunc, "kaas", "abc", time.Second); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected 1 call, got %d", got)
	}
}

func TestDeleteResourceWithRetry_NotFoundTreatedAsSuccess(t *testing.T) {
	withFastDeleteRetry(t)

	notFound := newResponseError("delete", "kaas", 404, nil, nil)
	deleteFunc := func() error { return notFound }

	if err := DeleteResourceWithRetry(context.Background(), deleteFunc, "kaas", "abc", time.Second); err != nil {
		t.Fatalf("expected nil for 404, got %v", err)
	}
}

func TestDeleteResourceWithRetry_SucceedsAfterRetry(t *testing.T) {
	withFastDeleteRetry(t)

	boom := newResponseError("delete", "kaas", 500, nil, nil)
	var calls int32
	deleteFunc := func() error {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			return boom
		}
		return nil
	}

	if err := DeleteResourceWithRetry(context.Background(), deleteFunc, "kaas", "abc", 5*time.Second); err != nil {
		t.Fatalf("expected nil after retry, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 calls, got %d", got)
	}
}

func TestDeleteResourceWithRetry_TimeoutReturnsError(t *testing.T) {
	withFastDeleteRetry(t)

	boom := newResponseError("delete", "kaas", 500, nil, nil)
	deleteFunc := func() error { return boom }

	err := DeleteResourceWithRetry(context.Background(), deleteFunc, "kaas", "abc", 20*time.Millisecond)
	if err == nil {
		t.Fatal("expected error on timeout, got nil")
	}
}

func TestDeleteResourceWithRetry_ContextCancelledReturnsError(t *testing.T) {
	withFastDeleteRetry(t)

	boom := newResponseError("delete", "kaas", 500, nil, nil)
	deleteFunc := func() error { return boom }

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := DeleteResourceWithRetry(ctx, deleteFunc, "kaas", "abc", time.Hour)
	if err == nil {
		t.Fatal("expected context-cancelled error, got nil")
	}
}

func TestDeleteResourceWithRetry_ExistsCheckerShortCircuitsAfterFailedDelete(t *testing.T) {
	withFastDeleteRetry(t)

	boom := newResponseError("delete", "kaas", 400, nil, nil)
	var deleteCalls, checkerCalls int32

	deleteFunc := func() error {
		atomic.AddInt32(&deleteCalls, 1)
		return boom
	}
	existsChecker := func(ctx context.Context) (bool, error) {
		atomic.AddInt32(&checkerCalls, 1)
		return true, nil // reports already gone
	}

	if err := DeleteResourceWithRetry(context.Background(), deleteFunc, "kaas", "abc", 5*time.Second, existsChecker); err != nil {
		t.Fatalf("expected nil when existsChecker reports gone, got %v", err)
	}
	if atomic.LoadInt32(&checkerCalls) < 1 {
		t.Fatal("expected existsChecker to be called at least once")
	}
}

// ── Verifies the intended caller pattern: deletion checkers use IsNotFound(err)
// on the GET response to return (true, nil).
func TestWaitForResourceDeleted_RecognisesNotFoundPattern(t *testing.T) {
	withFastPoll(t)

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

	if err := WaitForResourceDeleted(context.Background(), checker, "kaas", "abc", time.Second); err != nil {
		t.Fatalf("expected deletion to be detected via IsNotFound, got %v", err)
	}
}
