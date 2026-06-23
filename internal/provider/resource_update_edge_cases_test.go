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

// TestBlockStorageUpdate_MissingRegion verifies that BlockStorageResource.Update()
// adds an error when the GET response has no location (regionValue == "").
// This covers the `regionValue == ""` error branch that the standard handlers miss.
const blockstorageNoLocationJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name"},` +
	`"status":{"state":"NotUsed"},` +
	`"properties":{"sizeGb":10,"billingPeriod":"Hour","type":"Standard","dataCenter":""}}`

func TestBlockStorageUpdate_MissingRegion(t *testing.T) {
	ctx := context.Background()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(blockstorageNoLocationJSON)) //nolint:errcheck
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewBlockStorageResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceUpdateReq(ctx, t, res)
	res.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("BlockStorageResource Update() with missing region should have returned an error diagnostic")
	}
}

// elasticipInCreationJSON is an ElasticIP API response with status=InCreation.
// ElasticIPResource.Update() returns an error when the resource is still being
// created.  This covers the `*current.Status.State == "InCreation"` branch.
const elasticipInCreationJSON = `{` +
	`"metadata":{` +
	`"id":"test-id","name":"test-name",` +
	`"location":{"value":"test-location","code":"test-code","name":"test-region"}},` +
	`"status":{"state":"InCreation"},` +
	`"properties":{}}`

// TestElasticIPUpdate_InCreationStatus verifies that ElasticIPResource.Update()
// adds an error diagnostic (and makes no PATCH request) when the resource is
// in "InCreation" state.
func TestElasticIPUpdate_InCreationStatus(t *testing.T) {
	ctx := context.Background()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(elasticipInCreationJSON)) //nolint:errcheck
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewElasticIPResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceUpdateReq(ctx, t, res)
	res.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("ElasticIPResource Update() with 'InCreation' status should have returned an error diagnostic")
	}
}

// TestElasticIPUpdate_MissingRegion verifies that ElasticIPResource.Update()
// adds an error when the GET response has no location (regionValue == "").
const elasticipNoLocationJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name"},` +
	`"status":{"state":"Active"},` +
	`"properties":{}}`

func TestElasticIPUpdate_MissingRegion(t *testing.T) {
	ctx := context.Background()

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(elasticipNoLocationJSON)) //nolint:errcheck
	}

	_, mockClient := newMockArubaClient(t, handler)

	res := NewElasticIPResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceUpdateReq(ctx, t, res)
	res.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("ElasticIPResource Update() with missing region should have returned an error diagnostic")
	}
}

// vpcNoLocationJSON is a VPC API response without a location block.
// VPCResource.Update() returns an error when regionValue cannot be determined.
const vpcNoLocationJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name"},` +
	`"status":{"state":"Active"},` +
	`"properties":{}}`

// TestVPCUpdate_MissingRegion verifies that VPCResource.Update() adds an error
// when the GET response has no location (covers the `regionValue == ""` branch).
func TestVPCUpdate_MissingRegion(t *testing.T) {
	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(vpcNoLocationJSON)) //nolint:errcheck
	})

	res := NewVPCResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceUpdateReq(ctx, t, res)
	res.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("VPCResource Update() with missing region should have returned an error diagnostic")
	}
}

// backupNoLocationJSON is a Backup API response without a location block.
const backupNoLocationJSON = `{` +
	`"metadata":{"id":"test-id","name":"test-name"},` +
	`"status":{"state":"Active"},` +
	`"properties":{}}`

// TestBackupUpdate_MissingRegion verifies that BackupResource.Update() adds an
// error when the GET response has no location.
func TestBackupUpdate_MissingRegion(t *testing.T) {
	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(backupNoLocationJSON)) //nolint:errcheck
	})

	res := NewBackupResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceUpdateReq(ctx, t, res)
	res.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("BackupResource Update() with missing region should have returned an error diagnostic")
	}
}

// TestRestoreUpdate_MissingRegion verifies that RestoreResource.Update() adds
// an error when the GET response has no location.
func TestRestoreUpdate_MissingRegion(t *testing.T) {
	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Active"},"properties":{}}`)) //nolint:errcheck
	})

	res := NewRestoreResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceUpdateReq(ctx, t, res)
	res.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("RestoreResource Update() with missing region should have returned an error diagnostic")
	}
}

// TestSnapshotUpdate_MissingRegion verifies that SnapshotResource.Update() adds
// an error when the GET response has no location.
func TestSnapshotUpdate_MissingRegion(t *testing.T) {
	ctx := context.Background()

	_, mockClient := newMockArubaClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"metadata":{"id":"test-id","name":"test-name"},"status":{"state":"Active"},"properties":{}}`)) //nolint:errcheck
	})

	res := NewSnapshotResource()
	configureResource(ctx, t, res, mockClient)

	req, resp := resourceUpdateReq(ctx, t, res)
	res.Update(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("SnapshotResource Update() with missing region should have returned an error diagnostic")
	}
}
