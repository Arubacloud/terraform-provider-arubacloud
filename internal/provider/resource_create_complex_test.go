package provider

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// uriActiveJSON extends minimalActiveJSON with a metadata.uri so that
// Create() methods that GET a source volume/resource and require its URI
// (backup, restore) can proceed past the "URI not found in response" guard.
const uriActiveJSON = `{` +
	`"metadata":{` +
	`"id":"test-id","name":"test-name",` +
	`"uri":"/projects/p/providers/Aruba.Storage/volumes/test-vol-id"` +
	`},"status":{"state":"Active"}` +
	`}`

// createWithURIHandler serves uriActiveJSON (which includes a metadata.uri)
// for GET requests (source-resource lookup) and 201 + uriActiveJSON for POST
// (actual create).
func createWithURIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	//nolint:errcheck
	w.Write([]byte(uriActiveJSON))
}

// TestResourceCreate_HappyComplex covers resources that were excluded from
// TestResourceCreate_Happy because they require a metadata.uri in the GET
// response before the actual POST (backup and restore do a volume lookup).
func TestResourceCreate_HappyComplex(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	resources := []struct {
		name string
		newR func() resource.Resource
	}{
		{"backup", NewBackupResource},
		{"restore", NewRestoreResource},
	}

	for _, tc := range resources {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, createWithURIHandler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceCreateReq(ctx, t, res)
			res.Create(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Create() reported error with URI response: %v",
					tc.name, resp.Diagnostics)
			}
		})
	}
}
