package provider

import (
	"context"
	"net/http"
	"testing"
)

// TestKeypairRead_WithProperties covers keypair Read() path where the API
// returns a public key value in the response.
func TestKeypairRead_WithProperties(t *testing.T) {
	ctx := context.Background()

	kpJSON := `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Active"},` +
		`"properties":{"value":"ssh-rsa AAAA... test@test"}}`
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(kpJSON)) //nolint:errcheck
			return
		}
		apiError(w, http.StatusInternalServerError)
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewKeypairResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReq(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Keypair Read() reported error with properties: %v", resp.Diagnostics)
	}
}
