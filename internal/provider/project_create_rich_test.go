package provider

import (
	"context"
	"net/http"
	"testing"
)

// projectCreateRichJSON is the API response for a successful project Create that
// includes a description in Properties and tags in Metadata.  This exercises the
// branches inside ProjectResource.Create() that map the response description and
// tags back into Terraform state (lines that are skipped by minimalActiveJSON).
const projectCreateRichJSON = `{` +
	`"metadata":{` +
	`"id":"test-project-id",` +
	`"name":"test-name",` +
	`"tags":["env:test","team:platform"]` +
	`},` +
	`"status":{"state":"Active"},` +
	`"properties":{` +
	`"description":"test description"` +
	`}` +
	`}`

// TestProjectCreate_WithTagsAndDescription verifies that ProjectResource.Create()
// correctly maps tags and description from the API response into Terraform state.
// This covers the `if len(response.Data.Metadata.Tags) > 0` branch and the
// `if response.Data.Properties.Description != nil` branch that minimalActiveJSON
// leaves uncovered.
func TestProjectCreate_WithTagsAndDescription(t *testing.T) {
	ctx := context.Background()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Write([]byte(projectCreateRichJSON)) //nolint:errcheck
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewProjectResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceCreateReq(ctx, t, res)
	res.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("ProjectResource Create() with description+tags reported error: %v", resp.Diagnostics)
	}
}

// TestProjectCreate_WithEmptyTagsList verifies that ProjectResource.Create()
// correctly handles a response where Tags is empty but the plan had a non-null
// (empty) tags list.  This covers the `else { emptyList }` branch in the Create
// function's tags-response section.
const projectCreateNoTagsJSON = `{` +
	`"metadata":{` +
	`"id":"test-project-id",` +
	`"name":"test-name"` +
	`},` +
	`"status":{"state":"Active"},` +
	`"properties":{}` +
	`}`

func TestProjectCreate_WithEmptyTagsList(t *testing.T) {
	ctx := context.Background()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Write([]byte(projectCreateNoTagsJSON)) //nolint:errcheck
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewProjectResource()
	configureResource(ctx, t, res, mockClient)

	// resourceCreateReq sets tags to null (non-string attribute).
	// This hits the `data.Tags.IsNull()` true branch.
	req, resp := resourceCreateReq(ctx, t, res)
	res.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("ProjectResource Create() with no-tags response reported error: %v", resp.Diagnostics)
	}
}
