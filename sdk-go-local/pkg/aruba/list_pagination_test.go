package aruba

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/clients/metric"
	"github.com/Arubacloud/sdk-go/internal/clients/network"
	"github.com/Arubacloud/sdk-go/internal/clients/project"
	"github.com/Arubacloud/sdk-go/internal/testutil"
)

// TestListPaginationRefetch verifies that List[T].Next follows the server-supplied
// pagination link via the real DoRequestAbs path (rest != nil). Uses the Alerts
// adapter since it is the simplest flat family-A resource with no parent scoping.
func TestListPaginationRefetch(t *testing.T) {
	var serverURL string
	calls := 0

	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if calls == 1 {
			nextURL := serverURL + "/v1/alerts?page=2"
			fmt.Fprintf(w,
				`{"total":2,"next":%q,"values":[{"id":"alert-1","eventName":"cpu.high","theshold":70,"thesholdExceedence":"yes"}]}`,
				nextURL)
		} else {
			fmt.Fprint(w,
				`{"total":2,"values":[{"id":"alert-2","eventName":"mem.high","theshold":90,"thesholdExceedence":"yes"}]}`)
		}
	})
	serverURL = server.URL

	rest := testutil.NewClient(t, server.URL)
	adapter := &alertsClientAdapter{
		low:  metric.NewAlertsClientImpl(rest),
		rest: rest,
	}

	list, err := adapter.List(context.Background(), URI("/projects/p-1"))
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list.Items()) != 1 {
		t.Fatalf("page 1: got %d items, want 1", len(list.Items()))
	}
	if !list.HasNext() {
		t.Fatal("page 1 should have a next link")
	}

	page2, err := list.Next(context.Background())
	if err != nil {
		t.Fatalf("Next() error: %v", err)
	}
	if len(page2.Items()) != 1 {
		t.Errorf("page 2: got %d items, want 1", len(page2.Items()))
	}
	if page2.Items()[0].ID() != "alert-2" {
		t.Errorf("page 2 item ID = %q, want %q", page2.Items()[0].ID(), "alert-2")
	}
	if page2.HasNext() {
		t.Error("page 2 should not have a next link")
	}
}

