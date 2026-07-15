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

var _ Ref = (*VPCPeering)(nil)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestVPCPeering_FluentSetters(t *testing.T) {
	parent := &VPC{}
	parent.fromResponse(vpcTestResponse("v1", "my-vpc", "/projects/p1/providers/Aruba.Network/vpcs/v1", "p1"))

	p := NewVPCPeering().
		InVPC(parent).
		Named("my-peering").
		Tagged("peering").
		Tagged("cross-vpc").
		Tagged("peering"). // dedupe
		InRegion(RegionITBGBergamo).
		PeeredWith(URI("/projects/p2/network/vpcs/v2"))

	if p.Name() != "my-peering" {
		t.Errorf("Name() = %q", p.Name())
	}
	if tags := p.Tags(); len(tags) != 2 || tags[0] != "peering" || tags[1] != "cross-vpc" {
		t.Errorf("Tags() = %v", tags)
	}
	if p.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", p.Region())
	}
	if p.RemoteVPCURI() != "/projects/p2/network/vpcs/v2" {
		t.Errorf("RemoteVPCURI() = %q", p.RemoteVPCURI())
	}
	if p.VPCID() != "v1" {
		t.Errorf("VPCID() = %q", p.VPCID())
	}
	if p.ProjectID() != "p1" {
		t.Errorf("ProjectID() = %q", p.ProjectID())
	}
	if p.Err() != nil {
		t.Errorf("Err() = %v", p.Err())
	}

	p.Untagged("peering")
	if tags := p.Tags(); len(tags) != 1 || tags[0] != "cross-vpc" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	p.RetaggedAs("x", "y")
	if tags := p.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

// --------------------------------------------------------------------------
// IntoVPC — typed Ref
// --------------------------------------------------------------------------

func TestVPCPeering_IntoVPC_TypedRef(t *testing.T) {
	parent := &VPC{}
	parent.fromResponse(vpcTestResponse("v1", "my-vpc", "/projects/p1/providers/Aruba.Network/vpcs/v1", "p1"))

	p := NewVPCPeering().InVPC(parent)

	if p.VPCID() != "v1" {
		t.Errorf("VPCID() = %q", p.VPCID())
	}
	if p.ProjectID() != "p1" {
		t.Errorf("ProjectID() = %q", p.ProjectID())
	}
	if p.Err() != nil {
		t.Errorf("Err() = %v", p.Err())
	}
}

// --------------------------------------------------------------------------
// IntoVPC — URI Ref
// --------------------------------------------------------------------------

func TestVPCPeering_IntoVPC_URIRef(t *testing.T) {
	p := NewVPCPeering().InVPC(URI("/projects/p/network/vpcs/v"))

	if p.VPCID() != "v" {
		t.Errorf("VPCID() = %q", p.VPCID())
	}
	if p.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", p.ProjectID())
	}
	if p.Err() != nil {
		t.Errorf("Err() = %v", p.Err())
	}
}

// --------------------------------------------------------------------------
// IntoVPC — bad Ref
// --------------------------------------------------------------------------

func TestVPCPeering_IntoVPC_BadRef(t *testing.T) {
	p := NewVPCPeering().InVPC(URI("/garbage"))
	if p.Err() == nil {
		t.Error("expected Err() != nil for unresolvable Ref, got nil")
	}
}

// --------------------------------------------------------------------------
// WithRemoteVPC
// --------------------------------------------------------------------------

func TestVPCPeering_WithRemoteVPC_TypedRef(t *testing.T) {
	remote := &VPC{}
	remote.fromResponse(vpcTestResponse("v2", "remote-vpc", "/projects/p2/providers/Aruba.Network/vpcs/v2", "p2"))

	p := NewVPCPeering().PeeredWith(remote)

	if p.RemoteVPCURI() != "/projects/p2/providers/Aruba.Network/vpcs/v2" {
		t.Errorf("RemoteVPCURI() = %q", p.RemoteVPCURI())
	}
	if p.Err() != nil {
		t.Errorf("Err() = %v", p.Err())
	}
}

