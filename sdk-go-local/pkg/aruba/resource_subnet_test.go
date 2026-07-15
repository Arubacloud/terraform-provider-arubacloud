package aruba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Compile-time Ref satisfaction
// --------------------------------------------------------------------------

var _ Ref = (*Subnet)(nil)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestSubnet_FluentSetters(t *testing.T) {
	parent := &VPC{}
	parent.fromResponse(vpcTestResponse("vpc-1", "my-vpc", "/projects/p1/providers/Aruba.Network/vpcs/vpc-1", "p1"))

	s := NewSubnet().
		InVPC(parent).
		Named("my-subnet").
		Tagged("net").
		Tagged("infra").
		Tagged("net"). // dedupe
		InRegion(RegionITBGBergamo).
		OfType(SubnetTypeAdvanced).
		NotDefault().
		WithCIDR("10.0.0.0/24")

	if s.Name() != "my-subnet" {
		t.Errorf("Name() = %q", s.Name())
	}
	if tags := s.Tags(); len(tags) != 2 || tags[0] != "net" || tags[1] != "infra" {
		t.Errorf("Tags() = %v", tags)
	}
	if s.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", s.Region())
	}
	if s.Type() != SubnetTypeAdvanced {
		t.Errorf("Type() = %q", s.Type())
	}
	if s.IsDefault() {
		t.Error("IsDefault() should be false")
	}
	if s.CIDR() != "10.0.0.0/24" {
		t.Errorf("CIDR() = %q", s.CIDR())
	}
	if s.VPCID() != "vpc-1" {
		t.Errorf("VPCID() = %q", s.VPCID())
	}
	if s.ProjectID() != "p1" {
		t.Errorf("ProjectID() = %q", s.ProjectID())
	}
	if s.Err() != nil {
		t.Errorf("Err() = %v", s.Err())
	}

	s.Untagged("net")
	if tags := s.Tags(); len(tags) != 1 || tags[0] != "infra" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	s.RetaggedAs("x", "y")
	if tags := s.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

// --------------------------------------------------------------------------
// IntoVPC with bad Ref — must not panic, must surface error via Err()
// --------------------------------------------------------------------------

func TestSubnet_IntoVPC_BadRef(t *testing.T) {
	s := NewSubnet().InVPC(URI("/garbage"))
	if s.Err() == nil {
		t.Error("expected Err() != nil for unresolvable Ref, got nil")
	}
}

// --------------------------------------------------------------------------
// SubnetDHCPCommon sub-builder
// --------------------------------------------------------------------------

func TestSubnet_DHCPSubBuilder(t *testing.T) {
	d := NewSubnetDHCP().
		Enabled().
		WithRange("10.0.0.10", 50).
		WithRoutes(SubnetDHCPRouteCommon{Address: "0.0.0.0/0", Gateway: "10.0.0.1"}).
		WithRoutes(SubnetDHCPRouteCommon{Address: "192.168.0.0/16", Gateway: "10.0.0.2"}).
		WithDNSServers("1.1.1.1").
		WithDNSServers("8.8.8.8")

	if !d.IsEnabled() {
		t.Error("IsEnabled() should be true")
	}
	if d.RangeStart() != "10.0.0.10" {
		t.Errorf("RangeStart() = %q", d.RangeStart())
	}
	if d.RangeCount() != 50 {
		t.Errorf("RangeCount() = %d", d.RangeCount())
	}

	routes := d.Routes()
	if len(routes) != 2 {
		t.Fatalf("Routes() len = %d", len(routes))
	}
	if routes[0].Address != "0.0.0.0/0" || routes[0].Gateway != "10.0.0.1" {
		t.Errorf("routes[0] = %+v", routes[0])
	}
	if routes[1].Address != "192.168.0.0/16" || routes[1].Gateway != "10.0.0.2" {
		t.Errorf("routes[1] = %+v", routes[1])
	}

	dns := d.DNS()
	if len(dns) != 2 || dns[0] != "1.1.1.1" || dns[1] != "8.8.8.8" {
		t.Errorf("DNS() = %v", dns)
	}

	// Mutating the returned slice must not affect the builder.
	routes[0].Address = "tampered"
	if d.Routes()[0].Address != "0.0.0.0/0" {
		t.Error("Routes() defensive copy broken — mutation leaked into builder")
	}

	// Nil DHCP toType must return nil.
	var nilDHCP *SubnetDHCPCommon
	if nilDHCP.build() != nil {
		t.Error("nil.build() should return nil")
	}
}

