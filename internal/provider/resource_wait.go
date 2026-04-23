package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ErrWaitTimeout is returned by WaitForResourceActive or WaitForResourceDeleted
// when the resource does not reach the expected state within the configured
// timeout. Using a typed error lets callers distinguish a timeout (recoverable)
// from a hard failure.
type ErrWaitTimeout struct {
	ResourceType string
	ResourceID   string
	Timeout      time.Duration
	// Operation describes what the resource was waiting to do, e.g. "become active" or "be deleted".
	Operation string
}

func (e *ErrWaitTimeout) Error() string {
	op := e.Operation
	if op == "" {
		op = "become active"
	}
	return fmt.Sprintf("timeout waiting for %s %s to %s (timeout: %v)", e.ResourceType, e.ResourceID, op, e.Timeout)
}

// IsWaitTimeout reports whether err is an *ErrWaitTimeout.
func IsWaitTimeout(err error) bool {
	var t *ErrWaitTimeout
	return errors.As(err, &t)
}

// isFailedState returns true for terminal failure states from which the resource
// will never recover on its own.
func isFailedState(state string) bool {
	switch state {
	case "Failed", "Error", "Errored", "Faulted":
		return true
	}
	return false
}

// ReportWaitResult translates a WaitForResourceActive error into Terraform
// diagnostics. A timeout produces a Warning so the resource is NOT tainted and
// the next terraform apply will reconcile via Read. Any other error (including
// a terminal failure state) produces an Error.
func ReportWaitResult(diags *diag.Diagnostics, err error, resourceType, resourceID string) {
	if IsWaitTimeout(err) {
		diags.AddWarning(
			"Resource Provisioning In Progress",
			fmt.Sprintf("%s %q was created but did not become active within the timeout. "+
				"Run terraform apply again to reconcile. (%s)", resourceType, resourceID, err),
		)
	} else {
		diags.AddError("Resource Provisioning Failed", err.Error())
	}
}

// ResourceStateChecker is a function that checks the current state of a resource.
// Returns the state string and an error if the check failed.
type ResourceStateChecker func(ctx context.Context) (string, error)

// WaitForResourceActive waits for a resource to reach an active/ready state.
// It polls the resource status until it's not in a transitional state.
func WaitForResourceActive(ctx context.Context, checker ResourceStateChecker, resourceType, resourceID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(5 * time.Second) // Poll every 5 seconds
	defer ticker.Stop()

	tflog.Info(ctx, fmt.Sprintf("Waiting for %s %s to become active", resourceType, resourceID))

	consecutiveErrors := 0
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for %s %s", resourceType, resourceID)
		case <-ticker.C:
			if time.Now().After(deadline) {
				return &ErrWaitTimeout{ResourceType: resourceType, ResourceID: resourceID, Timeout: timeout, Operation: "become active"}
			}

			state, err := checker(ctx)
			if err != nil {
				consecutiveErrors++
				tflog.Warn(ctx, fmt.Sprintf("Error checking %s %s status (attempt %d): %v", resourceType, resourceID, consecutiveErrors, err))
				if consecutiveErrors >= 3 {
					return fmt.Errorf("giving up waiting for %s %s after %d consecutive check errors: %w", resourceType, resourceID, consecutiveErrors, err)
				}
				continue
			}
			consecutiveErrors = 0

			// Check if resource reached a terminal failure state.
			if isFailedState(state) {
				return fmt.Errorf("resource reached failed state: %s", state)
			}

			// Check if resource is in a ready state
			if isReadyState(state) {
				tflog.Info(ctx, fmt.Sprintf("%s %s is now active (state: %s)", resourceType, resourceID, state))
				return nil
			}

			tflog.Debug(ctx, fmt.Sprintf("%s %s is still in state: %s, waiting...", resourceType, resourceID, state))
		}
	}
}

// isReadyState checks if a resource state indicates it's ready to be used.
// Resources in "InCreation", "Creating", "Updating", or "Deleting" states are not ready.
func isReadyState(state string) bool {
	transitionalStates := []string{
		"InCreation",
		"Creating",
		"Updating",
		"Deleting",
		"Pending",
		"Provisioning",
	}

	for _, ts := range transitionalStates {
		if state == ts {
			return false
		}
	}

	// If not in a transitional state, consider it ready.
	// Common ready states: "Active", "NotUsed", "InUse", "Used", "Stopped", "Running", etc.
	return true
}

// ResourceDeletedChecker reports whether a resource has been fully deleted.
// Returns (true, nil) when confirmed gone (404 or equivalent).
// Returns (false, nil) when the resource still exists — keep polling.
// Returns (false, err) on unexpected errors.
type ResourceDeletedChecker func(ctx context.Context) (deleted bool, err error)

// waitForDeletedPollInterval is the interval between deletion checks.
// Overridable in tests so the polling loop does not force 10s waits.
var waitForDeletedPollInterval = 10 * time.Second

