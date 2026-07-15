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

var _ Ref = (*BlockStorage)(nil)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestBlockStorage_FluentSetters(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-1", "my-proj", "/projects/p-1"))

	bs := NewBlockStorage().
		InProject(proj).
		Named("my-bs").
		Tagged("storage").
		Tagged("data").
		Tagged("storage"). // dedupe
		InRegion(RegionITBGBergamo).
		SizedGB(50).
		OfType(BlockStorageTypeStandard).
		BilledBy(BillingPeriodHour).
		FromImage(VolumeImageLU22001).
		AsBootable()

	if bs.Name() != "my-bs" {
		t.Errorf("Name() = %q", bs.Name())
	}
	if tags := bs.Tags(); len(tags) != 2 || tags[0] != "storage" || tags[1] != "data" {
		t.Errorf("Tags() = %v", tags)
	}
	if bs.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", bs.Region())
	}
	if bs.SizeGB() != 50 {
		t.Errorf("Size() = %d", bs.SizeGB())
	}
	if bs.Type() != BlockStorageTypeStandard {
		t.Errorf("Type() = %q", bs.Type())
	}
	if bs.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", bs.BillingPeriod())
	}
	if bs.Image() != VolumeImageLU22001 {
		t.Errorf("Image() = %q", bs.Image())
	}
	if !bs.IsBootable() {
		t.Error("IsBootable() should be true")
	}
	if bs.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", bs.ProjectID())
	}
	if bs.Err() != nil {
		t.Errorf("Err() = %v", bs.Err())
	}

	bs.Untagged("storage")
	if tags := bs.Tags(); len(tags) != 1 || tags[0] != "data" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	bs.RetaggedAs("x", "y")
	if tags := bs.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

// --------------------------------------------------------------------------
// IntoProject — typed / URI / bad Ref
// --------------------------------------------------------------------------

func TestBlockStorage_IntoProject_TypedRef(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-42", "n", "/projects/p-42"))
	bs := NewBlockStorage().InProject(proj)
	if bs.ProjectID() != "p-42" {
		t.Errorf("ProjectID() = %q", bs.ProjectID())
	}
	if bs.Err() != nil {
		t.Errorf("Err() = %v", bs.Err())
	}
}

func TestBlockStorage_IntoProject_URIRef(t *testing.T) {
	bs := NewBlockStorage().InProject(URI("/projects/p-uri"))
	if bs.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", bs.ProjectID())
	}
	if bs.Err() != nil {
		t.Errorf("Err() = %v", bs.Err())
	}
}

func TestBlockStorage_IntoProject_BadRef(t *testing.T) {
	bs := NewBlockStorage().InProject(URI("/garbage"))
	if bs.Err() == nil {
		t.Error("expected Err() != nil for unresolvable Ref")
	}
}

// --------------------------------------------------------------------------
// FromSnapshot
// --------------------------------------------------------------------------

func TestBlockStorage_FromSnapshot_URIRef(t *testing.T) {
	snapURI := "/projects/p/providers/Aruba.Storage/snapshots/snap-1"
	bs := NewBlockStorage().FromSnapshot(URI(snapURI))
	if bs.SnapshotURI() != snapURI {
		t.Errorf("SnapshotURI() = %q", bs.SnapshotURI())
	}
	if bs.Err() != nil {
		t.Errorf("Err() = %v", bs.Err())
	}
}

func TestBlockStorage_FromSnapshot_TypedRef(t *testing.T) {
	// Any Ref works — use a URI-backed one to simulate a typed parent.
	snapURI := "/projects/p/providers/Aruba.Storage/snapshots/snap-42"
	bs := NewBlockStorage().FromSnapshot(URI(snapURI))
	if bs.SnapshotURI() != snapURI {
		t.Errorf("SnapshotURI() = %q", bs.SnapshotURI())
	}
	if bs.Err() != nil {
		t.Errorf("Err() = %v", bs.Err())
	}
}

