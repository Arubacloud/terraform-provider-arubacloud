package aruba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/clients/database"
	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Compile-time interface satisfaction
// --------------------------------------------------------------------------

var (
	_ Ref               = (*DBaaSBackup)(nil)
	_ Wrapper           = (*DBaaSBackup)(nil)
	_ withDBaaSBackupID = (*DBaaSBackup)(nil)
)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestDBaaSBackup_FluentSetters(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-1", "my-proj", "/projects/p-1"))

	dbaasURI := URI("/projects/p-1/providers/Aruba.Database/dbaas/d-1")
	dbURI := URI("/projects/p-1/providers/Aruba.Database/dbaas/d-1/databases/mydb")

	bkp := NewDBaaSBackup().
		InProject(proj).
		Named("my-dbaas-backup").
		Tagged("backup").
		Tagged("dbaas").
		Tagged("backup"). // dedupe
		InRegion(RegionITBGBergamo).
		FromDBaaS(dbaasURI).
		FromDatabase(dbURI).
		BilledBy(BillingPeriodHour)

	if bkp.Name() != "my-dbaas-backup" {
		t.Errorf("Name() = %q", bkp.Name())
	}
	if tags := bkp.Tags(); len(tags) != 2 || tags[0] != "backup" || tags[1] != "dbaas" {
		t.Errorf("Tags() = %v", tags)
	}
	if bkp.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", bkp.Region())
	}
	if bkp.DBaaSURI() != dbaasURI.URI() {
		t.Errorf("DBaaSURI() = %q", bkp.DBaaSURI())
	}
	if bkp.DatabaseURI() != dbURI.URI() {
		t.Errorf("DatabaseURI() = %q", bkp.DatabaseURI())
	}
	if bkp.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", bkp.BillingPeriod())
	}
	if bkp.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", bkp.ProjectID())
	}
	if bkp.Err() != nil {
		t.Errorf("Err() = %v", bkp.Err())
	}

	bkp.Untagged("backup")
	if tags := bkp.Tags(); len(tags) != 1 || tags[0] != "dbaas" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	bkp.RetaggedAs("x", "y")
	if tags := bkp.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

// --------------------------------------------------------------------------
// IntoProject
// --------------------------------------------------------------------------

func TestDBaaSBackup_IntoProject_TypedRef(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-42", "n", "/projects/p-42"))
	bkp := NewDBaaSBackup().InProject(proj)
	if bkp.ProjectID() != "p-42" {
		t.Errorf("ProjectID() = %q", bkp.ProjectID())
	}
	if bkp.Err() != nil {
		t.Errorf("Err() = %v", bkp.Err())
	}
}

func TestDBaaSBackup_IntoProject_URIRef(t *testing.T) {
	bkp := NewDBaaSBackup().InProject(URI("/projects/p-uri"))
	if bkp.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", bkp.ProjectID())
	}
	if bkp.Err() != nil {
		t.Errorf("Err() = %v", bkp.Err())
	}
}

func TestDBaaSBackup_IntoProject_BadRef(t *testing.T) {
	bkp := NewDBaaSBackup().InProject(URI("/garbage"))
	if bkp.Err() == nil {
		t.Error("expected Err() != nil for unresolvable Ref")
	}
}

// --------------------------------------------------------------------------
// FromDBaaS body-ref setter
// --------------------------------------------------------------------------

func TestDBaaSBackup_FromDBaaS_URIRef(t *testing.T) {
	uri := "/projects/p/providers/Aruba.Database/dbaas/d-1"
	bkp := NewDBaaSBackup().FromDBaaS(URI(uri))
	if bkp.DBaaSURI() != uri {
		t.Errorf("DBaaSURI() = %q", bkp.DBaaSURI())
	}
	if bkp.Err() != nil {
		t.Errorf("Err() = %v", bkp.Err())
	}
}

func TestDBaaSBackup_FromDBaaS_TypedRef(t *testing.T) {
	d := &DBaaS{}
	d.fromResponse(dbaasTestResponse("d-1", "n", "/projects/p/providers/Aruba.Database/dbaas/d-1"))

	bkp := NewDBaaSBackup().FromDBaaS(d)
	if bkp.DBaaSURI() != d.URI() {
		t.Errorf("DBaaSURI() = %q, want %q", bkp.DBaaSURI(), d.URI())
	}
	if bkp.Err() != nil {
		t.Errorf("Err() = %v", bkp.Err())
	}
}

