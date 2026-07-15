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
// Compile-time Ref satisfaction
// --------------------------------------------------------------------------

var _ Ref = (*Snapshot)(nil)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestSnapshot_FluentSetters(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-1", "my-proj", "/projects/p-1"))

	snap := NewSnapshot().
		InProject(proj).
		Named("my-snap").
		Tagged("backup").
		Tagged("storage").
		Tagged("backup"). // dedupe
		InRegion(RegionITBGBergamo).
		BilledBy(BillingPeriodHour)

	if snap.Name() != "my-snap" {
		t.Errorf("Name() = %q", snap.Name())
	}
	if tags := snap.Tags(); len(tags) != 2 || tags[0] != "backup" || tags[1] != "storage" {
		t.Errorf("Tags() = %v", tags)
	}
	if snap.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", snap.Region())
	}
	if snap.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", snap.BillingPeriod())
	}
	if snap.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", snap.ProjectID())
	}
	if snap.Err() != nil {
		t.Errorf("Err() = %v", snap.Err())
	}

	snap.Untagged("backup")
	if tags := snap.Tags(); len(tags) != 1 || tags[0] != "storage" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	snap.RetaggedAs("x", "y")
	if tags := snap.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

// --------------------------------------------------------------------------
// IntoProject — typed / URI / bad Ref
// --------------------------------------------------------------------------

func TestSnapshot_IntoProject_TypedRef(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-42", "n", "/projects/p-42"))
	snap := NewSnapshot().InProject(proj)
	if snap.ProjectID() != "p-42" {
		t.Errorf("ProjectID() = %q", snap.ProjectID())
	}
	if snap.Err() != nil {
		t.Errorf("Err() = %v", snap.Err())
	}
}

func TestSnapshot_IntoProject_URIRef(t *testing.T) {
	snap := NewSnapshot().InProject(URI("/projects/p-uri"))
	if snap.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", snap.ProjectID())
	}
	if snap.Err() != nil {
		t.Errorf("Err() = %v", snap.Err())
	}
}

func TestSnapshot_IntoProject_BadRef(t *testing.T) {
	snap := NewSnapshot().InProject(URI("/garbage"))
	if snap.Err() == nil {
		t.Error("expected Err() != nil for unresolvable Ref")
	}
}

// --------------------------------------------------------------------------
// FromVolume
// --------------------------------------------------------------------------

func TestSnapshot_FromVolume_URIRef(t *testing.T) {
	volURI := "/projects/p/providers/Aruba.Storage/blockStorages/bs-1"
	snap := NewSnapshot().FromVolume(URI(volURI))
	if snap.VolumeURI() != volURI {
		t.Errorf("VolumeURI() = %q", snap.VolumeURI())
	}
	if snap.Err() != nil {
		t.Errorf("Err() = %v", snap.Err())
	}
}

func TestSnapshot_FromVolume_TypedRef(t *testing.T) {
	// Simulate a typed BlockStorage with a known URI.
	bs := &BlockStorage{}
	bs.fromResponse(blockStorageTestResponse("bs-42", "n", "/projects/p/providers/Aruba.Storage/blockStorages/bs-42", "p"))

	snap := NewSnapshot().FromVolume(bs)
	if snap.VolumeURI() != bs.URI() {
		t.Errorf("VolumeURI() = %q, want %q", snap.VolumeURI(), bs.URI())
	}
	if snap.Err() != nil {
		t.Errorf("Err() = %v", snap.Err())
	}
}

