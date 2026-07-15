package aruba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Test helpers
// --------------------------------------------------------------------------

// kmipMakeKMSParent creates a KMS wrapper with specific projectID and kmsID
// for use as an IntoKMS parent in Kmip tests.
func kmipMakeKMSParent(projectID, kmsID string) *KMS {
	name := "test-kms"
	id := kmsID
	u := fmt.Sprintf("/projects/%s/providers/Aruba.Security/kms/%s", projectID, kmsID)
	resp := &types.KmsResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:                      &id,
			Name:                    &name,
			URI:                     &u,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{ID: projectID},
		},
	}
	k := &KMS{}
	k.projectScopedMixin = bindProjectScoped(&k.errMixin)
	k.fromResponse(resp)
	return k
}

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestKmip_FluentSetters(t *testing.T) {
	parent := kmipMakeKMSParent("p-1", "kms-1")
	km := NewKmip().
		InKMS(parent).
		Named("my-kmip")

	if km.Name() != "my-kmip" {
		t.Errorf("Name() = %q", km.Name())
	}
	if km.KMSID() != "kms-1" {
		t.Errorf("KMSID() = %q", km.KMSID())
	}
	if km.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", km.ProjectID())
	}
	if km.Err() != nil {
		t.Errorf("Err() = %v", km.Err())
	}
}

// --------------------------------------------------------------------------
// IntoKMS
// --------------------------------------------------------------------------

func TestKmip_IntoKMS_TypedRef(t *testing.T) {
	parent := kmipMakeKMSParent("p-42", "kms-42")
	km := NewKmip().InKMS(parent)
	if km.KMSID() != "kms-42" {
		t.Errorf("KMSID() = %q", km.KMSID())
	}
	if km.ProjectID() != "p-42" {
		t.Errorf("ProjectID() = %q", km.ProjectID())
	}
	if km.Err() != nil {
		t.Errorf("Err() = %v", km.Err())
	}
}

func TestKmip_IntoKMS_URIRef(t *testing.T) {
	km := NewKmip().InKMS(URI("/projects/p-uri/providers/Aruba.Security/kms/kms-uri"))
	if km.KMSID() != "kms-uri" {
		t.Errorf("KMSID() = %q", km.KMSID())
	}
	if km.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", km.ProjectID())
	}
	if km.Err() != nil {
		t.Errorf("Err() = %v", km.Err())
	}
}

func TestKmip_IntoKMS_BadRef_NoKMS(t *testing.T) {
	km := NewKmip().InKMS(URI("/projects/p-1"))
	if km.Err() == nil {
		t.Error("expected Err() != nil for URI missing kms segment")
	}
}

