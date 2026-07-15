package aruba

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

const testPassword = "Prova123456789AC@"

// --------------------------------------------------------------------------
// Compile-time satisfaction checks
// --------------------------------------------------------------------------

var (
	_ Ref     = (*User)(nil)
	_ Wrapper = (*User)(nil)
)

// --------------------------------------------------------------------------
// No Password() accessor — acceptance criterion
// --------------------------------------------------------------------------

func TestUser_NoPasswordAccessor(t *testing.T) {
	typ := reflect.TypeOf((*User)(nil))
	if _, ok := typ.MethodByName("Password"); ok {
		t.Error("*User must not expose a Password() accessor — write-only field")
	}
}

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestUser_FluentSetters(t *testing.T) {
	u := NewUser().
		InDBaaS(URI("/projects/p-1/providers/Aruba.Database/dbaas/d-1")).
		WithUsername("alice").
		WithPassword(testPassword)

	if u.Username() != "alice" {
		t.Errorf("Username() = %q", u.Username())
	}
	if u.ID() != "alice" {
		t.Errorf("ID() = %q", u.ID())
	}
	if u.DBaaSID() != "d-1" {
		t.Errorf("DBaaSID() = %q", u.DBaaSID())
	}
	if u.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", u.ProjectID())
	}
	if u.Err() != nil {
		t.Errorf("Err() = %v", u.Err())
	}
	// Password should not be readable through Username/ID/etc., only through RawRequest.
	wantPw := base64.StdEncoding.EncodeToString([]byte(testPassword))
	if u.RawRequest().Password != wantPw {
		t.Errorf("RawRequest().Password = %q, want %q", u.RawRequest().Password, wantPw)
	}
}

// --------------------------------------------------------------------------
// IntoDBaaS — typed / URI / bad Ref
// --------------------------------------------------------------------------

func TestUser_IntoDBaaS_TypedRef(t *testing.T) {
	dbaas := &DBaaS{}
	dbaas.fromResponse(dbaasTestResponse("d-1", "my-dbaas", "/projects/p-1/providers/Aruba.Database/dbaas/d-1"))

	u := NewUser().InDBaaS(dbaas)
	if u.DBaaSID() != "d-1" {
		t.Errorf("DBaaSID() = %q", u.DBaaSID())
	}
	if u.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", u.ProjectID())
	}
	if u.Err() != nil {
		t.Errorf("Err() = %v", u.Err())
	}
}

func TestUser_IntoDBaaS_URIRef(t *testing.T) {
	u := NewUser().InDBaaS(URI("/projects/p-uri/providers/Aruba.Database/dbaas/d-uri"))
	if u.DBaaSID() != "d-uri" {
		t.Errorf("DBaaSID() = %q", u.DBaaSID())
	}
	if u.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", u.ProjectID())
	}
	if u.Err() != nil {
		t.Errorf("Err() = %v", u.Err())
	}
}

func TestUser_IntoDBaaS_BadRef(t *testing.T) {
	u := NewUser().InDBaaS(URI("/something/garbage"))
	if u.Err() == nil {
		t.Error("expected Err() to be set for unresolvable parent")
	}
}

// --------------------------------------------------------------------------
// URI construction
// --------------------------------------------------------------------------

func TestUser_URI_Constructed(t *testing.T) {
	u := NewUser().
		InDBaaS(URI("/projects/p-1/providers/Aruba.Database/dbaas/d-1")).
		WithUsername("alice")
	want := "/projects/p-1/providers/Aruba.Database/dbaas/d-1/users/alice"
	if u.URI() != want {
		t.Errorf("URI() = %q, want %q", u.URI(), want)
	}
}

func TestUser_URI_MissingProjectID(t *testing.T) {
	u := &User{}
	u.dbaasID = "d-1"
	name := "alice"
	u.username = &name
	if u.URI() != "" {
		t.Errorf("URI() should be empty when projectID missing, got %q", u.URI())
	}
}

func TestUser_URI_MissingDBaaSID(t *testing.T) {
	u := &User{}
	u.projectID = "p-1"
	name := "alice"
	u.username = &name
	if u.URI() != "" {
		t.Errorf("URI() should be empty when dbaasID missing, got %q", u.URI())
	}
}

func TestUser_URI_MissingUsername(t *testing.T) {
	u := &User{}
	u.projectID = "p-1"
	u.dbaasID = "d-1"
	if u.URI() != "" {
		t.Errorf("URI() should be empty when username missing, got %q", u.URI())
	}
}

