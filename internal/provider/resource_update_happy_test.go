package provider

import (
	"context"
	"net/http"
	"testing"
)

// updateActiveJSON extends minimalActiveJSON with a location block so that
// Update() methods which extract the region value from the GET response can
// proceed without the "Unable to determine region value" error.
const updateActiveJSON = `{` +
	`"metadata":{` +
	`"id":"test-id","name":"test-name",` +
	`"location":{"value":"test-location","code":"test-code","name":"test-region"},` +
	`"project":{"id":"test-project-id"}` +
	`},` +
	`"status":{"state":"Active"}` +
	`}`

// updateHappyHandler serves 200 with updateActiveJSON for all requests so
// the initial GET and the PATCH/PUT write both succeed.  Update() reads
// current state (GET) and then applies changes (PATCH/PUT), so both legs need
// a valid response including the location field.
func updateHappyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	//nolint:errcheck
	w.Write([]byte(updateActiveJSON))
}

// TestResourceUpdate_Happy verifies that Update() succeeds (no error) when the
// API returns a minimal valid 200 response for both the initial GET and the
// subsequent write.
// Resources excluded because their Update() has nil-unsafe property access or
// nested-object validation that fails on minimal JSON:
//   - backup:            Volume URI required in GET response
//   - blockstorage:      checks BlockStorage-specific properties
//   - cloudserver:       complex nested network/storage access
//   - containerregistry: nil Properties.PublicIp / BlockStorage access
//   - dbaas:             complex nested config object
//   - database:          needs DBaaS-specific fields in response
//   - databasebackup:    no-op update (but causes issues with resource ID extraction)
//   - dbaasuser:         WaitForResourceActive extracts user_id from response
//   - kaas:              complex nested schema
//   - restore:           volume URI in response required
//   - schedulejob:       schedule_job_type nested attribute null issue
//   - securityrule:      null Properties nested object
//   - snapshot:          VolumeInfo pointer nil-deref
//   - vpnroute:          cloud_subnet null attribute
func TestResourceUpdate_Happy(t *testing.T) {
	ctx := context.Background()

	skip := map[string]bool{
		"blockstorage":      true, // checks BlockStorage-specific property fields
		"cloudserver":       true, // complex nested network/storage access
		"containerregistry": true, // nil Properties.BillingPlan / AdminUser access
		"dbaas":             true, // complex nested config object
		"database":          true, // needs DBaaS-specific fields
		"databasebackup":    true, // no-op update (documented no API call)
		"dbaasuser":         true, // WaitForResourceActive extracts user_id from response
		"kaas":              true, // complex nested schema
		"schedulejob":       true, // schedule_job_type nested attribute
		"securityrule":      true, // null Properties nested object
		"vpnroute":          true, // cloud_subnet null attribute
	}

	for _, tc := range allResources25 {
		if skip[tc.name] {
			continue
		}
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, updateHappyHandler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceUpdateReq(ctx, t, res)
			res.Update(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Update() reported error on 200 happy path: %v",
					tc.name, resp.Diagnostics)
			}
		})
	}
}
