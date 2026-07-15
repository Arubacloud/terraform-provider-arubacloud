package aruba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/clients/security"
	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Compile-time interface satisfaction
// --------------------------------------------------------------------------

var (
	_ Ref     = (*KMS)(nil)
	_ Wrapper = (*KMS)(nil)
)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestKMS_FluentSetters(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-1", "my-proj", "/projects/p-1"))

	k := NewKMS().
		InProject(proj).
		Named("my-kms").
		Tagged("security").
		Tagged("encryption").
		Tagged("security"). // dedupe
		InRegion(RegionITBGBergamo).
		BilledBy(BillingPeriodHour)

	if k.Name() != "my-kms" {
		t.Errorf("Name() = %q", k.Name())
	}
	if tags := k.Tags(); len(tags) != 2 || tags[0] != "security" || tags[1] != "encryption" {
		t.Errorf("Tags() = %v", tags)
	}
	if k.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", k.Region())
	}
	if k.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", k.BillingPeriod())
	}
	if k.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", k.ProjectID())
	}
	if k.Err() != nil {
		t.Errorf("Err() = %v", k.Err())
	}
}

// --------------------------------------------------------------------------
// IntoProject
// --------------------------------------------------------------------------

func TestKMS_IntoProject_TypedRef(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-42", "proj", "/projects/p-42"))
	k := NewKMS().InProject(proj)
	if k.ProjectID() != "p-42" {
		t.Errorf("ProjectID() = %q", k.ProjectID())
	}
	if k.Err() != nil {
		t.Errorf("Err() = %v", k.Err())
	}
}

func TestKMS_IntoProject_URIRef(t *testing.T) {
	k := NewKMS().InProject(URI("/projects/p-uri"))
	if k.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", k.ProjectID())
	}
}

func TestKMS_IntoProject_BadRef(t *testing.T) {
	k := NewKMS().InProject(URI("not-a-project-uri"))
	if k.Err() == nil {
		t.Error("expected Err() != nil for non-project URI")
	}
}

// --------------------------------------------------------------------------
// WithBillingPeriod round-trip
// --------------------------------------------------------------------------

func TestKMS_WithBillingPeriod_RoundTrip(t *testing.T) {
	k := NewKMS().BilledBy(BillingPeriodHour)
	req := k.RawRequest()
	if req.Properties.BillingPeriod == nil || *req.Properties.BillingPeriod != BillingPeriodHour {
		t.Errorf("Properties.BillingPeriod = %v", req.Properties.BillingPeriod)
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestKMS_ToRequest_FullyPopulated(t *testing.T) {
	k := NewKMS().
		InProject(URI("/projects/p")).
		Named("kms-name").
		Tagged("tag1").
		InRegion(RegionITBGBergamo).
		BilledBy(BillingPeriodHour)

	req := k.RawRequest()
	if req.Metadata.Name != "kms-name" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Metadata.Location.Value = %q", req.Metadata.Location.Value)
	}
	if req.Properties.BillingPeriod == nil || *req.Properties.BillingPeriod != BillingPeriodHour {
		t.Errorf("Properties.BillingPeriod = %v", req.Properties.BillingPeriod)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func kmsTestResponse(name string) *types.KmsResponse {
	id := "kms-1"
	uri := "/projects/p/providers/Aruba.Security/kms/kms-1"
	state := types.State("Active")
	return &types.KmsResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			URI:              &uri,
			Name:             func() *string { s := name; return &s }(),
			Tags:             []string{"tag1"},
			LocationResponse: &types.LocationResponse{Value: RegionITBGBergamo},
			ProjectMetadataResponse: &types.ProjectMetadataResponse{
				ID: "p",
			},
		},
		Properties: types.KmsPropertiesResponse{
			BillingPeriod: func() *BillingPeriod { v := BillingPeriodHour; return &v }(),
		},
		Status: types.ResourceStatusResponse{State: &state},
	}
}

func TestKMS_FromResponseHydration(t *testing.T) {
	k := &KMS{}
	k.fromResponse(kmsTestResponse("my-kms"))

	if k.Name() != "my-kms" {
		t.Errorf("Name() = %q", k.Name())
	}
	if k.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", k.ProjectID())
	}
	if k.ID() != "kms-1" {
		t.Errorf("ID() = %q", k.ID())
	}
	if k.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", k.Region())
	}
	if k.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", k.BillingPeriod())
	}
	if k.State() != "Active" {
		t.Errorf("State() = %q", k.State())
	}
}

func TestKMS_FromResponse_BackfillsProjectID_FromURI(t *testing.T) {
	id := "kms-x"
	uri := "/projects/proj-abc/providers/Aruba.Security/kms/kms-x"
	resp := &types.KmsResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  &id,
			URI: &uri,
		},
	}
	k := &KMS{}
	k.fromResponse(resp)
	if k.ProjectID() != "proj-abc" {
		t.Errorf("ProjectID() backfilled from URI = %q", k.ProjectID())
	}
}

