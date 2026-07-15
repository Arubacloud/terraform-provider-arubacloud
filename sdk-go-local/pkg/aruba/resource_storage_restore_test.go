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

var _ Ref = (*StorageRestore)(nil)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestStorageRestore_FluentSetters(t *testing.T) {
	bkp := &StorageBackup{}
	bkp.fromResponse(storageBackupTestResponse("bkp-1", "my-backup", "/projects/p/providers/Aruba.Storage/backups/bkp-1", "p"))

	r := NewStorageRestore().
		FromBackup(bkp).
		Named("my-restore").
		Tagged("restore").
		Tagged("storage").
		Tagged("restore"). // dedupe
		InRegion(RegionITBGBergamo).
		ToVolume(URI("/projects/p/providers/Aruba.Storage/blockStorages/bs-1"))

	if r.Name() != "my-restore" {
		t.Errorf("Name() = %q", r.Name())
	}
	if tags := r.Tags(); len(tags) != 2 || tags[0] != "restore" || tags[1] != "storage" {
		t.Errorf("Tags() = %v", tags)
	}
	if r.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", r.Region())
	}
	if r.TargetURI() != "/projects/p/providers/Aruba.Storage/blockStorages/bs-1" {
		t.Errorf("TargetURI() = %q", r.TargetURI())
	}
	if r.BackupID() != "bkp-1" {
		t.Errorf("BackupID() = %q", r.BackupID())
	}
	if r.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", r.ProjectID())
	}
	if r.Err() != nil {
		t.Errorf("Err() = %v", r.Err())
	}

	r.Untagged("restore")
	if tags := r.Tags(); len(tags) != 1 || tags[0] != "storage" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	r.RetaggedAs("x", "y")
	if tags := r.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

// --------------------------------------------------------------------------
// IntoBackup — typed / URI / bad Ref
// --------------------------------------------------------------------------

func TestStorageRestore_IntoBackup_TypedRef(t *testing.T) {
	bkp := &StorageBackup{}
	bkp.fromResponse(storageBackupTestResponse("bkp-42", "n", "/projects/p/providers/Aruba.Storage/backups/bkp-42", "p"))

	r := NewStorageRestore().FromBackup(bkp)
	if r.BackupID() != "bkp-42" {
		t.Errorf("BackupID() = %q", r.BackupID())
	}
	if r.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", r.ProjectID())
	}
	if r.Err() != nil {
		t.Errorf("Err() = %v", r.Err())
	}
}

func TestStorageRestore_IntoBackup_URIRef(t *testing.T) {
	r := NewStorageRestore().FromBackup(URI("/projects/p-uri/providers/Aruba.Storage/backups/bkp-uri"))
	if r.BackupID() != "bkp-uri" {
		t.Errorf("BackupID() = %q", r.BackupID())
	}
	if r.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", r.ProjectID())
	}
	if r.Err() != nil {
		t.Errorf("Err() = %v", r.Err())
	}
}

func TestStorageRestore_IntoBackup_BadRef(t *testing.T) {
	r := NewStorageRestore().FromBackup(URI("/something/else"))
	if r.Err() == nil {
		t.Error("expected Err() != nil for unresolvable Ref")
	}
}

// --------------------------------------------------------------------------
// ToVolume
// --------------------------------------------------------------------------

func TestStorageRestore_ToVolume_URIRef(t *testing.T) {
	volURI := "/projects/p/providers/Aruba.Storage/blockStorages/bs-1"
	r := NewStorageRestore().ToVolume(URI(volURI))
	if r.TargetURI() != volURI {
		t.Errorf("TargetURI() = %q", r.TargetURI())
	}
	if r.Err() != nil {
		t.Errorf("Err() = %v", r.Err())
	}
}

func TestStorageRestore_ToVolume_TypedRef(t *testing.T) {
	bs := &BlockStorage{}
	bs.fromResponse(blockStorageTestResponse("bs-42", "n", "/projects/p/providers/Aruba.Storage/blockStorages/bs-42", "p"))

	r := NewStorageRestore().ToVolume(bs)
	if r.TargetURI() != bs.URI() {
		t.Errorf("TargetURI() = %q, want %q", r.TargetURI(), bs.URI())
	}
	if r.Err() != nil {
		t.Errorf("Err() = %v", r.Err())
	}
}

