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

// keyMakeKMSParent creates a KMS wrapper with specific projectID and kmsID
// for use as an IntoKMS parent in Key tests.
func keyMakeKMSParent(projectID, kmsID string) *KMS {
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

func TestKey_FluentSetters(t *testing.T) {
	parent := keyMakeKMSParent("p-1", "kms-1")
	k := NewKey().
		InKMS(parent).
		Named("my-key").
		OfAlgorithm(KeyAlgorithmAes)

	if k.Name() != "my-key" {
		t.Errorf("Name() = %q", k.Name())
	}
	if k.Algorithm() != KeyAlgorithmAes {
		t.Errorf("Algorithm() = %q", k.Algorithm())
	}
	if k.KMSID() != "kms-1" {
		t.Errorf("KMSID() = %q", k.KMSID())
	}
	if k.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", k.ProjectID())
	}
	if k.Err() != nil {
		t.Errorf("Err() = %v", k.Err())
	}
}

// --------------------------------------------------------------------------
// IntoKMS
// --------------------------------------------------------------------------

func TestKey_IntoKMS_TypedRef(t *testing.T) {
	parent := keyMakeKMSParent("p-42", "kms-42")
	k := NewKey().InKMS(parent)
	if k.KMSID() != "kms-42" {
		t.Errorf("KMSID() = %q", k.KMSID())
	}
	if k.ProjectID() != "p-42" {
		t.Errorf("ProjectID() = %q", k.ProjectID())
	}
	if k.Err() != nil {
		t.Errorf("Err() = %v", k.Err())
	}
}

func TestKey_IntoKMS_URIRef(t *testing.T) {
	k := NewKey().InKMS(URI("/projects/p-uri/providers/Aruba.Security/kms/kms-uri"))
	if k.KMSID() != "kms-uri" {
		t.Errorf("KMSID() = %q", k.KMSID())
	}
	if k.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", k.ProjectID())
	}
	if k.Err() != nil {
		t.Errorf("Err() = %v", k.Err())
	}
}

func TestKey_IntoKMS_BadRef_NoKMS(t *testing.T) {
	k := NewKey().InKMS(URI("/projects/p-1"))
	if k.Err() == nil {
		t.Error("expected Err() != nil for URI missing kms segment")
	}
}

