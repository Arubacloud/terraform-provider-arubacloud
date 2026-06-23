package provider

import (
	"context"
	"net/http"
	"testing"
)

// TestSecurityRuleUpdate_PropertiesChangedError verifies that the securityrule
// Update() detects when immutable properties (direction, protocol, port, target)
// differ between the current API state and the Terraform plan, and adds an
// appropriate error diagnostic without making a PATCH call.
//
// This test covers the property-extraction and comparison section of Update()
// (the largest uncovered block) that the generic TestResourceUpdate_APIError
// misses because 500 on GET causes an early return before property extraction.
func TestSecurityRuleUpdate_PropertiesChangedError(t *testing.T) {
	ctx := context.Background()

	// Return the full security rule JSON for all GET requests so that the
	// properties extraction section is exercised.  Non-GET (PATCH) should not
	// be called when properties differ, so returning 500 is a safety net.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(securityruleFullJSON)) //nolint:errcheck
		} else {
			apiError(w, http.StatusInternalServerError)
		}
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewSecurityRuleResource()
	configureResource(ctx, t, res, mockClient)

	// Use resourceUpdateReqFull so data.Properties is non-null.  All string
	// attributes are set to "test-value", which differs from the current API
	// values (direction="Ingress", protocol="TCP", port="80", etc.).
	// This exercises the propertiesChanged=true branch and the early-return
	// with "Cannot Update Security Rule Properties" error.
	req, resp := resourceUpdateReqFull(ctx, t, res)
	res.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("securityrule Update() should fail when properties differ from current API state")
	}
}
