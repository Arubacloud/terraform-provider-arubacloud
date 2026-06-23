package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestSubnetConfigValidators_List verifies that SubnetResource implements
// ResourceWithConfigValidators and returns at least one validator.
func TestSubnetConfigValidators_List(t *testing.T) {
	ctx := context.Background()
	r := NewSubnetResource()

	rvc, ok := r.(resource.ResourceWithConfigValidators)
	if !ok {
		t.Fatal("SubnetResource does not implement ResourceWithConfigValidators")
	}

	validators := rvc.ConfigValidators(ctx)
	if len(validators) == 0 {
		t.Fatal("expected at least one config validator from SubnetResource.ConfigValidators()")
	}
}

// TestSubnetRouteAddressValidator_Description verifies that Description() and
// MarkdownDescription() return non-empty strings.
func TestSubnetRouteAddressValidator_Description(t *testing.T) {
	ctx := context.Background()
	v := subnetRouteAddressValidator{}

	desc := v.Description(ctx)
	if desc == "" {
		t.Error("Description() returned empty string")
	}

	mdDesc := v.MarkdownDescription(ctx)
	if mdDesc == "" {
		t.Error("MarkdownDescription() returned empty string")
	}
}

// buildSubnetConfig creates a tfsdk.Config for SubnetResource with the given
// attribute overrides.  All attributes not mentioned in overrides are set to
// null (all required attributes are strings, so the nil tftypes.Value is used).
func buildSubnetConfig(t *testing.T, overrides map[string]tftypes.Value) tfsdk.Config {
	t.Helper()
	ctx := context.Background()

	r := NewSubnetResource()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("buildSubnetConfig: schema root is not an object type")
	}

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, ty := range objType.AttributeTypes {
		if v, found := overrides[name]; found {
			attrs[name] = v
		} else {
			attrs[name] = tftypes.NewValue(ty, nil)
		}
	}

	return tfsdk.Config{
		Raw:    tftypes.NewValue(objType, attrs),
		Schema: schemaResp.Schema,
	}
}

// TestSubnetRouteAddressValidator_BasicType verifies that ValidateResource is
// a no-op for a subnet whose type is "Basic" (the validator only applies to
// "Advanced" subnets).
func TestSubnetRouteAddressValidator_BasicType(t *testing.T) {
	ctx := context.Background()
	v := subnetRouteAddressValidator{}

	config := buildSubnetConfig(t, map[string]tftypes.Value{
		"type": tftypes.NewValue(tftypes.String, "Basic"),
	})
	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	v.ValidateResource(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Basic subnet: expected no diagnostic, got: %v", resp.Diagnostics)
	}
}

// TestSubnetRouteAddressValidator_NullType verifies that ValidateResource is
// a no-op when type is null (unknown during plan).
func TestSubnetRouteAddressValidator_NullType(t *testing.T) {
	ctx := context.Background()
	v := subnetRouteAddressValidator{}

	// All attributes null — type is null.
	config := buildSubnetConfig(t, nil)
	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	v.ValidateResource(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("null type: expected no diagnostic, got: %v", resp.Diagnostics)
	}
}

// TestSubnetRouteAddressValidator_AdvancedNullNetwork verifies that
// ValidateResource skips validation when type is "Advanced" but network is
// null (the user may supply network later; validation would be premature).
func TestSubnetRouteAddressValidator_AdvancedNullNetwork(t *testing.T) {
	ctx := context.Background()
	v := subnetRouteAddressValidator{}

	config := buildSubnetConfig(t, map[string]tftypes.Value{
		"type": tftypes.NewValue(tftypes.String, "Advanced"),
		// network stays null (not provided)
	})
	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	v.ValidateResource(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Advanced with null network: expected no diagnostic, got: %v", resp.Diagnostics)
	}
}
