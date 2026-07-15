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

var _ Ref = (*ElasticIP)(nil)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestElasticIP_FluentSetters(t *testing.T) {
	parent := &Project{}
	parent.fromResponse(projectTestResponse("proj-1", "my-proj", "/projects/proj-1"))

	e := NewElasticIP().
		InProject(parent).
		Named("my-eip").
		Tagged("net").
		Tagged("public").
		Tagged("net"). // dedupe
		InRegion(RegionITBGBergamo).
		BilledBy(BillingPeriodHour)

	if e.Name() != "my-eip" {
		t.Errorf("Name() = %q", e.Name())
	}
	if tags := e.Tags(); len(tags) != 2 || tags[0] != "net" || tags[1] != "public" {
		t.Errorf("Tags() = %v", tags)
	}
	if e.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", e.Region())
	}
	if e.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", e.BillingPeriod())
	}
	if e.ProjectID() != "proj-1" {
		t.Errorf("ProjectID() = %q", e.ProjectID())
	}
	if e.Err() != nil {
		t.Errorf("Err() = %v", e.Err())
	}

	e.Untagged("net")
	if tags := e.Tags(); len(tags) != 1 || tags[0] != "public" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	e.RetaggedAs("x", "y")
	if tags := e.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

// --------------------------------------------------------------------------
// IntoProject with bad Ref
// --------------------------------------------------------------------------

