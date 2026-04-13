package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
)

// ProviderErrorCategory classifies an API error as semantic, transient, or technical.
type ProviderErrorCategory int

const (
	// ProviderErrorCategorySemantic represents true field-level validation errors (HTTP 4xx with a
	// non-empty Errors array). These are permanent — the resource spec is invalid and must be corrected.
	ProviderErrorCategorySemantic ProviderErrorCategory = iota + 1
	// ProviderErrorCategoryTransient represents temporary 4xx errors where the ErrorResponse carries
	// no field-level validation details (empty Errors array). Caused by transient conditions such as
	// a dependency resource in a wrong state.
	ProviderErrorCategoryTransient
	// ProviderErrorCategoryTechnical represents infrastructure or transient errors (network failures,
	// HTTP 5xx). These are candidates for retry.
	ProviderErrorCategoryTechnical
)

func (c ProviderErrorCategory) String() string {
	switch c {
	case ProviderErrorCategorySemantic:
		return "semantic"
	case ProviderErrorCategoryTransient:
		return "transient"
	case ProviderErrorCategoryTechnical:
		return "technical"
	default:
		return "unknown"
	}
}

// ProviderError is a structured error produced by Aruba CMP API interactions.
// It carries the error category, HTTP status code, RFC 7807 problem details,
// and the original Go error for transport-level failures.
type ProviderError struct {
	Category   ProviderErrorCategory
	StatusCode int
	Title      string
	Detail     string
	Instance   string
	Operation  string
	Resource   string
	Cause      error
}

// Error implements the error interface.
func (e *ProviderError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("failed to %s %q: %v", e.Operation, e.Resource, e.Cause)
	}
	parts := []string{
		fmt.Sprintf("failed to %s %q", e.Operation, e.Resource),
		fmt.Sprintf("status_code: %d", e.StatusCode),
		fmt.Sprintf("category: %s", e.Category),
	}
	if e.Title != "" {
		parts = append(parts, "title: "+e.Title)
	}
	if e.Detail != "" {
		parts = append(parts, "detail: "+e.Detail)
	}
	if e.Instance != "" {
		parts = append(parts, "instance: "+e.Instance)
	}
	return strings.Join(parts, ", ")
}

// Unwrap returns the underlying transport-level error, enabling errors.Is / errors.As chains.
func (e *ProviderError) Unwrap() error {
	return e.Cause
}

// sanitizeAPIString replaces tab and newline characters with a single space
// and collapses runs of spaces to prevent multi-line noise in error messages.
func sanitizeAPIString(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
}

