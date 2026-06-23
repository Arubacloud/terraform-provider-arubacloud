package provider

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// resourceReadReq builds a resource.ReadRequest whose State has every string
// attribute set to "test-<attr-name>" so that id / project_id / etc. are
// non-empty and the Read() method reaches the SDK call.  All non-string
// attributes (numbers, bools, lists, nested objects) are set to null.
func resourceReadReq(ctx context.Context, t *testing.T, r resource.Resource) (resource.ReadRequest, *resource.ReadResponse) {
	t.Helper()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("resourceReadReq: resource schema root is not an object type")
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
	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	return req, resp
}

// configureResource injects client into r via its Configure() method.
func configureResource(ctx context.Context, t *testing.T, r resource.Resource, client *ArubaCloudClient) {
	t.Helper()
	cfg, ok := r.(resource.ResourceWithConfigure)
	if !ok {
		t.Fatalf("configureResource: %T does not implement ResourceWithConfigure", r)
	}
	resp := &resource.ConfigureResponse{}
	cfg.Configure(ctx, resource.ConfigureRequest{ProviderData: client}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("configureResource: %v", resp.Diagnostics)
	}
}

// TestResourceRead_API404 verifies that Read() removes the resource from state
// (does NOT add an error diagnostic) when the API returns 404.
func TestResourceRead_API404(t *testing.T) {
	ctx := context.Background()

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
		{"kms", NewKMSResource},
		{"project", NewProjectResource},
		{"vpcpeering", NewVpcPeeringResource},
		{"vpntunnel", NewVPNTunnelResource},
		{"vpnroute", NewVPNRouteResource},
		{"dbaas", NewDBaaSResource},
		{"kaas", NewKaaSResource},
		{"containerregistry", NewContainerRegistryResource},
		{"restore", NewRestoreResource},
	}

	for _, tc := range resources {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
				apiError(w, http.StatusNotFound)
			})

			r := tc.newR()
			configureResource(ctx, t, r, mockClient)

			req, resp := resourceReadReq(ctx, t, r)
			r.Read(ctx, req, resp)

			// On 404, resources should remove themselves from state (no error).
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

// TestResourceRead_API500 verifies that Read() adds an error diagnostic when
// the API returns a non-404 error (e.g. 500).
func TestResourceRead_API500(t *testing.T) {
	ctx := context.Background()

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
		{"kms", NewKMSResource},
		{"project", NewProjectResource},
		{"vpcpeering", NewVpcPeeringResource},
		{"vpntunnel", NewVPNTunnelResource},
		{"vpnroute", NewVPNRouteResource},
		{"dbaas", NewDBaaSResource},
		{"kaas", NewKaaSResource},
		{"containerregistry", NewContainerRegistryResource},
		{"restore", NewRestoreResource},
	}

	for _, tc := range resources {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
				apiError(w, http.StatusInternalServerError)
			})

			r := tc.newR()
			configureResource(ctx, t, r, mockClient)

			req, resp := resourceReadReq(ctx, t, r)
			r.Read(ctx, req, resp)

			if !resp.Diagnostics.HasError() {
				t.Errorf("%s: Read() returned no error diagnostic for HTTP 500 response", tc.name)
			}
		})
	}
}