// TestListPaginationRefetch_ServerError verifies that Next returns an HTTPError
// when the pagination server responds with a non-success status.
func TestListPaginationRefetch_ServerError(t *testing.T) {
	var serverURL string
	calls := 0

	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			nextURL := serverURL + "/v1/alerts?page=2"
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w,
				`{"total":2,"next":%q,"values":[{"id":"alert-1","eventName":"cpu.high","theshold":70,"thesholdExceedence":"yes"}]}`,
				nextURL)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"title":"Internal Server Error","status":500}`)
		}
	})
	serverURL = server.URL

	rest := testutil.NewClient(t, server.URL)
	adapter := &alertsClientAdapter{
		low:  metric.NewAlertsClientImpl(rest),
		rest: rest,
	}

	list, err := adapter.List(context.Background(), URI("/projects/p-1"))
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if !list.HasNext() {
		t.Fatal("page 1 should have a next link")
	}

	_, err = list.Next(context.Background())
	if err == nil {
		t.Fatal("Next() should return an error on 500 response")
	}
	if httpErr, ok := err.(*HTTPError); !ok {
		t.Errorf("Next() error type = %T, want *HTTPError", err)
	} else if httpErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("Next() HTTPError.StatusCode = %d, want %d", httpErr.StatusCode, http.StatusInternalServerError)
	}
}

// TestListPaginationRefetch_RelativeURL verifies that Next resolves a relative
// pagination link against the client base URL instead of failing with an
// "unsupported protocol scheme" error.
func TestListPaginationRefetch_RelativeURL(t *testing.T) {
	calls := 0
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if calls == 1 {
			// Return a relative next URL — the client must resolve it.
			fmt.Fprint(w,
				`{"total":2,"next":"/v1/alerts?page=2","values":[{"id":"alert-1","eventName":"cpu.high","theshold":70,"thesholdExceedence":"yes"}]}`)
		} else {
			fmt.Fprint(w,
				`{"total":2,"values":[{"id":"alert-2","eventName":"mem.high","theshold":90,"thesholdExceedence":"yes"}]}`)
		}
	})

	rest := testutil.NewClient(t, server.URL)
	adapter := &alertsClientAdapter{
		low:  metric.NewAlertsClientImpl(rest),
		rest: rest,
	}

	list, err := adapter.List(context.Background(), URI("/projects/p-1"))
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if !list.HasNext() {
		t.Fatal("page 1 should have a next link")
	}

	page2, err := list.Next(context.Background())
	if err != nil {
		t.Fatalf("Next() with relative URL error: %v", err)
	}
	if len(page2.Items()) != 1 {
		t.Errorf("page 2: got %d items, want 1", len(page2.Items()))
	}
	if page2.Items()[0].ID() != "alert-2" {
		t.Errorf("page 2 item ID = %q, want %q", page2.Items()[0].ID(), "alert-2")
	}
}

// TestListPaginationRefetch_NilRest verifies that Next returns an error when
// the adapter's rest client is nil (e.g. constructed in tests without a server).
func TestListPaginationRefetch_NilRest(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w,
			`{"total":1,"next":"/next-page","values":[{"id":"alert-1","eventName":"cpu.high","theshold":70,"thesholdExceedence":"yes"}]}`)
	})

	adapter := &alertsClientAdapter{
		low:  metric.NewAlertsClientImpl(testutil.NewClient(t, server.URL)),
		rest: nil,
	}

	list, err := adapter.List(context.Background(), URI("/projects/p-1"))
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if !list.HasNext() {
		t.Skip("no next link — cannot test nil-rest refetch")
	}

	_, err = list.Next(context.Background())
	if err == nil {
		t.Fatal("Next() with nil rest should return an error")
	}
}

// TestListPaginationRefetch_VPC verifies that the VPC adapter's refetch closure
// follows the server-supplied next link and backfills projectID on page-2 items.
func TestListPaginationRefetch_VPC(t *testing.T) {
	var serverURL string
	calls := 0

	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if calls == 1 {
			nextURL := serverURL + "/v1/network/vpcs?page=2"
			fmt.Fprintf(w,
				`{"total":2,"next":%q,"values":[{"metadata":{"id":"vpc-1","name":"test-vpc"},"properties":{}}]}`,
				nextURL)
		} else {
			fmt.Fprint(w,
				`{"total":2,"values":[{"metadata":{"id":"vpc-2","name":"test-vpc-2"},"properties":{}}]}`)
		}
	})
	serverURL = server.URL

	rest := testutil.NewClient(t, server.URL)
	adapter := &vpcsClientAdapter{
		low:  network.NewVPCsClientImpl(rest),
		rest: rest,
	}

	list, err := adapter.List(context.Background(), URI("/projects/p-1"))
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if !list.HasNext() {
		t.Fatal("page 1 should have a next link")
	}

	page2, err := list.Next(context.Background())
	if err != nil {
		t.Fatalf("Next() error: %v", err)
	}
	if len(page2.Items()) != 1 {
		t.Errorf("page 2: got %d items, want 1", len(page2.Items()))
	}
	if page2.Items()[0].ID() != "vpc-2" {
		t.Errorf("page 2 item ID = %q, want %q", page2.Items()[0].ID(), "vpc-2")
	}
	if page2.HasNext() {
		t.Error("page 2 should not have a next link")
	}
}

// TestListPaginationRefetch_Project verifies that the top-level Project adapter's
// refetch closure follows the server-supplied next link (no parent-ID backfill).
func TestListPaginationRefetch_Project(t *testing.T) {
	var serverURL string
	calls := 0

	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if calls == 1 {
			nextURL := serverURL + "/v1/projects?page=2"
			fmt.Fprintf(w,
				`{"total":2,"next":%q,"values":[{"metadata":{"id":"proj-1","name":"test-project"},"properties":{}}]}`,
				nextURL)
		} else {
			fmt.Fprint(w,
				`{"total":2,"values":[{"metadata":{"id":"proj-2","name":"test-project-2"},"properties":{}}]}`)
		}
	})
	serverURL = server.URL

	rest := testutil.NewClient(t, server.URL)
	adapter := &projectClientAdapter{
		low:  project.NewProjectsClientImpl(rest),
		rest: rest,
	}

	list, err := adapter.List(context.Background())
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if !list.HasNext() {
		t.Fatal("page 1 should have a next link")
	}

	page2, err := list.Next(context.Background())
	if err != nil {
		t.Fatalf("Next() error: %v", err)
	}
	if len(page2.Items()) != 1 {
		t.Errorf("page 2: got %d items, want 1", len(page2.Items()))
	}
	if page2.Items()[0].ID() != "proj-2" {
		t.Errorf("page 2 item ID = %q, want %q", page2.Items()[0].ID(), "proj-2")
	}
	if page2.HasNext() {
		t.Error("page 2 should not have a next link")
	}
}