func TestDBaaSBackup_FromDBaaS_EmptyURI(t *testing.T) {
	bkp := NewDBaaSBackup().FromDBaaS(URI(""))
	if bkp.Err() == nil {
		t.Error("expected Err() != nil for empty DBaaS URI")
	}
	if bkp.DBaaSURI() != "" {
		t.Errorf("DBaaSURI() should remain empty, got %q", bkp.DBaaSURI())
	}
}

// --------------------------------------------------------------------------
// FromDatabase body-ref setter
// --------------------------------------------------------------------------

func TestDBaaSBackup_FromDatabase_URIRef(t *testing.T) {
	uri := "/projects/p/providers/Aruba.Database/dbaas/d-1/databases/mydb"
	bkp := NewDBaaSBackup().FromDatabase(URI(uri))
	if bkp.DatabaseURI() != uri {
		t.Errorf("DatabaseURI() = %q", bkp.DatabaseURI())
	}
	if bkp.Err() != nil {
		t.Errorf("Err() = %v", bkp.Err())
	}
}

func TestDBaaSBackup_FromDatabase_TypedRef(t *testing.T) {
	// Database.URI() is constructed from its ancestors, so we test via URI() directly.
	dbRef := URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/mydb")

	bkp := NewDBaaSBackup().FromDatabase(dbRef)
	if bkp.DatabaseURI() != dbRef.URI() {
		t.Errorf("DatabaseURI() = %q, want %q", bkp.DatabaseURI(), dbRef.URI())
	}
	if bkp.Err() != nil {
		t.Errorf("Err() = %v", bkp.Err())
	}
}

