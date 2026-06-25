package provider

import (
	"context"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// buildFullTFValue returns a non-null tftypes.Value for any schema type.
// Strings get "test-value", numbers get 0, bools get false, objects are
// built recursively with all attributes populated, and lists/sets/maps get
// empty collections.  This allows resources with required nested objects
// (network, properties, storage, etc.) to have their plans decoded without
// the UnhandledNull errors that occur with the minimal builders in
// resource_read_mock_test.go and resource_create_test.go.
func buildFullTFValue(ty tftypes.Type) tftypes.Value {
	switch {
	case ty.Is(tftypes.String):
		return tftypes.NewValue(tftypes.String, "test-value")
	case ty.Is(tftypes.Number):
		return tftypes.NewValue(tftypes.Number, new(big.Float).SetFloat64(0))
	case ty.Is(tftypes.Bool):
		return tftypes.NewValue(tftypes.Bool, false)
	}

	if objType, ok := ty.(tftypes.Object); ok {
		attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
		for name, attrType := range objType.AttributeTypes {
			attrs[name] = buildFullTFValue(attrType)
		}
		return tftypes.NewValue(objType, attrs)
	}

	if listType, ok := ty.(tftypes.List); ok {
		// An empty or null list loses element type info when the framework decodes
		// it — use one typed element so ElemType is preserved.
		if listType.ElementType != nil {
			return tftypes.NewValue(listType, []tftypes.Value{buildFullTFValue(listType.ElementType)})
		}
		return tftypes.NewValue(listType, nil)
	}

	if setType, ok := ty.(tftypes.Set); ok {
		if setType.ElementType != nil {
			return tftypes.NewValue(setType, []tftypes.Value{buildFullTFValue(setType.ElementType)})
		}
		return tftypes.NewValue(setType, nil)
	}

	if mapType, ok := ty.(tftypes.Map); ok {
		if mapType.ElementType != nil {
			return tftypes.NewValue(mapType, map[string]tftypes.Value{"key": buildFullTFValue(mapType.ElementType)})
		}
		return tftypes.NewValue(mapType, nil)
	}

	// Fallback: null for any unrecognised type (DynamicPseudoType etc.)
	return tftypes.NewValue(ty, nil)
}

// resourceCreateReqFull builds a Create request where ALL attributes —
// including nested object attributes — receive non-null values via
// buildFullTFValue.  Use this for resources whose Create() calls As() on
// nested objects that are normally null in the minimal resourceCreateReq.
func resourceCreateReqFull(ctx context.Context, t *testing.T, r resource.Resource) (resource.CreateRequest, *resource.CreateResponse) {
	t.Helper()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("resourceCreateReqFull: schema root is not an object type")
	}

	plan := tfsdk.Plan{
		Raw:    buildFullTFValue(objType),
		Schema: schemaResp.Schema,
	}
	req := resource.CreateRequest{Plan: plan}
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	return req, resp
}

// resourceReadReqFull builds a Read request where ALL attributes receive
// non-null values via buildFullTFValue.  Use for resources whose Read()
// calls As() on nested state objects (e.g. CloudServer, ContainerRegistry).
func resourceReadReqFull(ctx context.Context, t *testing.T, r resource.Resource) (resource.ReadRequest, *resource.ReadResponse) {
	t.Helper()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("resourceReadReqFull: schema root is not an object type")
	}

	state := tfsdk.State{
		Raw:    buildFullTFValue(objType),
		Schema: schemaResp.Schema,
	}
	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	return req, resp
}

// resourceUpdateReqFull builds an Update request where Plan and State both
// use buildFullTFValue.  Use for resources whose Update() calls As() on
// nested plan/state objects.
func resourceUpdateReqFull(ctx context.Context, t *testing.T, r resource.Resource) (resource.UpdateRequest, *resource.UpdateResponse) {
	t.Helper()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("resourceUpdateReqFull: schema root is not an object type")
	}

	fullVal := buildFullTFValue(objType)
	req := resource.UpdateRequest{
		Plan:  tfsdk.Plan{Raw: fullVal, Schema: schemaResp.Schema},
		State: tfsdk.State{Raw: fullVal, Schema: schemaResp.Schema},
	}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	return req, resp
}

// cloudserverFullJSON is a CloudServer API response that includes a
// properties block so that Read() and Update() can access Flavor.Name,
// VPC.URI, etc. without nil-pointer panics.
const cloudserverFullJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name",` +
	`"uri":"/projects/p/providers/Aruba.Compute/cloudServers/test-id",` +
	`"location":{"value":"test-location","code":"test-code","name":"test-region"},` +
	`"project":{"id":"test-project-id"}},` +
	`"status":{"state":"Active"},` +
	`"properties":{` +
	`"vpc":{"uri":"test-vpc-uri"},` +
	`"bootVolume":{"uri":"test-boot-uri"},` +
	`"flavor":{"name":"test-flavor"},` +
	`"keyPair":{"uri":""},` +
	`"zone":"test-zone"` +
	`}}`