func TestVPCPeering_WithRemoteVPC_URIRef(t *testing.T) {
	p := NewVPCPeering().PeeredWith(URI("/projects/p2/network/vpcs/v2"))

	if p.RemoteVPCURI() != "/projects/p2/network/vpcs/v2" {
		t.Errorf("RemoteVPCURI() = %q", p.RemoteVPCURI())
	}
	if p.Err() != nil {
		t.Errorf("Err() = %v", p.Err())
	}
}

func TestVPCPeering_WithRemoteVPC_EmptyURI(t *testing.T) {
	p := NewVPCPeering().PeeredWith(URI(""))

	if p.Err() == nil {
		t.Fatal("expected Err() != nil for empty URI remote VPC")
	}
	if !strings.Contains(p.Err().Error(), "empty URI") {
		t.Errorf("error = %q, expected 'empty URI'", p.Err().Error())
	}
	if p.RemoteVPCURI() != "" {
		t.Errorf("RemoteVPCURI() = %q, expected empty (setter must not apply on error)", p.RemoteVPCURI())
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestVPCPeering_ToRequestRoundTrip(t *testing.T) {
	p := NewVPCPeering().Named(
		"my-peering").
		Tagged("t1").
		Tagged("t2").
		InRegion(RegionITBGBergamo).
		PeeredWith(URI("/projects/p2/network/vpcs/v2"))

	req := p.RawRequest()

	if req.Metadata.Name != "my-peering" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if len(req.Metadata.Tags) != 2 {
		t.Errorf("Metadata.Tags = %v", req.Metadata.Tags)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Metadata.Location.Value = %q", req.Metadata.Location.Value)
	}
	if req.Properties.RemoteVPC == nil || req.Properties.RemoteVPC.URI != "/projects/p2/network/vpcs/v2" {
		t.Errorf("Properties.RemoteVPC = %v", req.Properties.RemoteVPC)
	}

	// Unset RemoteVPC must produce nil pointer (omitempty).
	p2 := NewVPCPeering().
		Named("bare")
	req2 := p2.RawRequest()
	if req2.Properties.RemoteVPC != nil {
		t.Errorf("Properties.RemoteVPC should be nil when not set, got %v", req2.Properties.RemoteVPC)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func vpcPeeringTestResponse(id, name, uri, projectID string) *types.VPCPeeringResponse {
	state := types.State("Active")
	loc := &types.LocationResponse{Value: RegionITBGBergamo}
	remoteURI := "/projects/p2/providers/Aruba.Network/vpcs/v2"
	return &types.VPCPeeringResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			URI:              &uri,
			Name:             &name,
			Tags:             []string{"peer-tag"},
			LocationResponse: loc,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{
				ID: projectID,
			},
		},
		Properties: types.VPCPeeringPropertiesResponse{
			LinkedResources: []types.LinkedResourceCommon{
				{URI: "/projects/p/providers/Aruba.Compute/cloudservers/cs1"},
			},
			RemoteVPC: &types.ReferenceResourceCommon{URI: remoteURI},
		},
		Status: types.ResourceStatusResponse{
			State: &state,
		},
	}
}

func TestVPCPeering_FromResponseHydration(t *testing.T) {
	p := &VPCPeering{}
	resp := vpcPeeringTestResponse("peer-1", "my-peering",
		"/projects/p1/providers/Aruba.Network/vpcs/v1/vpcPeerings/peer-1", "p1")
	p.fromResponse(resp)

	if p.ID() != "peer-1" {
		t.Errorf("ID() = %q", p.ID())
	}
	if p.URI() != "/projects/p1/providers/Aruba.Network/vpcs/v1/vpcPeerings/peer-1" {
		t.Errorf("URI() = %q", p.URI())
	}
	if p.VPCPeeringID() != "peer-1" {
		t.Errorf("VPCPeeringID() = %q", p.VPCPeeringID())
	}
	if p.Name() != "my-peering" {
		t.Errorf("Name() = %q", p.Name())
	}
	if tags := p.Tags(); len(tags) != 1 || tags[0] != "peer-tag" {
		t.Errorf("Tags() = %v", tags)
	}
	if p.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", p.Region())
	}
	if p.State() != "Active" {
		t.Errorf("State() = %q", p.State())
	}
	if linked := p.LinkedResources(); len(linked) != 1 {
		t.Errorf("LinkedResources() len = %d", len(linked))
	}
	if p.RemoteVPCURI() != "/projects/p2/providers/Aruba.Network/vpcs/v2" {
		t.Errorf("RemoteVPCURI() = %q", p.RemoteVPCURI())
	}
	if p.ProjectID() != "p1" {
		t.Errorf("ProjectID() = %q", p.ProjectID())
	}
	if p.VPCID() != "v1" {
		t.Errorf("VPCID() via URI fallback = %q", p.VPCID())
	}
	if p.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestVPCPeering_FromResponsePartial(t *testing.T) {
	// nil response is a no-op
	p := &VPCPeering{}
	p.fromResponse(nil)
	if p.ID() != "" || p.URI() != "" || p.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}
	if p.Raw() != nil {
		t.Error("Raw() should be nil before hydration")
	}

	// empty response — accessors must not panic; zero values expected
	p2 := &VPCPeering{}
	p2.fromResponse(&types.VPCPeeringResponse{})
	if p2.ID() != "" || p2.URI() != "" || p2.State() != "" {
		t.Error("empty response should yield zero accessor values")
	}
	if p2.RemoteVPCURI() != "" {
		t.Errorf("RemoteVPCURI() from empty response = %q", p2.RemoteVPCURI())
	}
}

func TestVPCPeering_FromResponseURIBackfill(t *testing.T) {
	uri := "/projects/p2/providers/Aruba.Network/vpcs/v2/vpcPeerings/peer-2"
	id := "peer-2"
	name := "uri-peering"
	resp := &types.VPCPeeringResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:   &id,
			URI:  &uri,
			Name: &name,
			// ProjectMetadataResponse intentionally nil
		},
	}
	p := &VPCPeering{}
	p.fromResponse(resp)

	if p.ProjectID() != "p2" {
		t.Errorf("ProjectID() via URI fallback = %q", p.ProjectID())
	}
	if p.VPCID() != "v2" {
		t.Errorf("VPCID() via URI fallback = %q", p.VPCID())
	}
}

