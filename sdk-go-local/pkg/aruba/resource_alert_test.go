package aruba

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/internal/clients/metric"
	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Test helpers
// --------------------------------------------------------------------------

func buildAlertTestAdapter(t *testing.T, handler http.HandlerFunc) *alertsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return &alertsClientAdapter{low: metric.NewAlertsClientImpl(testutil.NewClient(t, server.URL))}
}

var alertTestTime = time.Date(2025, 3, 10, 8, 0, 0, 0, time.UTC)

func alertMakeFullResponse() types.AlertResponse {
	return types.AlertResponse{
		ID:                   "alert-1",
		EventID:              "event-42",
		EventName:            "cpu.high",
		Username:             "alice",
		ServiceCategory:      "compute",
		ServiceTypology:      "vm",
		ResourceID:           "res-99",
		ServiceName:          "my-server",
		ResourceTypology:     "cloudserver",
		Metric:               "cpu_usage",
		LastReception:        alertTestTime,
		Rule:                 "cpu > 90%",
		Theshold:             90,
		UM:                   "%",
		Duration:             "5m",
		ThesholdExceedence:   "exceeded",
		Component:            "core",
		ClusterTypology:      "k8s",
		Cluster:              "prod-cluster",
		Clustername:          "prod",
		NodePool:             "default",
		SMS:                  true,
		Email:                true,
		Panel:                false,
		Hidden:               false,
		ExecutedAlertActions: []types.ExecutedAlertActionResponse{{ActionType: "email", Success: true}},
		Actions:              []types.AlertActionResponse{{Key: "retry"}},
	}
}

// --------------------------------------------------------------------------
// Hydration & accessors (no HTTP needed)
// --------------------------------------------------------------------------

func TestAlert_FromResponse_FullyPopulated(t *testing.T) {
	wire := alertMakeFullResponse()
	a := &Alert{}
	a.fromResponse(&wire)

	if a.ID() != "alert-1" {
		t.Errorf("ID() = %q", a.ID())
	}
	if a.URI() != "" {
		t.Errorf("URI() = %q, want empty", a.URI())
	}
	if a.EventID() != "event-42" {
		t.Errorf("EventID() = %q", a.EventID())
	}
	if a.EventName() != "cpu.high" {
		t.Errorf("EventName() = %q", a.EventName())
	}
	if a.Username() != "alice" {
		t.Errorf("Username() = %q", a.Username())
	}
	if a.ServiceCategory() != "compute" {
		t.Errorf("ServiceCategory() = %q", a.ServiceCategory())
	}
	if a.ServiceTypology() != "vm" {
		t.Errorf("ServiceTypology() = %q", a.ServiceTypology())
	}
	if a.ServiceName() != "my-server" {
		t.Errorf("ServiceName() = %q", a.ServiceName())
	}
	if a.ResourceID() != "res-99" {
		t.Errorf("ResourceID() = %q", a.ResourceID())
	}
	if a.ResourceTypology() != "cloudserver" {
		t.Errorf("ResourceTypology() = %q", a.ResourceTypology())
	}
	if a.Metric() != "cpu_usage" {
		t.Errorf("Metric() = %q", a.Metric())
	}
	if !a.LastReception().Equal(alertTestTime) {
		t.Errorf("LastReception() = %v", a.LastReception())
	}
	if a.Rule() != "cpu > 90%" {
		t.Errorf("Rule() = %q", a.Rule())
	}
	if a.Threshold() != 90 {
		t.Errorf("Threshold() = %d", a.Threshold())
	}
	if a.UM() != "%" {
		t.Errorf("UM() = %q", a.UM())
	}
	if a.Duration() != "5m" {
		t.Errorf("Duration() = %q", a.Duration())
	}
	if a.ThresholdExceedance() != "exceeded" {
		t.Errorf("ThresholdExceedance() = %q", a.ThresholdExceedance())
	}
	if a.Component() != "core" {
		t.Errorf("Component() = %q", a.Component())
	}
	if a.ClusterTypology() != "k8s" {
		t.Errorf("ClusterTypology() = %q", a.ClusterTypology())
	}
	if a.Cluster() != "prod-cluster" {
		t.Errorf("Cluster() = %q", a.Cluster())
	}
	if a.Clustername() != "prod" {
		t.Errorf("Clustername() = %q", a.Clustername())
	}
	if a.NodePool() != "default" {
		t.Errorf("NodePool() = %q", a.NodePool())
	}
	if !a.SMS() {
		t.Error("SMS() = false, want true")
	}
	if !a.Email() {
		t.Error("Email() = false, want true")
	}
	if a.Panel() {
		t.Error("Panel() = true, want false")
	}
	if a.Hidden() {
		t.Error("Hidden() = true, want false")
	}
	if len(a.ExecutedAlertActions()) != 1 {
		t.Errorf("ExecutedAlertActions() len = %d", len(a.ExecutedAlertActions()))
	}
	if len(a.Actions()) != 1 {
		t.Errorf("Actions() len = %d", len(a.Actions()))
	}
	if a.Raw() != &wire {
		t.Error("Raw() does not point to original wire struct")
	}
}