func TestSnapshot_FromVolume_EmptyURI(t *testing.T) {
	snap := NewSnapshot().FromVolume(URI(""))
	if snap.Err() == nil {
		t.Error("expected Err() != nil for empty volume URI")
	}
	if snap.VolumeURI() != "" {
		t.Errorf("VolumeURI() should remain empty, got %q", snap.VolumeURI())
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestSnapshot_ToRequestRoundTrip(t *testing.T) {
	volURI := "/projects/p/providers/Aruba.Storage/blockStorages/bs-1"
	snap := NewSnapshot().Named(
		"snap-rt").
		Tagged("t1").Tagged("t2").
		InRegion(RegionITBGBergamo).
		BilledBy(BillingPeriodHour).
		FromVolume(URI(volURI))

	req := snap.RawRequest()

	if req.Metadata.Name != "snap-rt" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if len(req.Metadata.Tags) != 2 {
		t.Errorf("Metadata.Tags = %v", req.Metadata.Tags)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Location.Value = %q", req.Metadata.Location.Value)
	}
	if req.Properties.BillingPeriod == nil || *req.Properties.BillingPeriod != BillingPeriodHour {
		t.Errorf("BillingPeriod = %v", req.Properties.BillingPeriod)
	}
	if req.Properties.Volume.URI != volURI {
		t.Errorf("Volume.URI = %q", req.Properties.Volume.URI)
	}
}

func TestSnapshot_ToRequest_UnsetOptionals_AreNilOrEmpty(t *testing.T) {
	snap := NewSnapshot().
		Named("bare")
	req := snap.RawRequest()

	if req.Properties.BillingPeriod == nil || *req.Properties.BillingPeriod != BillingPeriodHour {
		t.Errorf("BillingPeriod should default to Hour, got %v", req.Properties.BillingPeriod)
	}
	if req.Properties.Volume.URI != "" {
		t.Errorf("Volume.URI should be empty, got %q", req.Properties.Volume.URI)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func snapshotTestResponse(id, name, uri, projectID string) *types.SnapshotResponse {
	loc := &types.LocationResponse{Value: RegionITBGBergamo}
	state := types.State("Active")
	billingPeriod := BillingPeriodHour
	sizeGB := int32(20)
	boot := true
	zone := ZoneITBG1
	volURI := "/projects/p/providers/Aruba.Storage/blockStorages/bs-1"
	return &types.SnapshotResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			URI:              &uri,
			Name:             &name,
			Tags:             []string{"tag1"},
			LocationResponse: loc,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{
				ID: projectID,
			},
		},
		Properties: types.SnapshotPropertiesResponse{
			SizeGB:        &sizeGB,
			BillingPeriod: &billingPeriod,
			Type:          BlockStorageTypeStandard,
			Zone:          zone,
			Bootable:      &boot,
			Volume: &types.VolumeInfoResponse{
				URI: &volURI,
			},
		},
		Status: types.ResourceStatusResponse{
			State: &state,
			DisableStatusInfoResponse: &types.DisableStatusInfoResponse{
				IsDisabled: false,
			},
		},
	}
}

func TestSnapshot_FromResponseHydration(t *testing.T) {
	snap := &Snapshot{}
	resp := snapshotTestResponse("snap-1", "my-snap", "/projects/p1/providers/Aruba.Storage/snapshots/snap-1", "p1")
	snap.fromResponse(resp)

	if snap.ID() != "snap-1" {
		t.Errorf("ID() = %q", snap.ID())
	}
	if snap.URI() != "/projects/p1/providers/Aruba.Storage/snapshots/snap-1" {
		t.Errorf("URI() = %q", snap.URI())
	}
	if snap.SnapshotID() != "snap-1" {
		t.Errorf("SnapshotID() = %q", snap.SnapshotID())
	}
	if snap.Name() != "my-snap" {
		t.Errorf("Name() = %q", snap.Name())
	}
	if tags := snap.Tags(); len(tags) != 1 || tags[0] != "tag1" {
		t.Errorf("Tags() = %v", tags)
	}
	if snap.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", snap.Region())
	}
	if snap.State() != "Active" {
		t.Errorf("State() = %q", snap.State())
	}
	if snap.IsDisabled() {
		t.Error("IsDisabled() should be false")
	}
	if snap.SizeGB() != 20 {
		t.Errorf("Size() = %d", snap.SizeGB())
	}
	if snap.Type() != BlockStorageTypeStandard {
		t.Errorf("Type() = %q", snap.Type())
	}
	if snap.Zone() != ZoneITBG1 {
		t.Errorf("Zone() = %q", snap.Zone())
	}
	if snap.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", snap.BillingPeriod())
	}
	if !snap.IsBootable() {
		t.Error("IsBootable() should be true")
	}
	if snap.VolumeURI() != "/projects/p/providers/Aruba.Storage/blockStorages/bs-1" {
		t.Errorf("VolumeURI() = %q", snap.VolumeURI())
	}
	if snap.ProjectID() != "p1" {
		t.Errorf("ProjectID() = %q", snap.ProjectID())
	}
	if snap.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestSnapshot_FromResponsePartial(t *testing.T) {
	// nil response is a no-op
	snap := &Snapshot{}
	snap.fromResponse(nil)
	if snap.ID() != "" || snap.URI() != "" || snap.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}

	// empty response — accessors must not panic; zero values expected
	snap2 := &Snapshot{}
	snap2.fromResponse(&types.SnapshotResponse{})
	if snap2.ID() != "" || snap2.URI() != "" || snap2.State() != "" || snap2.BillingPeriod() != "" || snap2.VolumeURI() != "" {
		t.Error("empty response should yield zero accessor values")
	}
}

