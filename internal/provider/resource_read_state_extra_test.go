package provider

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// cloudserverFailedJSON is a CloudServer API response with state=Failed that
// also includes a full properties block.  This allows cloudserver's Read()
// to add only the warning (not additional errors from nil-property access) when
// the resource is in a terminal failure state.
const cloudserverFailedJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name",` +
	`"location":{"value":"test-location","code":"test-code","name":"test-region"},` +
	`"project":{"id":"test-project-id"}},` +
	`"status":{"state":"Failed"},` +
	`"properties":{` +
	`"vpc":{"uri":"test-vpc-uri"},` +
	`"bootVolume":{"uri":"test-boot-uri"},` +
	`"flavor":{"name":"test-flavor"},` +
	`"keyPair":{"uri":""},` +
	`"zone":"test-zone"` +
	`}}`

// dbaasFailedJSON is a DBaaS API response with state=Failed that includes
// the properties block so Read() can continue past the properties-mapping
// section without nil-pointer errors.
const dbaasFailedJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name",` +
	`"location":{"value":"test-location","code":"test-code","name":"test-region"},` +
	`"project":{"id":"test-project-id"}},` +
	`"status":{"state":"Failed"},` +
	`"properties":{` +
	`"engine":{"id":"postgres-14"},` +
	`"billingPlan":{"billingPeriod":"Hour"}` +
	`}}`

// TestResourceRead_FailedState_Extra verifies that complex resources (missing
// from the original TestResourceRead_FailedStateWarning because they require
// a full state or rich response JSON) correctly emit a Warning diagnostic and
// keep the resource in state when they encounter a terminal failure status.
func TestResourceRead_FailedState_Extra(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name    string
		newR    func() resource.Resource
		json    string
		useFull bool
	}{
		// kaas: the minimal failedJSON is safe because the properties-mapping
		// code inside kaas Read is nil-guarded throughout.
		{"kaas", NewKaaSResource, `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Failed"}}`, true},
		// cloudserver: needs full state (CloudServerNetworkModel can't hold null)
		// and a properties block so the post-warning mapping doesn't nil-panic.
		{"cloudserver", NewCloudServerResource, cloudserverFailedJSON, true},
		// dbaas: needs full state and a properties block.
		{"dbaas", NewDBaaSResource, dbaasFailedJSON, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := tc.json
			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(body)) //nolint:errcheck
					return
				}
				apiError(w, http.StatusInternalServerError)
			}

			_, mockClient := newMockArubaClient(t, handler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			var req resource.ReadRequest
			var resp *resource.ReadResponse
			if tc.useFull {
				req, resp = resourceReadReqFull(ctx, t, res)
			} else {
				req, resp = resourceReadReq(ctx, t, res)
			}
			res.Read(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Read() added error diagnostic for Failed state (expected warning only): %v",
					tc.name, resp.Diagnostics)
			}

			var hasWarning bool
			for _, d := range resp.Diagnostics {
				if d.Severity() == 2 {
					hasWarning = true
					break
				}
			}
			if !hasWarning {
				t.Errorf("%s: Read() did not add warning diagnostic for resource in Failed state", tc.name)
			}
		})
	}
}

// TestResourceRead_Provisioning_Extra verifies that extra resources (missing
// from TestResourceRead_RecoveryFromProvisioning) correctly transition from
// InCreation to Active during Read().
func TestResourceRead_Provisioning_Extra(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	cases := []struct {
		name string
		newR func() resource.Resource
	}{
		{"securityrule", NewSecurityRuleResource},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var callCount int32
			firstJSON := `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"InCreation"}}`
			// Return securityruleFullJSON for the second call so properties mapping succeeds.
			secondJSON := securityruleFullJSON

			handler := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					apiError(w, http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				if atomic.AddInt32(&callCount, 1) <= 1 {
					w.Write([]byte(firstJSON)) //nolint:errcheck
				} else {
					w.Write([]byte(secondJSON)) //nolint:errcheck
				}
			}

			_, mockClient := newMockArubaClient(t, handler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceReadReq(ctx, t, res)
			res.Read(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Read() reported error during provisioning recovery: %v",
					tc.name, resp.Diagnostics)
			}
			if resp.State.Raw.IsNull() {
				t.Errorf("%s: Read() removed resource from state unexpectedly after provisioning recovery",
					tc.name)
			}
		})
	}
}
