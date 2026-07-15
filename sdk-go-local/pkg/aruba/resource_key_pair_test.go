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

var _ Ref = (*KeyPair)(nil)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestKeyPair_FluentSetters(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-1", "my-project", "/projects/p-1"))

	kp := NewKeyPair().
		InProject(proj).
		Named("allow-ssh").
		Tagged("ssh-access").
		Tagged("ingress").
		Tagged("ssh-access"). // dedupe
		InRegion(RegionITBGBergamo).
		WithPublicKey("ssh-rsa AAAA...")

	if kp.Name() != "allow-ssh" {
		t.Errorf("Name() = %q", kp.Name())
	}
	if tags := kp.Tags(); len(tags) != 2 || tags[0] != "ssh-access" || tags[1] != "ingress" {
		t.Errorf("Tags() = %v", tags)
	}
	if kp.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", kp.Region())
	}
	if kp.PublicKey() != "ssh-rsa AAAA..." {
		t.Errorf("PublicKey() = %q", kp.PublicKey())
	}
	if kp.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", kp.ProjectID())
	}
	if kp.Err() != nil {
		t.Errorf("Err() = %v", kp.Err())
	}

	kp.Untagged("ssh-access")
	if tags := kp.Tags(); len(tags) != 1 || tags[0] != "ingress" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	kp.RetaggedAs("x", "y")
	if tags := kp.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

// --------------------------------------------------------------------------
// IntoProject — typed / URI / bad Ref
// --------------------------------------------------------------------------

func TestKeyPair_IntoProject_TypedRef(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-42", "n", "/projects/p-42"))

	kp := NewKeyPair().InProject(proj)
	if kp.ProjectID() != "p-42" {
		t.Errorf("ProjectID() = %q", kp.ProjectID())
	}
	if kp.Err() != nil {
		t.Errorf("Err() = %v", kp.Err())
	}
}

func TestKeyPair_IntoProject_URIRef(t *testing.T) {
	kp := NewKeyPair().InProject(URI("/projects/p-uri"))
	if kp.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", kp.ProjectID())
	}
	if kp.Err() != nil {
		t.Errorf("Err() = %v", kp.Err())
	}
}

func TestKeyPair_IntoProject_BadRef(t *testing.T) {
	kp := NewKeyPair().InProject(URI("/something/else"))
	if kp.Err() == nil {
		t.Error("expected Err() != nil for unresolvable Ref")
	}
}

// --------------------------------------------------------------------------
// WithPublicKey
// --------------------------------------------------------------------------

func TestKeyPair_WithPublicKey(t *testing.T) {
	kp := NewKeyPair().WithPublicKey("ssh-rsa AAAAB3Nza...")
	if kp.PublicKey() != "ssh-rsa AAAAB3Nza..." {
		t.Errorf("PublicKey() = %q", kp.PublicKey())
	}
	if kp.Err() != nil {
		t.Errorf("Err() = %v", kp.Err())
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestKeyPair_ToRequestRoundTrip(t *testing.T) {
	kp := NewKeyPair().Named(
		"my-keypair").
		Tagged("t1").Tagged("t2").
		InRegion(RegionITBGBergamo).
		WithPublicKey("ssh-rsa AAAA...")

	req := kp.RawRequest()

	if req.Metadata.Name != "my-keypair" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if len(req.Metadata.Tags) != 2 {
		t.Errorf("Metadata.Tags = %v", req.Metadata.Tags)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Location.Value = %q", req.Metadata.Location.Value)
	}
	if req.Properties.Value != "ssh-rsa AAAA..." {
		t.Errorf("Properties.Value = %q", req.Properties.Value)
	}
}

func TestKeyPair_ToRequest_UnsetPublicKey(t *testing.T) {
	kp := NewKeyPair().
		Named("bare")
	req := kp.RawRequest()
	if req.Properties.Value != "" {
		t.Errorf("Properties.Value should be empty, got %q", req.Properties.Value)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func keyPairTestResponse(id, name, uri string) *types.KeyPairResponse {
	loc := &types.LocationResponse{Value: RegionITBGBergamo}
	return &types.KeyPairResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			URI:              &uri,
			Name:             &name,
			Tags:             []string{"tag1"},
			LocationResponse: loc,
		},
		Properties: types.KeyPairPropertiesResponse{
			Value: "ssh-rsa AAAA...",
		},
	}
}

func TestKeyPair_FromResponseHydration(t *testing.T) {
	kp := &KeyPair{}
	resp := keyPairTestResponse("kp-1", "allow-ssh", "/projects/p/providers/Aruba.Compute/keyPairs/kp-1")
	kp.fromResponse(resp)

	if kp.ID() != "kp-1" {
		t.Errorf("ID() = %q", kp.ID())
	}
	if kp.URI() != "/projects/p/providers/Aruba.Compute/keyPairs/kp-1" {
		t.Errorf("URI() = %q", kp.URI())
	}
	if kp.KeyPairID() != "kp-1" {
		t.Errorf("KeyPairID() = %q", kp.KeyPairID())
	}
	if kp.Name() != "allow-ssh" {
		t.Errorf("Name() = %q", kp.Name())
	}
	if tags := kp.Tags(); len(tags) != 1 || tags[0] != "tag1" {
		t.Errorf("Tags() = %v", tags)
	}
	if kp.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", kp.Region())
	}
	if kp.PublicKey() != "ssh-rsa AAAA..." {
		t.Errorf("PublicKey() = %q", kp.PublicKey())
	}
	if kp.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
	// ProjectID backfilled from URI when ProjectMetadataResponse is nil
	if kp.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", kp.ProjectID())
	}
}

