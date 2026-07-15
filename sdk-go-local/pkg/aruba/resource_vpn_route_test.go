package aruba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Compile-time Ref satisfaction
// --------------------------------------------------------------------------

var _ Ref = (*VPNRoute)(nil)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestVPNRoute_FluentSetters(t *testing.T) {
	tun := &VPNTunnel{}
	tun.fromResponse(vpnTunnelTestResponse("t-1", "my-tunnel",
		"/projects/p/providers/Aruba.Network/vpnTunnels/t-1", "p"))

	r := NewVPNRoute().
		InVPNTunnel(tun).
		Named("my-route").
		Tagged("cloud").
		Tagged("vpn").
		Tagged("cloud"). // dedupe
		InRegion(RegionITBGBergamo).
		WithCloudSubnet("10.0.0.0/24").
		WithOnPremSubnet("192.168.0.0/24")

	if r.Name() != "my-route" {
		t.Errorf("Name() = %q", r.Name())
	}
	if tags := r.Tags(); len(tags) != 2 || tags[0] != "cloud" || tags[1] != "vpn" {
		t.Errorf("Tags() = %v", tags)
	}
	if r.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", r.Region())
	}
	if r.CloudSubnet() != "10.0.0.0/24" {
		t.Errorf("CloudSubnet() = %q", r.CloudSubnet())
	}
	if r.OnPremSubnet() != "192.168.0.0/24" {
		t.Errorf("OnPremSubnet() = %q", r.OnPremSubnet())
	}
	if r.VPNTunnelID() != "t-1" {
		t.Errorf("VPNTunnelID() = %q", r.VPNTunnelID())
	}
	if r.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", r.ProjectID())
	}
	if r.Err() != nil {
		t.Errorf("Err() = %v", r.Err())
	}

	r.Untagged("cloud")
	if tags := r.Tags(); len(tags) != 1 || tags[0] != "vpn" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	r.RetaggedAs("x", "y")
	if tags := r.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

// --------------------------------------------------------------------------
// IntoVPNTunnel — typed Ref
// --------------------------------------------------------------------------

func TestVPNRoute_IntoVPNTunnel_TypedRef(t *testing.T) {
	tun := &VPNTunnel{}
	tun.fromResponse(vpnTunnelTestResponse("t-1", "n",
		"/projects/p/providers/Aruba.Network/vpnTunnels/t-1", "p"))

	r := NewVPNRoute().InVPNTunnel(tun)

	if r.VPNTunnelID() != "t-1" {
		t.Errorf("VPNTunnelID() = %q", r.VPNTunnelID())
	}
	if r.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", r.ProjectID())
	}
	if r.Err() != nil {
		t.Errorf("Err() = %v", r.Err())
	}
}

// --------------------------------------------------------------------------
// IntoVPNTunnel — URI Ref (camelCase — production form; this exercises the mixin fix)
// --------------------------------------------------------------------------

func TestVPNRoute_IntoVPNTunnel_URIRef_CamelCase(t *testing.T) {
	r := NewVPNRoute().InVPNTunnel(URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1"))

	if r.Err() != nil {
		t.Fatalf("unexpected Err() = %v", r.Err())
	}
	if r.VPNTunnelID() != "t-1" {
		t.Errorf("VPNTunnelID() = %q", r.VPNTunnelID())
	}
	if r.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", r.ProjectID())
	}
}

// --------------------------------------------------------------------------
// IntoVPNTunnel — URI Ref (kebab form; mixin test form)
// --------------------------------------------------------------------------

// --------------------------------------------------------------------------
// IntoVPNTunnel — bad Ref
// --------------------------------------------------------------------------

