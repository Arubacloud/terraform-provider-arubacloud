package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
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

func TestCheckResponseErr_TypedExtractionWorks(t *testing.T) {
	errResp := &sdktypes.ErrorResponse{
		Title: strPtr("One or more validation errors occurred."),
		Errors: []sdktypes.ValidationError{
			{Field: "Tag", Message: "length must be at least 4"},
		},
	}
	rawBody := []byte(`{"title":"One or more validation errors occurred.","errors":[{"propertyName":"Tag","errorMessage":"length must be at least 4"}]}`)
	httpErr := &aruba.HTTPError{StatusCode: 400, ErrResp: errResp, Body: rawBody}

	err := CheckResponseErr("create", "Schedulejob", httpErr)

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

func TestCheckResponseErr_FallbackToRawBody(t *testing.T) {
	// Errors array has entries but Field/Message are empty (API uses different keys).
	errResp := &sdktypes.ErrorResponse{
		Title: strPtr("One or more validation errors occurred."),
		Errors: []sdktypes.ValidationError{
			{}, // empty Field and Message
		},
	}
	rawBody := []byte(`{"title":"One or more validation errors occurred.","errors":[{"propertyName":"CronExpression","errorMessage":"invalid cron format"}]}`)
	httpErr := &aruba.HTTPError{StatusCode: 400, ErrResp: errResp, Body: rawBody}

	err := CheckResponseErr("create", "Schedulejob", httpErr)

	if err.Category != ProviderErrorCategorySemantic {
		t.Errorf("expected semantic, got %v", err.Category)
	}
	if !contains(err.Detail, "CronExpression") {
		t.Errorf("expected raw fallback detail to contain CronExpression, got %q", err.Detail)
	}
}

func TestCheckResponseErr_NilBodyDoesNotPanic(t *testing.T) {
	errResp := &sdktypes.ErrorResponse{
		Title:  strPtr("One or more validation errors occurred."),
		Errors: []sdktypes.ValidationError{{}},
	}
	httpErr := &aruba.HTTPError{StatusCode: 400, ErrResp: errResp, Body: nil}

	err := CheckResponseErr("create", "Schedulejob", httpErr)

	if err.Category != ProviderErrorCategorySemantic {
		t.Errorf("expected semantic, got %v", err.Category)
	}
	// No detail, but no panic.
	_ = err.Error()
}

func TestCheckResponseErr_TransientNotConsultingRawBody(t *testing.T) {
	errResp := &sdktypes.ErrorResponse{
		Title:  strPtr("Conflict"),
		Errors: []sdktypes.ValidationError{}, // empty → transient
	}
	rawBody := []byte(`{"title":"Conflict","errors":[{"propertyName":"Name","errorMessage":"must be unique"}]}`)
	httpErr := &aruba.HTTPError{StatusCode: 409, ErrResp: errResp, Body: rawBody}

	err := CheckResponseErr("create", "Resource", httpErr)

	if err.Category != ProviderErrorCategoryTransient {
		t.Errorf("expected transient, got %v", err.Category)
	}
	// Raw body should not be consulted for transient errors (empty Errors array).
	if contains(err.Detail, "propertyName") {
		t.Errorf("transient errors should not include raw fallback, got %q", err.Detail)
	}
}

func TestCheckResponseErr_NilErrResp(t *testing.T) {
	err := CheckResponseErr("create", "Resource", &aruba.HTTPError{StatusCode: 500, ErrResp: nil})
	if err.Category != ProviderErrorCategoryTechnical {
		t.Errorf("expected technical, got %v", err.Category)
	}
	_ = err.Error()
}

func makeHTTPErr(statusCode int, errResp *sdktypes.ErrorResponse) error {
	return &aruba.HTTPError{StatusCode: statusCode, ErrResp: errResp}
}

func TestCheckResponseErr(t *testing.T) {
	title404 := "Not Found"
	title400 := "Validation failed"

	cases := []struct {
		name         string
		err          error
		wantNil      bool
		wantCategory ProviderErrorCategory
	}{
		{
			name:    "nil error → nil (success)",
			err:     nil,
			wantNil: true,
		},
		{
			name:         "404 Not Found (no validation errors) → transient",
			err:          makeHTTPErr(404, &sdktypes.ErrorResponse{Title: &title404}),
			wantCategory: ProviderErrorCategoryTransient,
		},
		{
			name: "400 with validation errors → semantic",
			err: makeHTTPErr(400, &sdktypes.ErrorResponse{
				Title: &title400,
				Errors: []sdktypes.ValidationError{
					{Field: "name", Message: "is required"},
				},
			}),
			wantCategory: ProviderErrorCategorySemantic,
		},
		{
			name:         "500 Internal Server Error → technical",
			err:          makeHTTPErr(500, nil),
			wantCategory: ProviderErrorCategoryTechnical,
		},
		{
			name:         "non-HTTP transport error → technical",
			err:          fmt.Errorf("dial tcp: connection refused"),
			wantCategory: ProviderErrorCategoryTechnical,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provErr := CheckResponseErr("create", "Resource", tc.err)
			if tc.wantNil {
				if provErr != nil {
					t.Errorf("expected nil, got %v", provErr)
				}
				return
			}
			if provErr == nil {
				t.Fatal("expected non-nil *ProviderError, got nil")
			}
			if provErr.Category != tc.wantCategory {
				t.Errorf("category: got %v, want %v", provErr.Category, tc.wantCategory)
			}
		})
	}
}

func TestCheckResponseErr_IsNotFound(t *testing.T) {
	title := "Not Found"
	err := CheckResponseErr("get", "VPC", makeHTTPErr(404, &sdktypes.ErrorResponse{Title: &title}))
	if err == nil {
		t.Fatal("expected non-nil error for 404")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound should return true for 404, got false; err=%v", err)
	}
}