// --------------------------------------------------------------------------
// Ref + ancestor ID satisfaction (runtime)
// --------------------------------------------------------------------------

func TestVPCPeering_RefSatisfaction(t *testing.T) {
	p := &VPCPeering{}
	p.fromResponse(vpcPeeringTestResponse("peer-99", "n",
		"/projects/p99/providers/Aruba.Network/vpcs/v99/vpcPeerings/peer-99", "p99"))

	// withVPCPeeringID typed path
	pid, ok := extractID(p, func(r Ref) (string, bool) {
		if w, ok := r.(withVPCPeeringID); ok {
			return w.VPCPeeringID(), true
		}
		return "", false
	}, "vpcPeerings")
	if !ok || pid != "peer-99" {
		t.Errorf("extractID via withVPCPeeringID = (%q, %v)", pid, ok)
	}

	// withVPCID typed path
	vid, ok := extractID(p, func(r Ref) (string, bool) {
		if w, ok := r.(withVPCID); ok {
			return w.VPCID(), true
		}
		return "", false
	}, "vpcs")
	if !ok || vid != "v99" {
		t.Errorf("extractID via withVPCID = (%q, %v)", vid, ok)
	}

	// withProjectID typed path
	projID, ok := extractID(p, func(r Ref) (string, bool) {
		if w, ok := r.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || projID != "p99" {
		t.Errorf("extractID via withProjectID = (%q, %v)", projID, ok)
	}
}

// --------------------------------------------------------------------------
// vpcPeeringIDsFromRef helper
// --------------------------------------------------------------------------

func TestVPCPeeringIDsFromRef_TypedRef(t *testing.T) {
	p := &VPCPeering{}
	p.fromResponse(vpcPeeringTestResponse("peer-1", "n",
		"/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer-1", "p"))
	pid, vid, peerid, err := vpcPeeringIDsFromRef(p)
	if err != nil || pid != "p" || vid != "v" || peerid != "peer-1" {
		t.Errorf("vpcPeeringIDsFromRef typed = (%q, %q, %q, %v)", pid, vid, peerid, err)
	}
}

func TestVPCPeeringIDsFromRef_URIRef_CamelCase(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer-1")
	pid, vid, peerid, err := vpcPeeringIDsFromRef(ref)
	if err != nil || pid != "p" || vid != "v" || peerid != "peer-1" {
		t.Errorf("vpcPeeringIDsFromRef camelCase = (%q, %q, %q, %v)", pid, vid, peerid, err)
	}
}

func TestVPCPeeringIDsFromRef_BadURI_MissingPeering(t *testing.T) {
	_, _, _, err := vpcPeeringIDsFromRef(URI("/projects/p/network/vpcs/v"))
	if err == nil {
		t.Error("expected error for URI without peering segment")
	}
}

func TestVPCPeeringIDsFromRef_BadURI_MissingAll(t *testing.T) {
	_, _, _, err := vpcPeeringIDsFromRef(URI("/something/else"))
	if err == nil {
		t.Error("expected error for totally invalid URI")
	}
}

// --------------------------------------------------------------------------
// vpcPeeringsClientAdapter — CRUD integration tests
// --------------------------------------------------------------------------

func buildVPCPeeringTestAdapter(t *testing.T, handler http.HandlerFunc) *vpcPeeringsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newVPCPeeringsClientAdapter(testutil.NewClient(t, server.URL))
}

const vpcPeeringSuccessBody = `{` +
	`"metadata":{"id":"peer-1","name":"my-peering","uri":"/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer-1","project":{"id":"p"}},` +
	`"properties":{"remoteVpc":{"uri":"/projects/p2/providers/Aruba.Network/vpcs/v2"}},` +
	`"status":{"state":"Active"}}`

func TestVPCPeeringsClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.VPCPeeringRequest
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, vpcPeeringSuccessBody)
	})

	vpc := &VPC{}
	vpc.fromResponse(vpcTestResponse("v", "my-vpc", "/projects/p/providers/Aruba.Network/vpcs/v", "p"))

	p := NewVPCPeering().
		InVPC(vpc).
		Named("my-peering").
		InRegion(RegionITBGBergamo).
		PeeredWith(URI("/projects/p2/network/vpcs/v2"))

	result, err := adapter.Create(context.Background(), p)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "peer-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-peering" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Metadata.Name != "my-peering" {
		t.Errorf("request Name = %q", gotBody.Metadata.Name)
	}
	if gotBody.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("request Location = %q", gotBody.Metadata.Location.Value)
	}
	if gotBody.Properties.RemoteVPC == nil || gotBody.Properties.RemoteVPC.URI != "/projects/p2/network/vpcs/v2" {
		t.Errorf("request RemoteVPC = %v", gotBody.Properties.RemoteVPC)
	}
}