func TestSnapshot_FromResponseURIBackfill(t *testing.T) {
	id := "snap-99"
	uri := "/projects/p-uri/providers/Aruba.Storage/snapshots/snap-99"
	state := types.State("")
	resp := &types.SnapshotResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  &id,
			URI: &uri,
		},
		Status: types.ResourceStatusResponse{State: &state},
	}
	snap := &Snapshot{}
	snap.fromResponse(resp)

	// ProjectMetadataResponse is nil → should backfill from URI.
	if snap.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() via URI backfill = %q", snap.ProjectID())
	}
}

func TestSnapshot_FromResponse_VolumeInfo_NilURI(t *testing.T) {
	id := "snap-1"
	uri := "/projects/p/providers/Aruba.Storage/snapshots/snap-1"
	state := types.State("")
	resp := &types.SnapshotResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  &id,
			URI: &uri,
		},
		Properties: types.SnapshotPropertiesResponse{
			Volume: &types.VolumeInfoResponse{
				URI: nil, // URI is nil
			},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}
	snap := &Snapshot{}
	snap.fromResponse(resp)

	if snap.VolumeURI() != "" {
		t.Errorf("VolumeURI() should be empty when VolumeInfo.URI is nil, got %q", snap.VolumeURI())
	}
}

// --------------------------------------------------------------------------
// Ref + ancestor ID satisfaction (runtime)
// --------------------------------------------------------------------------

