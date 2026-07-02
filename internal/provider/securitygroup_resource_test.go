package provider

import (
	"context"
	"net/http"
	"testing"
)

// TestSecurityGroupRead_WithProperties covers securitygroup Read() with a
// response that includes a properties block.
func TestSecurityGroupRead_WithProperties(t *testing.T) {
	ctx := context.Background()

	sgJSON := `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Active"},"properties":{"description":"test-sg","rulesCount":2}}`
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(sgJSON)) //nolint:errcheck
			return
		}
		apiError(w, http.StatusInternalServerError)
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewSecurityGroupResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReq(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("SecurityGroup Read() reported error with properties: %v", resp.Diagnostics)
	}
}