func TestBlockStorage_FromSnapshot_EmptyURI(t *testing.T) {
	bs := NewBlockStorage().FromSnapshot(URI(""))
	if bs.Err() == nil {
		t.Error("expected Err() != nil for empty snapshot URI")
	}
	if bs.SnapshotURI() != "" {
		t.Errorf("SnapshotURI() should remain empty, got %q", bs.SnapshotURI())
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestBlockStorage_ToRequestRoundTrip(t *testing.T) {
	snapURI := "/projects/p/providers/Aruba.Storage/snapshots/s1"
	bs := NewBlockStorage().Named(
		"bs-rt").
		Tagged("t1").Tagged("t2").
		InRegion(RegionITBGBergamo).
		SizedGB(30).
		OfType(BlockStorageTypePerformance).
		InZone(ZoneITBG1).
		BilledBy(BillingPeriodHour).
		AsBootable().
		FromImage(VolumeImageLU22001).
		FromSnapshot(URI(snapURI))

	req := bs.RawRequest()

	if req.Metadata.Name != "bs-rt" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if len(req.Metadata.Tags) != 2 {
		t.Errorf("Metadata.Tags = %v", req.Metadata.Tags)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Location.Value = %q", req.Metadata.Location.Value)
	}
	if req.Properties.SizeGB != 30 {
		t.Errorf("SizeGB = %d", req.Properties.SizeGB)
	}
	if req.Properties.Type != BlockStorageTypePerformance {
		t.Errorf("Type = %q", req.Properties.Type)
	}
	if req.Properties.Zone == nil || *req.Properties.Zone != ZoneITBG1 {
		t.Errorf("Zone = %v", req.Properties.Zone)
	}
	if req.Properties.BillingPeriod == nil || *req.Properties.BillingPeriod != BillingPeriodHour {
		t.Errorf("BillingPeriod = %v", req.Properties.BillingPeriod)
	}
	if req.Properties.Bootable == nil || !*req.Properties.Bootable {
		t.Error("Bootable should be true")
	}
	if req.Properties.Image == nil || *req.Properties.Image != VolumeImageLU22001 {
		t.Errorf("Image = %v", req.Properties.Image)
	}
	if req.Properties.Snapshot == nil || req.Properties.Snapshot.URI != snapURI {
		t.Errorf("Snapshot.URI = %v", req.Properties.Snapshot)
	}
}

func TestBlockStorage_ToRequest_UnsetOptionals_AreNilOrZero(t *testing.T) {
	bs := NewBlockStorage().
		Named("bare").SizedGB(10).OfType(BlockStorageTypeStandard)
	req := bs.RawRequest()

	// RawRequest delegates to toCreateRequest where bootable defaults to false.
	if req.Properties.Bootable == nil || *req.Properties.Bootable != false {
		t.Errorf("Bootable should default to false in Create request, got %v", req.Properties.Bootable)
	}
	if req.Properties.Image != nil {
		t.Errorf("Image should be nil, got %v", req.Properties.Image)
	}
	if req.Properties.Zone != nil {
		t.Errorf("Zone should be nil, got %v", req.Properties.Zone)
	}
	if req.Properties.Snapshot != nil {
		t.Errorf("Snapshot should be nil, got %v", req.Properties.Snapshot)
	}
	if req.Properties.BillingPeriod == nil || *req.Properties.BillingPeriod != BillingPeriodHour {
		t.Errorf("BillingPeriod should default to Hour, got %v", req.Properties.BillingPeriod)
	}
}

func TestBlockStorage_ToRequest_ZonalVsRegional(t *testing.T) {
	// Region only — Zone must be nil.
	bs1 := NewBlockStorage().InRegion(RegionITBGBergamo)
	req1 := bs1.RawRequest()
	if req1.Properties.Zone != nil {
		t.Errorf("Zone should be nil when only region set, got %v", req1.Properties.Zone)
	}
	if req1.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Location.Value = %q", req1.Metadata.Location.Value)
	}

	// Region + Zone — both set.
	bs2 := NewBlockStorage().InRegion(RegionITBGBergamo).InZone(ZoneITBG1)
	req2 := bs2.RawRequest()
	if req2.Properties.Zone == nil || *req2.Properties.Zone != ZoneITBG1 {
		t.Errorf("Zone = %v", req2.Properties.Zone)
	}
	if req2.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Location.Value = %q", req2.Metadata.Location.Value)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func blockStorageTestResponse(id, name, uri, projectID string) *types.BlockStorageResponse {
	loc := &types.LocationResponse{Value: RegionITBGBergamo}
	state := types.State("Active")
	img := VolumeImageLU22001
	boot := true
	zone := ZoneITBG1
	snap := &types.ReferenceResourceCommon{URI: "/projects/p/providers/Aruba.Storage/snapshots/s1"}
	return &types.BlockStorageResponse{
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
		Properties: types.BlockStoragePropertiesResponse{
			SizeGB:        20,
			Type:          BlockStorageTypeStandard,
			Zone:          zone,
			BillingPeriod: func() *BillingPeriod { v := BillingPeriodHour; return &v }(),
			Image:         &img,
			Bootable:      &boot,
			Snapshot:      snap,
			LinkedResources: []types.LinkedResourceCommon{
				{URI: "/projects/p/providers/Aruba.Compute/cloudservers/cs1", StrictCorrelation: true},
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

func TestBlockStorage_FromResponseHydration(t *testing.T) {
	bs := &BlockStorage{}
	resp := blockStorageTestResponse("bs-1", "my-bs", "/projects/p1/providers/Aruba.Storage/blockStorages/bs-1", "p1")
	bs.fromResponse(resp)

	if bs.ID() != "bs-1" {
		t.Errorf("ID() = %q", bs.ID())
	}
	if bs.URI() != "/projects/p1/providers/Aruba.Storage/blockStorages/bs-1" {
		t.Errorf("URI() = %q", bs.URI())
	}
	if bs.BlockStorageID() != "bs-1" {
		t.Errorf("BlockStorageID() = %q", bs.BlockStorageID())
	}
	if bs.Name() != "my-bs" {
		t.Errorf("Name() = %q", bs.Name())
	}
	if tags := bs.Tags(); len(tags) != 1 || tags[0] != "tag1" {
		t.Errorf("Tags() = %v", tags)
	}
	if bs.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", bs.Region())
	}
	if bs.State() != "Active" {
		t.Errorf("State() = %q", bs.State())
	}
	if bs.IsDisabled() {
		t.Error("IsDisabled() should be false")
	}
	if linked := bs.LinkedResources(); len(linked) != 1 {
		t.Errorf("LinkedResources() len = %d", len(linked))
	}
	if bs.SizeGB() != 20 {
		t.Errorf("Size() = %d", bs.SizeGB())
	}
	if bs.Type() != BlockStorageTypeStandard {
		t.Errorf("Type() = %q", bs.Type())
	}
	if bs.Zone() != ZoneITBG1 {
		t.Errorf("Zone() = %q", bs.Zone())
	}
	if bs.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", bs.BillingPeriod())
	}
	if bs.Image() != VolumeImageLU22001 {
		t.Errorf("Image() = %q", bs.Image())
	}
	if !bs.IsBootable() {
		t.Error("IsBootable() should be true")
	}
	if bs.SnapshotURI() != "/projects/p/providers/Aruba.Storage/snapshots/s1" {
		t.Errorf("SnapshotURI() = %q", bs.SnapshotURI())
	}
	if bs.ProjectID() != "p1" {
		t.Errorf("ProjectID() = %q", bs.ProjectID())
	}
	if bs.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestBlockStorage_FromResponsePartial(t *testing.T) {
	// nil response is a no-op
	bs := &BlockStorage{}
	bs.fromResponse(nil)
	if bs.ID() != "" || bs.URI() != "" || bs.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}

	// empty response — accessors must not panic; zero values expected
	bs2 := &BlockStorage{}
	bs2.fromResponse(&types.BlockStorageResponse{})
	if bs2.ID() != "" || bs2.URI() != "" || bs2.State() != "" || bs2.BillingPeriod() != "" || bs2.Image() != "" {
		t.Error("empty response should yield zero accessor values")
	}
}

func TestBlockStorage_FromResponseURIBackfill(t *testing.T) {
	id := "bs-99"
	uri := "/projects/p-uri/providers/Aruba.Storage/blockStorages/bs-99"
	state := types.State("")
	resp := &types.BlockStorageResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  &id,
			URI: &uri,
		},
		Status: types.ResourceStatusResponse{State: &state},
	}
	bs := &BlockStorage{}
	bs.fromResponse(resp)

	// ProjectMetadataResponse is nil → should backfill from URI.
	if bs.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() via URI backfill = %q", bs.ProjectID())
	}
}

