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
	_ Ref     = (*Database)(nil)
	_ Wrapper = (*Database)(nil)
)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestDatabase_FluentSetters(t *testing.T) {
	db := NewDatabase().
		InDBaaS(URI("/projects/p-1/providers/Aruba.Database/dbaas/d-1")).
		Named("mydb")

	if db.Name() != "mydb" {
		t.Errorf("Name() = %q", db.Name())
	}
	if db.ID() != "mydb" {
		t.Errorf("ID() = %q", db.ID())
	}
	if db.DatabaseID() != "mydb" {
		t.Errorf("DatabaseID() = %q", db.DatabaseID())
	}
	if db.DBaaSID() != "d-1" {
		t.Errorf("DBaaSID() = %q", db.DBaaSID())
	}
	if db.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", db.ProjectID())
	}
	if db.Err() != nil {
		t.Errorf("Err() = %v", db.Err())
	}
}

// --------------------------------------------------------------------------
// IntoDBaaS — typed / URI / bad Ref
// --------------------------------------------------------------------------

func TestDatabase_IntoDBaaS_TypedRef(t *testing.T) {
	dbaas := &DBaaS{}
	dbaas.fromResponse(dbaasTestResponse("d-1", "my-dbaas", "/projects/p-1/providers/Aruba.Database/dbaas/d-1"))

	db := NewDatabase().InDBaaS(dbaas)
	if db.DBaaSID() != "d-1" {
		t.Errorf("DBaaSID() = %q", db.DBaaSID())
	}
	if db.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", db.ProjectID())
	}
	if db.Err() != nil {
		t.Errorf("Err() = %v", db.Err())
	}
}

func TestDatabase_IntoDBaaS_URIRef(t *testing.T) {
	db := NewDatabase().InDBaaS(URI("/projects/p-uri/providers/Aruba.Database/dbaas/d-uri"))
	if db.DBaaSID() != "d-uri" {
		t.Errorf("DBaaSID() = %q", db.DBaaSID())
	}
	if db.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", db.ProjectID())
	}
	if db.Err() != nil {
		t.Errorf("Err() = %v", db.Err())
	}
}

func TestDatabase_IntoDBaaS_BadRef(t *testing.T) {
	db := NewDatabase().InDBaaS(URI("/something/garbage"))
	if db.Err() == nil {
		t.Error("expected Err() to be set for unresolvable parent")
	}
}

// --------------------------------------------------------------------------
// URI construction
// --------------------------------------------------------------------------

func TestDatabase_URI_Constructed(t *testing.T) {
	db := NewDatabase().
		InDBaaS(URI("/projects/p-1/providers/Aruba.Database/dbaas/d-1")).
		Named("mydb")

	want := "/projects/p-1/providers/Aruba.Database/dbaas/d-1/databases/mydb"
	if db.URI() != want {
		t.Errorf("URI() = %q, want %q", db.URI(), want)
	}
}

func TestDatabase_URI_MissingProjectID(t *testing.T) {
	db := &Database{}
	db.dbaasID = "d-1"
	name := "mydb"
	db.name = &name
	if db.URI() != "" {
		t.Errorf("URI() should be empty when projectID missing, got %q", db.URI())
	}
}

func TestDatabase_URI_MissingDBaaSID(t *testing.T) {
	db := &Database{}
	db.projectID = "p-1"
	name := "mydb"
	db.name = &name
	if db.URI() != "" {
		t.Errorf("URI() should be empty when dbaasID missing, got %q", db.URI())
	}
}

func TestDatabase_URI_MissingName(t *testing.T) {
	db := &Database{}
	db.projectID = "p-1"
	db.dbaasID = "d-1"
	if db.URI() != "" {
		t.Errorf("URI() should be empty when name missing, got %q", db.URI())
	}
}

// --------------------------------------------------------------------------
// toRequest
// --------------------------------------------------------------------------

