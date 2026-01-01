// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ResourceStateChecker is a function that checks the current state of a resource
// Returns the state string and an error if the check failed
type ResourceStateChecker func(ctx context.Context) (string, error)

// WaitForResourceActive waits for a resource to reach an active/ready state
// It polls the resource status until it's not in a transitional state
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

// isReadyState checks if a resource state indicates it's ready to be used
// Resources in "InCreation" or "Deleting" states are not ready
func isReadyState(state string) bool {
	transitionalStates := []string{
		"InCreation",
		"Creating",
		"Deleting",
		"Pending",
		"Provisioning",
	}

	for _, ts := range transitionalStates {
		if state == ts {
			return false
		}
	}

	// If not in a transitional state, consider it ready
	// Common ready states: "Active", "NotUsed", "InUse", "Used", "Stopped", "Running", etc.
	return true
}

// IsDependencyError checks if an API error should be retried.
// According to user requirements: retry on any error except 404 (Resource Not Found).
// We also check for dependency-related keywords in the error message for better logging.
func IsDependencyError(statusCode int, errorTitle, errorDetail *string) bool {
	// 404 means resource not found - don't retry, consider it already deleted
	if statusCode == 404 {
		return false
	}

	// For any other error (400, 409, 500, etc.), retry as it might be a dependency issue
	// The API might return 400 for dependency errors, but we'll retry on any error
	// to handle cases where dependencies are being deleted by Terraform

	// Build error message from title and detail for logging
	errorMsg := ""
	if errorTitle != nil {
		errorMsg = *errorTitle
	}
	if errorDetail != nil {
		if errorMsg != "" {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, *errorDetail)
		} else {
			errorMsg = *errorDetail
		}
	}

	// If error message contains dependency keywords, it's definitely a dependency error
	if errorMsg != "" && containsDependencyKeywords(errorMsg) {
		return true
	}

	// For any non-404 error, retry (might be dependency or transient issue)
	// This handles cases where the API returns errors while dependencies are being cleaned up
	return true
}

// containsDependencyKeywords checks if a string contains keywords that indicate dependency issues
func containsDependencyKeywords(s string) bool {
	keywords := []string{
		"dependency",
		"dependent",
		"depend",
		"cannot delete",
		"can't delete",
		"still in use",
		"in use",
		"has resources",
		"contains resources",
		"has subnets",
		"has security groups",
		"has securitygroup",
		"must be deleted first",
		"delete first",
		"remove first",
		"still exists",
		"associated",
		"attached",
		"linked",
	}

	lower := strings.ToLower(s)
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}

	return false
}

// DeleteResourceWithRetry is a generic function that attempts to delete a resource with retry logic.
// It retries deletion on any error (API errors, network errors, etc.) except 404 (Resource Not Found).
// This handles cases where resources cannot be deleted due to dependencies, transient issues, or network problems.
//
// All SDK responses follow the same structure:
//   - response.IsError() - method to check if error
//   - response.StatusCode - field with HTTP status code
//   - response.Error.Title - field with error title (pointer to string)
//   - response.Error.Detail - field with error detail (pointer to string)
//
// Parameters:
//   - ctx: Context for cancellation
//   - deleteFunc: Function that performs the delete operation and returns (response, error)
//   - extractErrorFunc: Function that extracts error details from the response
//   - resourceType: Human-readable resource type name (e.g., "VPC", "Subnet")
//   - resourceID: Resource identifier for logging
//   - timeout: Maximum time to wait for successful deletion
//
// The function will:
//   - Retry on any error except 404 (consider 404 as already deleted)
//   - Retry on network errors (EOF, connection reset, timeouts, etc.)
//   - Retry on API errors (400, 409, 500, etc.) that might indicate dependencies or transient issues
//   - Use exponential backoff (5s, 10s, 15s, up to 30s max)
//   - Log retry attempts with error details
//   - Return error only if timeout is reached
func DeleteResourceWithRetry(
	ctx context.Context,
	deleteFunc func() (interface{}, error),
	extractErrorFunc func(interface{}) (statusCode int, errorTitle *string, errorDetail *string, isError bool),
	resourceType, resourceID string,
	timeout time.Duration,
) error {
	deadline := time.Now().Add(timeout)
	retryInterval := 5 * time.Second // Start with 5 second intervals
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

			// Attempt deletion
			response, err := deleteFunc()
			if err != nil {
				// Network or other non-API errors - retry as they might be transient
				// Examples: EOF, connection reset, timeout, etc.
				errorMsg := err.Error()
				tflog.Info(ctx, fmt.Sprintf("%s %s deletion failed with network/connection error: %s. Retrying (attempt %d)...", resourceType, resourceID, errorMsg, attempt))

				// Wait before retrying (exponential backoff with max limit)
				waitTime := retryInterval
				if attempt > 1 {
					// Exponential backoff: 5s, 10s, 15s, 20s, 25s, 30s (max)
					waitTime = time.Duration(5*attempt) * time.Second
					if waitTime > maxRetryInterval {
						waitTime = maxRetryInterval
					}
				}

				select {
				case <-ctx.Done():
					return fmt.Errorf("context cancelled while waiting to delete %s %s", resourceType, resourceID)
				case <-time.After(waitTime):
					// Continue to next iteration
					continue
				}
			}

			// Extract error details using provided extractor function
			statusCode, errorTitle, errorDetail, isError := extractErrorFunc(response)

			// Check if deletion succeeded (no error)
			if !isError {
				tflog.Info(ctx, fmt.Sprintf("Successfully deleted %s %s", resourceType, resourceID))
				return nil
			}

			// Check if resource was already deleted (404) - don't retry, consider success
			if statusCode == 404 {
				tflog.Info(ctx, fmt.Sprintf("%s %s already deleted (404)", resourceType, resourceID))
				return nil
			}

			// For any other API error, retry (might be dependency or transient issue)
			// Build error message for logging
			errorMsg := ""
			if errorTitle != nil {
				errorMsg = *errorTitle
			}
			if errorDetail != nil {
				if errorMsg != "" {
					errorMsg = fmt.Sprintf("%s: %s", errorMsg, *errorDetail)
				} else {
					errorMsg = *errorDetail
				}
			}
			if errorMsg == "" {
				errorMsg = fmt.Sprintf("API error (status: %d)", statusCode)
			}

			tflog.Info(ctx, fmt.Sprintf("%s %s deletion failed: %s. Retrying (attempt %d)...", resourceType, resourceID, errorMsg, attempt))

			// Wait before retrying (exponential backoff with max limit)
			waitTime := retryInterval
			if attempt > 1 {
				// Exponential backoff: 5s, 10s, 15s, 20s, 25s, 30s (max)
				waitTime = time.Duration(5*attempt) * time.Second
				if waitTime > maxRetryInterval {
					waitTime = maxRetryInterval
				}
			}

			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled while waiting to delete %s %s", resourceType, resourceID)
			case <-time.After(waitTime):
				// Continue to next iteration
			}
		}
	}
}