func TestDBaaSBackup_FromDatabase_EmptyURI(t *testing.T) {
	bkp := NewDBaaSBackup().FromDatabase(URI(""))
	if bkp.Err() == nil {
		t.Error("expected Err() != nil for empty Database URI")
	}
	if bkp.DatabaseURI() != "" {
		t.Errorf("DatabaseURI() should remain empty, got %q", bkp.DatabaseURI())
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestDBaaSBackup_ToRequest(t *testing.T) {
	dbaasURI := "/projects/p/providers/Aruba.Database/dbaas/d-1"
	dbURI := "/projects/p/providers/Aruba.Database/dbaas/d-1/databases/mydb"

	bkp := NewDBaaSBackup().Named(
		"bkp-rt").
		Tagged("t1").Tagged("t2").
		InRegion(RegionITBGBergamo).
		FromDBaaS(URI(dbaasURI)).
		FromDatabase(URI(dbURI)).
		BilledBy(BillingPeriodHour)

	req := bkp.RawRequest()

	if req.Metadata.Name != "bkp-rt" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if len(req.Metadata.Tags) != 2 {
		t.Errorf("Metadata.Tags = %v", req.Metadata.Tags)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Location.Value = %q", req.Metadata.Location.Value)
	}
	if req.Properties.DBaaS.URI != dbaasURI {
		t.Errorf("Properties.DBaaS.URI = %q", req.Properties.DBaaS.URI)
	}
	if req.Properties.Database.Name != "mydb" {
		t.Errorf("Properties.Database.Name = %q, want %q", req.Properties.Database.Name, "mydb")
	}
	if req.Properties.BillingPlanCommon == nil || req.Properties.BillingPlanCommon.BillingPeriod == nil || *req.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodHour {
		t.Errorf("Properties.BillingPlanCommon.BillingPeriod = %v", req.Properties.BillingPlanCommon)
	}
}

func TestDBaaSBackup_ToRequest_ZoneFromLocation(t *testing.T) {
	bkp := NewDBaaSBackup().InRegion(RegionITBGBergamo)
	req := bkp.RawRequest()
	if req.Properties.Zone != Zone(RegionITBGBergamo) {
		t.Errorf("Properties.Zone = %q, want ITBG-Bergamo (auto-derived from location)", req.Properties.Zone)
	}
}

func TestDBaaSBackup_ToRequest_Empty(t *testing.T) {
	bkp := NewDBaaSBackup()
	req := bkp.RawRequest() // must not panic
	if req.Properties.Zone != "" {
		t.Errorf("Zone should be empty for bare wrapper, got %q", req.Properties.Zone)
	}
	if req.Properties.DBaaS.URI != "" {
		t.Errorf("DBaaS.URI should be empty, got %q", req.Properties.DBaaS.URI)
	}
	if req.Properties.Database.Name != "" {
		t.Errorf("Database.Name should be empty for bare wrapper, got %q", req.Properties.Database.Name)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration helpers
// --------------------------------------------------------------------------

func dbaasBackupTestResponse(name string) *types.BackupResponse {
	id := "bkp-1"
	uri := "/projects/p/providers/Aruba.Database/backups/bkp-1"
	state := types.State("Active")
	dbaasURI := "/projects/p/providers/Aruba.Database/dbaas/d-1"
	return &types.BackupResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			URI:              &uri,
			Name:             func() *string { s := name; return &s }(),
			Tags:             []string{"tag1"},
			LocationResponse: &types.LocationResponse{Value: RegionITBGBergamo},
			ProjectMetadataResponse: &types.ProjectMetadataResponse{
				ID: "p",
			},
		},
		Properties: types.BackupPropertiesResponse{
			Zone:     ZoneITBG1,
			DBaaS:    types.ReferenceResourceCommon{URI: dbaasURI},
			Database: types.DatabaseNameRef{Name: "mydb"},
			BillingPlanCommon: func() *types.BillingPlanCommon {
				v := BillingPeriodHour
				return &types.BillingPlanCommon{BillingPeriod: &v}
			}(),
			Storage: types.BackupStorageResponse{Size: 50},
		},
		Status: types.ResourceStatusResponse{
			State: &state,
		},
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration tests
// --------------------------------------------------------------------------

func TestDBaaSBackup_FromResponseHydration(t *testing.T) {
	bkp := &DBaaSBackup{}
	resp := dbaasBackupTestResponse("my-backup")
	bkp.fromResponse(resp)

	if bkp.ID() != "bkp-1" {
		t.Errorf("ID() = %q", bkp.ID())
	}
	if bkp.DBaaSBackupID() != "bkp-1" {
		t.Errorf("DBaaSBackupID() = %q", bkp.DBaaSBackupID())
	}
	if bkp.URI() != "/projects/p/providers/Aruba.Database/backups/bkp-1" {
		t.Errorf("URI() = %q", bkp.URI())
	}
	if bkp.Name() != "my-backup" {
		t.Errorf("Name() = %q", bkp.Name())
	}
	if tags := bkp.Tags(); len(tags) != 1 || tags[0] != "tag1" {
		t.Errorf("Tags() = %v", tags)
	}
	if bkp.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", bkp.Region())
	}
	if bkp.State() != "Active" {
		t.Errorf("State() = %q", bkp.State())
	}
	if bkp.DBaaSURI() != "/projects/p/providers/Aruba.Database/dbaas/d-1" {
		t.Errorf("DBaaSURI() = %q", bkp.DBaaSURI())
	}
	// After hydration from an API response, databaseRef holds the bare name
	// returned by the API, so DatabaseURI() and DatabaseName() both return it.
	if bkp.DatabaseURI() != "mydb" {
		t.Errorf("DatabaseURI() after hydration = %q, want %q", bkp.DatabaseURI(), "mydb")
	}
	if bkp.DatabaseName() != "mydb" {
		t.Errorf("DatabaseName() after hydration = %q, want %q", bkp.DatabaseName(), "mydb")
	}
	if bkp.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", bkp.BillingPeriod())
	}
	if bkp.Zone() != ZoneITBG1 {
		t.Errorf("Zone() = %q", bkp.Zone())
	}
	if bkp.SizeGB() != 50 {
		t.Errorf("Size() = %d", bkp.SizeGB())
	}
	if bkp.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", bkp.ProjectID())
	}
	if bkp.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestDBaaSBackup_FromResponse_NilSafe(t *testing.T) {
	bkp := &DBaaSBackup{}
	bkp.fromResponse(nil) // must not panic
	if bkp.ID() != "" || bkp.URI() != "" || bkp.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}
}

func TestDBaaSBackup_FromResponse_BackfillsProjectID_FromMetadata(t *testing.T) {
	resp := dbaasBackupTestResponse("n")
	// ProjectMetadataResponse.ID is "p" — should be backfilled directly.
	bkp := &DBaaSBackup{}
	bkp.fromResponse(resp)
	if bkp.ProjectID() != "p" {
		t.Errorf("ProjectID() from metadata = %q", bkp.ProjectID())
	}
}

func TestDBaaSBackup_FromResponse_BackfillsProjectID_FromURI(t *testing.T) {
	id := "bkp-99"
	uri := "/projects/p-uri/providers/Aruba.Database/backups/bkp-99"
	resp := &types.BackupResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  &id,
			URI: &uri,
			// No ProjectMetadataResponse — should backfill from URI.
		},
	}
	bkp := &DBaaSBackup{}
	bkp.fromResponse(resp)
	if bkp.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() via URI backfill = %q", bkp.ProjectID())
	}
}