func TestKeyPair_FromResponseURIBackfill(t *testing.T) {
	id := "kp-99"
	uri := "/projects/p-uri/providers/Aruba.Compute/keyPairs/kp-99"
	resp := &types.KeyPairResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  &id,
			URI: &uri,
		},
	}
	kp := &KeyPair{}
	kp.fromResponse(resp)

	if kp.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() via URI backfill = %q", kp.ProjectID())
	}
}

func TestKeyPair_FromResponse_NilSafe(t *testing.T) {
	kp := &KeyPair{}
	kp.fromResponse(nil)
	if kp.ID() != "" || kp.URI() != "" || kp.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}

	kp2 := &KeyPair{}
	kp2.fromResponse(&types.KeyPairResponse{})
	if kp2.ID() != "" || kp2.URI() != "" || kp2.PublicKey() != "" {
		t.Error("empty response should yield zero accessor values")
	}
}

// --------------------------------------------------------------------------
// Ref + ancestor ID satisfaction (runtime)
// --------------------------------------------------------------------------

func TestKeyPair_RefSatisfaction(t *testing.T) {
	kp := &KeyPair{}
	kp.fromResponse(keyPairTestResponse("kp-99", "n", "/projects/p99/providers/Aruba.Compute/keyPairs/kp-99"))

	// withKeyPairID typed path
	kid, ok := extractID(kp, func(ref Ref) (string, bool) {
		if w, ok := ref.(withKeyPairID); ok {
			return w.KeyPairID(), true
		}
		return "", false
	}, "keyPairs")
	if !ok || kid != "kp-99" {
		t.Errorf("extractID via withKeyPairID = (%q, %v)", kid, ok)
	}

	// withProjectID typed path
	pid, ok := extractID(kp, func(ref Ref) (string, bool) {
		if w, ok := ref.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || pid != "p99" {
		t.Errorf("extractID via withProjectID = (%q, %v)", pid, ok)
	}
}

// --------------------------------------------------------------------------
// keyPairIDsFromRef helper
// --------------------------------------------------------------------------

func TestKeyPairIDsFromRef_TypedRef(t *testing.T) {
	kp := &KeyPair{}
	kp.fromResponse(keyPairTestResponse("kp-1", "n", "/projects/p/providers/Aruba.Compute/keyPairs/kp-1"))
	pid, kid, err := keyPairIDsFromRef(kp)
	if err != nil || pid != "p" || kid != "kp-1" {
		t.Errorf("keyPairIDsFromRef typed = (%q, %q, %v)", pid, kid, err)
	}
}

func TestKeyPairIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Compute/keyPairs/kp-1")
	pid, kid, err := keyPairIDsFromRef(ref)
	if err != nil || pid != "p" || kid != "kp-1" {
		t.Errorf("keyPairIDsFromRef URI = (%q, %q, %v)", pid, kid, err)
	}
}

func TestKeyPairIDsFromRef_BadURI_MissingKeyPair(t *testing.T) {
	_, _, err := keyPairIDsFromRef(URI("/projects/p/providers/Aruba.Compute"))
	if err == nil {
		t.Error("expected error for URI without /keyPairs/<id>")
	}
}

func TestKeyPairIDsFromRef_BadURI_MissingProject(t *testing.T) {
	_, _, err := keyPairIDsFromRef(URI("/providers/Aruba.Compute/keyPairs/kp-1"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

func TestKeyPairIDsFromRef_BadURI_MissingAll(t *testing.T) {
	_, _, err := keyPairIDsFromRef(URI("/something/else"))
	if err == nil {
		t.Error("expected error for completely unrelated URI")
	}
}

// --------------------------------------------------------------------------
// keyPairsClientAdapter — HTTP mock tests
// --------------------------------------------------------------------------

func buildKeyPairsTestAdapter(t *testing.T, handler http.HandlerFunc) *keyPairsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newKeyPairsClientAdapter(testutil.NewClient(t, server.URL))
}

const keyPairSuccessBody = `{` +
	`"metadata":{"id":"kp-1","name":"allow-ssh","uri":"/projects/p/providers/Aruba.Compute/keyPairs/kp-1"},` +
	`"properties":{"value":"ssh-rsa AAAA..."}}`

func TestKeyPairsClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.KeyPairRequest
	adapter := buildKeyPairsTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "keyPairs") {
			t.Errorf("path %q should contain 'keypairs'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, keyPairSuccessBody)
	})

	kp := NewKeyPair().
		InProject(URI("/projects/p")).
		Named("allow-ssh").
		InRegion(RegionITBGBergamo).
		WithPublicKey("ssh-rsa AAAA...")

	result, err := adapter.Create(context.Background(), kp)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "kp-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "allow-ssh" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Metadata.Name != "allow-ssh" {
		t.Errorf("request Name = %q", gotBody.Metadata.Name)
	}
	if gotBody.Properties.Value != "ssh-rsa AAAA..." {
		t.Errorf("request Properties.Value = %q", gotBody.Properties.Value)
	}
}

func TestKeyPairsClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildKeyPairsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	_, err := adapter.Create(context.Background(), NewKeyPair().
		Named("x").WithPublicKey("ssh-rsa"))
	if err == nil {
		t.Fatal("expected error when KeyPair has no parent project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without parent project")
	}
}

func TestKeyPairsClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildKeyPairsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" field — triggers MetadataValidationError
		fmt.Fprint(w, `{"metadata":{"name":"kp","uri":"/projects/p/providers/Aruba.Compute/keyPairs/x"},"properties":{}}`)
	})

	kp := NewKeyPair().InProject(URI("/projects/p")).
		Named("kp").WithPublicKey("ssh-rsa AAAA...")
	result, err := adapter.Create(context.Background(), kp)
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

func TestKeyPairsClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildKeyPairsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "value is required", 422))
	})

	kp := NewKeyPair().InProject(URI("/projects/p")).WithPublicKey("ssh-rsa")
	result, err := adapter.Create(context.Background(), kp)
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

func TestKeyPairsClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildKeyPairsTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, keyPairSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Compute/keyPairs/kp-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "kp-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	if !containsSubstring(capturedPath, "keyPairs") {
		t.Errorf("path %q should contain 'keypairs'", capturedPath)
	}
}

func TestKeyPairsClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildKeyPairsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, keyPairSuccessBody)
	})

	existing := &KeyPair{}
	existing.fromResponse(keyPairTestResponse("kp-1", "n", "/projects/p/providers/Aruba.Compute/keyPairs/kp-1"))

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "kp-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
}

func TestKeyPairsClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildKeyPairsTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Compute/keyPairs/kp-1"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestKeyPairsClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildKeyPairsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "key pair not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Compute/keyPairs/missing"))
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

func TestKeyPairsClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildKeyPairsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"kp-1","name":"n1","uri":"/projects/p/providers/Aruba.Compute/keyPairs/kp-1"},"properties":{"value":"ssh-rsa AAAA..."}},`+
			`{"metadata":{"id":"kp-2","name":"n2","uri":"/projects/p/providers/Aruba.Compute/keyPairs/kp-2"},"properties":{"value":"ssh-rsa BBBB..."}}`+
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
	if items[0].ID() != "kp-1" || items[0].Name() != "n1" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[0].PublicKey() != "ssh-rsa AAAA..." {
		t.Errorf("items[0].PublicKey() = %q", items[0].PublicKey())
	}
	if items[1].ID() != "kp-2" || items[1].PublicKey() != "ssh-rsa BBBB..." {
		t.Errorf("items[1] ID=%q PublicKey=%q", items[1].ID(), items[1].PublicKey())
	}
	if items[0].ProjectID() != "p" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
}

// --------------------------------------------------------------------------
// Get — bad ref and non-2xx
// --------------------------------------------------------------------------

func TestKeyPairsClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildKeyPairsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := adapter.Get(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad ref")
	}
}

func TestKeyPairsClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildKeyPairsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "key pair not found", 404))
	})

	ref := URI("/projects/p/providers/Aruba.Compute/keyPairs/kp-missing")
	result, err := adapter.Get(context.Background(), ref)
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
	if result == nil {
		t.Fatal("result must be non-nil on non-2xx")
	}
}

// --------------------------------------------------------------------------
// Delete — bad ref
// --------------------------------------------------------------------------

func TestKeyPairsClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildKeyPairsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad ref")
	}
}

// --------------------------------------------------------------------------
// List — bad ref and non-2xx
// --------------------------------------------------------------------------

func TestKeyPairsClientAdapter_List_BadRef(t *testing.T) {
	adapter := buildKeyPairsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := adapter.List(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad project ref")
	}
}

func TestKeyPairsClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildKeyPairsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Forbidden", "access denied", 403))
	})

	_, err := adapter.List(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected error on 403")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

func TestKeyPair_FromResponse_SetsStatus(t *testing.T) {
	k := &KeyPair{}
	state := types.State("Active")
	k.fromResponse(&types.KeyPairResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if k.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", k.State())
	}
}

func TestKeyPair_WaitUntilActive_HappyPath(t *testing.T) {
	k := &KeyPair{}
	calls := 0
	state := types.State("InCreation")
	k.setRefresh(func(_ context.Context) error {
		calls++
		if calls >= 2 {
			state = "Active"
		}
		s := state
		k.setStatus(&types.ResourceStatusResponse{State: &s})
		return nil
	})
	if err := k.WaitUntilActive(context.Background(), fastOpts()...); err != nil {
		t.Fatalf("WaitUntilActive error: %v", err)
	}
	if k.State() != "Active" {
		t.Errorf("State() = %q after wait, want Active", k.State())
	}
}