// ExtractSDKError extracts error information from SDK responses using reflection.
// All SDK responses follow the same structure: StatusCode (int), Error.Title (*string), Error.Detail (*string)
// This function can be used by all Delete methods to extract error info generically.
func ExtractSDKError(response interface{}) (statusCode int, errorTitle *string, errorDetail *string, isError bool) {
	// First check if response has IsError() method
	type errorResponse interface {
		IsError() bool
	}
	resp, ok := response.(errorResponse)
	if !ok {
		return 0, nil, nil, false
	}
	if !resp.IsError() {
		return 0, nil, nil, false
	}

	// Use reflection to access StatusCode and Error fields
	v := reflect.ValueOf(response)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Access StatusCode field
	statusCodeField := v.FieldByName("StatusCode")
	if !statusCodeField.IsValid() || !statusCodeField.CanInterface() {
		return 0, nil, nil, true // It's an error but can't extract details
	}
	statusCode = int(statusCodeField.Int())

	// Access Error field
	errorField := v.FieldByName("Error")
	if !errorField.IsValid() {
		return statusCode, nil, nil, true
	}

	// Access Error.Title and Error.Detail
	var title, detail *string
	errorVal := errorField
	if errorVal.Kind() == reflect.Ptr {
		if errorVal.IsNil() {
			return statusCode, nil, nil, true
		}
		errorVal = errorVal.Elem()
	}

	if errorVal.Kind() == reflect.Struct {
		titleField := errorVal.FieldByName("Title")
		if titleField.IsValid() && titleField.CanInterface() && !titleField.IsNil() {
			if titlePtr, ok := titleField.Interface().(*string); ok {
				title = titlePtr
			}
		}

		detailField := errorVal.FieldByName("Detail")
		if detailField.IsValid() && detailField.CanInterface() && !detailField.IsNil() {
			if detailPtr, ok := detailField.Interface().(*string); ok {
				detail = detailPtr
			}
		}
	}

	return statusCode, title, detail, true
}

// RetryDeleteOperation is a helper that handles retry logic for delete operations.
// It should be called from Delete methods when an API error occurs (except 404).
// This function handles the retry loop with exponential backoff.
//
// Parameters:
//   - ctx: Context for cancellation
//   - deleteFunc: Function that performs the delete operation
//   - extractError: Function that extracts error info from response
//   - resourceType: Resource type name for logging
//   - resourceID: Resource ID for logging
//   - timeout: Maximum time to wait
//
// Returns error if timeout is reached, nil on success
func RetryDeleteOperation(
	ctx context.Context,
	deleteFunc func() (interface{}, error),
	extractError func(interface{}) (statusCode int, errorTitle *string, errorDetail *string, isError bool),
	resourceType, resourceID string,
	timeout time.Duration,
) error {
	return DeleteResourceWithRetry(ctx, deleteFunc, extractError, resourceType, resourceID, timeout)
}