// formatRawValidationErrors parses the raw HTTP response body and extracts the "errors" field
// as a human-readable string. Used as a fallback when the typed ValidationError extraction
// produces no output (e.g. the API uses different JSON key names than the SDK expects).
// Returns "" when nothing useful can be extracted.
func formatRawValidationErrors(rawBody []byte) string {
	if len(rawBody) == 0 {
		return ""
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &top); err != nil {
		return ""
	}
	errorsRaw, ok := top["errors"]
	if !ok {
		return ""
	}
	// Try to parse as an array of objects for a clean human-readable format.
	var entries []map[string]interface{}
	if err := json.Unmarshal(errorsRaw, &entries); err == nil && len(entries) > 0 {
		parts := make([]string, 0, len(entries))
		for _, entry := range entries {
			keys := make([]string, 0, len(entry))
			for k := range entry {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			pairs := make([]string, 0, len(keys))
			for _, k := range keys {
				pairs = append(pairs, fmt.Sprintf("%s: %v", k, entry[k]))
			}
			parts = append(parts, strings.Join(pairs, ", "))
		}
		return sanitizeAPIString(strings.Join(parts, "; "))
	}
	// Fallback: include raw JSON, truncated to avoid flooding the log.
	raw := sanitizeAPIString(string(errorsRaw))
	const maxLen = 500
	if len(raw) > maxLen {
		raw = raw[:maxLen] + "..."
	}
	return raw
}

// NewTransportError creates a ProviderError for a Go-level transport or network failure.
// Always classified as Technical.
func NewTransportError(operation, resource string, err error) *ProviderError {
	return &ProviderError{
		Category:  ProviderErrorCategoryTechnical,
		Operation: operation,
		Resource:  resource,
		Cause:     err,
	}
}

// newResponseError creates a ProviderError from a non-success HTTP response.
// For HTTP 4xx: Semantic when ErrorResponse.Errors is non-empty (field-level validation
// failures), Transient otherwise. For everything else: Technical.
func newResponseError(operation, resource string, statusCode int, errResp *sdktypes.ErrorResponse, rawBody []byte) *ProviderError {
	category := ProviderErrorCategoryTechnical
	if statusCode >= 400 && statusCode < 500 {
		if errResp != nil && len(errResp.Errors) > 0 {
			category = ProviderErrorCategorySemantic
		} else {
			category = ProviderErrorCategoryTransient
		}
	}

	title, detail, instance := "", "", ""
	if errResp != nil {
		if errResp.Title != nil {
			title = sanitizeAPIString(*errResp.Title)
		}
		if errResp.Detail != nil {
			detail = sanitizeAPIString(*errResp.Detail)
		}
		if errResp.Instance != nil {
			instance = sanitizeAPIString(*errResp.Instance)
		}
		if len(errResp.Errors) > 0 {
			parts := make([]string, 0, len(errResp.Errors))
			for _, ve := range errResp.Errors {
				switch {
				case ve.Field != "" && ve.Message != "":
					parts = append(parts, ve.Field+": "+ve.Message)
				case ve.Message != "":
					parts = append(parts, ve.Message)
				case ve.Field != "":
					parts = append(parts, ve.Field+": invalid")
				}
			}
			if len(parts) > 0 {
				validationDetail := sanitizeAPIString(strings.Join(parts, "; "))
				if detail != "" {
					detail = detail + " | Validation: " + validationDetail
				} else {
					detail = "Validation: " + validationDetail
				}
			} else {
				// Typed Field/Message extraction produced nothing; fall back to raw body.
				rawDetail := formatRawValidationErrors(rawBody)
				if rawDetail != "" {
					if detail != "" {
						detail = detail + " | Validation: " + rawDetail
					} else {
						detail = "Validation: " + rawDetail
					}
				}
			}
		}
	}

	return &ProviderError{
		Category:   category,
		StatusCode: statusCode,
		Title:      title,
		Detail:     detail,
		Instance:   instance,
		Operation:  operation,
		Resource:   resource,
	}
}

// CheckResponse inspects a typed SDK response and returns nil if the status code
// is 2xx, or a *ProviderError otherwise. A nil response is treated as a Technical error.
func CheckResponse[T any](operation, resource string, resp *sdktypes.Response[T]) error {
	if resp == nil {
		return &ProviderError{
			Category:  ProviderErrorCategoryTechnical,
			Operation: operation,
			Resource:  resource,
			Cause:     fmt.Errorf("nil response"),
		}
	}
	if resp.IsSuccess() {
		return nil
	}
	return newResponseError(operation, resource, resp.StatusCode, resp.Error, resp.RawBody)
}

// IsNotFound reports whether err (or any error in its chain) represents a 404 Not Found response.
func IsNotFound(err error) bool {
	var provErr *ProviderError
	return errors.As(err, &provErr) && provErr.StatusCode == 404
}

// ErrorIsSemantic reports whether err is a *ProviderError with category Semantic.
func ErrorIsSemantic(err error) bool {
	var provErr *ProviderError
	return errors.As(err, &provErr) && provErr.Category == ProviderErrorCategorySemantic
}

// ErrorIsTransient reports whether err is a *ProviderError with category Transient.
func ErrorIsTransient(err error) bool {
	var provErr *ProviderError
	return errors.As(err, &provErr) && provErr.Category == ProviderErrorCategoryTransient
}

// ErrorIsTechnical reports whether err is a *ProviderError with category Technical.
func ErrorIsTechnical(err error) bool {
	var provErr *ProviderError
	return errors.As(err, &provErr) && provErr.Category == ProviderErrorCategoryTechnical
}
