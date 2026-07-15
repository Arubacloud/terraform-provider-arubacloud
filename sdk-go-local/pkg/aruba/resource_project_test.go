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
// Compile-time Ref and withProjectID satisfaction
// --------------------------------------------------------------------------

var _ Ref = (*Project)(nil)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestProject_FluentSetters(t *testing.T) {
	p := NewProject().Named(
		"my-project").
		Tagged("a").
		Tagged("b").
		Tagged("a"). // dedupe
		DescribedAs("desc").
		AsDefault()

	if p.Name() != "my-project" {
		t.Errorf("Name() = %q", p.Name())
	}
	if tags := p.Tags(); len(tags) != 2 || tags[0] != "a" || tags[1] != "b" {
		t.Errorf("Tags() = %v", tags)
	}
	if p.Description() != "desc" {
		t.Errorf("Description() = %q", p.Description())
	}
	if !p.IsDefault() {
		t.Error("IsDefault() should be true")
	}

	p.Untagged("a")
	if tags := p.Tags(); len(tags) != 1 || tags[0] != "b" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	p.RetaggedAs("x", "y")
	if tags := p.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

func TestProject_Description_Unset(t *testing.T) {
	p := NewProject()
	if p.Description() != "" {
		t.Errorf("Description() of empty project = %q, want empty", p.Description())
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestProject_ToRequestRoundTrip(t *testing.T) {
	p := NewProject().Named(
		"proj").
		Tagged("t1").
		Tagged("t2").
		DescribedAs("d").
		AsDefault()

	req := p.toRequest()
	if req.Metadata.Name != "proj" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if len(req.Metadata.Tags) != 2 {
		t.Errorf("Metadata.Tags = %v", req.Metadata.Tags)
	}
	if req.Properties.Description == nil || *req.Properties.Description != "d" {
		t.Errorf("Properties.Description = %v", req.Properties.Description)
	}
	if !req.Properties.Default {
		t.Error("Properties.Default should be true")
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func projectTestResponse(id, name, uri string) *types.ProjectResponse {
	desc := "hello"
	return &types.ProjectResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:   &id,
			URI:  &uri,
			Name: &name,
			Tags: []string{"tag1"},
		},
		Properties: types.ProjectPropertiesResponse{
			Description: &desc,
			Default:     true,
		},
	}
}

func TestProject_FromResponseHydration(t *testing.T) {
	p := &Project{}
	resp := projectTestResponse("id-1", "n1", "/projects/id-1")
	p.fromResponse(resp)

	if p.ID() != "id-1" {
		t.Errorf("ID() = %q", p.ID())
	}
	if p.URI() != "/projects/id-1" {
		t.Errorf("URI() = %q", p.URI())
	}
	if p.Name() != "n1" {
		t.Errorf("Name() = %q", p.Name())
	}
	if tags := p.Tags(); len(tags) != 1 || tags[0] != "tag1" {
		t.Errorf("Tags() = %v", tags)
	}
	if p.Description() != "hello" {
		t.Errorf("Description() = %q", p.Description())
	}
	if !p.IsDefault() {
		t.Error("IsDefault() should be true")
	}
	if p.ProjectID() != "id-1" {
		t.Errorf("ProjectID() = %q (should equal own ID)", p.ProjectID())
	}
}

func TestProject_FromResponsePartial(t *testing.T) {
	// nil response is a no-op
	p := &Project{}
	p.fromResponse(nil)
	if p.ID() != "" || p.URI() != "" || p.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}

	// response with nil pointer fields — all accessors return zero values without panic
	p2 := &Project{}
	p2.fromResponse(&types.ProjectResponse{})
	if p2.ID() != "" || p2.URI() != "" || p2.Description() != "" {
		t.Error("empty response should yield zero accessor values")
	}
}

// --------------------------------------------------------------------------
// Ref + withProjectID interface satisfaction (runtime check)
// --------------------------------------------------------------------------

func TestProject_RefSatisfaction(t *testing.T) {
	p := &Project{}
	p.fromResponse(projectTestResponse("pid", "n", "/projects/pid"))

	// withProjectID path (typed assertion) should be preferred over URI parsing
	id, ok := extractID(p, func(r Ref) (string, bool) {
		if wp, ok := r.(withProjectID); ok {
			return wp.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || id != "pid" {
		t.Errorf("extractID via withProjectID = (%q, %v)", id, ok)
	}
}

// --------------------------------------------------------------------------
// projectIDFromRef
// --------------------------------------------------------------------------

func TestProjectIDFromRef_TypedRef(t *testing.T) {
	p := &Project{}
	p.fromResponse(projectTestResponse("abc", "n", "/projects/abc"))
	id, err := projectIDFromRef(p)
	if err != nil || id != "abc" {
		t.Errorf("projectIDFromRef typed = (%q, %v)", id, err)
	}
}

func TestProjectIDFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/xyz")
	id, err := projectIDFromRef(ref)
	if err != nil || id != "xyz" {
		t.Errorf("projectIDFromRef URI = (%q, %v)", id, err)
	}
}

func TestProjectIDFromRef_BadURI(t *testing.T) {
	ref := URI("/something/else")
	_, err := projectIDFromRef(ref)
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

// --------------------------------------------------------------------------
// Raw / RawJSON / RawYAML / RawRequest
// --------------------------------------------------------------------------

func TestProject_Raw_AfterHydrate(t *testing.T) {
	p := &Project{}
	resp := projectTestResponse("id-1", "n1", "/projects/id-1")
	p.fromResponse(resp)

	if p.Raw() == nil {
		t.Fatal("Raw() should be non-nil after fromResponse")
	}
	if p.Raw() != resp {
		t.Error("Raw() should return the stored response pointer")
	}
	if p.Raw().Metadata.ID == nil || *p.Raw().Metadata.ID != "id-1" {
		t.Errorf("Raw().Metadata.ID = %v", p.Raw().Metadata.ID)
	}
}

func TestProject_RawJSON_NilSafe(t *testing.T) {
	p := NewProject()
	if p.RawJSON() != nil {
		t.Error("RawJSON() before hydration should be nil")
	}

	p2 := &Project{}
	p2.fromResponse(projectTestResponse("id-2", "n2", "/projects/id-2"))
	b := p2.RawJSON()
	if len(b) == 0 {
		t.Error("RawJSON() after hydration should be non-empty")
	}
	if string(b[:1]) != "{" {
		t.Errorf("RawJSON() does not look like JSON: %s", string(b[:10]))
	}
}

func TestProject_RawRequest(t *testing.T) {
	p := NewProject().Named("myproj").DescribedAs("desc").AsDefault()
	req := p.RawRequest()
	if req.Metadata.Name != "myproj" {
		t.Errorf("RawRequest().Metadata.Name = %q", req.Metadata.Name)
	}
	if req.Properties.Description == nil || *req.Properties.Description != "desc" {
		t.Errorf("RawRequest().Properties.Description = %v", req.Properties.Description)
	}
	if !req.Properties.Default {
		t.Error("RawRequest().Properties.Default should be true")
	}
}

// --------------------------------------------------------------------------
// projectClientAdapter — CRUD integration tests (httptest-based)
// --------------------------------------------------------------------------

func buildTestAdapter(t *testing.T, handler http.HandlerFunc) *projectClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newProjectClientAdapter(testutil.NewClient(t, server.URL))
}

func TestProjectClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.ProjectRequest
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"metadata":{"id":"pid","name":"my-proj","uri":"/projects/pid"},"properties":{"default":false}}`)
	})

	p := NewProject().
		Named("my-proj")
	result, err := adapter.Create(context.Background(), p)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "pid" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-proj" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.URI() != "/projects/pid" {
		t.Errorf("URI() = %q", result.URI())
	}
	if result.Err() != nil {
		t.Errorf("Err() = %v", result.Err())
	}
	if gotBody.Metadata.Name != "my-proj" {
		t.Errorf("request Metadata.Name = %q", gotBody.Metadata.Name)
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
}

func TestProjectClientAdapter_Create_MetadataValidationError(t *testing.T) {
	// Server returns 201 but omits required "id" field — low-level Create wraps
	// *MetadataValidationError; adapter preserves it and returns non-nil *Project.
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"metadata":{"name":"p","uri":"/projects/x"},"properties":{}}`)
	})

	result, err := adapter.Create(context.Background(), NewProject().
		Named("p"))
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
	// HTTP envelope must be populated even on MetadataValidationError path
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d, want 201", result.StatusCode())
	}
}

func TestProjectClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "name is required", 422))
	})

	result, err := adapter.Create(context.Background(), NewProject())
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

func TestProjectClientAdapter_Create_SetterTimeError(t *testing.T) {
	// Simulate a setter-time error by manually injecting one
	p := NewProject()
	p.addErr(fmt.Errorf("setter error"))

	callCount := 0
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})
	_, err := adapter.Create(context.Background(), p)
	if err == nil {
		t.Fatal("expected setter-time error")
	}
	if callCount != 0 {
		t.Error("HTTP should not be called when setter errors present")
	}
}

func TestProjectClientAdapter_Get_URIRef(t *testing.T) {
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/pid" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"pid","name":"got","uri":"/projects/pid"},"properties":{}}`)
	})

	result, err := adapter.Get(context.Background(), URI("/projects/pid"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "pid" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "got" {
		t.Errorf("Name() = %q", result.Name())
	}
}

func TestProjectClientAdapter_Get_TypedRef(t *testing.T) {
	// Fetch returns a *Project; use it as Ref for a second Get to verify typed path
	var capturedPath string
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"p2","name":"n2","uri":"/projects/p2"},"properties":{}}`)
	})

	p := &Project{}
	p.fromResponse(projectTestResponse("p2", "n2", "/projects/p2"))

	result, err := adapter.Get(context.Background(), p)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if capturedPath != "/projects/p2" {
		t.Errorf("path = %q", capturedPath)
	}
	if result.ID() != "p2" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestProjectClientAdapter_Update_Success(t *testing.T) {
	var capturedBody types.ProjectRequest
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"pid","name":"renamed","uri":"/projects/pid"},"properties":{}}`)
	})

	p := &Project{}
	p.fromResponse(projectTestResponse("pid", "orig", "/projects/pid"))
	p.Named("renamed")

	result, err := adapter.Update(context.Background(), p)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if result.Name() != "renamed" {
		t.Errorf("Name() = %q", result.Name())
	}
	if capturedBody.Metadata.Name != "renamed" {
		t.Errorf("request Metadata.Name = %q", capturedBody.Metadata.Name)
	}
}