func TestKMS_FromResponse_Nil(t *testing.T) {
	k := &KMS{}
	k.fromResponse(nil) // must not panic
}

// --------------------------------------------------------------------------
// kmsIDsFromRef
// --------------------------------------------------------------------------

func TestKMSIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/proj-1/providers/Aruba.Security/kms/kms-42")
	pid, kid, err := kmsIDsFromRef(ref)
	if err != nil {
		t.Fatalf("kmsIDsFromRef error: %v", err)
	}
	if pid != "proj-1" {
		t.Errorf("projectID = %q", pid)
	}
	if kid != "kms-42" {
		t.Errorf("kmsID = %q", kid)
	}
}

func TestKMSIDsFromRef_TypedRef(t *testing.T) {
	k := &KMS{}
	k.fromResponse(kmsTestResponse("k"))
	pid, kid, err := kmsIDsFromRef(k)
	if err != nil {
		t.Fatalf("kmsIDsFromRef error: %v", err)
	}
	if pid != "p" {
		t.Errorf("projectID = %q", pid)
	}
	if kid != "kms-1" {
		t.Errorf("kmsID = %q", kid)
	}
}

func TestKMSIDsFromRef_BadURI_NoKMS(t *testing.T) {
	_, _, err := kmsIDsFromRef(URI("/projects/p/providers/Aruba.Security"))
	if err == nil {
		t.Error("expected error when kms segment missing")
	}
}

func TestKMSIDsFromRef_BadURI_NoProject(t *testing.T) {
	_, _, err := kmsIDsFromRef(URI("/providers/Aruba.Security/kms/k"))
	if err == nil {
		t.Error("expected error when project segment missing")
	}
}

// --------------------------------------------------------------------------
// HTTP-mock adapter helper
// --------------------------------------------------------------------------

func buildKMSTestAdapter(t *testing.T, handler http.HandlerFunc) *kmsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newKMSClientAdapter(testutil.NewClient(t, server.URL))
}

const kmsSuccessBody = `{` +
	`"metadata":{"id":"kms-1","name":"my-kms","uri":"/projects/p/providers/Aruba.Security/kms/kms-1","project":{"id":"p"}},` +
	`"properties":{"billingPeriod":"Hour"},` +
	`"status":{"state":"Active"}}`

// --------------------------------------------------------------------------
// Create adapter tests
// --------------------------------------------------------------------------

func TestKMSClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.KmsRequest
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "kms") {
			t.Errorf("path %q should contain 'kms'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, kmsSuccessBody)
	})

	k := NewKMS().
		InProject(URI("/projects/p")).
		Named("my-kms").
		InRegion(RegionITBGBergamo).
		BilledBy(BillingPeriodHour)

	result, err := adapter.Create(context.Background(), k)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "kms-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-kms" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Metadata.Name != "my-kms" {
		t.Errorf("request Metadata.Name = %q", gotBody.Metadata.Name)
	}
	if gotBody.Properties.BillingPeriod == nil || *gotBody.Properties.BillingPeriod != BillingPeriodHour {
		t.Errorf("request BillingPeriod = %v", gotBody.Properties.BillingPeriod)
	}
}

func TestKMSClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})
	_, err := adapter.Create(context.Background(), NewKMS().
		Named("x"))
	if err == nil {
		t.Fatal("expected error when KMS has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestKMSClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" — triggers MetadataValidationError from low-level Validate()
		fmt.Fprint(w, `{"metadata":{"name":"k","uri":"/projects/p/providers/Aruba.Security/kms/x"},"properties":{},"status":{}}`)
	})

	k := NewKMS().InProject(URI("/projects/p")).
		Named("k")
	result, err := adapter.Create(context.Background(), k)
	if err == nil {
		t.Fatal("expected MetadataValidationError, got nil")
	}
	var mvErr *types.MetadataValidationError
	if !errors.As(err, &mvErr) {
		t.Fatalf("expected *types.MetadataValidationError, got %T: %v", err, err)
	}
	if result == nil {
		t.Error("result wrapper should not be nil even on error")
	}
}

func TestKMSClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"bad request"}`)
	})
	_, err := adapter.Create(context.Background(), NewKMS().InProject(URI("/projects/p")).
		Named("k"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Update adapter tests
// --------------------------------------------------------------------------

func TestKMSClientAdapter_Update_Success(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kmsSuccessBody)
	})

	k := &KMS{}
	k.fromResponse(kmsTestResponse("my-kms"))
	k.BilledBy(BillingPeriodHour)

	result, err := adapter.Update(context.Background(), k)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if result.ID() != "kms-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestKMSClientAdapter_Update_NoID(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	k := NewKMS().InProject(URI("/projects/p")).
		Named("x")
	_, err := adapter.Update(context.Background(), k)
	if err == nil {
		t.Fatal("expected error when KMS has no ID")
	}
}

func TestKMSClientAdapter_Update_NoProject(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	k := &KMS{}
	id := "kms-1"
	k.setMeta(&types.ResourceMetadataResponse{ID: &id})
	_, err := adapter.Update(context.Background(), k)
	if err == nil {
		t.Fatal("expected error when KMS has no project")
	}
}

func TestKMSClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"not found"}`)
	})
	k := &KMS{}
	k.fromResponse(kmsTestResponse("k"))
	_, err := adapter.Update(context.Background(), k)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
}

// --------------------------------------------------------------------------
// Get adapter tests
// --------------------------------------------------------------------------

func TestKMSClientAdapter_Get_URIRef(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kmsSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Security/kms/kms-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "kms-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestKMSClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kmsSuccessBody)
	})

	k := &KMS{}
	k.fromResponse(kmsTestResponse("k"))
	result, err := adapter.Get(context.Background(), k)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "kms-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

// --------------------------------------------------------------------------
// Delete adapter tests
// --------------------------------------------------------------------------

func TestKMSClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	k := &KMS{}
	k.fromResponse(kmsTestResponse("k"))
	if err := adapter.Delete(context.Background(), k); err != nil {
		t.Errorf("Delete error: %v", err)
	}
}

func TestKMSClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"not found"}`)
	})
	k := &KMS{}
	k.fromResponse(kmsTestResponse("k"))
	err := adapter.Delete(context.Background(), k)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
}

// --------------------------------------------------------------------------
// List adapter tests
// --------------------------------------------------------------------------

const kmsListBody = `{"total":2,"values":[` +
	`{"metadata":{"id":"kms-1","name":"kms-one","uri":"/projects/p/providers/Aruba.Security/kms/kms-1","project":{"id":"p"}},"properties":{"billingPeriod":"Hour"},"status":{}},` +
	`{"metadata":{"id":"kms-2","name":"kms-two","uri":"/projects/p/providers/Aruba.Security/kms/kms-2","project":{"id":"p"}},"properties":{"billingPeriod":"Hour"},"status":{}}` +
	`]}`

func TestKMSClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kmsListBody)
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
	if items[0].Name() != "kms-one" {
		t.Errorf("items[0].Name() = %q", items[0].Name())
	}
	if items[1].Name() != "kms-two" {
		t.Errorf("items[1].Name() = %q", items[1].Name())
	}
}

// --------------------------------------------------------------------------
// Setter delegation (RemoveTag, ReplaceTags, InRegion)
// --------------------------------------------------------------------------

func TestKMS_SetterDelegation(t *testing.T) {
	k := NewKMS().
		Tagged("a").
		Tagged("b").
		Tagged("c").
		Untagged("b").
		RetaggedAs("x", "y").
		InRegion(RegionITBGBergamo)

	tags := k.Tags()
	if len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("Tags() after ReplaceTags = %v", tags)
	}
	if k.Region() != RegionITBGBergamo {
		t.Errorf("Region() after InRegion = %q", k.Region())
	}
}

// --------------------------------------------------------------------------
// URI and Raw accessors after hydration
// --------------------------------------------------------------------------

func TestKMS_URI_AfterHydration(t *testing.T) {
	k := &KMS{}
	k.fromResponse(kmsTestResponse("u"))
	if k.URI() == "" {
		t.Error("URI() should not be empty after hydration")
	}
	if k.Raw() == nil {
		t.Error("Raw() should not be nil after hydration")
	}
}

func TestKMS_RawRequest_NoHydration(t *testing.T) {
	_ = NewKMS().RawRequest() // must not panic
}

// --------------------------------------------------------------------------
// Accessors at zero value (no hydration)
// --------------------------------------------------------------------------

func TestKMS_Accessors_ZeroValue(t *testing.T) {
	k := NewKMS()
	if k.URI() != "" {
		t.Errorf("URI() zero = %q", k.URI())
	}
	if k.Raw() != nil {
		t.Errorf("Raw() zero = %v", k.Raw())
	}
	if k.BillingPeriod() != "" {
		t.Errorf("BillingPeriod() zero = %q", k.BillingPeriod())
	}
}

// --------------------------------------------------------------------------
// BillingPeriod — request-side fallback (response is nil)
// --------------------------------------------------------------------------

func TestKMS_BillingPeriod_RequestFallback(t *testing.T) {
	k := NewKMS().BilledBy(BillingPeriodHour)
	// response is nil, so it falls through to the request-side value
	if k.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() request fallback = %q", k.BillingPeriod())
	}
}

// --------------------------------------------------------------------------
// Get adapter — BadRef and NonTwoXX
// --------------------------------------------------------------------------

func TestKMSClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// URI has no project or kms segments
	_, err := adapter.Get(context.Background(), URI("/providers/Aruba.Security"))
	if err == nil {
		t.Fatal("expected error for unparseable ref")
	}
}

func TestKMSClientAdapter_Get_NetworkError(t *testing.T) {
	adapter := &kmsClientAdapter{low: security.NewKMSClientImpl(testutil.NewBrokenClient(t, "http://localhost:9"))}
	_, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Security/kms/kms-1"))
	if err == nil {
		t.Fatal("expected network error from broken client")
	}
}

func TestKMSClientAdapter_Get_ProjectIDFallback(t *testing.T) {
	// Server returns a response that lacks project metadata — ensures the
	// "if out.projectID == """ guard fires and restores the extracted projectID.
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// No "project" field and no project in URI — projectID won't be set by fromResponse.
		id := "kms-99"
		fmt.Fprintf(w, `{"metadata":{"id":%q},"properties":{},"status":{}}`, id)
	})
	ref := URI("/projects/p/providers/Aruba.Security/kms/kms-99")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	// projectID must be restored from the extracted ref IDs.
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() after fallback = %q", result.ProjectID())
	}
}

func TestKMSClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"not found"}`)
	})
	ref := URI("/projects/p/providers/Aruba.Security/kms/kms-1")
	_, err := adapter.Get(context.Background(), ref)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Delete adapter — BadRef
