package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

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

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for %s %s", resourceType, resourceID)
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for %s %s to become active (timeout: %v)", resourceType, resourceID, timeout)
			}

			state, err := checker(ctx)
			if err != nil {
				tflog.Warn(ctx, fmt.Sprintf("Error checking %s %s status: %v", resourceType, resourceID, err))
				continue
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