func TestElasticIP_IntoProject_BadRef(t *testing.T) {
	e := NewElasticIP().InProject(URI("/garbage"))
	if e.Err() == nil {
		t.Error("expected Err() != nil for unresolvable Ref, got nil")
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestElasticIP_ToRequestRoundTrip(t *testing.T) {
	e := NewElasticIP().Named(
		"eip-1").
		Tagged("t1").
		Tagged("t2").
		InRegion(RegionITBGBergamo).
		BilledBy(BillingPeriodHour)

	req := e.RawRequest()

	if req.Metadata.Name != "eip-1" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if len(req.Metadata.Tags) != 2 {
		t.Errorf("Metadata.Tags = %v", req.Metadata.Tags)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Location.Value = %q", req.Metadata.Location.Value)
	}
	if req.Properties.BillingPlanCommon == nil || req.Properties.BillingPlanCommon.BillingPeriod == nil || *req.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodHour {
		t.Errorf("BillingPlanCommon.BillingPeriod (wire) = %v", req.Properties.BillingPlanCommon)
	}

	// No billing period set → defaults to Hour on the wire.
	e2 := NewElasticIP().
		Named("bare")
	req2 := e2.RawRequest()
	if req2.Properties.BillingPlanCommon == nil || req2.Properties.BillingPlanCommon.BillingPeriod == nil || *req2.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodHour {
		t.Errorf("unset BillingPeriod should default to Hour in billingPlan on wire, got %v", req2.Properties.BillingPlanCommon)
	}
}

// --------------------------------------------------------------------------
// BillingPlanCommon wire round-trip
// --------------------------------------------------------------------------

func TestElasticIP_BillingPeriod_WireBillingPlan(t *testing.T) {
	for _, period := range []BillingPeriod{BillingPeriodHour, BillingPeriodMonth, BillingPeriodYear} {
		tc := period
		t.Run(string(tc), func(t *testing.T) {
			req := NewElasticIP().Named("x").InRegion(RegionITBGBergamo).BilledBy(tc).RawRequest()
			if req.Properties.BillingPlanCommon == nil || req.Properties.BillingPlanCommon.BillingPeriod == nil {
				t.Fatal("BillingPlanCommon or BillingPlanCommon.BillingPeriod is nil")
			}
			if got := *req.Properties.BillingPlanCommon.BillingPeriod; got != tc {
				t.Errorf("toRequest wire = %q, want %q", got, tc)
			}
			std := tc
			e := &ElasticIP{}
			e.fromResponse(&types.ElasticIPResponse{
				Properties: types.ElasticIPPropertiesResponse{
					BillingPlanCommon: &types.BillingPlanCommon{BillingPeriod: &std},
				},
			})
			if e.BillingPeriod() != tc {
				t.Errorf("fromResponse getter = %q, want %q", e.BillingPeriod(), tc)
			}
		})
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func elasticIPTestResponse(id, name, uri, projectID string) *types.ElasticIPResponse {
	loc := &types.LocationResponse{Value: RegionITBGBergamo}
	state := types.State("Active")
	addr := "1.2.3.4"
	return &types.ElasticIPResponse{
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
		Properties: types.ElasticIPPropertiesResponse{
			BillingPlanCommon: &types.BillingPlanCommon{BillingPeriod: func() *BillingPeriod { v := BillingPeriodHour; return &v }()},
			Address:           &addr,
			LinkedResources: []types.LinkedResourceCommon{
				{URI: "/projects/p/providers/Aruba.Compute/cloudservers/cs1", StrictCorrelation: true},
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

func TestElasticIP_FromResponseHydration(t *testing.T) {
	e := &ElasticIP{}
	resp := elasticIPTestResponse("eip-1", "my-eip", "/projects/p1/providers/Aruba.Network/elasticIps/eip-1", "p1")
	e.fromResponse(resp)

	if e.ID() != "eip-1" {
		t.Errorf("ID() = %q", e.ID())
	}
	if e.URI() != "/projects/p1/providers/Aruba.Network/elasticIps/eip-1" {
		t.Errorf("URI() = %q", e.URI())
	}
	if e.ElasticIPID() != "eip-1" {
		t.Errorf("ElasticIPID() = %q", e.ElasticIPID())
	}
	if e.Name() != "my-eip" {
		t.Errorf("Name() = %q", e.Name())
	}
	if tags := e.Tags(); len(tags) != 1 || tags[0] != "tag1" {
		t.Errorf("Tags() = %v", tags)
	}
	if e.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", e.Region())
	}
	if e.State() != "Active" {
		t.Errorf("State() = %q", e.State())
	}
	if e.IsDisabled() {
		t.Error("IsDisabled() should be false")
	}
	if linked := e.LinkedResources(); len(linked) != 1 {
		t.Errorf("LinkedResources() len = %d", len(linked))
	}
	if e.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", e.BillingPeriod())
	}
	if e.Address() != "1.2.3.4" {
		t.Errorf("Address() = %q", e.Address())
	}
	if e.ProjectID() != "p1" {
		t.Errorf("ProjectID() = %q", e.ProjectID())
	}
	if e.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestElasticIP_FromResponsePartial(t *testing.T) {
	// nil response is a no-op
	e := &ElasticIP{}
	e.fromResponse(nil)
	if e.ID() != "" || e.URI() != "" || e.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}

	// empty response — accessors must not panic; zero values expected
	e2 := &ElasticIP{}
	e2.fromResponse(&types.ElasticIPResponse{})
	if e2.ID() != "" || e2.URI() != "" || e2.State() != "" || e2.BillingPeriod() != "" || e2.Address() != "" {
		t.Error("empty response should yield zero accessor values")
	}
}

// --------------------------------------------------------------------------
// Ref + ancestor ID satisfaction (runtime)
// --------------------------------------------------------------------------

func TestElasticIP_RefSatisfaction(t *testing.T) {
	e := &ElasticIP{}
	e.fromResponse(elasticIPTestResponse("eip-99", "n", "/projects/p99/providers/Aruba.Network/elasticIps/eip-99", "p99"))

	// withElasticIPID typed path
	eid, ok := extractID(e, func(r Ref) (string, bool) {
		if w, ok := r.(withElasticIPID); ok {
			return w.ElasticIPID(), true
		}
		return "", false
	}, "elasticIps")
	if !ok || eid != "eip-99" {
		t.Errorf("extractID via withElasticIPID = (%q, %v)", eid, ok)
	}

	// withProjectID typed path
	pid, ok := extractID(e, func(r Ref) (string, bool) {
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
// elasticIPIDsFromRef helper
// --------------------------------------------------------------------------

func TestElasticIPIDsFromRef_TypedRef(t *testing.T) {
	e := &ElasticIP{}
	e.fromResponse(elasticIPTestResponse("eid", "n", "/projects/p/providers/Aruba.Network/elasticIps/eid", "p"))
	pid, eid, err := elasticIPIDsFromRef(e)
	if err != nil || pid != "p" || eid != "eid" {
		t.Errorf("elasticIPIDsFromRef typed = (%q, %q, %v)", pid, eid, err)
	}
}

func TestElasticIPIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Network/elasticIps/e1")
	pid, eid, err := elasticIPIDsFromRef(ref)
	if err != nil || pid != "p" || eid != "e1" {
		t.Errorf("elasticIPIDsFromRef URI = (%q, %q, %v)", pid, eid, err)
	}
}

func TestElasticIPIDsFromRef_BadURI(t *testing.T) {
	_, _, err := elasticIPIDsFromRef(URI("/something/else"))
	if err == nil {
		t.Error("expected error for URI without /elasticIps/<id>")
	}
}

// --------------------------------------------------------------------------
// elasticIPsClientAdapter — CRUD integration tests
// --------------------------------------------------------------------------

func buildElasticIPTestAdapter(t *testing.T, handler http.HandlerFunc) *elasticIPsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newElasticIPsClientAdapter(testutil.NewClient(t, server.URL))
}

const elasticIPSuccessBody = `{` +
	`"metadata":{"id":"eid","name":"my-eip","uri":"/projects/p/providers/Aruba.Network/elasticIps/eid","project":{"id":"p"}},` +
	`"properties":{"billingPlan":{"billingPeriod":"Hour"},"address":"1.2.3.4"},` +
	`"status":{"state":"Active"}}`

func TestElasticIPsClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.ElasticIPRequest
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, elasticIPSuccessBody)
	})

	eip := NewElasticIP().
		InProject(URI("/projects/p")).
		Named("my-eip").
		InRegion(RegionITBGBergamo).
		BilledBy(BillingPeriodHour)

	result, err := adapter.Create(context.Background(), eip)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "eid" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-eip" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Metadata.Name != "my-eip" {
		t.Errorf("request Name = %q", gotBody.Metadata.Name)
	}
	if gotBody.Properties.BillingPlanCommon == nil || gotBody.Properties.BillingPlanCommon.BillingPeriod == nil || *gotBody.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodHour {
		t.Errorf("request BillingPlanCommon.BillingPeriod (wire) = %v", gotBody.Properties.BillingPlanCommon)
	}
}

func TestElasticIPsClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	_, err := adapter.Create(context.Background(), NewElasticIP().
		Named("x"))
	if err == nil {
		t.Fatal("expected error when elastic IP has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestElasticIPsClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" field — triggers MetadataValidationError
		fmt.Fprint(w, `{"metadata":{"name":"eip","uri":"/projects/p/providers/Aruba.Network/elasticIps/x"},"properties":{},"status":{}}`)
	})

	eip := NewElasticIP().InProject(URI("/projects/p")).
		Named("eip")
	result, err := adapter.Create(context.Background(), eip)
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

func TestElasticIPsClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "name is required", 422))
	})

	eip := NewElasticIP().InProject(URI("/projects/p"))
	result, err := adapter.Create(context.Background(), eip)
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

func TestElasticIPsClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, elasticIPSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Network/elasticIps/eid")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "eid" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	wantPath := "/projects/p/providers/Aruba.Network/elasticIps/eid"
	if capturedPath != wantPath {
		t.Errorf("path = %q, want %q", capturedPath, wantPath)
	}
}

func TestElasticIPsClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, elasticIPSuccessBody)
	})

	existing := &ElasticIP{}
	existing.fromResponse(elasticIPTestResponse("eid", "n", "/projects/p/providers/Aruba.Network/elasticIps/eid", "p"))

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "eid" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestElasticIPsClientAdapter_Update_Success(t *testing.T) {
	var capturedBody types.ElasticIPRequest
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"eid","name":"renamed","uri":"/projects/p/providers/Aruba.Network/elasticIps/eid","project":{"id":"p"}},"properties":{"billingPlan":{"billingPeriod":"Hour"}},"status":{}}`)
	})

	e := &ElasticIP{}
	e.fromResponse(elasticIPTestResponse("eid", "orig", "/projects/p/providers/Aruba.Network/elasticIps/eid", "p"))
	e.Named("renamed").BilledBy(BillingPeriodHour)

	result, err := adapter.Update(context.Background(), e)
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

func TestElasticIPsClientAdapter_Update_NoID(t *testing.T) {
	callCount := 0
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	e := NewElasticIP().InProject(URI("/projects/p")).
		Named("x")
	_, err := adapter.Update(context.Background(), e)
	if err == nil {
		t.Fatal("expected error when elastic IP has no ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when ID is missing")
	}
}

func TestElasticIPsClientAdapter_Update_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	e := &ElasticIP{}
	e.fromResponse(&types.ElasticIPResponse{
		Metadata: types.ResourceMetadataResponse{
			ID: strPtr("eid"),
		},
	})

	_, err := adapter.Update(context.Background(), e)
	if err == nil {
		t.Fatal("expected error when elastic IP has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestElasticIPsClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Network/elasticIps/eid"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestElasticIPsClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "elastic IP not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Network/elasticIps/missing"))
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

func TestElasticIPsClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"e1","name":"n1","uri":"/projects/p/providers/Aruba.Network/elasticIps/e1","project":{"id":"p"}},"properties":{"billingPlan":{"billingPeriod":"Hour"}},"status":{}},`+
			`{"metadata":{"id":"e2","name":"n2","uri":"/projects/p/providers/Aruba.Network/elasticIps/e2","project":{"id":"p"}},"properties":{"billingPlan":{"billingPeriod":"Hour"}},"status":{}}`+
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
	if items[0].ID() != "e1" || items[0].Name() != "n1" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[0].BillingPeriod() != BillingPeriodHour {
		t.Errorf("items[0].BillingPeriod() = %q", items[0].BillingPeriod())
	}
	if items[1].ID() != "e2" || items[1].BillingPeriod() != BillingPeriodHour {
		t.Errorf("items[1] ID=%q BillingPeriod=%q", items[1].ID(), items[1].BillingPeriod())
	}
	if items[0].ProjectID() != "p" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
}

