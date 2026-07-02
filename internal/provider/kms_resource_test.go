package provider

import (
	"context"
	"net/http"
	"testing"
)

// TestKMSRead_WithBillingPeriod covers the kms.Properties.BillingPeriod branch
// that is only reached when the properties block in the API response has a
// non-empty billingPeriod field.
func TestKMSRead_WithBillingPeriod(t *testing.T) {
	ctx := context.Background()

	kmsJSON := `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Active"},"properties":{"billingPeriod":"Hour"}}`
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(kmsJSON)) //nolint:errcheck
			return
		}
		apiError(w, http.StatusInternalServerError)
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewKMSResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReq(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("KMS Read() reported error with billingPeriod: %v", resp.Diagnostics)
	}
}
