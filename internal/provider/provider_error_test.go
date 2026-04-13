package provider

import (
	"testing"

	sdktypes "github.com/Arubacloud/sdk-go/pkg/types"
)

func strPtr(s string) *string { return &s }

func TestFormatRawValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		rawBody  []byte
		wantPart string // substring that must appear (or "" to check empty return)
		wantLen  int    // 0 means "not empty", -1 means "must be empty"
	}{
		{
			name:    "nil input",
			rawBody: nil,
			wantLen: -1,
		},
		{
			name:    "empty input",
			rawBody: []byte{},
			wantLen: -1,
		},
		{
			name:    "invalid JSON",
			rawBody: []byte(`not-json`),
			wantLen: -1,
		},
		{
			name:    "no errors key",
			rawBody: []byte(`{"title":"oops","status":400}`),
			wantLen: -1,
		},
		{
			name:     "non-standard keys propertyName/errorMessage",
			rawBody:  []byte(`{"errors":[{"propertyName":"CronExpression","errorMessage":"invalid cron format"}]}`),
			wantPart: "CronExpression",
		},
		{
			name:     "standard field/message keys",
			rawBody:  []byte(`{"errors":[{"field":"Tag","message":"length must be at least 4"}]}`),
			wantPart: "Tag",
		},
		{
			name:     "multiple entries",
			rawBody:  []byte(`{"errors":[{"propertyName":"Name","errorMessage":"required"},{"propertyName":"Tag","errorMessage":"too short"}]}`),
			wantPart: "Name",
		},
		{
			name:     "errors as map (non-standard ASP.NET format) falls back to raw string",
			rawBody:  []byte(`{"errors":{"Tag":["too short"],"Name":["required"]}}`),
			wantPart: "Tag",
		},
		{
			name: "long error truncated",
			rawBody: func() []byte {
				long := make([]byte, 600)
				for i := range long {
					long[i] = 'x'
				}
				return []byte(`{"errors":"` + string(long) + `"}`)
			}(),
			wantPart: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatRawValidationErrors(tt.rawBody)
			if tt.wantLen == -1 {
				if got != "" {
					t.Errorf("expected empty string, got %q", got)
				}
				return
			}
			if got == "" {
				t.Errorf("expected non-empty string, got empty")
				return
			}
			if tt.wantPart != "" && !contains(got, tt.wantPart) {
				t.Errorf("expected %q to contain %q", got, tt.wantPart)
			}
		})
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

func TestNewResponseError_TypedExtractionWorks(t *testing.T) {
	errResp := &sdktypes.ErrorResponse{
		Title: strPtr("One or more validation errors occurred."),
		Errors: []sdktypes.ValidationError{
			{Field: "Tag", Message: "length must be at least 4"},
		},
	}
	rawBody := []byte(`{"title":"One or more validation errors occurred.","errors":[{"propertyName":"Tag","errorMessage":"length must be at least 4"}]}`)

	err := newResponseError("create", "Schedulejob", 400, errResp, rawBody)

	if err.Category != ProviderErrorCategorySemantic {
		t.Errorf("expected semantic, got %v", err.Category)
	}
	if !contains(err.Detail, "Tag: length must be at least 4") {
		t.Errorf("expected typed extraction in detail, got %q", err.Detail)
	}
	// Raw fallback must NOT be appended when typed extraction succeeds.
	if contains(err.Detail, "propertyName") {
		t.Errorf("raw fallback should not appear when typed extraction succeeded, got %q", err.Detail)
	}
}

func TestNewResponseError_FallbackToRawBody(t *testing.T) {
	// Errors array has entries but Field/Message are empty (API uses different keys).
	errResp := &sdktypes.ErrorResponse{
		Title: strPtr("One or more validation errors occurred."),
		Errors: []sdktypes.ValidationError{
			{}, // empty Field and Message
		},
	}
	rawBody := []byte(`{"title":"One or more validation errors occurred.","errors":[{"propertyName":"CronExpression","errorMessage":"invalid cron format"}]}`)

	err := newResponseError("create", "Schedulejob", 400, errResp, rawBody)

	if err.Category != ProviderErrorCategorySemantic {
		t.Errorf("expected semantic, got %v", err.Category)
	}
	if !contains(err.Detail, "CronExpression") {
		t.Errorf("expected raw fallback detail to contain CronExpression, got %q", err.Detail)
	}
}

func TestNewResponseError_NilRawBodyDoesNotPanic(t *testing.T) {
	errResp := &sdktypes.ErrorResponse{
		Title:  strPtr("One or more validation errors occurred."),
		Errors: []sdktypes.ValidationError{{}},
	}

	err := newResponseError("create", "Schedulejob", 400, errResp, nil)

	if err.Category != ProviderErrorCategorySemantic {
		t.Errorf("expected semantic, got %v", err.Category)
	}
	// No detail, but no panic.
	_ = err.Error()
}

func TestNewResponseError_TransientNotConsultingRawBody(t *testing.T) {
	errResp := &sdktypes.ErrorResponse{
		Title:  strPtr("Conflict"),
		Errors: []sdktypes.ValidationError{}, // empty → transient
	}
	rawBody := []byte(`{"title":"Conflict","errors":[{"propertyName":"Name","errorMessage":"must be unique"}]}`)

	err := newResponseError("create", "Resource", 409, errResp, rawBody)

	if err.Category != ProviderErrorCategoryTransient {
		t.Errorf("expected transient, got %v", err.Category)
	}
	// Raw body should not be consulted for transient errors (empty Errors array).
	if contains(err.Detail, "propertyName") {
		t.Errorf("transient errors should not include raw fallback, got %q", err.Detail)
	}
}

func TestNewResponseError_NilErrResp(t *testing.T) {
	err := newResponseError("create", "Resource", 500, nil, nil)
	if err.Category != ProviderErrorCategoryTechnical {
		t.Errorf("expected technical, got %v", err.Category)
	}
	_ = err.Error()
}
