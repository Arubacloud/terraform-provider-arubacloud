package aruba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/internal/clients/schedule"
	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Compile-time interface satisfaction
// --------------------------------------------------------------------------

var (
	_ Ref     = (*Job)(nil)
	_ Wrapper = (*Job)(nil)
)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestJob_FluentSetters(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-1", "my-proj", "/projects/p-1"))

	j := NewJob().
		InProject(proj).
		Named("my-job").
		Tagged("env:prod").
		Tagged("schedule").
		Tagged("env:prod"). // dedupe
		InRegion(RegionITBGBergamo).
		Enabled()

	if j.Name() != "my-job" {
		t.Errorf("Name() = %q", j.Name())
	}
	if tags := j.Tags(); len(tags) != 2 || tags[0] != "env:prod" || tags[1] != "schedule" {
		t.Errorf("Tags() = %v", tags)
	}
	if j.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", j.Region())
	}
	if !j.IsEnabled() {
		t.Error("Enabled() should be true")
	}
	if j.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", j.ProjectID())
	}
	if j.Err() != nil {
		t.Errorf("Err() = %v", j.Err())
	}
}

// --------------------------------------------------------------------------
// IntoProject
// --------------------------------------------------------------------------

func TestJob_IntoProject_TypedRef(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-42", "proj", "/projects/p-42"))
	j := NewJob().InProject(proj)
	if j.ProjectID() != "p-42" {
		t.Errorf("ProjectID() = %q", j.ProjectID())
	}
	if j.Err() != nil {
		t.Errorf("Err() = %v", j.Err())
	}
}

func TestJob_IntoProject_URIRef(t *testing.T) {
	j := NewJob().InProject(URI("/projects/p-uri"))
	if j.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", j.ProjectID())
	}
}

func TestJob_IntoProject_BadRef(t *testing.T) {
	j := NewJob().InProject(URI("not-a-project-uri"))
	if j.Err() == nil {
		t.Error("expected Err() != nil for non-project URI")
	}
}

// --------------------------------------------------------------------------
// WithEnabled
// --------------------------------------------------------------------------

func TestJob_WithEnabled_True(t *testing.T) {
	j := NewJob().Enabled()
	req := j.RawRequest()
	if req.Properties.Enabled == nil || !*req.Properties.Enabled {
		t.Error("Enabled should be true")
	}
}

func TestJob_WithEnabled_False(t *testing.T) {
	j := NewJob().Disabled()
	if j.enabled == nil || *j.enabled != false {
		t.Error("enabled *bool should be set to false")
	}
}

// --------------------------------------------------------------------------
// Schedule setters — happy paths
// --------------------------------------------------------------------------

func TestJob_OneShotAt(t *testing.T) {
	ts := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	j := NewJob().OneShotAt(ts)
	if j.Err() != nil {
		t.Fatalf("Err() = %v", j.Err())
	}
	if j.JobType() != JobTypeOneShot {
		t.Errorf("JobType() = %q", j.JobType())
	}
	if j.scheduleAt == nil || *j.scheduleAt != "2026-05-01T12:00:00Z" {
		t.Errorf("scheduleAt = %v", j.scheduleAt)
	}
}

func TestJob_WithCron(t *testing.T) {
	j := NewJob().WithCron("0 8 * * 1-5")
	if j.Err() != nil {
		t.Fatalf("Err() = %v", j.Err())
	}
	if j.JobType() != JobTypeRecurring {
		t.Errorf("JobType() = %q", j.JobType())
	}
	if j.Cron() != "0 8 * * 1-5" {
		t.Errorf("Cron() = %q", j.Cron())
	}
}

func TestJob_RecurringUntil(t *testing.T) {
	ts := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	j := NewJob().RecurringUntil(ts)
	if j.Err() != nil {
		t.Fatalf("Err() = %v", j.Err())
	}
	if j.JobType() != JobTypeRecurring {
		t.Errorf("JobType() = %q", j.JobType())
	}
	if j.executeUntil == nil || *j.executeUntil != "2026-12-31T00:00:00Z" {
		t.Errorf("executeUntil = %v", j.executeUntil)
	}
}

func TestJob_WithCron_And_RecurringUntil(t *testing.T) {
	ts := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	j := NewJob().WithCron("0 8 * * 1-5").RecurringUntil(ts)
	if j.Err() != nil {
		t.Fatalf("Cron+Until should not error, got: %v", j.Err())
	}
	if j.JobType() != JobTypeRecurring {
		t.Errorf("JobType() = %q", j.JobType())
	}
	if j.Cron() != "0 8 * * 1-5" {
		t.Errorf("Cron() = %q", j.Cron())
	}
	if j.executeUntil == nil {
		t.Error("executeUntil should be set")
	}
}