func TestDatabase_ToRequest(t *testing.T) {
	db := NewDatabase().
		Named("mydb")
	req := db.toRequest()
	if req.Name != "mydb" {
		t.Errorf("toRequest().Name = %q", req.Name)
	}
}

func TestDatabase_ToRequest_Empty(t *testing.T) {
	db := &Database{}
	req := db.toRequest()
	if req.Name != "" {
		t.Errorf("toRequest().Name should be empty, got %q", req.Name)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func databaseTestResponse(name string) *types.DatabaseResponse {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	creator := "admin@example.com"
	return &types.DatabaseResponse{
		Name:         name,
		CreationDate: &ts,
		CreatedBy:    &creator,
	}
}

func TestDatabase_FromResponseHydration(t *testing.T) {
	resp := databaseTestResponse("mydb")
	db := &Database{}
	db.fromResponse(resp)

	if db.Name() != "mydb" {
		t.Errorf("Name() = %q", db.Name())
	}
	if db.ID() != "mydb" {
		t.Errorf("ID() = %q", db.ID())
	}
	if db.DatabaseID() != "mydb" {
		t.Errorf("DatabaseID() = %q", db.DatabaseID())
	}
	if db.CreatedAt().IsZero() {
		t.Error("CreatedAt() should be non-zero")
	}
	if db.CreatedBy() != "admin@example.com" {
		t.Errorf("CreatedBy() = %q", db.CreatedBy())
	}
	if db.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestDatabase_FromResponse_NilSafe(t *testing.T) {
	db := &Database{}
	db.fromResponse(nil) // must not panic
	if db.ID() != "" || db.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}

	db2 := &Database{}
	db2.fromResponse(&types.DatabaseResponse{}) // empty response — Name is ""
	if db2.ID() != "" || db2.Name() != "" {
		t.Error("empty response should yield zero accessor values")
	}
}

// --------------------------------------------------------------------------
// databaseIDsFromRef helper
// --------------------------------------------------------------------------

func TestDatabaseIDsFromRef_TypedRef(t *testing.T) {
	db := NewDatabase().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		Named("mydb")

	pid, did, name, err := databaseIDsFromRef(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid != "p" || did != "d-1" || name != "mydb" {
		t.Errorf("databaseIDsFromRef typed = (%q, %q, %q)", pid, did, name)
	}
}

func TestDatabaseIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/mydb")
	pid, did, name, err := databaseIDsFromRef(ref)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid != "p" || did != "d-1" || name != "mydb" {
		t.Errorf("databaseIDsFromRef URI = (%q, %q, %q)", pid, did, name)
	}
}

func TestDatabaseIDsFromRef_BadURI_MissingDatabase(t *testing.T) {
	_, _, _, err := databaseIDsFromRef(URI("/projects/p/providers/Aruba.Database/dbaas/d-1"))
	if err == nil {
		t.Error("expected error for URI without /databases/<name>")
	}
}

func TestDatabaseIDsFromRef_BadURI_MissingDBaaS(t *testing.T) {
	_, _, _, err := databaseIDsFromRef(URI("/projects/p/providers/Aruba.Database/databases/mydb"))
	if err == nil {
		t.Error("expected error for URI without /dbaas/<id>")
	}
}

func TestDatabaseIDsFromRef_BadURI_MissingProject(t *testing.T) {
	_, _, _, err := databaseIDsFromRef(URI("/providers/Aruba.Database/dbaas/d-1/databases/mydb"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

// --------------------------------------------------------------------------
// databasesClientAdapter — HTTP mock tests
// --------------------------------------------------------------------------

func buildDatabaseTestAdapter(t *testing.T, handler http.HandlerFunc) *databasesClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newDatabasesClientAdapter(testutil.NewClient(t, server.URL))
}

const databaseSuccessBody = `{"name":"my-db","createdBy":"admin@example.com"}`

func TestDatabasesClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.DatabaseRequest
	callCount := 0
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "databases") {
			t.Errorf("path %q should contain 'databases'", r.URL.Path)
		}
		if !containsSubstring(r.URL.Path, "dbaas") {
			t.Errorf("path %q should contain 'dbaas'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, databaseSuccessBody)
	})

	db := NewDatabase().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		Named("my-db")

	result, err := adapter.Create(context.Background(), db)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
	if result.Name() != "my-db" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Name != "my-db" {
		t.Errorf("wire body Name = %q", gotBody.Name)
	}
}

func TestDatabasesClientAdapter_Create_NoParent(t *testing.T) {
	callCount := 0
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
	})
	db := NewDatabase().Named("my-db") // no IntoDBaaS
	_, err := adapter.Create(context.Background(), db)
	if err == nil {
		t.Error("expected error when parent DBaaS is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestDatabasesClientAdapter_Create_NoName(t *testing.T) {
	callCount := 0
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
	})
	db := NewDatabase().InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")) // no Named
	_, err := adapter.Create(context.Background(), db)
	if err == nil {
		t.Error("expected error when name is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestDatabasesClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, `{"title":"Unprocessable","detail":"invalid name"}`)
	})

	db := NewDatabase().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		Named("my-db")

	_, err := adapter.Create(context.Background(), db)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

func TestDatabasesClientAdapter_Update_Success(t *testing.T) {
	var gotBody types.DatabaseRequest
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "databases/my-db") {
			t.Errorf("path %q should contain 'databases/my-db'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, databaseSuccessBody)
	})

	db := NewDatabase().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		Named("my-db")

	result, err := adapter.Update(context.Background(), db)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if result.Name() != "my-db" {
		t.Errorf("Name() = %q", result.Name())
	}
	if gotBody.Name != "my-db" {
		t.Errorf("wire body Name = %q", gotBody.Name)
	}
}

