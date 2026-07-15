package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"k8s.io/utils/ptr"
)

func TestListEvents(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && r.URL.Path == "/projects/test-project/providers/Aruba.Audit/events" {
				w.WriteHeader(http.StatusOK)
				resp := types.AuditEventListResponse{
					ListResponse: types.ListResponse{Total: 1},
					Values: []types.AuditEventResponse{
						{
							SeverityLevel: "Information",
							LogFormat: types.EventLogFormatVersionResponse{
								Version: "1.0",
							},
							Timestamp: time.Now(),
							Operation: types.EventOperationResponse{
								ID:    "Microsoft.Compute/virtualMachines/start/action",
								Value: ptr.To("Start Virtual Machine"),
							},
							Event: types.EventInfoResponse{
								ID:    "event-123",
								Value: ptr.To("Virtual Machine Started"),
								Type:  "operational",
							},
							Category: types.EventCategoryResponse{
								Value:       "Administrative",
								Description: ptr.To("Administrative operations"),
							},
							Origin:  "user",
							Channel: "Operation",
							Status: types.EventStatusResponse{
								Value:       "Succeeded",
								Description: ptr.To("Operation completed successfully"),
								Code:        ptr.To(int32(200)),
							},
							Identity: types.EventIdentityResponse{
								Caller: types.EventCallerResponse{
									Subject:  "user@example.com",
									Username: ptr.To("testuser"),
									Company:  ptr.To("TestCompany"),
									TenantID: ptr.To("tenant-123"),
								},
							},
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewEventsClientImpl(c)

		resp, err := svc.List(context.Background(), "test-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil || len(resp.Data.Values) != 1 {
			t.Errorf("expected 1 audit event")
		}
		if resp.Data.Values[0].SeverityLevel != "Information" {
			t.Errorf("expected severity level 'Information', got %s", resp.Data.Values[0].SeverityLevel)
		}
		if resp.Data.Values[0].Operation.ID != "Microsoft.Compute/virtualMachines/start/action" {
			t.Errorf("expected operation id 'Microsoft.Compute/virtualMachines/start/action', got %s", resp.Data.Values[0].Operation.ID)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && r.URL.Path == "/projects/test-project/providers/Aruba.Audit/events" {
				w.WriteHeader(http.StatusOK)
				resp := types.AuditEventListResponse{
					ListResponse: types.ListResponse{Total: 0},
					Values:       []types.AuditEventResponse{},
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewEventsClientImpl(c)

		resp, err := svc.List(context.Background(), "test-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil || len(resp.Data.Values) != 0 {
			t.Errorf("expected 0 audit events")
		}
	})

	t.Run("with pagination", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && r.URL.Path == "/projects/test-project/providers/Aruba.Audit/events" {
				limit := r.URL.Query().Get("limit")
				offset := r.URL.Query().Get("offset")
				if limit != "10" || offset != "5" {
					t.Errorf("expected limit=10 and offset=5, got limit=%s and offset=%s", limit, offset)
				}
				w.WriteHeader(http.StatusOK)
				resp := types.AuditEventListResponse{
					ListResponse: types.ListResponse{Total: 100},
					Values: []types.AuditEventResponse{
						{
							SeverityLevel: "Warning",
							LogFormat:     types.EventLogFormatVersionResponse{Version: "1.0"},
							Timestamp:     time.Now(),
							Operation:     types.EventOperationResponse{ID: "test-operation"},
							Event:         types.EventInfoResponse{ID: "event-456", Type: "operational"},
							Category:      types.EventCategoryResponse{Value: "Security"},
							Origin:        "system",
							Channel:       "Security",
							Status:        types.EventStatusResponse{Value: "Failed", Code: ptr.To(int32(403))},
							Identity:      types.EventIdentityResponse{Caller: types.EventCallerResponse{Subject: "system"}},
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewEventsClientImpl(c)

		params := &types.RequestParameters{
			Limit:  ptr.To(int32(10)),
			Offset: ptr.To(int32(5)),
		}
		resp, err := svc.List(context.Background(), "test-project", params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Total != 100 {
			t.Errorf("expected total 100, got %d", resp.Data.Total)
		}
		if len(resp.Data.Values) != 1 {
			t.Errorf("expected 1 audit event, got %d", len(resp.Data.Values))
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "project not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewEventsClientImpl(c)

		resp, err := svc.List(context.Background(), "test-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 response")
		}
		if resp.Error == nil {
			t.Fatalf("expected resp.Error to be populated")
		}
		if resp.Error.Title == nil || *resp.Error.Title != "Not Found" {
			t.Errorf("expected title 'Not Found', got %v", resp.Error.Title)
		}
	})

	t.Run("bad gateway non-json", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, "Bad Gateway")
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewEventsClientImpl(c)

		resp, err := svc.List(context.Background(), "test-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.StatusCode != http.StatusBadGateway {
			t.Fatalf("expected 502 response")
		}
		if resp.Error != nil {
			t.Errorf("expected resp.Error to be nil for non-JSON body, got %+v", resp.Error)
		}
		if string(resp.RawBody) != "Bad Gateway" {
			t.Errorf("expected RawBody 'Bad Gateway', got %q", string(resp.RawBody))
		}
	})

	t.Run("network error", func(t *testing.T) {
		c := testutil.NewBrokenClient(t, "http://unused.invalid")
		svc := NewEventsClientImpl(c)

		_, err := svc.List(context.Background(), "test-project", nil)
		if err == nil {
			t.Fatal("expected a network error, got nil")
		}
	})

	t.Run("nil params injects default api-version", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("api-version"); got != "1.0" {
				t.Errorf("expected api-version=1.0, got %q", got)
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"total":0,"values":[]}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewEventsClientImpl(c)

		if _, err := svc.List(context.Background(), "test-project", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
