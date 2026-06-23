package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// resourcesForExtraErrorTests lists resources that are missing from the
// existing TestResourceRead_API404 / TestResourceRead_API500 test lists in
// resource_read_mock_test.go. Adding them here covers their error-handling
// branches (404 → state removal, 500 → error diagnostic) without modifying
// the original table-driven tests.
var resourcesForExtraErrorTests = []struct {
	name string
	newR func() resource.Resource
}{
	{"databasegrant", NewDatabaseGrantResource},
	{"vpcpeeringroute", NewVpcPeeringRouteResource},
	{"dbaasuser", NewDBaaSUserResource},
	{"database", NewDatabaseResource},
	{"databasebackup", NewDatabaseBackupResource},
	{"securityrule", NewSecurityRuleResource},
	{"schedulejob", NewScheduleJobResource},
	{"cloudserver", NewCloudServerResource},
}

// TestResourceRead_API404_Extra verifies that the extra resources (missing from
// the original TestResourceRead_API404) correctly remove themselves from state
// (no error diagnostic) when the API returns 404.
func TestResourceRead_API404_Extra(t *testing.T) {
	ctx := context.Background()

	for _, tc := range resourcesForExtraErrorTests {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
				apiError(w, http.StatusNotFound)
			})

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceReadReq(ctx, t, res)
			res.Read(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Read() added error diagnostic on 404 (expected state removal only): %v",
					tc.name, resp.Diagnostics)
			}
			if !resp.State.Raw.IsNull() {
				t.Errorf("%s: Read() did not remove resource from state on 404", tc.name)
			}
		})
	}
}

// TestResourceRead_API500_Extra verifies that the extra resources (missing from
// the original TestResourceRead_API500) add an error diagnostic when the API
// returns 500.
func TestResourceRead_API500_Extra(t *testing.T) {
	ctx := context.Background()

	for _, tc := range resourcesForExtraErrorTests {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
				apiError(w, http.StatusInternalServerError)
			})

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceReadReq(ctx, t, res)
			res.Read(ctx, req, resp)

			if !resp.Diagnostics.HasError() {
				t.Errorf("%s: Read() returned no error diagnostic for HTTP 500 response", tc.name)
			}
		})
	}
}