// --------------------------------------------------------------------------
// toRequest
// --------------------------------------------------------------------------

func TestUser_ToRequest(t *testing.T) {
	u := NewUser().WithUsername("alice").WithPassword(testPassword)
	req := u.toRequest()
	if req.Username != "alice" {
		t.Errorf("toRequest().Username = %q", req.Username)
	}
	wantPw := base64.StdEncoding.EncodeToString([]byte(testPassword))
	if req.Password != wantPw {
		t.Errorf("toRequest().Password = %q, want %q", req.Password, wantPw)
	}
}

func TestUser_ToRequest_EncodesPassword(t *testing.T) {
	u := NewUser().WithPassword(testPassword)
	encoded := u.toRequest().Password
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("toRequest().Password is not valid base64: %v", err)
	}
	if string(decoded) != testPassword {
		t.Errorf("decoded password = %q, want %q", string(decoded), testPassword)
	}
}

func TestUser_ToRequest_Empty(t *testing.T) {
	u := &User{}
	req := u.toRequest()
	if req.Username != "" {
		t.Errorf("toRequest().Username should be empty, got %q", req.Username)
	}
	if req.Password != "" {
		t.Errorf("toRequest().Password should be empty, got %q", req.Password)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func userTestResponse(username string) *types.UserResponse {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	creator := "admin@example.com"
	return &types.UserResponse{
		Username:     username,
		CreationDate: &ts,
		CreatedBy:    &creator,
	}
}

func TestUser_FromResponseHydration(t *testing.T) {
	resp := userTestResponse("alice")
	u := &User{}
	u.fromResponse(resp)

	if u.Username() != "alice" {
		t.Errorf("Username() = %q", u.Username())
	}
	if u.ID() != "alice" {
		t.Errorf("ID() = %q", u.ID())
	}
	if u.CreatedAt().IsZero() {
		t.Error("CreatedAt() should be non-zero")
	}
	if u.CreatedBy() != "admin@example.com" {
		t.Errorf("CreatedBy() = %q", u.CreatedBy())
	}
	if u.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestUser_FromResponse_NilSafe(t *testing.T) {
	u := &User{}
	u.fromResponse(nil) // must not panic
	if u.ID() != "" || u.Username() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}

	u2 := &User{}
	u2.fromResponse(&types.UserResponse{}) // empty response — Username is ""
	if u2.ID() != "" || u2.Username() != "" {
		t.Error("empty response should yield zero accessor values")
	}
}

func TestUser_FromResponse_PreservesPassword(t *testing.T) {
	u := NewUser().WithPassword(testPassword)
	u.fromResponse(userTestResponse("alice"))
	// The locally-set password must survive hydration.
	wantPw := base64.StdEncoding.EncodeToString([]byte(testPassword))
	if u.RawRequest().Password != wantPw {
		t.Errorf("fromResponse clobbered the locally-set password; got %q", u.RawRequest().Password)
	}
}

// --------------------------------------------------------------------------
// userIDsFromRef helper
// --------------------------------------------------------------------------

func TestUserIDsFromRef_TypedRef(t *testing.T) {
	u := NewUser().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		WithUsername("alice")
	pid, did, name, err := userIDsFromRef(u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid != "p" || did != "d-1" || name != "alice" {
		t.Errorf("userIDsFromRef typed = (%q, %q, %q)", pid, did, name)
	}
}

func TestUserIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/users/alice")
	pid, did, name, err := userIDsFromRef(ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid != "p" || did != "d-1" || name != "alice" {
		t.Errorf("userIDsFromRef URI = (%q, %q, %q)", pid, did, name)
	}
}

func TestUserIDsFromRef_BadURI_MissingUsers(t *testing.T) {
	_, _, _, err := userIDsFromRef(URI("/projects/p/providers/Aruba.Database/dbaas/d-1"))
	if err == nil {
		t.Error("expected error for URI without /users/<name>")
	}
}

func TestUserIDsFromRef_BadURI_MissingDBaaS(t *testing.T) {
	_, _, _, err := userIDsFromRef(URI("/projects/p/providers/Aruba.Database/users/alice"))
	if err == nil {
		t.Error("expected error for URI without /dbaas/<id>")
	}
}

func TestUserIDsFromRef_BadURI_MissingProject(t *testing.T) {
	_, _, _, err := userIDsFromRef(URI("/providers/Aruba.Database/dbaas/d-1/users/alice"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

// --------------------------------------------------------------------------
// usersClientAdapter — HTTP mock tests
// --------------------------------------------------------------------------

func buildUserTestAdapter(t *testing.T, handler http.HandlerFunc) *usersClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newUsersClientAdapter(testutil.NewClient(t, server.URL))
}

const userSuccessBody = `{"username":"my-user","createdBy":"admin@example.com"}`

func TestUsersClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.UserRequest
	callCount := 0
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "users") {
			t.Errorf("path %q should contain 'users'", r.URL.Path)
		}
		if !containsSubstring(r.URL.Path, "dbaas") {
			t.Errorf("path %q should contain 'dbaas'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, userSuccessBody)
	})

	u := NewUser().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		WithUsername("my-user").
		WithPassword(testPassword)

	result, err := adapter.Create(context.Background(), u)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
	if result.Username() != "my-user" {
		t.Errorf("Username() = %q", result.Username())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Username != "my-user" {
		t.Errorf("wire body Username = %q", gotBody.Username)
	}
	wantPw := base64.StdEncoding.EncodeToString([]byte(testPassword))
	if gotBody.Password != wantPw {
		t.Errorf("wire body Password = %q, want %q", gotBody.Password, wantPw)
	}
}

func TestUsersClientAdapter_Create_NoParent(t *testing.T) {
	callCount := 0
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
	})
	u := NewUser().WithUsername("alice").WithPassword(testPassword) // no IntoDBaaS
	_, err := adapter.Create(context.Background(), u)
	if err == nil {
		t.Error("expected error when parent DBaaS is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestUsersClientAdapter_Create_NoUsername(t *testing.T) {
	callCount := 0
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
	})
	u := NewUser().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		WithPassword(testPassword) // no WithUsername
	_, err := adapter.Create(context.Background(), u)
	if err == nil {
		t.Error("expected error when username is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestUsersClientAdapter_Create_NoPassword(t *testing.T) {
	callCount := 0
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
	})
	u := NewUser().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		WithUsername("alice") // no WithPassword
	_, err := adapter.Create(context.Background(), u)
	if err == nil {
		t.Error("expected error when password is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestUsersClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, `{"title":"Unprocessable","detail":"invalid username"}`)
	})

	u := NewUser().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		WithUsername("my-user").
		WithPassword(testPassword)
	_, err := adapter.Create(context.Background(), u)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

func TestUsersClientAdapter_Update_Success(t *testing.T) {
	var gotBody types.UserRequest
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "users/my-user") {
			t.Errorf("path %q should contain 'users/my-user'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, userSuccessBody)
	})

	u := NewUser().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		WithUsername("my-user").
		WithPassword(testPassword)

	result, err := adapter.Update(context.Background(), u)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if result.Username() != "my-user" {
		t.Errorf("Username() = %q", result.Username())
	}
	if gotBody.Username != "my-user" {
		t.Errorf("wire body Username = %q", gotBody.Username)
	}
	wantPw := base64.StdEncoding.EncodeToString([]byte(testPassword))
	if gotBody.Password != wantPw {
		t.Errorf("wire body Password = %q, want %q", gotBody.Password, wantPw)
	}
}

func TestUsersClientAdapter_Update_NoIDs(t *testing.T) {
	callCount := 0
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
	})
	u := NewUser() // no IntoDBaaS, no WithUsername, no WithPassword
	_, err := adapter.Update(context.Background(), u)
	if err == nil {
		t.Error("expected error when IDs are missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestUsersClientAdapter_Update_NoPassword(t *testing.T) {
	callCount := 0
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
	})
	u := NewUser().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		WithUsername("alice") // no WithPassword
	_, err := adapter.Update(context.Background(), u)
	if err == nil {
		t.Error("expected error when password is missing — Update rotates the password")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestUsersClientAdapter_Get_URIRef(t *testing.T) {
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if !containsSubstring(r.URL.Path, "users/my-user") {
			t.Errorf("path %q should contain 'users/my-user'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, userSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/users/my-user")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.Username() != "my-user" {
		t.Errorf("Username() = %q", result.Username())
	}
	if result.DBaaSID() != "d-1" {
		t.Errorf("DBaaSID() = %q", result.DBaaSID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
}

func TestUsersClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, userSuccessBody)
	})

	u := NewUser().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		WithUsername("my-user")

	result, err := adapter.Get(context.Background(), u)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.Username() != "my-user" {
		t.Errorf("Username() = %q", result.Username())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
}

func TestUsersClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/users/my-user")
	if err := adapter.Delete(context.Background(), ref); err != nil {
		t.Errorf("Delete error: %v", err)
	}
}

func TestUsersClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"title":"Not Found","detail":"user not found"}`)
	})
	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/users/my-user")
	err := adapter.Delete(context.Background(), ref)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Zero-value accessors (Shape F — covers the nil-response branch)
// --------------------------------------------------------------------------

func TestUser_Accessors_ZeroValue(t *testing.T) {
	u := &User{}
	if !u.CreatedAt().IsZero() {
		t.Error("CreatedAt() zero should be zero time")
	}
	if u.CreatedBy() != "" {
		t.Errorf("CreatedBy() zero = %q", u.CreatedBy())
	}
}

// --------------------------------------------------------------------------
// Create — conflict response
// --------------------------------------------------------------------------

func TestUsersClientAdapter_Create_Conflict(t *testing.T) {
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `{"title":"Conflict","detail":"username taken"}`)
	})
	u := NewUser().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		WithUsername("alice").
		WithPassword(testPassword)
	_, err := adapter.Create(context.Background(), u)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusConflict {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Update — extra guard paths
// --------------------------------------------------------------------------

func TestUsersClientAdapter_Update_NoDBaaS(t *testing.T) {
	callCount := 0
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	// Has username and password but no parent DBaaS
	u := NewUser().WithUsername("alice").WithPassword(testPassword)
	_, err := adapter.Update(context.Background(), u)
	if err == nil {
		t.Fatal("expected error when DBaaS is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestUsersClientAdapter_Update_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	// Has username, password, dbaasID but no projectID
	u := &User{}
	u.dbaasID = "d-1"
	name := "alice"
	u.username = &name
	pw := testPassword
	u.password = &pw
	_, err := adapter.Update(context.Background(), u)
	if err == nil {
		t.Fatal("expected error when projectID is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestUsersClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `{"title":"Conflict","detail":"concurrent update"}`)
	})
	u := NewUser().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		WithUsername("alice").
		WithPassword(testPassword)
	_, err := adapter.Update(context.Background(), u)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusConflict {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Get — bad Ref and non-2xx
// --------------------------------------------------------------------------

func TestUsersClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.Get(context.Background(), URI("/something/bad"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
}

func TestUsersClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"title":"Not Found","detail":"user not found"}`)
	})
	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/users/alice")
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
// Delete — bad Ref
// --------------------------------------------------------------------------

func TestUsersClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := adapter.Delete(context.Background(), URI("/something/bad"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
}

// --------------------------------------------------------------------------
// Create — errMixin path
// --------------------------------------------------------------------------

func TestUsersClientAdapter_Create_ErrMixin(t *testing.T) {
	callCount := 0
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
	})
	// IntoDBaaS with bad URI sets errMixin
	u := NewUser().InDBaaS(URI("/garbage")).WithUsername("alice").WithPassword(testPassword)
	_, err := adapter.Create(context.Background(), u)
	if err == nil {
		t.Fatal("expected error from errMixin")
	}
	if callCount != 0 {
		t.Errorf("expected 0 HTTP calls, got %d", callCount)
	}
}

// --------------------------------------------------------------------------
// List — bad parent ref and non-2xx
// --------------------------------------------------------------------------

func TestUsersClientAdapter_List_BadRef(t *testing.T) {
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/something/bad"))
	if err == nil {
		t.Fatal("expected error for bad parent Ref")
	}
}

func TestUsersClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"title":"Forbidden","detail":"not allowed"}`)
	})
	parent := URI("/projects/p/providers/Aruba.Database/dbaas/d-1")
	_, err := adapter.List(context.Background(), parent)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

func TestUsersClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildUserTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if !containsSubstring(r.URL.Path, "users") {
			t.Errorf("path %q should contain 'users'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"values":[{"username":"alice"},{"username":"bob"}]}`)
	})

	parent := URI("/projects/p/providers/Aruba.Database/dbaas/d-1")
	list, err := adapter.List(context.Background(), parent)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	items := list.Items()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Username() != "alice" {
		t.Errorf("items[0].Username() = %q", items[0].Username())
	}
	if items[1].Username() != "bob" {
		t.Errorf("items[1].Username() = %q", items[1].Username())
	}
	for i, item := range items {
		if item.DBaaSID() != "d-1" {
			t.Errorf("items[%d].DBaaSID() = %q", i, item.DBaaSID())
		}
		if item.ProjectID() != "p" {
			t.Errorf("items[%d].ProjectID() = %q", i, item.ProjectID())
		}
	}
	if list.Total() != 2 {
		t.Errorf("Total() = %d", list.Total())
	}
}