func TestVPNRoute_IntoVPNTunnel_BadRef(t *testing.T) {
	r := NewVPNRoute().InVPNTunnel(URI("/garbage"))
	if r.Err() == nil {
		t.Error("expected Err() != nil for unresolvable Ref")
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestVPNRoute_ToRequestRoundTrip(t *testing.T) {
	r := NewVPNRoute().Named(
		"my-route").
		Tagged("t1").
		InRegion(RegionITBGBergamo).
		WithCloudSubnet("10.1.0.0/16").
		WithOnPremSubnet("172.16.0.0/12")

	req := r.RawRequest()

	if req.Metadata.Name != "my-route" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if len(req.Metadata.Tags) != 1 || req.Metadata.Tags[0] != "t1" {
		t.Errorf("Metadata.Tags = %v", req.Metadata.Tags)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Metadata.Location.Value = %q", req.Metadata.Location.Value)
	}
	if req.Properties.CloudSubnet != "10.1.0.0/16" {
		t.Errorf("Properties.CloudSubnet = %q", req.Properties.CloudSubnet)
	}
	if req.Properties.OnPremSubnet != "172.16.0.0/12" {
		t.Errorf("Properties.OnPremSubnet = %q", req.Properties.OnPremSubnet)
	}
}

// --------------------------------------------------------------------------
// toRequest — unset subnets emit empty strings (plain string, not pointer)
// --------------------------------------------------------------------------

func TestVPNRoute_ToRequest_UnsetSubnets_AreEmpty(t *testing.T) {
	req := NewVPNRoute().
		Named("bare").RawRequest()
	if req.Properties.CloudSubnet != "" {
		t.Errorf("CloudSubnet = %q, want empty", req.Properties.CloudSubnet)
	}
	if req.Properties.OnPremSubnet != "" {
		t.Errorf("OnPremSubnet = %q, want empty", req.Properties.OnPremSubnet)
	}
}

// --------------------------------------------------------------------------
// Ref + ancestor ID satisfaction (runtime)
// --------------------------------------------------------------------------

func TestVPNRoute_RefSatisfaction(t *testing.T) {
	r := &VPNRoute{}
	r.fromResponse(vpnRouteTestResponse("r-1", "my-route",
		"/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1", "p"))
	r.vpnTunnelID = "t-1"

	rid, ok := extractID(r, func(ref Ref) (string, bool) {
		if w, ok := ref.(withVPNRouteID); ok {
			return w.VPNRouteID(), true
		}
		return "", false
	}, "vpnRoutes")
	if !ok || rid != "r-1" {
		t.Errorf("extractID via withVPNRouteID = (%q, %v)", rid, ok)
	}

	tid, ok := extractID(r, func(ref Ref) (string, bool) {
		if w, ok := ref.(withVPNTunnelID); ok {
			return w.VPNTunnelID(), true
		}
		return "", false
	}, "vpnTunnels")
	if !ok || tid != "t-1" {
		t.Errorf("extractID via withVPNTunnelID = (%q, %v)", tid, ok)
	}

	pid, ok := extractID(r, func(ref Ref) (string, bool) {
		if w, ok := ref.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || pid != "p" {
		t.Errorf("extractID via withProjectID = (%q, %v)", pid, ok)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func vpnRouteTestResponse(id, name, uri, projectID string) *types.VPNRouteResponse {
	state := types.State("Active")
	cloud := "10.0.0.0/24"
	onPrem := "192.168.0.0/24"
	loc := &types.LocationResponse{Value: RegionITBGBergamo}
	return &types.VPNRouteResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			URI:              &uri,
			Name:             &name,
			Tags:             []string{"route-tag"},
			LocationResponse: loc,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{
				ID: projectID,
			},
		},
		Properties: types.VPNRoutePropertiesResponse{
			CloudSubnet:  types.SubnetCIDROrRef{CIDR: cloud},
			OnPremSubnet: onPrem,
			LinkedResources: []types.LinkedResourceCommon{
				{URI: "/projects/p/providers/Aruba.Network/vpnTunnels/t-1"},
			},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}
}

func TestVPNRoute_FromResponseHydration(t *testing.T) {
	r := &VPNRoute{}
	resp := vpnRouteTestResponse("r-1", "my-route",
		"/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1", "p")
	r.fromResponse(resp)

	if r.ID() != "r-1" {
		t.Errorf("ID() = %q", r.ID())
	}
	if r.URI() != "/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1" {
		t.Errorf("URI() = %q", r.URI())
	}
	if r.VPNRouteID() != "r-1" {
		t.Errorf("VPNRouteID() = %q", r.VPNRouteID())
	}
	if r.Name() != "my-route" {
		t.Errorf("Name() = %q", r.Name())
	}
	if tags := r.Tags(); len(tags) != 1 || tags[0] != "route-tag" {
		t.Errorf("Tags() = %v", tags)
	}
	if r.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", r.Region())
	}
	if r.State() != "Active" {
		t.Errorf("State() = %q", r.State())
	}
	if r.CloudSubnet() != "10.0.0.0/24" {
		t.Errorf("CloudSubnet() = %q", r.CloudSubnet())
	}
	if r.OnPremSubnet() != "192.168.0.0/24" {
		t.Errorf("OnPremSubnet() = %q", r.OnPremSubnet())
	}
	if r.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", r.ProjectID())
	}
	if r.vpnTunnelID != "t-1" {
		t.Errorf("vpnTunnelID = %q", r.vpnTunnelID)
	}
	if linked := r.LinkedResources(); len(linked) != 1 {
		t.Errorf("LinkedResources() len = %d", len(linked))
	}
	if r.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestVPNRoute_FromResponsePartial(t *testing.T) {
	r := &VPNRoute{}
	r.fromResponse(nil)
	if r.ID() != "" || r.URI() != "" || r.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}
	if r.Raw() != nil {
		t.Error("Raw() should be nil before hydration")
	}

	r2 := &VPNRoute{}
	r2.fromResponse(&types.VPNRouteResponse{})
	if r2.ID() != "" || r2.URI() != "" || r2.State() != "" {
		t.Error("empty response should yield zero accessor values")
	}
}

func TestVPNRoute_FromResponseURIBackfill_CamelCase(t *testing.T) {
	uri := "/projects/p2/providers/Aruba.Network/vpnTunnels/t-2/vpnRoutes/r-2"
	id := "r-2"
	name := "uri-route"
	resp := &types.VPNRouteResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:   &id,
			URI:  &uri,
			Name: &name,
			// ProjectMetadataResponse intentionally nil
		},
	}
	r := &VPNRoute{}
	r.fromResponse(resp)

	if r.ProjectID() != "p2" {
		t.Errorf("ProjectID() via URI fallback = %q", r.ProjectID())
	}
	if r.vpnTunnelID != "t-2" {
		t.Errorf("vpnTunnelID via URI fallback = %q", r.vpnTunnelID)
	}
}

// --------------------------------------------------------------------------
// vpnRouteIDsFromRef helper
// --------------------------------------------------------------------------

func TestVPNRouteIDsFromRef_TypedRef(t *testing.T) {
	r := &VPNRoute{}
	r.fromResponse(vpnRouteTestResponse("r-1", "n",
		"/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1", "p"))
	r.vpnTunnelID = "t-1"

	pid, tid, rid, err := vpnRouteIDsFromRef(r)
	if err != nil || pid != "p" || tid != "t-1" || rid != "r-1" {
		t.Errorf("vpnRouteIDsFromRef typed = (%q, %q, %q, %v)", pid, tid, rid, err)
	}
}

func TestVPNRouteIDsFromRef_URIRef_CamelCase(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1")
	pid, tid, rid, err := vpnRouteIDsFromRef(ref)
	if err != nil || pid != "p" || tid != "t-1" || rid != "r-1" {
		t.Errorf("vpnRouteIDsFromRef camelCase = (%q, %q, %q, %v)", pid, tid, rid, err)
	}
}

func TestVPNRouteIDsFromRef_BadURI_MissingRoute(t *testing.T) {
	_, _, _, err := vpnRouteIDsFromRef(URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1"))
	if err == nil {
		t.Error("expected error for URI missing route segment")
	}
}

func TestVPNRouteIDsFromRef_BadURI_MissingTunnel(t *testing.T) {
	_, _, _, err := vpnRouteIDsFromRef(URI("/projects/p/vpnRoutes/r-1"))
	if err == nil {
		t.Error("expected error for URI missing tunnel segment")
	}
}

func TestVPNRouteIDsFromRef_BadURI_MissingProject(t *testing.T) {
	_, _, _, err := vpnRouteIDsFromRef(URI("/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1"))
	if err == nil {
		t.Error("expected error for URI missing project segment")
	}
}

func TestVPNRouteIDsFromRef_BadURI_MissingAll(t *testing.T) {
	_, _, _, err := vpnRouteIDsFromRef(URI("/something/else"))
	if err == nil {
		t.Error("expected error for totally invalid URI")
	}
}

// --------------------------------------------------------------------------
// vpnRoutesClientAdapter — CRUD integration tests
// --------------------------------------------------------------------------

func buildVPNRouteTestAdapter(t *testing.T, handler http.HandlerFunc) *vpnRoutesClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newVPNRoutesClientAdapter(testutil.NewClient(t, server.URL))
}

const vpnRouteSuccessBody = `{` +
	`"metadata":{` +
	`"id":"r-1","name":"my-route",` +
	`"uri":"/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1",` +
	`"project":{"id":"p"}` +
	`},` +
	`"properties":{` +
	`"cloudSubnet":"10.0.0.0/24","onPremSubnet":"192.168.0.0/24"` +
	`},` +
	`"status":{"state":"Active"}}`

func TestVPNRoutesClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.VPNRouteRequest
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, vpnRouteSuccessBody)
	})

	route := NewVPNRoute().
		InVPNTunnel(URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1")).
		Named("my-route").
		InRegion(RegionITBGBergamo).
		WithCloudSubnet("10.0.0.0/24").
		WithOnPremSubnet("192.168.0.0/24")

	result, err := adapter.Create(context.Background(), route)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "r-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-route" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Metadata.Name != "my-route" {
		t.Errorf("request Name = %q", gotBody.Metadata.Name)
	}
	if gotBody.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("request Location = %q", gotBody.Metadata.Location.Value)
	}
	if gotBody.Properties.CloudSubnet != "10.0.0.0/24" {
		t.Errorf("request CloudSubnet = %q", gotBody.Properties.CloudSubnet)
	}
	if gotBody.Properties.OnPremSubnet != "192.168.0.0/24" {
		t.Errorf("request OnPremSubnet = %q", gotBody.Properties.OnPremSubnet)
	}
}