// --------------------------------------------------------------------------
// Schedule setters — mode-conflict errors
// --------------------------------------------------------------------------

func TestJob_OneShotAt_Then_WithCron_Errors(t *testing.T) {
	ts := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	j := NewJob().OneShotAt(ts).WithCron("0 8 * * 1-5")
	if j.Err() == nil {
		t.Error("expected error mixing OneShot and Recurring")
	}
}

func TestJob_OneShotAt_Then_RecurringUntil_Errors(t *testing.T) {
	ts := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	j := NewJob().OneShotAt(ts).RecurringUntil(ts)
	if j.Err() == nil {
		t.Error("expected error mixing OneShot and Recurring")
	}
}

func TestJob_WithCron_Then_OneShotAt_Errors(t *testing.T) {
	ts := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	j := NewJob().WithCron("0 8 * * 1-5").OneShotAt(ts)
	if j.Err() == nil {
		t.Error("expected error mixing Recurring and OneShot")
	}
}

func TestJob_StartingAt_NeutralOnEmpty(t *testing.T) {
	ts := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	j := NewJob().StartingAt(ts)
	if j.Err() != nil {
		t.Fatalf("unexpected error: %v", j.Err())
	}
	if j.JobType() != "" {
		t.Errorf("JobType() = %q; want empty (mode-neutral)", j.JobType())
	}
	req := j.RawRequest()
	if req.Properties.ScheduleAt == nil {
		t.Fatal("ScheduleAt should be set")
	}
	if got, want := *req.Properties.ScheduleAt, "2026-05-01T12:00:00Z"; got != want {
		t.Errorf("ScheduleAt = %q; want %q", got, want)
	}
}

func TestJob_StartingAt_CompatibleWithRecurring(t *testing.T) {
	start := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	j := NewJob().WithCron("0 10 * * *").StartingAt(start).RecurringUntil(end)
	if j.Err() != nil {
		t.Fatalf("unexpected error: %v", j.Err())
	}
	if j.JobType() != JobTypeRecurring {
		t.Errorf("JobType() = %q; want %q", j.JobType(), JobTypeRecurring)
	}
	req := j.RawRequest()
	if req.Properties.Cron == nil || *req.Properties.Cron != "0 10 * * *" {
		t.Errorf("Cron = %v; want %q", req.Properties.Cron, "0 10 * * *")
	}
	if req.Properties.ScheduleAt == nil {
		t.Error("ScheduleAt should be set")
	}
	if req.Properties.ExecuteUntil == nil {
		t.Error("ExecuteUntil should be set")
	}
}

func TestJob_OneShotAt_StillForcesOneShotMode(t *testing.T) {
	ts := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	j := NewJob().OneShotAt(ts).WithCron("0 8 * * 1-5")
	if j.Err() == nil {
		t.Error("expected error mixing OneShot (OneShotAt) and Recurring (WithCron)")
	}
}

func TestJob_OfType_PopulatesScheduleJobType(t *testing.T) {
	j := NewJob().OfType(JobTypeOneShot)
	if j.Err() != nil {
		t.Fatalf("unexpected error: %v", j.Err())
	}
	if j.JobType() != JobTypeOneShot {
		t.Errorf("JobType() = %q; want %q", j.JobType(), JobTypeOneShot)
	}
	if got := j.RawRequest().Properties.JobType; got != JobTypeOneShot {
		t.Errorf("wire JobType = %q; want %q", got, JobTypeOneShot)
	}
}

func TestJob_OfType_CompatibleWithMatchingOneShotAt(t *testing.T) {
	ts := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	j := NewJob().OfType(JobTypeOneShot).OneShotAt(ts)
	if j.Err() != nil {
		t.Fatalf("unexpected error: %v", j.Err())
	}
	req := j.RawRequest()
	if req.Properties.JobType != JobTypeOneShot {
		t.Errorf("JobType = %q; want %q", req.Properties.JobType, JobTypeOneShot)
	}
	if req.Properties.ScheduleAt == nil {
		t.Error("ScheduleAt should be set")
	}
}

func TestJob_OfType_CompatibleWithMatchingWithCron(t *testing.T) {
	j := NewJob().OfType(JobTypeRecurring).WithCron("0 10 * * *")
	if j.Err() != nil {
		t.Fatalf("unexpected error: %v", j.Err())
	}
	if j.JobType() != JobTypeRecurring {
		t.Errorf("JobType() = %q; want %q", j.JobType(), JobTypeRecurring)
	}
}