func TestElasticIPsClientAdapter_Delete_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestElasticIPsClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.Get(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
}

func TestElasticIPsClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "elastic IP not found", 404))
	})
	_, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Network/elasticIps/eid"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

func TestElasticIPsClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "elastic IP not found", 404))
	})
	e := &ElasticIP{}
	e.fromResponse(elasticIPTestResponse("eid", "my-eip", "/projects/p/providers/Aruba.Network/elasticIps/eid", "p"))
	_, err := adapter.Update(context.Background(), e)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
}

func TestElasticIPsClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Forbidden", "access denied", 403))
	})
	_, err := adapter.List(context.Background(), URI("/projects/p"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

func TestElasticIPsClientAdapter_List_BadProjectRef(t *testing.T) {
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/not-a-project"))
	if err == nil {
		t.Fatal("expected error for bad project Ref")
	}
}

func TestElasticIPIDsFromRef_BadURI_MissingProject(t *testing.T) {
	// URI has elasticIps segment but no projects segment
	_, _, err := elasticIPIDsFromRef(URI("/providers/Aruba.Network/elasticIps/eid"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

func TestElasticIPsClientAdapter_Create_WithBuilderError(t *testing.T) {
	callCount := 0
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})
	e := NewElasticIP().InProject(URI("/garbage"))
	_, err := adapter.Create(context.Background(), e)
	if err == nil {
		t.Fatal("expected error for builder error")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when builder has errors")
	}
}

func TestElasticIPsClientAdapter_Get_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newElasticIPsClientAdapter(testutil.NewClient(t, server.URL))
	result, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Network/elasticIps/eid"))
	if err == nil {
		t.Fatal("expected transport error")
	}
	_ = result
}

func TestElasticIPsClientAdapter_Update_WithBuilderError(t *testing.T) {
	callCount := 0
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	e := NewElasticIP().InProject(URI("/garbage"))
	_, err := adapter.Update(context.Background(), e)
	if err == nil {
		t.Fatal("expected error for builder error")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when builder has errors")
	}
}

func TestElasticIPsClientAdapter_Update_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newElasticIPsClientAdapter(testutil.NewClient(t, server.URL))
	e := &ElasticIP{}
	e.fromResponse(elasticIPTestResponse("eid", "my-eip", "/projects/p/providers/Aruba.Network/elasticIps/eid", "p"))
	_, err := adapter.Update(context.Background(), e)
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestElasticIPsClientAdapter_Delete_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newElasticIPsClientAdapter(testutil.NewClient(t, server.URL))
	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Network/elasticIps/eid"))
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestElasticIPsClientAdapter_List_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newElasticIPsClientAdapter(testutil.NewClient(t, server.URL))
	_, err := adapter.List(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestElasticIPsClientAdapter_List_ProjectIDBackfill(t *testing.T) {
	// Items without projectID in metadata or URI: triggers e.projectID = projectID backfill
	adapter := buildElasticIPTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":1,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"eip-x","name":"eip-x"},"properties":{},"status":{}}`+
			`]}`)
	})

	list, err := adapter.List(context.Background(), URI("/projects/proj-x"))
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

func TestElasticIP_FromResponse_URIBackfillProjectID(t *testing.T) {
	// Ensure the URI-based projectID backfill branch is exercised.
	uri := "/projects/p-uri/providers/Aruba.Network/elasticIps/eip-42"
	id := "eip-42"
	name := "backfilled"
	resp := &types.ElasticIPResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:   &id,
			URI:  &uri,
			Name: &name,
			// ProjectMetadataResponse intentionally nil
		},
	}
	e := &ElasticIP{}
	e.fromResponse(resp)
	if e.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() via URI fallback = %q", e.ProjectID())
	}
}

