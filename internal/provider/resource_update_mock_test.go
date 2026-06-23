package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// resourceUpdateReq builds a resource.UpdateRequest whose Plan and State both
// have every string attribute set to "test-<attr-name>".  Using the same
// values in plan and state means Update() cannot early-return on "immutable
// field changed" checks; it will proceed to the first API call (GET or
// PUT/PATCH), where the mock server's 500 triggers an error diagnostic.
func resourceUpdateReq(ctx context.Context, t *testing.T, r resource.Resource) (resource.UpdateRequest, *resource.UpdateResponse) {
	t.Helper()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("resourceUpdateReq: resource schema root is not an object type")
	}

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, ty := range objType.AttributeTypes {
		if ty.Is(tftypes.String) {
			attrs[name] = tftypes.NewValue(tftypes.String, "test-"+name)
		} else {
			attrs[name] = tftypes.NewValue(ty, nil)
		}
	}

	tfVal := tftypes.NewValue(objType, attrs)
	req := resource.UpdateRequest{
		Plan:  tfsdk.Plan{Raw: tfVal, Schema: schemaResp.Schema},
		State: tfsdk.State{Raw: tfVal, Schema: schemaResp.Schema},
	}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	return req, resp
}

// resourcesWithAPIUpdate is the subset of allResources25 whose Update()
// makes at least one SDK API call.  DatabaseBackup.Update() is a documented
// no-op (adds a warning and saves state unchanged) so it is excluded from
// error-path assertions that expect an error diagnostic.
var resourcesWithAPIUpdate = []struct {
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
	{"databasegrant", NewDatabaseGrantResource},
	{"kaas", NewKaaSResource},
	{"containerregistry", NewContainerRegistryResource},
	{"schedulejob", NewScheduleJobResource},
}

// TestResourceUpdate_API500 verifies that Update() adds an error diagnostic
// when every API request returns 500.  Resources whose Update() makes at
// least one SDK call are exercised (databasebackup is excluded — its Update
// is a documented no-op that adds only a warning).
func TestResourceUpdate_API500(t *testing.T) {
	ctx := context.Background()

	for _, tc := range resourcesWithAPIUpdate {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
				apiError(w, http.StatusInternalServerError)
			})

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceUpdateReq(ctx, t, res)
			res.Update(ctx, req, resp)

			if !resp.Diagnostics.HasError() {
				t.Errorf("%s: Update() returned no error diagnostic for HTTP 500 response", tc.name)
			}
		})
	}
}

// TestResourceUpdate_API404 verifies that Update() adds an error diagnostic
// when the initial GET returns 404 (resource not found mid-update).
// Keypair.Update() is excluded because it correctly handles 404 as
// "resource gone" by removing it from state without an error diagnostic.
// Databasebackup is excluded because its Update() makes no API call.
func TestResourceUpdate_API404(t *testing.T) {
	ctx := context.Background()

	// Keypair.Update() treats 404 as "resource gone" → RemoveResource, no error.
	skip := map[string]bool{"keypair": true}

	for _, tc := range resourcesWithAPIUpdate {
		if skip[tc.name] {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
				apiError(w, http.StatusNotFound)
			})

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceUpdateReq(ctx, t, res)
			res.Update(ctx, req, resp)

			if !resp.Diagnostics.HasError() {
				t.Errorf("%s: Update() returned no error diagnostic for HTTP 404 response", tc.name)
			}
		})
	}
}
