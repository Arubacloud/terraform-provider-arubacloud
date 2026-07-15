package aruba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/internal/clients/database"
	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Compile-time satisfaction checks
// --------------------------------------------------------------------------

var (
	_ Ref     = (*Grant)(nil)
	_ Wrapper = (*Grant)(nil)
)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestGrant_FluentSetters(t *testing.T) {
	g := NewGrant().
		InDatabase(URI("/projects/p-1/providers/Aruba.Database/dbaas/d-1/databases/db-1")).
		ForUser("alice").
		OfRole("READ_WRITE")

	if g.Username() != "alice" {
		t.Errorf("UserName() = %q", g.Username())
	}
	if g.RoleName() != "READ_WRITE" {
		t.Errorf("RoleName() = %q", g.RoleName())
	}
	if g.ID() != "" {
		t.Errorf("ID() should be empty before a Get, got %q", g.ID())
	}
	if g.DatabaseID() != "db-1" {
		t.Errorf("DatabaseID() = %q", g.DatabaseID())
	}
	if g.DBaaSID() != "d-1" {
		t.Errorf("DBaaSID() = %q", g.DBaaSID())
	}
	if g.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", g.ProjectID())
	}
	if g.Err() != nil {
		t.Errorf("Err() = %v", g.Err())
	}
}

// --------------------------------------------------------------------------
// IntoDatabase — typed / URI / bad Ref
// --------------------------------------------------------------------------

func TestGrant_IntoDatabase_TypedRef(t *testing.T) {
	db := NewDatabase().
		InDBaaS(URI("/projects/p-1/providers/Aruba.Database/dbaas/d-1")).
		Named("db-1")

	g := NewGrant().InDatabase(db)
	if g.DatabaseID() != "db-1" {
		t.Errorf("DatabaseID() = %q", g.DatabaseID())
	}
	if g.DBaaSID() != "d-1" {
		t.Errorf("DBaaSID() = %q", g.DBaaSID())
	}
	if g.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", g.ProjectID())
	}
	if g.Err() != nil {
		t.Errorf("Err() = %v", g.Err())
	}
}

func TestGrant_IntoDatabase_URIRef(t *testing.T) {
	g := NewGrant().InDatabase(URI("/projects/p-uri/providers/Aruba.Database/dbaas/d-uri/databases/db-uri"))
	if g.DatabaseID() != "db-uri" {
		t.Errorf("DatabaseID() = %q", g.DatabaseID())
	}
	if g.DBaaSID() != "d-uri" {
		t.Errorf("DBaaSID() = %q", g.DBaaSID())
	}
	if g.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", g.ProjectID())
	}
	if g.Err() != nil {
		t.Errorf("Err() = %v", g.Err())
	}
}

func TestGrant_IntoDatabase_BadRef(t *testing.T) {
	g := NewGrant().InDatabase(URI("/something/garbage"))
	if g.Err() == nil {
		t.Error("expected Err() to be set for unresolvable parent")
	}
}

// --------------------------------------------------------------------------
// URI construction
// --------------------------------------------------------------------------

func TestGrant_URI_Constructed(t *testing.T) {
	g := NewGrant().
		InDatabase(URI("/projects/p-1/providers/Aruba.Database/dbaas/d-1/databases/db-1"))
	gid := "g-1"
	g.id = &gid

	want := "/projects/p-1/providers/Aruba.Database/dbaas/d-1/databases/db-1/grants/g-1"
	if g.URI() != want {
		t.Errorf("URI() = %q, want %q", g.URI(), want)
	}
}

func TestGrant_URI_MissingProjectID(t *testing.T) {
	g := &Grant{}
	g.databaseID = "db-1"
	g.dbaasID = "d-1"
	gid := "g-1"
	g.id = &gid
	if g.URI() != "" {
		t.Errorf("URI() should be empty when projectID missing, got %q", g.URI())
	}
}