func TestKmip_IntoKMS_BadRef_NoProject(t *testing.T) {
	km := NewKmip().InKMS(URI("/providers/Aruba.Security/kms/kms-1"))
	if km.Err() == nil {
		t.Error("expected Err() != nil for URI missing project segment")
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestKmip_ToRequest_FullyPopulated(t *testing.T) {
	km := NewKmip().
		Named("my-kmip-service")
	req := km.RawRequest()
	if req.Name != "my-kmip-service" {
		t.Errorf("Name = %q", req.Name)
	}
}

func TestKmip_ToRequest_ZeroState(t *testing.T) {
	req := NewKmip().RawRequest()
	if req.Name != "" {
		t.Errorf("Name = %q, want empty", req.Name)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

const kmipSuccessBody = `{` +
	`"id":"kmip-1",` +
	`"name":"my-kmip",` +
	`"type":"Kmip",` +
	`"status":"Active",` +
	`"creationDate":"2024-01-01T00:00:00Z"` +
	`}`

const kmipCertBody = `{"key":"-----BEGIN PRIVATE KEY-----","cert":"-----BEGIN CERTIFICATE-----"}`

func TestKmip_FromResponseHydration(t *testing.T) {
	id := "kmip-1"
	name := "my-kmip"
	typ := "Kmip"
	status := ServiceStatusActive
	creation := "2024-01-01T00:00:00Z"
	deletion := "2024-12-31T00:00:00Z"

	resp := &types.KmipResponse{
		ID:           &id,
		Name:         &name,
		Type:         &typ,
		Status:       &status,
		CreationDate: &creation,
		DeletionDate: &deletion,
	}

	km := NewKmip()
	km.projectID = "p-1"
	km.kmsID = "kms-1"
	km.fromResponse(resp)

	if km.ID() != "kmip-1" {
		t.Errorf("ID() = %q", km.ID())
	}
	if km.KmipID() != "kmip-1" {
		t.Errorf("KmipID() = %q", km.KmipID())
	}
	if km.Name() != "my-kmip" {
		t.Errorf("Name() = %q", km.Name())
	}
	if km.Type() != "Kmip" {
		t.Errorf("Type() = %q", km.Type())
	}
	if km.KmipStatus() != "Active" {
		t.Errorf("KmipStatus() = %q", km.KmipStatus())
	}
	if km.CreationDate() != "2024-01-01T00:00:00Z" {
		t.Errorf("CreationDate() = %q", km.CreationDate())
	}
	if km.DeletionDate() != "2024-12-31T00:00:00Z" {
		t.Errorf("DeletionDate() = %q", km.DeletionDate())
	}
	if km.Raw() == nil {
		t.Error("Raw() nil after fromResponse")
	}
	if km.URI() == "" {
		t.Error("URI() empty after hydration with projectID+kmsID set")
	}
}

func TestKmip_FromResponse_Nil(t *testing.T) {
	km := NewKmip()
	km.fromResponse(nil)
	if km.response != nil {
		t.Error("response should remain nil after fromResponse(nil)")
	}
}

// --------------------------------------------------------------------------
// URI construction
// --------------------------------------------------------------------------

func TestKmip_URI_AllAncestorsPresent(t *testing.T) {
	km := NewKmip()
	km.projectID = "p-1"
	km.kmsID = "kms-1"
	id := "kmip-1"
	km.fromResponse(&types.KmipResponse{ID: &id})
	want := "/projects/p-1/providers/Aruba.Security/kms/kms-1/kmips/kmip-1"
	if km.URI() != want {
		t.Errorf("URI() = %q, want %q", km.URI(), want)
	}
}

func TestKmip_URI_MissingKmipID(t *testing.T) {
	km := NewKmip()
	km.projectID = "p-1"
	km.kmsID = "kms-1"
	if km.URI() != "" {
		t.Errorf("URI() = %q, want empty when kmipID missing", km.URI())
	}
}

func TestKmip_URI_MissingKMSID(t *testing.T) {
	km := NewKmip()
	km.projectID = "p-1"
	id := "kmip-1"
	km.fromResponse(&types.KmipResponse{ID: &id})
	if km.URI() != "" {
		t.Errorf("URI() = %q, want empty when kmsID missing", km.URI())
	}
}

func TestKmip_URI_MissingProjectID(t *testing.T) {
	km := NewKmip()
	km.kmsID = "kms-1"
	id := "kmip-1"
	km.fromResponse(&types.KmipResponse{ID: &id})
	if km.URI() != "" {
		t.Errorf("URI() = %q, want empty when projectID missing", km.URI())
	}
}

// --------------------------------------------------------------------------
// Zero-value accessors (Shape F)
// --------------------------------------------------------------------------

func TestKmip_Accessors_ZeroValue(t *testing.T) {
	km := NewKmip()
	if km.ID() != "" {
		t.Errorf("ID() = %q, want empty", km.ID())
	}
	if km.KmipID() != "" {
		t.Errorf("KmipID() = %q, want empty", km.KmipID())
	}
	if km.Name() != "" {
		t.Errorf("Name() = %q, want empty", km.Name())
	}
	if km.Type() != "" {
		t.Errorf("Type() = %q, want empty", km.Type())
	}
	if km.KmipStatus() != "" {
		t.Errorf("KmipStatus() = %q, want empty", km.KmipStatus())
	}
	if km.CreationDate() != "" {
		t.Errorf("CreationDate() = %q, want empty", km.CreationDate())
	}
	if km.DeletionDate() != "" {
		t.Errorf("DeletionDate() = %q, want empty", km.DeletionDate())
	}
	if km.URI() != "" {
		t.Errorf("URI() = %q, want empty", km.URI())
	}
	if km.Raw() != nil {
		t.Error("Raw() non-nil on zero Kmip")
	}
}

func TestKmip_Name_LocalFallback(t *testing.T) {
	km := NewKmip().
		Named("fallback-name")
	if km.Name() != "fallback-name" {
		t.Errorf("Name() = %q, want fallback-name", km.Name())
	}
}

// --------------------------------------------------------------------------
// kmipIDsFromRef
// --------------------------------------------------------------------------

func TestKmipIDsFromRef_TypedRef(t *testing.T) {
	km := NewKmip()
	km.projectID = "p-1"
	km.kmsID = "kms-1"
	id := "kmip-1"
	km.fromResponse(&types.KmipResponse{ID: &id})

	pid, kid, kmipID, err := kmipIDsFromRef(km)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid != "p-1" || kid != "kms-1" || kmipID != "kmip-1" {
		t.Errorf("got (%q,%q,%q)", pid, kid, kmipID)
	}
}

func TestKmipIDsFromRef_URIRef(t *testing.T) {
	pid, kid, kmipID, err := kmipIDsFromRef(URI("/projects/p-2/providers/Aruba.Security/kms/kms-2/kmips/kmip-2"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid != "p-2" || kid != "kms-2" || kmipID != "kmip-2" {
		t.Errorf("got (%q,%q,%q)", pid, kid, kmipID)
	}
}

func TestKmipIDsFromRef_BadURI_NoKmip(t *testing.T) {
	_, _, _, err := kmipIDsFromRef(URI("/projects/p/providers/Aruba.Security/kms/kms-1"))
	if err == nil {
		t.Error("expected error for URI missing kmips segment")
	}
}

func TestKmipIDsFromRef_BadURI_NoKMS(t *testing.T) {
	_, _, _, err := kmipIDsFromRef(URI("/projects/p/providers/Aruba.Security/kmips/kmip-1"))
	if err == nil {
		t.Error("expected error for URI missing kms segment")
	}
}

func TestKmipIDsFromRef_BadURI_NoProject(t *testing.T) {
	_, _, _, err := kmipIDsFromRef(URI("/providers/Aruba.Security/kms/kms-1/kmips/kmip-1"))
	if err == nil {
		t.Error("expected error for URI missing projects segment")
	}
}

// --------------------------------------------------------------------------
// HTTP-mock adapter helper
// --------------------------------------------------------------------------

func buildKmipTestAdapter(t *testing.T, handler http.HandlerFunc) *kmipsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newKmipsClientAdapter(testutil.NewClient(t, server.URL))
}

// --------------------------------------------------------------------------
// Create adapter tests
// --------------------------------------------------------------------------

func TestKmipsClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.KmipRequest
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, kmipSuccessBody)
	})

	parent := kmipMakeKMSParent("p-1", "kms-1")
	km := NewKmip().InKMS(parent).
		Named("my-kmip")

	result, err := adapter.Create(context.Background(), km)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.KmipID() != "kmip-1" {
		t.Errorf("KmipID() = %q", result.KmipID())
	}
	if result.Name() != "my-kmip" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Name != "my-kmip" {
		t.Errorf("request Name = %q", gotBody.Name)
	}
}

func TestKmipsClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})
	km := NewKmip().
		Named("x")
	_, err := adapter.Create(context.Background(), km)
	if err == nil {
		t.Fatal("expected error when Kmip has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestKmipsClientAdapter_Create_NoKMS(t *testing.T) {
	callCount := 0
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})
	km := NewKmip()
	km.projectID = "p-1"
	km.Named("x")
	_, err := adapter.Create(context.Background(), km)
	if err == nil {
		t.Fatal("expected error when Kmip has no KMS ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without KMS ID")
	}
}

