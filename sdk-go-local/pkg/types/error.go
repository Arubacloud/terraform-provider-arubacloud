package types

import "encoding/json"

// ValidationError is one field-level validation error (e.g. 400 payloads with an "errors" array).
type ValidationError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
}

// ErrorResponse represents API error bodies: RFC 7807-style fields, optional validation errors, trace id,
// and Extensions for any additional JSON properties.
//
// Examples:
//
//	// 400 validation
//	{
//	  "title": "One or more validation error occurred.",
//	  "status": 400,
//	  "errors": [{"field": "Tag", "message": "length must be at least 4"}]
//	}
//
//	// 404 Not Found
//	{
//	  "type": "https://tools.ietf.org/html/rfc9110#section-15.5.5",
//	  "title": "Not Found",
//	  "status": 404,
//	  "traceId": "00-86dd2c3651e19162ea8007a9c961c43a-a89dcb232bcfacd4-00"
//	}
type ErrorResponse struct {
	// Type A URI reference that identifies the problem type (nullable)
	Type *string `json:"type,omitempty"`

	// Title A short, human-readable summary of the problem type (nullable)
	Title *string `json:"title,omitempty"`

	// Status The HTTP status code (nullable)
	Status *int32 `json:"status,omitempty"`

	// Detail A human-readable explanation specific to this occurrence of the problem (nullable)
	Detail *string `json:"detail,omitempty"`

	// Instance A URI reference that identifies the specific occurrence of the problem (nullable)
	Instance *string `json:"instance,omitempty"`

	// Errors Field-level validation errors
	Errors []ValidationError `json:"errors,omitempty"`

	// TraceID Correlation id for support / logging (e.g. W3C trace context), JSON key "traceId"
	TraceID *string `json:"traceId,omitempty"`

	// Extensions holds JSON properties not mapped to the fields above (forward compatibility).
	Extensions map[string]interface{} `json:"-"`
}

// UnmarshalJSON implements json.Unmarshaler: typed fields plus unknown keys in Extensions.
func (e *ErrorResponse) UnmarshalJSON(data []byte) error {
	type Alias ErrorResponse
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	e.Extensions = make(map[string]interface{})
	knownFields := map[string]bool{
		"type":     true,
		"title":    true,
		"status":   true,
		"detail":   true,
		"instance": true,
		"errors":   true,
		"traceId":  true,
	}

	for key, value := range raw {
		if !knownFields[key] {
			e.Extensions[key] = value
		}
	}

	return nil
}

// MarshalJSON implements json.Marshaler: merges Extensions into the output object.
func (e *ErrorResponse) MarshalJSON() ([]byte, error) {
	type Alias ErrorResponse
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	if len(e.Extensions) == 0 {
		return data, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	for key, value := range e.Extensions {
		result[key] = value
	}

	return json.Marshal(result)
}