func TestGrant_URI_MissingDBaaSID(t *testing.T) {
	g := &Grant{}
	g.projectID = "p-1"
	g.databaseID = "db-1"
	gid := "g-1"
	g.id = &gid
	if g.URI() != "" {
		t.Errorf("URI() should be empty when dbaasID missing, got %q", g.URI())
	}
}

func TestGrant_URI_MissingDatabaseID(t *testing.T) {
	g := &Grant{}
	g.projectID = "p-1"
	g.dbaasID = "d-1"
	gid := "g-1"
	g.id = &gid
	if g.URI() != "" {
		t.Errorf("URI() should be empty when databaseID missing, got %q", g.URI())
	}
}

func TestGrant_URI_MissingGrantID(t *testing.T) {
	g := &Grant{}
	g.projectID = "p-1"
	g.dbaasID = "d-1"
	g.databaseID = "db-1"
	if g.URI() != "" {
		t.Errorf("URI() should be empty when grantID missing, got %q", g.URI())
	}
}

// --------------------------------------------------------------------------
// toRequest
// --------------------------------------------------------------------------

func TestGrant_ToRequest(t *testing.T) {
	g := NewGrant().ForUser("alice").OfRole("READ_WRITE")
	req := g.toRequest()
	if req.User.Username != "alice" {
		t.Errorf("toRequest().User.Username = %q", req.User.Username)
	}
	if req.Role.Name != "READ_WRITE" {
		t.Errorf("toRequest().Role.Name = %q", req.Role.Name)
	}
}