func TestAlert_FromResponse_Nil(t *testing.T) {
	a := &Alert{}
	a.fromResponse(nil)
	if a.response != nil {
		t.Error("fromResponse(nil) should not set response")
	}
}

func TestAlert_Accessors_ZeroValue(t *testing.T) {
	a := &Alert{}
	if a.ID() != "" {
		t.Errorf("ID() on zero = %q", a.ID())
	}
	if a.URI() != "" {
		t.Errorf("URI() on zero = %q", a.URI())
	}
	if a.EventID() != "" {
		t.Error("EventID() on zero should be empty")
	}
	if a.EventName() != "" {
		t.Error("EventName() on zero should be empty")
	}
	if a.Username() != "" {
		t.Error("Username() on zero should be empty")
	}
	if a.ServiceCategory() != "" {
		t.Error("ServiceCategory() on zero should be empty")
	}
	if a.ServiceTypology() != "" {
		t.Error("ServiceTypology() on zero should be empty")
	}
	if a.ServiceName() != "" {
		t.Error("ServiceName() on zero should be empty")
	}
	if a.ResourceID() != "" {
		t.Error("ResourceID() on zero should be empty")
	}
	if a.ResourceTypology() != "" {
		t.Error("ResourceTypology() on zero should be empty")
	}
	if a.Metric() != "" {
		t.Error("Metric() on zero should be empty")
	}
	if !a.LastReception().IsZero() {
		t.Error("LastReception() on zero should be zero time")
	}
	if a.Rule() != "" {
		t.Error("Rule() on zero should be empty")
	}
	if a.Threshold() != 0 {
		t.Errorf("Threshold() on zero = %d", a.Threshold())
	}
	if a.UM() != "" {
		t.Error("UM() on zero should be empty")
	}
	if a.Duration() != "" {
		t.Error("Duration() on zero should be empty")
	}
	if a.ThresholdExceedance() != "" {
		t.Error("ThresholdExceedance() on zero should be empty")
	}
	if a.Component() != "" {
		t.Error("Component() on zero should be empty")
	}
	if a.ClusterTypology() != "" {
		t.Error("ClusterTypology() on zero should be empty")
	}
	if a.Cluster() != "" {
		t.Error("Cluster() on zero should be empty")
	}
	if a.Clustername() != "" {
		t.Error("Clustername() on zero should be empty")
	}
	if a.NodePool() != "" {
		t.Error("NodePool() on zero should be empty")
	}
	if a.SMS() {
		t.Error("SMS() on zero should be false")
	}
	if a.Email() {
		t.Error("Email() on zero should be false")
	}
	if a.Panel() {
		t.Error("Panel() on zero should be false")
	}
	if a.Hidden() {
		t.Error("Hidden() on zero should be false")
	}
	if a.ExecutedAlertActions() != nil {
		t.Error("ExecutedAlertActions() on zero should be nil")
	}
	if a.Actions() != nil {
		t.Error("Actions() on zero should be nil")
	}
	if a.Raw() != nil {
		t.Error("Raw() on zero should be nil")
	}
	if a.ProjectID() != "" {
		t.Error("ProjectID() on zero should be empty")
	}
}

func TestAlert_ID_FromResponseID(t *testing.T) {
	a := &Alert{}
	a.fromResponse(&types.AlertResponse{ID: "alert-xyz"})
	if a.ID() != "alert-xyz" {
		t.Errorf("ID() = %q, want alert-xyz", a.ID())
	}
}

func TestAlert_URI_AlwaysEmpty(t *testing.T) {
	wire := alertMakeFullResponse()
	a := &Alert{}
	a.fromResponse(&wire)
	a.projectID = "p-1"
	if a.URI() != "" {
		t.Errorf("URI() = %q; alerts have no individual fetch URI", a.URI())
	}
}

// --------------------------------------------------------------------------
// Issue #211 acceptance criteria — typo normalization
// --------------------------------------------------------------------------