func TestStorageRestore_ToVolume_EmptyURI(t *testing.T) {
	r := NewStorageRestore().ToVolume(URI(""))
	if r.Err() == nil {
		t.Error("expected Err() != nil for empty target URI")
	}
	if r.TargetURI() != "" {
		t.Errorf("TargetURI() should remain empty, got %q", r.TargetURI())
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestStorageRestore_ToRequestRoundTrip(t *testing.T) {
	volURI := "/projects/p/providers/Aruba.Storage/blockStorages/bs-1"
	r := NewStorageRestore().Named(
		"rt-restore").
		Tagged("t1").Tagged("t2").
		InRegion(RegionITBGBergamo).
		ToVolume(URI(volURI))

	req := r.RawRequest()

	if req.Metadata.Name != "rt-restore" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if len(req.Metadata.Tags) != 2 {
		t.Errorf("Metadata.Tags = %v", req.Metadata.Tags)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Location.Value = %q", req.Metadata.Location.Value)
	}
	if req.Properties.Target.URI != volURI {
		t.Errorf("Properties.Target.URI = %q", req.Properties.Target.URI)
	}
}

func TestStorageRestore_ToRequest_UnsetTarget(t *testing.T) {
	r := NewStorageRestore().
		Named("bare")
	req := r.RawRequest()

	if req.Properties.Target.URI != "" {
		t.Errorf("Target.URI should be empty, got %q", req.Properties.Target.URI)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func storageRestoreTestResponse(id, name, uri, targetURI string) *types.StorageRestoreResponse {
	loc := &types.LocationResponse{Value: RegionITBGBergamo}
	state := types.State("Active")
	return &types.StorageRestoreResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			URI:              &uri,
			Name:             &name,
			Tags:             []string{"tag1"},
			LocationResponse: loc,
		},
		Properties: types.StorageRestorePropertiesResponse{
			Destination: types.ReferenceResourceCommon{URI: targetURI},
		},
		Status: types.ResourceStatusResponse{
			State: &state,
		},
	}
}

func TestStorageRestore_FromResponseHydration(t *testing.T) {
	targetURI := "/projects/p/providers/Aruba.Storage/blockStorages/bs-1"
	r := &StorageRestore{}
	r.backupID = "bkp-1" // pre-set to simulate IntoBackup
	resp := storageRestoreTestResponse("r-1", "my-restore", "/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/r-1", targetURI)
	r.fromResponse(resp)

	if r.ID() != "r-1" {
		t.Errorf("ID() = %q", r.ID())
	}
	if r.URI() != "/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/r-1" {
		t.Errorf("URI() = %q", r.URI())
	}
	if r.RestoreID() != "r-1" {
		t.Errorf("RestoreID() = %q", r.RestoreID())
	}
	if r.Name() != "my-restore" {
		t.Errorf("Name() = %q", r.Name())
	}
	if tags := r.Tags(); len(tags) != 1 || tags[0] != "tag1" {
		t.Errorf("Tags() = %v", tags)
	}
	if r.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", r.Region())
	}
	if r.State() != "Active" {
		t.Errorf("State() = %q", r.State())
	}
	if r.TargetURI() != targetURI {
		t.Errorf("TargetURI() = %q", r.TargetURI())
	}
	if r.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestStorageRestore_FromResponseURIBackfill(t *testing.T) {
	id := "r-99"
	uri := "/projects/p-uri/providers/Aruba.Storage/backups/bkp-99/restores/r-99"
	state := types.State("")
	resp := &types.StorageRestoreResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  &id,
			URI: &uri,
		},
		Status: types.ResourceStatusResponse{State: &state},
	}
	r := &StorageRestore{}
	r.fromResponse(resp)

	// Neither ProjectMetadataResponse nor pre-set IDs → backfill from URI.
	if r.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() via URI backfill = %q", r.ProjectID())
	}
	if r.backupID != "bkp-99" {
		t.Errorf("backupID via URI backfill = %q", r.backupID)
	}
}

func TestStorageRestore_FromResponse_NilSafe(t *testing.T) {
	r := &StorageRestore{}
	r.fromResponse(nil)
	if r.ID() != "" || r.URI() != "" || r.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}

	r2 := &StorageRestore{}
	r2.fromResponse(&types.StorageRestoreResponse{})
	if r2.ID() != "" || r2.URI() != "" || r2.State() != "" || r2.TargetURI() != "" {
		t.Error("empty response should yield zero accessor values")
	}
}

// --------------------------------------------------------------------------
// Ref + ancestor ID satisfaction (runtime)
// --------------------------------------------------------------------------