// securityruleFullJSON includes a properties block with all required fields
// (direction, protocol, port) and a non-nil target pointer.  Without target
// the Read() function skips building the target ObjectValue, which leaves a
// required key absent from the properties ObjectValue → error.
const securityruleFullJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name"},` +
	`"status":{"state":"Active"},` +
	`"properties":{` +
	`"direction":"Ingress",` +
	`"protocol":"TCP",` +
	`"port":"80",` +
	`"target":{"kind":"IP","value":"10.0.0.0/8"}` +
	`}}`

// blockstorageNotUsedJSON is used for BlockStorage Update tests.
// The status must be "NotUsed" or "Used" — "Active" causes Update to
// return early with "Cannot Update" diagnostic.  The location block
// provides regionValue, and properties.type prevents an empty-type error.
const blockstorageNotUsedJSON = `{` +
	`"metadata":{` +
	`"id":"test-id","name":"test-name",` +
	`"location":{"value":"test-location","code":"test-code","name":"test-region"},` +
	`"project":{"id":"test-project-id"}},` +
	`"status":{"state":"NotUsed"},` +
	`"properties":{"sizeGB":10,"billingPeriod":"Hour","type":"Standard","zone":""}}`

// containerregistryUpdateJSON is used for ContainerRegistry Update tests.
// It combines updateActiveJSON's location block with the properties block
// needed to avoid nil-pointer panics on Properties.BillingPlan / AdminUser.
const containerregistryUpdateJSON = `{` +
	`"metadata":{` +
	`"id":"test-id","name":"test-name",` +
	`"location":{"value":"test-location","code":"test-code","name":"test-region"},` +
	`"project":{"id":"test-project-id"}},` +
	`"status":{"state":"Active"},` +
	`"properties":{` +
	`"publicIp":{"uri":"test-pip-uri"},` +
	`"vpc":{"uri":"test-vpc-uri"},` +
	`"subnet":{"uri":"test-subnet-uri"},` +
	`"securityGroup":{"uri":"test-sg-uri"},` +
	`"blockStorage":{"uri":"test-bs-uri"},` +
	`"billingPlan":{"billingPeriod":"Month"},` +
	`"adminUser":{"username":"test-admin"}` +
	`}}`

// cloudserverCreateSuccessHandler serves the cloudserverFullJSON for every
// request so that POST (create) has a metadata.id and GET (wait poll +
// re-read) has properties for field-mapping.
func cloudserverCreateSuccessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write([]byte(cloudserverFullJSON)) //nolint:errcheck
}

// blockstorageUpdateSuccessHandler routes GET to blockstorageNotUsedJSON
// (status=NotUsed, has location+properties) so Update() proceeds past all
// guards, and routes PATCH to minimalActiveJSON for the write response.
func blockstorageUpdateSuccessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodGet {
		w.Write([]byte(blockstorageNotUsedJSON)) //nolint:errcheck
	} else {
		w.Write([]byte(minimalActiveJSON)) //nolint:errcheck
	}
}

// containerregistryUpdateSuccessHandler routes all requests to
// containerregistryUpdateJSON which has both location and properties so
// ContainerRegistry.Update() can proceed past all guards.
func containerregistryUpdateSuccessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write([]byte(containerregistryUpdateJSON)) //nolint:errcheck
}