// --------------------------------------------------------------------------

func TestKMSClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := adapter.Delete(context.Background(), URI("/providers/Aruba.Security"))
	if err == nil {
		t.Fatal("expected error for unparseable ref")
	}
}

func TestKMSClientAdapter_Delete_NetworkError(t *testing.T) {
	adapter := &kmsClientAdapter{low: security.NewKMSClientImpl(testutil.NewBrokenClient(t, "http://localhost:9"))}
	k := &KMS{}
	k.fromResponse(kmsTestResponse("k"))
	err := adapter.Delete(context.Background(), k)
	if err == nil {
		t.Fatal("expected network error from broken client")
	}
}

// --------------------------------------------------------------------------
// List adapter — NonTwoXX and bad parent ref
// --------------------------------------------------------------------------

func TestKMSClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"forbidden"}`)
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

func TestKMSClientAdapter_List_BadParentRef(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/no-project-here"))
	if err == nil {
		t.Fatal("expected error for parent ref with no project")
	}
}

// --------------------------------------------------------------------------
// newKMSClientAdapter nil-rest branch
// --------------------------------------------------------------------------

func TestNewKMSClientAdapter_NilRest(t *testing.T) {
	// Exercises the rest == nil guard; the returned adapter has no low-level client
	// but must not panic on construction.
	a := newKMSClientAdapter(nil)
	if a == nil {
		t.Fatal("expected non-nil adapter")
	}
}

// --------------------------------------------------------------------------
// Update — Err() pre-condition
// --------------------------------------------------------------------------

func TestKMSClientAdapter_Update_ErrSet(t *testing.T) {
	adapter := buildKMSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	k := NewKMS().InProject(URI("not-a-project-uri")) // sets Err()
	_, err := adapter.Update(context.Background(), k)
	if err == nil {
		t.Fatal("expected error when KMS has Err() set")
	}
}

// --------------------------------------------------------------------------
// Reflective guards
// --------------------------------------------------------------------------

func TestKMSClient_HasUpdateMethod(t *testing.T) {
	iface := reflect.TypeOf((*KMSClient)(nil)).Elem()
	if _, ok := iface.MethodByName("Update"); !ok {
		t.Error("KMSClient interface is missing the Update method")
	}
}

func TestSecurityClient_HasKeysMethod(t *testing.T) {
	iface := reflect.TypeOf((*SecurityClient)(nil)).Elem()
	if _, ok := iface.MethodByName("Keys"); !ok {
		t.Error("SecurityClient interface is missing the Keys method")
	}
}

func TestSecurityClient_HasKmipsMethod(t *testing.T) {
	iface := reflect.TypeOf((*SecurityClient)(nil)).Elem()
	if _, ok := iface.MethodByName("Kmips"); !ok {
		t.Error("SecurityClient interface is missing the Kmips method")
	}
}

func TestKMS_FromResponse_SetsStatus(t *testing.T) {
	k := &KMS{}
	state := types.State("Active")
	k.fromResponse(&types.KmsResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if k.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", k.State())
	}
}

func TestKMSClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kmsSuccessBody)
	})
	adapter := newKMSClientAdapter(testutil.NewClient(t, server.URL))
	k, err := adapter.Get(context.Background(), URI("/projects/proj-1/providers/Aruba.Security/kms/kms-42"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&k.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned KMS")
	}
}
