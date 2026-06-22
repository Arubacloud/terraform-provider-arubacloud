package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
)

// newMockArubaClient spins up a single httptest.Server that serves both the
// OAuth2 token endpoint (POST /token) and the Aruba API.  apiHandler is
// invoked for every request whose path is not "/token".
//
// The returned ArubaCloudClient is pre-configured with the mock SDK client so
// it can be injected directly into any resource / datasource via Configure().
func newMockArubaClient(t *testing.T, apiHandler http.HandlerFunc) (*httptest.Server, *ArubaCloudClient) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "mock-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
			return
		}
		apiHandler(w, r)
	}))
	t.Cleanup(srv.Close)

	opts := aruba.DefaultOptions("test-key", "test-secret").
		WithBaseURL(srv.URL).
		WithTokenIssuerURL(srv.URL + "/token")

	sdkClient, err := aruba.NewClient(opts)
	if err != nil {
		t.Fatalf("newMockArubaClient: failed to create SDK client: %v", err)
	}

	return srv, &ArubaCloudClient{
		ApiKey:          "test-key",
		ApiSecret:       "test-secret",
		Client:          sdkClient,
		ResourceTimeout: 10 * time.Minute,
	}
}

// apiError writes an RFC-7807 problem-details JSON body with the given HTTP
// status code.  Pass statusCode 404 or 500 to exercise the two most common
// API error branches in Read() methods.
func apiError(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"title":  http.StatusText(statusCode),
		"status": statusCode,
	})
}