func TestStorageRestore_RefSatisfaction(t *testing.T) {
	r := &StorageRestore{}
	r.fromResponse(storageRestoreTestResponse("r-99", "n", "/projects/p99/providers/Aruba.Storage/backups/bkp-99/restores/r-99", "/projects/p/providers/Aruba.Storage/blockStorages/bs-1"))

	// withRestoreID typed path
	rid, ok := extractID(r, func(ref Ref) (string, bool) {
		if w, ok := ref.(withRestoreID); ok {
			return w.RestoreID(), true
		}
		return "", false
	}, "restores")
	if !ok || rid != "r-99" {
		t.Errorf("extractID via withRestoreID = (%q, %v)", rid, ok)
	}

	// withBackupID typed path
	bid, ok := extractID(r, func(ref Ref) (string, bool) {
		if w, ok := ref.(withBackupID); ok {
			return w.BackupID(), true
		}
		return "", false
	}, "backups")
	if !ok || bid != "bkp-99" {
		t.Errorf("extractID via withBackupID = (%q, %v)", bid, ok)
	}

	// withProjectID typed path
	pid, ok := extractID(r, func(ref Ref) (string, bool) {
		if w, ok := ref.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || pid != "p99" {
		t.Errorf("extractID via withProjectID = (%q, %v)", pid, ok)
	}
}

// --------------------------------------------------------------------------
// restoreIDsFromRef helper
// --------------------------------------------------------------------------

func TestRestoreIDsFromRef_TypedRef(t *testing.T) {
	r := &StorageRestore{}
	r.fromResponse(storageRestoreTestResponse("r-1", "n", "/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/r-1", ""))
	pid, bid, rid, err := restoreIDsFromRef(r)
	if err != nil || pid != "p" || bid != "bkp-1" || rid != "r-1" {
		t.Errorf("restoreIDsFromRef typed = (%q, %q, %q, %v)", pid, bid, rid, err)
	}
}

func TestRestoreIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/r-1")
	pid, bid, rid, err := restoreIDsFromRef(ref)
	if err != nil || pid != "p" || bid != "bkp-1" || rid != "r-1" {
		t.Errorf("restoreIDsFromRef URI = (%q, %q, %q, %v)", pid, bid, rid, err)
	}
}

func TestRestoreIDsFromRef_BadURI_MissingRestore(t *testing.T) {
	_, _, _, err := restoreIDsFromRef(URI("/projects/p/providers/Aruba.Storage/backups/bkp-1"))
	if err == nil {
		t.Error("expected error for URI without /restores/<id>")
	}
}

func TestRestoreIDsFromRef_BadURI_MissingBackup(t *testing.T) {
	_, _, _, err := restoreIDsFromRef(URI("/projects/p/providers/Aruba.Storage/restores/r-1"))
	if err == nil {
		t.Error("expected error for URI without /backups/<id>")
	}
}

func TestRestoreIDsFromRef_BadURI_MissingProject(t *testing.T) {
	_, _, _, err := restoreIDsFromRef(URI("/providers/Aruba.Storage/backups/bkp-1/restores/r-1"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

func TestRestoreIDsFromRef_BadURI_MissingAll(t *testing.T) {
	_, _, _, err := restoreIDsFromRef(URI("/something/else"))
	if err == nil {
		t.Error("expected error for completely unrelated URI")
	}
}

// --------------------------------------------------------------------------
// storageRestoresClientAdapter — fake low-level client
// --------------------------------------------------------------------------

type fakeStorageRestoreLowLevel struct {
	createFunc func(ctx context.Context, projectID, backupID string, body types.StorageRestoreRequest, params *types.RequestParameters) (*types.Response[types.StorageRestoreResponse], error)
	getFunc    func(ctx context.Context, projectID, backupID, restoreID string, params *types.RequestParameters) (*types.Response[types.StorageRestoreResponse], error)
	updateFunc func(ctx context.Context, projectID, backupID, restoreID string, body types.StorageRestoreRequest, params *types.RequestParameters) (*types.Response[types.StorageRestoreResponse], error)
	deleteFunc func(ctx context.Context, projectID, backupID, restoreID string, params *types.RequestParameters) (*types.Response[any], error)
	listFunc   func(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[types.StorageRestoreListResponse], error)
}

func (f *fakeStorageRestoreLowLevel) Create(ctx context.Context, projectID, backupID string, body types.StorageRestoreRequest, params *types.RequestParameters) (*types.Response[types.StorageRestoreResponse], error) {
	return f.createFunc(ctx, projectID, backupID, body, params)
}
func (f *fakeStorageRestoreLowLevel) Get(ctx context.Context, projectID, backupID, restoreID string, params *types.RequestParameters) (*types.Response[types.StorageRestoreResponse], error) {
	return f.getFunc(ctx, projectID, backupID, restoreID, params)
}
func (f *fakeStorageRestoreLowLevel) Update(ctx context.Context, projectID, backupID, restoreID string, body types.StorageRestoreRequest, params *types.RequestParameters) (*types.Response[types.StorageRestoreResponse], error) {
	return f.updateFunc(ctx, projectID, backupID, restoreID, body, params)
}
func (f *fakeStorageRestoreLowLevel) Delete(ctx context.Context, projectID, backupID, restoreID string, params *types.RequestParameters) (*types.Response[any], error) {
	return f.deleteFunc(ctx, projectID, backupID, restoreID, params)
}
func (f *fakeStorageRestoreLowLevel) List(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[types.StorageRestoreListResponse], error) {
	return f.listFunc(ctx, projectID, backupID, params)
}

// --------------------------------------------------------------------------
// storageRestoresClientAdapter — HTTP mock tests
// --------------------------------------------------------------------------

func buildStorageRestoresTestAdapter(t *testing.T, handler http.HandlerFunc) *storageRestoresClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newStorageRestoresClientAdapter(testutil.NewClient(t, server.URL))
}

const storageRestoreSuccessBody = `{` +
	`"metadata":{"id":"r-1","name":"my-restore","uri":"/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/r-1"},` +
	`"properties":{"destinationVolume":{"uri":"/projects/p/providers/Aruba.Storage/blockStorages/bs-1"}},` +
	`"status":{"state":"Active"}}`

func TestStorageRestoresClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.StorageRestoreRequest
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "restores") {
			t.Errorf("path %q should contain 'restores'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, storageRestoreSuccessBody)
	})

	bkp := &StorageBackup{}
	bkp.fromResponse(storageBackupTestResponse("bkp-1", "n", "/projects/p/providers/Aruba.Storage/backups/bkp-1", "p"))

	r := NewStorageRestore().
		FromBackup(bkp).
		Named("my-restore").
		InRegion(RegionITBGBergamo).
		ToVolume(URI("/projects/p/providers/Aruba.Storage/blockStorages/bs-1"))

	result, err := adapter.Create(context.Background(), r)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "r-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-restore" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Metadata.Name != "my-restore" {
		t.Errorf("request Name = %q", gotBody.Metadata.Name)
	}
	if gotBody.Properties.Target.URI != "/projects/p/providers/Aruba.Storage/blockStorages/bs-1" {
		t.Errorf("request Target.URI = %q", gotBody.Properties.Target.URI)
	}
}

func TestStorageRestoresClientAdapter_Create_NoBackup(t *testing.T) {
	callCount := 0
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	_, err := adapter.Create(context.Background(), NewStorageRestore().
		Named("x").ToVolume(URI("/v")))
	if err == nil {
		t.Fatal("expected error when StorageRestore has no parent backup")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without parent backup")
	}
}

func TestStorageRestoresClientAdapter_Create_NoTarget(t *testing.T) {
	callCount := 0
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	bkp := &StorageBackup{}
	bkp.fromResponse(storageBackupTestResponse("bkp-1", "n", "/projects/p/providers/Aruba.Storage/backups/bkp-1", "p"))

	_, err := adapter.Create(context.Background(), NewStorageRestore().FromBackup(bkp).
		Named("x"))
	if err == nil {
		t.Fatal("expected error when StorageRestore has no target")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without target")
	}
}

func TestStorageRestoresClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" field — triggers MetadataValidationError
		fmt.Fprint(w, `{"metadata":{"name":"r","uri":"/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/x"},"properties":{},"status":{}}`)
	})

	bkp := &StorageBackup{}
	bkp.fromResponse(storageBackupTestResponse("bkp-1", "n", "/projects/p/providers/Aruba.Storage/backups/bkp-1", "p"))

	r := NewStorageRestore().FromBackup(bkp).
		Named("r").ToVolume(URI("/v"))
	result, err := adapter.Create(context.Background(), r)
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

func TestStorageRestoresClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "target is required", 422))
	})

	bkp := &StorageBackup{}
	bkp.fromResponse(storageBackupTestResponse("bkp-1", "n", "/projects/p/providers/Aruba.Storage/backups/bkp-1", "p"))

	r := NewStorageRestore().FromBackup(bkp).ToVolume(URI("/v"))
	result, err := adapter.Create(context.Background(), r)
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

func TestStorageRestoresClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, storageRestoreSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/r-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "r-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.BackupID() != "bkp-1" {
		t.Errorf("BackupID() = %q", result.BackupID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	if !containsSubstring(capturedPath, "restores") {
		t.Errorf("path %q should contain 'restores'", capturedPath)
	}
}

func TestStorageRestoresClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, storageRestoreSuccessBody)
	})

	existing := &StorageRestore{}
	existing.fromResponse(storageRestoreTestResponse("r-1", "n", "/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/r-1", ""))

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "r-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.BackupID() != "bkp-1" {
		t.Errorf("BackupID() = %q", result.BackupID())
	}
}

func TestStorageRestoresClientAdapter_Update_Success(t *testing.T) {
	targetURI := "/projects/p/providers/Aruba.Storage/blockStorages/bs-new"
	var capturedBody types.StorageRestoreRequest

	respBody := storageRestoreTestResponse("r-1", "my-restore", "/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/r-1", targetURI)
	resp := &types.Response[types.StorageRestoreResponse]{
		StatusCode: http.StatusOK,
		Data:       respBody,
	}

	fake := &fakeStorageRestoreLowLevel{
		updateFunc: func(_ context.Context, _, _, _ string, body types.StorageRestoreRequest, _ *types.RequestParameters) (*types.Response[types.StorageRestoreResponse], error) {
			capturedBody = body
			return resp, nil
		},
	}
	adapter := &storageRestoresClientAdapter{low: fake}

	r := &StorageRestore{}
	r.fromResponse(storageRestoreTestResponse("r-1", "my-restore", "/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/r-1", "/projects/p/providers/Aruba.Storage/blockStorages/bs-1"))
	r.backupID = "bkp-1"
	r.projectID = "p"
	r.ToVolume(URI(targetURI))

	result, err := adapter.Update(context.Background(), r)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if result.TargetURI() != targetURI {
		t.Errorf("TargetURI() = %q", result.TargetURI())
	}
	if capturedBody.Properties.Target.URI != targetURI {
		t.Errorf("request Target.URI = %q", capturedBody.Properties.Target.URI)
	}
}