func TestJob_OfType_ConflictsWithOppositeMode(t *testing.T) {
	ts := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	j := NewJob().OfType(JobTypeRecurring).OneShotAt(ts)
	if j.Err() == nil {
		t.Error("expected error mixing Recurring (OfType) and OneShot (OneShotAt)")
	}
	j2 := NewJob().OfType(JobTypeOneShot).WithCron("0 8 * * 1-5")
	if j2.Err() == nil {
		t.Error("expected error mixing OneShot (OfType) and Recurring (WithCron)")
	}
}

// --------------------------------------------------------------------------
// JobStep sub-builder
// --------------------------------------------------------------------------

func TestJobStep_Build_Basic(t *testing.T) {
	s := NewJobStep().
		Named("restart").
		Targeting(URI("/projects/p/providers/Aruba.Compute/cloudServers/srv-1")).
		WithAction("/projects/p/providers/Aruba.Compute/cloudServers/srv-1/providers/Aruba.Compute/actions/reboot").
		WithVerb(HTTPVerbPOST).
		WithBody(`{"force":true}`)

	out := s.build()
	if out.Name == nil || *out.Name != "restart" {
		t.Errorf("Name = %v", out.Name)
	}
	if out.ResourceURI != "/projects/p/providers/Aruba.Compute/cloudServers/srv-1" {
		t.Errorf("ResourceURI = %q", out.ResourceURI)
	}
	if out.HttpVerb != HTTPVerbPOST {
		t.Errorf("HttpVerb = %q", out.HttpVerb)
	}
	if out.Body == nil || *out.Body != `{"force":true}` {
		t.Errorf("Body = %v", out.Body)
	}
}

func TestJobStep_OfResource_EmptyURI_Errors(t *testing.T) {
	s := NewJobStep().Targeting(URI(""))
	if s.Err() == nil {
		t.Error("expected Err() != nil for empty resource URI")
	}
}

func TestJob_AddStep_DrainErrors(t *testing.T) {
	step := NewJobStep().Targeting(URI("")) // adds error to step
	j := NewJob().WithSteps(step)
	if j.Err() == nil {
		t.Error("expected step errors to be drained into job")
	}
}

