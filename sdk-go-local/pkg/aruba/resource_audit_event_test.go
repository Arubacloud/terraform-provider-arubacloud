package aruba

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/internal/clients/audit"
	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Test helpers
// --------------------------------------------------------------------------

func buildAuditEventTestAdapter(t *testing.T, handler http.HandlerFunc) *auditEventsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return &auditEventsClientAdapter{low: audit.NewEventsClientImpl(testutil.NewClient(t, server.URL))}
}

// auditTestTimestamp is a fixed time used across hydration tests.
var auditTestTimestamp = time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

// auditMakeFullEvent returns a fully-populated types.AuditEventResponse for hydration tests.
func auditMakeFullEvent() types.AuditEventResponse {
	catID := "cat-42"
	typID := "typ-99"
	title := "Resource Created"
	regionName := "IT-MIL"
	az := "az1"
	subStatusVal := "sub-ok"
	subStatusDesc := "sub-description"
	return types.AuditEventResponse{
		SeverityLevel: "Info",
		LogFormat:     types.EventLogFormatVersionResponse{Version: "1.0"},
		Timestamp:     auditTestTimestamp,
		Operation:     types.EventOperationResponse{ID: "op-1", Value: strPtr("Create")},
		Event:         types.EventInfoResponse{ID: "evt-1", Value: strPtr("evt-value"), Type: "write"},
		Category:      types.EventCategoryResponse{Value: "resource", Description: strPtr("Resource category")},
		Region:        &types.EventRegionInfoResponse{Name: &regionName, AvailabilityZone: &az},
		Origin:        "portal",
		Channel:       "web",
		Status:        types.EventStatusResponse{Value: "Success", Description: strPtr("ok"), Code: int32Ptr(200)},
		SubStatus:     &types.EventSubStatusResponse{Value: &subStatusVal, Description: &subStatusDesc},
		Identity: types.EventIdentityResponse{
			Caller: types.EventCallerResponse{
				Subject:  "user-sub",
				Username: strPtr("alice"),
				Company:  strPtr("ArubaCloud"),
				TenantID: strPtr("tenant-1"),
			},
		},
		Properties: map[string]interface{}{"key": "val"},
		Actions:    []types.EventActionResponse{{Key: strPtr("action-1")}},
		CategoryID: &catID,
		TypologyID: &typID,
		Title:      &title,
	}
}

func int32Ptr(n int32) *int32 { return &n }

// --------------------------------------------------------------------------
// Hydration & accessors (no HTTP needed)
// --------------------------------------------------------------------------

