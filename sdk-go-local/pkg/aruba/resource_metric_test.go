package aruba

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/clients/metric"
	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Test helpers
// --------------------------------------------------------------------------

func buildMetricTestAdapter(t *testing.T, handler http.HandlerFunc) *metricsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return &metricsClientAdapter{low: metric.NewMetricsClientImpl(testutil.NewClient(t, server.URL))}
}

func metricMakeFullResponse() types.MetricResponse {
	return types.MetricResponse{
		ReferenceID:   "metric-ref-1",
		Name:          "cpu_usage",
		ReferenceName: "CPU Usage",
		Metadata: []types.MetricMetadataResponse{
			{Field: "unit", Value: "%"},
			{Field: "aggregation", Value: "avg"},
		},
		Data: []types.MetricDataResponse{
			{Time: "2025-03-10T08:00:00Z", Measure: "72.5"},
			{Time: "2025-03-10T08:05:00Z", Measure: "68.1"},
		},
	}
}

// --------------------------------------------------------------------------
// Hydration & accessors (no HTTP needed)
// --------------------------------------------------------------------------

func TestMetric_FromResponse_FullyPopulated(t *testing.T) {
	wire := metricMakeFullResponse()
	m := &Metric{}
	m.fromResponse(&wire)

	if m.ID() != "metric-ref-1" {
		t.Errorf("ID() = %q, want metric-ref-1", m.ID())
	}
	if m.URI() != "" {
		t.Errorf("URI() = %q, want empty", m.URI())
	}
	if m.ReferenceID() != "metric-ref-1" {
		t.Errorf("ReferenceID() = %q", m.ReferenceID())
	}
	if m.Name() != "cpu_usage" {
		t.Errorf("Name() = %q", m.Name())
	}
	if m.ReferenceName() != "CPU Usage" {
		t.Errorf("ReferenceName() = %q", m.ReferenceName())
	}
	if len(m.Metadata()) != 2 {
		t.Errorf("Metadata() len = %d, want 2", len(m.Metadata()))
	}
	if m.Metadata()[0].Field != "unit" || m.Metadata()[0].Value != "%" {
		t.Errorf("Metadata()[0] = %+v", m.Metadata()[0])
	}
	if m.Metadata()[1].Field != "aggregation" || m.Metadata()[1].Value != "avg" {
		t.Errorf("Metadata()[1] = %+v", m.Metadata()[1])
	}
	if len(m.Data()) != 2 {
		t.Errorf("Data() len = %d, want 2", len(m.Data()))
	}
	if m.Data()[0].Time != "2025-03-10T08:00:00Z" || m.Data()[0].Measure != "72.5" {
		t.Errorf("Data()[0] = %+v", m.Data()[0])
	}
	if m.Data()[1].Time != "2025-03-10T08:05:00Z" || m.Data()[1].Measure != "68.1" {
		t.Errorf("Data()[1] = %+v", m.Data()[1])
	}
	if m.Raw() != &wire {
		t.Error("Raw() does not point to original wire struct")
	}
}

func TestMetric_FromResponse_Nil(t *testing.T) {
	m := &Metric{}
	m.fromResponse(nil)
	if m.response != nil {
		t.Error("fromResponse(nil) should not set response")
	}
}

func TestMetric_Accessors_ZeroValue(t *testing.T) {
	m := &Metric{}
	if m.ID() != "" {
		t.Errorf("ID() on zero = %q", m.ID())
	}
	if m.URI() != "" {
		t.Errorf("URI() on zero = %q", m.URI())
	}
	if m.ReferenceID() != "" {
		t.Error("ReferenceID() on zero should be empty")
	}
	if m.Name() != "" {
		t.Error("Name() on zero should be empty")
	}
	if m.ReferenceName() != "" {
		t.Error("ReferenceName() on zero should be empty")
	}
	if m.Metadata() != nil {
		t.Error("Metadata() on zero should be nil")
	}
	if m.Data() != nil {
		t.Error("Data() on zero should be nil")
	}
	if m.Raw() != nil {
		t.Error("Raw() on zero should be nil")
	}
	if m.ProjectID() != "" {
		t.Error("ProjectID() on zero should be empty")
	}
}

func TestMetric_ID_FromReferenceID(t *testing.T) {
	m := &Metric{}
	m.fromResponse(&types.MetricResponse{ReferenceID: "metric-xyz"})
	if m.ID() != "metric-xyz" {
		t.Errorf("ID() = %q, want metric-xyz", m.ID())
	}
}

func TestMetric_URI_AlwaysEmpty(t *testing.T) {
	wire := metricMakeFullResponse()
	m := &Metric{}
	m.fromResponse(&wire)
	m.projectID = "p-1"
	if m.URI() != "" {
		t.Errorf("URI() = %q; metrics have no individual fetch URI", m.URI())
	}
}

func TestMetric_Metadata_PreservesOrderAndContent(t *testing.T) {
	entries := []types.MetricMetadataResponse{
		{Field: "unit", Value: "%"},
		{Field: "aggregation", Value: "avg"},
		{Field: "type", Value: "gauge"},
	}
	m := &Metric{}
	m.fromResponse(&types.MetricResponse{Metadata: entries})
	got := m.Metadata()
	if !reflect.DeepEqual(got, entries) {
		t.Errorf("Metadata() = %+v, want %+v", got, entries)
	}
}

