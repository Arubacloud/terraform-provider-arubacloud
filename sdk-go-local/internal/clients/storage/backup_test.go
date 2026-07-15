package storage

import (
	"context"
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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"total":1,"values":[{"metadata":{"name":"test-backup"},"properties":{"type":"Full","sourceVolume":{"uri":"/projects/test-project/providers/Aruba.Storage/blockStorages/volume-123"},"retentionDays":10,"billingPeriod":"Monthly"},"status":{"state":"active"}}]}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		resp, err := svc.List(context.Background(), "test-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Data.Values) != 1 {
			t.Errorf("expected 1 backup, got %d", len(resp.Data.Values))
		}
		if resp.Data.Values[0].Metadata.Name == nil || *resp.Data.Values[0].Metadata.Name != "test-backup" {
			t.Errorf("expected name 'test-backup', got %v", resp.Data.Values[0].Metadata.Name)
		}
		if resp.Data.Values[0].Properties.Type != types.StorageBackupTypeFull {
			t.Errorf("expected type 'Full', got %q", resp.Data.Values[0].Properties.Type)
		}
	})

	t.Run("empty project", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewBackupClientImpl(c)
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
		svc := NewBackupClientImpl(c)
		resp, err := svc.List(context.Background(), "test-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
		if resp.Error == nil || resp.Error.Title == nil || *resp.Error.Title != "Not Found" {
			t.Errorf("expected error title 'Not Found', got %v", resp.Error)
		}
	})

	t.Run("bad gateway non-json", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, "Bad Gateway")
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		resp, err := svc.List(context.Background(), "test-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusBadGateway {
			t.Errorf("expected status 502, got %d", resp.StatusCode)
		}
		if resp.Error != nil {
			t.Errorf("expected nil Error, got %v", resp.Error)
		}
		if string(resp.RawBody) != "Bad Gateway" {
			t.Errorf("expected raw body 'Bad Gateway', got %q", string(resp.RawBody))
		}
	})

	t.Run("network error", func(t *testing.T) {
		c := testutil.NewBrokenClient(t, "http://unused.invalid")
		svc := NewBackupClientImpl(c)
		_, err := svc.List(context.Background(), "test-project", nil)
		if err == nil {
			t.Fatal("expected transport error, got nil")
		}
	})

	t.Run("nil params injects default api-version", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("api-version"); got != "1.0" {
				t.Errorf("expected api-version=1.0, got %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"total":0,"values":[]}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		resp, err := svc.List(context.Background(), "test-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestGetBackup(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"metadata":{"name":"test-backup","id":"backup-123"},"properties":{"type":"Full","sourceVolume":{"uri":"/projects/test-project/providers/Aruba.Storage/blockStorages/volume-123"},"retentionDays":10,"billingPeriod":"Monthly"},"status":{"state":"active"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		resp, err := svc.Get(context.Background(), "test-project", "backup-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "test-backup" {
			t.Errorf("expected name 'test-backup', got %v", resp.Data.Metadata.Name)
		}
		if resp.Data.Properties.Type != types.StorageBackupTypeFull {
			t.Errorf("expected type 'Full', got %q", resp.Data.Properties.Type)
		}
		if resp.Data.Properties.RetentionDays == nil || *resp.Data.Properties.RetentionDays != 10 {
			t.Errorf("expected RetentionDays 10, got %v", resp.Data.Properties.RetentionDays)
		}
	})

	t.Run("empty project", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewBackupClientImpl(c)
		_, err := svc.Get(context.Background(), "", "backup-123", nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty backup ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewBackupClientImpl(c)
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
		svc := NewBackupClientImpl(c)
		resp, err := svc.Get(context.Background(), "test-project", "backup-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
		if resp.Error == nil || resp.Error.Title == nil || *resp.Error.Title != "Not Found" {
			t.Errorf("expected error title 'Not Found', got %v", resp.Error)
		}
	})

	t.Run("bad gateway non-json", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, "Bad Gateway")
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		resp, err := svc.Get(context.Background(), "test-project", "backup-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusBadGateway {
			t.Errorf("expected status 502, got %d", resp.StatusCode)
		}
		if resp.Error != nil {
			t.Errorf("expected nil Error, got %v", resp.Error)
		}
		if string(resp.RawBody) != "Bad Gateway" {
			t.Errorf("expected raw body 'Bad Gateway', got %q", string(resp.RawBody))
		}
	})

	t.Run("network error", func(t *testing.T) {
		c := testutil.NewBrokenClient(t, "http://unused.invalid")
		svc := NewBackupClientImpl(c)
		_, err := svc.Get(context.Background(), "test-project", "backup-123", nil)
		if err == nil {
			t.Fatal("expected transport error, got nil")
		}
	})

	t.Run("nil params injects default api-version", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("api-version"); got != "1.0" {
				t.Errorf("expected api-version=1.0, got %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"metadata":{"name":"x"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		resp, err := svc.Get(context.Background(), "test-project", "backup-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestCreateBackup(t *testing.T) {
	t.Run("successful create", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"name":"new-backup","id":"backup-456","uri":"/projects/test-project/providers/Aruba.Storage/backups/backup-456"},"properties":{"type":"Full","sourceVolume":{"uri":"/projects/test-project/providers/Aruba.Storage/blockStorages/volume-456"},"retentionDays":20,"billingPeriod":"Yearly"},"status":{"state":"creating"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		body := types.StorageBackupRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{Name: "new-backup"},
			},
			Properties: types.StorageBackupPropertiesRequest{
				StorageBackupType: types.StorageBackupTypeFull,
				Origin:            types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Storage/blockStorages/volume-456"},
				RetentionDays:     ptr.To(20),
				BillingPeriod:     (*types.BillingPeriod)(ptr.To("Yearly")),
			},
		}
		resp, err := svc.Create(context.Background(), "test-project", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "new-backup" {
			t.Errorf("expected name 'new-backup', got %v", resp.Data.Metadata.Name)
		}
		if resp.Data.Metadata.ID == nil || *resp.Data.Metadata.ID != "backup-456" {
			t.Errorf("expected ID 'backup-456', got %v", resp.Data.Metadata.ID)
		}
		if resp.Data.Metadata.URI == nil || *resp.Data.Metadata.URI != "/projects/test-project/providers/Aruba.Storage/backups/backup-456" {
			t.Errorf("expected URI, got %v", resp.Data.Metadata.URI)
		}
		if resp.Data.Status.State == nil || *resp.Data.Status.State != "creating" {
			t.Errorf("expected state 'creating', got %v", resp.Data.Status.State)
		}
	})

	t.Run("empty project", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewBackupClientImpl(c)
		_, err := svc.Create(context.Background(), "", types.StorageBackupRequest{}, nil)
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
		svc := NewBackupClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.StorageBackupRequest{}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
		if resp.Error == nil || resp.Error.Title == nil || *resp.Error.Title != "Not Found" {
			t.Errorf("expected error title 'Not Found', got %v", resp.Error)
		}
	})

	// TODO(TD-010): Create uses the manual response-build flow and silently swallows
	// non-JSON unmarshal errors; resp.Error will be nil even though the body is non-JSON.
	t.Run("bad gateway non-json", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, "Bad Gateway")
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.StorageBackupRequest{}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusBadGateway {
			t.Errorf("expected status 502, got %d", resp.StatusCode)
		}
		if string(resp.RawBody) != "Bad Gateway" {
			t.Errorf("expected raw body 'Bad Gateway', got %q", string(resp.RawBody))
		}
	})

	t.Run("network error", func(t *testing.T) {
		c := testutil.NewBrokenClient(t, "http://unused.invalid")
		svc := NewBackupClientImpl(c)
		_, err := svc.Create(context.Background(), "test-project", types.StorageBackupRequest{}, nil)
		if err == nil {
			t.Fatal("expected transport error, got nil")
		}
	})

	t.Run("nil params injects default api-version", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("api-version"); got != "1.0" {
				t.Errorf("expected api-version=1.0, got %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"id":"x","uri":"/x","name":"x"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.StorageBackupRequest{}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("expected status 201, got %d", resp.StatusCode)
		}
	})

	t.Run("successful create missing id", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"uri":"/projects/test-project/providers/Aruba.Storage/backups/res-123","name":"test-name"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.StorageBackupRequest{}, nil)
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
			fmt.Fprint(w, `{"metadata":{"id":"res-123","uri":"/projects/test-project/providers/Aruba.Storage/backups/res-123"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.StorageBackupRequest{}, nil)
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

func TestUpdateBackup(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"metadata":{"name":"updated-backup","id":"backup-123"},"properties":{"type":"Incremental","sourceVolume":{"uri":"/projects/test-project/providers/Aruba.Storage/blockStorages/volume-123"},"retentionDays":30,"billingPeriod":"Monthly"},"status":{"state":"updating"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		body := types.StorageBackupRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{Name: "updated-backup"},
			},
			Properties: types.StorageBackupPropertiesRequest{
				StorageBackupType: types.StorageBackupTypeIncremental,
				Origin:            types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Storage/blockStorages/volume-123"},
				RetentionDays:     ptr.To(30),
				BillingPeriod:     (*types.BillingPeriod)(ptr.To("Monthly")),
			},
		}
		resp, err := svc.Update(context.Background(), "test-project", "backup-123", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "updated-backup" {
			t.Errorf("expected name 'updated-backup', got %v", resp.Data.Metadata.Name)
		}
		if resp.Data.Properties.Type != types.StorageBackupTypeIncremental {
			t.Errorf("expected type 'Incremental', got %q", resp.Data.Properties.Type)
		}
		if resp.Data.Properties.RetentionDays == nil || *resp.Data.Properties.RetentionDays != 30 {
			t.Errorf("expected RetentionDays 30, got %v", resp.Data.Properties.RetentionDays)
		}
	})

	t.Run("empty project", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewBackupClientImpl(c)
		_, err := svc.Update(context.Background(), "", "backup-123", types.StorageBackupRequest{}, nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty backup ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewBackupClientImpl(c)
		_, err := svc.Update(context.Background(), "test-project", "", types.StorageBackupRequest{}, nil)
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
		svc := NewBackupClientImpl(c)
		resp, err := svc.Update(context.Background(), "test-project", "backup-123", types.StorageBackupRequest{}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
		if resp.Error == nil || resp.Error.Title == nil || *resp.Error.Title != "Not Found" {
			t.Errorf("expected error title 'Not Found', got %v", resp.Error)
		}
	})

	// TODO(TD-010): Update uses the manual response-build flow and silently swallows
	// non-JSON unmarshal errors; resp.Error will be nil even though the body is non-JSON.
	t.Run("bad gateway non-json", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, "Bad Gateway")
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		resp, err := svc.Update(context.Background(), "test-project", "backup-123", types.StorageBackupRequest{}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusBadGateway {
			t.Errorf("expected status 502, got %d", resp.StatusCode)
		}
		if string(resp.RawBody) != "Bad Gateway" {
			t.Errorf("expected raw body 'Bad Gateway', got %q", string(resp.RawBody))
		}
	})

	t.Run("network error", func(t *testing.T) {
		c := testutil.NewBrokenClient(t, "http://unused.invalid")
		svc := NewBackupClientImpl(c)
		_, err := svc.Update(context.Background(), "test-project", "backup-123", types.StorageBackupRequest{}, nil)
		if err == nil {
			t.Fatal("expected transport error, got nil")
		}
	})

	t.Run("nil params injects default api-version", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("api-version"); got != "1.0" {
				t.Errorf("expected api-version=1.0, got %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"metadata":{"name":"x"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		resp, err := svc.Update(context.Background(), "test-project", "backup-123", types.StorageBackupRequest{}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestDeleteBackup(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		_, err := svc.Delete(context.Background(), "test-project", "backup-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty project", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewBackupClientImpl(c)
		_, err := svc.Delete(context.Background(), "", "backup-123", nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty backup ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewBackupClientImpl(c)
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
		svc := NewBackupClientImpl(c)
		resp, err := svc.Delete(context.Background(), "test-project", "backup-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
		if resp.Error == nil || resp.Error.Title == nil || *resp.Error.Title != "Not Found" {
			t.Errorf("expected error title 'Not Found', got %v", resp.Error)
		}
	})

	t.Run("bad gateway non-json", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, "Bad Gateway")
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewBackupClientImpl(c)
		resp, err := svc.Delete(context.Background(), "test-project", "backup-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusBadGateway {
			t.Errorf("expected status 502, got %d", resp.StatusCode)
		}
		if resp.Error != nil {
			t.Errorf("expected nil Error, got %v", resp.Error)
		}
		if string(resp.RawBody) != "Bad Gateway" {
			t.Errorf("expected raw body 'Bad Gateway', got %q", string(resp.RawBody))
		}
	})

	t.Run("network error", func(t *testing.T) {
		c := testutil.NewBrokenClient(t, "http://unused.invalid")
		svc := NewBackupClientImpl(c)
		_, err := svc.Delete(context.Background(), "test-project", "backup-123", nil)
		if err == nil {
			t.Fatal("expected transport error, got nil")
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
		svc := NewBackupClientImpl(c)
		resp, err := svc.Delete(context.Background(), "test-project", "backup-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", resp.StatusCode)
		}
	})
}