func TestAuditEvent_FromResponse_FullyPopulated(t *testing.T) {
	wire := auditMakeFullEvent()
	ev := &AuditEvent{}
	ev.fromResponse(&wire)

	if ev.ID() != "evt-1" {
		t.Errorf("ID() = %q", ev.ID())
	}
	if ev.URI() != "" {
		t.Errorf("URI() = %q, want empty", ev.URI())
	}
	if !ev.CreatedAt().Equal(auditTestTimestamp) {
		t.Errorf("CreatedAt() = %v", ev.CreatedAt())
	}
	if ev.SeverityLevel() != "Info" {
		t.Errorf("SeverityLevel() = %q", ev.SeverityLevel())
	}
	if ev.Origin() != "portal" {
		t.Errorf("Origin() = %q", ev.Origin())
	}
	if ev.Channel() != "web" {
		t.Errorf("Channel() = %q", ev.Channel())
	}
	if ev.LogFormat().Version != "1.0" {
		t.Errorf("LogFormat().Version = %q", ev.LogFormat().Version)
	}
	if ev.Operation().ID != "op-1" {
		t.Errorf("Operation().ID = %q", ev.Operation().ID)
	}
	if ev.Event().ID != "evt-1" {
		t.Errorf("Event().ID = %q", ev.Event().ID)
	}
	if ev.Event().Type != "write" {
		t.Errorf("Event().Type = %q", ev.Event().Type)
	}
	if ev.Category().Value != "resource" {
		t.Errorf("Category().Value = %q", ev.Category().Value)
	}
	if ev.Region() != "IT-MIL" {
		t.Errorf("Region() = %q, want IT-MIL", ev.Region())
	}
	if ev.Zone() != "az1" {
		t.Errorf("Zone() = %q, want az1", ev.Zone())
	}
	if ev.Status().Value != "Success" {
		t.Errorf("Status().Value = %q", ev.Status().Value)
	}
	if ev.SubStatus() == nil || *ev.SubStatus().Value != "sub-ok" {
		t.Error("SubStatus() unexpected")
	}
	if ev.Identity().Caller.Subject != "user-sub" {
		t.Errorf("Identity().Caller.Subject = %q", ev.Identity().Caller.Subject)
	}
	if ev.Properties()["key"] != "val" {
		t.Error("Properties() unexpected")
	}
	if len(ev.Actions()) != 1 {
		t.Errorf("Actions() len = %d", len(ev.Actions()))
	}
	if ev.CategoryID() != "cat-42" {
		t.Errorf("CategoryID() = %q", ev.CategoryID())
	}
	if ev.TypologyID() != "typ-99" {
		t.Errorf("TypologyID() = %q", ev.TypologyID())
	}
	if ev.Title() != "Resource Created" {
		t.Errorf("Title() = %q", ev.Title())
	}
	if ev.Raw() != &wire {
		t.Error("Raw() does not point to original wire struct")
	}
}

func TestAuditEvent_FromResponse_Nil(t *testing.T) {
	ev := &AuditEvent{}
	ev.fromResponse(nil)
	if ev.response != nil {
		t.Error("fromResponse(nil) should not set response")
	}
}

func TestAuditEvent_Accessors_ZeroValue(t *testing.T) {
	ev := &AuditEvent{}
	if ev.ID() != "" {
		t.Errorf("ID() on zero = %q", ev.ID())
	}
	if ev.URI() != "" {
		t.Errorf("URI() on zero = %q", ev.URI())
	}
	if !ev.CreatedAt().IsZero() {
		t.Error("CreatedAt() on zero should be zero time")
	}
	if ev.SeverityLevel() != "" {
		t.Error("SeverityLevel() on zero should be empty")
	}
	if ev.Origin() != "" {
		t.Error("Origin() on zero should be empty")
	}
	if ev.Channel() != "" {
		t.Error("Channel() on zero should be empty")
	}
	if ev.LogFormat() != (types.EventLogFormatVersionResponse{}) {
		t.Error("LogFormat() on zero should be zero value")
	}
	if ev.Operation() != (types.EventOperationResponse{}) {
		t.Error("Operation() on zero should be zero value")
	}
	if ev.Event() != (types.EventInfoResponse{}) {
		t.Error("Event() on zero should be zero value")
	}
	if ev.Category() != (types.EventCategoryResponse{}) {
		t.Error("Category() on zero should be zero value")
	}
	if ev.Region() != "" {
		t.Errorf("Region() on zero should be empty, got %q", ev.Region())
	}
	if ev.Zone() != "" {
		t.Errorf("Zone() on zero should be empty, got %q", ev.Zone())
	}
	if ev.Status().Value != "" {
		t.Error("Status() on zero should have empty Value")
	}
	if ev.SubStatus() != nil {
		t.Error("SubStatus() on zero should be nil")
	}
	if ev.Identity().Caller.Subject != "" {
		t.Error("Identity() on zero should have empty Caller.Subject")
	}
	if ev.Properties() != nil {
		t.Error("Properties() on zero should be nil")
	}
	if ev.Actions() != nil {
		t.Error("Actions() on zero should be nil")
	}
	if ev.CategoryID() != "" {
		t.Error("CategoryID() on zero should be empty")
	}
	if ev.TypologyID() != "" {
		t.Error("TypologyID() on zero should be empty")
	}
	if ev.Title() != "" {
		t.Error("Title() on zero should be empty")
	}
	if ev.Raw() != nil {
		t.Error("Raw() on zero should be nil")
	}
	if ev.ProjectID() != "" {
		t.Error("ProjectID() on zero should be empty")
	}
}