func TestIsNotFound(t *testing.T) {
	if IsNotFound(nil) {
		t.Error("nil error should not be not-found")
	}
	notFound := &ProviderError{StatusCode: 404, Category: ProviderErrorCategoryTransient}
	if !IsNotFound(notFound) {
		t.Error("404 error should be not-found")
	}
	other := &ProviderError{StatusCode: 500, Category: ProviderErrorCategoryTechnical}
	if IsNotFound(other) {
		t.Error("500 error should not be not-found")
	}
}

func TestErrorCategoryHelpers(t *testing.T) {
	semantic := &ProviderError{Category: ProviderErrorCategorySemantic}
	transient := &ProviderError{Category: ProviderErrorCategoryTransient}
	technical := &ProviderError{Category: ProviderErrorCategoryTechnical}

	if !ErrorIsSemantic(semantic) {
		t.Error("expected semantic")
	}
	if ErrorIsSemantic(transient) {
		t.Error("transient should not be semantic")
	}
	if !ErrorIsTransient(transient) {
		t.Error("expected transient")
	}
	if !ErrorIsTechnical(technical) {
		t.Error("expected technical")
	}
	if ErrorIsTechnical(nil) {
		t.Error("nil should not be technical")
	}
}

// ── sanitizeAPIString ─────────────────────────────────────────────────────────

func TestSanitizeAPIString(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"hello", "hello"},
		{"tab\there", "tab here"},
		{"new\nline", "new line"},
		{"multiple  spaces", "multiple spaces"},
		{"\t\n  \t", ""},
		{"  leading and trailing  ", "leading and trailing"},
		{"multiple   spaces   here", "multiple spaces here"},
	}
	for _, tc := range cases {
		if got := sanitizeAPIString(tc.in); got != tc.want {
			t.Errorf("sanitizeAPIString(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// ── ProviderErrorCategory.String ─────────────────────────────────────────────

func TestProviderErrorCategory_String(t *testing.T) {
	cases := []struct {
		cat  ProviderErrorCategory
		want string
	}{
		{ProviderErrorCategorySemantic, "semantic"},
		{ProviderErrorCategoryTransient, "transient"},
		{ProviderErrorCategoryTechnical, "technical"},
		{ProviderErrorCategory(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.cat.String(); got != tc.want {
			t.Errorf("ProviderErrorCategory(%d).String() = %q, want %q", tc.cat, got, tc.want)
		}
	}
}

// ── ProviderError.Error ───────────────────────────────────────────────────────

func TestProviderError_Error(t *testing.T) {
	// With Cause: message should mention operation, resource, and cause text.
	cause := errors.New("network failure")
	e1 := &ProviderError{
		Category:  ProviderErrorCategoryTechnical,
		Operation: "create",
		Resource:  "vpc",
		Cause:     cause,
	}
	msg1 := e1.Error()
	for _, sub := range []string{"create", "vpc", "network failure"} {
		if !strings.Contains(msg1, sub) {
			t.Errorf("ProviderError.Error() with Cause = %q, missing %q", msg1, sub)
		}
	}

	// Without Cause, full fields: message should contain status code, category,
	// title, detail, and instance.
	e2 := &ProviderError{
		Category:   ProviderErrorCategorySemantic,
		StatusCode: 400,
		Title:      "Bad Request",
		Detail:     "name is required",
		Instance:   "/api/v1/vpc",
		Operation:  "create",
		Resource:   "vpc",
	}
	msg2 := e2.Error()
	for _, sub := range []string{"400", "Bad Request", "name is required", "/api/v1/vpc"} {
		if !strings.Contains(msg2, sub) {
			t.Errorf("ProviderError.Error() without Cause = %q, missing %q", msg2, sub)
		}
	}

	// Without Cause, empty optional fields: no panic.
	e3 := &ProviderError{
		Category:   ProviderErrorCategoryTechnical,
		StatusCode: 500,
		Operation:  "delete",
		Resource:   "vpc",
	}
	msg3 := e3.Error()
	if !strings.Contains(msg3, "500") || !strings.Contains(msg3, "delete") {
		t.Errorf("ProviderError.Error() minimal = %q, expected status+operation", msg3)
	}
}

// ── ProviderError.Unwrap ──────────────────────────────────────────────────────

func TestProviderError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	e := &ProviderError{Cause: cause, Operation: "create", Resource: "vpc"}
	if !errors.Is(e, cause) {
		t.Error("errors.Is should find cause via Unwrap")
	}

	// No Cause → Unwrap returns nil.
	e2 := &ProviderError{StatusCode: 400, Operation: "create", Resource: "vpc"}
	if errors.Unwrap(e2) != nil {
		t.Error("Unwrap should return nil when Cause is not set")
	}
}

// ── NewTransportError ─────────────────────────────────────────────────────────

func TestNewTransportError(t *testing.T) {
	cause := errors.New("dial error")
	err := NewTransportError("create", "vpc", cause)

	if err.Category != ProviderErrorCategoryTechnical {
		t.Errorf("expected Technical category, got %v", err.Category)
	}
	if err.Operation != "create" {
		t.Errorf("expected operation 'create', got %q", err.Operation)
	}
	if err.Resource != "vpc" {
		t.Errorf("expected resource 'vpc', got %q", err.Resource)
	}
	if !errors.Is(err, cause) {
		t.Error("expected errors.Is to find cause via Unwrap")
	}
	if !ErrorIsTechnical(err) {
		t.Error("ErrorIsTechnical should return true for transport error")
	}
	if ErrorIsTransient(err) || ErrorIsSemantic(err) {
		t.Error("transport error should not be transient or semantic")
	}
}
