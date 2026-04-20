package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ErrWaitTimeout is returned by WaitForResourceActive when the resource does not
// reach an active state within the configured timeout. Using a typed error lets
// callers distinguish a timeout (recoverable) from a hard failure.
type ErrWaitTimeout struct {
	ResourceType string
	ResourceID   string
	Timeout      time.Duration
}

func (e *ErrWaitTimeout) Error() string {
	return fmt.Sprintf("timeout waiting for %s %s to become active (timeout: %v)", e.ResourceType, e.ResourceID, e.Timeout)
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
				return &ErrWaitTimeout{ResourceType: resourceType, ResourceID: resourceID, Timeout: timeout}
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

// WaitForResourceDeleted polls until the resource is confirmed deleted (checker returns true)
// or the timeout elapses. Up to 3 consecutive checker errors are tolerated before giving up,
// mirroring the behaviour of WaitForResourceActive.
func WaitForResourceDeleted(ctx context.Context, checker ResourceDeletedChecker, resourceType, resourceID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	tflog.Info(ctx, fmt.Sprintf("Waiting for %s %s to be deleted", resourceType, resourceID))

	consecutiveErrors := 0
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for %s %s deletion", resourceType, resourceID)
		case <-ticker.C:
			if time.Now().After(deadline) {
				return &ErrWaitTimeout{ResourceType: resourceType, ResourceID: resourceID, Timeout: timeout}
			}

			deleted, err := checker(ctx)
			if err != nil {
				consecutiveErrors++
				tflog.Warn(ctx, fmt.Sprintf("Error checking deletion of %s %s (attempt %d): %v", resourceType, resourceID, consecutiveErrors, err))
				if consecutiveErrors >= 3 {
					return fmt.Errorf("giving up waiting for %s %s deletion after %d consecutive check errors: %w", resourceType, resourceID, consecutiveErrors, err)
				}
				continue
			}
			consecutiveErrors = 0

			if deleted {
				tflog.Info(ctx, fmt.Sprintf("%s %s has been deleted", resourceType, resourceID))
				return nil
			}

			tflog.Debug(ctx, fmt.Sprintf("%s %s still exists, waiting for deletion...", resourceType, resourceID))
		}
	}
}

// DeleteResourceWithRetry attempts to delete a resource with retry logic.
// deleteFunc must return nil on success, NewTransportError on network failure, or
// CheckResponse on API failure. A 404 response is treated as success (already deleted).
// Retries use exponential backoff up to 30 s between attempts.
func DeleteResourceWithRetry(
	ctx context.Context,
	deleteFunc func() error,
	resourceType, resourceID string,
	timeout time.Duration,
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