// --------------------------------------------------------------------------
// dbaasBackupIDsFromRef helper
// --------------------------------------------------------------------------

func TestDBaaSBackupIDsFromRef_TypedRef(t *testing.T) {
	bkp := &DBaaSBackup{}
	bkp.fromResponse(dbaasBackupTestResponse("n"))
	// white-box: set projectID so typed path is exercised
	bkp.projectID = "p"
	pid, bid, err := dbaasBackupIDsFromRef(bkp)
	if err != nil || pid != "p" || bid != "bkp-1" {
		t.Errorf("dbaasBackupIDsFromRef typed = (%q, %q, %v)", pid, bid, err)
	}
}

func TestDBaaSBackupIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Database/backups/b-1")
	pid, bid, err := dbaasBackupIDsFromRef(ref)
	if err != nil || pid != "p" || bid != "b-1" {
		t.Errorf("dbaasBackupIDsFromRef URI = (%q, %q, %v)", pid, bid, err)
	}
}

func TestDBaaSBackupIDsFromRef_BadURI_MissingBackups(t *testing.T) {
	_, _, err := dbaasBackupIDsFromRef(URI("/projects/p/providers/Aruba.Database/something/else"))
	if err == nil {
		t.Error("expected error for URI without /backups/<id>")
	}
}

func TestDBaaSBackupIDsFromRef_BadURI_MissingProject(t *testing.T) {
	_, _, err := dbaasBackupIDsFromRef(URI("/providers/Aruba.Database/backups/b-1"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

// --------------------------------------------------------------------------
// fakeDBaaSBackupLowLevel — body-capture tests
// --------------------------------------------------------------------------

type fakeDBaaSBackupLowLevel struct {
	createFunc func(ctx context.Context, projectID string, body types.BackupRequest, params *types.RequestParameters) (*types.Response[types.BackupResponse], error)
	getFunc    func(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[types.BackupResponse], error)
	deleteFunc func(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[any], error)
	listFunc   func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.DBaaSBackupListResponse], error)
}

func (f *fakeDBaaSBackupLowLevel) Create(ctx context.Context, projectID string, body types.BackupRequest, params *types.RequestParameters) (*types.Response[types.BackupResponse], error) {
	return f.createFunc(ctx, projectID, body, params)
}
func (f *fakeDBaaSBackupLowLevel) Get(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[types.BackupResponse], error) {
	return f.getFunc(ctx, projectID, backupID, params)
}
func (f *fakeDBaaSBackupLowLevel) Delete(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[any], error) {
	return f.deleteFunc(ctx, projectID, backupID, params)
}
func (f *fakeDBaaSBackupLowLevel) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.DBaaSBackupListResponse], error) {
	return f.listFunc(ctx, projectID, params)
}

// --------------------------------------------------------------------------
// HTTP-mock adapter helper
// --------------------------------------------------------------------------

func buildDBaaSBackupsTestAdapter(t *testing.T, handler http.HandlerFunc) *dbaasBackupsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newDBaaSBackupsClientAdapter(testutil.NewClient(t, server.URL))
}

