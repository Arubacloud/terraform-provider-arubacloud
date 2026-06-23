package provider

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// resourceDeleteReq builds a resource.DeleteRequest whose State has every
// string attribute set to "test-<attr-name>" so that id / project_id / etc.
// are non-empty and Delete() proceeds past the empty-ID guard to the API call.
func resourceDeleteReq(ctx context.Context, t *testing.T, r resource.Resource) (resource.DeleteRequest, *resource.DeleteResponse) {
	t.Helper()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("resourceDeleteReq: resource schema root is not an object type")
	}

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, ty := range objType.AttributeTypes {
		if ty.Is(tftypes.String) {
			attrs[name] = tftypes.NewValue(tftypes.String, "test-"+name)
		} else {
			attrs[name] = tftypes.NewValue(ty, nil)
		}
	}

	state := tfsdk.State{
		Raw:    tftypes.NewValue(objType, attrs),
		Schema: schemaResp.Schema,
	}
	return resource.DeleteRequest{State: state}, &resource.DeleteResponse{}
}

// allResources25 is the canonical list of all 25 resources under test.
// Tests that need a subset maintain their own skip maps or slices.
var allResources25 = []struct {
	name string
	newR func() resource.Resource
}{
	{"vpc", NewVPCResource},
	{"subnet", NewSubnetResource},
	{"securitygroup", NewSecurityGroupResource},
	{"securityrule", NewSecurityRuleResource},
	{"elasticip", NewElasticIPResource},
	{"keypair", NewKeypairResource},
	{"blockstorage", NewBlockStorageResource},
	{"snapshot", NewSnapshotResource},
	{"backup", NewBackupResource},
	{"restore", NewRestoreResource},
	{"kms", NewKMSResource},
	{"project", NewProjectResource},
	{"cloudserver", NewCloudServerResource},
	{"vpcpeering", NewVpcPeeringResource},
	{"vpcpeeringroute", NewVpcPeeringRouteResource},
	{"vpntunnel", NewVPNTunnelResource},
	{"vpnroute", NewVPNRouteResource},
	{"dbaas", NewDBaaSResource},
	{"dbaasuser", NewDBaaSUserResource},
	{"database", NewDatabaseResource},
	{"databasebackup", NewDatabaseBackupResource},
	{"databasegrant", NewDatabaseGrantResource},
	{"kaas", NewKaaSResource},
	{"containerregistry", NewContainerRegistryResource},
	{"schedulejob", NewScheduleJobResource},
}

// TestResourceDelete_Success verifies that Delete() succeeds (no error
// diagnostic) when the API returns 204 for DELETE and 404 for the polling
// GET (resource confirmed gone immediately on first deletion check).
func TestResourceDelete_Success(t *testing.T) {
	ctx := context.Background()

	for _, tc := range allResources25 {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodDelete {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				// GET requests from deletionChecker and WaitForResourceDeleted
				apiError(w, http.StatusNotFound)
			})

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceDeleteReq(ctx, t, res)
			res.Delete(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Delete() reported error on 204/404 path: %v",
					tc.name, resp.Diagnostics)
			}
		})
	}
}

// TestResourceDelete_APIError verifies that Delete() adds an error diagnostic
// when the API consistently returns 500, causing the retry loop to time out.
// deleteRetryBaseWait is set to 1ms and ResourceTimeout to 50ms so the loop
// exits quickly without slowing the test suite.
func TestResourceDelete_APIError(t *testing.T) {
	oldBaseWait := deleteRetryBaseWait
	deleteRetryBaseWait = 1 * time.Millisecond
	t.Cleanup(func() { deleteRetryBaseWait = oldBaseWait })

	for _, tc := range allResources25 {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			_, mockClient := newMockArubaClientFast(t, func(w http.ResponseWriter, r *http.Request) {
				apiError(w, http.StatusInternalServerError)
			})

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceDeleteReq(ctx, t, res)
			res.Delete(ctx, req, resp)

			if !resp.Diagnostics.HasError() {
				t.Errorf("%s: Delete() returned no error for persistent 500 response", tc.name)
			}
		})
	}
}