// --------------------------------------------------------------------------
// Ref + ancestor ID satisfaction (runtime)
// --------------------------------------------------------------------------

func TestBlockStorage_RefSatisfaction(t *testing.T) {
	bs := &BlockStorage{}
	bs.fromResponse(blockStorageTestResponse("bs-99", "n", "/projects/p99/providers/Aruba.Storage/blockStorages/bs-99", "p99"))

	// withBlockStorageID typed path
	bid, ok := extractID(bs, func(r Ref) (string, bool) {
		if w, ok := r.(withBlockStorageID); ok {
			return w.BlockStorageID(), true
		}
		return "", false
	}, "blockStorages")
	if !ok || bid != "bs-99" {
		t.Errorf("extractID via withBlockStorageID = (%q, %v)", bid, ok)
	}

	// withProjectID typed path
	pid, ok := extractID(bs, func(r Ref) (string, bool) {
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
// blockStorageIDsFromRef helper
// --------------------------------------------------------------------------

func TestBlockStorageIDsFromRef_TypedRef(t *testing.T) {
	bs := &BlockStorage{}
	bs.fromResponse(blockStorageTestResponse("bid", "n", "/projects/p/providers/Aruba.Storage/blockStorages/bid", "p"))
	pid, bid, err := blockStorageIDsFromRef(bs)
	if err != nil || pid != "p" || bid != "bid" {
		t.Errorf("blockStorageIDsFromRef typed = (%q, %q, %v)", pid, bid, err)
	}
}

func TestBlockStorageIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Storage/blockStorages/bs-1")
	pid, bid, err := blockStorageIDsFromRef(ref)
	if err != nil || pid != "p" || bid != "bs-1" {
		t.Errorf("blockStorageIDsFromRef URI = (%q, %q, %v)", pid, bid, err)
	}
}

func TestBlockStorageIDsFromRef_BadURI_MissingBlockStorage(t *testing.T) {
	_, _, err := blockStorageIDsFromRef(URI("/projects/p/providers/Aruba.Storage/something/else"))
	if err == nil {
		t.Error("expected error for URI without /blockStorages/<id>")
	}
}

func TestBlockStorageIDsFromRef_BadURI_MissingProject(t *testing.T) {
	_, _, err := blockStorageIDsFromRef(URI("/providers/Aruba.Storage/blockStorages/bs-1"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

func TestBlockStorageIDsFromRef_BadURI_MissingAll(t *testing.T) {
	_, _, err := blockStorageIDsFromRef(URI("/something/else"))
	if err == nil {
		t.Error("expected error for completely unrelated URI")
	}
}

// --------------------------------------------------------------------------
// volumesClientAdapter — CRUD integration tests
// --------------------------------------------------------------------------

func buildVolumesTestAdapter(t *testing.T, handler http.HandlerFunc) *volumesClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newVolumesClientAdapter(testutil.NewClient(t, server.URL))
}

const blockStorageSuccessBody = `{` +
	`"metadata":{"id":"bs-1","name":"my-bs","uri":"/projects/p/providers/Aruba.Storage/blockStorages/bs-1","project":{"id":"p"}},` +
	`"properties":{"sizeGB":20,"type":"Standard","billingPeriod":"Hour","zone":"ITBG-1"},` +
	`"status":{"state":"Active"}}`

func TestVolumesClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.BlockStorageRequest
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "blockStorages") {
			t.Errorf("path %q should contain 'blockstorages'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, blockStorageSuccessBody)
	})

	bs := NewBlockStorage().
		InProject(URI("/projects/p")).
		Named("my-bs").
		InRegion(RegionITBGBergamo).
		SizedGB(20).
		OfType(BlockStorageTypeStandard).
		BilledBy(BillingPeriodHour)

	result, err := adapter.Create(context.Background(), bs)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "bs-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-bs" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Metadata.Name != "my-bs" {
		t.Errorf("request Name = %q", gotBody.Metadata.Name)
	}
	if gotBody.Properties.SizeGB != 20 {
		t.Errorf("request SizeGB = %d", gotBody.Properties.SizeGB)
	}
}

func TestVolumesClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	_, err := adapter.Create(context.Background(), NewBlockStorage().
		Named("x"))
	if err == nil {
		t.Fatal("expected error when BlockStorage has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestVolumesClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" field — triggers MetadataValidationError
		fmt.Fprint(w, `{"metadata":{"name":"bs","uri":"/projects/p/providers/Aruba.Storage/blockStorages/x"},"properties":{},"status":{}}`)
	})

	bs := NewBlockStorage().InProject(URI("/projects/p")).
		Named("bs")
	result, err := adapter.Create(context.Background(), bs)
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

func TestVolumesClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "name is required", 422))
	})

	bs := NewBlockStorage().InProject(URI("/projects/p"))
	result, err := adapter.Create(context.Background(), bs)
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

func TestVolumesClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, blockStorageSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Storage/blockStorages/bs-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "bs-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	if !containsSubstring(capturedPath, "blockStorages") {
		t.Errorf("path %q should contain 'blockstorages'", capturedPath)
	}
}

func TestVolumesClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, blockStorageSuccessBody)
	})

	existing := &BlockStorage{}
	existing.fromResponse(blockStorageTestResponse("bs-1", "n", "/projects/p/providers/Aruba.Storage/blockStorages/bs-1", "p"))

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "bs-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestVolumesClientAdapter_Update_Success(t *testing.T) {
	var capturedBody types.BlockStorageRequest
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"bs-1","name":"my-bs","uri":"/projects/p/providers/Aruba.Storage/blockStorages/bs-1","project":{"id":"p"}},"properties":{"sizeGB":40,"type":"Standard","billingPeriod":"Hour"},"status":{}}`)
	})

	bs := &BlockStorage{}
	bs.fromResponse(blockStorageTestResponse("bs-1", "my-bs", "/projects/p/providers/Aruba.Storage/blockStorages/bs-1", "p"))
	bs.SizedGB(40)

	result, err := adapter.Update(context.Background(), bs)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if result.SizeGB() != 40 {
		t.Errorf("Size() = %d", result.SizeGB())
	}
	if capturedBody.Properties.SizeGB != 40 {
		t.Errorf("request SizeGB = %d", capturedBody.Properties.SizeGB)
	}
}

func TestVolumesClientAdapter_Update_NoID(t *testing.T) {
	callCount := 0
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	bs := NewBlockStorage().InProject(URI("/projects/p")).
		Named("x")
	_, err := adapter.Update(context.Background(), bs)
	if err == nil {
		t.Fatal("expected error when BlockStorage has no ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when ID is missing")
	}
}

func TestVolumesClientAdapter_Update_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	bs := &BlockStorage{}
	bs.fromResponse(&types.BlockStorageResponse{
		Metadata: types.ResourceMetadataResponse{
			ID: strPtr("bs-1"),
		},
		Status: types.ResourceStatusResponse{},
	})

	_, err := adapter.Update(context.Background(), bs)
	if err == nil {
		t.Fatal("expected error when BlockStorage has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestVolumesClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Storage/blockStorages/bs-1"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestVolumesClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "block storage not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Storage/blockStorages/missing"))
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

func TestVolumesClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"bs-1","name":"n1","uri":"/projects/p/providers/Aruba.Storage/blockStorages/bs-1","project":{"id":"p"}},"properties":{"sizeGB":10,"type":"Standard","billingPeriod":"Hour"},"status":{}},`+
			`{"metadata":{"id":"bs-2","name":"n2","uri":"/projects/p/providers/Aruba.Storage/blockStorages/bs-2","project":{"id":"p"}},"properties":{"sizeGB":20,"type":"Performance","billingPeriod":"Month"},"status":{}}`+
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
	if items[0].ID() != "bs-1" || items[0].Name() != "n1" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[0].SizeGB() != 10 {
		t.Errorf("items[0].SizeGB() = %d", items[0].SizeGB())
	}
	if items[1].ID() != "bs-2" || items[1].Type() != BlockStorageTypePerformance {
		t.Errorf("items[1] ID=%q Type=%q", items[1].ID(), items[1].Type())
	}
	if items[0].ProjectID() != "p" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
}