const dbaasBackupSuccessBody = `{` +
	`"metadata":{"id":"bkp-1","name":"my-backup","uri":"/projects/p/providers/Aruba.Database/backups/bkp-1","project":{"id":"p"}},` +
	`"properties":{"datacenter":"ITBG-1","dbaas":{"uri":"/projects/p/providers/Aruba.Database/dbaas/d-1"},"database":{"name":"mydb"},"billingPlan":{"billingPeriod":"Hour"}},` +
	`"status":{"state":"Active"}}`

// --------------------------------------------------------------------------
// Create adapter tests
// --------------------------------------------------------------------------

func TestDBaaSBackupsClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.BackupRequest
	adapter := buildDBaaSBackupsTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "backups") {
			t.Errorf("path %q should contain 'backups'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, dbaasBackupSuccessBody)
	})

	bkp := NewDBaaSBackup().
		InProject(URI("/projects/p")).
		Named("my-backup").
		InRegion(RegionITBGBergamo).
		FromDBaaS(URI("/projects/p/providers/Aruba.Database/dbaas/d-1")).
		FromDatabase(URI("/projects/p/providers/Aruba.Database/dbaas/d-1/databases/mydb")).
		BilledBy(BillingPeriodHour)

	result, err := adapter.Create(context.Background(), bkp)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "bkp-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-backup" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	// Wire body assertions
	if gotBody.Metadata.Name != "my-backup" {
		t.Errorf("request Metadata.Name = %q", gotBody.Metadata.Name)
	}
	if gotBody.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("request Metadata.Location.Value = %q", gotBody.Metadata.Location.Value)
	}
	if gotBody.Properties.Zone != Zone(RegionITBGBergamo) {
		t.Errorf("request Properties.Zone = %q", gotBody.Properties.Zone)
	}
	if gotBody.Properties.DBaaS.URI != "/projects/p/providers/Aruba.Database/dbaas/d-1" {
		t.Errorf("request Properties.DBaaS.URI = %q", gotBody.Properties.DBaaS.URI)
	}
	if gotBody.Properties.Database.Name != "mydb" {
		t.Errorf("request Properties.Database.Name = %q, want %q", gotBody.Properties.Database.Name, "mydb")
	}
	if gotBody.Properties.BillingPlanCommon == nil || gotBody.Properties.BillingPlanCommon.BillingPeriod == nil || *gotBody.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodHour {
		t.Errorf("request Properties.BillingPlanCommon.BillingPeriod = %v", gotBody.Properties.BillingPlanCommon)
	}
}

func TestDBaaSBackupsClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildDBaaSBackupsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	_, err := adapter.Create(context.Background(), NewDBaaSBackup().
		Named("x"))
	if err == nil {
		t.Fatal("expected error when DBaaSBackup has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestDBaaSBackupsClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildDBaaSBackupsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" field — triggers MetadataValidationError from low-level Validate()
		fmt.Fprint(w, `{"metadata":{"name":"bkp","uri":"/projects/p/providers/Aruba.Database/backups/x"},"properties":{},"status":{}}`)
	})

	bkp := NewDBaaSBackup().InProject(URI("/projects/p")).
		Named("bkp")
	result, err := adapter.Create(context.Background(), bkp)
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

func TestDBaaSBackupsClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildDBaaSBackupsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "name is required", 422))
	})

	bkp := NewDBaaSBackup().InProject(URI("/projects/p"))
	result, err := adapter.Create(context.Background(), bkp)
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

