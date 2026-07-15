package aruba

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// errMixin
// --------------------------------------------------------------------------

func TestErrMixin_Empty(t *testing.T) {
	var m errMixin
	if m.Err() != nil {
		t.Errorf("expected nil error, got %v", m.Err())
	}
}

func TestErrMixin_SingleError(t *testing.T) {
	var m errMixin
	sentinel := errors.New("oops")
	m.addErr(sentinel)
	if !errors.Is(m.Err(), sentinel) {
		t.Errorf("Err() does not wrap sentinel: %v", m.Err())
	}
}

func TestErrMixin_MultipleErrors(t *testing.T) {
	var m errMixin
	e1 := errors.New("first")
	e2 := errors.New("second")
	m.addErr(e1)
	m.addErr(e2)
	err := m.Err()
	if !errors.Is(err, e1) {
		t.Errorf("Err() does not wrap e1: %v", err)
	}
	if !errors.Is(err, e2) {
		t.Errorf("Err() does not wrap e2: %v", err)
	}
}

func TestErrMixin_AddNilIsNoop(t *testing.T) {
	var m errMixin
	m.addErr(nil)
	if m.Err() != nil {
		t.Errorf("expected nil after adding nil error, got %v", m.Err())
	}
}

// --------------------------------------------------------------------------
// metadataMixin
// --------------------------------------------------------------------------

func TestMetadataMixin(t *testing.T) {
	var m metadataMixin
	m.named("hello")
	if m.Name() != "hello" {
		t.Errorf("Name() = %q", m.Name())
	}

	m.addTag("a")
	m.addTag("b")
	m.addTag("a") // duplicate — should not appear twice
	if got := m.Tags(); len(got) != 2 {
		t.Errorf("Tags() length = %d, want 2; tags=%v", len(got), got)
	}

	m.removeTag("a")
	tags := m.Tags()
	if len(tags) != 1 || tags[0] != "b" {
		t.Errorf("after Untagged(a): %v", tags)
	}

	m.removeTag("nonexistent") // no-op
	if len(m.Tags()) != 1 {
		t.Errorf("RemoveTag of missing tag changed slice")
	}

	m.replaceTags("x", "y", "z")
	if got := m.Tags(); len(got) != 3 {
		t.Errorf("after ReplaceTags: %v", got)
	}

	req := m.toMetadata()
	if req.Name != "hello" {
		t.Errorf("toMetadata().Name = %q", req.Name)
	}
	if len(req.Tags) != 3 {
		t.Errorf("toMetadata().Tags = %v", req.Tags)
	}
}

// --------------------------------------------------------------------------
// regionalMixin
// --------------------------------------------------------------------------

func TestRegionalMixin(t *testing.T) {
	var m regionalMixin
	m.inRegion("eu-west")
	if m.Region() != "eu-west" {
		t.Errorf("Region() = %q", m.Region())
	}
	m.inRegion("us-east")
	if m.Region() != "us-east" {
		t.Errorf("Region() after second inRegion = %q", m.Region())
	}
	loc := m.toLocation()
	if loc.Value != "us-east" {
		t.Errorf("toLocation().Value = %q", loc.Value)
	}
}

// --------------------------------------------------------------------------
// responseMetadataMixin
// --------------------------------------------------------------------------

func TestResponseMetadataMixin_SetMeta(t *testing.T) {
	var m responseMetadataMixin
	id := "abc"
	m.setMeta(&types.ResourceMetadataResponse{ID: &id})
	if m.ID() != "abc" {
		t.Errorf("ID() after setMeta = %q", m.ID())
	}
}

func TestResponseMetadataMixin_Nil(t *testing.T) {
	var m responseMetadataMixin
	if m.ID() != "" {
		t.Errorf("ID() on nil meta = %q", m.ID())
	}
	if m.RespURI() != "" {
		t.Errorf("RespURI() on nil meta = %q", m.RespURI())
	}
	if m.Project() != "" {
		t.Errorf("Project() on nil meta = %q", m.Project())
	}
	if !m.CreatedAt().IsZero() {
		t.Errorf("CreatedAt() should be zero, got %v", m.CreatedAt())
	}
	if !m.UpdatedAt().IsZero() {
		t.Errorf("UpdatedAt() should be zero, got %v", m.UpdatedAt())
	}
	if m.Version() != "" {
		t.Errorf("Version() on nil meta = %q", m.Version())
	}
	if m.CreatedBy() != "" {
		t.Errorf("CreatedBy() on nil meta = %q", m.CreatedBy())
	}
	if m.UpdatedBy() != "" {
		t.Errorf("UpdatedBy() on nil meta = %q", m.UpdatedBy())
	}
	if m.CreatedUser() != "" {
		t.Errorf("CreatedUser() on nil meta = %q", m.CreatedUser())
	}
	if m.UpdatedUser() != "" {
		t.Errorf("UpdatedUser() on nil meta = %q", m.UpdatedUser())
	}
}