func TestGrant_ToRequest_Empty(t *testing.T) {
	g := &Grant{}
	req := g.toRequest()
	if req.User.Username != "" {
		t.Errorf("toRequest().User.Username should be empty, got %q", req.User.Username)
	}
	if req.Role.Name != "" {
		t.Errorf("toRequest().Role.Name should be empty, got %q", req.Role.Name)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func grantTestResponse(userName, roleName, dbName string) *types.GrantResponse {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	creator := "admin@example.com"
	return &types.GrantResponse{
		User:         types.GrantUserCommon{Username: userName},
		Role:         types.GrantRoleCommon{Name: roleName},
		Database:     types.GrantDatabaseResponse{Name: dbName},
		CreationDate: &ts,
		CreatedBy:    &creator,
	}
}

func TestGrant_FromResponseHydration(t *testing.T) {
	resp := grantTestResponse("alice", "READ_WRITE", "my-db")
	g := &Grant{}
	g.fromResponse(resp)

	if g.Username() != "alice" {
		t.Errorf("UserName() = %q", g.Username())
	}
	if g.RoleName() != "READ_WRITE" {
		t.Errorf("RoleName() = %q", g.RoleName())
	}
	if g.DatabaseName() != "my-db" {
		t.Errorf("DatabaseName() = %q", g.DatabaseName())
	}
	if g.CreatedAt().IsZero() {
		t.Error("CreatedAt() should be non-zero")
	}
	if g.CreatedBy() != "admin@example.com" {
		t.Errorf("CreatedBy() = %q", g.CreatedBy())
	}
	if g.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestGrant_FromResponse_NilSafe(t *testing.T) {
	g := &Grant{}
	g.fromResponse(nil) // must not panic
	if g.ID() != "" || g.Username() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}

	g2 := &Grant{}
	g2.fromResponse(&types.GrantResponse{}) // empty response
	if g2.Username() != "" || g2.RoleName() != "" {
		t.Error("empty response should yield zero accessor values")
	}
}

func TestGrant_FromResponse_DoesNotTouchID(t *testing.T) {
	g := &Grant{}
	gid := "g-existing"
	g.id = &gid
	g.fromResponse(grantTestResponse("alice", "READ_WRITE", "my-db"))
	// fromResponse must NOT clear or overwrite g.id — GrantResponse has no id field.
	if g.ID() != "g-existing" {
		t.Errorf("fromResponse clobbered g.id; ID() = %q, want %q", g.ID(), "g-existing")
	}
}

// --------------------------------------------------------------------------
// grantIDsFromRef helper
// --------------------------------------------------------------------------

func TestGrantIDsFromRef_TypedRef(t *testing.T) {
	g := NewGrant().
		InDatabase(URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1"))
	gid := "g-1"
	g.id = &gid

	pid, did, dbid, gotGID, err := grantIDsFromRef(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid != "p" || did != "d-1" || dbid != "db-1" || gotGID != "g-1" {
		t.Errorf("grantIDsFromRef typed = (%q, %q, %q, %q)", pid, did, dbid, gotGID)
	}
}

func TestGrantIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1/grants/g-1")
	pid, did, dbid, gid, err := grantIDsFromRef(ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid != "p" || did != "d-1" || dbid != "db-1" || gid != "g-1" {
		t.Errorf("grantIDsFromRef URI = (%q, %q, %q, %q)", pid, did, dbid, gid)
	}
}

func TestGrantIDsFromRef_BadURI_MissingGrants(t *testing.T) {
	_, _, _, _, err := grantIDsFromRef(URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1"))
	if err == nil {
		t.Error("expected error for URI without /grants/<id>")
	}
}

func TestGrantIDsFromRef_BadURI_MissingDatabases(t *testing.T) {
	_, _, _, _, err := grantIDsFromRef(URI("/projects/p/providers/Aruba.Database/dbaas/d-1/grants/g-1"))
	if err == nil {
		t.Error("expected error for URI without /databases/<id>")
	}
}

func TestGrantIDsFromRef_BadURI_MissingDBaaS(t *testing.T) {
	_, _, _, _, err := grantIDsFromRef(URI("/projects/p/providers/Aruba.Database/databases/db-1/grants/g-1"))
	if err == nil {
		t.Error("expected error for URI without /dbaas/<id>")
	}
}

func TestGrantIDsFromRef_BadURI_MissingProject(t *testing.T) {
	_, _, _, _, err := grantIDsFromRef(URI("/providers/Aruba.Database/dbaas/d-1/databases/db-1/grants/g-1"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

// --------------------------------------------------------------------------
// grantsClientAdapter — HTTP mock tests
// --------------------------------------------------------------------------

func buildGrantTestAdapter(t *testing.T, handler http.HandlerFunc) *grantsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newGrantsClientAdapter(testutil.NewClient(t, server.URL))
}

const grantSuccessBody = `{"user":{"username":"alice"},"role":{"name":"READ_WRITE"},"database":{"name":"db-1"},"createdBy":"admin@example.com"}`

// --------------------------------------------------------------------------
// WaitUntilGone — adapter-level test (Family-B path)
// --------------------------------------------------------------------------

func TestGrantsClientAdapter_WaitUntilGone(t *testing.T) {
	callCount := 0
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, grantSuccessBody)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "grant not found", 404))
		}
	})

	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1/grants/g-1")
	grant, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if err := grant.WaitUntilGone(context.Background(), fastOpts()...); err != nil {
		t.Fatalf("WaitUntilGone error: %v", err)
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 calls (1 Get + 1 refresh returning 404), got %d", callCount)
	}
}

func TestGrantsClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.GrantRequest
	callCount := 0
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "databases/db-1/grants") {
			t.Errorf("path %q should contain 'databases/db-1/grants'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, grantSuccessBody)
	})

	g := NewGrant().
		InDatabase(URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1")).
		ForUser("alice").
		OfRole("READ_WRITE")

	result, err := adapter.Create(context.Background(), g)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
	if result.Username() != "alice" {
		t.Errorf("UserName() = %q", result.Username())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.User.Username != "alice" {
		t.Errorf("wire body User.Username = %q", gotBody.User.Username)
	}
	if gotBody.Role.Name != "READ_WRITE" {
		t.Errorf("wire body Role.Name = %q", gotBody.Role.Name)
	}
	// Wire-schema gap: GrantResponse has no id field — ID() stays empty after Create.
	if result.ID() != "" {
		t.Errorf("ID() should be empty after Create (wire-schema gap), got %q", result.ID())
	}
}

func TestGrantsClientAdapter_Create_NoParent(t *testing.T) {
	callCount := 0
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
	})
	g := NewGrant().ForUser("alice").OfRole("READ_WRITE") // no IntoDatabase
	_, err := adapter.Create(context.Background(), g)
	if err == nil {
		t.Error("expected error when parent database is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestGrantsClientAdapter_Create_NoUsername(t *testing.T) {
	callCount := 0
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
	})
	g := NewGrant().
		InDatabase(URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1")).
		OfRole("READ_WRITE") // no WithUserName
	_, err := adapter.Create(context.Background(), g)
	if err == nil {
		t.Error("expected error when username is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestGrantsClientAdapter_Create_NoRoleName(t *testing.T) {
	callCount := 0
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
	})
	g := NewGrant().
		InDatabase(URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1")).
		ForUser("alice") // no WithRoleName
	_, err := adapter.Create(context.Background(), g)
	if err == nil {
		t.Error("expected error when role is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestGrantsClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, `{"title":"Unprocessable","detail":"invalid grant"}`)
	})

	g := NewGrant().
		InDatabase(URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1")).
		ForUser("alice").
		OfRole("READ_WRITE")
	_, err := adapter.Create(context.Background(), g)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

func TestGrantsClientAdapter_Update_Success(t *testing.T) {
	var gotBody types.GrantRequest
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "grants/g-789") {
			t.Errorf("path %q should contain 'grants/g-789'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, grantSuccessBody)
	})

	g := NewGrant().
		InDatabase(URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1")).
		ForUser("alice").
		OfRole("READ_WRITE")
	gid := "g-789"
	g.id = &gid // white-box: simulate prior Get populating the opaque ID

	result, err := adapter.Update(context.Background(), g)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if result.Username() != "alice" {
		t.Errorf("UserName() = %q", result.Username())
	}
	if gotBody.User.Username != "alice" {
		t.Errorf("wire body User.Username = %q", gotBody.User.Username)
	}
	if gotBody.Role.Name != "READ_WRITE" {
		t.Errorf("wire body Role.Name = %q", gotBody.Role.Name)
	}
}

func TestGrantsClientAdapter_Update_NoID(t *testing.T) {
	// Wire-schema constraint: Update requires an opaque grantID from a prior Get.
	callCount := 0
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
	})
	g := NewGrant().
		InDatabase(URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1")).
		ForUser("alice").
		OfRole("READ_WRITE") // all ancestor IDs set, but no g.id
	_, err := adapter.Update(context.Background(), g)
	if err == nil {
		t.Error("expected error when grant ID is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestGrantsClientAdapter_Update_NoIDs(t *testing.T) {
	callCount := 0
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
	})
	g := NewGrant() // bare — no IntoDatabase, no WithUserName, no WithRoleName
	_, err := adapter.Update(context.Background(), g)
	if err == nil {
		t.Error("expected error when all IDs missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestGrantsClientAdapter_Get_URIRef(t *testing.T) {
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if !containsSubstring(r.URL.Path, "grants/g-1") {
			t.Errorf("path %q should contain 'grants/g-1'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, grantSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1/grants/g-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "g-1" {
		t.Errorf("ID() = %q, want %q", result.ID(), "g-1")
	}
	if result.DatabaseID() != "db-1" {
		t.Errorf("DatabaseID() = %q", result.DatabaseID())
	}
	if result.DBaaSID() != "d-1" {
		t.Errorf("DBaaSID() = %q", result.DBaaSID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	if result.Username() != "alice" {
		t.Errorf("UserName() = %q", result.Username())
	}
}

func TestGrantsClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, grantSuccessBody)
	})

	g := NewGrant().
		InDatabase(URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1"))
	gid := "g-1"
	g.id = &gid // white-box: typed Ref with populated ID

	result, err := adapter.Get(context.Background(), g)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	if result.DatabaseID() != "db-1" {
		t.Errorf("DatabaseID() = %q", result.DatabaseID())
	}
}

func TestGrantsClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1/grants/g-1")
	if err := adapter.Delete(context.Background(), ref); err != nil {
		t.Errorf("Delete error: %v", err)
	}
}

func TestGrantsClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"title":"Not Found","detail":"grant not found"}`)
	})
	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1/grants/g-1")
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
// RawRequest (0% → covers the method)
// --------------------------------------------------------------------------

func TestGrant_RawRequest(t *testing.T) {
	g := NewGrant().ForUser("alice").OfRole("READ_WRITE")
	req := g.RawRequest()
	if req.User.Username != "alice" {
		t.Errorf("RawRequest().User.Username = %q", req.User.Username)
	}
	if req.Role.Name != "READ_WRITE" {
		t.Errorf("RawRequest().Role.Name = %q", req.Role.Name)
	}
	// Zero value — must not panic
	_ = NewGrant().RawRequest()
}

// --------------------------------------------------------------------------
// Zero-value accessors (Shape F — covers the nil-response branch)
// --------------------------------------------------------------------------

func TestGrant_Accessors_ZeroValue(t *testing.T) {
	g := &Grant{}
	if g.DatabaseName() != "" {
		t.Errorf("DatabaseName() zero = %q", g.DatabaseName())
	}
	if !g.CreatedAt().IsZero() {
		t.Error("CreatedAt() zero should be zero time")
	}
	if g.CreatedBy() != "" {
		t.Errorf("CreatedBy() zero = %q", g.CreatedBy())
	}
}

// --------------------------------------------------------------------------
// Create — extra guard paths
// --------------------------------------------------------------------------

func TestGrantsClientAdapter_Create_NoDBaaS(t *testing.T) {
	callCount := 0
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
	})
	// Has database in path but no DBaaS — bad URI produces no dbaasID
	g := NewGrant().
		InDatabase(URI("/projects/p/providers/Aruba.Database/databases/db-1")).
		ForUser("alice").
		OfRole("READ_WRITE")
	_, err := adapter.Create(context.Background(), g)
	if err == nil {
		t.Fatal("expected error when DBaaS is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

// --------------------------------------------------------------------------
// Update — extra guard paths
// --------------------------------------------------------------------------

func TestGrantsClientAdapter_Update_NoDatabase(t *testing.T) {
	callCount := 0
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
	})
	// Has grantID but no database
	g := &Grant{}
	gid := "g-1"
	g.id = &gid
	g.dbaasID = "d-1"
	g.projectID = "p"
	uname := "alice"
	g.username = &uname
	rname := "READ_WRITE"
	g.roleName = &rname
	_, err := adapter.Update(context.Background(), g)
	if err == nil {
		t.Fatal("expected error when databaseID is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestGrantsClientAdapter_Update_NoDBaaS(t *testing.T) {
	callCount := 0
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
	})
	// Has grantID and databaseID but no dbaasID
	g := &Grant{}
	gid := "g-1"
	g.id = &gid
	g.databaseID = "db-1"
	g.projectID = "p"
	uname := "alice"
	g.username = &uname
	rname := "READ_WRITE"
	g.roleName = &rname
	_, err := adapter.Update(context.Background(), g)
	if err == nil {
		t.Fatal("expected error when dbaasID is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestGrantsClientAdapter_Update_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
	})
	// Has grantID, databaseID, dbaasID but no projectID
	g := &Grant{}
	gid := "g-1"
	g.id = &gid
	g.databaseID = "db-1"
	g.dbaasID = "d-1"
	uname := "alice"
	g.username = &uname
	rname := "READ_WRITE"
	g.roleName = &rname
	_, err := adapter.Update(context.Background(), g)
	if err == nil {
		t.Fatal("expected error when projectID is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestGrantsClientAdapter_Update_NoUsername(t *testing.T) {
	callCount := 0
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
	})
	g := &Grant{}
	gid := "g-1"
	g.id = &gid
	g.databaseID = "db-1"
	g.dbaasID = "d-1"
	g.projectID = "p"
	rname := "READ_WRITE"
	g.roleName = &rname
	_, err := adapter.Update(context.Background(), g)
	if err == nil {
		t.Fatal("expected error when username is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestGrantsClientAdapter_Update_NoRoleName(t *testing.T) {
	callCount := 0
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
	})
	g := &Grant{}
	gid := "g-1"
	g.id = &gid
	g.databaseID = "db-1"
	g.dbaasID = "d-1"
	g.projectID = "p"
	uname := "alice"
	g.username = &uname
	_, err := adapter.Update(context.Background(), g)
	if err == nil {
		t.Fatal("expected error when roleName is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestGrantsClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `{"title":"Conflict","detail":"concurrent update"}`)
	})
	g := NewGrant().
		InDatabase(URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1")).
		ForUser("alice").
		OfRole("READ_WRITE")
	gid := "g-1"
	g.id = &gid
	_, err := adapter.Update(context.Background(), g)
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

func TestGrantsClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.Get(context.Background(), URI("/something/bad"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
}

func TestGrantsClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"title":"Not Found","detail":"grant not found"}`)
	})
	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1/grants/g-1")
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
// Create — broken client (network error from LL)
// --------------------------------------------------------------------------

