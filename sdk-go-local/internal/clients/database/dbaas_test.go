package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"k8s.io/utils/ptr"
)

func statePtr(s types.State) *types.State { return &s }

func TestListDBaaS(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && r.URL.Path == "/projects/test-project/providers/Aruba.Database/dbaas" {
				w.WriteHeader(http.StatusOK)
				resp := types.DBaaSListResponse{
					ListResponse: types.ListResponse{Total: 1},
					Values: []types.DBaaSResponse{
						{
							Metadata: types.ResourceMetadataResponse{
								Name: ptr.To("dbaas-1"),
								ID:   ptr.To("dbaas-123"),
							},
							Properties: types.DBaaSPropertiesResponse{
								Engine: &types.DBaaSEngineResponse{
									Type:    ptr.To("MySQL"),
									Version: ptr.To("8.0"),
								},
							},
							Status: types.ResourceStatusResponse{State: statePtr(types.State("active"))},
						},
					},
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)

		resp, err := svc.List(context.Background(), "test-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil || len(resp.Data.Values) != 1 {
			t.Fatalf("expected 1 DBaaS")
		}
		if resp.Data.Values[0].Metadata.Name == nil || *resp.Data.Values[0].Metadata.Name != "dbaas-1" {
			t.Errorf("expected name 'dbaas-1'")
		}
	})

	t.Run("empty project", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewDBaaSClientImpl(c)
		_, err := svc.List(context.Background(), "", nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)

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
		svc := NewDBaaSClientImpl(c)

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
		svc := NewDBaaSClientImpl(c)
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
		svc := NewDBaaSClientImpl(c)
		if _, err := svc.List(context.Background(), "test-project", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetDBaaS(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && r.URL.Path == "/projects/test-project/providers/Aruba.Database/dbaas/dbaas-123" {
				w.WriteHeader(http.StatusOK)
				resp := types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{
						Name: ptr.To("dbaas-1"),
						ID:   ptr.To("dbaas-123"),
					},
					Properties: types.DBaaSPropertiesResponse{
						Engine: &types.DBaaSEngineResponse{
							Type:    ptr.To("MySQL"),
							Version: ptr.To("8.0"),
						},
						Flavor: &types.DBaaSFlavorResponse{
							Name: ptr.To("M4-8"),
						},
					},
					Status: types.ResourceStatusResponse{State: statePtr(types.State("active"))},
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)

		resp, err := svc.Get(context.Background(), "test-project", "dbaas-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "dbaas-1" {
			t.Errorf("expected name 'dbaas-1'")
		}
		if resp.Data.Status.State == nil || *resp.Data.Status.State != "active" {
			t.Errorf("expected state 'active'")
		}
	})

	t.Run("empty project", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewDBaaSClientImpl(c)
		_, err := svc.Get(context.Background(), "", "dbaas-123", nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty DBaaS ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewDBaaSClientImpl(c)
		_, err := svc.Get(context.Background(), "test-project", "", nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)

		resp, err := svc.Get(context.Background(), "test-project", "dbaas-123", nil)
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
		svc := NewDBaaSClientImpl(c)

		resp, err := svc.Get(context.Background(), "test-project", "dbaas-123", nil)
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
		svc := NewDBaaSClientImpl(c)
		_, err := svc.Get(context.Background(), "test-project", "dbaas-123", nil)
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
			fmt.Fprint(w, `{}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)
		if _, err := svc.Get(context.Background(), "test-project", "dbaas-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCreateDBaaS(t *testing.T) {
	t.Run("successful create", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && r.URL.Path == "/projects/test-project/providers/Aruba.Database/dbaas" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				resp := types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{
						Name: ptr.To("dbaas-1"),
						ID:   ptr.To("dbaas-123"),
						URI:  ptr.To("/projects/test-project/providers/Aruba.Database/dbaas/dbaas-123"),
					},
					Properties: types.DBaaSPropertiesResponse{
						Engine: &types.DBaaSEngineResponse{Type: ptr.To("MySQL")},
					},
					Status: types.ResourceStatusResponse{State: statePtr(types.State("creating"))},
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)

		body := types.DBaaSRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{Name: "dbaas-1"},
			},
			Properties: types.DBaaSPropertiesRequest{},
		}

		resp, err := svc.Create(context.Background(), "test-project", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "dbaas-1" {
			t.Errorf("expected name 'dbaas-1'")
		}
		if resp.Data.Metadata.ID == nil || *resp.Data.Metadata.ID != "dbaas-123" {
			t.Errorf("expected ID 'dbaas-123', got %v", resp.Data.Metadata.ID)
		}
		if resp.Data.Metadata.URI == nil || *resp.Data.Metadata.URI != "/projects/test-project/providers/Aruba.Database/dbaas/dbaas-123" {
			t.Errorf("expected URI, got %v", resp.Data.Metadata.URI)
		}
		if resp.Data.Status.State == nil || *resp.Data.Status.State != "creating" {
			t.Errorf("expected state 'creating'")
		}
	})

	t.Run("empty project", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewDBaaSClientImpl(c)
		_, err := svc.Create(context.Background(), "", types.DBaaSRequest{}, nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)

		resp, err := svc.Create(context.Background(), "test-project", types.DBaaSRequest{}, nil)
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
		// TODO(TD-010): Create's manual response build silently swallows non-JSON
		// unmarshal errors (diverges from ParseResponseBody which logs at DEBUG).
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, "Bad Gateway")
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)

		resp, err := svc.Create(context.Background(), "test-project", types.DBaaSRequest{}, nil)
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
		svc := NewDBaaSClientImpl(c)
		_, err := svc.Create(context.Background(), "test-project", types.DBaaSRequest{}, nil)
		if err == nil {
			t.Fatal("expected a network error, got nil")
		}
	})

	t.Run("nil params injects default api-version", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("api-version"); got != "1.0" {
				t.Errorf("expected api-version=1.0, got %q", got)
			}
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"id":"x","uri":"/x","name":"x"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)
		if _, err := svc.Create(context.Background(), "test-project", types.DBaaSRequest{}, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("successful create missing id", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"uri":"/projects/test-project/providers/Aruba.Database/dbaas/res-123","name":"test-name"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.DBaaSRequest{}, nil)
		if err == nil {
			t.Fatal("expected metadata validation error, got nil")
		}
		var mvErr *types.MetadataValidationError
		if !errors.As(err, &mvErr) {
			t.Fatalf("expected *types.MetadataValidationError, got %T: %v", err, err)
		}
		if len(mvErr.Missing) != 1 || mvErr.Missing[0] != "id" {
			t.Errorf("expected missing=[id], got %v", mvErr.Missing)
		}
		if resp == nil {
			t.Fatal("expected partial response alongside error")
		}
	})

	t.Run("successful create missing name", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"id":"res-123","uri":"/projects/test-project/providers/Aruba.Database/dbaas/res-123"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.DBaaSRequest{}, nil)
		if err == nil {
			t.Fatal("expected metadata validation error, got nil")
		}
		var mvErr *types.MetadataValidationError
		if !errors.As(err, &mvErr) {
			t.Fatalf("expected *types.MetadataValidationError, got %T: %v", err, err)
		}
		if len(mvErr.Missing) != 1 || mvErr.Missing[0] != "name" {
			t.Errorf("expected missing=[name], got %v", mvErr.Missing)
		}
		if resp == nil {
			t.Fatal("expected partial response alongside error")
		}
	})
}

func TestUpdateDBaaS(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "PUT" && r.URL.Path == "/projects/test-project/providers/Aruba.Database/dbaas/dbaas-123" {
				w.WriteHeader(http.StatusOK)
				resp := types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{
						Name: ptr.To("dbaas-updated"),
						ID:   ptr.To("dbaas-123"),
					},
					Properties: types.DBaaSPropertiesResponse{
						Engine: &types.DBaaSEngineResponse{Type: ptr.To("MySQL")},
					},
					Status: types.ResourceStatusResponse{State: statePtr(types.State("updating"))},
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)

		body := types.DBaaSRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{Name: "dbaas-updated"},
			},
			Properties: types.DBaaSPropertiesRequest{},
		}

		resp, err := svc.Update(context.Background(), "test-project", "dbaas-123", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "dbaas-updated" {
			t.Errorf("expected name 'dbaas-updated'")
		}
		if resp.Data.Status.State == nil || *resp.Data.Status.State != "updating" {
			t.Errorf("expected state 'updating'")
		}
	})

	t.Run("empty project", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewDBaaSClientImpl(c)
		_, err := svc.Update(context.Background(), "", "dbaas-123", types.DBaaSRequest{}, nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty DBaaS ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewDBaaSClientImpl(c)
		_, err := svc.Update(context.Background(), "test-project", "", types.DBaaSRequest{}, nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)

		resp, err := svc.Update(context.Background(), "test-project", "dbaas-123", types.DBaaSRequest{}, nil)
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
		// TODO(TD-010): Update's manual response build silently swallows non-JSON
		// unmarshal errors (diverges from ParseResponseBody which logs at DEBUG).
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, "Bad Gateway")
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)

		resp, err := svc.Update(context.Background(), "test-project", "dbaas-123", types.DBaaSRequest{}, nil)
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
		svc := NewDBaaSClientImpl(c)
		_, err := svc.Update(context.Background(), "test-project", "dbaas-123", types.DBaaSRequest{}, nil)
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
			fmt.Fprint(w, `{}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)
		if _, err := svc.Update(context.Background(), "test-project", "dbaas-123", types.DBaaSRequest{}, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestDeleteDBaaS(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "DELETE" && r.URL.Path == "/projects/test-project/providers/Aruba.Database/dbaas/dbaas-123" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)

		_, err := svc.Delete(context.Background(), "test-project", "dbaas-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty project", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewDBaaSClientImpl(c)
		_, err := svc.Delete(context.Background(), "", "dbaas-123", nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty DBaaS ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewDBaaSClientImpl(c)
		_, err := svc.Delete(context.Background(), "test-project", "", nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)

		resp, err := svc.Delete(context.Background(), "test-project", "dbaas-123", nil)
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
		svc := NewDBaaSClientImpl(c)

		resp, err := svc.Delete(context.Background(), "test-project", "dbaas-123", nil)
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
		svc := NewDBaaSClientImpl(c)
		_, err := svc.Delete(context.Background(), "test-project", "dbaas-123", nil)
		if err == nil {
			t.Fatal("expected a network error, got nil")
		}
	})

	t.Run("nil params injects default api-version", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("api-version"); got != "1.0" {
				t.Errorf("expected api-version=1.0, got %q", got)
			}
			w.WriteHeader(http.StatusNoContent)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewDBaaSClientImpl(c)
		if _, err := svc.Delete(context.Background(), "test-project", "dbaas-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