func TestMetric_Data_PreservesOrderAndContent(t *testing.T) {
	datapoints := []types.MetricDataResponse{
		{Time: "2025-01-01T00:00:00Z", Measure: "10.0"},
		{Time: "2025-01-01T00:05:00Z", Measure: "20.0"},
		{Time: "2025-01-01T00:10:00Z", Measure: "15.5"},
	}
	m := &Metric{}
	m.fromResponse(&types.MetricResponse{Data: datapoints})
	got := m.Data()
	if !reflect.DeepEqual(got, datapoints) {
		t.Errorf("Data() = %+v, want %+v", got, datapoints)
	}
}

// --------------------------------------------------------------------------
// Adapter — happy paths
// --------------------------------------------------------------------------

const metricListBody = `{` +
	`"total":2,` +
	`"self":"/self","prev":"","next":"/next","first":"/first","last":"/last",` +
	`"values":[` +
	`{"referenceId":"metric-1","name":"cpu_usage","referenceName":"CPU Usage",` +
	`"metadata":[{"field":"unit","value":"%"}],"data":[{"time":"2025-03-10T08:00:00Z","measure":"72.5"}]},` +
	`{"referenceId":"metric-2","name":"mem_usage","referenceName":"Memory Usage",` +
	`"metadata":[{"field":"unit","value":"MB"}],"data":[{"time":"2025-03-10T08:00:00Z","measure":"1024"}]}` +
	`]}`

func TestMetricsClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildMetricTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, metricListBody)
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
	if items[0].ID() != "metric-1" {
		t.Errorf("items[0].ID() = %q", items[0].ID())
	}
	if items[0].Name() != "cpu_usage" {
		t.Errorf("items[0].Name() = %q", items[0].Name())
	}
	if items[0].ReferenceID() != "metric-1" {
		t.Errorf("items[0].ReferenceID() = %q", items[0].ReferenceID())
	}
	if items[0].ReferenceName() != "CPU Usage" {
		t.Errorf("items[0].ReferenceName() = %q", items[0].ReferenceName())
	}
	if items[0].ProjectID() != "p-1" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
	if items[0].StatusCode() != http.StatusOK {
		t.Errorf("items[0].StatusCode() = %d", items[0].StatusCode())
	}
	if len(items[0].Metadata()) != 1 {
		t.Errorf("items[0].Metadata() len = %d, want 1", len(items[0].Metadata()))
	}
	if len(items[0].Data()) != 1 {
		t.Errorf("items[0].Data() len = %d, want 1", len(items[0].Data()))
	}
	if items[1].ID() != "metric-2" {
		t.Errorf("items[1].ID() = %q", items[1].ID())
	}
}

func TestMetricsClientAdapter_List_Empty(t *testing.T) {
	adapter := buildMetricTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestMetricsClientAdapter_List_BadProjectRef(t *testing.T) {
	callCount := 0
	adapter := buildMetricTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestMetricsClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildMetricTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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

func TestMetricsClientAdapter_List_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := &metricsClientAdapter{low: metric.NewMetricsClientImpl(testutil.NewClient(t, server.URL))}
	_, err := adapter.List(context.Background(), URI("/projects/p-1"))
	if err == nil {
		t.Fatal("expected transport error from connection hijack")
	}
}

// --------------------------------------------------------------------------
// Constructor & contract
// --------------------------------------------------------------------------

func TestNewMetricsClientAdapter_Nil(t *testing.T) {
	a := newMetricsClientAdapter(nil)
	if a == nil {
		t.Fatal("newMetricsClientAdapter(nil) returned nil")
	}
	if a.low != nil {
		t.Error("expected low == nil when rest is nil")
	}
}

func TestNewMetricsClientAdapter_NonNil(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	client := testutil.NewClient(t, server.URL)
	a := newMetricsClientAdapter(client)
	if a == nil {
		t.Fatal("newMetricsClientAdapter(client) returned nil")
	}
	if a.low == nil {
		t.Error("expected low != nil when rest is non-nil")
	}
}

func TestMetricsClient_OnlyHasList(t *testing.T) {
	iface := reflect.TypeOf((*MetricsClient)(nil)).Elem()
	if iface.NumMethod() != 1 {
		t.Errorf("MetricsClient has %d methods, want exactly 1 (List)", iface.NumMethod())
	}
	if iface.Method(0).Name != "List" {
		t.Errorf("MetricsClient.Method(0) = %q, want List", iface.Method(0).Name)
	}
	for i := range iface.NumMethod() {
		name := iface.Method(i).Name
		if name == "Create" || name == "Update" || name == "Delete" || name == "Get" {
			t.Errorf("MetricsClient must not expose %s — metrics are read-only (Family C)", name)
		}
	}
}

// --------------------------------------------------------------------------
// Refetch placeholder coverage
// --------------------------------------------------------------------------

func TestMetricsClientAdapter_List_RefetchReturnsError(t *testing.T) {
	adapter := buildMetricTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":1,"next":"/next-page","values":[`+
			`{"referenceId":"metric-1","name":"cpu_usage"}`+
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
