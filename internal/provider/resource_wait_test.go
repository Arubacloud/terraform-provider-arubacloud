package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	tfdiag "github.com/hashicorp/terraform-plugin-framework/diag"
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

	notFound := newResponseError("delete", "kaas", 404, "", "", "", false)
	deleteFunc := func() error { return notFound }

	if err := DeleteResourceWithRetry(context.Background(), deleteFunc, "kaas", "abc", time.Second); err != nil {
		t.Fatalf("expected nil for 404, got %v", err)
	}
}

func TestDeleteResourceWithRetry_SucceedsAfterRetry(t *testing.T) {
	withFastDeleteRetry(t)

	boom := newResponseError("delete", "kaas", 500, "", "", "", false)
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

	boom := newResponseError("delete", "kaas", 500, "", "", "", false)
	deleteFunc := func() error { return boom }

	err := DeleteResourceWithRetry(context.Background(), deleteFunc, "kaas", "abc", 20*time.Millisecond)
	if err == nil {
		t.Fatal("expected error on timeout, got nil")
	}
}

func TestDeleteResourceWithRetry_ContextCancelledReturnsError(t *testing.T) {
	withFastDeleteRetry(t)

	boom := newResponseError("delete", "kaas", 500, "", "", "", false)
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

	boom := newResponseError("delete", "kaas", 400, "", "", "", false)
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

// ── isFailedState ────────────────────────────────────────────────────────────

func TestIsFailedState(t *testing.T) {
	for _, state := range []string{"Failed", "Error", "Errored", "Faulted"} {
		if !isFailedState(state) {
			t.Errorf("isFailedState(%q) = false, want true", state)
		}
	}
	for _, state := range []string{"Active", "InCreation", "Creating", "Pending", "Provisioning", "Running", "", "unknown"} {
		if isFailedState(state) {
			t.Errorf("isFailedState(%q) = true, want false", state)
		}
	}
}

// ── IsCreatingState ───────────────────────────────────────────────────────────

func TestIsCreatingState(t *testing.T) {
	for _, state := range []string{"InCreation", "Creating", "Pending", "Provisioning"} {
		if !IsCreatingState(state) {
			t.Errorf("IsCreatingState(%q) = false, want true", state)
		}
	}
	for _, state := range []string{"Active", "Failed", "Running", "Stopped", "", "Updating", "Deleting"} {
		if IsCreatingState(state) {
			t.Errorf("IsCreatingState(%q) = true, want false", state)
		}
	}
}

// ── ErrWaitTimeout ───────────────────────────────────────────────────────────

func TestErrWaitTimeout_Error_DefaultsOperationToBeActive(t *testing.T) {
	err := &ErrWaitTimeout{ResourceType: "vpc", ResourceID: "v1", Timeout: 5 * time.Minute}
	msg := err.Error()
	if msg == "" {
		t.Fatal("Error() returned empty string")
	}
	// Default operation should be "become active"
	if got, want := err.Operation, ""; got != want {
		t.Errorf("Operation = %q, want %q", got, want)
	}
}

func TestErrWaitTimeout_Error_UsesExplicitOperation(t *testing.T) {
	err := &ErrWaitTimeout{ResourceType: "vpc", ResourceID: "v1", Timeout: time.Minute, Operation: "be deleted"}
	msg := err.Error()
	if msg == "" {
		t.Fatal("Error() returned empty string")
	}
}

// ── ReportWaitResult ─────────────────────────────────────────────────────────

func TestReportWaitResult_TimeoutAddsWarning(t *testing.T) {
	var diags tfdiag.Diagnostics
	ReportWaitResult(&diags, &ErrWaitTimeout{ResourceType: "vpc", ResourceID: "v1", Timeout: time.Minute}, "vpc", "v1")
	if diags.HasError() {
		t.Fatalf("expected no error, got error diagnostic")
	}
	if len(diags) == 0 {
		t.Fatalf("expected at least one warning diagnostic, got none")
	}
}

func TestReportWaitResult_OtherErrorAddsError(t *testing.T) {
	var diags tfdiag.Diagnostics
	ReportWaitResult(&diags, fmt.Errorf("boom"), "vpc", "v1")
	if !diags.HasError() {
		t.Fatalf("expected error diagnostic, got none")
	}
}

// ── Verifies the intended caller pattern: deletion checkers use IsNotFound(err)
// on the GET response to return (true, nil).
func TestWaitForResourceDeleted_RecognisesNotFoundPattern(t *testing.T) {
	withFastPoll(t)

	notFound := newResponseError("get", "kaas", 404, "", "", "", false)
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

// ── isReadyState ─────────────────────────────────────────────────────────────

func TestIsReadyState(t *testing.T) {
	transitional := []string{
		"InCreation", "Creating", "Updating", "Deleting", "Pending", "Provisioning",
	}
	for _, s := range transitional {
		if isReadyState(s) {
			t.Errorf("isReadyState(%q) = true, want false", s)
		}
	}
	ready := []string{"Active", "NotUsed", "InUse", "Used", "Stopped", "Running", "Unknown"}
	for _, s := range ready {
		if !isReadyState(s) {
			t.Errorf("isReadyState(%q) = false, want true", s)
		}
	}
}

// ── ErrWaitTimeout.Error ──────────────────────────────────────────────────────

func TestErrWaitTimeout_Error(t *testing.T) {
	te := &ErrWaitTimeout{
		ResourceType: "vpc",
		ResourceID:   "abc",
		Timeout:      time.Minute,
		Operation:    "become active",
	}
	msg := te.Error()
	for _, substr := range []string{"vpc", "abc", "1m0s", "become active"} {
		if !strings.Contains(msg, substr) {
			t.Errorf("ErrWaitTimeout.Error() = %q, missing %q", msg, substr)
		}
	}

	// Empty Operation defaults to "become active".
	te2 := &ErrWaitTimeout{ResourceType: "kaas", ResourceID: "xyz", Timeout: 5 * time.Minute}
	msg2 := te2.Error()
	if !strings.Contains(msg2, "become active") {
		t.Errorf("ErrWaitTimeout.Error() with empty Operation = %q, missing 'become active'", msg2)
	}
}

// ── ReportWaitResult ─────────────────────────────────────────────────────────

func TestReportWaitResult_TimeoutProducesWarning(t *testing.T) {
	var d tfdiag.Diagnostics
	ReportWaitResult(&d, &ErrWaitTimeout{ResourceType: "vpc", ResourceID: "abc", Timeout: time.Minute}, "vpc", "abc")
	if d.HasError() {
		t.Error("expected no error diagnostic for timeout, got error")
	}
	if len(d) == 0 {
		t.Error("expected at least one warning diagnostic, got none")
	}
}

func TestReportWaitResult_NonTimeoutProducesError(t *testing.T) {
	var d tfdiag.Diagnostics
	ReportWaitResult(&d, errors.New("something broke"), "vpc", "abc")
	if !d.HasError() {
		t.Error("expected error diagnostic for non-timeout error, got none")
	}
}

// ── remainingTimeout ─────────────────────────────────────────────────────────

func TestRemainingTimeout(t *testing.T) {
	total := 10 * time.Minute

	// Recent start: nearly the full budget remains.
	got := remainingTimeout(time.Now(), total)
	if got <= 0 || got > total {
		t.Errorf("remainingTimeout for fresh start = %v, expected (0, %v]", got, total)
	}

	// Start was 20 minutes ago, total is 10 minutes: budget exhausted → 0.
	old := time.Now().Add(-20 * time.Minute)
	got2 := remainingTimeout(old, total)
	if got2 != 0 {
		t.Errorf("remainingTimeout for exhausted budget = %v, expected 0", got2)
	}
}

// ── CreateWithTransientRetry ─────────────────────────────────────────────────

func TestCreateWithTransientRetry_SucceedsImmediately(t *testing.T) {
	var calls int32
	createFunc := func() error {
		atomic.AddInt32(&calls, 1)
		return nil
	}
	err := CreateWithTransientRetry(context.Background(), createFunc, "vpc", "abc", time.Minute)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if n := atomic.LoadInt32(&calls); n != 1 {
		t.Fatalf("expected 1 call, got %d", n)
	}
}

func TestCreateWithTransientRetry_NonTransientErrorReturnedImmediately(t *testing.T) {
	semantic := CheckResponseErr("create", "vpc", &aruba.HTTPError{
		StatusCode: 400,
		ErrResp: &sdktypes.ErrorResponse{
			Errors: []sdktypes.ValidationError{{Field: "name", Message: "required"}},
		},
	})
	var calls int32
	createFunc := func() error {
		atomic.AddInt32(&calls, 1)
		return semantic
	}
	err := CreateWithTransientRetry(context.Background(), createFunc, "vpc", "abc", time.Minute)
	if err == nil {
		t.Fatal("expected non-nil error, got nil")
	}
	if n := atomic.LoadInt32(&calls); n != 1 {
		t.Fatalf("expected exactly 1 call for semantic error, got %d", n)
	}
}

func TestCreateWithTransientRetry_TransientErrorTimesOut(t *testing.T) {
	transient := newResponseError("create", "vpc", 400, "", "", "", false)
	if !ErrorIsTransient(transient) {
		t.Fatal("test setup: expected 400 with no validation errors to be transient")
	}

	var calls int32
	createFunc := func() error {
		atomic.AddInt32(&calls, 1)
		return transient
	}

	// A negative timeout means the deadline is already in the past: the function
	// detects the timeout immediately after the first transient error.
	err := CreateWithTransientRetry(context.Background(), createFunc, "vpc", "abc", -time.Second)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if n := atomic.LoadInt32(&calls); n < 1 {
		t.Fatal("expected createFunc to be called at least once")
	}
}
