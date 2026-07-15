package compute

import (
	"context"
	"encoding/base64"
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

func TestListCloudServers(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		apiCalled := false
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			apiCalled = true
			w.WriteHeader(http.StatusOK)
			resp := types.CloudServerListResponse{
				ListResponse: types.ListResponse{Total: 2},
				Values: []types.CloudServerResponse{
					{Metadata: types.ResourceMetadataResponse{Name: ptr.To("server-1")}},
				},
			}
			json.NewEncoder(w).Encode(resp)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)

		resp, err := svc.List(context.Background(), "test-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !apiCalled {
			t.Error("API endpoint was not called")
		}
		if resp.Data.Total != 2 {
			t.Errorf("expected total 2, got %d", resp.Data.Total)
		}
	})

	t.Run("empty project ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewCloudServersClientImpl(c)
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
		svc := NewCloudServersClientImpl(c)
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
		svc := NewCloudServersClientImpl(c)
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
		svc := NewCloudServersClientImpl(c)
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
		svc := NewCloudServersClientImpl(c)
		if _, err := svc.List(context.Background(), "test-project", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetCloudServer(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			resp := types.CloudServerResponse{
				Metadata:   types.ResourceMetadataResponse{Name: ptr.To("my-server")},
				Properties: types.CloudServerPropertiesResponse{Zone: "ITBG-1"},
			}
			json.NewEncoder(w).Encode(resp)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)

		resp, err := svc.Get(context.Background(), "test-project", "server-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "my-server" {
			t.Errorf("expected name 'my-server', got '%v'", resp.Data.Metadata.Name)
		}
	})

	t.Run("empty project ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewCloudServersClientImpl(c)
		_, err := svc.Get(context.Background(), "", "server-123", nil)
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
		svc := NewCloudServersClientImpl(c)
		resp, err := svc.Get(context.Background(), "test-project", "server-123", nil)
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
		svc := NewCloudServersClientImpl(c)
		resp, err := svc.Get(context.Background(), "test-project", "server-123", nil)
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
		svc := NewCloudServersClientImpl(c)
		_, err := svc.Get(context.Background(), "test-project", "server-123", nil)
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
		svc := NewCloudServersClientImpl(c)
		if _, err := svc.Get(context.Background(), "test-project", "server-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCloudServerRequestOmitsOptionalFields(t *testing.T) {
	t.Run("omits optional fields", func(t *testing.T) {
		req := types.CloudServerRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{Name: "test-server"},
				Location:                types.LocationRequest{Value: "ITBG-Bergamo"},
			},
			Properties: types.CloudServerPropertiesRequest{
				Zone: "ITBG-1",
				VPC:  types.ReferenceResourceCommon{URI: "/vpcs/123"},
				BootVolume: types.ReferenceResourceCommon{
					URI: "/blockStorages/456",
				},
			},
		}

		b, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}

		var envelope map[string]any
		if err := json.Unmarshal(b, &envelope); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		props, _ := envelope["properties"].(map[string]any)
		if _, ok := props["elasticIp"]; ok {
			t.Errorf("expected elasticIp to be omitted when nil, got: %s", b)
		}
		if _, ok := props["keyPair"]; ok {
			t.Errorf("expected keyPair to be omitted when nil, got: %s", b)
		}
	})
}