func TestStorageRestoresClientAdapter_Update_NoID(t *testing.T) {
	callCount := 0
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	bkp := &StorageBackup{}
	bkp.fromResponse(storageBackupTestResponse("bkp-1", "n", "/projects/p/providers/Aruba.Storage/backups/bkp-1", "p"))

	r := NewStorageRestore().FromBackup(bkp).
		Named("x")
	_, err := adapter.Update(context.Background(), r)
	if err == nil {
		t.Fatal("expected error when StorageRestore has no ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when ID is missing")
	}
}

func TestStorageRestoresClientAdapter_Update_NoBackup(t *testing.T) {
	callCount := 0
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	r := &StorageRestore{}
	id := "r-1"
	r.fromResponse(&types.StorageRestoreResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id},
		Status:   types.ResourceStatusResponse{},
	})

	_, err := adapter.Update(context.Background(), r)
	if err == nil {
		t.Fatal("expected error when StorageRestore has no parent backup")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without parent backup")
	}
}

func TestStorageRestoresClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/r-1"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestStorageRestoresClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "restore not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/missing"))
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

func TestNewStorageRestoresClientAdapter_Nil(t *testing.T) {
	adapter := newStorageRestoresClientAdapter(nil)
	if adapter == nil {
		t.Fatal("newStorageRestoresClientAdapter(nil) returned nil")
	}
}

func TestStorageRestoresClientAdapter_Update_Err(t *testing.T) {
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := NewStorageRestore().ToVolume(URI("")) // seeds an error
	_, err := adapter.Update(context.Background(), r)
	if err == nil {
		t.Fatal("expected error when StorageRestore has a pre-existing Err()")
	}
}

// --------------------------------------------------------------------------
// Additional adapter coverage tests (Get_BadRef, Get_NonTwoXX, Update_NonTwoXX,
// Delete_BadRef, List_NonTwoXX, List_BadRef)
// --------------------------------------------------------------------------

func TestStorageRestoresClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := adapter.Get(context.Background(), URI("/something/unrelated"))
	if err == nil {
		t.Fatal("expected error for unresolvable Ref")
	}
}

func TestStorageRestoresClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "restore not found", 404))
	})

	ref := URI("/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/missing")
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