func TestProjectClientAdapter_Update_NoID(t *testing.T) {
	callCount := 0
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	_, err := adapter.Update(context.Background(), NewProject().
		Named("x"))
	if err == nil {
		t.Fatal("expected error when project has no ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when ID is missing")
	}
}

func TestProjectClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if err := adapter.Delete(context.Background(), URI("/projects/pid")); err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestProjectClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "project not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/missing"))
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

func TestProjectClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"/projects","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"p1","name":"n1","uri":"/projects/p1"},"properties":{}},`+
			`{"metadata":{"id":"p2","name":"n2","uri":"/projects/p2"},"properties":{}}`+
			`]}`)
	})

	list, err := adapter.List(context.Background())
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
	if items[0].ID() != "p1" || items[0].Name() != "n1" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[1].ID() != "p2" || items[1].Name() != "n2" {
		t.Errorf("items[1] = {%q, %q}", items[1].ID(), items[1].Name())
	}
}

func TestProjectClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Forbidden", "access denied", 403))
	})

	_, err := adapter.List(context.Background())
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

// --------------------------------------------------------------------------
// Get — bad ref and non-2xx
// --------------------------------------------------------------------------

func TestProjectClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := adapter.Get(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad ref")
	}
}

func TestProjectClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "project not found", 404))
	})

	result, err := adapter.Get(context.Background(), URI("/projects/missing"))
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
// Update — setter error, non-2xx
// --------------------------------------------------------------------------

func TestProjectClientAdapter_Update_SetterError(t *testing.T) {
	callCount := 0
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	p := &Project{}
	p.fromResponse(projectTestResponse("pid", "orig", "/projects/pid"))
	p.addErr(fmt.Errorf("setter error"))

	_, err := adapter.Update(context.Background(), p)
	if err == nil {
		t.Fatal("expected setter-time error")
	}
	if callCount != 0 {
		t.Error("HTTP should not be called when setter errors present")
	}
}

func TestProjectClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "name is required", 422))
	})

	p := &Project{}
	p.fromResponse(projectTestResponse("pid", "orig", "/projects/pid"))

	result, err := adapter.Update(context.Background(), p)
	if err == nil {
		t.Fatal("expected error on 422")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
	if result == nil {
		t.Fatal("result must be non-nil on non-2xx")
	}
}

// --------------------------------------------------------------------------
// Delete — bad ref
// --------------------------------------------------------------------------

func TestProjectClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad ref")
	}
}