func TestKey_IntoKMS_BadRef_NoProject(t *testing.T) {
	k := NewKey().InKMS(URI("/providers/Aruba.Security/kms/kms-1"))
	if k.Err() == nil {
		t.Error("expected Err() != nil for URI missing project segment")
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestKey_ToRequest_FullyPopulated(t *testing.T) {
	k := NewKey().
		Named("enc-key").OfAlgorithm(KeyAlgorithmRsa)
	req := k.RawRequest()
	if req.Name != "enc-key" {
		t.Errorf("Name = %q", req.Name)
	}
	if req.Algorithm != KeyAlgorithmRsa {
		t.Errorf("Algorithm = %q", req.Algorithm)
	}
}

func TestKey_ToRequest_ZeroState(t *testing.T) {
	req := NewKey().RawRequest()
	if req.Name != "" {
		t.Errorf("Name = %q, want empty", req.Name)
	}
	if req.Algorithm != "" {
		t.Errorf("Algorithm = %q, want empty", req.Algorithm)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

const keySuccessBody = `{` +
	`"keyId":"key-1",` +
	`"privateKeyId":"pk-1",` +
	`"name":"my-key",` +
	`"algorithm":"Aes",` +
	`"creationSource":"Cmp",` +
	`"type":"Symmetric",` +
	`"status":"Active"}`

func TestKey_FromResponseHydration(t *testing.T) {
	algo := KeyAlgorithmAes
	cs := KeyCreationSourceCmp
	kt := KeyTypeSymmetric
	ks := KeyStatusActive
	keyID := "key-1"
	pkID := "pk-1"
	name := "my-key"

	resp := &types.KeyResponse{
		KeyID:          &keyID,
		PrivateKeyID:   &pkID,
		Name:           &name,
		Algorithm:      &algo,
		CreationSource: &cs,
		Type:           &kt,
		Status:         &ks,
	}

	k := NewKey()
	k.projectID = "p-1"
	k.kmsID = "kms-1"
	k.fromResponse(resp)

	if k.ID() != "key-1" {
		t.Errorf("ID() = %q", k.ID())
	}
	if k.KeyID() != "key-1" {
		t.Errorf("KeyID() = %q", k.KeyID())
	}
	if k.PrivateKeyID() != "pk-1" {
		t.Errorf("PrivateKeyID() = %q", k.PrivateKeyID())
	}
	if k.Name() != "my-key" {
		t.Errorf("Name() = %q", k.Name())
	}
	if k.Algorithm() != KeyAlgorithmAes {
		t.Errorf("Algorithm() = %q", k.Algorithm())
	}
	if k.CreationSource() != string(KeyCreationSourceCmp) {
		t.Errorf("CreationSource() = %q", k.CreationSource())
	}
	if k.Type() != string(KeyTypeSymmetric) {
		t.Errorf("Type() = %q", k.Type())
	}
	if k.KeyStatus() != "Active" {
		t.Errorf("KeyStatus() = %q", k.KeyStatus())
	}
	if k.Raw() == nil {
		t.Error("Raw() nil after fromResponse")
	}
	if k.URI() == "" {
		t.Error("URI() empty after hydration with projectID+kmsID set")
	}
}

func TestKey_FromResponse_Nil(t *testing.T) {
	k := NewKey()
	k.fromResponse(nil)
	if k.response != nil {
		t.Error("response should remain nil after fromResponse(nil)")
	}
}

// --------------------------------------------------------------------------
// URI construction
// --------------------------------------------------------------------------

func TestKey_URI_AllAncestorsPresent(t *testing.T) {
	k := NewKey()
	k.projectID = "p-1"
	k.kmsID = "kms-1"
	keyID := "key-1"
	k.fromResponse(&types.KeyResponse{KeyID: &keyID})
	want := "/projects/p-1/providers/Aruba.Security/kms/kms-1/keys/key-1"
	if k.URI() != want {
		t.Errorf("URI() = %q, want %q", k.URI(), want)
	}
}

func TestKey_URI_MissingKeyID(t *testing.T) {
	k := NewKey()
	k.projectID = "p-1"
	k.kmsID = "kms-1"
	if k.URI() != "" {
		t.Errorf("URI() = %q, want empty when keyID missing", k.URI())
	}
}

func TestKey_URI_MissingKMSID(t *testing.T) {
	k := NewKey()
	k.projectID = "p-1"
	keyID := "key-1"
	k.fromResponse(&types.KeyResponse{KeyID: &keyID})
	if k.URI() != "" {
		t.Errorf("URI() = %q, want empty when kmsID missing", k.URI())
	}
}

func TestKey_URI_MissingProjectID(t *testing.T) {
	k := NewKey()
	k.kmsID = "kms-1"
	keyID := "key-1"
	k.fromResponse(&types.KeyResponse{KeyID: &keyID})
	if k.URI() != "" {
		t.Errorf("URI() = %q, want empty when projectID missing", k.URI())
	}
}

// --------------------------------------------------------------------------
// Zero-value accessors (Shape F)
// --------------------------------------------------------------------------

func TestKey_Accessors_ZeroValue(t *testing.T) {
	k := NewKey()
	if k.ID() != "" {
		t.Errorf("ID() = %q, want empty", k.ID())
	}
	if k.KeyID() != "" {
		t.Errorf("KeyID() = %q, want empty", k.KeyID())
	}
	if k.Name() != "" {
		t.Errorf("Name() = %q, want empty", k.Name())
	}
	if k.Algorithm() != "" {
		t.Errorf("Algorithm() = %q, want empty", k.Algorithm())
	}
	if k.Type() != "" {
		t.Errorf("Type() = %q, want empty", k.Type())
	}
	if k.KeyStatus() != "" {
		t.Errorf("KeyStatus() = %q, want empty", k.KeyStatus())
	}
	if k.CreationSource() != "" {
		t.Errorf("CreationSource() = %q, want empty", k.CreationSource())
	}
	if k.PrivateKeyID() != "" {
		t.Errorf("PrivateKeyID() = %q, want empty", k.PrivateKeyID())
	}
	if k.URI() != "" {
		t.Errorf("URI() = %q, want empty", k.URI())
	}
	if k.Raw() != nil {
		t.Error("Raw() non-nil on zero Key")
	}
}

// Algorithm falls back to local field (no response)
func TestKey_Algorithm_LocalFallback(t *testing.T) {
	k := NewKey().OfAlgorithm(KeyAlgorithmRsa)
	if k.Algorithm() != KeyAlgorithmRsa {
		t.Errorf("Algorithm() = %q, want Rsa", k.Algorithm())
	}
}

// Name falls back to local field (no response)
func TestKey_Name_LocalFallback(t *testing.T) {
	k := NewKey().
		Named("fallback-name")
	if k.Name() != "fallback-name" {
		t.Errorf("Name() = %q, want fallback-name", k.Name())
	}
}

// --------------------------------------------------------------------------
// keyIDsFromRef
// --------------------------------------------------------------------------

func TestKeyIDsFromRef_TypedRef(t *testing.T) {
	k := NewKey()
	k.projectID = "p-1"
	k.kmsID = "kms-1"
	keyID := "key-1"
	k.fromResponse(&types.KeyResponse{KeyID: &keyID})

	pid, kid, kID, err := keyIDsFromRef(k)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid != "p-1" || kid != "kms-1" || kID != "key-1" {
		t.Errorf("got (%q,%q,%q)", pid, kid, kID)
	}
}

func TestKeyIDsFromRef_URIRef(t *testing.T) {
	pid, kid, kID, err := keyIDsFromRef(URI("/projects/p-2/providers/Aruba.Security/kms/kms-2/keys/key-2"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid != "p-2" || kid != "kms-2" || kID != "key-2" {
		t.Errorf("got (%q,%q,%q)", pid, kid, kID)
	}
}

func TestKeyIDsFromRef_BadURI_NoKey(t *testing.T) {
	_, _, _, err := keyIDsFromRef(URI("/projects/p/providers/Aruba.Security/kms/kms-1"))
	if err == nil {
		t.Error("expected error for URI missing keys segment")
	}
}

func TestKeyIDsFromRef_BadURI_NoKMS(t *testing.T) {
	_, _, _, err := keyIDsFromRef(URI("/projects/p/providers/Aruba.Security/keys/key-1"))
	if err == nil {
		t.Error("expected error for URI missing kms segment")
	}
}

func TestKeyIDsFromRef_BadURI_NoProject(t *testing.T) {
	_, _, _, err := keyIDsFromRef(URI("/providers/Aruba.Security/kms/kms-1/keys/key-1"))
	if err == nil {
		t.Error("expected error for URI missing projects segment")
	}
}

// --------------------------------------------------------------------------
// HTTP-mock adapter helper
// --------------------------------------------------------------------------

func buildKeyTestAdapter(t *testing.T, handler http.HandlerFunc) *keysClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newKeysClientAdapter(testutil.NewClient(t, server.URL))
}

// --------------------------------------------------------------------------
// Create adapter tests
// --------------------------------------------------------------------------

func TestKeysClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.KeyRequest
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, keySuccessBody)
	})

	parent := keyMakeKMSParent("p-1", "kms-1")
	k := NewKey().InKMS(parent).
		Named("my-key").OfAlgorithm(KeyAlgorithmAes)

	result, err := adapter.Create(context.Background(), k)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.KeyID() != "key-1" {
		t.Errorf("KeyID() = %q", result.KeyID())
	}
	if result.Name() != "my-key" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.Algorithm() != KeyAlgorithmAes {
		t.Errorf("Algorithm() = %q", result.Algorithm())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Name != "my-key" {
		t.Errorf("request Name = %q", gotBody.Name)
	}
}

func TestKeysClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})
	k := NewKey().
		Named("x").OfAlgorithm(KeyAlgorithmAes)
	_, err := adapter.Create(context.Background(), k)
	if err == nil {
		t.Fatal("expected error when Key has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestKeysClientAdapter_Create_NoKMS(t *testing.T) {
	callCount := 0
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})
	k := NewKey()
	k.projectID = "p-1" // project set but not KMS
	k.Named("x").OfAlgorithm(KeyAlgorithmAes)
	_, err := adapter.Create(context.Background(), k)
	if err == nil {
		t.Fatal("expected error when Key has no KMS ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without KMS ID")
	}
}