func TestJob_AddStep_Nil(t *testing.T) {
	j := NewJob().WithSteps(nil)
	if j.Err() != nil {
		t.Errorf("WithSteps(nil) should not error: %v", j.Err())
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestJob_ToRequest_OneShot(t *testing.T) {
	ts := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	j := NewJob().
		InProject(URI("/projects/p")).
		Named("one-shot-job").
		InRegion(RegionITBGBergamo).
		Enabled().
		OneShotAt(ts).
		WithSteps(NewJobStep().
			Named("step-1").
			Targeting(URI("/projects/p/providers/Aruba.Compute/cloudServers/s-1")).
			WithAction("/projects/p/providers/Aruba.Compute/cloudServers/s-1/actions/start").
			WithVerb(HTTPVerbPOST))

	req := j.RawRequest()
	if req.Metadata.Name != "one-shot-job" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if req.Properties.JobType != JobTypeOneShot {
		t.Errorf("JobType = %q", req.Properties.JobType)
	}
	if req.Properties.ScheduleAt == nil || *req.Properties.ScheduleAt != "2026-05-01T12:00:00Z" {
		t.Errorf("ScheduleAt = %v", req.Properties.ScheduleAt)
	}
	if req.Properties.Enabled == nil || !*req.Properties.Enabled {
		t.Error("Enabled should be true")
	}
	if len(req.Properties.Steps) != 1 {
		t.Fatalf("Steps len = %d", len(req.Properties.Steps))
	}
	if req.Properties.Steps[0].HttpVerb != HTTPVerbPOST {
		t.Errorf("Step.HttpVerb = %q", req.Properties.Steps[0].HttpVerb)
	}
}

func TestJob_ToRequest_Recurring(t *testing.T) {
	until := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	j := NewJob().
		InProject(URI("/projects/p")).
		Named("cron-job").
		WithCron("0 8 * * 1-5").
		RecurringUntil(until)

	req := j.RawRequest()
	if req.Properties.JobType != JobTypeRecurring {
		t.Errorf("JobType = %q", req.Properties.JobType)
	}
	if req.Properties.Cron == nil || *req.Properties.Cron != "0 8 * * 1-5" {
		t.Errorf("Cron = %v", req.Properties.Cron)
	}
	if req.Properties.ExecuteUntil == nil || *req.Properties.ExecuteUntil != "2026-12-31T00:00:00Z" {
		t.Errorf("ExecuteUntil = %v", req.Properties.ExecuteUntil)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func jobTestResponse(name string) *types.JobResponse {
	id := "job-1"
	uri := "/projects/p/providers/Aruba.Schedule/jobs/job-1"
	state := types.State("Active")
	schedAt := "2026-05-01T12:00:00Z"
	cronExpr := "0 8 * * 1-5"
	execUntil := "2026-12-31T00:00:00Z"
	jt := JobTypeOneShot
	resURI := "/projects/p/providers/Aruba.Compute/cloudServers/s-1"
	actionURI := "/projects/p/providers/Aruba.Compute/cloudServers/s-1/actions/start"
	verb := string(HTTPVerbPOST)
	stepName := "step-1"
	return &types.JobResponse{
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
		Properties: types.JobPropertiesResponse{
			Enabled:      true,
			JobType:      jt,
			ScheduleAt:   &schedAt,
			ExecuteUntil: &execUntil,
			Cron:         &cronExpr,
			Steps: []types.JobStepResponse{
				{
					Name:        &stepName,
					ResourceURI: &resURI,
					ActionURI:   &actionURI,
					HttpVerb:    &verb,
				},
			},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}
}

func TestJob_FromResponseHydration(t *testing.T) {
	j := &Job{}
	j.fromResponse(jobTestResponse("my-job"))

	if j.Name() != "my-job" {
		t.Errorf("Name() = %q", j.Name())
	}
	if j.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", j.ProjectID())
	}
	if j.ID() != "job-1" {
		t.Errorf("ID() = %q", j.ID())
	}
	if j.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", j.Region())
	}
	if !j.IsEnabled() {
		t.Error("Enabled() should be true")
	}
	if j.JobType() != JobTypeOneShot {
		t.Errorf("JobType() = %q", j.JobType())
	}
	if j.scheduleAt == nil || *j.scheduleAt != "2026-05-01T12:00:00Z" {
		t.Errorf("scheduleAt = %v", j.scheduleAt)
	}
	if j.cron == nil || *j.cron != "0 8 * * 1-5" {
		t.Errorf("cron = %v", j.cron)
	}
	if j.executeUntil == nil || *j.executeUntil != "2026-12-31T00:00:00Z" {
		t.Errorf("executeUntil = %v", j.executeUntil)
	}
	if len(j.steps) != 1 {
		t.Fatalf("steps len = %d", len(j.steps))
	}
	step := j.steps[0]
	if step.name == nil || *step.name != "step-1" {
		t.Errorf("steps[0].name = %v", step.name)
	}
	if step.resourceURI == nil || *step.resourceURI != "/projects/p/providers/Aruba.Compute/cloudServers/s-1" {
		t.Errorf("steps[0].resourceURI = %v", step.resourceURI)
	}
	if step.httpVerb == nil || *step.httpVerb != HTTPVerbPOST {
		t.Errorf("steps[0].httpVerb = %v", step.httpVerb)
	}
}

func TestJob_FromResponse_BackfillsProjectID_FromURI(t *testing.T) {
	id := "job-x"
	uri := "/projects/proj-abc/providers/Aruba.Schedule/jobs/job-x"
	resp := &types.JobResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  &id,
			URI: &uri,
		},
		Properties: types.JobPropertiesResponse{},
	}
	j := &Job{}
	j.fromResponse(resp)
	if j.ProjectID() != "proj-abc" {
		t.Errorf("ProjectID() backfilled from URI = %q", j.ProjectID())
	}
}

func TestJob_FromResponse_Nil(t *testing.T) {
	j := &Job{}
	j.fromResponse(nil) // must not panic
}

// --------------------------------------------------------------------------
// jobIDsFromRef
// --------------------------------------------------------------------------

func TestJobIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/proj-1/providers/Aruba.Schedule/jobs/job-42")
	pid, jid, err := jobIDsFromRef(ref)
	if err != nil {
		t.Fatalf("jobIDsFromRef error: %v", err)
	}
	if pid != "proj-1" {
		t.Errorf("projectID = %q", pid)
	}
	if jid != "job-42" {
		t.Errorf("jobID = %q", jid)
	}
}

func TestJobIDsFromRef_TypedRef(t *testing.T) {
	j := &Job{}
	j.fromResponse(jobTestResponse("j"))
	pid, jid, err := jobIDsFromRef(j)
	if err != nil {
		t.Fatalf("jobIDsFromRef error: %v", err)
	}
	if pid != "p" {
		t.Errorf("projectID = %q", pid)
	}
	if jid != "job-1" {
		t.Errorf("jobID = %q", jid)
	}
}

func TestJobIDsFromRef_BadURI_NoJob(t *testing.T) {
	_, _, err := jobIDsFromRef(URI("/projects/p/providers/Aruba.Schedule"))
	if err == nil {
		t.Error("expected error when job segment missing")
	}
}

