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

var _ Ref = (*VPC)(nil)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestVPC_FluentSetters(t *testing.T) {
	// Seed a parent project via fromResponse so IntoProject can extract ID.
	parent := &Project{}
	parent.fromResponse(projectTestResponse("proj-1", "my-proj", "/projects/proj-1"))

	v := NewVPC().
		InProject(parent).
		Named("my-vpc").
		Tagged("net").
		Tagged("infra").
		Tagged("net"). // dedupe
		InRegion(RegionITBGBergamo).
		AsDefault().
		WithPreset()

	if v.Name() != "my-vpc" {
		t.Errorf("Name() = %q", v.Name())
	}
	if tags := v.Tags(); len(tags) != 2 || tags[0] != "net" || tags[1] != "infra" {
		t.Errorf("Tags() = %v", tags)
	}
	if v.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", v.Region())
	}
	if !v.IsDefault() {
		t.Error("IsDefault() should be true")
	}
	if !v.IsPreset() {
		t.Error("IsPreset() should be true")
	}
	if v.ProjectID() != "proj-1" {
		t.Errorf("ProjectID() = %q", v.ProjectID())
	}
	if v.Err() != nil {
		t.Errorf("Err() = %v", v.Err())
	}

	v.Untagged("net")
	if tags := v.Tags(); len(tags) != 1 || tags[0] != "infra" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	v.RetaggedAs("x", "y")
	if tags := v.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

// --------------------------------------------------------------------------
// IntoProject with bad Ref — must not panic, must surface error via Err()
// --------------------------------------------------------------------------

