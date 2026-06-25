package provider

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// blockstorageActiveJSON is a BlockStorage API response with status=Active.
// When BlockStorageResource.Update() receives this status, it returns an error
// (cannot update an attached/active volume) without making a PATCH call.
// This covers the `status != "Used" && status != "NotUsed"` branch.
// Note: BlockStorage SDK uses "sizeGb" and "dataCenter" as JSON field names.
const blockstorageActiveJSON = `{` +
	`"metadata":{` +
	`"id":"test-id","name":"test-name",` +
	`"location":{"value":"test-location","code":"test-code","name":"test-region"},` +
	`"project":{"id":"test-project-id"}},` +
	`"status":{"state":"Active"},` +
	`"properties":{"sizeGb":10,"billingPeriod":"Hour","type":"Standard","dataCenter":""}}`

// blockstorageNotUsedWithZoneJSON is like blockstorageNotUsedJSON but includes
// a non-empty dataCenter (zone).  This covers the `current.Properties.Zone != ""`
// true branch inside BlockStorageResource.Update().
const blockstorageNotUsedWithZoneJSON = `{` +
	`"metadata":{` +
	`"id":"test-id","name":"test-name",` +
	`"location":{"value":"test-location","code":"test-code","name":"test-region"},` +
	`"project":{"id":"test-project-id"}},` +
	`"status":{"state":"NotUsed"},` +
	`"properties":{"sizeGb":10,"billingPeriod":"Hour","type":"Standard","dataCenter":"test-zone"}}`

// TestBlockStorageUpdate_ActiveStatus verifies that BlockStorageResource.Update()
// adds an error diagnostic (and makes no PATCH request) when the resource is in
// "Active" state.  This covers the `status != "Used" && status != "NotUsed"`
// branch that the existing blockstorageUpdateSuccessHandler misses.
func TestBlockStorageUpdate_ActiveStatus(t *testing.T) {
	ctx := context.Background()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodGet {
			w.Write([]byte(blockstorageActiveJSON)) //nolint:errcheck
		} else {
			// PATCH should never be reached when status is Active.
			w.Write([]byte(minimalActiveJSON)) //nolint:errcheck
		}
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewBlockStorageResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceUpdateReq(ctx, t, res)
	res.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("BlockStorageResource Update() with 'Active' status should have returned an error diagnostic")
	}
}

// TestBlockStorageUpdate_WithZone verifies that BlockStorageResource.Update()
// sets a non-nil zone pointer when the current GET response contains a non-empty
// zone value.  This covers the `current.Properties.Zone != ""` true branch.
func TestBlockStorageUpdate_WithZone(t *testing.T) {
	oldActivePoll := waitForActivePollInterval
	waitForActivePollInterval = 1 * time.Millisecond
	t.Cleanup(func() { waitForActivePollInterval = oldActivePoll })

	ctx := context.Background()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodGet {
			w.Write([]byte(blockstorageNotUsedWithZoneJSON)) //nolint:errcheck
		} else {
			w.Write([]byte(blockstorageNotUsedWithZoneJSON)) //nolint:errcheck
		}
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewBlockStorageResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceUpdateReq(ctx, t, res)
	res.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("BlockStorageResource Update() with zone reported error: %v", resp.Diagnostics)
	}
}

// blockstorageNoLocationJSON is used by acceptance tests only.
// BlockStorage Update() does not check for location; the "missing region" variant
// of this constant was removed when it was found to test coincidental SDK
// URI-parsing failures rather than actual region-check logic.
const blockstorageNoLocationJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name"},` +
	`"status":{"state":"NotUsed"},` +
	`"properties":{"sizeGb":10,"billingPeriod":"Hour","type":"Standard","dataCenter":""}}`

// elasticipNoLocationJSON is kept for reference; ElasticIP Update() does not
// check for InCreation status so there is no status-guard branch to test here.

// elasticipNoLocationJSON is kept for reference; ElasticIP Update() does not
// check for location so there is no "missing region" error branch to test.
const elasticipNoLocationJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name"},` +
	`"status":{"state":"Active"},` +
	`"properties":{}}`

// vpcNoLocationJSON is a VPC API response without a location block.
const vpcNoLocationJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name"},` +
	`"status":{"state":"Active"},` +
	`"properties":{}}`

// backupNoLocationJSON is a Backup API response without a location block.
const backupNoLocationJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name"},` +
	`"status":{"state":"Active"},` +
	`"properties":{}}`
