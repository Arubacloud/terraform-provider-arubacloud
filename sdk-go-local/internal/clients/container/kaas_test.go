package container

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

func TestListKaaS(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && r.URL.Path == "/projects/test-project/providers/Aruba.Container/kaas" {
				w.WriteHeader(http.StatusOK)
				resp := types.KaaSListResponse{
					ListResponse: types.ListResponse{Total: 1},
					Values: []types.KaaSResponse{
						{
							Metadata: types.ResourceMetadataResponse{
								Name: ptr.To("test-kaas"),
							},
							Properties: types.KaaSPropertiesResponse{
								Preset: false,
								HA:     ptr.To(true),
								KubernetesVersion: types.KubernetesVersionInfoResponse{
									Value:       ptr.To("1.28.0"),
									Recommended: true,
								},
							},
							Status: types.ResourceStatusResponse{
								State: ptr.To(types.State("active")),
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
		svc := NewKaaSClientImpl(c)

		resp, err := svc.List(context.Background(), "test-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil || len(resp.Data.Values) != 1 {
			t.Errorf("expected 1 KaaS cluster")
		}
		if resp.Data.Values[0].Metadata.Name == nil || *resp.Data.Values[0].Metadata.Name != "test-kaas" {
			t.Errorf("expected name 'test-kaas'")
		}
		if resp.Data.Values[0].Properties.HA == nil || !*resp.Data.Values[0].Properties.HA {
			t.Errorf("expected HA to be true")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKaaSClientImpl(c)

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
		svc := NewKaaSClientImpl(c)

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
		svc := NewKaaSClientImpl(c)

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
		svc := NewKaaSClientImpl(c)

		if _, err := svc.List(context.Background(), "test-project", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetKaaS(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && r.URL.Path == "/projects/test-project/providers/Aruba.Container/kaas/kaas-123" {
				w.WriteHeader(http.StatusOK)
				resp := types.KaaSResponse{
					Metadata: types.ResourceMetadataResponse{
						Name: ptr.To("test-kaas"),
						ID:   ptr.To("kaas-123"),
					},
					Properties: types.KaaSPropertiesResponse{
						Preset: false,
						HA:     ptr.To(true),
						KubernetesVersion: types.KubernetesVersionInfoResponse{
							Value:       ptr.To("1.28.0"),
							Recommended: true,
						},
						NodePools: func() *[]types.NodePoolPropertiesResponse {
							pools := []types.NodePoolPropertiesResponse{
								{
									Name:        ptr.To("default-pool"),
									Nodes:       ptr.To(int32(3)),
									Instance:    &types.KaaSNodePoolInstanceResponse{Name: ptr.To("small")},
									DataCenter:  &types.KaaSNodePoolDataCenterResponse{Code: ptr.To("dc-01")},
									Autoscaling: false,
								},
							}
							return &pools
						}(),
						ManagementIP: ptr.To("10.0.0.100"),
					},
					Status: types.ResourceStatusResponse{
						State: ptr.To(types.State("active")),
					},
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKaaSClientImpl(c)

		resp, err := svc.Get(context.Background(), "test-project", "kaas-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "test-kaas" {
			t.Errorf("expected name 'test-kaas'")
		}
		if resp.Data.Properties.KubernetesVersion.Value == nil || *resp.Data.Properties.KubernetesVersion.Value != "1.28.0" {
			val := ""
			if resp.Data.Properties.KubernetesVersion.Value != nil {
				val = *resp.Data.Properties.KubernetesVersion.Value
			}
			t.Errorf("expected Kubernetes version '1.28.0', got %s", val)
		}
		if resp.Data.Properties.NodePools == nil || len(*resp.Data.Properties.NodePools) != 1 {
			t.Errorf("expected 1 node pool")
		}
		if resp.Data.Properties.ManagementIP == nil || *resp.Data.Properties.ManagementIP != "10.0.0.100" {
			t.Errorf("expected management IP '10.0.0.100'")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKaaSClientImpl(c)

		resp, err := svc.Get(context.Background(), "test-project", "kaas-123", nil)
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
		svc := NewKaaSClientImpl(c)

		resp, err := svc.Get(context.Background(), "test-project", "kaas-123", nil)
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
		svc := NewKaaSClientImpl(c)

		_, err := svc.Get(context.Background(), "test-project", "kaas-123", nil)
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
		svc := NewKaaSClientImpl(c)

		if _, err := svc.Get(context.Background(), "test-project", "kaas-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCreateKaaS(t *testing.T) {
	t.Run("successful create", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && r.URL.Path == "/projects/test-project/providers/Aruba.Container/kaas" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				resp := types.KaaSResponse{
					Metadata: types.ResourceMetadataResponse{
						Name: ptr.To("new-kaas"),
						ID:   ptr.To("kaas-456"),
						URI:  ptr.To("/projects/test-project/providers/Aruba.Container/kaas/kaas-456"),
					},
					Properties: types.KaaSPropertiesResponse{
						Preset: false,
						HA:     ptr.To(true),
						KubernetesVersion: types.KubernetesVersionInfoResponse{
							Value: ptr.To("1.28.0"),
						},
					},
					Status: types.ResourceStatusResponse{
						State: ptr.To(types.State("creating")),
					},
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKaaSClientImpl(c)

		body := types.KaaSRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: "new-kaas",
				},
			},
			Properties: types.KaaSPropertiesRequest{
				Preset: ptr.To(false),
				HA:     ptr.To(true),
				KubernetesVersion: types.KubernetesVersionInfoRequest{
					Value: "1.28.0",
				},
			},
		}

		resp, err := svc.Create(context.Background(), "test-project", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "new-kaas" {
			t.Errorf("expected name 'new-kaas'")
		}
		if resp.Data.Metadata.ID == nil || *resp.Data.Metadata.ID != "kaas-456" {
			t.Errorf("expected ID 'kaas-456', got %v", resp.Data.Metadata.ID)
		}
		if resp.Data.Metadata.URI == nil || *resp.Data.Metadata.URI != "/projects/test-project/providers/Aruba.Container/kaas/kaas-456" {
			t.Errorf("expected URI, got %v", resp.Data.Metadata.URI)
		}
		if resp.Data.Status.State == nil || *resp.Data.Status.State != "creating" {
			t.Errorf("expected state 'creating'")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKaaSClientImpl(c)

		body := types.KaaSRequest{}
		resp, err := svc.Create(context.Background(), "test-project", body, nil)
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
		svc := NewKaaSClientImpl(c)

		body := types.KaaSRequest{}
		resp, err := svc.Create(context.Background(), "test-project", body, nil)
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
		svc := NewKaaSClientImpl(c)

		body := types.KaaSRequest{}
		_, err := svc.Create(context.Background(), "test-project", body, nil)
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
		svc := NewKaaSClientImpl(c)

		body := types.KaaSRequest{}
		if _, err := svc.Create(context.Background(), "test-project", body, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("successful create missing id", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"uri":"/projects/test-project/providers/Aruba.Container/kaas/res-123","name":"test-name"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKaaSClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.KaaSRequest{}, nil)
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
			fmt.Fprint(w, `{"metadata":{"id":"res-123","uri":"/projects/test-project/providers/Aruba.Container/kaas/res-123"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKaaSClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.KaaSRequest{}, nil)
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

func TestUpdateKaaS(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "PUT" && r.URL.Path == "/projects/test-project/providers/Aruba.Container/kaas/kaas-123" {
				w.WriteHeader(http.StatusOK)
				resp := types.KaaSResponse{
					Metadata: types.ResourceMetadataResponse{
						Name: ptr.To("updated-kaas"),
						ID:   ptr.To("kaas-123"),
					},
					Properties: types.KaaSPropertiesResponse{
						Preset: false,
						HA:     ptr.To(true),
						KubernetesVersion: types.KubernetesVersionInfoResponse{
							Value: ptr.To("1.29.0"),
						},
					},
					Status: types.ResourceStatusResponse{
						State: ptr.To(types.State("updating")),
					},
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKaaSClientImpl(c)

		body := types.KaaSUpdateRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: "updated-kaas",
				},
			},
			Properties: types.KaaSPropertiesUpdateRequest{
				KubernetesVersion: types.KubernetesVersionInfoUpdateRequest{
					Value: "1.29.0",
				},
				NodePools: []types.NodePoolPropertiesRequest{
					{
						Name:     "default-pool",
						Nodes:    3,
						Instance: "K4A8",
						Zone:     "ITBG-1",
					},
				},
				HA: ptr.To(true),
			},
		}

		resp, err := svc.Update(context.Background(), "test-project", "kaas-123", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "updated-kaas" {
			t.Errorf("expected name 'updated-kaas'")
		}
		if resp.Data.Properties.KubernetesVersion.Value == nil || *resp.Data.Properties.KubernetesVersion.Value != "1.29.0" {
			t.Errorf("expected Kubernetes version '1.29.0'")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKaaSClientImpl(c)

		body := types.KaaSUpdateRequest{}
		resp, err := svc.Update(context.Background(), "test-project", "kaas-123", body, nil)
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
		svc := NewKaaSClientImpl(c)

		body := types.KaaSUpdateRequest{}
		resp, err := svc.Update(context.Background(), "test-project", "kaas-123", body, nil)
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
		svc := NewKaaSClientImpl(c)

		body := types.KaaSUpdateRequest{}
		_, err := svc.Update(context.Background(), "test-project", "kaas-123", body, nil)
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
		svc := NewKaaSClientImpl(c)

		body := types.KaaSUpdateRequest{}
		if _, err := svc.Update(context.Background(), "test-project", "kaas-123", body, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestDeleteKaaS(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "DELETE" && r.URL.Path == "/projects/test-project/providers/Aruba.Container/kaas/kaas-123" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKaaSClientImpl(c)

		_, err := svc.Delete(context.Background(), "test-project", "kaas-123", nil)
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
		svc := NewKaaSClientImpl(c)

		resp, err := svc.Delete(context.Background(), "test-project", "kaas-123", nil)
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
		svc := NewKaaSClientImpl(c)

		resp, err := svc.Delete(context.Background(), "test-project", "kaas-123", nil)
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
		svc := NewKaaSClientImpl(c)

		_, err := svc.Delete(context.Background(), "test-project", "kaas-123", nil)
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
		svc := NewKaaSClientImpl(c)

		if _, err := svc.Delete(context.Background(), "test-project", "kaas-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestDownloadKubeconfig(t *testing.T) {
	t.Run("successful download", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && r.URL.Path == "/projects/test-project/providers/Aruba.Container/kaas/kaas-123/download" {
				w.WriteHeader(http.StatusOK)
				resp := types.KaaSKubeconfigResponse{
					Name:    "kubeconfig.yaml",
					Content: "YXBpVmVyc2lvbjogdjEKa2luZDogQ29uZmlnCmNsdXN0ZXJzOgotIGNsdXN0ZXI6CiAgICBjZXJ0aWZpY2F0ZS1hdXRob3JpdHktZGF0YTogTFMwdExTMUNSVWRKVGlCRFJWSlVTVVpKUTBGVVJTMHRMUzB0Q2sxSlNVUkZVeTB4TlMwdExTMHRDazFKU1VSRlV5MHhOUzB0TFMwdENrMUo=",
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewKaaSClientImpl(c)

		resp, err := svc.DownloadKubeconfig(context.Background(), "test-project", "kaas-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Name != "kubeconfig.yaml" {
			t.Errorf("expected filename 'kubeconfig.yaml', got %s", resp.Data.Name)
		}
		if resp.Data.Content == "" {
			t.Errorf("expected non-empty content")
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
		svc := NewKaaSClientImpl(c)

		resp, err := svc.DownloadKubeconfig(context.Background(), "test-project", "kaas-123", nil)
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
		svc := NewKaaSClientImpl(c)

		resp, err := svc.DownloadKubeconfig(context.Background(), "test-project", "kaas-123", nil)
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
		svc := NewKaaSClientImpl(c)

		_, err := svc.DownloadKubeconfig(context.Background(), "test-project", "kaas-123", nil)
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
		svc := NewKaaSClientImpl(c)

		if _, err := svc.DownloadKubeconfig(context.Background(), "test-project", "kaas-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