func TestVPC_IntoProject_BadRef(t *testing.T) {
	v := NewVPC().InProject(URI("/garbage"))
	if v.Err() == nil {
		t.Error("expected Err() != nil for unresolvable Ref, got nil")
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestVPC_ToRequestRoundTrip(t *testing.T) {
	v := NewVPC().Named(
		"vpc-1").
		Tagged("t1").
		Tagged("t2").
		InRegion(RegionITBGBergamo).
		AsDefault().
		WithoutPreset()

	req := v.RawRequest()

	if req.Metadata.Name != "vpc-1" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if len(req.Metadata.Tags) != 2 {
		t.Errorf("Metadata.Tags = %v", req.Metadata.Tags)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Location.Value = %q", req.Metadata.Location.Value)
	}
	if req.Properties.Properties == nil {
		t.Fatal("Properties.Properties should be non-nil when Default or Preset set")
	}
	if req.Properties.Properties.Default == nil || !*req.Properties.Properties.Default {
		t.Error("Properties.Properties.Default should be true")
	}
	if req.Properties.Properties.Preset == nil || *req.Properties.Properties.Preset {
		t.Error("Properties.Properties.Preset should be false")
	}

	// New construction default: Default is explicitly set to false, so
	// Properties.Properties is non-nil and Default points to false.
	v2 := NewVPC().
		Named("bare")
	req2 := v2.RawRequest()
	if req2.Properties.Properties == nil {
		t.Fatal("Properties.Properties should be non-nil after construction (Default defaults to false)")
	}
	if req2.Properties.Properties.Default == nil || *req2.Properties.Properties.Default {
		t.Errorf("Default should default to *false, got %v", req2.Properties.Properties.Default)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func vpcTestResponse(id, name, uri, projectID string) *types.VPCResponse {
	loc := &types.LocationResponse{Value: RegionITBGBergamo}
	state := types.State("Active")
	return &types.VPCResponse{
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
		Properties: types.VPCPropertiesResponse{
			Default: true,
			LinkedResources: []types.LinkedResourceCommon{
				{URI: "/projects/p/network/vpcs/v/subnets/s1", StrictCorrelation: true},
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

func TestVPC_FromResponseHydration(t *testing.T) {
	v := &VPC{}
	resp := vpcTestResponse("vpc-1", "my-vpc", "/projects/p/providers/Aruba.Network/vpcs/vpc-1", "p")
	v.fromResponse(resp)

	if v.ID() != "vpc-1" {
		t.Errorf("ID() = %q", v.ID())
	}
	if v.URI() != "/projects/p/providers/Aruba.Network/vpcs/vpc-1" {
		t.Errorf("URI() = %q", v.URI())
	}
	if v.VPCID() != "vpc-1" {
		t.Errorf("VPCID() = %q", v.VPCID())
	}
	if v.Name() != "my-vpc" {
		t.Errorf("Name() = %q", v.Name())
	}
	if tags := v.Tags(); len(tags) != 1 || tags[0] != "tag1" {
		t.Errorf("Tags() = %v", tags)
	}
	if v.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", v.Region())
	}
	if v.State() != "Active" {
		t.Errorf("State() = %q", v.State())
	}
	if v.IsDisabled() {
		t.Error("IsDisabled() should be false")
	}
	if linked := v.LinkedResources(); len(linked) != 1 {
		t.Errorf("LinkedResources() len = %d", len(linked))
	}
	if !v.IsDefault() {
		t.Error("IsDefault() should be true")
	}
	if v.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", v.ProjectID())
	}
	if v.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestVPC_FromResponsePartial(t *testing.T) {
	// nil response is a no-op
	v := &VPC{}
	v.fromResponse(nil)
	if v.ID() != "" || v.URI() != "" || v.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}

	// response with empty inner fields — accessors must not panic
	v2 := &VPC{}
	v2.fromResponse(&types.VPCResponse{})
	if v2.ID() != "" || v2.URI() != "" || v2.State() != "" {
		t.Error("empty response should yield zero accessor values")
	}
}

// --------------------------------------------------------------------------
// Ref + ancestor ID satisfaction (runtime)
// --------------------------------------------------------------------------

func TestVPC_RefSatisfaction(t *testing.T) {
	v := &VPC{}
	v.fromResponse(vpcTestResponse("vpc-99", "n", "/projects/p99/providers/Aruba.Network/vpcs/vpc-99", "p99"))

	// withVPCID typed path
	vpcID, ok := extractID(v, func(r Ref) (string, bool) {
		if w, ok := r.(withVPCID); ok {
			return w.VPCID(), true
		}
		return "", false
	}, "vpcs")
	if !ok || vpcID != "vpc-99" {
		t.Errorf("extractID via withVPCID = (%q, %v)", vpcID, ok)
	}

	// withProjectID typed path
	projectID, ok := extractID(v, func(r Ref) (string, bool) {
		if w, ok := r.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || projectID != "p99" {
		t.Errorf("extractID via withProjectID = (%q, %v)", projectID, ok)
	}

	// A child wrapper using vpcScopedMixin can inherit both IDs from the VPC.
	child := bindVPCScoped(&errMixin{})
	child.intoVPC(v)
	if child.VPCID() != "vpc-99" {
		t.Errorf("child VPCID() = %q", child.VPCID())
	}
	if child.ProjectID() != "p99" {
		t.Errorf("child ProjectID() = %q", child.ProjectID())
	}
}

// --------------------------------------------------------------------------
// vpcIDsFromRef helper
// --------------------------------------------------------------------------

func TestVPCIDsFromRef_TypedRef(t *testing.T) {
	v := &VPC{}
	v.fromResponse(vpcTestResponse("vid", "n", "/projects/p/providers/Aruba.Network/vpcs/vid", "p"))
	pid, vid, err := vpcIDsFromRef(v)
	if err != nil || pid != "p" || vid != "vid" {
		t.Errorf("vpcIDsFromRef typed = (%q, %q, %v)", pid, vid, err)
	}
}

func TestVPCIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Network/vpcs/v")
	pid, vid, err := vpcIDsFromRef(ref)
	if err != nil || pid != "p" || vid != "v" {
		t.Errorf("vpcIDsFromRef URI = (%q, %q, %v)", pid, vid, err)
	}
}

func TestVPCIDsFromRef_BadURI(t *testing.T) {
	_, _, err := vpcIDsFromRef(URI("/something/else"))
	if err == nil {
		t.Error("expected error for URI without /vpcs/<id>")
	}
}

// --------------------------------------------------------------------------
// vpcsClientAdapter — CRUD integration tests
// --------------------------------------------------------------------------

func buildVPCTestAdapter(t *testing.T, handler http.HandlerFunc) *vpcsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newVPCsClientAdapter(testutil.NewClient(t, server.URL))
}

func TestVPCsClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.VPCRequest
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"metadata":{"id":"vid","name":"my-vpc","uri":"/projects/p/providers/Aruba.Network/vpcs/vid","project":{"id":"p"}},"properties":{"default":false},"status":{"state":"Creating"}}`)
	})

	vpc := NewVPC().
		InProject(URI("/projects/p")).
		Named("my-vpc").
		NotDefault()

	result, err := adapter.Create(context.Background(), vpc)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "vid" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-vpc" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Metadata.Name != "my-vpc" {
		t.Errorf("request Name = %q", gotBody.Metadata.Name)
	}
}

func TestVPCsClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	_, err := adapter.Create(context.Background(), NewVPC().
		Named("x"))
	if err == nil {
		t.Fatal("expected error when VPC has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestVPCsClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" field — triggers MetadataValidationError
		fmt.Fprint(w, `{"metadata":{"name":"v","uri":"/projects/p/providers/Aruba.Network/vpcs/x"},"properties":{},"status":{}}`)
	})

	vpc := NewVPC().InProject(URI("/projects/p")).
		Named("v")
	result, err := adapter.Create(context.Background(), vpc)
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

func TestVPCsClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "name is required", 422))
	})

	vpc := NewVPC().InProject(URI("/projects/p"))
	result, err := adapter.Create(context.Background(), vpc)
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

func TestVPCsClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"v","name":"got","uri":"/projects/p/providers/Aruba.Network/vpcs/v","project":{"id":"p"}},"properties":{"default":false},"status":{}}`)
	})

	result, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "v" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "got" {
		t.Errorf("Name() = %q", result.Name())
	}
	wantPath := "/projects/p/providers/Aruba.Network/vpcs/v"
	if capturedPath != wantPath {
		t.Errorf("path = %q, want %q", capturedPath, wantPath)
	}
}

func TestVPCsClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"vid","name":"n","uri":"/projects/p/providers/Aruba.Network/vpcs/vid","project":{"id":"p"}},"properties":{"default":false},"status":{}}`)
	})

	v := &VPC{}
	v.fromResponse(vpcTestResponse("vid", "n", "/projects/p/providers/Aruba.Network/vpcs/vid", "p"))

	result, err := adapter.Get(context.Background(), v)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "vid" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestVPCsClientAdapter_Update_Success(t *testing.T) {
	var capturedBody types.VPCRequest
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"vid","name":"renamed","uri":"/projects/p/providers/Aruba.Network/vpcs/vid","project":{"id":"p"}},"properties":{"default":false},"status":{}}`)
	})

	v := &VPC{}
	v.fromResponse(vpcTestResponse("vid", "orig", "/projects/p/providers/Aruba.Network/vpcs/vid", "p"))
	v.Named("renamed")

	result, err := adapter.Update(context.Background(), v)
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