func TestAlert_TypoNormalization_Threshold(t *testing.T) {
	a := &Alert{}
	a.fromResponse(&types.AlertResponse{Theshold: 42})
	if a.Threshold() != 42 {
		t.Errorf("Threshold() = %d, want 42 (reads misspelled Theshold field)", a.Threshold())
	}
}

func TestAlert_TypoNormalization_ThresholdExceedance(t *testing.T) {
	a := &Alert{}
	a.fromResponse(&types.AlertResponse{ThesholdExceedence: "exceeded"})
	if a.ThresholdExceedance() != "exceeded" {
		t.Errorf("ThresholdExceedance() = %q, want %q (reads misspelled ThesholdExceedence field)", a.ThresholdExceedance(), "exceeded")
	}
}

// --------------------------------------------------------------------------
// Adapter — happy paths
// --------------------------------------------------------------------------

const alertListBody = `{` +
	`"total":2,` +
	`"self":"/self","prev":"","next":"/next","first":"/first","last":"/last",` +
	`"values":[` +
	`{"id":"alert-1","eventName":"cpu.high","theshold":80,"thesholdExceedence":"yes","severityLevel":"Warning"},` +
	`{"id":"alert-2","eventName":"mem.low","theshold":10,"thesholdExceedence":"no","severityLevel":"Info"}` +
	`]}`

func TestAlertsClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildAlertTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, alertListBody)
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
	if items[0].ID() != "alert-1" {
		t.Errorf("items[0].ID() = %q", items[0].ID())
	}
	if items[0].EventName() != "cpu.high" {
		t.Errorf("items[0].EventName() = %q", items[0].EventName())
	}
	if items[0].ProjectID() != "p-1" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
	if items[0].StatusCode() != http.StatusOK {
		t.Errorf("items[0].StatusCode() = %d", items[0].StatusCode())
	}
	if items[0].Threshold() != 80 {
		t.Errorf("items[0].Threshold() = %d, want 80", items[0].Threshold())
	}
	if items[0].ThresholdExceedance() != "yes" {
		t.Errorf("items[0].ThresholdExceedance() = %q, want yes", items[0].ThresholdExceedance())
	}
	if items[1].ID() != "alert-2" {
		t.Errorf("items[1].ID() = %q", items[1].ID())
	}
}

func TestAlertsClientAdapter_List_Empty(t *testing.T) {
	adapter := buildAlertTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestAlertsClientAdapter_List_BadProjectRef(t *testing.T) {
	callCount := 0
	adapter := buildAlertTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestAlertsClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildAlertTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestAlertsClientAdapter_List_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := &alertsClientAdapter{low: metric.NewAlertsClientImpl(testutil.NewClient(t, server.URL))}
	_, err := adapter.List(context.Background(), URI("/projects/p-1"))
	if err == nil {
		t.Fatal("expected transport error from connection hijack")
	}
}

// --------------------------------------------------------------------------
// Constructor & contract
// --------------------------------------------------------------------------

func TestNewAlertsClientAdapter_Nil(t *testing.T) {
	a := newAlertsClientAdapter(nil)
	if a == nil {
		t.Fatal("newAlertsClientAdapter(nil) returned nil")
	}
	if a.low != nil {
		t.Error("expected low == nil when rest is nil")
	}
}

func TestNewAlertsClientAdapter_NonNil(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	client := testutil.NewClient(t, server.URL)
	a := newAlertsClientAdapter(client)
	if a == nil {
		t.Fatal("newAlertsClientAdapter(client) returned nil")
	}
	if a.low == nil {
		t.Error("expected low != nil when rest is non-nil")
	}
}

func TestAlertsClient_OnlyHasList(t *testing.T) {
	iface := reflect.TypeOf((*AlertsClient)(nil)).Elem()
	if iface.NumMethod() != 1 {
		t.Errorf("AlertsClient has %d methods, want exactly 1 (List)", iface.NumMethod())
	}
	if iface.Method(0).Name != "List" {
		t.Errorf("AlertsClient.Method(0) = %q, want List", iface.Method(0).Name)
	}
	for i := range iface.NumMethod() {
		name := iface.Method(i).Name
		if name == "Create" || name == "Update" || name == "Delete" || name == "Get" {
			t.Errorf("AlertsClient must not expose %s — alerts are read-only (Family C)", name)
		}
	}
}

// --------------------------------------------------------------------------
// Refetch placeholder coverage
// --------------------------------------------------------------------------

func TestAlertsClientAdapter_List_RefetchReturnsError(t *testing.T) {
	adapter := buildAlertTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":1,"next":"/next-page","values":[`+
			`{"id":"alert-1","eventName":"cpu.high","theshold":70,"thesholdExceedence":"yes"}`+
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
