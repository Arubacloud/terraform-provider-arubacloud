package compute

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

func TestListKeyPairs(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			resp := types.KeyPairListResponse{
				ListResponse: types.ListResponse{Total: 1},
				Values: []types.KeyPairResponse{
					{Metadata: types.ResourceMetadataResponse{Name: ptr.To("my-keypair")}},
				},
			}
			json.NewEncoder(w).Encode(resp)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKeyPairsClientImpl(c)

		resp, err := svc.List(context.Background(), "test-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Data.Total != 1 {
			t.Errorf("expected total 1, got %d", resp.Data.Total)
		}
		if len(resp.Data.Values) != 1 {
			t.Errorf("expected 1 keypair, got %d", len(resp.Data.Values))
		}
	})

	t.Run("empty project ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewKeyPairsClientImpl(c)
		_, err := svc.List(context.Background(), "", nil)
		if err == nil {
			t.Error("expected error for empty project ID")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKeyPairsClientImpl(c)
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
		svc := NewKeyPairsClientImpl(c)
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
		svc := NewKeyPairsClientImpl(c)
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
		svc := NewKeyPairsClientImpl(c)
		if _, err := svc.List(context.Background(), "test-project", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetKeyPair(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			resp := types.KeyPairResponse{
				Metadata: types.ResourceMetadataResponse{Name: ptr.To("my-keypair")},
			}
			json.NewEncoder(w).Encode(resp)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKeyPairsClientImpl(c)

		resp, err := svc.Get(context.Background(), "test-project", "keypair-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "my-keypair" {
			t.Errorf("expected name 'my-keypair', got '%v'", resp.Data.Metadata.Name)
		}
	})

	t.Run("empty keypair ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewKeyPairsClientImpl(c)
		_, err := svc.Get(context.Background(), "test-project", "", nil)
		if err == nil {
			t.Error("expected error for empty keypair ID")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKeyPairsClientImpl(c)
		resp, err := svc.Get(context.Background(), "test-project", "keypair-123", nil)
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
		svc := NewKeyPairsClientImpl(c)
		resp, err := svc.Get(context.Background(), "test-project", "keypair-123", nil)
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
		svc := NewKeyPairsClientImpl(c)
		_, err := svc.Get(context.Background(), "test-project", "keypair-123", nil)
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
		svc := NewKeyPairsClientImpl(c)
		if _, err := svc.Get(context.Background(), "test-project", "keypair-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCreateKeyPair(t *testing.T) {
	req := types.KeyPairRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: types.ResourceMetadataRequest{Name: "new-keypair"},
			Location:                types.LocationRequest{Value: "ITBG-Bergamo"},
		},
		Properties: types.KeyPairPropertiesRequest{Value: "ssh-rsa AAAAB3Nza..."},
	}

	t.Run("successful create", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"id":"res-123","uri":"/projects/test-project/providers/Aruba.Compute/keyPairs/res-123","name":"new-keypair"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKeyPairsClientImpl(c)

		resp, err := svc.Create(context.Background(), "test-project", req, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("expected status 201, got %d", resp.StatusCode)
		}
		if resp.Data.Metadata.ID == nil || *resp.Data.Metadata.ID != "res-123" {
			t.Errorf("expected ID 'res-123', got %v", resp.Data.Metadata.ID)
		}
		if resp.Data.Metadata.URI == nil || *resp.Data.Metadata.URI != "/projects/test-project/providers/Aruba.Compute/keyPairs/res-123" {
			t.Errorf("expected URI, got %v", resp.Data.Metadata.URI)
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "new-keypair" {
			t.Errorf("expected name 'new-keypair', got %v", resp.Data.Metadata.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKeyPairsClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", req, nil)
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
		// TODO(TD-010): Create's manual response build silently swallows non-JSON unmarshal
		// errors (diverges from ParseResponseBody which logs at DEBUG).
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, "Bad Gateway")
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKeyPairsClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", req, nil)
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
		svc := NewKeyPairsClientImpl(c)
		_, err := svc.Create(context.Background(), "test-project", req, nil)
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
		svc := NewKeyPairsClientImpl(c)
		if _, err := svc.Create(context.Background(), "test-project", req, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("successful create missing id", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"uri":"/projects/test-project/providers/Aruba.Compute/keyPairs/res-123","name":"new-keypair"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKeyPairsClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.KeyPairRequest{}, nil)
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
			fmt.Fprint(w, `{"metadata":{"id":"res-123","uri":"/projects/test-project/providers/Aruba.Compute/keyPairs/res-123"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKeyPairsClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.KeyPairRequest{}, nil)
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

func TestDeleteKeyPair(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKeyPairsClientImpl(c)

		_, err := svc.Delete(context.Background(), "test-project", "keypair-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKeyPairsClientImpl(c)
		resp, err := svc.Delete(context.Background(), "test-project", "keypair-123", nil)
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
		svc := NewKeyPairsClientImpl(c)
		resp, err := svc.Delete(context.Background(), "test-project", "keypair-123", nil)
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
		svc := NewKeyPairsClientImpl(c)
		_, err := svc.Delete(context.Background(), "test-project", "keypair-123", nil)
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
		svc := NewKeyPairsClientImpl(c)
		if _, err := svc.Delete(context.Background(), "test-project", "keypair-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