func TestSubnet_DHCPSubBuilder_NoRange(t *testing.T) {
	d := NewSubnetDHCP().Enabled()
	if d.RangeStart() != "" || d.RangeCount() != 0 {
		t.Error("unset range should yield zero values")
	}
	if d.Routes() != nil {
		t.Error("Routes() should be nil when empty")
	}
	if d.DNS() != nil {
		t.Error("DNS() should be nil when empty")
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestSubnet_ToRequestRoundTrip(t *testing.T) {
	s := NewSubnet().Named(
		"sn-1").
		Tagged("t1").
		Tagged("t2").
		InRegion(RegionITBGBergamo).
		OfType(SubnetTypeBasic).
		AsDefault().
		WithCIDR("10.1.2.0/24").
		WithDHCP(NewSubnetDHCP().
			Enabled().
			WithRange("10.1.2.10", 20).
			WithRoutes(SubnetDHCPRouteCommon{Address: "0.0.0.0/0", Gateway: "10.1.2.1"}).
			WithDNSServers("9.9.9.9"))

	req := s.RawRequest()

	if req.Metadata.Name != "sn-1" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if len(req.Metadata.Tags) != 2 {
		t.Errorf("Metadata.Tags = %v", req.Metadata.Tags)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Location.Value = %q", req.Metadata.Location.Value)
	}
	if req.Properties.Type != SubnetTypeBasic {
		t.Errorf("Properties.Type = %q", req.Properties.Type)
	}
	if req.Properties.Default == nil || !*req.Properties.Default {
		t.Error("Properties.Default should be true")
	}
	if req.Properties.Network == nil || req.Properties.Network.Address != "10.1.2.0/24" {
		t.Errorf("Properties.Network = %+v", req.Properties.Network)
	}
	if req.Properties.DHCP == nil {
		t.Fatal("Properties.DHCP should be non-nil")
	}
	if !req.Properties.DHCP.Enabled {
		t.Error("DHCP.Enabled should be true")
	}
	if req.Properties.DHCP.Range == nil || req.Properties.DHCP.Range.Start != "10.1.2.10" || req.Properties.DHCP.Range.Count != 20 {
		t.Errorf("DHCP.Range = %+v", req.Properties.DHCP.Range)
	}
	if len(req.Properties.DHCP.Routes) != 1 || req.Properties.DHCP.Routes[0].Gateway != "10.1.2.1" {
		t.Errorf("DHCP.Routes = %v", req.Properties.DHCP.Routes)
	}
	if len(req.Properties.DHCP.DNS) != 1 || req.Properties.DHCP.DNS[0] != "9.9.9.9" {
		t.Errorf("DHCP.DNS = %v", req.Properties.DHCP.DNS)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func subnetTestResponse(id, name, uri, projectID string) *types.SubnetResponse {
	state := types.State("Active")
	loc := &types.LocationResponse{Value: RegionITBGBergamo}
	addr := "10.0.0.0/24"
	return &types.SubnetResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			URI:              &uri,
			Name:             &name,
			Tags:             []string{"tag1"},
			LocationResponse: loc,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{
				ID: projectID,
			},
		},
		Properties: types.SubnetPropertiesResponse{
			Type:    SubnetTypeAdvanced,
			Default: true,
			Network: &types.SubnetNetworkCommon{Address: addr},
			DHCP: &types.SubnetDHCPCommon{
				Enabled: true,
				Range:   &types.SubnetDHCPRangeCommon{Start: "10.0.0.10", Count: 50},
				DNS:     []string{"8.8.8.8"},
			},
			LinkedResources: []types.LinkedResourceCommon{
				{URI: uri + "/linked-res", StrictCorrelation: false},
			},
		},
		Status: types.ResourceStatusResponse{
			State: &state,
			DisableStatusInfoResponse: &types.DisableStatusInfoResponse{
				IsDisabled: false,
			},
		},
	}
}

func TestSubnet_FromResponseHydration(t *testing.T) {
	uri := "/projects/p1/providers/Aruba.Network/vpcs/v1/subnets/s1"
	resp := subnetTestResponse("s1", "my-subnet", uri, "p1")

	s := &Subnet{}
	s.fromResponse(resp)

	if s.ID() != "s1" {
		t.Errorf("ID() = %q", s.ID())
	}
	if s.SubnetID() != "s1" {
		t.Errorf("SubnetID() = %q", s.SubnetID())
	}
	if s.URI() != uri {
		t.Errorf("URI() = %q", s.URI())
	}
	if s.Name() != "my-subnet" {
		t.Errorf("Name() = %q", s.Name())
	}
	if tags := s.Tags(); len(tags) != 1 || tags[0] != "tag1" {
		t.Errorf("Tags() = %v", tags)
	}
	if s.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", s.Region())
	}
	if s.State() != "Active" {
		t.Errorf("State() = %q", s.State())
	}
	if s.IsDisabled() {
		t.Error("IsDisabled() should be false")
	}
	if linked := s.LinkedResources(); len(linked) != 1 {
		t.Errorf("LinkedResources() len = %d", len(linked))
	}
	if s.Type() != SubnetTypeAdvanced {
		t.Errorf("Type() = %q", s.Type())
	}
	if !s.IsDefault() {
		t.Error("IsDefault() should be true")
	}
	if s.CIDR() != "10.0.0.0/24" {
		t.Errorf("CIDR() = %q", s.CIDR())
	}
	if s.DHCP() == nil || !s.DHCP().IsEnabled() {
		t.Error("DHCP() should be non-nil and enabled")
	}
	if s.DHCP().RangeStart() != "10.0.0.10" || s.DHCP().RangeCount() != 50 {
		t.Errorf("DHCP range = (%q, %d)", s.DHCP().RangeStart(), s.DHCP().RangeCount())
	}
	if dns := s.DHCP().DNS(); len(dns) != 1 || dns[0] != "8.8.8.8" {
		t.Errorf("DHCP.DNS() = %v", dns)
	}
	if s.ProjectID() != "p1" {
		t.Errorf("ProjectID() = %q", s.ProjectID())
	}
	if s.VPCID() != "v1" {
		t.Errorf("VPCID() = %q", s.VPCID())
	}
	if s.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestSubnet_FromResponsePartial(t *testing.T) {
	// nil response is a no-op
	s := &Subnet{}
	s.fromResponse(nil)
	if s.ID() != "" || s.URI() != "" || s.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}

	// empty response — accessors must not panic
	s2 := &Subnet{}
	s2.fromResponse(&types.SubnetResponse{})
	if s2.ID() != "" || s2.URI() != "" || s2.State() != "" {
		t.Error("empty response should yield zero accessor values")
	}
	if s2.CIDR() != "" || s2.Type() != "" || s2.DHCP() != nil {
		t.Error("empty properties should yield zero accessor values")
	}
}

// --------------------------------------------------------------------------
// Ref + ancestor-ID satisfaction (runtime)
// --------------------------------------------------------------------------

func TestSubnet_RefSatisfaction(t *testing.T) {
	uri := "/projects/p99/providers/Aruba.Network/vpcs/v99/subnets/s99"
	s := &Subnet{}
	s.fromResponse(subnetTestResponse("s99", "n", uri, "p99"))

	// withSubnetID typed path
	sid, ok := extractID(s, func(r Ref) (string, bool) {
		if w, ok := r.(withSubnetID); ok {
			return w.SubnetID(), true
		}
		return "", false
	}, "subnets")
	if !ok || sid != "s99" {
		t.Errorf("extractID via withSubnetID = (%q, %v)", sid, ok)
	}

	// withVPCID typed path
	vid, ok := extractID(s, func(r Ref) (string, bool) {
		if w, ok := r.(withVPCID); ok {
			return w.VPCID(), true
		}
		return "", false
	}, "vpcs")
	if !ok || vid != "v99" {
		t.Errorf("extractID via withVPCID = (%q, %v)", vid, ok)
	}

	// withProjectID typed path
	pid, ok := extractID(s, func(r Ref) (string, bool) {
		if w, ok := r.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || pid != "p99" {
		t.Errorf("extractID via withProjectID = (%q, %v)", pid, ok)
	}
}

// --------------------------------------------------------------------------
// subnetIDsFromRef helper
// --------------------------------------------------------------------------

func TestSubnetIDsFromRef_TypedRef(t *testing.T) {
	uri := "/projects/p/providers/Aruba.Network/vpcs/v/subnets/s"
	s := &Subnet{}
	s.fromResponse(subnetTestResponse("s", "n", uri, "p"))
	pid, vid, sid, err := subnetIDsFromRef(s)
	if err != nil || pid != "p" || vid != "v" || sid != "s" {
		t.Errorf("subnetIDsFromRef typed = (%q, %q, %q, %v)", pid, vid, sid, err)
	}
}

func TestSubnetIDsFromRef_URIRef_APIForm(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Network/vpcs/v/subnets/s")
	pid, vid, sid, err := subnetIDsFromRef(ref)
	if err != nil || pid != "p" || vid != "v" || sid != "s" {
		t.Errorf("subnetIDsFromRef API URI = (%q, %q, %q, %v)", pid, vid, sid, err)
	}
}

func TestSubnetIDsFromRef_URIRef_NetworkForm(t *testing.T) {
	ref := URI("/projects/p/network/vpcs/v/subnets/s")
	pid, vid, sid, err := subnetIDsFromRef(ref)
	if err != nil || pid != "p" || vid != "v" || sid != "s" {
		t.Errorf("subnetIDsFromRef network URI = (%q, %q, %q, %v)", pid, vid, sid, err)
	}
}

func TestSubnetIDsFromRef_BadURI_MissingSubnet(t *testing.T) {
	_, _, _, err := subnetIDsFromRef(URI("/projects/p/providers/Aruba.Network/vpcs/v"))
	if err == nil {
		t.Error("expected error for URI without /subnets/<id>")
	}
}

func TestSubnetIDsFromRef_BadURI_MissingAll(t *testing.T) {
	_, _, _, err := subnetIDsFromRef(URI("/something/else"))
	if err == nil {
		t.Error("expected error for unrecognisable URI")
	}
}

// --------------------------------------------------------------------------
// subnetsClientAdapter — CRUD integration tests
// --------------------------------------------------------------------------

func buildSubnetTestAdapter(t *testing.T, handler http.HandlerFunc) *subnetsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newSubnetsClientAdapter(testutil.NewClient(t, server.URL))
}

const subnetSuccessBody = `{` +
	`"metadata":{"id":"sid","name":"my-subnet","uri":"/projects/p/providers/Aruba.Network/vpcs/v/subnets/sid","project":{"id":"p"}},` +
	`"properties":{"type":"Advanced","default":false,"network":{"address":"10.0.0.0/24"}},` +
	`"status":{"state":"Creating"}}`

func TestSubnetsClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.SubnetRequest
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, subnetSuccessBody)
	})

	vpc := &VPC{}
	vpc.fromResponse(vpcTestResponse("v", "my-vpc", "/projects/p/providers/Aruba.Network/vpcs/v", "p"))

	s := NewSubnet().
		InVPC(vpc).
		Named("my-subnet").
		OfType(SubnetTypeAdvanced).
		WithCIDR("10.0.0.0/24")

	result, err := adapter.Create(context.Background(), s)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "sid" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-subnet" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Metadata.Name != "my-subnet" {
		t.Errorf("request Name = %q", gotBody.Metadata.Name)
	}
	if gotBody.Properties.Type != SubnetTypeAdvanced {
		t.Errorf("request Type = %q", gotBody.Properties.Type)
	}
}