func TestKeysClientAdapter_Create_ErrMixin(t *testing.T) {
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	k := NewKey().InKMS(URI("not-a-valid-kms-uri"))
	_, err := adapter.Create(context.Background(), k)
	if err == nil {
		t.Fatal("expected error from errMixin when IntoKMS failed")
	}
}

func TestKeysClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"bad request"}`)
	})
	parent := keyMakeKMSParent("p-1", "kms-1")
	_, err := adapter.Create(context.Background(), NewKey().InKMS(parent).
		Named("k").OfAlgorithm(KeyAlgorithmAes))
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

func TestKeysClientAdapter_Get_URIRef(t *testing.T) {
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, keySuccessBody)
	})

	result, err := adapter.Get(context.Background(), URI(
		"/projects/p-1/providers/Aruba.Security/kms/kms-1/keys/key-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.KeyID() != "key-1" {
		t.Errorf("KeyID() = %q", result.KeyID())
	}
}

func TestKeysClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, keySuccessBody)
	})

	k := NewKey()
	k.projectID = "p-1"
	k.kmsID = "kms-1"
	id := "key-1"
	k.fromResponse(&types.KeyResponse{KeyID: &id})

	result, err := adapter.Get(context.Background(), k)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.KeyID() != "key-1" {
		t.Errorf("KeyID() = %q", result.KeyID())
	}
}

func TestKeysClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.Get(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected error for bad ref missing key/kms IDs")
	}
}

func TestKeysClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	_, err := adapter.Get(context.Background(), URI(
		"/projects/p-1/providers/Aruba.Security/kms/kms-1/keys/key-1"))
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

func TestKeysClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := adapter.Delete(context.Background(), URI(
		"/projects/p-1/providers/Aruba.Security/kms/kms-1/keys/key-1"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestKeysClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := adapter.Delete(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected error for bad ref")
	}
}

func TestKeysClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	err := adapter.Delete(context.Background(), URI(
		"/projects/p-1/providers/Aruba.Security/kms/kms-1/keys/key-1"))
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

const keyListBody = `{` +
	`"total":2,` +
	`"values":[` +
	`{"keyId":"key-1","name":"k1","algorithm":"Aes"},` +
	`{"keyId":"key-2","name":"k2","algorithm":"Rsa"}` +
	`]}`

func TestKeysClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, keyListBody)
	})

	parent := keyMakeKMSParent("p-1", "kms-1")
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
	if items[0].KeyID() != "key-1" {
		t.Errorf("items[0].KeyID() = %q", items[0].KeyID())
	}
	if items[1].Algorithm() != KeyAlgorithmRsa {
		t.Errorf("items[1].Algorithm() = %q", items[1].Algorithm())
	}
}

func TestKeysClientAdapter_List_BadParentRef(t *testing.T) {
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected error for parent ref missing KMS ID")
	}
}

func TestKeysClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildKeyTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	parent := keyMakeKMSParent("p-1", "kms-1")
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
// Nil-rest guard
// --------------------------------------------------------------------------

func TestNewKeysClientAdapter_Nil(t *testing.T) {
	a := newKeysClientAdapter(nil)
	if a == nil {
		t.Fatal("newKeysClientAdapter(nil) returned nil")
	}
}

// --------------------------------------------------------------------------
// Reflective guard: KeysClient must NOT expose Update (Family B, no-Update)
// --------------------------------------------------------------------------

func TestKeysClient_NoUpdateMethod(t *testing.T) {
	iface := reflect.TypeOf((*KeysClient)(nil)).Elem()
	for i := range iface.NumMethod() {
		if iface.Method(i).Name == "Update" {
			t.Error("KeysClient must not expose an Update method — Key is a no-Update Family B resource")
		}
	}
}
