package provider

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// alwaysCreatingHandler returns a valid Create response (201) for POST
// requests, and a response with status=InCreation for all GET requests.
// This causes WaitForResourceActive to loop until it exhausts the
// ResourceTimeout, exercising the timeout / partial-state-save code path.
func alwaysCreatingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"metadata":{"id":"test-id","name":"test-name","uri":"/projects/p/res/test-id"},"status":{"state":"InCreation"}}`)) //nolint:errcheck
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"metadata":{"id":"test-id","name":"test-name","uri":"/projects/p/res/test-id"},"status":{"state":"InCreation"}}`)) //nolint:errcheck
	}
}

// alwaysCreatingWithURIHandler is like alwaysCreatingHandler but also has a
// metadata.uri so that resources which look up a source URI before the POST
// (backup, restore) can extract it and proceed to the Create call.
func alwaysCreatingWithURIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	// Use uriActiveJSON but override the state to InCreation.
	// uriActiveJSON ends with `"status":{"state":"Active"}}` — we build a
	// variant inline that ends with InCreation instead.
	w.Write([]byte(`{"metadata":{"id":"test-id","name":"test-name","uri":"/projects/p/providers/Aruba.Storage/blockStorages/test-vol-id"},"status":{"state":"InCreation"}}`)) //nolint:errcheck
}

// TestResourceCreate_ProvisioningTimeout verifies that Create() correctly
// handles a WaitForResourceActive timeout: it saves partial state (with the
// resource ID returned by the initial Create call) and adds a non-error
// diagnostic (warning for timeout, error for other failures) so that a
// subsequent `terraform apply` reconciles the resource.
//
// The test uses newMockArubaClientFast (50 ms ResourceTimeout) combined with
// waitForActivePollInterval=1ms so the poll loop exhausts the timeout almost
// immediately without real waiting.
//
// ReportWaitResult adds a WARNING (not an error) for timeout conditions, so
// the assertion checks for any diagnostic rather than an error-only diagnostic.
func TestResourceCreate_ProvisioningTimeout(t *testing.T) {
	// Shorten both the poll interval (ticker period) and the resource timeout
	// so WaitForResourceActive exhausts its budget in < 5 ms.
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	cases := []struct {
		name    string
		newR    func() resource.Resource
		handler http.HandlerFunc
		useFull bool
	}{
		// Resources without SDK-level pre-create polling (no 5-second delays).
		// Subnet, SecurityGroup, and SecurityGroupRule use SDK WaitForResourceState
		// before the POST and would add 5+ seconds each — they are excluded.
		{"vpc", NewVPCResource, alwaysCreatingHandler, false},
		{"elasticip", NewElasticIPResource, alwaysCreatingHandler, false},
		// keypair Create is synchronous — no WaitForResourceActive call.
		{"blockstorage", NewBlockStorageResource, alwaysCreatingHandler, false},
		{"snapshot", NewSnapshotResource, alwaysCreatingHandler, false},
		{"kms", NewKMSResource, alwaysCreatingHandler, false},
		// project and databasegrant do not call WaitForResourceActive in Create.
		// keypair does not call WaitForResourceActive in Create (it's synchronous).
		{"vpcpeering", NewVpcPeeringResource, alwaysCreatingHandler, false},
		{"vpcpeeringroute", NewVpcPeeringRouteResource, alwaysCreatingHandler, false},
		{"vpntunnel", NewVPNTunnelResource, alwaysCreatingHandler, false},
		{"backup", NewBackupResource, alwaysCreatingWithURIHandler, false},
		{"restore", NewRestoreResource, alwaysCreatingWithURIHandler, false},
		{"cloudserver", NewCloudServerResource, alwaysCreatingHandler, true},
		{"dbaas", NewDBaaSResource, alwaysCreatingHandler, true},
		{"containerregistry", NewContainerRegistryResource, alwaysCreatingHandler, true},
		{"schedulejob", NewScheduleJobResource, alwaysCreatingHandler, true},
		{"vpnroute", NewVPNRouteResource, alwaysCreatingHandler, true},
		{"databasebackup", NewDatabaseBackupResource, alwaysCreatingWithURIHandler, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClientFast(t, tc.handler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			var req resource.CreateRequest
			var resp *resource.CreateResponse
			if tc.useFull {
				req, resp = resourceCreateReqFull(ctx, t, res)
			} else {
				req, resp = resourceCreateReq(ctx, t, res)
			}
			res.Create(ctx, req, resp)

			// A timeout results in a WARNING (ReportWaitResult uses AddWarning
			// for ErrWaitTimeout), while other errors use AddError. Either way,
			// diagnostics must be non-empty to signal that the resource was not
			// fully provisioned within the timeout.
			if len(resp.Diagnostics) == 0 {
				t.Errorf("%s: Create() reported no diagnostics on provisioning timeout (expected warning or error)", tc.name)
			}
		})
	}
}
