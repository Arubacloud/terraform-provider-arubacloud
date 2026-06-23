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

// buildSecurityRuleCreatePlanWithProtocol creates a CreateRequest for the
// SecurityRuleResource where properties.protocol is set to the given value
// and properties.port is null (omitted).  All other string attributes use the
// "test-<name>" default from resourceCreateReq.
//
// This is used to exercise the port-clearing and JSON-manipulation code path
// that runs when protocol is "Any" or "ICMP".
func buildSecurityRuleCreatePlanWithProtocol(ctx context.Context, t *testing.T, protocol string) (resource.CreateRequest, *resource.CreateResponse) {
	t.Helper()

	r := NewSecurityRuleResource()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatal("buildSecurityRuleCreatePlanWithProtocol: schema root is not an object type")
	}

	// Get the tftypes for the nested properties and target objects.
	propsTFType, ok := objType.AttributeTypes["properties"].(tftypes.Object)
	if !ok {
		t.Fatal("buildSecurityRuleCreatePlanWithProtocol: properties attribute is not an object type")
	}
	targetTFType, ok := propsTFType.AttributeTypes["target"].(tftypes.Object)
	if !ok {
		t.Fatal("buildSecurityRuleCreatePlanWithProtocol: target attribute is not an object type")
	}

	// Build target: kind = "IP", value = "0.0.0.0/0"
	targetVal := tftypes.NewValue(targetTFType, map[string]tftypes.Value{
		"kind":  tftypes.NewValue(tftypes.String, "IP"),
		"value": tftypes.NewValue(tftypes.String, "0.0.0.0/0"),
	})

	// Build properties: protocol = caller-provided, port = null
	propsVal := tftypes.NewValue(propsTFType, map[string]tftypes.Value{
		"direction": tftypes.NewValue(tftypes.String, "Ingress"),
		"protocol":  tftypes.NewValue(tftypes.String, protocol),
		"port":      tftypes.NewValue(tftypes.String, nil), // null → port == "" in Create()
		"target":    targetVal,
	})

	// Build root: set all string attributes to "test-<name>",
	// override properties with the custom value, leave non-strings null.
	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, ty := range objType.AttributeTypes {
		switch name {
		case "properties":
			attrs[name] = propsVal
		default:
			if ty.Is(tftypes.String) {
				attrs[name] = tftypes.NewValue(tftypes.String, "test-"+name)
			} else {
				attrs[name] = tftypes.NewValue(ty, nil)
			}
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

// TestSecurityRuleCreate_AnyProtocol verifies that Create() succeeds when
// protocol is "Any" (no port required).  This exercises the port-clearing
// branch (strings.EqualFold(protocol, "Any")), the JSON-marshalling path that
// omits the port field, and the CRITICAL rebuild guard.
func TestSecurityRuleCreate_AnyProtocol(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Write([]byte(minimalActiveJSON)) //nolint:errcheck
	})

	res := NewSecurityRuleResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := buildSecurityRuleCreatePlanWithProtocol(ctx, t, "Any")
	res.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("SecurityRuleResource Create() with 'Any' protocol reported error: %v", resp.Diagnostics)
	}
}

// TestSecurityRuleCreate_ICMPProtocol verifies that Create() succeeds when
// protocol is "ICMP".  This exercises the same port-clearing branch as the
// "Any" test but via the ICMP condition.
func TestSecurityRuleCreate_ICMPProtocol(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Write([]byte(minimalActiveJSON)) //nolint:errcheck
	})

	res := NewSecurityRuleResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := buildSecurityRuleCreatePlanWithProtocol(ctx, t, "ICMP")
	res.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("SecurityRuleResource Create() with 'ICMP' protocol reported error: %v", resp.Diagnostics)
	}
}
