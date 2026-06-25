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

// uriActiveJSON extends minimalActiveJSON with a metadata.uri so that
// Create() methods which GET a source resource before the actual POST
// (backup, restore do a volume lookup) can extract the URI and proceed.
const uriActiveJSON = `{` +
	`"metadata":{` +
	`"id":"test-id","name":"test-name",` +
	`"uri":"/projects/p/providers/Aruba.Storage/blockStorages/test-vol-id"` +
	`},"status":{"state":"Active"}` +
	`}`

// resourceCreateReq builds a resource.CreateRequest whose Plan has every
// string attribute set to "test-<attr-name>" so that required string fields
// (project_id, vpc_id, etc.) are non-empty and Create() reaches the API call.
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

// createSuccessHandler serves 201 with minimalActiveJSON for POST (the actual
// create call) and 200 with minimalActiveJSON for GET (SDK-internal VPC/SG
// pre-polling + provider-level WaitForResourceActive checks).
func createSuccessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write([]byte(minimalActiveJSON)) //nolint:errcheck
}

// createSuccessWithURIHandler is a variant of createSuccessHandler that
// returns uriActiveJSON (which includes a metadata.uri) so that resources
// which GET a source volume before the actual POST (backup, restore) can
// extract the URI and proceed past the "URI not found in response" guard.
func createSuccessWithURIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write([]byte(uriActiveJSON)) //nolint:errcheck
}

// createAPIErrorHandler routes GET requests to return "Active" state JSON
// (so SDK-internal pre-create polling resolves after one cycle) and returns
// HTTP 500 for all write operations.  Subnet, SecurityGroup, and
// SecurityGroupRule SDK clients call WaitForResourceState before the actual
// POST; without this routing the test would block for the SDK's default
// 5-second poll interval.
func createAPIErrorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(minimalActiveJSON)) //nolint:errcheck
		return
	}
	apiError(w, http.StatusInternalServerError)
}

// TestResourceCreate_Success verifies that Create() succeeds (no error) when
// the API returns a valid 201 response.
//
// Resources excluded because their Create() requires response fields beyond
// minimalActiveJSON, or has nil-unsafe Properties access that panics on minimal
// JSON:
//   - containerregistry: nil Properties.PublicIp / BlockStorage access
//   - cloudserver:       requires non-empty ID in response metadata
//   - securityrule:      null Properties object → "missing attribute" error
//   - dbaas:             nested config object fails validation
//   - kaas:              complex nested schema with nil-unsafe access
//   - vpnroute:          cloud_subnet must be a non-null String
//   - dbaasuser:         WaitForResourceActive extracts user_id from response
//   - database:          WaitForResourceActive extracts database_id from response
//   - databasebackup:    requires DBaaS URI in response
//   - schedulejob:       schedule_job_type nested attribute null issue
func TestResourceCreate_Success(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	cases := []struct {
		name    string
		newR    func() resource.Resource
		handler http.HandlerFunc
	}{
		{"vpc", NewVPCResource, createSuccessHandler},
		{"subnet", NewSubnetResource, createSuccessHandler},
		{"securitygroup", NewSecurityGroupResource, createSuccessHandler},
		{"elasticip", NewElasticIPResource, createSuccessHandler},
		{"keypair", NewKeypairResource, createSuccessHandler},
		{"blockstorage", NewBlockStorageResource, createSuccessHandler},
		{"snapshot", NewSnapshotResource, createSuccessHandler},
		{"kms", NewKMSResource, createSuccessHandler},
		{"project", NewProjectResource, createSuccessHandler},
		{"vpcpeering", NewVpcPeeringResource, createSuccessHandler},
		{"vpcpeeringroute", NewVpcPeeringRouteResource, createSuccessHandler},
		{"vpntunnel", NewVPNTunnelResource, createSuccessHandler},
		{"databasegrant", NewDatabaseGrantResource, createSuccessHandler},
		// backup and restore do a source-volume GET before the actual POST
		// and require metadata.uri in the response body.
		{"backup", NewBackupResource, createSuccessWithURIHandler},
		{"restore", NewRestoreResource, createSuccessWithURIHandler},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, tc.handler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceCreateReq(ctx, t, res)
			res.Create(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Create() reported error: %v", tc.name, resp.Diagnostics)
			}
		})
	}
}

// TestResourceCreate_APIError verifies that Create() adds an error diagnostic
// when the API returns 500 for the actual write operation, for all 25 resources.
// GET requests (SDK-internal VPC/SG pre-create polling) are answered with
// "Active" JSON by createAPIErrorHandler so polling exits after one cycle.
func TestResourceCreate_APIError(t *testing.T) {
	ctx := context.Background()

	for _, tc := range allResources25 {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, createAPIErrorHandler)

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
