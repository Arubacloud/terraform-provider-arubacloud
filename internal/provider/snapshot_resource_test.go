package provider

import (
	"context"
	"net/http"
	"testing"
)

// TestSnapshotRead_WithProperties covers snapshot Read() with properties so
// the volume-URI branch and billingPeriod branch are exercised.
func TestSnapshotRead_WithProperties(t *testing.T) {
	ctx := context.Background()

	snapJSON := `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Active"},` +
		`"properties":{"volume":{"uri":"/test/vol-id"},"billingPeriod":"Hour"}}`
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(snapJSON)) //nolint:errcheck
			return
		}
		apiError(w, http.StatusInternalServerError)
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewSnapshotResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReq(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Snapshot Read() reported error with properties: %v", resp.Diagnostics)
	}
}
