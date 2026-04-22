package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestLogAndAppendAPIError(t *testing.T) {
	var diags diag.Diagnostics
	err := errors.New("connection refused")

	LogAndAppendAPIError(context.TODO(), &diags, "API Error", err, map[string]any{
		"project_id":  "proj-123",
		"resource_id": "res-456",
	})

	if !diags.HasError() {
		t.Fatal("expected diagnostic error, got none")
	}
	if diags[0].Summary() != "API Error" {
		t.Errorf("summary: got %q, want %q", diags[0].Summary(), "API Error")
	}
	if diags[0].Detail() != err.Error() {
		t.Errorf("detail: got %q, want %q", diags[0].Detail(), err.Error())
	}
}

func TestLogAndAppendAPIErrorPreservesFields(t *testing.T) {
	var diags diag.Diagnostics
	err := errors.New("timeout")
	fields := map[string]any{"project_id": "p", "vpc_id": "v"}

	LogAndAppendAPIError(context.TODO(), &diags, "Read Error", err, fields)

	// Original fields map must not be mutated.
	if _, ok := fields["error"]; ok {
		t.Error("LogAndAppendAPIError mutated the caller's fields map")
	}
	if len(diags) != 1 {
		t.Errorf("expected 1 diagnostic, got %d", len(diags))
	}
}
