package provider

import (
	"context"
	"net/http"
	"testing"
)

// TestBackupRead_WithProperties covers backup Read() with a properties block
// that includes retentionDays and billingPeriod to cover those pointer branches.
func TestBackupRead_WithProperties(t *testing.T) {
	ctx := context.Background()

	backupJSON := `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Active"},` +
		`"properties":{"retentionDays":7,"billingPeriod":"Month","origin":{"uri":"/volumes/test-vol"}}}`
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(backupJSON)) //nolint:errcheck
			return
		}
		apiError(w, http.StatusInternalServerError)
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewBackupResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReq(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Backup Read() reported error with properties: %v", resp.Diagnostics)
	}
}