func TestResponseMetadataMixin_Populated(t *testing.T) {
	id := "res-123"
	uri := "/projects/p/vpcs/v"
	ver := "1"
	proj := "proj-1"
	createdBy := "aru-297647"
	updatedBy := "aru-111111"
	createdUser := "alice"
	updatedUser := "bob"
	now := time.Now().UTC().Truncate(time.Second)
	later := now.Add(time.Hour)

	m := responseMetadataMixin{
		meta: &types.ResourceMetadataResponse{
			ID:      &id,
			URI:     &uri,
			Version: &ver,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{
				ID: proj,
			},
			CreationDate: &now,
			UpdateDate:   &later,
			CreatedBy:    &createdBy,
			UpdatedBy:    &updatedBy,
			CreatedUser:  &createdUser,
			UpdatedUser:  &updatedUser,
		},
	}

	if m.ID() != id {
		t.Errorf("ID() = %q, want %q", m.ID(), id)
	}
	if m.RespURI() != uri {
		t.Errorf("RespURI() = %q, want %q", m.RespURI(), uri)
	}
	if m.Project() != proj {
		t.Errorf("Project() = %q, want %q", m.Project(), proj)
	}
	if m.Version() != ver {
		t.Errorf("Version() = %q, want %q", m.Version(), ver)
	}
	if !m.CreatedAt().Equal(now) {
		t.Errorf("CreatedAt() = %v, want %v", m.CreatedAt(), now)
	}
	if !m.UpdatedAt().Equal(later) {
		t.Errorf("UpdatedAt() = %v, want %v", m.UpdatedAt(), later)
	}
	if m.CreatedBy() != createdBy {
		t.Errorf("CreatedBy() = %q, want %q", m.CreatedBy(), createdBy)
	}
	if m.UpdatedBy() != updatedBy {
		t.Errorf("UpdatedBy() = %q, want %q", m.UpdatedBy(), updatedBy)
	}
	if m.CreatedUser() != createdUser {
		t.Errorf("CreatedUser() = %q, want %q", m.CreatedUser(), createdUser)
	}
	if m.UpdatedUser() != updatedUser {
		t.Errorf("UpdatedUser() = %q, want %q", m.UpdatedUser(), updatedUser)
	}
}

// --------------------------------------------------------------------------
// linkedMixin
// --------------------------------------------------------------------------

func TestLinkedMixin(t *testing.T) {
	var m linkedMixin
	m.setLinked([]types.LinkedResourceCommon{{URI: "/foo", StrictCorrelation: true}})
	got := m.LinkedResources()
	if len(got) != 1 || got[0].URI != "/foo" {
		t.Errorf("LinkedResources() = %v", got)
	}
}

// --------------------------------------------------------------------------
// httpEnvelopeMixin
// --------------------------------------------------------------------------

func TestHTTPEnvelopeMixin(t *testing.T) {
	var m httpEnvelopeMixin

	title := "Not Found"
	status := int32(404)
	resp := &types.Response[struct{}]{
		StatusCode:   404,
		Headers:      http.Header{"X-Trace": []string{"abc"}},
		RawBody:      []byte(`{"title":"Not Found"}`),
		HTTPResponse: &http.Response{StatusCode: 404},
		Error: &types.ErrorResponse{
			Title:  &title,
			Status: &status,
		},
	}
	populateHTTPEnvelope(&m, resp)

	if m.StatusCode() != 404 {
		t.Errorf("StatusCode() = %d", m.StatusCode())
	}
	if m.Headers().Get("X-Trace") != "abc" {
		t.Errorf("Headers() = %v", m.Headers())
	}
	httpResp, raw := m.RawHTTP()
	if httpResp == nil || httpResp.StatusCode != 404 {
		t.Errorf("RawHTTP() http.Response = %v", httpResp)
	}
	if string(raw) != `{"title":"Not Found"}` {
		t.Errorf("RawHTTP() raw = %q", raw)
	}
	if m.RawError() == nil || *m.RawError().Title != "Not Found" {
		t.Errorf("RawError() = %v", m.RawError())
	}
}

func TestHTTPEnvelopeMixin_NilResponse(t *testing.T) {
	var m httpEnvelopeMixin
	populateHTTPEnvelope[struct{}](&m, nil)
	if m.StatusCode() != 0 {
		t.Errorf("StatusCode() after nil response = %d", m.StatusCode())
	}
}
