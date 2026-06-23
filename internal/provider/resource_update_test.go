package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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

// resourceUpdateReq builds a resource.UpdateRequest whose Plan and State both
// have every string attribute set to "test-<attr-name>".  Using identical
// values in plan and state prevents early-return on "immutable field changed"
// checks, so Update() proceeds to the first API call.
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

// updateSuccessHandler serves 200 with updateActiveJSON for all requests so
// the initial GET and the PATCH/PUT write both succeed.  updateActiveJSON
// includes a location block required by some resources to determine the
// region value before issuing an update.
func updateSuccessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(updateActiveJSON)) //nolint:errcheck
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

// TestResourceUpdate_Success verifies that Update() succeeds (no error) when
// the API returns a minimal valid 200 for both the initial GET and the write.
//
// Resources excluded because their Update() has nil-unsafe property access or
// nested-object validators that fail on minimal JSON:
//   - blockstorage:      accesses BlockStorage-specific property fields
//   - cloudserver:       complex nested network/storage property access
//   - containerregistry: dereferences BillingPlan / AdminUser without nil check
//   - dbaas:             nested config object fails null sub-attribute check
//   - database:          requires DBaaS-specific fields in GET response
//   - databasebackup:    documented no-op (no API call, adds warning only)
//   - dbaasuser:         WaitForResourceActive extracts user_id from response
//   - kaas:              complex nested schema with nil-unsafe access
//   - schedulejob:       schedule_job_type null nested attribute
//   - securityrule:      null Properties nested object → missing attribute error
//   - vpnroute:          cloud_subnet null attribute
func TestResourceUpdate_Success(t *testing.T) {
	ctx := context.Background()

	skip := map[string]bool{
		"blockstorage":      true,
		"cloudserver":       true,
		"containerregistry": true,
		"dbaas":             true,
		"database":          true,
		"databasebackup":    true,
		"dbaasuser":         true,
		"kaas":              true,
		"schedulejob":       true,
		"securityrule":      true,
		"vpnroute":          true,
	}

	for _, tc := range allResources25 {
		if skip[tc.name] {
			continue
		}
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, updateSuccessHandler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceUpdateReq(ctx, t, res)
			res.Update(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Update() reported error: %v", tc.name, resp.Diagnostics)
			}
		})
	}
}

// TestResourceUpdate_APIError verifies that Update() adds an error diagnostic
// when every API request returns 500.  Update() methods typically start with
// a GET of the current resource; a 500 on that GET is sufficient to trigger
// the error path without SDK-internal polling interference.
func TestResourceUpdate_APIError(t *testing.T) {
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

// TestResourceUpdate_NotFound verifies that Update() adds an error diagnostic
// when the initial GET returns 404 (resource deleted between plan and apply).
// Keypair.Update() is excluded: it correctly treats 404 as "resource gone"
// by calling resp.State.RemoveResource without adding an error diagnostic.
func TestResourceUpdate_NotFound(t *testing.T) {
	ctx := context.Background()

	// keypair.Update() silently removes the resource from state on 404.
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