func TestStorageRestoresClientAdapter_Update_NonTwoXX(t *testing.T) {
	targetURI := "/projects/p/providers/Aruba.Storage/blockStorages/bs-new"
	fake := &fakeStorageRestoreLowLevel{
		updateFunc: func(_ context.Context, _, _, _ string, _ types.StorageRestoreRequest, _ *types.RequestParameters) (*types.Response[types.StorageRestoreResponse], error) {
			return &types.Response[types.StorageRestoreResponse]{
				StatusCode: http.StatusUnprocessableEntity,
			}, nil
		},
	}
	adapter := &storageRestoresClientAdapter{low: fake}

	r := &StorageRestore{}
	r.fromResponse(storageRestoreTestResponse("r-1", "my-restore", "/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/r-1", "/v"))
	r.backupID = "bkp-1"
	r.projectID = "p"
	r.ToVolume(URI(targetURI))

	_, err := adapter.Update(context.Background(), r)
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

func TestStorageRestoresClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/something/unrelated"))
	if err == nil {
		t.Fatal("expected error for unresolvable Ref")
	}
}

func TestStorageRestoresClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Forbidden", "not authorized", 403))
	})

	bkp := &StorageBackup{}
	bkp.fromResponse(storageBackupTestResponse("bkp-1", "n", "/projects/p/providers/Aruba.Storage/backups/bkp-1", "p"))

	_, err := adapter.List(context.Background(), bkp)
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

func TestStorageRestoresClientAdapter_List_BadRef(t *testing.T) {
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := adapter.List(context.Background(), URI("/something/unrelated"))
	if err == nil {
		t.Fatal("expected error for unresolvable backup Ref")
	}
}

func TestStorageRestoresClientAdapter_List_LowLevelError(t *testing.T) {
	// Exercises the err != nil branch of a.low.List.
	fake := &fakeStorageRestoreLowLevel{
		listFunc: func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.StorageRestoreListResponse], error) {
			return nil, fmt.Errorf("network error")
		},
	}
	adapter := &storageRestoresClientAdapter{low: fake}

	bkp := &StorageBackup{}
	bkp.fromResponse(storageBackupTestResponse("bkp-1", "n", "/projects/p/providers/Aruba.Storage/backups/bkp-1", "p"))

	_, err := adapter.List(context.Background(), bkp)
	if err == nil {
		t.Fatal("expected error from low-level list failure")
	}
}

func TestStorageRestoresClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildStorageRestoresTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"r-1","name":"n1","uri":"/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/r-1"},"properties":{"destinationVolume":{"uri":"/projects/p/providers/Aruba.Storage/blockStorages/bs-1"}},"status":{}},`+
			`{"metadata":{"id":"r-2","name":"n2","uri":"/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/r-2"},"properties":{"destinationVolume":{"uri":"/projects/p/providers/Aruba.Storage/blockStorages/bs-2"}},"status":{}}`+
			`]}`)
	})

	bkp := &StorageBackup{}
	bkp.fromResponse(storageBackupTestResponse("bkp-1", "n", "/projects/p/providers/Aruba.Storage/backups/bkp-1", "p"))

	list, err := adapter.List(context.Background(), bkp)
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
	if items[0].ID() != "r-1" || items[0].Name() != "n1" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[0].TargetURI() != "/projects/p/providers/Aruba.Storage/blockStorages/bs-1" {
		t.Errorf("items[0].TargetURI() = %q", items[0].TargetURI())
	}
	if items[1].ID() != "r-2" || items[1].TargetURI() != "/projects/p/providers/Aruba.Storage/blockStorages/bs-2" {
		t.Errorf("items[1] ID=%q TargetURI=%q", items[1].ID(), items[1].TargetURI())
	}
	if items[0].BackupID() != "bkp-1" {
		t.Errorf("items[0].BackupID() = %q", items[0].BackupID())
	}
	if items[0].ProjectID() != "p" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
}

func TestStorageRestore_FromResponse_SetsStatus(t *testing.T) {
	r := &StorageRestore{}
	state := types.State("Active")
	r.fromResponse(&types.StorageRestoreResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if r.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", r.State())
	}
}

func TestStorageRestoresClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, storageRestoreSuccessBody)
	})
	adapter := newStorageRestoresClientAdapter(testutil.NewClient(t, server.URL))
	restore, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Storage/backups/bkp-1/restores/r-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&restore.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned StorageRestore")
	}
}