func TestAuditEvent_ID_FromEvent(t *testing.T) {
	ev := &AuditEvent{}
	ev.fromResponse(&types.AuditEventResponse{Event: types.EventInfoResponse{ID: "evt-123"}})
	if ev.ID() != "evt-123" {
		t.Errorf("ID() = %q, want evt-123", ev.ID())
	}

	ev2 := &AuditEvent{}
	ev2.fromResponse(&types.AuditEventResponse{}) // Event.ID is ""
	if ev2.ID() != "" {
		t.Errorf("ID() = %q, want empty", ev2.ID())
	}
}

func TestAuditEvent_URI_AlwaysEmpty(t *testing.T) {
	wire := auditMakeFullEvent()
	ev := &AuditEvent{}
	ev.fromResponse(&wire)
	ev.projectID = "some-project"
	if ev.URI() != "" {
		t.Errorf("URI() = %q; audit events have no individual fetch URI", ev.URI())
	}
}

func TestAuditEvent_CreatedAt_FromTimestamp(t *testing.T) {
	ts := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	ev := &AuditEvent{}
	ev.fromResponse(&types.AuditEventResponse{Timestamp: ts})
	if !ev.CreatedAt().Equal(ts) {
		t.Errorf("CreatedAt() = %v, want %v", ev.CreatedAt(), ts)
	}
}

func TestAuditEvent_OptionalStringPointers_NilAndPresent(t *testing.T) {
	ev := &AuditEvent{}
	ev.fromResponse(&types.AuditEventResponse{})

	if ev.CategoryID() != "" {
		t.Errorf("CategoryID() nil ptr = %q", ev.CategoryID())
	}
	if ev.TypologyID() != "" {
		t.Errorf("TypologyID() nil ptr = %q", ev.TypologyID())
	}
	if ev.Title() != "" {
		t.Errorf("Title() nil ptr = %q", ev.Title())
	}

	catID := "cat-1"
	typID := "typ-2"
	title := "My Title"
	ev2 := &AuditEvent{}
	ev2.fromResponse(&types.AuditEventResponse{CategoryID: &catID, TypologyID: &typID, Title: &title})

	if ev2.CategoryID() != "cat-1" {
		t.Errorf("CategoryID() = %q", ev2.CategoryID())
	}
	if ev2.TypologyID() != "typ-2" {
		t.Errorf("TypologyID() = %q", ev2.TypologyID())
	}
	if ev2.Title() != "My Title" {
		t.Errorf("Title() = %q", ev2.Title())
	}
}

// --------------------------------------------------------------------------
// Adapter — happy paths
// --------------------------------------------------------------------------

const auditEventListBody = `{` +
	`"total":2,` +
	`"self":"/self","prev":"","next":"/next","first":"/first","last":"/last",` +
	`"values":[` +
	`{"severityLevel":"Info","event":{"id":"evt-1","type":"write"},"@timestamp":"2025-01-15T12:00:00Z","status":{"value":"Success"},"operation":{"id":"op-1"},"category":{"value":"res"},"identity":{"caller":{"subject":"s"}},"logFormat":{"version":"1.0"},"origin":"api","channel":"web"},` +
	`{"severityLevel":"Warning","event":{"id":"evt-2","type":"read"},"@timestamp":"2025-01-16T00:00:00Z","status":{"value":"Failed"},"operation":{"id":"op-2"},"category":{"value":"res"},"identity":{"caller":{"subject":"s"}},"logFormat":{"version":"1.0"},"origin":"cli","channel":"api"}` +
	`]}`

func TestAuditEventsClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildAuditEventTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, auditEventListBody)
	})

	list, err := adapter.List(context.Background(), URI("/projects/p-1"))
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if list.Total() != 2 {
		t.Errorf("Total() = %d, want 2", list.Total())
	}
	items := list.Items()
	if len(items) != 2 {
		t.Fatalf("len(Items()) = %d, want 2", len(items))
	}
	if items[0].ID() != "evt-1" {
		t.Errorf("items[0].ID() = %q", items[0].ID())
	}
	if items[0].SeverityLevel() != "Info" {
		t.Errorf("items[0].SeverityLevel() = %q", items[0].SeverityLevel())
	}
	if items[0].ProjectID() != "p-1" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
	if items[0].StatusCode() != http.StatusOK {
		t.Errorf("items[0].StatusCode() = %d", items[0].StatusCode())
	}
	if items[1].SeverityLevel() != "Warning" {
		t.Errorf("items[1].SeverityLevel() = %q", items[1].SeverityLevel())
	}
}

func TestAuditEventsClientAdapter_List_Empty(t *testing.T) {
	adapter := buildAuditEventTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":0,"values":[]}`)
	})

	list, err := adapter.List(context.Background(), URI("/projects/p-1"))
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if list.Total() != 0 {
		t.Errorf("Total() = %d, want 0", list.Total())
	}
	if len(list.Items()) != 0 {
		t.Errorf("len(Items()) = %d, want 0", len(list.Items()))
	}
}

// --------------------------------------------------------------------------
// Adapter — error paths
// --------------------------------------------------------------------------

func TestAuditEventsClientAdapter_List_BadProjectRef(t *testing.T) {
	callCount := 0
	adapter := buildAuditEventTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	_, err := adapter.List(context.Background(), URI("/garbage"))
	if err == nil {
		t.Fatal("expected error for bad project ref")
	}
	if callCount != 0 {
		t.Errorf("HTTP was called %d times; should not have been called", callCount)
	}
}

func TestAuditEventsClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildAuditEventTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := adapter.List(context.Background(), URI("/projects/p-1"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want 500", httpErr.StatusCode)
	}
}

func TestAuditEventsClientAdapter_List_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := &auditEventsClientAdapter{low: audit.NewEventsClientImpl(testutil.NewClient(t, server.URL))}
	_, err := adapter.List(context.Background(), URI("/projects/p-1"))
	if err == nil {
		t.Fatal("expected transport error from connection hijack")
	}
}

// --------------------------------------------------------------------------
// Constructor & contract
// --------------------------------------------------------------------------

func TestNewAuditEventsClientAdapter_Nil(t *testing.T) {
	a := newAuditEventsClientAdapter(nil)
	if a == nil {
		t.Fatal("newAuditEventsClientAdapter(nil) returned nil")
	}
	if a.low != nil {
		t.Error("expected low == nil when rest is nil")
	}
}

func TestNewAuditEventsClientAdapter_NonNil(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	client := testutil.NewClient(t, server.URL)
	a := newAuditEventsClientAdapter(client)
	if a == nil {
		t.Fatal("newAuditEventsClientAdapter(client) returned nil")
	}
	if a.low == nil {
		t.Error("expected low != nil when rest is non-nil")
	}
}

func TestEventsClient_OnlyHasList(t *testing.T) {
	iface := reflect.TypeOf((*EventsClient)(nil)).Elem()
	if iface.NumMethod() != 1 {
		t.Errorf("EventsClient has %d methods, want exactly 1 (List)", iface.NumMethod())
	}
	if iface.Method(0).Name != "List" {
		t.Errorf("EventsClient.Method(0) = %q, want List", iface.Method(0).Name)
	}
	for i := range iface.NumMethod() {
		name := iface.Method(i).Name
		if name == "Create" || name == "Update" || name == "Delete" || name == "Get" {
			t.Errorf("EventsClient must not expose %s — audit events are read-only (Family C)", name)
		}
	}
}

