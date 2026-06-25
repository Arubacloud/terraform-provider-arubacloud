package provider

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// createRichJSON is a create-response that includes all optional metadata
// fields returned by the API (id, uri, tags).  Using this handler covers the
// "URI non-nil" and "len(Tags) > 0" branches in Create() functions that are
// skipped when the handler returns minimalActiveJSON.
const createRichJSON = `{` +
	`"metadata":{` +
	`"id":"test-id",` +
	`"name":"test-name",` +
	`"uri":"/test/resource/test-id",` +
	`"tags":["env:test","team:platform"]` +
	`},"status":{"state":"Active"}}`

// createSuccessRichHandler returns createRichJSON for both POST and GET
// requests.  POST (the actual create) receives 201; GET receives 200.  This
// is used by TestResourceCreate_RichResponse to cover the URI / tags response
// branches in Create() for a batch of resources.
func createSuccessRichHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write([]byte(createRichJSON)) //nolint:errcheck
}

// createSuccessRichWithURIHandler is like createSuccessRichHandler but uses
// a response that contains metadata.uri (required by backup/restore Create()
// which perform a GET to resolve the source volume URI before the POST).
const createRichWithSourceURI = `{` +
	`"metadata":{` +
	`"id":"test-id",` +
	`"name":"test-name",` +
	`"uri":"/projects/p/providers/Aruba.Storage/blockStorages/test-vol-id",` +
	`"tags":["env:test"]` +
	`},"status":{"state":"Active"}}`

func createSuccessRichWithURIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write([]byte(createRichWithSourceURI)) //nolint:errcheck
}

// TestResourceCreate_RichResponse repeats TestResourceCreate_Success with a
// handler that returns uri and tags in the response body.  This covers the
// "metadata.URI non-nil → data.Uri = ..." and "len(Tags) > 0 → data.Tags = ..."
// branches that minimalActiveJSON leaves as uncovered "false" paths in the
// Create() functions below.
func TestResourceCreate_RichResponse(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	cases := []struct {
		name    string
		newR    func() resource.Resource
		handler http.HandlerFunc
	}{
		{"vpc", NewVPCResource, createSuccessRichHandler},
		{"subnet", NewSubnetResource, createSuccessRichHandler},
		{"securitygroup", NewSecurityGroupResource, createSuccessRichHandler},
		{"elasticip", NewElasticIPResource, createSuccessRichHandler},
		{"keypair", NewKeypairResource, createSuccessRichHandler},
		{"blockstorage", NewBlockStorageResource, createSuccessRichHandler},
		{"snapshot", NewSnapshotResource, createSuccessRichHandler},
		{"kms", NewKMSResource, createSuccessRichHandler},
		{"vpcpeering", NewVpcPeeringResource, createSuccessRichHandler},
		{"vpcpeeringroute", NewVpcPeeringRouteResource, createSuccessRichHandler},
		{"vpntunnel", NewVPNTunnelResource, createSuccessRichHandler},
		{"databasegrant", NewDatabaseGrantResource, createSuccessRichHandler},
		// backup / restore pre-fetch a source volume URI before the actual POST.
		{"backup", NewBackupResource, createSuccessRichWithURIHandler},
		{"restore", NewRestoreResource, createSuccessRichWithURIHandler},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, mockClient := newMockArubaClient(t, tc.handler)

			res := tc.newR()
			configureResource(ctx, t, res, mockClient)

			req, resp := resourceCreateReq(ctx, t, res)
			res.Create(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Errorf("%s: Create() with rich response reported error: %v", tc.name, resp.Diagnostics)
			}
		})
	}
}