// --------------------------------------------------------------------------
// Zero-value accessor tests (Shape F — 66.7% accessors)
// --------------------------------------------------------------------------

func TestBlockStorage_Accessors_ZeroValue(t *testing.T) {
	bs := NewBlockStorage()
	if bs.Type() != "" {
		t.Errorf("Type() zero value = %q, want \"\"", bs.Type())
	}
	if bs.IsBootable() != false {
		t.Error("IsBootable() zero value should be false")
	}
}

func TestNewVolumesClientAdapter_Nil(t *testing.T) {
	// Exercises the nil-rest branch of newVolumesClientAdapter.
	adapter := newVolumesClientAdapter(nil)
	if adapter == nil {
		t.Fatal("newVolumesClientAdapter(nil) returned nil")
	}
}

func TestVolumesClientAdapter_Update_Err(t *testing.T) {
	// Exercise the Err() pre-check in Update.
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	bs := NewBlockStorage().FromSnapshot(URI("")) // seeds an error
	_, err := adapter.Update(context.Background(), bs)
	if err == nil {
		t.Fatal("expected error when BlockStorage has a pre-existing Err()")
	}
}

// --------------------------------------------------------------------------
// Additional adapter coverage tests (Get_BadRef, Update_NonTwoXX, Delete_BadRef,
// List_NonTwoXX)
// --------------------------------------------------------------------------

func TestVolumesClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := adapter.Get(context.Background(), URI("/something/unrelated"))
	if err == nil {
		t.Fatal("expected error for unresolvable Ref")
	}
}

func TestVolumesClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "block storage not found", 404))
	})

	ref := URI("/projects/p/providers/Aruba.Storage/blockStorages/missing")
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

func TestVolumesClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "size invalid", 422))
	})

	bs := &BlockStorage{}
	bs.fromResponse(blockStorageTestResponse("bs-1", "my-bs", "/projects/p/providers/Aruba.Storage/blockStorages/bs-1", "p"))
	bs.SizedGB(99999)

	_, err := adapter.Update(context.Background(), bs)
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

func TestVolumesClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/something/unrelated"))
	if err == nil {
		t.Fatal("expected error for unresolvable Ref")
	}
}

func TestVolumesClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestVolumesClientAdapter_List_BadRef(t *testing.T) {
	adapter := buildVolumesTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := adapter.List(context.Background(), URI("/something/unrelated"))
	if err == nil {
		t.Fatal("expected error for unresolvable project Ref")
	}
}

// containsSubstring reports whether s contains substr.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestBlockStorage_FromResponse_SetsStatus(t *testing.T) {
	b := &BlockStorage{}
	state := types.State("NotUsed")
	b.fromResponse(&types.BlockStorageResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if b.State() != types.StateNotUsed {
		t.Errorf("State() = %q after fromResponse, want NotUsed", b.State())
	}
}

func TestBlockStorage_WaitUntilNotUsed_HappyPath(t *testing.T) {
	b := &BlockStorage{}
	calls := 0
	state := types.State("InCreation")
	b.setRefresh(func(_ context.Context) error {
		calls++
		if calls >= 2 {
			state = "NotUsed"
		}
		s := state
		b.setStatus(&types.ResourceStatusResponse{State: &s})
		return nil
	})
	if err := b.WaitUntilNotUsed(context.Background(), fastOpts()...); err != nil {
		t.Fatalf("WaitUntilNotUsed error: %v", err)
	}
	if b.State() != "NotUsed" {
		t.Errorf("State() = %q after wait, want NotUsed", b.State())
	}
}

func TestBlockStorage_WaitUntilUsed_HappyPath(t *testing.T) {
	for _, attachedState := range []types.State{types.StateInUse, types.StateUsed, types.StateReserved} {
		t.Run(string(attachedState), func(t *testing.T) {
			b := &BlockStorage{}
			calls := 0
			state := types.State("InCreation")
			b.setRefresh(func(_ context.Context) error {
				calls++
				if calls >= 2 {
					state = attachedState
				}
				s := state
				b.setStatus(&types.ResourceStatusResponse{State: &s})
				return nil
			})
			if err := b.WaitUntilUsed(context.Background(), fastOpts()...); err != nil {
				t.Fatalf("WaitUntilUsed error for %q: %v", attachedState, err)
			}
			if b.State() != attachedState {
				t.Errorf("State() = %q after wait, want %q", b.State(), attachedState)
			}
		})
	}
}

func TestVolumesClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, blockStorageSuccessBody)
	})
	adapter := newVolumesClientAdapter(testutil.NewClient(t, server.URL))
	bs, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Storage/blockStorages/bs-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&bs.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned BlockStorage")
	}
}

func TestBlockStorage_UpdateRequest_OmitsUnsetBootable(t *testing.T) {
	// When the API's Get response omits the bootable field (common when the value
	// is the platform default), fromResponse leaves b.bootable nil. The old
	// toRequest() would send bootable=false via blockStorageBootable, silently
	// clobbering a server-side bootable=true. toUpdateRequest must omit the
	// field instead so the server preserves its own value.
	bs := &BlockStorage{}
	bs.fromResponse(&types.BlockStorageResponse{
		Metadata:   types.ResourceMetadataResponse{ID: lo("bs-1"), Name: lo("vol")},
		Properties: types.BlockStoragePropertiesResponse{}, // bootable absent in API response
	})

	req := bs.toUpdateRequest()
	if req.Properties.Bootable != nil {
		t.Errorf("Update request bootable should be nil (omitted) when never set by caller, got %v", *req.Properties.Bootable)
	}
}

func TestBlockStorage_UpdateRequest_ExplicitBootable(t *testing.T) {
	bs := NewBlockStorage().Named("vol").SizedGB(10).NotBootable()
	req := bs.toUpdateRequest()
	if req.Properties.Bootable == nil || *req.Properties.Bootable != false {
		t.Errorf("Update request bootable should be false when UnsetBootable was called, got %v", req.Properties.Bootable)
	}
}

func lo(s string) *string { return &s }
