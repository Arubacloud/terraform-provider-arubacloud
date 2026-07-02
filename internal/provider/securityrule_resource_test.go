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

func TestNormalizeProtocol(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"any", "Any"},
		{"ANY", "Any"},
		{"Any", "Any"},
		{"tcp", "TCP"},
		{"TCP", "TCP"},
		{"udp", "UDP"},
		{"UDP", "UDP"},
		{"icmp", "ICMP"},
		{"ICMP", "ICMP"},
		{"", ""},
		{"other", "Other"},
	}
	for _, tc := range cases {
		if got := normalizeProtocol(tc.in); got != tc.want {
			t.Errorf("normalizeProtocol(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeTargetKind(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"IP", "IP"},
		{"ip", "IP"},
		{"Ip", "IP"},
		{"SecurityGroup", "SecurityGroup"},
		{"securitygroup", "SecurityGroup"},
		{"SECURITYGROUP", "SecurityGroup"},
		{"", ""},
		{"unknown", "unknown"},
	}
	for _, tc := range cases {
		if got := normalizeTargetKind(tc.in); got != tc.want {
			t.Errorf("normalizeTargetKind(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

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

// TestSecurityRuleUpdate_PropertiesChangedError verifies that the securityrule
// Update() detects when immutable properties (direction, protocol, port, target)
// differ between the current API state and the Terraform plan, and adds an
// appropriate error diagnostic without making a PATCH call.
//
// This test covers the property-extraction and comparison section of Update()
// (the largest uncovered block) that the generic TestResourceUpdate_APIError
// misses because 500 on GET causes an early return before property extraction.
func TestSecurityRuleUpdate_PropertiesChangedError(t *testing.T) {
	ctx := context.Background()

	// Return the full security rule JSON for all GET requests so that the
	// properties extraction section is exercised.  Non-GET (PATCH) should not
	// be called when properties differ, so returning 500 is a safety net.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(securityruleFullJSON)) //nolint:errcheck
		} else {
			apiError(w, http.StatusInternalServerError)
		}
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewSecurityRuleResource()
	configureResource(ctx, t, res, mockClient)

	// Use resourceUpdateReqFull so data.Properties is non-null.  All string
	// attributes are set to "test-value", which differs from the current API
	// values (direction="Ingress", protocol="TCP", port="80", etc.).
	// This exercises the propertiesChanged=true branch and the early-return
	// with "Cannot Update Security Rule Properties" error.
	req, resp := resourceUpdateReqFull(ctx, t, res)
	res.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("securityrule Update() should fail when properties differ from current API state")
	}
}