// WaitForResourceDeleted polls until the resource is confirmed deleted (checker returns true)
// or the timeout elapses. Up to 3 consecutive checker errors are tolerated before giving up,
// mirroring the behaviour of WaitForResourceActive.
//
// An immediate check is performed before the first ticker tick so that resources
// already gone at call time are detected without the 10 s polling delay.
func WaitForResourceDeleted(ctx context.Context, checker ResourceDeletedChecker, resourceType, resourceID string, timeout time.Duration) error {
	tflog.Info(ctx, "waiting for resource deletion", map[string]interface{}{
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"timeout":       timeout.String(),
	})

	consecutiveErrors := 0
	checkDeletion := func() (bool, error) {
		deleted, err := checker(ctx)
		if err != nil {
			consecutiveErrors++
			tflog.Warn(ctx, "deletion check error", map[string]interface{}{
				"resource_type":      resourceType,
				"resource_id":        resourceID,
				"consecutive_errors": consecutiveErrors,
				"error":              err.Error(),
			})
			if consecutiveErrors >= 3 {
				return false, fmt.Errorf("giving up waiting for %s %s deletion after %d consecutive check errors: %w", resourceType, resourceID, consecutiveErrors, err)
			}
			return false, nil
		}
		consecutiveErrors = 0
		if deleted {
			tflog.Info(ctx, "resource confirmed deleted", map[string]interface{}{
				"resource_type": resourceType,
				"resource_id":   resourceID,
			})
			return true, nil
		}
		tflog.Debug(ctx, "resource still exists, waiting for deletion", map[string]interface{}{
			"resource_type": resourceType,
			"resource_id":   resourceID,
		})
		return false, nil
	}

	// Immediate check — returns promptly if the resource is already gone.
	if deleted, err := checkDeletion(); err != nil {
		return err
	} else if deleted {
		return nil
	}

	ticker := time.NewTicker(waitForDeletedPollInterval)
	defer ticker.Stop()

	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for %s %s deletion", resourceType, resourceID)
		case <-timeoutTimer.C:
			return &ErrWaitTimeout{ResourceType: resourceType, ResourceID: resourceID, Timeout: timeout, Operation: "be deleted"}
		case <-ticker.C:
			deleted, err := checkDeletion()
			if err != nil {
				return err
			}
			if deleted {
				return nil
			}
		}
	}
}

// remainingTimeout returns how much of the original timeout budget is left
// since start, clamped to zero. Pass this as the timeout to WaitForResourceDeleted
// so that DeleteResourceWithRetry and WaitForResourceDeleted share a single
// overall budget instead of each consuming the full ResourceTimeout.
func remainingTimeout(start time.Time, total time.Duration) time.Duration {
	r := total - time.Since(start)
	if r < 0 {
		r = 0
	}
	return r
}

// DeleteResourceWithRetry attempts to delete a resource with retry logic.
// deleteFunc must return nil on success, NewTransportError on network failure, or
// CheckResponse on API failure. A 404 response is treated as success (already deleted).
// Retries use exponential backoff up to 30 s between attempts.
//
// An optional existsChecker (same type as ResourceDeletedChecker) can be
// provided as the last argument. When present, it is called via GET before
// each retry so that a resource already deleted by the API (e.g. returns 400
// on a second DELETE but 404 on GET) is recognised as gone immediately.
func DeleteResourceWithRetry(
	ctx context.Context,
	deleteFunc func() error,
	resourceType, resourceID string,
	timeout time.Duration,
	existsChecker ...ResourceDeletedChecker,
) error {
	deadline := time.Now().Add(timeout)
	maxRetryInterval := 30 * time.Second
	attempt := 0

	tflog.Info(ctx, fmt.Sprintf("Attempting to delete %s %s", resourceType, resourceID))

	for {
		attempt++
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while deleting %s %s", resourceType, resourceID)
		default:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting to delete %s %s (timeout: %v, attempts: %d)", resourceType, resourceID, timeout, attempt)
			}

			err := deleteFunc()
			if err == nil {
				tflog.Info(ctx, fmt.Sprintf("Successfully deleted %s %s", resourceType, resourceID))
				return nil
			}
			if IsNotFound(err) {
				tflog.Info(ctx, fmt.Sprintf("%s %s already deleted (404)", resourceType, resourceID))
				return nil
			}

			// Before waiting and retrying, check if the resource is already gone
			// via GET. This handles APIs that return 400 on a second DELETE of an
			// already-deleted (or still-deleting) resource instead of 404.
			if len(existsChecker) > 0 && existsChecker[0] != nil {
				if deleted, _ := existsChecker[0](ctx); deleted {
					tflog.Info(ctx, fmt.Sprintf("%s %s confirmed gone via GET — treating as deleted", resourceType, resourceID))
					return nil
				}
			}

			tflog.Info(ctx, fmt.Sprintf("%s %s deletion failed: %s. Retrying (attempt %d)...", resourceType, resourceID, err, attempt))

			waitTime := time.Duration(5*attempt) * time.Second
			if waitTime > maxRetryInterval {
				waitTime = maxRetryInterval
			}
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled while waiting to delete %s %s", resourceType, resourceID)
			case <-time.After(waitTime):
			}
		}
	}
}