func TestDBaaSBackupsClientAdapter_Create_WithBodyRefs_ViaFake(t *testing.T) {
	dbaasURI := "/projects/p/providers/Aruba.Database/dbaas/d-1"
	dbURI := "/projects/p/providers/Aruba.Database/dbaas/d-1/databases/mydb"

	var captured types.BackupRequest
	resp := &types.Response[types.BackupResponse]{
		StatusCode: http.StatusCreated,
		Data:       dbaasBackupTestResponse("bkp"),
	}
	fake := &fakeDBaaSBackupLowLevel{
		createFunc: func(_ context.Context, _ string, body types.BackupRequest, _ *types.RequestParameters) (*types.Response[types.BackupResponse], error) {
			captured = body
			return resp, nil
		},
	}
	adapter := &dbaasBackupsClientAdapter{low: fake}

	bkp := NewDBaaSBackup().
		InProject(URI("/projects/p")).
		InRegion(RegionITBGBergamo).
		FromDBaaS(URI(dbaasURI)).
		FromDatabase(URI(dbURI)).
		BilledBy(BillingPeriodHour)

	_, err := adapter.Create(context.Background(), bkp)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if captured.Properties.DBaaS.URI != dbaasURI {
		t.Errorf("captured DBaaS.URI = %q", captured.Properties.DBaaS.URI)
	}
	if captured.Properties.Database.Name != "mydb" {
		t.Errorf("captured Database.Name = %q, want %q", captured.Properties.Database.Name, "mydb")
	}
	if captured.Properties.Zone != Zone(RegionITBGBergamo) {
		t.Errorf("captured Zone = %q", captured.Properties.Zone)
	}
	if captured.Properties.BillingPlanCommon == nil || captured.Properties.BillingPlanCommon.BillingPeriod == nil || *captured.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodHour {
		t.Errorf("captured BillingPlanCommon.BillingPeriod = %v", captured.Properties.BillingPlanCommon)
	}
}

// --------------------------------------------------------------------------
// Get adapter tests
// --------------------------------------------------------------------------

func TestDBaaSBackupsClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildDBaaSBackupsTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, dbaasBackupSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Database/backups/bkp-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "bkp-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	if !containsSubstring(capturedPath, "backups") {
		t.Errorf("path %q should contain 'backups'", capturedPath)
	}
}

func TestDBaaSBackupsClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildDBaaSBackupsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, dbaasBackupSuccessBody)
	})

	existing := &DBaaSBackup{}
	existing.fromResponse(dbaasBackupTestResponse("my-backup"))
	existing.projectID = "p"

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "bkp-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

// --------------------------------------------------------------------------
// Delete adapter tests
// --------------------------------------------------------------------------

func TestDBaaSBackupsClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildDBaaSBackupsTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Database/backups/bkp-1"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestDBaaSBackupsClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildDBaaSBackupsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "backup not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Database/backups/missing"))
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
// List adapter tests
// --------------------------------------------------------------------------

// --------------------------------------------------------------------------
// InRegion (0% → covers that branch)
// --------------------------------------------------------------------------

func TestDBaaSBackup_InRegion(t *testing.T) {
	bkp := NewDBaaSBackup().InRegion(RegionITBGBergamo)
	if bkp.Region() != RegionITBGBergamo {
		t.Errorf("Region() after InRegion = %q", bkp.Region())
	}
}

// --------------------------------------------------------------------------
// Zero-value accessors (Shape F — covers the nil-response branch)
// --------------------------------------------------------------------------

func TestDBaaSBackup_Accessors_ZeroValue(t *testing.T) {
	bkp := &DBaaSBackup{}
	if bkp.SizeGB() != 0 {
		t.Errorf("Size() zero = %d", bkp.SizeGB())
	}
	if bkp.Zone() != "" {
		t.Errorf("Zone() zero = %q", bkp.Zone())
	}
}

// --------------------------------------------------------------------------
// Get — bad Ref and non-2xx
// --------------------------------------------------------------------------

func TestDBaaSBackupsClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildDBaaSBackupsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.Get(context.Background(), URI("/something/not-a-backup"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
}

func TestDBaaSBackupsClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildDBaaSBackupsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "backup not found", 404))
	})
	ref := URI("/projects/p/providers/Aruba.Database/backups/bkp-1")
	_, err := adapter.Get(context.Background(), ref)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("HTTPError.StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Delete — bad Ref
// --------------------------------------------------------------------------

func TestDBaaSBackupsClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildDBaaSBackupsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := adapter.Delete(context.Background(), URI("/something/bad"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
}

// --------------------------------------------------------------------------
// List — non-2xx response
// --------------------------------------------------------------------------

func TestDBaaSBackupsClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildDBaaSBackupsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Forbidden", "not allowed", 403))
	})
	_, err := adapter.List(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected error on 403")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusForbidden {
		t.Errorf("HTTPError.StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Get — broken client
// --------------------------------------------------------------------------

func TestDBaaSBackupsClientAdapter_Get_BrokenClient(t *testing.T) {
	adapter := &dbaasBackupsClientAdapter{low: database.NewBackupsClientImpl(testutil.NewBrokenClient(t, "http://localhost:9"))}
	_, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Database/backups/bkp-1"))
	if err == nil {
		t.Fatal("expected network error from broken client")
	}
}

// --------------------------------------------------------------------------
// List — bad parent ref
// --------------------------------------------------------------------------

func TestDBaaSBackupsClientAdapter_List_BadRef(t *testing.T) {
	adapter := buildDBaaSBackupsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/something/bad"))
	if err == nil {
		t.Fatal("expected error for bad parent Ref")
	}
}

func TestDBaaSBackupsClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildDBaaSBackupsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"bkp-1","name":"n1","uri":"/projects/p/providers/Aruba.Database/backups/bkp-1","project":{"id":"p"}},"properties":{"datacenter":"ITBG-1","dbaas":{"uri":"/projects/p/providers/Aruba.Database/dbaas/d-1"},"database":{"name":"mydb"},"billingPlan":{"billingPeriod":"Hour"}},"status":{}},`+
			`{"metadata":{"id":"bkp-2","name":"n2","uri":"/projects/p/providers/Aruba.Database/backups/bkp-2","project":{"id":"p"}},"properties":{"datacenter":"ITBG-1","dbaas":{"uri":"/projects/p/providers/Aruba.Database/dbaas/d-1"},"database":{"name":"mydb"},"billingPlan":{"billingPeriod":"Hour"}},"status":{}}`+
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
	if items[0].ID() != "bkp-1" || items[0].Name() != "n1" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[0].BillingPeriod() != BillingPeriodHour {
		t.Errorf("items[0].BillingPeriod() = %q", items[0].BillingPeriod())
	}
	if items[1].ID() != "bkp-2" || items[1].BillingPeriod() != BillingPeriodHour {
		t.Errorf("items[1] ID=%q BillingPeriod=%q", items[1].ID(), items[1].BillingPeriod())
	}
	if items[0].ProjectID() != "p" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
}

// --------------------------------------------------------------------------
// Reflective check: BackupsClient has no Update method
// --------------------------------------------------------------------------

func TestDBaaSBackupsClient_NoUpdateMethod(t *testing.T) {
	iface := reflect.TypeOf((*BackupsClient)(nil)).Elem()
	for i := range iface.NumMethod() {
		if iface.Method(i).Name == "Update" {
			t.Fatal("BackupsClient must NOT have an Update method (underlying API does not support it)")
		}
	}
}

func TestDBaaSBackup_FromResponse_SetsStatus(t *testing.T) {
	b := &DBaaSBackup{}
	state := types.State("Active")
	b.fromResponse(&types.BackupResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if b.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", b.State())
	}
}

func TestDBaaSBackup_InZone_OverridesRegionDerivedZone(t *testing.T) {
	b := NewDBaaSBackup().
		InRegion(RegionITBGBergamo).
		InZone(ZoneITBG1)
	req := b.RawRequest()
	if req.Properties.Zone != ZoneITBG1 {
		t.Errorf("Zone = %q, want %q", req.Properties.Zone, ZoneITBG1)
	}
}

func TestDBaaSBackup_NoInZone_DerivesZoneFromRegion(t *testing.T) {
	b := NewDBaaSBackup().InRegion(RegionITBGBergamo)
	req := b.RawRequest()
	if req.Properties.Zone != Zone(RegionITBGBergamo) {
		t.Errorf("Zone = %q, want region-derived %q", req.Properties.Zone, Zone(RegionITBGBergamo))
	}
}

func TestDBaaSBackupsClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, dbaasBackupSuccessBody)
	})
	adapter := newDBaaSBackupsClientAdapter(testutil.NewClient(t, server.URL))
	bkp, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Database/backups/bkp-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&bkp.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned DBaaSBackup")
	}
}