func TestDatabasesClientAdapter_Update_NoIDs(t *testing.T) {
	callCount := 0
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
	})
	db := NewDatabase() // no IntoDBaaS, no Named
	_, err := adapter.Update(context.Background(), db)
	if err == nil {
		t.Error("expected error when IDs are missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestDatabasesClientAdapter_Get_URIRef(t *testing.T) {
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if !containsSubstring(r.URL.Path, "databases/my-db") {
			t.Errorf("path %q should contain 'databases/my-db'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, databaseSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/my-db")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.Name() != "my-db" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.DBaaSID() != "d-1" {
		t.Errorf("DBaaSID() = %q", result.DBaaSID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
}

func TestDatabasesClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, databaseSuccessBody)
	})

	db := NewDatabase().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		Named("my-db")

	result, err := adapter.Get(context.Background(), db)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.Name() != "my-db" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
}

func TestDatabasesClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/my-db")
	if err := adapter.Delete(context.Background(), ref); err != nil {
		t.Errorf("Delete error: %v", err)
	}
}

func TestDatabasesClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"title":"Not Found","detail":"database not found"}`)
	})
	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/my-db")
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

func TestDatabase_RawRequest(t *testing.T) {
	db := NewDatabase().
		Named("mydb")
	req := db.RawRequest()
	if req.Name != "mydb" {
		t.Errorf("RawRequest().Name = %q", req.Name)
	}
	// Also test zero value
	_ = NewDatabase().RawRequest() // must not panic
}

// --------------------------------------------------------------------------
// Zero-value accessors (Shape F — covers the nil-response branch)
// --------------------------------------------------------------------------

func TestDatabase_Accessors_ZeroValue(t *testing.T) {
	db := &Database{}
	if !db.CreatedAt().IsZero() {
		t.Error("CreatedAt() zero should be zero time")
	}
	if db.CreatedBy() != "" {
		t.Errorf("CreatedBy() zero = %q", db.CreatedBy())
	}
}

// --------------------------------------------------------------------------
// Update — no dbaasID path
// --------------------------------------------------------------------------

func TestDatabasesClientAdapter_Update_NoDBaaSID(t *testing.T) {
	callCount := 0
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
	})
	// Has name and projectID but no dbaasID
	db := &Database{}
	db.projectID = "p"
	name := "mydb"
	db.name = &name
	_, err := adapter.Update(context.Background(), db)
	if err == nil {
		t.Fatal("expected error when dbaasID is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

// --------------------------------------------------------------------------
// Update — extra guard paths
// --------------------------------------------------------------------------

func TestDatabasesClientAdapter_Update_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	// Has name and dbaasID but no projectID
	db := &Database{}
	db.dbaasID = "d-1"
	name := "mydb"
	db.name = &name
	_, err := adapter.Update(context.Background(), db)
	if err == nil {
		t.Fatal("expected error when projectID is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

func TestDatabasesClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `{"title":"Conflict","detail":"concurrent update"}`)
	})
	db := NewDatabase().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		Named("my-db")

	_, err := adapter.Update(context.Background(), db)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusConflict {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Get — bad Ref
// --------------------------------------------------------------------------

func TestDatabasesClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.Get(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
}

// --------------------------------------------------------------------------
// Get — non-2xx response
// --------------------------------------------------------------------------

func TestDatabasesClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"title":"Not Found","detail":"database not found"}`)
	})
	ref := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/my-db")
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

func TestDatabasesClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := adapter.Delete(context.Background(), URI("/something/bad"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
}

// --------------------------------------------------------------------------
// Create — no dbaasID path
// --------------------------------------------------------------------------

func TestDatabasesClientAdapter_Create_NoDBaaSID(t *testing.T) {
	callCount := 0
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
	})
	// Has projectID but no dbaasID
	db := &Database{}
	db.projectID = "p"
	name := "mydb"
	db.name = &name
	_, err := adapter.Create(context.Background(), db)
	if err == nil {
		t.Fatal("expected error when dbaasID is missing")
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls, got %d", callCount)
	}
}

// --------------------------------------------------------------------------
// Create — errMixin path
// --------------------------------------------------------------------------

func TestDatabasesClientAdapter_Create_ErrMixin(t *testing.T) {
	callCount := 0
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
	})
	// IntoDBaaS with bad URI sets errMixin
	db := NewDatabase().InDBaaS(URI("/garbage")).
		Named("mydb")
	_, err := adapter.Create(context.Background(), db)
	if err == nil {
		t.Fatal("expected error from errMixin")
	}
	if callCount != 0 {
		t.Errorf("expected 0 HTTP calls, got %d", callCount)
	}
}

// --------------------------------------------------------------------------
// Update — broken client
// --------------------------------------------------------------------------

func TestDatabasesClientAdapter_Update_BrokenClient(t *testing.T) {
	adapter := &databasesClientAdapter{low: database.NewDatabasesClientImpl(testutil.NewBrokenClient(t, "http://localhost:9"))}
	db := NewDatabase().
		InDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		Named("my-db")

	_, err := adapter.Update(context.Background(), db)
	if err == nil {
		t.Fatal("expected network error from broken client")
	}
}

// --------------------------------------------------------------------------
// List — bad parent ref and non-2xx
// --------------------------------------------------------------------------

func TestDatabasesClientAdapter_List_BadRef(t *testing.T) {
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/something/bad"))
	if err == nil {
		t.Fatal("expected error for bad parent Ref")
	}
}

func TestDatabasesClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestDatabasesClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildDatabaseTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if !containsSubstring(r.URL.Path, "databases") {
			t.Errorf("path %q should contain 'databases'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"values":[{"name":"db-a"},{"name":"db-b"}]}`)
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
	if items[0].Name() != "db-a" {
		t.Errorf("items[0].Name() = %q", items[0].Name())
	}
	if items[1].Name() != "db-b" {
		t.Errorf("items[1].Name() = %q", items[1].Name())
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
