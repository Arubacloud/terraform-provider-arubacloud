package types

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Resource Metadata Request
type ResourceMetadataRequest struct {
	Name string   `json:"name"`
	Tags []string `json:"tags,omitempty"`
}

// Regional Resource Metadata Request
type RegionalResourceMetadataRequest struct {
	ResourceMetadataRequest
	Location *LocationRequest `json:"location,omitempty"`
}

type LocationRequest struct {
	Value Region `json:"value"`
}

// Resource Metadata Response
type LocationResponse struct {
	Code    string `json:"code,omitempty"`
	Country string `json:"country,omitempty"`
	Name    string `json:"region,omitempty"`
	City    string `json:"city,omitempty"`
	Value   Region `json:"value,omitempty"`
}

type ProjectMetadataResponse struct {
	ID string `json:"id,omitempty"`
}

type ResourceRequest struct {
	Metadata *ResourceMetadataRequest `json:"metadata"`
}

type TypologyMetadataResponse struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type CategoryMetadataResponse struct {
	Name     string                   `json:"name,omitempty"`
	Provider string                   `json:"provider,omitempty"`
	Typology TypologyMetadataResponse `json:"typology,omitempty"`
}

type ResourceMetadataResponse struct {
	ID                      *string                   `json:"id,omitempty"`
	URI                     *string                   `json:"uri,omitempty"`
	Name                    *string                   `json:"name,omitempty"`
	LocationResponse        *LocationResponse         `json:"location,omitempty"`
	ProjectMetadataResponse *ProjectMetadataResponse  `json:"project,omitempty"`
	Tags                    []string                  `json:"tags,omitempty"`
	Category                *CategoryMetadataResponse `json:"category,omitempty"`
	CreationDate            *time.Time                `json:"creationDate,omitempty"`
	CreatedBy               *string                   `json:"createdBy,omitempty"`
	UpdateDate              *time.Time                `json:"updateDate,omitempty"`
	UpdatedBy               *string                   `json:"updatedBy,omitempty"`
	Version                 *string                   `json:"version,omitempty"`
	CreatedUser             *string                   `json:"createdUser,omitempty"`
	UpdatedUser             *string                   `json:"updatedUser,omitempty"`
}

// Status
type PreviousStatusResponse struct {
	State        *State     `json:"state,omitempty"`
	CreationDate *time.Time `json:"creationDate,omitempty"`
}

type DisableStatusInfoResponse struct {
	IsDisabled bool     `json:"isDisabled,omitempty"`
	Reasons    []string `json:"reasons,omitempty"`
}

type ResourceStatusResponse struct {
	State                     *State                     `json:"state,omitempty"`
	CreationDate              *time.Time                 `json:"creationDate,omitempty"`
	DisableStatusInfoResponse *DisableStatusInfoResponse `json:"disableStatusInfo,omitempty"`
	PreviousStatusResponse    *PreviousStatusResponse    `json:"previousStatus,omitempty"`
	FailureReason             *string                    `json:"failureReason,omitempty"`
}

// LinkedResourceCommon represents a resource linked
type LinkedResourceCommon struct {
	// URI of the linked resource
	URI string `json:"uri"`

	// StrictCorrelation indicates strict correlation with the resource
	StrictCorrelation bool `json:"strictCorrelation"`
}

// BillingPeriod identifies the billing cycle for a resource.
//
// The platform accepts "Hour" as the primary value. Additional periods may
// be available — consult the resource-specific API documentation for the
// authoritative list.
type BillingPeriod string

const (
	BillingPeriodHour  BillingPeriod = "Hour"
	BillingPeriodMonth BillingPeriod = "Month"
	BillingPeriodYear  BillingPeriod = "Year"
)

// BillingPlanCommon is the nested wire wrapper used by resources whose API encodes
// billing inside a billingPlan object rather than as a flat billingPeriod field.
type BillingPlanCommon struct {
	BillingPeriod *BillingPeriod `json:"billingPeriod,omitempty"`
}

type ReferenceResourceCommon struct {
	URI string `json:"uri"`
}

type ListResponse struct {
	// Total number of items
	Total int64 `json:"total"`

	// Self link to current page
	Self string `json:"self"`

	// Prev link to previous page
	Prev string `json:"prev"`

	// Next link to next page
	Next string `json:"next"`

	// First link to first page
	First string `json:"first"`

	// Last link to last page
	Last string `json:"last"`
}

// BaseList returns the embedded pagination/total metadata. Promoted onto every
// per-resource list payload (VPCList, AlertsListResponse, …) via Go's method-
// promotion rules, so a generic helper can extract pagination fields uniformly.
func (lr ListResponse) BaseList() ListResponse { return lr }

// Response wraps an HTTP response with parsed data
type Response[T any] struct {
	// Data contains the parsed response body (for 2xx responses)
	Data *T

	// Error contains the parsed error response (for 4xx/5xx responses)
	Error *ErrorResponse

	// HTTPResponse is the underlying HTTP response
	HTTPResponse *http.Response

	// StatusCode is the HTTP status code
	StatusCode int

	// Headers contains the response headers
	Headers http.Header

	// RawBody contains the raw response body (if requested)
	RawBody []byte
}

// MetadataValidationError reports which required identity fields were absent
// in a successful API response. Callers can use errors.As to branch on this
// specifically and distinguish API-contract violations from network errors.
type MetadataValidationError struct {
	Missing []string
}

func (e *MetadataValidationError) Error() string {
	return fmt.Sprintf("response metadata missing required field(s): %s", strings.Join(e.Missing, ", "))
}

// Validate returns a *MetadataValidationError if ID or Name are absent or
// empty. The Aruba Cloud API must populate both on a successful Create
// response; callers rely on them to identify the resource.
func (m *ResourceMetadataResponse) Validate() error {
	if m == nil {
		return &MetadataValidationError{Missing: []string{"metadata"}}
	}
	var missing []string
	if m.ID == nil || *m.ID == "" {
		missing = append(missing, "id")
	}
	if m.Name == nil || *m.Name == "" {
		missing = append(missing, "name")
	}
	if len(missing) > 0 {
		return &MetadataValidationError{Missing: missing}
	}
	return nil
}

// IsSuccess returns true if the status code is 2xx
func (r *Response[T]) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsError returns true if the status code is 4xx or 5xx
func (r *Response[T]) IsError() bool {
	return r.StatusCode >= 400
}