func TestKmipsClientAdapter_Create_ErrMixin(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	km := NewKmip().InKMS(URI("not-a-valid-kms-uri"))
	_, err := adapter.Create(context.Background(), km)
	if err == nil {
		t.Fatal("expected error from errMixin when IntoKMS failed")
	}
}

func TestKmipsClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"bad request"}`)
	})
	parent := kmipMakeKMSParent("p-1", "kms-1")
	_, err := adapter.Create(context.Background(), NewKmip().InKMS(parent).
		Named("km"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Get adapter tests
// --------------------------------------------------------------------------

func TestKmipsClientAdapter_Get_URIRef(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kmipSuccessBody)
	})

	result, err := adapter.Get(context.Background(), URI(
		"/projects/p-1/providers/Aruba.Security/kms/kms-1/kmips/kmip-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.KmipID() != "kmip-1" {
		t.Errorf("KmipID() = %q", result.KmipID())
	}
}

func TestKmipsClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kmipSuccessBody)
	})

	km := NewKmip()
	km.projectID = "p-1"
	km.kmsID = "kms-1"
	id := "kmip-1"
	km.fromResponse(&types.KmipResponse{ID: &id})

	result, err := adapter.Get(context.Background(), km)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.KmipID() != "kmip-1" {
		t.Errorf("KmipID() = %q", result.KmipID())
	}
}

func TestKmipsClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.Get(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected error for bad ref missing kmip/kms IDs")
	}
}

func TestKmipsClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	_, err := adapter.Get(context.Background(), URI(
		"/projects/p-1/providers/Aruba.Security/kms/kms-1/kmips/kmip-1"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Delete adapter tests
// --------------------------------------------------------------------------

func TestKmipsClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := adapter.Delete(context.Background(), URI(
		"/projects/p-1/providers/Aruba.Security/kms/kms-1/kmips/kmip-1"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestKmipsClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := adapter.Delete(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected error for bad ref")
	}
}

func TestKmipsClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	err := adapter.Delete(context.Background(), URI(
		"/projects/p-1/providers/Aruba.Security/kms/kms-1/kmips/kmip-1"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// List adapter tests
// --------------------------------------------------------------------------

const kmipListBody = `{` +
	`"total":2,` +
	`"values":[` +
	`{"id":"kmip-1","name":"km1","status":"Active"},` +
	`{"id":"kmip-2","name":"km2","status":"Pending"}` +
	`]}`

func TestKmipsClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kmipListBody)
	})

	parent := kmipMakeKMSParent("p-1", "kms-1")
	list, err := adapter.List(context.Background(), parent)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if list.Total() != 2 {
		t.Errorf("Total() = %d", list.Total())
	}
	items := list.Items()
	if len(items) != 2 {
		t.Fatalf("len(Items()) = %d", len(items))
	}
	if items[0].KmipID() != "kmip-1" {
		t.Errorf("items[0].KmipID() = %q", items[0].KmipID())
	}
	if items[1].KmipStatus() != "Pending" {
		t.Errorf("items[1].KmipStatus() = %q", items[1].KmipStatus())
	}
}

func TestKmipsClientAdapter_List_BadParentRef(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected error for parent ref missing KMS ID")
	}
}

func TestKmipsClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	parent := kmipMakeKMSParent("p-1", "kms-1")
	_, err := adapter.List(context.Background(), parent)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Download adapter tests
// --------------------------------------------------------------------------

func TestKmipsClientAdapter_Download_Success(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kmipCertBody)
	})

	cert, err := adapter.Download(context.Background(), URI(
		"/projects/p-1/providers/Aruba.Security/kms/kms-1/kmips/kmip-1"))
	if err != nil {
		t.Fatalf("Download error: %v", err)
	}
	if cert == nil {
		t.Fatal("Download returned nil cert")
	}
	if cert.Key() == "" {
		t.Error("cert.Key() is empty")
	}
	if cert.Cert() == "" {
		t.Error("cert.Cert() is empty")
	}
}

func TestKmipsClientAdapter_Download_LowLevelError(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		// Hijack the connection to force a network-level error.
		hj, ok := w.(http.Hijacker)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	_, err := adapter.Download(context.Background(), URI(
		"/projects/p-1/providers/Aruba.Security/kms/kms-1/kmips/kmip-1"))
	if err == nil {
		t.Fatal("expected error from low-level client on connection drop")
	}
}

func TestKmipsClientAdapter_Download_BadRef(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.Download(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected error for bad ref missing kmip/kms IDs")
	}
}

func TestKmipsClientAdapter_Download_NonTwoXX(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	_, err := adapter.Download(context.Background(), URI(
		"/projects/p-1/providers/Aruba.Security/kms/kms-1/kmips/kmip-1"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// KmipCertificate wrapper tests
// --------------------------------------------------------------------------

func TestKmipCertificate_Accessors_NilSafe(t *testing.T) {
	var c *KmipCertificate
	if c.Key() != "" {
		t.Errorf("Key() on nil receiver = %q, want empty", c.Key())
	}
	if c.Cert() != "" {
		t.Errorf("Cert() on nil receiver = %q, want empty", c.Cert())
	}
	if c.Raw() != nil {
		t.Error("Raw() on nil receiver should be nil")
	}

	withNilResp := &KmipCertificate{response: nil}
	if withNilResp.Key() != "" {
		t.Errorf("Key() with nil response = %q, want empty", withNilResp.Key())
	}
	if withNilResp.Cert() != "" {
		t.Errorf("Cert() with nil response = %q, want empty", withNilResp.Cert())
	}
	if withNilResp.Raw() != nil {
		t.Error("Raw() with nil response should be nil")
	}
}

func TestKmipCertificate_Accessors_Populated(t *testing.T) {
	c := &KmipCertificate{response: &types.KmipCertificateResponse{
		Key:  "-----BEGIN PRIVATE KEY-----",
		Cert: "-----BEGIN CERTIFICATE-----",
	}}
	if c.Key() != "-----BEGIN PRIVATE KEY-----" {
		t.Errorf("Key() = %q", c.Key())
	}
	if c.Cert() != "-----BEGIN CERTIFICATE-----" {
		t.Errorf("Cert() = %q", c.Cert())
	}
	if c.Raw() == nil {
		t.Error("Raw() should not be nil after population")
	}
}

func TestKmipsClientAdapter_Download_ReturnsWrapper(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kmipCertBody)
	})

	cert, err := adapter.Download(context.Background(), URI(
		"/projects/p-1/providers/Aruba.Security/kms/kms-1/kmips/kmip-1"))
	if err != nil {
		t.Fatalf("Download error: %v", err)
	}
	if cert == nil {
		t.Fatal("Download returned nil cert")
	}
	if _, ok := any(cert).(*KmipCertificate); !ok {
		t.Errorf("Download returned %T, want *KmipCertificate", cert)
	}
	if cert.Raw() == nil {
		t.Error("Raw() should not be nil on a successful download")
	}
}

// --------------------------------------------------------------------------
// Nil-rest guard
// --------------------------------------------------------------------------

func TestNewKmipsClientAdapter_Nil(t *testing.T) {
	a := newKmipsClientAdapter(nil)
	if a == nil {
		t.Fatal("newKmipsClientAdapter(nil) returned nil")
	}
}

// --------------------------------------------------------------------------
// WaitUntilCertificateAvailable tests
// --------------------------------------------------------------------------

func TestKmip_WaitUntilCertificateAvailable_HappyPath(t *testing.T) {
	km := &Kmip{}
	calls := 0
	status := ServiceStatusInCreation
	km.setRefresh(func(_ context.Context) error {
		calls++
		if calls >= 2 {
			status = ServiceStatusCertificateAvailable
		}
		st := status
		km.response = &types.KmipResponse{Status: &st}
		return nil
	})
	if err := km.WaitUntilCertificateAvailable(context.Background(), fastOpts()...); err != nil {
		t.Fatalf("WaitUntilCertificateAvailable error: %v", err)
	}
	if km.KmipStatus() != string(ServiceStatusCertificateAvailable) {
		t.Errorf("KmipStatus() = %q after wait", km.KmipStatus())
	}
}

func TestKmip_WaitUntilCertificateAvailable_TerminalFailure(t *testing.T) {
	km := &Kmip{}
	failed := ServiceStatusFailed
	km.setRefresh(func(_ context.Context) error {
		km.response = &types.KmipResponse{Status: &failed}
		return nil
	})
	if err := km.WaitUntilCertificateAvailable(context.Background(), fastOpts()...); err == nil {
		t.Fatal("expected error for terminal failure state, got nil")
	}
}

func TestKmip_WaitUntilCertificateAvailable_NotHydrated(t *testing.T) {
	km := &Kmip{}
	err := km.WaitUntilCertificateAvailable(context.Background())
	if err == nil {
		t.Fatal("expected error when refresh callback is not set")
	}
}

func TestKmipsClientAdapter_Create_InjectsRefresh(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, kmipSuccessBody)
	})
	parent := kmipMakeKMSParent("p-1", "kms-1")
	km, err := adapter.Create(context.Background(), NewKmip().InKMS(parent).
		Named("my-kmip"))
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if km.refresh == nil {
		t.Error("Create should inject a refresh callback into the returned Kmip")
	}
}

func TestKmipsClientAdapter_Get_InjectsRefresh(t *testing.T) {
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kmipSuccessBody)
	})
	km, err := adapter.Get(context.Background(), URI(
		"/projects/p-1/providers/Aruba.Security/kms/kms-1/kmips/kmip-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if km.refresh == nil {
		t.Error("Get should inject a refresh callback into the returned Kmip")
	}
}

// --------------------------------------------------------------------------
// WaitUntilReady (alias for WaitUntilCertificateAvailable)
// --------------------------------------------------------------------------

func TestKmip_WaitUntilReady_HappyPath(t *testing.T) {
	km := &Kmip{}
	calls := 0
	status := ServiceStatusInCreation
	km.setRefresh(func(_ context.Context) error {
		calls++
		if calls >= 2 {
			status = ServiceStatusCertificateAvailable
		}
		s := status
		km.response = &types.KmipResponse{Status: &s}
		return nil
	})
	if err := km.WaitUntilReady(context.Background(), fastOpts()...); err != nil {
		t.Fatalf("WaitUntilReady error: %v", err)
	}
	if km.KmipStatus() != string(ServiceStatusCertificateAvailable) {
		t.Errorf("KmipStatus() = %q after wait, want %q", km.KmipStatus(), ServiceStatusCertificateAvailable)
	}
}

func TestKmip_WaitUntilReady_AcceptsActive(t *testing.T) {
	km := &Kmip{}
	calls := 0
	status := ServiceStatusInCreation
	km.setRefresh(func(_ context.Context) error {
		calls++
		if calls >= 2 {
			status = ServiceStatusActive
		}
		s := status
		km.response = &types.KmipResponse{Status: &s}
		return nil
	})
	if err := km.WaitUntilReady(context.Background(), fastOpts()...); err != nil {
		t.Fatalf("WaitUntilReady should accept Active as ready, got error: %v", err)
	}
	if km.KmipStatus() != string(ServiceStatusActive) {
		t.Errorf("KmipStatus() = %q after wait, want %q", km.KmipStatus(), ServiceStatusActive)
	}
}

// --------------------------------------------------------------------------
// WaitUntilGone — adapter-level test (Kmip path)
// --------------------------------------------------------------------------

func TestKmipsClientAdapter_WaitUntilGone(t *testing.T) {
	callCount := 0
	adapter := buildKmipTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, kmipSuccessBody)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "kmip not found", 404))
		}
	})

	km, err := adapter.Get(context.Background(), URI("/projects/p-1/providers/Aruba.Security/kms/kms-1/kmips/kmip-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if err := km.WaitUntilGone(context.Background(), fastOpts()...); err != nil {
		t.Fatalf("WaitUntilGone error: %v", err)
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 calls (1 Get + 1 refresh returning 404), got %d", callCount)
	}
}

// --------------------------------------------------------------------------
// Reflective guard: KmipsClient must NOT expose Update (Family B, no-Update)
// --------------------------------------------------------------------------

func TestKmipsClient_NoUpdateMethod(t *testing.T) {
	iface := reflect.TypeOf((*KmipsClient)(nil)).Elem()
	for i := range iface.NumMethod() {
		if iface.Method(i).Name == "Update" {
			t.Error("KmipsClient must not expose an Update method — Kmip is a no-Update Family B resource")
		}
	}
}