func TestVPNRoutesClientAdapter_Create_NoTunnel(t *testing.T) {
	callCount := 0
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	_, err := adapter.Create(context.Background(), NewVPNRoute().
		Named("x"))
	if err == nil {
		t.Fatal("expected error when route has no parent tunnel")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without parent tunnel")
	}
}

func TestVPNRoutesClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" — triggers MetadataValidationError
		fmt.Fprint(w, `{"metadata":{"name":"route","uri":"/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/x"},"properties":{},"status":{}}`)
	})

	route := NewVPNRoute().
		InVPNTunnel(URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1")).
		Named("route")

	result, err := adapter.Create(context.Background(), route)
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

func TestVPNRoutesClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "name is required", 422))
	})

	route := NewVPNRoute().
		InVPNTunnel(URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1"))
	result, err := adapter.Create(context.Background(), route)
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

func TestVPNRoutesClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, vpnRouteSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "r-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	if result.StatusCode() != http.StatusOK {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if !strings.Contains(capturedPath, "vpnRoutes") {
		t.Errorf("path = %q, expected vpnRoutes segment", capturedPath)
	}
}

func TestVPNRoutesClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, vpnRouteSuccessBody)
	})

	existing := &VPNRoute{}
	existing.fromResponse(vpnRouteTestResponse("r-99", "n",
		"/projects/p2/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-99", "p2"))
	existing.vpnTunnelID = "t-1"

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "r-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestVPNRoutesClientAdapter_Update_Success(t *testing.T) {
	var capturedBody types.VPNRouteRequest
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"r-1","name":"renamed","uri":"/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1","project":{"id":"p"}},"properties":{},"status":{}}`)
	})

	r := &VPNRoute{}
	r.fromResponse(vpnRouteTestResponse("r-1", "orig",
		"/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1", "p"))
	r.vpnTunnelID = "t-1"
	r.Named("renamed")

	result, err := adapter.Update(context.Background(), r)
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

func TestVPNRoutesClientAdapter_Update_NoID(t *testing.T) {
	callCount := 0
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	r := NewVPNRoute().
		InVPNTunnel(URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1")).
		Named("x")

	_, err := adapter.Update(context.Background(), r)
	if err == nil {
		t.Fatal("expected error when route has no ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without ID")
	}
}

func TestVPNRoutesClientAdapter_Update_NoTunnel(t *testing.T) {
	callCount := 0
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	r := &VPNRoute{}
	id := "r-1"
	r.meta = &types.ResourceMetadataResponse{ID: &id}
	// vpnTunnelID intentionally empty
	_, err := adapter.Update(context.Background(), r)
	if err == nil {
		t.Fatal("expected error when route has no parent tunnel")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without parent tunnel")
	}
}

func TestVPNRoutesClientAdapter_Delete_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestVPNRoutesClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	ref := URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1")
	if err := adapter.Delete(context.Background(), ref); err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestVPNRoutesClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "route not found", 404))
	})

	ref := URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1")
	err := adapter.Delete(context.Background(), ref)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("HTTPError.StatusCode = %d", httpErr.StatusCode)
	}
}

// InRegion exercises the 0% branch.
func TestVPNRoute_InRegion(t *testing.T) {
	r := NewVPNRoute().
		Tagged("a").
		Tagged("b").
		Untagged("a").
		RetaggedAs("x", "y").
		InRegion("ITMI-Milano-1")

	if r.Region() != "ITMI-Milano-1" {
		t.Errorf("Region() = %q", r.Region())
	}
	if tags := r.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("Tags() = %v", tags)
	}
}

func TestVPNRoutesClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "vpn route not found", 404))
	})

	ref := URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/missing")
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

func TestVPNRoutesClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "vpn route not found", 404))
	})

	r := &VPNRoute{}
	r.fromResponse(vpnRouteTestResponse("r-1", "my-route",
		"/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1", "p"))
	_, err := adapter.Update(context.Background(), r)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
}

func TestVPNRoutesClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Forbidden", "access denied", 403))
	})

	_, err := adapter.List(context.Background(),
		URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
}

func TestVPNRouteIDsFromRef_BadURI_MissingTunnelID(t *testing.T) {
	// URI has vpnRoutes but no vpnTunnels segment
	_, _, _, err := vpnRouteIDsFromRef(URI("/projects/p/vpnRoutes/route"))
	if err == nil {
		t.Error("expected error for URI without vpn tunnel segment")
	}
}

func TestVPNRouteIDsFromRef_BadURI_MissingProjectID(t *testing.T) {
	// URI has vpnRoutes+vpnTunnels but no projects
	_, _, _, err := vpnRouteIDsFromRef(URI("/providers/Aruba.Network/vpnTunnels/t/vpnRoutes/route"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

func TestVPNRoutesClientAdapter_Create_WithBuilderError(t *testing.T) {
	callCount := 0
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})
	r := NewVPNRoute().InVPNTunnel(URI("/garbage"))
	_, err := adapter.Create(context.Background(), r)
	if err == nil {
		t.Fatal("expected error for builder error")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when builder has errors")
	}
}

func TestVPNRoutesClientAdapter_Get_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestVPNRoutesClientAdapter_Get_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPNRoutesClientAdapter(testutil.NewClient(t, server.URL))
	result, err := adapter.Get(context.Background(),
		URI("/projects/p/providers/Aruba.Network/vpnTunnels/t/vpnRoutes/r"))
	if err == nil {
		t.Fatal("expected transport error")
	}
	_ = result
}

func TestVPNRoutesClientAdapter_Update_WithBuilderError(t *testing.T) {
	callCount := 0
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	r := NewVPNRoute().InVPNTunnel(URI("/garbage"))
	_, err := adapter.Update(context.Background(), r)
	if err == nil {
		t.Fatal("expected error for builder error")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when builder has errors")
	}
}

func TestVPNRoutesClientAdapter_Update_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPNRoutesClientAdapter(testutil.NewClient(t, server.URL))
	route := &VPNRoute{}
	route.fromResponse(vpnRouteTestResponse("r-1", "route-a",
		"/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1", "p"))
	_, err := adapter.Update(context.Background(), route)
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestVPNRoutesClientAdapter_Delete_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPNRoutesClientAdapter(testutil.NewClient(t, server.URL))
	err := adapter.Delete(context.Background(),
		URI("/projects/p/providers/Aruba.Network/vpnTunnels/t/vpnRoutes/r"))
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestVPNRoutesClientAdapter_List_BadTunnelRef(t *testing.T) {
	callCount := 0
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad tunnel Ref")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made for bad tunnel Ref")
	}
}

func TestVPNRoutesClientAdapter_List_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPNRoutesClientAdapter(testutil.NewClient(t, server.URL))
	_, err := adapter.List(context.Background(),
		URI("/projects/p/providers/Aruba.Network/vpnTunnels/t"))
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestVPNRoutesClientAdapter_List_AncestorIDBackfill(t *testing.T) {
	// Items without ancestor IDs in metadata/URI: triggers vpnTunnelID/projectID backfill
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":1,"values":[`+
			`{"metadata":{"id":"r-x","name":"route-x"},"properties":{"cloudSubnet":"10.0.0.0/24","onPremSubnet":"192.168.0.0/24"},"status":{}}`+
			`]}`)
	})

	list, err := adapter.List(context.Background(),
		URI("/projects/proj-x/providers/Aruba.Network/vpnTunnels/tunnel-x"))
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
	if items[0].VPNTunnelID() != "tunnel-x" {
		t.Errorf("VPNTunnelID() after backfill = %q, want %q", items[0].VPNTunnelID(), "tunnel-x")
	}
}

func TestVPNRoutesClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"values":[`+
			`{"metadata":{"id":"r-1","name":"route-1","uri":"/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1","project":{"id":"p"}},"properties":{"cloudSubnet":"10.0.0.0/24","onPremSubnet":"192.168.0.0/24"},"status":{}},`+
			`{"metadata":{"id":"r-2","name":"route-2","uri":"/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-2","project":{"id":"p"}},"properties":{"cloudSubnet":"10.1.0.0/24","onPremSubnet":"192.168.1.0/24"},"status":{}}`+
			`]}`)
	})

	tunnel := URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1")
	list, err := adapter.List(context.Background(), tunnel)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	items := list.Items()
	if len(items) != 2 {
		t.Fatalf("Items() len = %d, want 2", len(items))
	}
	if items[0].ID() != "r-1" {
		t.Errorf("items[0].ID() = %q", items[0].ID())
	}
	if items[1].ID() != "r-2" {
		t.Errorf("items[1].ID() = %q", items[1].ID())
	}
	for i, item := range items {
		if item.ProjectID() != "p" {
			t.Errorf("items[%d].ProjectID() = %q", i, item.ProjectID())
		}
		if item.vpnTunnelID != "t-1" {
			t.Errorf("items[%d].vpnTunnelID = %q", i, item.vpnTunnelID)
		}
	}
}

func TestVPNRoute_FromResponse_SetsStatus(t *testing.T) {
	r := &VPNRoute{}
	state := types.State("Active")
	r.fromResponse(&types.VPNRouteResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if r.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", r.State())
	}
}

func TestVPNRoutesClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, vpnRouteSuccessBody)
	})
	adapter := newVPNRoutesClientAdapter(testutil.NewClient(t, server.URL))
	route, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&route.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned VPNRoute")
	}
}

// vpnRouteCreateObjectSubnetBody simulates the Create response where cloudSubnet
// is returned as a full subnet resource object rather than a plain CIDR string.
const vpnRouteCreateObjectSubnetBody = `{` +
	`"metadata":{` +
	`"id":"r-1","name":"my-route",` +
	`"uri":"/projects/p/providers/Aruba.Network/vpnTunnels/t-1/vpnRoutes/r-1",` +
	`"project":{"id":"p"}` +
	`},` +
	`"properties":{` +
	`"cloudSubnet":{` +
	`"metadata":{"id":"sub-1","name":"my-subnet"},` +
	`"properties":{"network":{"address":"10.1.0.0/24"}}` +
	`},` +
	`"onPremSubnet":"192.168.129.0/24"` +
	`},` +
	`"status":{"state":"Active"}}`

func TestVPNRoutesClientAdapter_Create_CloudSubnetAsObject(t *testing.T) {
	adapter := buildVPNRouteTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, vpnRouteCreateObjectSubnetBody)
	})

	route := NewVPNRoute().
		InVPNTunnel(URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1")).
		Named("my-route").
		InRegion(RegionITBGBergamo).
		WithCloudSubnet("10.1.0.0/24").
		WithOnPremSubnet("192.168.129.0/24")

	result, err := adapter.Create(context.Background(), route)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.CloudSubnet() != "10.1.0.0/24" {
		t.Errorf("CloudSubnet() = %q, want %q", result.CloudSubnet(), "10.1.0.0/24")
	}
	if result.OnPremSubnet() != "192.168.129.0/24" {
		t.Errorf("OnPremSubnet() = %q, want %q", result.OnPremSubnet(), "192.168.129.0/24")
	}
}

func TestVPNRouteRef(t *testing.T) {
	ref := VPNRouteRef("p-1", "tun-1", "rt-1")
	want := "/projects/p-1/providers/Aruba.Network/vpnTunnels/tun-1/vpnRoutes/rt-1"
	if ref.URI() != want {
		t.Errorf("VPNRouteRef URI = %q, want %q", ref.URI(), want)
	}
	ids := parseURIIDs(ref.URI())
	if ids["projects"] != "p-1" || ids["vpnTunnels"] != "tun-1" || ids["vpnRoutes"] != "rt-1" {
		t.Errorf("parseURIIDs = %v", ids)
	}
}

// --------------------------------------------------------------------------
// VPNTunnelURI and CloudSubnetCIDR getters
// --------------------------------------------------------------------------

func TestVPNRoute_VPNTunnelURI_NilResponse(t *testing.T) {
	r := &VPNRoute{}
	if got := r.VPNTunnelURI(); got != "" {
		t.Errorf("VPNTunnelURI() on nil response = %q, want \"\"", got)
	}
}

func TestVPNRoute_VPNTunnelURI_NilTunnel(t *testing.T) {
	r := &VPNRoute{}
	r.response = &types.VPNRouteResponse{}
	if got := r.VPNTunnelURI(); got != "" {
		t.Errorf("VPNTunnelURI() with nil VPNTunnel = %q, want \"\"", got)
	}
}

func TestVPNRoute_VPNTunnelURI_Populated(t *testing.T) {
	tunnelURI := "/projects/p/providers/Aruba.Network/vpnTunnels/t-1"
	r := &VPNRoute{}
	r.response = &types.VPNRouteResponse{
		Properties: types.VPNRoutePropertiesResponse{
			VPNTunnel: &types.ReferenceResourceCommon{URI: tunnelURI},
		},
	}
	if got := r.VPNTunnelURI(); got != tunnelURI {
		t.Errorf("VPNTunnelURI() = %q, want %q", got, tunnelURI)
	}
}

func TestVPNRoute_CloudSubnetCIDR_Alias(t *testing.T) {
	r := NewVPNRoute().WithCloudSubnet("10.1.2.0/24")
	if got := r.CloudSubnetCIDR(); got != "10.1.2.0/24" {
		t.Errorf("CloudSubnetCIDR() = %q, want %q", got, "10.1.2.0/24")
	}
}

func TestVPNRoute_CloudSubnetCIDR_NilResponse(t *testing.T) {
	r := &VPNRoute{}
	if got := r.CloudSubnetCIDR(); got != "" {
		t.Errorf("CloudSubnetCIDR() on nil = %q, want \"\"", got)
	}
}