func TestVPCPeeringsClientAdapter_Create_NoVPC(t *testing.T) {
	callCount := 0
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	_, err := adapter.Create(context.Background(), NewVPCPeering().
		Named("x"))
	if err == nil {
		t.Fatal("expected error when peering has no VPC")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without VPC")
	}
}

func TestVPCPeeringsClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" field — triggers MetadataValidationError
		fmt.Fprint(w, `{"metadata":{"name":"peering","uri":"/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/x"},"properties":{},"status":{}}`)
	})

	vpc := &VPC{}
	vpc.fromResponse(vpcTestResponse("v", "my-vpc", "/projects/p/providers/Aruba.Network/vpcs/v", "p"))

	p := NewVPCPeering().InVPC(vpc).
		Named("peering")
	result, err := adapter.Create(context.Background(), p)
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

func TestVPCPeeringsClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "name is required", 422))
	})

	vpc := &VPC{}
	vpc.fromResponse(vpcTestResponse("v", "my-vpc", "/projects/p/providers/Aruba.Network/vpcs/v", "p"))

	p := NewVPCPeering().InVPC(vpc)
	result, err := adapter.Create(context.Background(), p)
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

func TestVPCPeeringsClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, vpcPeeringSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "peer-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	if result.VPCID() != "v" {
		t.Errorf("VPCID() = %q", result.VPCID())
	}
	if result.StatusCode() != http.StatusOK {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if !strings.Contains(capturedPath, "vpcPeerings") {
		t.Errorf("path = %q, expected vpcPeerings segment", capturedPath)
	}
}

func TestVPCPeeringsClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, vpcPeeringSuccessBody)
	})

	existing := &VPCPeering{}
	existing.fromResponse(vpcPeeringTestResponse("peer-1", "n",
		"/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer-1", "p"))

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "peer-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestVPCPeeringsClientAdapter_Update_Success(t *testing.T) {
	var capturedBody types.VPCPeeringRequest
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"peer-1","name":"renamed","uri":"/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer-1","project":{"id":"p"}},"properties":{},"status":{}}`)
	})

	p := &VPCPeering{}
	p.fromResponse(vpcPeeringTestResponse("peer-1", "orig",
		"/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer-1", "p"))
	p.Named("renamed")

	result, err := adapter.Update(context.Background(), p)
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

func TestVPCPeeringsClientAdapter_Update_NoID(t *testing.T) {
	callCount := 0
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	p := NewVPCPeering().InVPC(URI("/projects/p/network/vpcs/v")).
		Named("x")
	_, err := adapter.Update(context.Background(), p)
	if err == nil {
		t.Fatal("expected error when peering has no ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when ID is missing")
	}
}

func TestVPCPeeringsClientAdapter_Update_NoVPC(t *testing.T) {
	callCount := 0
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	p := &VPCPeering{}
	id := "peer-1"
	p.fromResponse(&types.VPCPeeringResponse{
		Metadata: types.ResourceMetadataResponse{
			ID: &id,
		},
	})

	_, err := adapter.Update(context.Background(), p)
	if err == nil {
		t.Fatal("expected error when peering has no VPC")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without VPC")
	}
}

func TestVPCPeeringsClientAdapter_Delete_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestVPCPeeringsClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer-1"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestVPCPeeringsClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "vpc peering not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/missing"))
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

// InRegion exercises the 0% branch.
func TestVPCPeering_InRegion(t *testing.T) {
	p := NewVPCPeering().
		Tagged("a").
		Tagged("b").
		Untagged("a").
		RetaggedAs("x", "y").
		InRegion("ITMI-Milano-1")

	if p.Region() != "ITMI-Milano-1" {
		t.Errorf("Region() = %q", p.Region())
	}
	if tags := p.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("Tags() = %v", tags)
	}
}

func TestVPCPeeringsClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "vpc peering not found", 404))
	})

	ref := URI("/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/missing")
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

func TestVPCPeeringsClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "vpc peering not found", 404))
	})

	p := &VPCPeering{}
	p.fromResponse(vpcPeeringTestResponse("peer-1", "my-peering",
		"/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer-1", "p"))
	_, err := adapter.Update(context.Background(), p)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
}

func TestVPCPeeringsClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestVPCPeeringIDsFromRef_BadURI_MissingVPC(t *testing.T) {
	// URI has vpcPeerings but no vpcs segment
	_, _, _, err := vpcPeeringIDsFromRef(URI("/projects/p/vpcPeerings/peer"))
	if err == nil {
		t.Error("expected error for URI without /vpcs/<id>")
	}
}

func TestVPCPeeringIDsFromRef_BadURI_MissingProject(t *testing.T) {
	// URI has vpcPeerings+vpcs but no projects
	_, _, _, err := vpcPeeringIDsFromRef(URI("/providers/Aruba.Network/vpcs/v/vpcPeerings/peer"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

func TestVPCPeeringsClientAdapter_Create_WithBuilderError(t *testing.T) {
	callCount := 0
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})
	peering := NewVPCPeering().InVPC(URI("/garbage"))
	_, err := adapter.Create(context.Background(), peering)
	if err == nil {
		t.Fatal("expected error for builder error")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when builder has errors")
	}
}

func TestVPCPeeringsClientAdapter_Get_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestVPCPeeringsClientAdapter_Get_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPCPeeringsClientAdapter(testutil.NewClient(t, server.URL))
	result, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer"))
	if err == nil {
		t.Fatal("expected transport error")
	}
	_ = result
}

func TestVPCPeeringsClientAdapter_Update_WithBuilderError(t *testing.T) {
	callCount := 0
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	peering := NewVPCPeering().InVPC(URI("/garbage"))
	_, err := adapter.Update(context.Background(), peering)
	if err == nil {
		t.Fatal("expected error for builder error")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when builder has errors")
	}
}

func TestVPCPeeringsClientAdapter_Update_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPCPeeringsClientAdapter(testutil.NewClient(t, server.URL))
	peering := &VPCPeering{}
	peering.fromResponse(vpcPeeringTestResponse("peer-1", "peering-a", "/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer-1", "p"))
	_, err := adapter.Update(context.Background(), peering)
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestVPCPeeringsClientAdapter_Delete_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPCPeeringsClientAdapter(testutil.NewClient(t, server.URL))
	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer"))
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestVPCPeeringsClientAdapter_List_BadVPCRef(t *testing.T) {
	callCount := 0
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestVPCPeeringsClientAdapter_List_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPCPeeringsClientAdapter(testutil.NewClient(t, server.URL))
	_, err := adapter.List(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v"))
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestVPCPeeringsClientAdapter_List_AncestorIDBackfill(t *testing.T) {
	// Items without ancestor IDs in metadata/URI: triggers vpcID/projectID backfill
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":1,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"peer-x","name":"peering-x"},"properties":{},"status":{}}`+
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
	if items[0].VPCID() != "vpc-x" {
		t.Errorf("VPCID() after backfill = %q, want %q", items[0].VPCID(), "vpc-x")
	}
}

func TestVPCPeeringsClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildVPCPeeringTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"peer-1","name":"peering-a","uri":"/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer-1","project":{"id":"p"}},"properties":{},"status":{"state":"Active"}},`+
			`{"metadata":{"id":"peer-2","name":"peering-b","uri":"/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer-2","project":{"id":"p"}},"properties":{},"status":{"state":"Inactive"}}`+
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
	if items[0].ID() != "peer-1" || items[0].Name() != "peering-a" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[1].ID() != "peer-2" || items[1].State() != "Inactive" {
		t.Errorf("items[1] ID=%q State=%q", items[1].ID(), items[1].State())
	}
	if items[0].VPCID() != "v" {
		t.Errorf("items[0].VPCID() = %q", items[0].VPCID())
	}
	if items[0].ProjectID() != "p" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
}

func TestVPCPeering_FromResponse_SetsStatus(t *testing.T) {
	p := &VPCPeering{}
	state := types.State("Active")
	p.fromResponse(&types.VPCPeeringResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if p.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", p.State())
	}
}

func TestVPCPeeringsClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, vpcPeeringSuccessBody)
	})
	adapter := newVPCPeeringsClientAdapter(testutil.NewClient(t, server.URL))
	peering, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v/vpcPeerings/peer-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&peering.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned VPCPeering")
	}
}

func TestVPCPeeringRef(t *testing.T) {
	ref := VPCPeeringRef("p-1", "vpc-1", "peer-1")
	want := "/projects/p-1/providers/Aruba.Network/vpcs/vpc-1/vpcPeerings/peer-1"
	if ref.URI() != want {
		t.Errorf("VPCPeeringRef URI = %q, want %q", ref.URI(), want)
	}
	ids := parseURIIDs(ref.URI())
	if ids["projects"] != "p-1" || ids["vpcs"] != "vpc-1" || ids["vpcPeerings"] != "peer-1" {
		t.Errorf("parseURIIDs = %v", ids)
	}
}
