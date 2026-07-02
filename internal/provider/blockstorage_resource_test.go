package provider

import (
	"context"
	"net/http"
	"testing"
)

// TestBlockStorageRead_WithProperties covers blockstorage Read() with a
// properties block that includes bootable=true and a non-empty image field.
func TestBlockStorageRead_WithProperties(t *testing.T) {
	ctx := context.Background()

	bsJSON := `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Active"},` +
		`"properties":{"sizeGB":50,"billingPeriod":"Hour","type":"Standard","zone":"test-zone","bootable":true,"image":"test-image"}}`
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(bsJSON)) //nolint:errcheck
			return
		}
		apiError(w, http.StatusInternalServerError)
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewBlockStorageResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReq(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("BlockStorage Read() reported error with properties: %v", resp.Diagnostics)
	}
}