func TestSnapshot_RefSatisfaction(t *testing.T) {
	snap := &Snapshot{}
	snap.fromResponse(snapshotTestResponse("snap-99", "n", "/projects/p99/providers/Aruba.Storage/snapshots/snap-99", "p99"))

	// withSnapshotID typed path
	sid, ok := extractID(snap, func(r Ref) (string, bool) {
		if w, ok := r.(withSnapshotID); ok {
			return w.SnapshotID(), true
		}
		return "", false
	}, "snapshots")
	if !ok || sid != "snap-99" {
		t.Errorf("extractID via withSnapshotID = (%q, %v)", sid, ok)
	}

	// withProjectID typed path
	pid, ok := extractID(snap, func(r Ref) (string, bool) {
		if w, ok := r.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || pid != "p99" {
		t.Errorf("extractID via withProjectID = (%q, %v)", pid, ok)
	}
}

// --------------------------------------------------------------------------
// snapshotIDsFromRef helper
// --------------------------------------------------------------------------

func TestSnapshotIDsFromRef_TypedRef(t *testing.T) {
	snap := &Snapshot{}
	snap.fromResponse(snapshotTestResponse("sid", "n", "/projects/p/providers/Aruba.Storage/snapshots/sid", "p"))
	pid, sid, err := snapshotIDsFromRef(snap)
	if err != nil || pid != "p" || sid != "sid" {
		t.Errorf("snapshotIDsFromRef typed = (%q, %q, %v)", pid, sid, err)
	}
}

func TestSnapshotIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Storage/snapshots/snap-1")
	pid, sid, err := snapshotIDsFromRef(ref)
	if err != nil || pid != "p" || sid != "snap-1" {
		t.Errorf("snapshotIDsFromRef URI = (%q, %q, %v)", pid, sid, err)
	}
}

func TestSnapshotIDsFromRef_BadURI_MissingSnapshot(t *testing.T) {
	_, _, err := snapshotIDsFromRef(URI("/projects/p/providers/Aruba.Storage/something/else"))
	if err == nil {
		t.Error("expected error for URI without /snapshots/<id>")
	}
}

func TestSnapshotIDsFromRef_BadURI_MissingProject(t *testing.T) {
	_, _, err := snapshotIDsFromRef(URI("/providers/Aruba.Storage/snapshots/snap-1"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

func TestSnapshotIDsFromRef_BadURI_MissingAll(t *testing.T) {
	_, _, err := snapshotIDsFromRef(URI("/something/else"))
	if err == nil {
		t.Error("expected error for completely unrelated URI")
	}
}

// --------------------------------------------------------------------------
// snapshotsClientAdapter — fake low-level client for body tests
// --------------------------------------------------------------------------

// fakeSnapshotLowLevel is a hand-rolled implementation of snapshotLowLevelClient.
type fakeSnapshotLowLevel struct {
	createFunc func(ctx context.Context, projectID string, body types.SnapshotRequest, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error)
	getFunc    func(ctx context.Context, projectID, snapshotID string, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error)
	updateFunc func(ctx context.Context, projectID, snapshotID string, body types.SnapshotRequest, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error)
	deleteFunc func(ctx context.Context, projectID, snapshotID string, params *types.RequestParameters) (*types.Response[any], error)
	listFunc   func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.SnapshotListResponse], error)
}

func (f *fakeSnapshotLowLevel) Create(ctx context.Context, projectID string, body types.SnapshotRequest, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
	return f.createFunc(ctx, projectID, body, params)
}
func (f *fakeSnapshotLowLevel) Get(ctx context.Context, projectID, snapshotID string, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
	return f.getFunc(ctx, projectID, snapshotID, params)
}
func (f *fakeSnapshotLowLevel) Update(ctx context.Context, projectID, snapshotID string, body types.SnapshotRequest, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
	return f.updateFunc(ctx, projectID, snapshotID, body, params)
}
func (f *fakeSnapshotLowLevel) Delete(ctx context.Context, projectID, snapshotID string, params *types.RequestParameters) (*types.Response[any], error) {
	return f.deleteFunc(ctx, projectID, snapshotID, params)
}
func (f *fakeSnapshotLowLevel) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.SnapshotListResponse], error) {
	return f.listFunc(ctx, projectID, params)
}

// --------------------------------------------------------------------------
// snapshotsClientAdapter — HTTP mock tests (no Volume.URI to skip wait-for-active)
// --------------------------------------------------------------------------

func buildSnapshotsTestAdapter(t *testing.T, handler http.HandlerFunc) *snapshotsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newSnapshotsClientAdapter(testutil.NewClient(t, server.URL))
}

const snapshotSuccessBody = `{` +
	`"metadata":{"id":"snap-1","name":"my-snap","uri":"/projects/p/providers/Aruba.Storage/snapshots/snap-1","project":{"id":"p"}},` +
	`"properties":{"sizeGb":20,"type":"Standard","billingPeriod":"Hour","dataCenter":"ITBG-1"},` +
	`"status":{"state":"Active"}}`

func TestSnapshotsClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.SnapshotRequest
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "snapshots") {
			t.Errorf("path %q should contain 'snapshots'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, snapshotSuccessBody)
	})

	// No OfVolume → Volume.URI == "" → no BlockStorage dependency.
	snap := NewSnapshot().
		InProject(URI("/projects/p")).
		Named("my-snap").
		InRegion(RegionITBGBergamo).
		BilledBy(BillingPeriodHour)

	result, err := adapter.Create(context.Background(), snap)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "snap-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-snap" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Metadata.Name != "my-snap" {
		t.Errorf("request Name = %q", gotBody.Metadata.Name)
	}
	if gotBody.Properties.BillingPeriod == nil || *gotBody.Properties.BillingPeriod != BillingPeriodHour {
		t.Errorf("request BillingPeriod = %v", gotBody.Properties.BillingPeriod)
	}
}

func TestSnapshotsClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	_, err := adapter.Create(context.Background(), NewSnapshot().
		Named("x"))
	if err == nil {
		t.Fatal("expected error when Snapshot has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestSnapshotsClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" field — triggers MetadataValidationError
		fmt.Fprint(w, `{"metadata":{"name":"snap","uri":"/projects/p/providers/Aruba.Storage/snapshots/x"},"properties":{},"status":{}}`)
	})

	snap := NewSnapshot().InProject(URI("/projects/p")).
		Named("snap")
	result, err := adapter.Create(context.Background(), snap)
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
}

func TestSnapshotsClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "name is required", 422))
	})

	snap := NewSnapshot().InProject(URI("/projects/p"))
	result, err := adapter.Create(context.Background(), snap)
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

// TestSnapshotsClientAdapter_Create_WithVolume uses the fake low-level client
// to assert the Volume.URI is wired correctly in the request body.
func TestSnapshotsClientAdapter_Create_WithVolume(t *testing.T) {
	volURI := "/projects/p/providers/Aruba.Storage/blockStorages/bs-1"
	var capturedBody types.SnapshotRequest

	snapResp := snapshotTestResponse("snap-1", "my-snap", "/projects/p/providers/Aruba.Storage/snapshots/snap-1", "p")
	resp := &types.Response[types.SnapshotResponse]{
		StatusCode: http.StatusCreated,
		Data:       snapResp,
	}

	fake := &fakeSnapshotLowLevel{
		createFunc: func(_ context.Context, _ string, body types.SnapshotRequest, _ *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
			capturedBody = body
			return resp, nil
		},
	}
	adapter := &snapshotsClientAdapter{low: fake}

	snap := NewSnapshot().
		InProject(URI("/projects/p")).
		Named("my-snap").
		FromVolume(URI(volURI))

	result, err := adapter.Create(context.Background(), snap)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if capturedBody.Properties.Volume.URI != volURI {
		t.Errorf("Volume.URI in request = %q", capturedBody.Properties.Volume.URI)
	}
	if result.ID() != "snap-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestSnapshotsClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, snapshotSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Storage/snapshots/snap-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "snap-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	if !containsSubstring(capturedPath, "snapshots") {
		t.Errorf("path %q should contain 'snapshots'", capturedPath)
	}
}

func TestSnapshotsClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, snapshotSuccessBody)
	})

	existing := &Snapshot{}
	existing.fromResponse(snapshotTestResponse("snap-1", "n", "/projects/p/providers/Aruba.Storage/snapshots/snap-1", "p"))

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "snap-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestSnapshotsClientAdapter_Update_Success(t *testing.T) {
	var capturedBody types.SnapshotRequest
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"snap-1","name":"my-snap","uri":"/projects/p/providers/Aruba.Storage/snapshots/snap-1","project":{"id":"p"}},"properties":{"billingPeriod":"Hour"},"status":{}}`)
	})

	snap := &Snapshot{}
	snap.fromResponse(snapshotTestResponse("snap-1", "my-snap", "/projects/p/providers/Aruba.Storage/snapshots/snap-1", "p"))
	snap.BilledBy(BillingPeriodHour)

	result, err := adapter.Update(context.Background(), snap)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if result.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", result.BillingPeriod())
	}
	if capturedBody.Properties.BillingPeriod == nil || *capturedBody.Properties.BillingPeriod != BillingPeriodHour {
		t.Errorf("request BillingPeriod = %v", capturedBody.Properties.BillingPeriod)
	}
}

func TestSnapshotsClientAdapter_Update_NoID(t *testing.T) {
	callCount := 0
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	snap := NewSnapshot().InProject(URI("/projects/p")).
		Named("x")
	_, err := adapter.Update(context.Background(), snap)
	if err == nil {
		t.Fatal("expected error when Snapshot has no ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when ID is missing")
	}
}

func TestSnapshotsClientAdapter_Update_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	snap := &Snapshot{}
	snap.fromResponse(&types.SnapshotResponse{
		Metadata: types.ResourceMetadataResponse{
			ID: strPtr("snap-1"),
		},
		Status: types.ResourceStatusResponse{},
	})

	_, err := adapter.Update(context.Background(), snap)
	if err == nil {
		t.Fatal("expected error when Snapshot has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestSnapshotsClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Storage/snapshots/snap-1"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestSnapshotsClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "snapshot not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Storage/snapshots/missing"))
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

// --------------------------------------------------------------------------
// Zero-value accessor tests (Shape F — 66.7% accessors)
// --------------------------------------------------------------------------

func TestSnapshot_Accessors_ZeroValue(t *testing.T) {
	snap := NewSnapshot()
	if snap.SizeGB() != 0 {
		t.Errorf("Size() zero value = %d, want 0", snap.SizeGB())
	}
	if snap.Type() != "" {
		t.Errorf("Type() zero value = %q, want \"\"", snap.Type())
	}
	if snap.IsBootable() != false {
		t.Error("IsBootable() zero value should be false")
	}
}

func TestNewSnapshotsClientAdapter_Nil(t *testing.T) {
	adapter := newSnapshotsClientAdapter(nil)
	if adapter == nil {
		t.Fatal("newSnapshotsClientAdapter(nil) returned nil")
	}
}

func TestSnapshotsClientAdapter_Update_Err(t *testing.T) {
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	snap := NewSnapshot().FromVolume(URI("")) // seeds an error
	_, err := adapter.Update(context.Background(), snap)
	if err == nil {
		t.Fatal("expected error when Snapshot has a pre-existing Err()")
	}
}

// --------------------------------------------------------------------------
// Additional adapter coverage tests
// --------------------------------------------------------------------------

func TestSnapshotsClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := adapter.Get(context.Background(), URI("/something/unrelated"))
	if err == nil {
		t.Fatal("expected error for unresolvable Ref")
	}
}

func TestSnapshotsClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "snapshot not found", 404))
	})

	ref := URI("/projects/p/providers/Aruba.Storage/snapshots/missing")
	_, err := adapter.Get(context.Background(), ref)
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

func TestSnapshotsClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "invalid billing period", 422))
	})

	snap := &Snapshot{}
	snap.fromResponse(snapshotTestResponse("snap-1", "my-snap", "/projects/p/providers/Aruba.Storage/snapshots/snap-1", "p"))
	snap.BilledBy(BillingPeriodHour)

	_, err := adapter.Update(context.Background(), snap)
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
}

func TestSnapshotsClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/something/unrelated"))
	if err == nil {
		t.Fatal("expected error for unresolvable Ref")
	}
}

func TestSnapshotsClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Forbidden", "not authorized", 403))
	})

	_, err := adapter.List(context.Background(), URI("/projects/p"))
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

func TestSnapshotsClientAdapter_List_BadRef(t *testing.T) {
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := adapter.List(context.Background(), URI("/something/unrelated"))
	if err == nil {
		t.Fatal("expected error for unresolvable project Ref")
	}
}

func TestSnapshotsClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildSnapshotsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"snap-1","name":"n1","uri":"/projects/p/providers/Aruba.Storage/snapshots/snap-1","project":{"id":"p"}},"properties":{"sizeGb":10,"type":"Standard","billingPeriod":"Hour","dataCenter":"ITBG-1"},"status":{}},`+
			`{"metadata":{"id":"snap-2","name":"n2","uri":"/projects/p/providers/Aruba.Storage/snapshots/snap-2","project":{"id":"p"}},"properties":{"sizeGb":20,"type":"Performance","billingPeriod":"Month","dataCenter":"ITBG-2"},"status":{}}`+
			`]}`)
	})

	list, err := adapter.List(context.Background(), URI("/projects/p"))
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
	if items[0].ID() != "snap-1" || items[0].Name() != "n1" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[0].SizeGB() != 10 {
		t.Errorf("items[0].SizeGB() = %d", items[0].SizeGB())
	}
	if items[1].ID() != "snap-2" || items[1].Type() != BlockStorageTypePerformance {
		t.Errorf("items[1] ID=%q Type=%q", items[1].ID(), items[1].Type())
	}
	if items[0].ProjectID() != "p" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
}

func TestSnapshot_FromResponse_SetsStatus(t *testing.T) {
	s := &Snapshot{}
	state := types.State("Active")
	s.fromResponse(&types.SnapshotResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if s.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", s.State())
	}
}

func TestSnapshotsClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, snapshotSuccessBody)
	})
	adapter := newSnapshotsClientAdapter(testutil.NewClient(t, server.URL))
	snap, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Storage/snapshots/snap-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&snap.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned Snapshot")
	}
}
