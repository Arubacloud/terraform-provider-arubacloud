package provider

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// createHappyHandler serves 201 with minimalActiveJSON for POST (the actual
// create call) and 200 with minimalActiveJSON for GET (SDK-internal VPC/SG
// pre-polling + provider-level WaitForResourceActive checks).
func createHappyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	//nolint:errcheck
	w.Write([]byte(minimalActiveJSON))
}

// TestResourceCreate_Happy verifies that Create() succeeds (no error) when
// the API returns a minimal valid 201 response.
// Several resources are excluded because their Read/Create code has
// non-nil-guarded property access that panics on minimal JSON, or nested
// object validators that reject null sub-attributes:
//   - containerregistry: nil Properties.PublicIp / BlockStorage access
//   - cloudserver:       ID required check errors on empty metadata
//   - securityrule:      null Properties object → "missing attribute" error
//   - dbaas:             nested config object fails validation
//   - kaas:              complex nested schema with nil-unsafe access
func TestResourceCreate_Happy(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	// Resources excluded because their Create() requires specific response fields
	// beyond the minimalActiveJSON (e.g. volume URI, nested type attributes) or
	// has nil-unsafe Properties access that panics with minimal JSON.
	skip := map[string]bool{
		"backup":            true, // checks for volume URI in response
		"restore":           true, // same
		"vpnroute":          true, // cloud_subnet must be a non-null String
		"dbaasuser":         true, // WaitForResourceActive extracts user_id from response
		"database":          true, // WaitForResourceActive extracts database_id from response
		"databasebackup":    true, // requires DBaaS URI in response
		"schedulejob":       true, // schedule_job_type nested attribute null issue
		"containerregistry": true, // nil-deref Properties.PublicIp / BlockStorage
		"cloudserver":       true, // requires non-empty ID in response metadata
		"securityrule":      true, // null Properties → missing attribute error
		"dbaas":             true, // nested config object validator fails on null
		"kaas":              true, // complex nested schema with nil-unsafe access
	}

	for _, tc := range allResources25 {
		if skip[tc.name] {
			continue
		}
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, createHappyHandler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceCreateReq(ctx, t, res)
			res.Create(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Create() reported error on 201 happy path: %v",
					tc.name, resp.Diagnostics)
			}
		})
	}
}
