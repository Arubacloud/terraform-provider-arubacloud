package types

import "time"

// EventOperationResponse represents an operation in the audit log
type EventOperationResponse struct {
	ID    string  `json:"id"`
	Value *string `json:"value,omitempty"`
}

// EventInfoResponse represents event information
type EventInfoResponse struct {
	ID    string  `json:"id"`
	Value *string `json:"value,omitempty"`
	Type  string  `json:"type"`
}

// EventCategoryResponse represents the event category
type EventCategoryResponse struct {
	Value       string  `json:"value"`
	Description *string `json:"description,omitempty"`
}

// EventRegionInfoResponse represents the region information in an audit log event.
type EventRegionInfoResponse struct {
	Name             *string `json:"name,omitempty"`
	AvailabilityZone *string `json:"availabilityZone,omitempty"`
}

// EventStatusResponse represents the status of the event
type EventStatusResponse struct {
	Value       string                 `json:"value"`
	Description *string                `json:"description,omitempty"`
	Code        *int32                 `json:"code,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// EventSubStatusResponse represents the sub-status of the event
type EventSubStatusResponse struct {
	Value       *string                `json:"value,omitempty"`
	Description *string                `json:"description,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// EventCallerResponse represents the caller identity
type EventCallerResponse struct {
	Subject  string  `json:"subject"`
	Username *string `json:"username,omitempty"`
	Company  *string `json:"company,omitempty"`
	TenantID *string `json:"tenantId,omitempty"`
}

// EventIdentityResponse represents the identity information
type EventIdentityResponse struct {
	Caller     EventCallerResponse    `json:"caller"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// EventActionResponse represents an available action
type EventActionResponse struct {
	Key        *string `json:"key,omitempty"`
	Disabled   *bool   `json:"disabled,omitempty"`
	Executable *bool   `json:"executable,omitempty"`
}

// EventLogFormatVersionResponse represents the log format version
type EventLogFormatVersionResponse struct {
	Version string `json:"version"`
}

// AuditEventResponse represents the complete audit event response
type AuditEventResponse struct {
	SeverityLevel string                        `json:"severityLevel"`
	LogFormat     EventLogFormatVersionResponse `json:"logFormat"`
	Timestamp     time.Time                     `json:"@timestamp"`
	Operation     EventOperationResponse        `json:"operation"`
	Event         EventInfoResponse             `json:"event"`
	Category      EventCategoryResponse         `json:"category"`
	Region        *EventRegionInfoResponse      `json:"region,omitempty"`
	Origin        string                        `json:"origin"`
	Channel       string                        `json:"channel"`
	Status        EventStatusResponse           `json:"status"`
	SubStatus     *EventSubStatusResponse       `json:"subStatus,omitempty"`
	Identity      EventIdentityResponse         `json:"identity"`
	Properties    map[string]interface{}        `json:"properties,omitempty"`
	Actions       []EventActionResponse         `json:"actions,omitempty"`
	CategoryID    *string                       `json:"categoryId,omitempty"`
	TypologyID    *string                       `json:"typologyId,omitempty"`
	Title         *string                       `json:"title,omitempty"`
}

// AuditEventListResponse represents a paginated list of audit events
type AuditEventListResponse struct {
	ListResponse
	Values []AuditEventResponse `json:"values"`
}