func TestGrantsClientAdapter_Create_BrokenClient(t *testing.T) {
	adapter := &grantsClientAdapter{low: database.NewGrantsClientImpl(testutil.NewBrokenClient(t, "http://localhost:9"))}
	g := NewGrant().
		InDatabase(URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1")).
		ForUser("alice").
		OfRole("READ_WRITE")
	_, err := adapter.Create(context.Background(), g)
	if err == nil {
		t.Fatal("expected network error from broken client")
	}
}

// --------------------------------------------------------------------------
// Create — errMixin path
// --------------------------------------------------------------------------

func TestGrantsClientAdapter_Create_ErrMixin(t *testing.T) {
	callCount := 0
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
	})
	// IntoDatabase with bad URI sets errMixin
	g := NewGrant().InDatabase(URI("/garbage")).ForUser("alice").OfRole("READ_WRITE")
	_, err := adapter.Create(context.Background(), g)
	if err == nil {
		t.Fatal("expected error from errMixin")
	}
	if callCount != 0 {
		t.Errorf("expected 0 HTTP calls, got %d", callCount)
	}
}

// --------------------------------------------------------------------------
// Delete — bad ref and broken client
// --------------------------------------------------------------------------

func TestGrantsClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := adapter.Delete(context.Background(), URI("/something/bad"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
}

// --------------------------------------------------------------------------
// List — bad parent ref and non-2xx
// --------------------------------------------------------------------------

func TestGrantsClientAdapter_List_BadRef(t *testing.T) {
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/something/bad"))
	if err == nil {
		t.Fatal("expected error for bad parent Ref")
	}
}

func TestGrantsClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"title":"Forbidden","detail":"not allowed"}`)
	})
	parent := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1")
	_, err := adapter.List(context.Background(), parent)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

func TestGrantsClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildGrantTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if !containsSubstring(r.URL.Path, "grants") {
			t.Errorf("path %q should contain 'grants'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"values":[{"user":{"username":"alice"},"role":{"name":"READ_WRITE"},"database":{"name":"db-1"}},{"user":{"username":"bob"},"role":{"name":"READ_ONLY"},"database":{"name":"db-1"}}]}`)
	})

	parent := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/db-1")
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
		if item.DatabaseID() != "db-1" {
			t.Errorf("items[%d].DatabaseID() = %q", i, item.DatabaseID())
		}
		if item.DBaaSID() != "d-1" {
			t.Errorf("items[%d].DBaaSID() = %q", i, item.DBaaSID())
		}
		if item.ProjectID() != "p" {
			t.Errorf("items[%d].ProjectID() = %q", i, item.ProjectID())
		}
		// Wire-schema gap: GrantResponse has no id field — List items have empty ID().
		if item.ID() != "" {
			t.Errorf("items[%d].ID() should be empty (wire-schema gap), got %q", i, item.ID())
		}
	}
	if list.Total() != 2 {
		t.Errorf("Total() = %d", list.Total())
	}
}
