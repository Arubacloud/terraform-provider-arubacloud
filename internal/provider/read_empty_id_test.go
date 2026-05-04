package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// These tests demonstrate and guard against the Read-empty-ID bug across every
// resource in the provider. The Plugin Framework contract says Read should
// treat an empty / null / unknown own id as "the resource is gone" and call
// resp.State.RemoveResource(ctx) so the caller plans a Create. Returning a
// hard diagnostic instead breaks `terraform apply -refresh-only`,
// state-recovery flows, and Crossplane/upjet Observe.
//
// TestAllResourcesReadEmptyID_RemovesState sweeps every resource registered
// on the provider so new resources are covered automatically and can't
// silently reintroduce the buggy pattern. TestVPCRead_EmptyParentID_StillErrors
// pins the companion invariant a fix must preserve: an empty *parent* id
// (e.g. project_id) is a genuine misconfiguration and must still produce a
// diagnostic.

type emptyIDCase struct {
	name string
	val  tftypes.Value
}

func emptyIDCases() []emptyIDCase {
	return []emptyIDCase{
		{"empty string", tftypes.NewValue(tftypes.String, "")},
		{"null", tftypes.NewValue(tftypes.String, nil)},
		{"unknown", tftypes.NewValue(tftypes.String, tftypes.UnknownValue)},
	}
}

// stateForEmptyIDRead builds a tfsdk.State matching the resource's schema
// where `id` is set to idVal and every other top-level string attribute
// (typical home of parent ids like project_id / vpc_id) is filled with a
// non-empty placeholder so the "own id empty" branch is isolated from the
// "parent id empty" branch. Non-string attributes are left null.
func stateForEmptyIDRead(ctx context.Context, t *testing.T, r resource.Resource, idVal tftypes.Value) tfsdk.State {
	t.Helper()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema error: %v", schemaResp.Diagnostics)
	}
	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("schema top type is not an object")
	}
	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, ty := range objType.AttributeTypes {
		switch {
		case name == "id":
			attrs[name] = idVal
		case ty.Is(tftypes.String):
			attrs[name] = tftypes.NewValue(tftypes.String, "placeholder")
		default:
			attrs[name] = tftypes.NewValue(ty, nil)
		}
	}
	return tfsdk.State{
		Raw:    tftypes.NewValue(objType, attrs),
		Schema: schemaResp.Schema,
	}
}

func TestAllResourcesReadEmptyID_RemovesState(t *testing.T) {
	ctx := context.Background()
	p := New("test")()

	resources := p.Resources(ctx)
	if len(resources) == 0 {
		t.Fatal("provider registered zero resources; sweep would be vacuous")
	}

	for _, mk := range resources {
		r := mk()

		metaResp := &resource.MetadataResponse{}
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "arubacloud"}, metaResp)
		name := metaResp.TypeName

		schemaResp := &resource.SchemaResponse{}
		r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
		if schemaResp.Diagnostics.HasError() {
			t.Errorf("%s: schema returned diagnostics: %v", name, schemaResp.Diagnostics)
			continue
		}
		if _, ok := schemaResp.Schema.Attributes["id"]; !ok {
			t.Errorf("%s: schema has no top-level id attribute, cannot sweep", name)
			continue
		}

		for _, tc := range emptyIDCases() {
			t.Run(fmt.Sprintf("%s/%s", name, tc.name), func(t *testing.T) {
				state := stateForEmptyIDRead(ctx, t, r, tc.val)
				req := resource.ReadRequest{State: state}
				resp := &resource.ReadResponse{State: tfsdk.State{Schema: state.Schema}}

				// Some buggy Read paths fall through to an unconfigured API
				// client and panic on a nil deref — treat that as a (louder)
				// flavour of the same bug.
				defer func() {
					if rec := recover(); rec != nil {
						t.Fatalf(
							"%s Read with own id=%s panicked instead of handling the empty-id case.\n"+
								"Panic: %v",
							name, tc.name, rec,
						)
					}
				}()

				r.Read(ctx, req, resp)

				if resp.Diagnostics.HasError() {
					t.Fatalf(
						"%s Read with own id=%s returned a hard error, violating the framework contract.\n"+
							"Expected: resp.State.RemoveResource(ctx) so the caller plans a Create.\n"+
							"Got diagnostics: %v",
						name, tc.name, resp.Diagnostics,
					)
				}
				if !resp.State.Raw.IsNull() {
					t.Fatalf(
						"%s Read with own id=%s did not remove the resource from state.\n"+
							"Expected: resp.State.Raw.IsNull() == true.\n"+
							"Got: %v",
						name, tc.name, resp.State.Raw,
					)
				}
			})
		}
	}
}

// TestVPCRead_EmptyParentID_StillErrors pins the companion invariant: a
// missing parent id (project_id) is a real misconfiguration and must keep
// erroring. A fix that calls RemoveResource unconditionally would flip this
// test to green, so it acts as a guard against over-broad fixes.
func TestVPCRead_EmptyParentID_StillErrors(t *testing.T) {
	ctx := context.Background()
	r := &VPCResource{}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema error: %v", schemaResp.Diagnostics)
	}

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("schema top type is not an object")
	}
	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, ty := range objType.AttributeTypes {
		attrs[name] = tftypes.NewValue(ty, nil)
	}
	attrs["id"] = tftypes.NewValue(tftypes.String, "some-vpc-id")
	attrs["project_id"] = tftypes.NewValue(tftypes.String, "")

	req := resource.ReadRequest{
		State: tfsdk.State{
			Raw:    tftypes.NewValue(objType, attrs),
			Schema: schemaResp.Schema,
		},
	}
	resp := &resource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Read(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("Read with empty parent project_id should have errored (genuine misconfiguration), got no diagnostics")
	}
}
