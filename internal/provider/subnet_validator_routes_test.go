package provider

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// buildSubnetAdvancedWithRoutes creates a tfsdk.Config for SubnetResource with
// type = "Advanced", a network CIDR of subnetCIDR, and the given DHCP routes.
// Each route element is a map{"address": CIDR, "gateway": IP}.
//
// The tftypes.Object hierarchy is derived from the resource schema so it
// remains correct even if the schema changes field types.
func buildSubnetAdvancedWithRoutes(t *testing.T, subnetCIDR string, routes []map[string]string) tfsdk.Config {
	t.Helper()
	ctx := context.Background()

	r := NewSubnetResource()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatal("buildSubnetAdvancedWithRoutes: schema root is not an object type")
	}

	// Extract nested attribute types from the schema-derived tftypes hierarchy.
	networkTFType, ok := objType.AttributeTypes["network"].(tftypes.Object)
	if !ok {
		t.Fatal("network attribute is not an object type")
	}
	dhcpTFType, ok := networkTFType.AttributeTypes["dhcp"].(tftypes.Object)
	if !ok {
		t.Fatal("dhcp attribute is not an object type")
	}
	routesListType, ok := dhcpTFType.AttributeTypes["routes"].(tftypes.List)
	if !ok {
		t.Fatal("routes attribute is not a list type")
	}
	routeObjType, ok := routesListType.ElementType.(tftypes.Object)
	if !ok {
		t.Fatal("routes element type is not an object type")
	}
	rangeTFType, ok := dhcpTFType.AttributeTypes["range"].(tftypes.Object)
	if !ok {
		t.Fatal("range attribute is not an object type")
	}
	dnsListType, ok := dhcpTFType.AttributeTypes["dns"].(tftypes.List)
	if !ok {
		t.Fatal("dns attribute is not a list type")
	}

	// Build route tftypes.Values.
	routeVals := make([]tftypes.Value, len(routes))
	for i, r := range routes {
		routeVals[i] = tftypes.NewValue(routeObjType, map[string]tftypes.Value{
			"address": tftypes.NewValue(tftypes.String, r["address"]),
			"gateway": tftypes.NewValue(tftypes.String, r["gateway"]),
		})
	}

	// Build range object.
	rangeVal := tftypes.NewValue(rangeTFType, map[string]tftypes.Value{
		"start": tftypes.NewValue(tftypes.String, "10.0.0.100"),
		"count": tftypes.NewValue(tftypes.Number, new(big.Float).SetFloat64(50)),
	})

	// Build DHCP object.
	dhcpVal := tftypes.NewValue(dhcpTFType, map[string]tftypes.Value{
		"enabled": tftypes.NewValue(tftypes.Bool, true),
		"range":   rangeVal,
		"routes":  tftypes.NewValue(routesListType, routeVals),
		"dns":     tftypes.NewValue(dnsListType, []tftypes.Value{}),
	})

	// Build network object.
	networkVal := tftypes.NewValue(networkTFType, map[string]tftypes.Value{
		"address": tftypes.NewValue(tftypes.String, subnetCIDR),
		"dhcp":    dhcpVal,
	})

	// Build root object: use "Advanced" for type, the custom network value,
	// and "test-<name>" for all other string attributes.
	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, ty := range objType.AttributeTypes {
		switch name {
		case "type":
			attrs[name] = tftypes.NewValue(tftypes.String, "Advanced")
		case "network":
			attrs[name] = networkVal
		default:
			if ty.Is(tftypes.String) {
				attrs[name] = tftypes.NewValue(tftypes.String, "test-"+name)
			} else {
				attrs[name] = tftypes.NewValue(ty, nil)
			}
		}
	}

	return tfsdk.Config{
		Raw:    tftypes.NewValue(objType, attrs),
		Schema: schemaResp.Schema,
	}
}

// TestSubnetRouteAddressValidator_ValidRoute verifies that ValidateResource
// produces no diagnostics when the route address is within the subnet CIDR.
// This covers the full validator path through network.As → DHCP.As → routes
// loop → cidrContains (returning true).
func TestSubnetRouteAddressValidator_ValidRoute(t *testing.T) {
	ctx := context.Background()
	v := subnetRouteAddressValidator{}

	config := buildSubnetAdvancedWithRoutes(t, "10.0.0.0/24",
		[]map[string]string{{"address": "10.0.0.0/25", "gateway": "10.0.0.1"}})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	v.ValidateResource(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("ValidateResource with valid route reported error: %v", resp.Diagnostics)
	}
}

// TestSubnetRouteAddressValidator_RouteOutsideCIDR verifies that
// ValidateResource adds an attribute error when the route address is outside
// the subnet CIDR.  This covers the `!cidrContains(subnetNet, routeNet)` branch.
func TestSubnetRouteAddressValidator_RouteOutsideCIDR(t *testing.T) {
	ctx := context.Background()
	v := subnetRouteAddressValidator{}

	config := buildSubnetAdvancedWithRoutes(t, "192.168.0.0/24",
		[]map[string]string{{"address": "10.0.0.0/8", "gateway": "192.168.0.1"}})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	v.ValidateResource(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("ValidateResource with route outside subnet CIDR should have reported an error")
	}
}

// TestSubnetRouteAddressValidator_InvalidRouteCIDR verifies that
// ValidateResource adds an attribute error when the route address is not valid
// CIDR notation.  This covers the `err != nil` branch inside the routes loop.
func TestSubnetRouteAddressValidator_InvalidRouteCIDR(t *testing.T) {
	ctx := context.Background()
	v := subnetRouteAddressValidator{}

	config := buildSubnetAdvancedWithRoutes(t, "10.0.0.0/24",
		[]map[string]string{{"address": "not-a-cidr", "gateway": "10.0.0.1"}})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	v.ValidateResource(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("ValidateResource with invalid route CIDR should have reported an error")
	}
}

// TestSubnetRouteAddressValidator_EmptyRouteAddress verifies that
// ValidateResource skips routes whose address is an empty string.
func TestSubnetRouteAddressValidator_EmptyRouteAddress(t *testing.T) {
	ctx := context.Background()
	v := subnetRouteAddressValidator{}

	config := buildSubnetAdvancedWithRoutes(t, "10.0.0.0/24",
		[]map[string]string{{"address": "", "gateway": "10.0.0.1"}})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	v.ValidateResource(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("ValidateResource with empty route address should not report error: %v", resp.Diagnostics)
	}
}

// TestSubnetRouteAddressValidator_MultipleRoutes verifies that all routes in a
// list are validated (both valid and invalid routes in the same list).
func TestSubnetRouteAddressValidator_MultipleRoutes(t *testing.T) {
	ctx := context.Background()
	v := subnetRouteAddressValidator{}

	config := buildSubnetAdvancedWithRoutes(t, "10.0.0.0/24",
		[]map[string]string{
			{"address": "10.0.0.0/25", "gateway": "10.0.0.1"},    // valid
			{"address": "192.168.1.0/24", "gateway": "10.0.0.1"}, // outside CIDR
		})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	v.ValidateResource(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("ValidateResource with mixed routes should report error for out-of-CIDR route")
	}
}
