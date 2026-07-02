package provider

import (
	"context"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ── cidrContains ──────────────────────────────────────────────────────────────

func TestCidrContains(t *testing.T) {
	parseCIDR := func(s string) *net.IPNet {
		t.Helper()
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			t.Fatalf("ParseCIDR(%q): %v", s, err)
		}
		return n
	}

	cases := []struct {
		parent string
		child  string
		want   bool
	}{
		// Subnet is contained in the parent.
		{"10.0.0.0/8", "10.1.2.0/24", true},
		// Equal networks are considered contained.
		{"10.0.0.0/24", "10.0.0.0/24", true},
		// Child prefix longer than parent — still a subset.
		{"192.168.0.0/16", "192.168.1.0/24", true},
		// Child has a different network address — not contained.
		{"10.0.0.0/24", "10.0.1.0/24", false},
		// Parent is a subnet of child (narrower mask) — not contained.
		{"10.0.0.0/24", "10.0.0.0/16", false},
		// IPv4 vs IPv6 — different bit widths, never contained.
		{"10.0.0.0/8", "::1/128", false},
	}

	for _, tc := range cases {
		parent, child := parseCIDR(tc.parent), parseCIDR(tc.child)
		if got := cidrContains(parent, child); got != tc.want {
			t.Errorf("cidrContains(%q, %q) = %v, want %v", tc.parent, tc.child, got, tc.want)
		}
	}
}

// subnetAdvancedJSON is a Subnet API response with a full properties block
// including type=Advanced, a network address, DHCP enabled with a range,
// routes, and DNS entries.  This exercises the DHCP-mapping code paths in
// SubnetResource.Read() that are skipped by minimalActiveJSON.
const subnetAdvancedJSON = `{` +
	`"metadata":{` +
	`"id":"test-id","name":"test-name",` +
	`"uri":"/subnets/test-id",` +
	`"location":{"value":"test-location"},` +
	`"tags":["env:test"]` +
	`},` +
	`"status":{"state":"Active"},` +
	`"properties":{` +
	`"type":"Advanced",` +
	`"network":{"address":"10.0.0.0/24"},` +
	`"dhcp":{` +
	`"enabled":true,` +
	`"range":{"start":"10.0.0.100","count":50},` +
	`"routes":[{"address":"10.0.0.0/24","gateway":"10.0.0.1"}],` +
	`"dns":["8.8.8.8","8.8.4.4"]` +
	`}` +
	`}` +
	`}`

// subnetBasicJSON is a Basic Subnet response with no network block.  This
// exercises the shouldSetNetwork=false branch (type != "Advanced" and no
// network in state) which sets the entire network block to null.
const subnetBasicNoNetworkJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name"},` +
	`"status":{"state":"Active"},` +
	`"properties":{"type":"Basic"}` +
	`}`

// TestSubnetResource_Read_WithDHCP verifies that SubnetResource.Read() correctly
// maps a response that includes a full DHCP block (range, routes, DNS).  Using
// resourceReadReqFull provides non-null network in state so shouldSetNetwork is
// true and the DHCP-mapping section is exercised.
func TestSubnetResource_Read_WithDHCP(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(subnetAdvancedJSON)) //nolint:errcheck
	})

	res := NewSubnetResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceReadReqFull(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("SubnetResource Read() reported error with DHCP response: %v", resp.Diagnostics)
	}
	if resp.State.Raw.IsNull() {
		t.Fatal("SubnetResource Read() removed resource from state unexpectedly")
	}
}

// TestSubnetResource_Read_BasicNoNetwork verifies that SubnetResource.Read() with
// a Basic subnet response and null network in state correctly sets
// data.Network to null (the shouldSetNetwork=false / else branch).
func TestSubnetResource_Read_BasicNoNetwork(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(subnetBasicNoNetworkJSON)) //nolint:errcheck
	})

	res := NewSubnetResource()
	configureResource(ctx, t, res, mockClient)

	// Use minimal Read request (network is null in state) with a Basic subnet type.
	req, resp := resourceReadReq(ctx, t, res)
	res.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("SubnetResource Read() reported error for Basic subnet: %v", resp.Diagnostics)
	}
}

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
