// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// FormatAPIError formats an API error response with detailed validation errors
// from the Extensions field, providing user-friendly error messages.
func FormatAPIError(ctx context.Context, err *sdktypes.ErrorResponse, baseMsg string, logContext map[string]interface{}) string {
	if err == nil {
		return baseMsg
	}

	errorMsg := baseMsg
	if err.Title != nil {
		errorMsg = fmt.Sprintf("%s: %s", errorMsg, *err.Title)
	}
	if err.Detail != nil {
		errorMsg = fmt.Sprintf("%s - %s", errorMsg, *err.Detail)
	}

	// Add validation errors if present in Extensions
	if len(err.Extensions) > 0 {
		if errorsArray, ok := err.Extensions["errors"].([]interface{}); ok {
			errorMsg += "\n\nValidation Errors:"
			for _, errItem := range errorsArray {
				if errMap, ok := errItem.(map[string]interface{}); ok {
					fieldName := errMap["fieldName"]
					errorMessage := errMap["errorMessage"]
					errorMsg += fmt.Sprintf("\n  - %v: %v", fieldName, errorMessage)
				}
			}
		} else {
			// If not in expected format, include all extensions
			errorMsg += "\n\nAdditional Error Details:"
			for key, value := range err.Extensions {
				errorMsg += fmt.Sprintf("\n  - %s: %v", key, value)
			}
		}
	}

	// Build error details for logging
	errorDetails := logContext
	if errorDetails == nil {
		errorDetails = make(map[string]interface{})
	}
	if err.Title != nil {
		errorDetails["error_title"] = *err.Title
	}
	if err.Detail != nil {
		errorDetails["error_detail"] = *err.Detail
	}
	if err.Status != nil {
		errorDetails["error_status"] = *err.Status
	}
	if err.Type != nil {
		errorDetails["error_type"] = *err.Type
	}
	if err.Extensions != nil {
		errorDetails["error_extensions"] = err.Extensions
	}

	// Log full error response JSON
	if errorJSON, jsonErr := json.MarshalIndent(err, "", "  "); jsonErr == nil {
		tflog.Error(ctx, "Full API error response JSON", map[string]interface{}{
			"error_json": string(errorJSON),
		})
	}

	tflog.Error(ctx, "API request failed", errorDetails)

	return errorMsg
}