func TestVPCsClientAdapter_Update_NoID(t *testing.T) {
	callCount := 0
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	v := NewVPC().InProject(URI("/projects/p")).
		Named("x")
	_, err := adapter.Update(context.Background(), v)
	if err == nil {
		t.Fatal("expected error when VPC has no ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when ID is missing")
	}
}

func TestVPCsClientAdapter_Update_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	v := &VPC{}
	v.fromResponse(&types.VPCResponse{
		Metadata: types.ResourceMetadataResponse{
			ID: strPtr("vid"),
		},
	})

	_, err := adapter.Update(context.Background(), v)
	if err == nil {
		t.Fatal("expected error when VPC has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestVPCsClientAdapter_Delete_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestVPCsClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestVPCsClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "vpc not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/missing"))
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

// IsDefault and IsPreset zero-value tests cover the 66.7% branches.
func TestVPC_IsDefault_ZeroValue(t *testing.T) {
	v := NewVPC()
	if v.IsDefault() {
		t.Error("IsDefault() on zero value should be false")
	}
}

func TestVPC_IsPreset_ZeroValue(t *testing.T) {
	v := NewVPC()
	if v.IsPreset() {
		t.Error("IsPreset() on zero value should be false")
	}
}

func TestVPCsClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "vpc not found", 404))
	})

	ref := URI("/projects/p/providers/Aruba.Network/vpcs/missing")
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

func TestVPCsClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "vpc not found", 404))
	})

	v := &VPC{}
	v.fromResponse(vpcTestResponse("vpc-1", "my-vpc", "/projects/p/providers/Aruba.Network/vpcs/vpc-1", "p"))
	_, err := adapter.Update(context.Background(), v)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
}

func TestVPCsClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Forbidden", "access denied", 403))
	})

	_, err := adapter.List(context.Background(), URI("/projects/p"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
}

func TestVPCsClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"v1","name":"n1","uri":"/projects/p/providers/Aruba.Network/vpcs/v1","project":{"id":"p"}},"properties":{"default":false},"status":{}},`+
			`{"metadata":{"id":"v2","name":"n2","uri":"/projects/p/providers/Aruba.Network/vpcs/v2","project":{"id":"p"}},"properties":{"default":true},"status":{}}`+
			`]}`)
	})

	list, err := adapter.List(context.Background(), URI("/projects/p"))
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
	if items[0].ID() != "v1" || items[0].Name() != "n1" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[1].ID() != "v2" || items[1].IsDefault() != true {
		t.Errorf("items[1] ID=%q IsDefault=%v", items[1].ID(), items[1].IsDefault())
	}
}

// strPtr is a test helper local to this file.
func strPtr(s string) *string { return &s }

func TestVPC_RawJSON_RoundTrip(t *testing.T) {
	v := &VPC{}
	wire := vpcTestResponse("vpc-1", "my-vpc", "/projects/p/providers/Aruba.Network/vpcs/vpc-1", "p")
	v.fromResponse(wire)
	b := v.RawJSON()
	if len(b) == 0 {
		t.Fatal("RawJSON() returned empty")
	}
	var back types.VPCResponse
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if back.Metadata.ID == nil || *back.Metadata.ID != "vpc-1" {
		t.Errorf("round-trip lost ID: %+v", back.Metadata)
	}
	if back.Metadata.Name == nil || *back.Metadata.Name != "my-vpc" {
		t.Errorf("round-trip lost Name: %+v", back.Metadata)
	}
}

func TestVPC_RawJSON_NilSafe(t *testing.T) {
	v := &VPC{}
	if v.RawJSON() != nil {
		t.Error("RawJSON() on zero-value VPC should return nil")
	}
	if v.RawYAML() != nil {
		t.Error("RawYAML() on zero-value VPC should return nil")
	}
}

func TestVPCsClientAdapter_Create_WithBuilderError(t *testing.T) {
	callCount := 0
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})
	v := NewVPC().InProject(URI("/garbage"))
	_, err := adapter.Create(context.Background(), v)
	if err == nil {
		t.Fatal("expected error for builder error")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when builder has errors")
	}
}

func TestVPCsClientAdapter_Get_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestVPCsClientAdapter_Get_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPCsClientAdapter(testutil.NewClient(t, server.URL))
	result, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v"))
	if err == nil {
		t.Fatal("expected transport error")
	}
	_ = result
}

func TestVPCsClientAdapter_Update_WithBuilderError(t *testing.T) {
	callCount := 0
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	v := NewVPC().InProject(URI("/garbage"))
	_, err := adapter.Update(context.Background(), v)
	if err == nil {
		t.Fatal("expected error for builder error")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when builder has errors")
	}
}

func TestVPCsClientAdapter_Update_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPCsClientAdapter(testutil.NewClient(t, server.URL))
	v := &VPC{}
	v.fromResponse(vpcTestResponse("vid", "n", "/projects/p/providers/Aruba.Network/vpcs/vid", "p"))
	_, err := adapter.Update(context.Background(), v)
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestVPCsClientAdapter_Delete_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPCsClientAdapter(testutil.NewClient(t, server.URL))
	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/v"))
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestVPCsClientAdapter_List_BadProjectRef(t *testing.T) {
	callCount := 0
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/garbage"))
	if err == nil {
		t.Fatal("expected error for bad project Ref")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made for bad project Ref")
	}
}

func TestVPCsClientAdapter_List_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPCsClientAdapter(testutil.NewClient(t, server.URL))
	_, err := adapter.List(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestVPCIDsFromRef_MissingProjectID(t *testing.T) {
	// URI has vpcs segment but no projects segment
	_, _, err := vpcIDsFromRef(URI("/providers/Aruba.Network/vpcs/v"))
	if err == nil {
		t.Error("expected error when project ID is missing from URI")
	}
}

func TestVPC_FromResponse_SetsStatus(t *testing.T) {
	v := &VPC{}
	state := types.State("Active")
	v.fromResponse(&types.VPCResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if v.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", v.State())
	}
}

func TestVPCsClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"v-1","name":"my-vpc","uri":"/projects/p-1/providers/Aruba.Network/vpcs/v-1","project":{"id":"p-1"}},"properties":{"default":false},"status":{"state":"Active"}}`)
	})
	adapter := newVPCsClientAdapter(testutil.NewClient(t, server.URL))
	vpc, err := adapter.Get(context.Background(), URI("/projects/p-1/providers/Aruba.Network/vpcs/v-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&vpc.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned VPC")
	}
}