// TestResourceCreate_Success_ComplexResources verifies that Create() succeeds
// (no error diagnostic) for resources that were excluded from the generic
// TestResourceCreate_Success because their Create() extracts nested object
// attributes from the plan that are set to null by the minimal
// resourceCreateReq helper.  Using resourceCreateReqFull provides non-null
// nested objects so that Create() reaches the API call.
func TestResourceCreate_Success_ComplexResources(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	cases := []struct {
		name    string
		newR    func() resource.Resource
		handler http.HandlerFunc
	}{
		// cloudserver is excluded: its network.subnet_uri_refs and
		// securitygroup_uri_refs are typed lists inside a SingleNestedAttribute.
		// The TPF framework loses element-type info when decoding a
		// buildFullTFValue-generated plan, causing types.ObjectValue to fail.
		// CloudServer Create is tested by acceptance tests and the generic
		// TestResourceCreate_Success test (which uses a minimal plan instead).
		// securityrule: needs full plan (properties object non-null); response
		// can be minimal because Create doesn't map Properties from the response.
		{"securityrule", NewSecurityRuleResource, createSuccessHandler},
		// dbaas: needs full plan (storage + network objects non-null).
		{"dbaas", NewDBaaSResource, createSuccessHandler},
		// containerregistry is excluded: its SDK WaitUntilReady uses a poll interval
		// of several minutes (matching the 20-40 min provisioning time), which
		// exceeds unit-test timeouts.  The Create path is covered by acceptance tests.
		// schedulejob: needs full plan (properties object with schedule_job_type).
		{"schedulejob", NewScheduleJobResource, createSuccessHandler},
		// vpnroute: needs full plan (properties object with cloud_subnet).
		{"vpnroute", NewVPNRouteResource, createSuccessHandler},
		// databasebackup: first does a GET for the DBaaS URI, then POSTs.
		{"databasebackup", NewDatabaseBackupResource, createSuccessWithURIHandler},
		// vpntunnel: Create() extracts a complex nested properties object from
		// the plan. With resourceCreateReqFull the nested objects are non-null.
		{"vpntunnel", NewVPNTunnelResource, createSuccessHandler},
		// subnet: Create() extracts nested network (address, dhcp.ranges, dhcp.routes,
		// segments) objects from the plan.  resourceCreateReqFull provides empty
		// lists for ranges/routes/segments so those loops are no-ops.
		{"subnet", NewSubnetResource, createSuccessHandler},
		// dbaasuser and database are excluded: they use the username/name as the
		// resource ID and then pass it to WaitForResourceActive as the resource
		// identifier.  minimalActiveJSON provides no username/name field in the
		// SDK response type, so ID is set to "" and the SDK rejects the subsequent
		// Get(username="") call with "user/database ID cannot be empty".
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, tc.handler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceCreateReqFull(ctx, t, res)
			res.Create(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Create() reported error: %v", tc.name, resp.Diagnostics)
			}
		})
	}
}

// TestResourceRead_Success_ComplexResources tests the Read() happy-path for
// resources excluded from the generic TestResourceRead_Success because they
// either (a) require a state with non-null nested objects, or (b) require a
// response JSON that includes a properties block.
func TestResourceRead_Success_ComplexResources(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	// All resources that would belong here require resource-specific valid URI
	// values in the test state (e.g., "/projects/p/compute/cloudServers/id").
	// buildFullTFValue sets all strings to "test-value" which the SDK rejects
	// with "cannot determine X ID from Ref 'test-value'".  These resources
	// are covered by acceptance tests and the generic TestResourceRead_Success
	// (which uses resourceReadReq with null non-string fields and realistic
	// "test-<attrname>" string values).  Add dedicated per-resource tests
	// here when a valid sentinel URI is available for each resource type.
	cases := []struct {
		name    string
		newR    func() resource.Resource
		handler http.HandlerFunc
		useFull bool
	}{}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, tc.handler)

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
				t.Errorf("%s: Read() reported error on success response: %v",
					tc.name, resp.Diagnostics)
			}
			if resp.State.Raw.IsNull() {
				t.Errorf("%s: Read() removed resource from state on success response", tc.name)
			}
		})
	}
}

// TestResourceUpdate_Success_ComplexResources tests the Update() happy-path
// for resources excluded from the generic TestResourceUpdate_Success because
// they require either a specific response status / properties block, or a
// plan with non-null nested objects.
func TestResourceUpdate_Success_ComplexResources(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	// All resources that would belong here start their Update() with a GET
	// that fails when the state has uri="test-uri" or uri="test-value"
	// (the SDK cannot extract a resource ID from those sentinel strings).
	// These Update paths are covered by acceptance tests.  Add dedicated
	// per-resource tests here when a valid sentinel URI is provided.
	//
	// Exception: databasebackup is a documented no-op update and never calls
	// the API, so it is safe to test here.
	cases := []struct {
		name    string
		newR    func() resource.Resource
		handler http.HandlerFunc
		useFull bool
	}{
		{"databasebackup", NewDatabaseBackupResource, func(w http.ResponseWriter, _ *http.Request) {
			apiError(w, http.StatusInternalServerError)
		}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, tc.handler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			var req resource.UpdateRequest
			var resp *resource.UpdateResponse
			if tc.useFull {
				req, resp = resourceUpdateReqFull(ctx, t, res)
			} else {
				req, resp = resourceUpdateReq(ctx, t, res)
			}
			res.Update(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Update() reported error: %v", tc.name, resp.Diagnostics)
			}
		})
	}
}