func TestJobIDsFromRef_BadURI_NoProject(t *testing.T) {
	_, _, err := jobIDsFromRef(URI("/providers/Aruba.Schedule/jobs/j"))
	if err == nil {
		t.Error("expected error when project segment missing")
	}
}

// --------------------------------------------------------------------------
// HTTP-mock adapter helper
// --------------------------------------------------------------------------

func buildJobsTestAdapter(t *testing.T, handler http.HandlerFunc) *jobsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newJobsClientAdapter(testutil.NewClient(t, server.URL))
}

const jobSuccessBody = `{` +
	`"metadata":{"id":"job-1","name":"my-job","uri":"/projects/p/providers/Aruba.Schedule/jobs/job-1","project":{"id":"p"}},` +
	`"properties":{"enabled":true,"scheduleJobType":"OneShot","scheduleAt":"2026-05-01T12:00:00Z"},` +
	`"status":{"state":"Active"}}`

// --------------------------------------------------------------------------
// Create adapter tests
// --------------------------------------------------------------------------

func TestJobsClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.JobRequest
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "jobs") {
			t.Errorf("path %q should contain 'jobs'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, jobSuccessBody)
	})

	ts := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	j := NewJob().
		InProject(URI("/projects/p")).
		Named("my-job").
		InRegion(RegionITBGBergamo).
		Enabled().
		OneShotAt(ts)

	result, err := adapter.Create(context.Background(), j)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "job-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-job" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Metadata.Name != "my-job" {
		t.Errorf("request Metadata.Name = %q", gotBody.Metadata.Name)
	}
	if gotBody.Properties.JobType != JobTypeOneShot {
		t.Errorf("request JobType = %q", gotBody.Properties.JobType)
	}
}

func TestJobsClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})
	_, err := adapter.Create(context.Background(), NewJob().
		Named("x"))
	if err == nil {
		t.Fatal("expected error when Job has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestJobsClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" — triggers MetadataValidationError from low-level Validate()
		fmt.Fprint(w, `{"metadata":{"name":"j","uri":"/projects/p/providers/Aruba.Schedule/jobs/x"},"properties":{},"status":{}}`)
	})

	j := NewJob().InProject(URI("/projects/p")).
		Named("j")
	result, err := adapter.Create(context.Background(), j)
	if err == nil {
		t.Fatal("expected MetadataValidationError, got nil")
	}
	var mvErr *types.MetadataValidationError
	if !errors.As(err, &mvErr) {
		t.Fatalf("expected *types.MetadataValidationError, got %T: %v", err, err)
	}
	if result == nil {
		t.Error("result wrapper should not be nil even on error")
	}
}

func TestJobsClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"bad request"}`)
	})
	_, err := adapter.Create(context.Background(), NewJob().InProject(URI("/projects/p")).
		Named("j"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Update adapter tests
// --------------------------------------------------------------------------

func TestJobsClientAdapter_Update_Success(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, jobSuccessBody)
	})

	// Load from response, then only update non-schedule fields to avoid mode conflict.
	j := &Job{}
	j.fromResponse(jobTestResponse("my-job"))
	j.Named("my-job-updated").Disabled()

	result, err := adapter.Update(context.Background(), j)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if result.ID() != "job-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestJobsClientAdapter_Update_NoID(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	j := NewJob().InProject(URI("/projects/p")).
		Named("x")
	_, err := adapter.Update(context.Background(), j)
	if err == nil {
		t.Fatal("expected error when Job has no ID")
	}
}

func TestJobsClientAdapter_Update_NoProject(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	j := &Job{}
	id := "job-1"
	j.setMeta(&types.ResourceMetadataResponse{ID: &id})
	_, err := adapter.Update(context.Background(), j)
	if err == nil {
		t.Fatal("expected error when Job has no project")
	}
}

func TestJobsClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"not found"}`)
	})
	j := &Job{}
	j.fromResponse(jobTestResponse("j"))
	_, err := adapter.Update(context.Background(), j)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
}

// --------------------------------------------------------------------------
// Get adapter tests
// --------------------------------------------------------------------------

func TestJobsClientAdapter_Get_URIRef(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, jobSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Schedule/jobs/job-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "job-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestJobsClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, jobSuccessBody)
	})

	j := &Job{}
	j.fromResponse(jobTestResponse("j"))
	result, err := adapter.Get(context.Background(), j)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "job-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

// --------------------------------------------------------------------------
// Delete adapter tests
// --------------------------------------------------------------------------

func TestJobsClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	j := &Job{}
	j.fromResponse(jobTestResponse("j"))
	if err := adapter.Delete(context.Background(), j); err != nil {
		t.Errorf("Delete error: %v", err)
	}
}

func TestJobsClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"not found"}`)
	})
	j := &Job{}
	j.fromResponse(jobTestResponse("j"))
	err := adapter.Delete(context.Background(), j)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
}

// --------------------------------------------------------------------------
// List adapter tests
// --------------------------------------------------------------------------

const jobListBody = `{"total":2,"values":[` +
	`{"metadata":{"id":"job-1","name":"job-one","uri":"/projects/p/providers/Aruba.Schedule/jobs/job-1","project":{"id":"p"}},"properties":{},"status":{}},` +
	`{"metadata":{"id":"job-2","name":"job-two","uri":"/projects/p/providers/Aruba.Schedule/jobs/job-2","project":{"id":"p"}},"properties":{},"status":{}}` +
	`]}`

func TestJobsClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, jobListBody)
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
	if items[0].Name() != "job-one" {
		t.Errorf("items[0].Name() = %q", items[0].Name())
	}
	if items[1].Name() != "job-two" {
		t.Errorf("items[1].Name() = %q", items[1].Name())
	}
}

// --------------------------------------------------------------------------
// Setter delegation (RemoveTag, ReplaceTags, InRegion)
// --------------------------------------------------------------------------

func TestJob_SetterDelegation(t *testing.T) {
	j := NewJob().
		Tagged("a").
		Tagged("b").
		Tagged("c").
		Untagged("b").
		RetaggedAs("x", "y").
		InRegion(RegionITBGBergamo)

	tags := j.Tags()
	if len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("Tags() after ReplaceTags = %v", tags)
	}
	if j.Region() != RegionITBGBergamo {
		t.Errorf("Region() after InRegion = %q", j.Region())
	}
}

// --------------------------------------------------------------------------
// URI and Raw accessors after hydration
// --------------------------------------------------------------------------

func TestJob_URI_AfterHydration(t *testing.T) {
	j := &Job{}
	j.fromResponse(jobTestResponse("u"))
	if j.URI() == "" {
		t.Error("URI() should not be empty after hydration")
	}
	if j.Raw() == nil {
		t.Error("Raw() should not be nil after hydration")
	}
}

func TestJob_RawRequest_NoHydration(t *testing.T) {
	_ = NewJob().RawRequest() // must not panic
}

// --------------------------------------------------------------------------
// Accessors at zero value (no hydration)
// --------------------------------------------------------------------------

func TestJob_Accessors_ZeroValue(t *testing.T) {
	j := NewJob()
	if j.URI() != "" {
		t.Errorf("URI() zero = %q", j.URI())
	}
	if j.Raw() != nil {
		t.Errorf("Raw() zero = %v", j.Raw())
	}
	if j.IsEnabled() != false {
		t.Error("Enabled() zero should be false")
	}
	if j.JobType() != "" {
		t.Errorf("JobType() zero = %q", j.JobType())
	}
	if j.Cron() != "" {
		t.Errorf("Cron() zero = %q", j.Cron())
	}
}

// --------------------------------------------------------------------------
// Enabled, JobType, Cron — request-side fallbacks (response is nil)
// --------------------------------------------------------------------------

func TestJob_Enabled_RequestFallback(t *testing.T) {
	j := NewJob().Enabled()
	if !j.IsEnabled() {
		t.Error("Enabled() request fallback should be true")
	}
}

func TestJob_JobType_RequestFallback(t *testing.T) {
	ts := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	j := NewJob().OneShotAt(ts)
	if j.JobType() != JobTypeOneShot {
		t.Errorf("JobType() request fallback = %q", j.JobType())
	}
}

func TestJob_Cron_RequestFallback(t *testing.T) {
	j := NewJob().WithCron("0 * * * *")
	if j.Cron() != "0 * * * *" {
		t.Errorf("Cron() request fallback = %q", j.Cron())
	}
}

// --------------------------------------------------------------------------
// Get adapter — BadRef and NonTwoXX
// --------------------------------------------------------------------------

func TestJobsClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.Get(context.Background(), URI("/providers/Aruba.Schedule"))
	if err == nil {
		t.Fatal("expected error for unparseable ref")
	}
}

func TestJobsClientAdapter_Get_NetworkError(t *testing.T) {
	adapter := &jobsClientAdapter{low: schedule.NewJobsClientImpl(testutil.NewBrokenClient(t, "http://localhost:9"))}
	_, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Schedule/jobs/job-1"))
	if err == nil {
		t.Fatal("expected network error from broken client")
	}
}

func TestJobsClientAdapter_Get_ProjectIDFallback(t *testing.T) {
	// Server returns a response with no project metadata, so fromResponse won't set
	// projectID — the "if out.projectID == """ guard must restore it from the ref.
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		id := "job-99"
		fmt.Fprintf(w, `{"metadata":{"id":%q},"properties":{},"status":{}}`, id)
	})
	ref := URI("/projects/p/providers/Aruba.Schedule/jobs/job-99")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() after fallback = %q", result.ProjectID())
	}
}

func TestJobsClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"not found"}`)
	})
	ref := URI("/projects/p/providers/Aruba.Schedule/jobs/job-1")
	_, err := adapter.Get(context.Background(), ref)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Delete adapter — BadRef