func TestCreateCloudServer(t *testing.T) {
	cloudInitContent := "#cloud-config\npackage_update: true\n"
	userData := base64.StdEncoding.EncodeToString([]byte(cloudInitContent))
	req := types.CloudServerRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: types.ResourceMetadataRequest{Name: "new-server"},
			Location:                types.LocationRequest{Value: "ITBG-Bergamo"},
		},
		Properties: types.CloudServerPropertiesRequest{
			Zone:     "ITBG-1",
			UserData: ptr.To(userData),
		},
	}

	t.Run("successful create", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"id":"res-123","uri":"/projects/test-project/providers/Aruba.Compute/cloudServers/res-123","name":"new-server"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)

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
		if resp.Data.Metadata.URI == nil || *resp.Data.Metadata.URI != "/projects/test-project/providers/Aruba.Compute/cloudServers/res-123" {
			t.Errorf("expected URI, got %v", resp.Data.Metadata.URI)
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "new-server" {
			t.Errorf("expected name 'new-server', got %v", resp.Data.Metadata.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)
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
		svc := NewCloudServersClientImpl(c)
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
		svc := NewCloudServersClientImpl(c)
		_, err := svc.Create(context.Background(), "test-project", req, nil)
		if err == nil {
			t.Fatal("expected a network error, got nil")
		}
	})

	t.Run("nil params injects default api-version", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("api-version"); got != "1.1" {
				t.Errorf("expected api-version=1.1, got %q", got)
			}
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"id":"x","uri":"/x","name":"x"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)
		if _, err := svc.Create(context.Background(), "test-project", req, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("successful create missing id", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"uri":"/projects/test-project/providers/Aruba.Compute/cloudServers/res-123","name":"new-server"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.CloudServerRequest{}, nil)
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
			fmt.Fprint(w, `{"metadata":{"id":"res-123","uri":"/projects/test-project/providers/Aruba.Compute/cloudServers/res-123"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.CloudServerRequest{}, nil)
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

func TestDeleteCloudServer(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)

		_, err := svc.Delete(context.Background(), "test-project", "server-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty project ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewCloudServersClientImpl(c)
		_, err := svc.Delete(context.Background(), "", "server-123", nil)
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
		svc := NewCloudServersClientImpl(c)
		resp, err := svc.Delete(context.Background(), "test-project", "server-123", nil)
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
		svc := NewCloudServersClientImpl(c)
		resp, err := svc.Delete(context.Background(), "test-project", "server-123", nil)
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
		svc := NewCloudServersClientImpl(c)
		_, err := svc.Delete(context.Background(), "test-project", "server-123", nil)
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
		svc := NewCloudServersClientImpl(c)
		if _, err := svc.Delete(context.Background(), "test-project", "server-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestPowerOnCloudServer(t *testing.T) {
	t.Run("successful power on", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/projects/test-project/providers/Aruba.Compute/cloudServers/server-123/poweron" {
				t.Errorf("expected poweron path, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			resp := types.CloudServerResponse{
				Metadata:   types.ResourceMetadataResponse{Name: ptr.To("my-server")},
				Properties: types.CloudServerPropertiesResponse{Zone: "ITBG-1"},
				Status:     types.ResourceStatusResponse{State: statePtr(types.State("active"))},
			}
			json.NewEncoder(w).Encode(resp)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)

		resp, err := svc.PowerOn(context.Background(), "test-project", "server-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "my-server" {
			t.Errorf("expected name 'my-server', got '%v'", resp.Data.Metadata.Name)
		}
		if !resp.IsSuccess() {
			t.Errorf("expected successful response, got status code %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)
		resp, err := svc.PowerOn(context.Background(), "test-project", "server-123", nil)
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
		svc := NewCloudServersClientImpl(c)
		resp, err := svc.PowerOn(context.Background(), "test-project", "server-123", nil)
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
		svc := NewCloudServersClientImpl(c)
		_, err := svc.PowerOn(context.Background(), "test-project", "server-123", nil)
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
		svc := NewCloudServersClientImpl(c)
		if _, err := svc.PowerOn(context.Background(), "test-project", "server-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestPowerOffCloudServer(t *testing.T) {
	t.Run("successful power off", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/projects/test-project/providers/Aruba.Compute/cloudServers/server-123/poweroff" {
				t.Errorf("expected poweroff path, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			resp := types.CloudServerResponse{
				Metadata:   types.ResourceMetadataResponse{Name: ptr.To("my-server")},
				Properties: types.CloudServerPropertiesResponse{Zone: "ITBG-1"},
				Status:     types.ResourceStatusResponse{State: statePtr(types.State("stopped"))},
			}
			json.NewEncoder(w).Encode(resp)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)

		resp, err := svc.PowerOff(context.Background(), "test-project", "server-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "my-server" {
			t.Errorf("expected name 'my-server', got '%v'", resp.Data.Metadata.Name)
		}
		if !resp.IsSuccess() {
			t.Errorf("expected successful response, got status code %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)
		resp, err := svc.PowerOff(context.Background(), "test-project", "server-123", nil)
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
		svc := NewCloudServersClientImpl(c)
		resp, err := svc.PowerOff(context.Background(), "test-project", "server-123", nil)
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
		svc := NewCloudServersClientImpl(c)
		_, err := svc.PowerOff(context.Background(), "test-project", "server-123", nil)
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
		svc := NewCloudServersClientImpl(c)
		if _, err := svc.PowerOff(context.Background(), "test-project", "server-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestSetPasswordCloudServer(t *testing.T) {
	req := types.CloudServerPasswordRequest{Password: "newPassword123"}

	t.Run("successful set password", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/projects/test-project/providers/Aruba.Compute/cloudServers/server-123/password" {
				t.Errorf("expected password path, got %s", r.URL.Path)
			}
			var reqBody types.CloudServerPasswordRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("failed to decode request body: %v", err)
			}
			if reqBody.Password != "newPassword123" {
				t.Errorf("expected password 'newPassword123', got '%s'", reqBody.Password)
			}
			w.WriteHeader(http.StatusOK)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)

		resp, err := svc.SetPassword(context.Background(), "test-project", "server-123", req, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resp.IsSuccess() {
			t.Errorf("expected successful response, got status code %d", resp.StatusCode)
		}
	})

	t.Run("empty project ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewCloudServersClientImpl(c)
		_, err := svc.SetPassword(context.Background(), "", "server-123", req, nil)
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
		svc := NewCloudServersClientImpl(c)
		resp, err := svc.SetPassword(context.Background(), "test-project", "server-123", req, nil)
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
		svc := NewCloudServersClientImpl(c)
		resp, err := svc.SetPassword(context.Background(), "test-project", "server-123", req, nil)
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
		svc := NewCloudServersClientImpl(c)
		_, err := svc.SetPassword(context.Background(), "test-project", "server-123", req, nil)
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
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)
		if _, err := svc.SetPassword(context.Background(), "test-project", "server-123", req, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAssociateSubnetsCloudServer(t *testing.T) {
	body := types.CloudServerAssociateSubnetsRequest{
		SubnetsToAssociate:    []types.ReferenceResourceCommon{{URI: "/subnets/s-1"}},
		SubnetsToDisassociate: []types.ReferenceResourceCommon{{URI: "/subnets/s-2"}},
	}

	t.Run("successful associate/disassociate", func(t *testing.T) {
		var gotBody types.CloudServerAssociateSubnetsRequest
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/projects/test-project/providers/Aruba.Compute/cloudServers/server-123/associateDisassociateSubnets" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Errorf("decode body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{"metadata":{"id":"cs-1","name":"srv","uri":"/projects/test-project/providers/Aruba.Compute/cloudServers/cs-1"},"properties":{}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)

		resp, err := svc.AssociateSubnets(context.Background(), "test-project", "server-123", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("expected 202, got %d", resp.StatusCode)
		}
		if len(gotBody.SubnetsToAssociate) != 1 || gotBody.SubnetsToAssociate[0].URI != "/subnets/s-1" {
			t.Errorf("SubnetsToAssociate = %v", gotBody.SubnetsToAssociate)
		}
		if len(gotBody.SubnetsToDisassociate) != 1 || gotBody.SubnetsToDisassociate[0].URI != "/subnets/s-2" {
			t.Errorf("SubnetsToDisassociate = %v", gotBody.SubnetsToDisassociate)
		}
	})

	t.Run("empty project ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewCloudServersClientImpl(c)
		_, err := svc.AssociateSubnets(context.Background(), "", "server-123", body, nil)
		if err == nil {
			t.Error("expected error for empty project ID")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)
		resp, err := svc.AssociateSubnets(context.Background(), "test-project", "server-123", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 response")
		}
	})

	t.Run("nil params injects default api-version", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("api-version"); got != "1.0" {
				t.Errorf("expected api-version=1.0, got %q", got)
			}
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)
		if _, err := svc.AssociateSubnets(context.Background(), "test-project", "server-123", body, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAssociateSecurityGroupsCloudServer(t *testing.T) {
	body := types.CloudServerAssociateSecurityGroupsRequest{
		SecurityGroupsToAssociate: []types.ReferenceResourceCommon{{URI: "/sgs/sg-1"}},
	}

	t.Run("successful associate", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/projects/test-project/providers/Aruba.Compute/cloudServers/server-123/associateDisassociateSecurityGroups" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)

		resp, err := svc.AssociateSecurityGroups(context.Background(), "test-project", "server-123", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("expected 202, got %d", resp.StatusCode)
		}
	})

	t.Run("empty server ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewCloudServersClientImpl(c)
		_, err := svc.AssociateSecurityGroups(context.Background(), "test-project", "", body, nil)
		if err == nil {
			t.Error("expected error for empty server ID")
		}
	})
}

func TestAssociateElasticIPsCloudServer(t *testing.T) {
	body := types.CloudServerAssociateElasticIPsRequest{
		ElasticIPsToDisassociate: []types.ReferenceResourceCommon{{URI: "/eips/eip-1"}},
	}

	t.Run("successful disassociate", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/projects/test-project/providers/Aruba.Compute/cloudServers/server-123/associateDisassociateElasticIPs" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)

		resp, err := svc.AssociateElasticIPs(context.Background(), "test-project", "server-123", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("expected 202, got %d", resp.StatusCode)
		}
	})

	t.Run("empty project ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewCloudServersClientImpl(c)
		_, err := svc.AssociateElasticIPs(context.Background(), "", "server-123", body, nil)
		if err == nil {
			t.Error("expected error for empty project ID")
		}
	})
}

func TestAttachDetachDataVolumesCloudServer(t *testing.T) {
	body := types.CloudServerAttachDetachDataVolumesRequest{
		VolumesToAttach: []types.ReferenceResourceCommon{{URI: "/vols/v-1"}},
		VolumesToDetach: []types.ReferenceResourceCommon{{URI: "/vols/v-2"}},
	}

	t.Run("successful attach/detach", func(t *testing.T) {
		var gotBody types.CloudServerAttachDetachDataVolumesRequest
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/projects/test-project/providers/Aruba.Compute/cloudServers/server-123/attachDetachDataVolumes" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Errorf("decode body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewCloudServersClientImpl(c)

		resp, err := svc.AttachDetachDataVolumes(context.Background(), "test-project", "server-123", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusAccepted {
			t.Errorf("expected 202, got %d", resp.StatusCode)
		}
		if len(gotBody.VolumesToAttach) != 1 || gotBody.VolumesToAttach[0].URI != "/vols/v-1" {
			t.Errorf("VolumesToAttach = %v", gotBody.VolumesToAttach)
		}
		if len(gotBody.VolumesToDetach) != 1 || gotBody.VolumesToDetach[0].URI != "/vols/v-2" {
			t.Errorf("VolumesToDetach = %v", gotBody.VolumesToDetach)
		}
	})

	t.Run("empty project ID", func(t *testing.T) {
		c := testutil.NewClient(t, "http://unused.invalid")
		svc := NewCloudServersClientImpl(c)
		_, err := svc.AttachDetachDataVolumes(context.Background(), "", "server-123", body, nil)
		if err == nil {
			t.Error("expected error for empty project ID")
		}
	})

	t.Run("network error", func(t *testing.T) {
		c := testutil.NewBrokenClient(t, "http://unused.invalid")
		svc := NewCloudServersClientImpl(c)
		_, err := svc.AttachDetachDataVolumes(context.Background(), "test-project", "server-123", body, nil)
		if err == nil {
			t.Fatal("expected a network error, got nil")
		}
	})
}