// --------------------------------------------------------------------------
// Refetch placeholder coverage
// --------------------------------------------------------------------------

// --------------------------------------------------------------------------
// EventTypeName getter
// --------------------------------------------------------------------------

func TestAuditEvent_EventTypeName_NilResponse(t *testing.T) {
	e := &AuditEvent{}
	if got := e.EventTypeName(); got != "" {
		t.Errorf("EventTypeName() = %q, want empty", got)
	}
}

func TestAuditEvent_EventTypeName_FromResponse(t *testing.T) {
	ev := auditMakeFullEvent()
	e := &AuditEvent{}
	e.fromResponse(&ev)
	if got := e.EventTypeName(); got != "write" {
		t.Errorf("EventTypeName() = %q, want write", got)
	}
}

// --------------------------------------------------------------------------
// IdentityName getter
// --------------------------------------------------------------------------

func TestAuditEvent_IdentityName_NilResponse(t *testing.T) {
	e := &AuditEvent{}
	if got := e.IdentityName(); got != "" {
		t.Errorf("IdentityName() = %q, want empty", got)
	}
}

func TestAuditEvent_IdentityName_NoUsername(t *testing.T) {
	ev := types.AuditEventResponse{Identity: types.EventIdentityResponse{Caller: types.EventCallerResponse{Subject: "sub"}}}
	e := &AuditEvent{}
	e.fromResponse(&ev)
	if got := e.IdentityName(); got != "" {
		t.Errorf("IdentityName() = %q, want empty when username nil", got)
	}
}

func TestAuditEvent_IdentityName_FromResponse(t *testing.T) {
	ev := auditMakeFullEvent()
	e := &AuditEvent{}
	e.fromResponse(&ev)
	if got := e.IdentityName(); got != "alice" {
		t.Errorf("IdentityName() = %q, want alice", got)
	}
}

// --------------------------------------------------------------------------
// OperationName getter
// --------------------------------------------------------------------------

func TestAuditEvent_OperationName_NilResponse(t *testing.T) {
	e := &AuditEvent{}
	if got := e.OperationName(); got != "" {
		t.Errorf("OperationName() = %q, want empty", got)
	}
}

func TestAuditEvent_OperationName_NoValue(t *testing.T) {
	ev := types.AuditEventResponse{Operation: types.EventOperationResponse{ID: "op-1"}}
	e := &AuditEvent{}
	e.fromResponse(&ev)
	if got := e.OperationName(); got != "" {
		t.Errorf("OperationName() = %q, want empty when Value nil", got)
	}
}

func TestAuditEvent_OperationName_FromResponse(t *testing.T) {
	ev := auditMakeFullEvent()
	e := &AuditEvent{}
	e.fromResponse(&ev)
	if got := e.OperationName(); got != "Create" {
		t.Errorf("OperationName() = %q, want Create", got)
	}
}

func TestAuditEventsClientAdapter_List_RefetchReturnsError(t *testing.T) {
	adapter := buildAuditEventTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Include a non-empty next URL so list.HasNext() is true.
		fmt.Fprint(w, `{"total":1,"next":"/next-page","values":[`+
			`{"event":{"id":"e1","type":"write"},"@timestamp":"2025-01-01T00:00:00Z","status":{"value":"ok"},"operation":{"id":"op1"},"category":{"value":"res"},"identity":{"caller":{"subject":"s"}},"logFormat":{"version":"1.0"},"origin":"api","channel":"web","severityLevel":"Info"}`+
			`]}`)
	})

	list, err := adapter.List(context.Background(), URI("/projects/p-1"))
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if !list.HasNext() {
		t.Skip("no next page in response — cannot test refetch")
	}
	_, err = list.Next(context.Background())
	if err == nil {
		t.Fatal("expected refetch error")
	}
}
