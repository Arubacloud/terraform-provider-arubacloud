package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// resourceCreateReq builds a resource.CreateRequest whose Plan has every
// string attribute set to "test-<attr-name>" so that project_id and other
// required string fields are non-empty and Create() reaches the API call.
// Non-string attributes (numbers, bools, lists, objects) are set to null.
func resourceCreateReq(ctx context.Context, t *testing.T, r resource.Resource) (resource.CreateRequest, *resource.CreateResponse) {
	t.Helper()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("resourceCreateReq: resource schema root is not an object type")
	}

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, ty := range objType.AttributeTypes {
		if ty.Is(tftypes.String) {
			attrs[name] = tftypes.NewValue(tftypes.String, "test-"+name)
		} else {
			attrs[name] = tftypes.NewValue(ty, nil)
		}
	}

	plan := tfsdk.Plan{
		Raw:    tftypes.NewValue(objType, attrs),
		Schema: schemaResp.Schema,
	}
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	return req, resp
}

// createRoutingHandler returns a mock handler that returns a minimal "Active"
// state JSON for GET requests (so SDK-internal polling for VPC/SecurityGroup
// readiness resolves after one poll cycle) and an HTTP 500 for all writes.
// Three SDK clients (Subnets, SecurityGroups, SecurityGroupRules) call
// WaitForResourceState before issuing the actual POST, so without this
// routing the test would block for the SDK's default 5-second poll interval.
func createRoutingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		//nolint:errcheck
		w.Write([]byte(`{"status":{"state":"Active"},"metadata":{"id":"test-id"}}`))
		return
	}
	apiError(w, http.StatusInternalServerError)
}

// TestResourceCreate_API500 verifies that Create() adds an error diagnostic
// when the API returns 500 for the actual write operation, for all 25 resources.
// GET requests (used by SDK-internal polling for VPC/SG readiness) are handled
// by createRoutingHandler and return "Active" so polling completes quickly.
func TestResourceCreate_API500(t *testing.T) {
	ctx := context.Background()

	for _, tc := range allResources25 {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, createRoutingHandler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceCreateReq(ctx, t, res)
			res.Create(ctx, req, resp)

			if !resp.Diagnostics.HasError() {
				t.Errorf("%s: Create() returned no error diagnostic for HTTP 500 response", tc.name)
			}
		})
	}
}
