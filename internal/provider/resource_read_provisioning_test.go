package provider

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// transitioningReadHandler builds a stateful handler that returns
// status=firstState on the first GET request and status=secondState on every
// subsequent request.  All non-GET requests return 500 (write operations are
// not expected during Read).
//
// This simulates a resource that is still provisioning when Read() is called
// (firstState = "InCreation" / "Provisioning"), transitions to the ready state
// after the provider-side WaitForResourceActive loop polls once, and then
// allows the final re-read to succeed (secondState = "Active").
func transitioningReadHandler(firstState, secondState string) http.HandlerFunc {
	var callCount int32
	firstJSON := `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"` + firstState + `"}}`
	secondJSON := `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"` + secondState + `"}}`

	return func(w http.ResponseWriter, r *http.Request) {
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
}

// TestResourceRead_RecoveryFromProvisioning verifies that Read() correctly
// handles the case where the resource is still in a transitional state when
// first polled.  The first GET returns status=InCreation; the
// WaitForResourceActive loop polls again and receives status=Active; then the
// resource is re-read and state is saved without errors.
//
// This covers the `case IsCreatingState(st)` branch inside Read() that is
// skipped in TestResourceRead_Success (which always returns Active immediately).
func TestResourceRead_RecoveryFromProvisioning(t *testing.T) {
	// Use very short poll interval so the test doesn't actually wait 5 seconds.
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	// Resources whose Read() has the IsCreatingState branch and can be tested
	// with minimal state (no null-object panics when handling the response).
	// Resources that require non-null nested objects in state (cloudserver,
	// containerregistry, dbaas) are excluded because resourceReadReq sets
	// nested objects to null — the re-read after provisioning will fail trying
	// to map the active-state response back to nested state objects.
	resources := []struct {
		name string
		newR func() resource.Resource
	}{
		{"vpc", NewVPCResource},
		{"subnet", NewSubnetResource},
		{"securitygroup", NewSecurityGroupResource},
		{"elasticip", NewElasticIPResource},
		{"keypair", NewKeypairResource},
		{"blockstorage", NewBlockStorageResource},
		{"snapshot", NewSnapshotResource},
		{"backup", NewBackupResource},
		{"restore", NewRestoreResource},
		{"kms", NewKMSResource},
		{"project", NewProjectResource},
		{"vpcpeering", NewVpcPeeringResource},
		{"vpcpeeringroute", NewVpcPeeringRouteResource},
		{"vpntunnel", NewVPNTunnelResource},
		{"vpnroute", NewVPNRouteResource},
		{"dbaasuser", NewDBaaSUserResource},
		{"database", NewDatabaseResource},
		{"databasebackup", NewDatabaseBackupResource},
		{"databasegrant", NewDatabaseGrantResource},
		{"kaas", NewKaaSResource},
		{"schedulejob", NewScheduleJobResource},
	}

	for _, tc := range resources {
		t.Run(tc.name, func(t *testing.T) {
			handler := transitioningReadHandler("InCreation", "Active")
			_, mockClient := newMockArubaClient(t, handler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceReadReq(ctx, t, res)
			res.Read(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Read() reported error during provisioning recovery: %v",
					tc.name, resp.Diagnostics)
			}
			// State should be non-null after a successful Read from Active state.
			if resp.State.Raw.IsNull() {
				t.Errorf("%s: Read() removed resource from state unexpectedly after provisioning recovery", tc.name)
			}
		})
	}
}

// TestResourceRead_FailedStateWarning verifies that Read() adds a warning
// diagnostic when the resource is in a terminal failure state, and preserves
// the resource in state (so Terraform can plan a replace or destroy).
func TestResourceRead_FailedStateWarning(t *testing.T) {
	ctx := context.Background()

	failedJSON := `{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Failed"}}`
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(failedJSON)) //nolint:errcheck
			return
		}
		apiError(w, http.StatusInternalServerError)
	}

	// Resources that expose the isFailedState branch in their Read() function.
	// Resources that do NOT check isFailedState (keypair, dbaasuser, database,
	// databasegrant) are excluded to avoid false-failure assertions.
	resources := []struct {
		name string
		newR func() resource.Resource
	}{
		{"vpc", NewVPCResource},
		{"subnet", NewSubnetResource},
		{"securitygroup", NewSecurityGroupResource},
		{"elasticip", NewElasticIPResource},
		{"blockstorage", NewBlockStorageResource},
		{"snapshot", NewSnapshotResource},
		{"backup", NewBackupResource},
		{"restore", NewRestoreResource},
		{"kms", NewKMSResource},
		{"vpcpeering", NewVpcPeeringResource},
		{"vpcpeeringroute", NewVpcPeeringRouteResource},
		{"vpntunnel", NewVPNTunnelResource},
		{"vpnroute", NewVPNRouteResource},
		{"databasebackup", NewDatabaseBackupResource},
		{"schedulejob", NewScheduleJobResource},
	}

	for _, tc := range resources {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, handler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceReadReq(ctx, t, res)
			res.Read(ctx, req, resp)

			// A "Failed" state must result in a warning, not an error.
			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Read() added error diagnostic for Failed state (expected warning only): %v",
					tc.name, resp.Diagnostics)
			}

			var hasWarning bool
			for _, d := range resp.Diagnostics {
				if d.Severity() == 2 { // Warning
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
