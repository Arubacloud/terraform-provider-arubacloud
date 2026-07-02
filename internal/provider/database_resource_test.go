package provider

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// databaseCreateSuccessJSON is a response for the Database Create endpoint
// that includes a top-level "name" field so that DatabaseResponse.Name is
// non-empty and the provider can set data.Id = "testdb".
// The WaitForResourceActive checker then calls Get("testdb") rather than
// Get("") which the SDK would reject.
const databaseCreateSuccessJSON = `{"name":"testdb","status":{"state":"Active"}}`

func databaseCreateSuccessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write([]byte(databaseCreateSuccessJSON)) //nolint:errcheck
}

// TestDatabaseCreate_Success verifies that database Create() succeeds when the
// API response includes the database name (used as the resource ID) so that
// the WaitForResourceActive poll can call Get with the right ID.
func TestDatabaseCreate_Success(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, databaseCreateSuccessHandler)

	res := NewDatabaseResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceCreateReq(ctx, t, res)
	res.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Database Create() reported error: %v", resp.Diagnostics)
	}
}

// TestDatabaseUpdate_Success verifies that database Update() succeeds when the
// API returns valid data.  The Update needs projectID and dbaasID from state,
// and the GET response needs the current database details.
func TestDatabaseUpdate_Success(t *testing.T) {
	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"testdb","status":{"state":"Active"}}`)) //nolint:errcheck
	})

	res := NewDatabaseResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceUpdateReq(ctx, t, res)
	res.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Database Update() reported error: %v", resp.Diagnostics)
	}
}