// --------------------------------------------------------------------------

func TestJobsClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := adapter.Delete(context.Background(), URI("/providers/Aruba.Schedule"))
	if err == nil {
		t.Fatal("expected error for unparseable ref")
	}
}

func TestJobsClientAdapter_Delete_NetworkError(t *testing.T) {
	adapter := &jobsClientAdapter{low: schedule.NewJobsClientImpl(testutil.NewBrokenClient(t, "http://localhost:9"))}
	j := &Job{}
	j.fromResponse(jobTestResponse("j"))
	err := adapter.Delete(context.Background(), j)
	if err == nil {
		t.Fatal("expected network error from broken client")
	}
}

// --------------------------------------------------------------------------
// List adapter — NonTwoXX and bad parent ref
// --------------------------------------------------------------------------

func TestJobsClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"forbidden"}`)
	})
	_, err := adapter.List(context.Background(), URI("/projects/p"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

func TestJobsClientAdapter_List_BadParentRef(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/no-project-here"))
	if err == nil {
		t.Fatal("expected error for parent ref with no project")
	}
}

// --------------------------------------------------------------------------
// newJobsClientAdapter nil-rest branch
// --------------------------------------------------------------------------

func TestNewJobsClientAdapter_NilRest(t *testing.T) {
	a := newJobsClientAdapter(nil)
	if a == nil {
		t.Fatal("expected non-nil adapter")
	}
}

// --------------------------------------------------------------------------
// Update — Err() pre-condition
// --------------------------------------------------------------------------

func TestJobsClientAdapter_Update_ErrSet(t *testing.T) {
	adapter := buildJobsTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	j := NewJob().InProject(URI("not-a-project-uri")) // sets Err()
	_, err := adapter.Update(context.Background(), j)
	if err == nil {
		t.Fatal("expected error when Job has Err() set")
	}
}

// --------------------------------------------------------------------------
// Cron — response not nil but Cron field is nil
// --------------------------------------------------------------------------

func TestJob_Cron_ResponseNilCron(t *testing.T) {
	j := &Job{}
	id := "job-x"
	uri := "/projects/p/providers/Aruba.Schedule/jobs/job-x"
	jt := JobTypeOneShot
	resp := &types.JobResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, URI: &uri},
		Properties: types.JobPropertiesResponse{
			Enabled: true,
			JobType: jt,
			Cron:    nil, // explicitly nil
		},
	}
	j.fromResponse(resp)
	// Cron() should fall through to the cron field (also nil), returning ""
	if j.Cron() != "" {
		t.Errorf("Cron() with nil response.Cron = %q", j.Cron())
	}
}

// --------------------------------------------------------------------------
// Cron — response not nil and Cron field not nil (response-branch)
// --------------------------------------------------------------------------

func TestJob_Cron_ResponseBranch(t *testing.T) {
	cronExpr := "0 12 * * 1"
	j := &Job{}
	// Directly set the response to exercise the response-first branch of Cron().
	j.response = &types.JobResponse{
		Properties: types.JobPropertiesResponse{
			Cron: &cronExpr,
		},
	}
	if j.Cron() != cronExpr {
		t.Errorf("Cron() response branch = %q, want %q", j.Cron(), cronExpr)
	}
}

// --------------------------------------------------------------------------
// Reflective guard
// --------------------------------------------------------------------------

func TestJobsClient_HasUpdateMethod(t *testing.T) {
	iface := reflect.TypeOf((*JobsClient)(nil)).Elem()
	if _, ok := iface.MethodByName("Update"); !ok {
		t.Error("JobsClient interface is missing the Update method")
	}
}

func TestJob_FromResponse_SetsStatus(t *testing.T) {
	j := &Job{}
	state := types.State("Active")
	j.fromResponse(&types.JobResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if j.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", j.State())
	}
}

func TestJobsClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, jobSuccessBody)
	})
	adapter := newJobsClientAdapter(testutil.NewClient(t, server.URL))
	j, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Schedule/jobs/job-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&j.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned Job")
	}
}

// --------------------------------------------------------------------------
// ScheduleAt getter
// --------------------------------------------------------------------------

func TestJob_ScheduleAt_NilResponse(t *testing.T) {
	j := &Job{}
	if got := j.ScheduleAt(); !got.IsZero() {
		t.Errorf("ScheduleAt() = %v, want zero", got)
	}
}

func TestJob_ScheduleAt_NoScheduleAt(t *testing.T) {
	j := &Job{}
	j.fromResponse(&types.JobResponse{})
	if got := j.ScheduleAt(); !got.IsZero() {
		t.Errorf("ScheduleAt() = %v, want zero", got)
	}
}

func TestJob_ScheduleAt_FromResponse(t *testing.T) {
	ts := "2025-06-01T10:00:00Z"
	j := &Job{}
	j.fromResponse(&types.JobResponse{
		Properties: types.JobPropertiesResponse{ScheduleAt: &ts},
	})
	want := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	if got := j.ScheduleAt(); !got.Equal(want) {
		t.Errorf("ScheduleAt() = %v, want %v", got, want)
	}
}

func TestJob_ScheduleAt_LocalSetter(t *testing.T) {
	want := time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)
	j := NewJob().OneShotAt(want)
	if got := j.ScheduleAt(); !got.Equal(want) {
		t.Errorf("ScheduleAt() = %v, want %v", got, want)
	}
}

// --------------------------------------------------------------------------
// ExecuteUntil getter
// --------------------------------------------------------------------------

func TestJob_ExecuteUntil_NilResponse(t *testing.T) {
	j := &Job{}
	if got := j.ExecuteUntil(); !got.IsZero() {
		t.Errorf("ExecuteUntil() = %v, want zero", got)
	}
}

func TestJob_ExecuteUntil_FromResponse(t *testing.T) {
	ts := "2025-12-31T23:59:59Z"
	j := &Job{}
	j.fromResponse(&types.JobResponse{
		Properties: types.JobPropertiesResponse{ExecuteUntil: &ts},
	})
	want := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	if got := j.ExecuteUntil(); !got.Equal(want) {
		t.Errorf("ExecuteUntil() = %v, want %v", got, want)
	}
}

func TestJob_ExecuteUntil_LocalSetter(t *testing.T) {
	want := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	j := NewJob().WithCron("0 * * * *").RecurringUntil(want)
	if got := j.ExecuteUntil(); !got.Equal(want) {
		t.Errorf("ExecuteUntil() = %v, want %v", got, want)
	}
}

// --------------------------------------------------------------------------
// NextExecutionAt getter
// --------------------------------------------------------------------------

func TestJob_NextExecutionAt_NilResponse(t *testing.T) {
	j := &Job{}
	if got := j.NextExecutionAt(); !got.IsZero() {
		t.Errorf("NextExecutionAt() = %v, want zero", got)
	}
}

func TestJob_NextExecutionAt_NoNextExecution(t *testing.T) {
	j := &Job{}
	j.fromResponse(&types.JobResponse{})
	if got := j.NextExecutionAt(); !got.IsZero() {
		t.Errorf("NextExecutionAt() = %v, want zero", got)
	}
}

func TestJob_NextExecutionAt_FromResponse(t *testing.T) {
	ts := "2025-06-15T08:00:00Z"
	j := &Job{}
	j.fromResponse(&types.JobResponse{
		Properties: types.JobPropertiesResponse{NextExecution: &ts},
	})
	want := time.Date(2025, 6, 15, 8, 0, 0, 0, time.UTC)
	if got := j.NextExecutionAt(); !got.Equal(want) {
		t.Errorf("NextExecutionAt() = %v, want %v", got, want)
	}
}

// --------------------------------------------------------------------------
// Steps getter
// --------------------------------------------------------------------------

func TestJob_Steps_NilWhenEmpty(t *testing.T) {
	j := &Job{}
	if steps := j.Steps(); steps != nil {
		t.Errorf("Steps() = %v, want nil", steps)
	}
}

func TestJob_Steps_ReturnsConfigured(t *testing.T) {
	step := NewJobStep().Named("step-1").WithAction("/actions/start").WithVerb(HTTPVerbPOST)
	j := NewJob().WithSteps(step)
	steps := j.Steps()
	if len(steps) != 1 {
		t.Fatalf("Steps() len = %d, want 1", len(steps))
	}
}

func TestJob_Steps_FromResponse(t *testing.T) {
	stepName := "step-from-resp"
	j := &Job{}
	j.fromResponse(&types.JobResponse{
		Properties: types.JobPropertiesResponse{
			Steps: []types.JobStepResponse{
				{Name: &stepName},
			},
		},
	})
	steps := j.Steps()
	if len(steps) != 1 {
		t.Fatalf("Steps() len = %d, want 1", len(steps))
	}
}