func TestSubnetsClientAdapter_Create_NoVPC(t *testing.T) {
	callCount := 0
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	_, err := adapter.Create(context.Background(), NewSubnet().
		Named("x"))
	if err == nil {
		t.Fatal("expected error when subnet has no VPC")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without VPC")
	}
}

func TestSubnetsClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" field — triggers MetadataValidationError
		fmt.Fprint(w, `{"metadata":{"name":"sn","uri":"/projects/p/providers/Aruba.Network/vpcs/v/subnets/x"},"properties":{},"status":{}}`)
	})

	vpc := &VPC{}
	vpc.fromResponse(vpcTestResponse("v", "n", "/projects/p/providers/Aruba.Network/vpcs/v", "p"))

	result, err := adapter.Create(context.Background(), NewSubnet().InVPC(vpc).
		Named("sn"))
	if err == nil {
		t.Fatal("expected MetadataValidationError, got nil")
	}
	var mvErr *types.MetadataValidationError
	if !errors.As(err, &mvErr) {
		t.Fatalf("expected *types.MetadataValidationError, got %T: %v", err, err)
	}
	if result == nil {
		t.Fatal("result must be non-nil alongside MetadataValidationError")
	}
}

func TestSubnetsClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "name is required", 422))
	})

	vpc := &VPC{}
	vpc.fromResponse(vpcTestResponse("v", "n", "/projects/p/providers/Aruba.Network/vpcs/v", "p"))

	result, err := adapter.Create(context.Background(), NewSubnet().InVPC(vpc))
	if err == nil {
		t.Fatal("expected error on 422")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("HTTPError.StatusCode = %d", httpErr.StatusCode)
	}
	if result == nil {
		t.Fatal("result must be non-nil on non-2xx")
	}
}

func TestSubnetsClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, subnetSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Network/vpcs/v/subnets/sid")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "sid" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.VPCID() != "v" {
		t.Errorf("VPCID() = %q", result.VPCID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	wantPath := "/projects/p/providers/Aruba.Network/vpcs/v/subnets/sid"
	if capturedPath != wantPath {
		t.Errorf("path = %q, want %q", capturedPath, wantPath)
	}
}

func TestSubnetsClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, subnetSuccessBody)
	})

	uri := "/projects/p/providers/Aruba.Network/vpcs/v/subnets/sid"
	existing := &Subnet{}
	existing.fromResponse(subnetTestResponse("sid", "my-subnet", uri, "p"))

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "sid" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestSubnetsClientAdapter_Update_Success(t *testing.T) {
	var capturedBody types.SubnetRequest
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"sid","name":"renamed","uri":"/projects/p/providers/Aruba.Network/vpcs/v/subnets/sid","project":{"id":"p"}},"properties":{},"status":{}}`)
	})

	uri := "/projects/p/providers/Aruba.Network/vpcs/v/subnets/sid"
	s := &Subnet{}
	s.fromResponse(subnetTestResponse("sid", "orig", uri, "p"))
	s.Named("renamed")

	result, err := adapter.Update(context.Background(), s)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if result.Name() != "renamed" {
		t.Errorf("Name() = %q", result.Name())
	}
	if capturedBody.Metadata.Name != "renamed" {
		t.Errorf("request Name = %q", capturedBody.Metadata.Name)
	}
}

func TestSubnetsClientAdapter_Update_NoID(t *testing.T) {
	callCount := 0
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	vpc := &VPC{}
	vpc.fromResponse(vpcTestResponse("v", "n", "/projects/p/providers/Aruba.Network/vpcs/v", "p"))

	s := NewSubnet().InVPC(vpc).
		Named("x")
	_, err := adapter.Update(context.Background(), s)
	if err == nil {
		t.Fatal("expected error when subnet has no ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when ID is missing")
	}
}

func TestSubnetsClientAdapter_Update_NoVPC(t *testing.T) {
	callCount := 0
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	id := "sid"
	s := &Subnet{}
	s.fromResponse(&types.SubnetResponse{
		Metadata: types.ResourceMetadataResponse{
			ID: &id,
		},
	})

	_, err := adapter.Update(context.Background(), s)
	if err == nil {
		t.Fatal("expected error when subnet has no VPC")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without VPC")
	}
}

func TestSubnetsClientAdapter_Delete_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	err := adapter.Delete(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made for bad Ref")
	}
}

func TestSubnetsClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v/subnets/sid"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestSubnetsClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "subnet not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v/subnets/missing"))
	if err == nil {
		t.Fatal("expected error on 404")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

func TestDHCPFromType_Nil(t *testing.T) {
	// Covers the dhcpFromType(nil) early-return branch.
	if got := dhcpFromType(nil); got != nil {
		t.Errorf("dhcpFromType(nil) = %v, want nil", got)
	}
}

func TestDHCPFromType_WithRoutes(t *testing.T) {
	// Covers the t.Routes > 0 branch in dhcpFromType.
	dhcp := &types.SubnetDHCPCommon{
		Enabled: true,
		Routes:  []types.SubnetDHCPRouteCommon{{Address: "10.0.0.0/8", Gateway: "10.0.0.1"}},
	}
	got := dhcpFromType(dhcp)
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if len(got.inner.Routes) != 1 {
		t.Errorf("Routes len = %d, want 1", len(got.inner.Routes))
	}
}

// IsDefault zero-value test covers the 66.7% branch.
func TestSubnet_IsDefault_ZeroValue(t *testing.T) {
	s := NewSubnet()
	if s.IsDefault() {
		t.Error("IsDefault() on zero value should be false")
	}
}

func TestSubnetsClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "subnet not found", 404))
	})

	ref := URI("/projects/p/providers/Aruba.Network/vpcs/v/subnets/missing")
	result, err := adapter.Get(context.Background(), ref)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
	if result == nil {
		t.Fatal("result must be non-nil on non-2xx")
	}
}

func TestSubnetsClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "subnet not found", 404))
	})

	s := &Subnet{}
	s.fromResponse(subnetTestResponse("s-1", "my-subnet", "/projects/p/providers/Aruba.Network/vpcs/v/subnets/s-1", "p"))
	_, err := adapter.Update(context.Background(), s)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
}

func TestSubnetsClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Forbidden", "access denied", 403))
	})

	_, err := adapter.List(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
}

func TestSubnetIDsFromRef_BadURI_MissingVPC(t *testing.T) {
	// URI has subnets segment but no vpcs segment
	_, _, _, err := subnetIDsFromRef(URI("/projects/p/subnets/s"))
	if err == nil {
		t.Error("expected error for URI without /vpcs/<id>")
	}
}

func TestSubnetIDsFromRef_BadURI_MissingProject(t *testing.T) {
	// URI has subnets+vpcs segments but no projects segment
	_, _, _, err := subnetIDsFromRef(URI("/providers/Aruba.Network/vpcs/v/subnets/s"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

func TestSubnetsClientAdapter_Create_WithBuilderError(t *testing.T) {
	callCount := 0
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})
	s := NewSubnet().InVPC(URI("/garbage"))
	_, err := adapter.Create(context.Background(), s)
	if err == nil {
		t.Fatal("expected error for builder error")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when builder has errors")
	}
}

func TestSubnetsClientAdapter_Get_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	result, err := adapter.Get(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
	if result != nil {
		t.Error("result should be nil on bad Ref")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made for bad Ref")
	}
}

func TestSubnetsClientAdapter_Get_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newSubnetsClientAdapter(testutil.NewClient(t, server.URL))
	result, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v/subnets/s"))
	if err == nil {
		t.Fatal("expected transport error")
	}
	_ = result
}

func TestSubnetsClientAdapter_Update_WithBuilderError(t *testing.T) {
	callCount := 0
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	s := NewSubnet().InVPC(URI("/garbage"))
	_, err := adapter.Update(context.Background(), s)
	if err == nil {
		t.Fatal("expected error for builder error")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when builder has errors")
	}
}

func TestSubnetsClientAdapter_Update_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newSubnetsClientAdapter(testutil.NewClient(t, server.URL))
	s := &Subnet{}
	s.fromResponse(subnetTestResponse("s1", "n1", "/projects/p/providers/Aruba.Network/vpcs/v/subnets/s1", "p"))
	_, err := adapter.Update(context.Background(), s)
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestSubnetsClientAdapter_Delete_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newSubnetsClientAdapter(testutil.NewClient(t, server.URL))
	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v/subnets/s"))
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestSubnetsClientAdapter_List_BadVPCRef(t *testing.T) {
	callCount := 0
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad VPC Ref")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made for bad VPC Ref")
	}
}

func TestSubnetsClientAdapter_List_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newSubnetsClientAdapter(testutil.NewClient(t, server.URL))
	_, err := adapter.List(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v"))
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestSubnetsClientAdapter_List_ProjectIDBackfill(t *testing.T) {
	// Response items without projectID in metadata or URI: triggers s.projectID = projectID backfill
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":1,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"s-x","name":"s-x"},"properties":{"type":"Basic"},"status":{}}`+
			`]}`)
	})

	list, err := adapter.List(context.Background(), URI("/projects/proj-x/providers/Aruba.Network/vpcs/vpc-x"))
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	items := list.Items()
	if len(items) != 1 {
		t.Fatalf("Items() len = %d", len(items))
	}
	if items[0].ProjectID() != "proj-x" {
		t.Errorf("ProjectID() after backfill = %q, want %q", items[0].ProjectID(), "proj-x")
	}
}

func TestSubnetsClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildSubnetTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"s1","name":"n1","uri":"/projects/p/providers/Aruba.Network/vpcs/v/subnets/s1","project":{"id":"p"}},"properties":{"type":"Basic","default":false},"status":{}},`+
			`{"metadata":{"id":"s2","name":"n2","uri":"/projects/p/providers/Aruba.Network/vpcs/v/subnets/s2","project":{"id":"p"}},"properties":{"type":"Advanced","default":true},"status":{}}`+
			`]}`)
	})

	list, err := adapter.List(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v"))
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if list.Total() != 2 {
		t.Errorf("Total() = %d", list.Total())
	}
	items := list.Items()
	if len(items) != 2 {
		t.Fatalf("Items() len = %d", len(items))
	}
	if items[0].ID() != "s1" || items[0].Name() != "n1" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[0].VPCID() != "v" || items[0].ProjectID() != "p" {
		t.Errorf("items[0] ancestor IDs = VPC:%q Project:%q", items[0].VPCID(), items[0].ProjectID())
	}
	if items[1].ID() != "s2" || !items[1].IsDefault() {
		t.Errorf("items[1] ID=%q IsDefault=%v", items[1].ID(), items[1].IsDefault())
	}
}

func TestSubnet_FromResponse_SetsStatus(t *testing.T) {
	s := &Subnet{}
	state := types.State("Active")
	s.fromResponse(&types.SubnetResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if s.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", s.State())
	}
}

func TestSubnetsClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, subnetSuccessBody)
	})
	adapter := newSubnetsClientAdapter(testutil.NewClient(t, server.URL))
	sub, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v/subnets/sid"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&sub.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned Subnet")
	}
}

func TestSubnetRef(t *testing.T) {
	ref := SubnetRef("p-1", "vpc-1", "sn-1")
	want := "/projects/p-1/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1"
	if ref.URI() != want {
		t.Errorf("SubnetRef URI = %q, want %q", ref.URI(), want)
	}
	ids := parseURIIDs(ref.URI())
	if ids["projects"] != "p-1" || ids["vpcs"] != "vpc-1" || ids["subnets"] != "sn-1" {
		t.Errorf("parseURIIDs = %v", ids)
	}
}

// --------------------------------------------------------------------------
// Network getter
// --------------------------------------------------------------------------

func TestSubnet_Network_Unset(t *testing.T) {
	s := &Subnet{}
	if got := s.Network(); got != "" {
		t.Errorf("Network() = %q, want empty", got)
	}
}

func TestSubnet_Network_WithCIDR(t *testing.T) {
	s := NewSubnet().WithCIDR("192.168.1.0/24")
	if got := s.Network(); got != "192.168.1.0/24" {
		t.Errorf("Network() = %q, want 192.168.1.0/24", got)
	}
}

func TestSubnet_Network_AliasForCIDR(t *testing.T) {
	uri := "/projects/p/providers/Aruba.Network/vpcs/v/subnets/s-1"
	s := &Subnet{}
	s.fromResponse(subnetTestResponse("s-1", "my-subnet", uri, "p"))
	if got, want := s.Network(), s.CIDR(); got != want {
		t.Errorf("Network() = %q, CIDR() = %q — should be equal", got, want)
	}
}

func TestSubnet_Network_FromResponse(t *testing.T) {
	uri := "/projects/p/providers/Aruba.Network/vpcs/v/subnets/s-1"
	s := &Subnet{}
	s.fromResponse(subnetTestResponse("s-1", "my-subnet", uri, "p"))
	if got := s.Network(); got != "10.0.0.0/24" {
		t.Errorf("Network() = %q, want 10.0.0.0/24", got)
	}
}