func TestVPCsClientAdapter_Get_BackfillsProjectID(t *testing.T) {
	// API response does not include project metadata; projectID must be
	// backfilled from the Ref so that a subsequent Update does not fail.
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"v-1","name":"my-vpc","uri":"/projects/p-1/providers/Aruba.Network/vpcs/v-1"},"properties":{"default":false},"status":{}}`)
	})
	adapter := newVPCsClientAdapter(testutil.NewClient(t, server.URL))
	vpc, err := adapter.Get(context.Background(), URI("/projects/p-1/providers/Aruba.Network/vpcs/v-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if vpc.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q after Get without project metadata, want %q", vpc.ProjectID(), "p-1")
	}
}

func TestVPCsClientAdapter_List_BackfillsProjectID(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"values":[{"metadata":{"id":"v-1","name":"vpc1","uri":"/projects/p-1/providers/Aruba.Network/vpcs/v-1"},"properties":{"default":false},"status":{}}],"total":1}`)
	})
	adapter := newVPCsClientAdapter(testutil.NewClient(t, server.URL))
	list, err := adapter.List(context.Background(), URI("/projects/p-1"))
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list.Items()) == 0 {
		t.Fatal("expected at least one item")
	}
	if list.Items()[0].ProjectID() != "p-1" {
		t.Errorf("Items()[0].ProjectID() = %q, want %q", list.Items()[0].ProjectID(), "p-1")
	}
}

func TestVPCRef(t *testing.T) {
	ref := VPCRef("p-1", "vpc-1")
	want := "/projects/p-1/providers/Aruba.Network/vpcs/vpc-1"
	if ref.URI() != want {
		t.Errorf("VPCRef URI = %q, want %q", ref.URI(), want)
	}
	ids := parseURIIDs(ref.URI())
	if ids["projects"] != "p-1" || ids["vpcs"] != "vpc-1" {
		t.Errorf("parseURIIDs = %v", ids)
	}
}

func TestVPCsClientAdapter_WaitUntilGone(t *testing.T) {
	callCount := 0
	adapter := buildVPCTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"metadata":{"id":"vid","name":"my-vpc","uri":"/projects/p/providers/Aruba.Network/vpcs/vid","project":{"id":"p"}},"properties":{"default":false},"status":{"state":"Active"}}`)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "vpc not found", 404))
		}
	})

	vpc, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Network/vpcs/vid"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if err := vpc.WaitUntilGone(context.Background(), fastOpts()...); err != nil {
		t.Fatalf("WaitUntilGone error: %v", err)
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 calls (1 Get + 1 refresh returning 404), got %d", callCount)
	}
}
