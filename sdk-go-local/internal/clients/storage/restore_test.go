package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"k8s.io/utils/ptr"
)

func newRestoreSvc(t *testing.T, baseURL string) *restoreClientImpl {
	t.Helper()
	c := testutil.NewClient(t, baseURL)
	return NewRestoreClientImpl(c, NewBackupClientImpl(c))
}

func TestListRestores(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"total":1,"values":[{"metadata":{"name":"test-restore"},"properties":{"destinationVolume":{"uri":"/projects/test-project/providers/Aruba.Storage/blockStorages/vol-789"}}}]}`)
		})
		svc := newRestoreSvc(t, server.URL)
		resp, err := svc.List(context.Background(), "test-project", "backup-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Data.Values) != 1 {
			t.Errorf("expected 1 restore, got %d", len(resp.Data.Values))
		}
		if resp.Data.Values[0].Metadata.Name == nil || *resp.Data.Values[0].Metadata.Name != "test-restore" {
			t.Errorf("expected name 'test-restore', got %v", resp.Data.Values[0].Metadata.Name)
		}
		if resp.Data.Values[0].Properties.Destination.URI == "" {
			t.Errorf("expected Destination URI to be set")
		}
	})

	t.Run("empty project", func(t *testing.T) {
		svc := newRestoreSvc(t, "http://unused.invalid")
		_, err := svc.List(context.Background(), "", "backup-123", nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty backup ID", func(t *testing.T) {
		svc := newRestoreSvc(t, "http://unused.invalid")
		_, err := svc.List(context.Background(), "test-project", "", nil)
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
		svc := newRestoreSvc(t, server.URL)
		resp, err := svc.List(context.Background(), "test-project", "backup-123", nil)
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
		svc := newRestoreSvc(t, server.URL)
		resp, err := svc.List(context.Background(), "test-project", "backup-123", nil)
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
		svc := NewRestoreClientImpl(c, NewBackupClientImpl(c))
		_, err := svc.List(context.Background(), "test-project", "backup-123", nil)
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
		svc := newRestoreSvc(t, server.URL)
		resp, err := svc.List(context.Background(), "test-project", "backup-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestGetRestore(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"metadata":{"name":"test-restore","id":"restore-123"},"properties":{"destinationVolume":{"uri":"/projects/test-project/providers/Aruba.Storage/blockStorages/vol-789"}}}`)
		})
		svc := newRestoreSvc(t, server.URL)
		resp, err := svc.Get(context.Background(), "test-project", "backup-123", "restore-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "test-restore" {
			t.Errorf("expected name 'test-restore', got %v", resp.Data.Metadata.Name)
		}
		if resp.Data.Properties.Destination.URI == "" {
			t.Errorf("expected Destination URI to be set")
		}
	})

	t.Run("empty project", func(t *testing.T) {
		svc := newRestoreSvc(t, "http://unused.invalid")
		_, err := svc.Get(context.Background(), "", "backup-123", "restore-123", nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty backup ID", func(t *testing.T) {
		svc := newRestoreSvc(t, "http://unused.invalid")
		_, err := svc.Get(context.Background(), "test-project", "", "restore-123", nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty restore ID", func(t *testing.T) {
		svc := newRestoreSvc(t, "http://unused.invalid")
		_, err := svc.Get(context.Background(), "test-project", "backup-123", "", nil)
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
		svc := newRestoreSvc(t, server.URL)
		resp, err := svc.Get(context.Background(), "test-project", "backup-123", "restore-123", nil)
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
		svc := newRestoreSvc(t, server.URL)
		resp, err := svc.Get(context.Background(), "test-project", "backup-123", "restore-123", nil)
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
		svc := NewRestoreClientImpl(c, NewBackupClientImpl(c))
		_, err := svc.Get(context.Background(), "test-project", "backup-123", "restore-123", nil)
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
		svc := newRestoreSvc(t, server.URL)
		resp, err := svc.Get(context.Background(), "test-project", "backup-123", "restore-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestCreateRestore(t *testing.T) {
	t.Run("successful create", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"name":"new-restore","id":"restore-456","uri":"/projects/test-project/providers/Aruba.Storage/restores/restore-456"},"properties":{"destinationVolume":{"uri":"/projects/test-project/providers/Aruba.Storage/blockStorages/vol-789"}},"status":{"state":"creating"}}`)
		})
		svc := newRestoreSvc(t, server.URL)
		body := types.StorageRestoreRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{Name: "new-restore"},
			},
			Properties: types.StorageRestorePropertiesRequest{
				Target: types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Storage/blockStorages/vol-789"},
			},
		}
		resp, err := svc.Create(context.Background(), "test-project", "backup-123", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "new-restore" {
			t.Errorf("expected name 'new-restore', got %v", resp.Data.Metadata.Name)
		}
		if resp.Data.Metadata.ID == nil || *resp.Data.Metadata.ID != "restore-456" {
			t.Errorf("expected ID 'restore-456', got %v", resp.Data.Metadata.ID)
		}
		if resp.Data.Metadata.URI == nil || *resp.Data.Metadata.URI != "/projects/test-project/providers/Aruba.Storage/restores/restore-456" {
			t.Errorf("expected URI, got %v", resp.Data.Metadata.URI)
		}
		if resp.Data.Status.State == nil || *resp.Data.Status.State != "creating" {
			t.Errorf("expected state 'creating', got %v", resp.Data.Status.State)
		}
	})

	t.Run("empty project", func(t *testing.T) {
		svc := newRestoreSvc(t, "http://unused.invalid")
		_, err := svc.Create(context.Background(), "", "backup-123", types.StorageRestoreRequest{}, nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty backup ID", func(t *testing.T) {
		svc := newRestoreSvc(t, "http://unused.invalid")
		_, err := svc.Create(context.Background(), "test-project", "", types.StorageRestoreRequest{}, nil)
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
		svc := newRestoreSvc(t, server.URL)
		body := types.StorageRestoreRequest{Properties: types.StorageRestorePropertiesRequest{Target: types.ReferenceResourceCommon{URI: "dummy"}}}
		resp, err := svc.Create(context.Background(), "test-project", "backup-123", body, nil)
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
		svc := newRestoreSvc(t, server.URL)
		body := types.StorageRestoreRequest{Properties: types.StorageRestorePropertiesRequest{Target: types.ReferenceResourceCommon{URI: "dummy"}}}
		resp, err := svc.Create(context.Background(), "test-project", "backup-123", body, nil)
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
		svc := NewRestoreClientImpl(c, NewBackupClientImpl(c))
		body := types.StorageRestoreRequest{Properties: types.StorageRestorePropertiesRequest{Target: types.ReferenceResourceCommon{URI: "dummy"}}}
		_, err := svc.Create(context.Background(), "test-project", "backup-123", body, nil)
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
		svc := newRestoreSvc(t, server.URL)
		body := types.StorageRestoreRequest{Properties: types.StorageRestorePropertiesRequest{Target: types.ReferenceResourceCommon{URI: "dummy"}}}
		resp, err := svc.Create(context.Background(), "test-project", "backup-123", body, nil)
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
			fmt.Fprint(w, `{"metadata":{"uri":"/projects/test-project/providers/Aruba.Storage/restores/res-123","name":"test-name"}}`)
		})
		svc := newRestoreSvc(t, server.URL)
		body := types.StorageRestoreRequest{Properties: types.StorageRestorePropertiesRequest{Target: types.ReferenceResourceCommon{URI: "dummy"}}}
		resp, err := svc.Create(context.Background(), "test-project", "backup-123", body, nil)
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
			fmt.Fprint(w, `{"metadata":{"id":"res-123","uri":"/projects/test-project/providers/Aruba.Storage/restores/res-123"}}`)
		})
		svc := newRestoreSvc(t, server.URL)
		body := types.StorageRestoreRequest{Properties: types.StorageRestorePropertiesRequest{Target: types.ReferenceResourceCommon{URI: "dummy"}}}
		resp, err := svc.Create(context.Background(), "test-project", "backup-123", body, nil)
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

func TestUpdateRestore(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"metadata":{"name":"updated-restore","id":"restore-123"},"properties":{"destinationVolume":{"uri":"/projects/test-project/providers/Aruba.Storage/blockStorages/vol-789"}},"status":{"state":"updating"}}`)
		})
		svc := newRestoreSvc(t, server.URL)
		body := types.StorageRestoreRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{Name: "updated-restore"},
			},
			Properties: types.StorageRestorePropertiesRequest{
				Target: types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Storage/blockStorages/vol-789"},
			},
		}
		resp, err := svc.Update(context.Background(), "test-project", "backup-123", "restore-123", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "updated-restore" {
			t.Errorf("expected name 'updated-restore', got %v", resp.Data.Metadata.Name)
		}
		if resp.Data.Status.State == nil || *resp.Data.Status.State != "updating" {
			t.Errorf("expected state 'updating', got %v", resp.Data.Status.State)
		}
	})

	t.Run("empty project", func(t *testing.T) {
		svc := newRestoreSvc(t, "http://unused.invalid")
		_, err := svc.Update(context.Background(), "", "backup-123", "restore-123", types.StorageRestoreRequest{}, nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty backup ID", func(t *testing.T) {
		svc := newRestoreSvc(t, "http://unused.invalid")
		_, err := svc.Update(context.Background(), "test-project", "", "restore-123", types.StorageRestoreRequest{}, nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty restore ID", func(t *testing.T) {
		svc := newRestoreSvc(t, "http://unused.invalid")
		_, err := svc.Update(context.Background(), "test-project", "backup-123", "", types.StorageRestoreRequest{}, nil)
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
		svc := newRestoreSvc(t, server.URL)
		resp, err := svc.Update(context.Background(), "test-project", "backup-123", "restore-123", types.StorageRestoreRequest{}, nil)
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
		svc := newRestoreSvc(t, server.URL)
		resp, err := svc.Update(context.Background(), "test-project", "backup-123", "restore-123", types.StorageRestoreRequest{}, nil)
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
		svc := NewRestoreClientImpl(c, NewBackupClientImpl(c))
		_, err := svc.Update(context.Background(), "test-project", "backup-123", "restore-123", types.StorageRestoreRequest{}, nil)
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
		svc := newRestoreSvc(t, server.URL)
		resp, err := svc.Update(context.Background(), "test-project", "backup-123", "restore-123", types.StorageRestoreRequest{}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestDeleteRestore(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
		svc := newRestoreSvc(t, server.URL)
		_, err := svc.Delete(context.Background(), "test-project", "backup-123", "restore-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty project", func(t *testing.T) {
		svc := newRestoreSvc(t, "http://unused.invalid")
		_, err := svc.Delete(context.Background(), "", "backup-123", "restore-123", nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty backup ID", func(t *testing.T) {
		svc := newRestoreSvc(t, "http://unused.invalid")
		_, err := svc.Delete(context.Background(), "test-project", "", "restore-123", nil)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
	})

	t.Run("empty restore ID", func(t *testing.T) {
		svc := newRestoreSvc(t, "http://unused.invalid")
		_, err := svc.Delete(context.Background(), "test-project", "backup-123", "", nil)
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
		svc := newRestoreSvc(t, server.URL)
		resp, err := svc.Delete(context.Background(), "test-project", "backup-123", "restore-123", nil)
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
		svc := newRestoreSvc(t, server.URL)
		resp, err := svc.Delete(context.Background(), "test-project", "backup-123", "restore-123", nil)
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
		svc := NewRestoreClientImpl(c, NewBackupClientImpl(c))
		_, err := svc.Delete(context.Background(), "test-project", "backup-123", "restore-123", nil)
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
		svc := newRestoreSvc(t, server.URL)
		resp, err := svc.Delete(context.Background(), "test-project", "backup-123", "restore-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", resp.StatusCode)
		}
	})
}

func TestNewRestoreClientImpl_panicsOnNilBackupClient(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on nil backupClient but got none")
		}
		if !strings.Contains(fmt.Sprint(r), "backupClient") {
			t.Fatalf("expected panic message to mention backupClient, got: %v", r)
		}
	}()
	NewRestoreClientImpl(nil, nil)
}

func TestValidateStorageRestore(t *testing.T) {
	t.Run("empty project", func(t *testing.T) {
		err := types.ValidateStorageRestore("", "backup-123", nil)
		if err == nil {
			t.Fatal("expected error for empty project, got nil")
		}
	})

	t.Run("empty backup ID", func(t *testing.T) {
		err := types.ValidateStorageRestore("test-project", "", nil)
		if err == nil {
			t.Fatal("expected error for empty backup ID, got nil")
		}
	})

	t.Run("nil restoreID is valid", func(t *testing.T) {
		err := types.ValidateStorageRestore("test-project", "backup-123", nil)
		if err != nil {
			t.Fatalf("expected nil error for nil restoreID, got %v", err)
		}
	})

	t.Run("empty-string-pointer restoreID", func(t *testing.T) {
		err := types.ValidateStorageRestore("test-project", "backup-123", ptr.To(""))
		if err == nil {
			t.Fatal("expected error for empty-string-pointer restoreID, got nil")
		}
	})

	t.Run("non-empty restoreID is valid", func(t *testing.T) {
		err := types.ValidateStorageRestore("test-project", "backup-123", ptr.To("restore-123"))
		if err != nil {
			t.Fatalf("expected nil error for non-empty restoreID, got %v", err)
		}
	})
}
