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

func TestListBackups(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && r.URL.Path == "/projects/test-project/providers/Aruba.Database/backups" {
				w.WriteHeader(http.StatusOK)
				resp := types.DBaaSBackupListResponse{
					ListResponse: types.ListResponse{Total: 1},
					Values: []types.BackupResponse{
						{
							Metadata: types.ResourceMetadataResponse{
								Name: ptr.To("backup-1"),
								ID:   ptr.To("bk-123"),
							},
							Properties: types.BackupPropertiesResponse{
								Zone:     "ITBG-1",
								DBaaS:    types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Database/dbaas/dbaas-1"},
								Database: types.DatabaseNameRef{Name: "db-1"},
								BillingPlanCommon: func() *types.BillingPlanCommon {
									v := types.BillingPeriod("Hour")
									return &types.BillingPlanCommon{BillingPeriod: &v}
								}(),
								Storage: types.BackupStorageResponse{Size: 10},
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
		svc := NewBackupsClientImpl(c)

		resp, err := svc.List(context.Background(), "test-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil || len(resp.Data.Values) != 1 {
			t.Fatalf("expected 1 backup")
		}
		if resp.Data.Values[0].Metadata.Name == nil || *resp.Data.Values[0].Metadata.Name != "backup-1" {
			t.Errorf("expected name 'backup-1'")
		}
		if resp.Data.Values[0].Properties.Zone != "ITBG-1" {
			t.Errorf("expected zone 'ITBG-1', got %s", resp.Data.Values[0].Properties.Zone)
		}
	})

	t.Run("empty project", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewBackupsClientImpl(c)
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
		svc := NewBackupsClientImpl(c)

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
		svc := NewBackupsClientImpl(c)

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
		svc := NewBackupsClientImpl(c)
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
		svc := NewBackupsClientImpl(c)
		if _, err := svc.List(context.Background(), "test-project", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetBackup(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && r.URL.Path == "/projects/test-project/providers/Aruba.Database/backups/bk-123" {
				w.WriteHeader(http.StatusOK)
				resp := types.BackupResponse{
					Metadata: types.ResourceMetadataResponse{
						Name: ptr.To("backup-1"),
						ID:   ptr.To("bk-123"),
					},
					Properties: types.BackupPropertiesResponse{
						Zone:     "ITBG-1",
						DBaaS:    types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Database/dbaas/dbaas-1"},
						Database: types.DatabaseNameRef{Name: "db-1"},
						BillingPlanCommon: func() *types.BillingPlanCommon {
							v := types.BillingPeriod("Hour")
							return &types.BillingPlanCommon{BillingPeriod: &v}
						}(),
						Storage: types.BackupStorageResponse{Size: 10},
					},
					Status: types.ResourceStatusResponse{State: statePtr(types.State("active"))},
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupsClientImpl(c)

		resp, err := svc.Get(context.Background(), "test-project", "bk-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "backup-1" {
			t.Errorf("expected name 'backup-1'")
		}
		if resp.Data.Properties.Zone != "ITBG-1" {
			t.Errorf("expected zone 'ITBG-1', got %s", resp.Data.Properties.Zone)
		}
	})

	t.Run("empty project", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewBackupsClientImpl(c)
		_, err := svc.Get(context.Background(), "", "bk-123", nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty backup ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewBackupsClientImpl(c)
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
		svc := NewBackupsClientImpl(c)

		resp, err := svc.Get(context.Background(), "test-project", "bk-123", nil)
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
		svc := NewBackupsClientImpl(c)

		resp, err := svc.Get(context.Background(), "test-project", "bk-123", nil)
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
		svc := NewBackupsClientImpl(c)
		_, err := svc.Get(context.Background(), "test-project", "bk-123", nil)
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
		svc := NewBackupsClientImpl(c)
		if _, err := svc.Get(context.Background(), "test-project", "bk-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCreateBackup(t *testing.T) {
	t.Run("successful create", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && r.URL.Path == "/projects/test-project/providers/Aruba.Database/backups" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				resp := types.BackupResponse{
					Metadata: types.ResourceMetadataResponse{
						Name: ptr.To("backup-1"),
						ID:   ptr.To("bk-123"),
						URI:  ptr.To("/projects/test-project/providers/Aruba.Database/backups/bk-123"),
					},
					Properties: types.BackupPropertiesResponse{
						Zone:     "ITBG-1",
						DBaaS:    types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Database/dbaas/dbaas-1"},
						Database: types.DatabaseNameRef{Name: "db-1"},
						BillingPlanCommon: func() *types.BillingPlanCommon {
							v := types.BillingPeriod("Hour")
							return &types.BillingPlanCommon{BillingPeriod: &v}
						}(),
					},
					Status: types.ResourceStatusResponse{State: statePtr(types.State("creating"))},
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupsClientImpl(c)

		body := types.BackupRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{Name: "backup-1"},
			},
			Properties: types.BackupPropertiesRequest{
				Zone:     "ITBG-1",
				DBaaS:    types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Database/dbaas/dbaas-1"},
				Database: types.DatabaseNameRef{Name: "db-1"},
				BillingPlanCommon: func() *types.BillingPlanCommon {
					v := types.BillingPeriod("Hour")
					return &types.BillingPlanCommon{BillingPeriod: &v}
				}(),
			},
		}

		resp, err := svc.Create(context.Background(), "test-project", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "backup-1" {
			t.Errorf("expected name 'backup-1'")
		}
		if resp.Data.Metadata.ID == nil || *resp.Data.Metadata.ID != "bk-123" {
			t.Errorf("expected ID 'bk-123', got %v", resp.Data.Metadata.ID)
		}
		if resp.Data.Metadata.URI == nil || *resp.Data.Metadata.URI != "/projects/test-project/providers/Aruba.Database/backups/bk-123" {
			t.Errorf("expected URI, got %v", resp.Data.Metadata.URI)
		}
		if resp.Data.Status.State == nil || *resp.Data.Status.State != "creating" {
			t.Errorf("expected state 'creating'")
		}
	})

	t.Run("empty project", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewBackupsClientImpl(c)
		_, err := svc.Create(context.Background(), "", types.BackupRequest{}, nil)
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
		svc := NewBackupsClientImpl(c)

		resp, err := svc.Create(context.Background(), "test-project", types.BackupRequest{}, nil)
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
		svc := NewBackupsClientImpl(c)

		resp, err := svc.Create(context.Background(), "test-project", types.BackupRequest{}, nil)
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
		svc := NewBackupsClientImpl(c)
		_, err := svc.Create(context.Background(), "test-project", types.BackupRequest{}, nil)
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
		svc := NewBackupsClientImpl(c)
		if _, err := svc.Create(context.Background(), "test-project", types.BackupRequest{}, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("successful create missing id", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"uri":"/projects/test-project/providers/Aruba.Database/backups/res-123","name":"test-name"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupsClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.BackupRequest{}, nil)
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
			fmt.Fprint(w, `{"metadata":{"id":"res-123","uri":"/projects/test-project/providers/Aruba.Database/backups/res-123"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupsClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.BackupRequest{}, nil)
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

func TestDeleteBackup(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "DELETE" && r.URL.Path == "/projects/test-project/providers/Aruba.Database/backups/bk-123" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupsClientImpl(c)

		_, err := svc.Delete(context.Background(), "test-project", "bk-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty project", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewBackupsClientImpl(c)
		_, err := svc.Delete(context.Background(), "", "bk-123", nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty backup ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewBackupsClientImpl(c)
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
		svc := NewBackupsClientImpl(c)

		resp, err := svc.Delete(context.Background(), "test-project", "bk-123", nil)
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
		svc := NewBackupsClientImpl(c)

		resp, err := svc.Delete(context.Background(), "test-project", "bk-123", nil)
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
		svc := NewBackupsClientImpl(c)
		_, err := svc.Delete(context.Background(), "test-project", "bk-123", nil)
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
		svc := NewBackupsClientImpl(c)
		if _, err := svc.Delete(context.Background(), "test-project", "bk-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