func TestElasticIP_FromResponse_SetsStatus(t *testing.T) {
	e := &ElasticIP{}
	state := types.State("NotUsed")
	e.fromResponse(&types.ElasticIPResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if e.State() != types.StateNotUsed {
		t.Errorf("State() = %q after fromResponse, want NotUsed", e.State())
	}
}

func TestElasticIP_WaitUntilNotUsed_HappyPath(t *testing.T) {
	e := &ElasticIP{}
	calls := 0
	state := types.State("InCreation")
	e.setRefresh(func(_ context.Context) error {
		calls++
		if calls >= 2 {
			state = "NotUsed"
		}
		s := state
		e.setStatus(&types.ResourceStatusResponse{State: &s})
		return nil
	})
	if err := e.WaitUntilNotUsed(context.Background(), fastOpts()...); err != nil {
		t.Fatalf("WaitUntilNotUsed error: %v", err)
	}
	if e.State() != "NotUsed" {
		t.Errorf("State() = %q after wait, want NotUsed", e.State())
	}
}

func TestElasticIP_WaitUntilUsed_HappyPath(t *testing.T) {
	for _, attachedState := range []types.State{types.StateInUse, types.StateUsed, types.StateReserved} {
		t.Run(string(attachedState), func(t *testing.T) {
			e := &ElasticIP{}
			calls := 0
			state := types.State("InCreation")
			e.setRefresh(func(_ context.Context) error {
				calls++
				if calls >= 2 {
					state = attachedState
				}
				s := state
				e.setStatus(&types.ResourceStatusResponse{State: &s})
				return nil
			})
			if err := e.WaitUntilUsed(context.Background(), fastOpts()...); err != nil {
				t.Fatalf("WaitUntilUsed error for %q: %v", attachedState, err)
			}
			if e.State() != attachedState {
				t.Errorf("State() = %q after wait, want %q", e.State(), attachedState)
			}
		})
	}
}

func TestElasticIPsClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, elasticIPSuccessBody)
	})
	adapter := newElasticIPsClientAdapter(testutil.NewClient(t, server.URL))
	eip, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Network/elasticIps/eid"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&eip.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned ElasticIP")
	}
}

func TestElasticIPRef(t *testing.T) {
	ref := ElasticIPRef("p-1", "eip-1")
	want := "/projects/p-1/providers/Aruba.Network/elasticIps/eip-1"
	if ref.URI() != want {
		t.Errorf("ElasticIPRef URI = %q, want %q", ref.URI(), want)
	}
	ids := parseURIIDs(ref.URI())
	if ids["projects"] != "p-1" || ids["elasticIps"] != "eip-1" {
		t.Errorf("parseURIIDs = %v", ids)
	}
}

// --------------------------------------------------------------------------
// AssociatedResourceURI getter
// --------------------------------------------------------------------------

func TestElasticIP_AssociatedResourceURI_NoLinked(t *testing.T) {
	e := &ElasticIP{}
	if got := e.AssociatedResourceURI(); got != "" {
		t.Errorf("AssociatedResourceURI() with no linked = %q, want \"\"", got)
	}
}

func TestElasticIP_AssociatedResourceURI_WithLinked(t *testing.T) {
	e := &ElasticIP{}
	linkedURI := "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"
	e.setLinked([]types.LinkedResourceCommon{{URI: linkedURI}})
	if got := e.AssociatedResourceURI(); got != linkedURI {
		t.Errorf("AssociatedResourceURI() = %q, want %q", got, linkedURI)
	}
}

func TestElasticIP_AssociatedResourceURI_FirstLinked(t *testing.T) {
	e := &ElasticIP{}
	first := "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"
	second := "/projects/p/providers/Aruba.Network/loadBalancers/lb-1"
	e.setLinked([]types.LinkedResourceCommon{{URI: first}, {URI: second}})
	if got := e.AssociatedResourceURI(); got != first {
		t.Errorf("AssociatedResourceURI() = %q, want first = %q", got, first)
	}
}
