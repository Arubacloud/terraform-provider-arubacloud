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

func TestListContainerRegistry(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && r.URL.Path == "/projects/test-project/providers/Aruba.Container/registries" {
				w.WriteHeader(http.StatusOK)
				resp := types.ContainerRegistryListResponse{
					ListResponse: types.ListResponse{Total: 1},
					Values: []types.ContainerRegistryResponse{
						{
							Metadata: types.ResourceMetadataResponse{
								Name: ptr.To("test-registry"),
							},
							Properties: types.ContainerRegistryPropertiesResponse{
								VPC: types.ReferenceResourceCommon{
									URI: *ptr.To("/projects/test-project/providers/Aruba.Network/vpcs/vpc-1"),
								},
								Subnet: types.ReferenceResourceCommon{
									URI: *ptr.To("/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/subnets/subnet-124"),
								},
								SecurityGroup: types.ReferenceResourceCommon{
									URI: *ptr.To("/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-67890"),
								},
								PublicIp: types.ReferenceResourceCommon{
									URI: *ptr.To("/projects/test-project/providers/Aruba.Network/elasticips/eip-12345"),
								},
								BlockStorage: types.ReferenceResourceCommon{
									URI: *ptr.To("/projects/test-project/providers/Aruba.Storage/blockStorages/bs-54321"),
								},
								BillingPlanCommon: func() *types.BillingPlanCommon {
									v := types.BillingPeriodHour
									return &types.BillingPlanCommon{BillingPeriod: &v}
								}(),
								AdminUser: &types.UserCredentialCommon{
									Username: "admin",
								},
								ConcurrentUsers: ptr.To("100"),
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
		svc := NewContainerRegistryClientImpl(c)

		resp, err := svc.List(context.Background(), "test-project", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatalf("resp is nil")
		}
		if resp.Data == nil {
			t.Fatalf("resp.Data is nil")
		}
		if resp.Data.Values == nil {
			t.Fatalf("resp.Data.Values is nil")
		}
		if len(resp.Data.Values) != 1 {
			t.Errorf("expected 1 Container Registry")
		}
		if resp.Data.Values[0].Metadata.Name == nil || *resp.Data.Values[0].Metadata.Name != "test-registry" {
			t.Errorf("expected name 'test-registry'")
		}
		if resp.Data.Values[0].Properties.PublicIp.URI != "/projects/test-project/providers/Aruba.Network/elasticips/eip-12345" {
			t.Errorf("expected PublicIp URI")
		}
		if resp.Data.Values[0].Properties.VPC.URI != "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1" {
			t.Errorf("expected VPC URI")
		}
		if resp.Data.Values[0].Properties.Subnet.URI != "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/subnets/subnet-124" {
			t.Errorf("expected Subnet URI")
		}
		if resp.Data.Values[0].Properties.SecurityGroup.URI != "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-67890" {
			t.Errorf("expected SecurityGroup URI")
		}
		if resp.Data.Values[0].Properties.BlockStorage.URI != "/projects/test-project/providers/Aruba.Storage/blockStorages/bs-54321" {
			t.Errorf("expected BlockStorage URI")
		}
		if resp.Data.Values[0].Properties.BillingPlanCommon == nil || resp.Data.Values[0].Properties.BillingPlanCommon.BillingPeriod == nil || *resp.Data.Values[0].Properties.BillingPlanCommon.BillingPeriod != "Hour" {
			t.Errorf("expected BillingPlanCommon.BillingPeriod Hour")
		}
		if resp.Data.Values[0].Properties.AdminUser == nil || resp.Data.Values[0].Properties.AdminUser.Username != "admin" {
			t.Errorf("expected AdminUser username")
		}
		if resp.Data.Values[0].Properties.ConcurrentUsers == nil || *resp.Data.Values[0].Properties.ConcurrentUsers != "100" {
			t.Errorf("expected ConcurrentUsers 100")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewContainerRegistryClientImpl(c)

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
		svc := NewContainerRegistryClientImpl(c)

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
		svc := NewContainerRegistryClientImpl(c)

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
		svc := NewContainerRegistryClientImpl(c)

		if _, err := svc.List(context.Background(), "test-project", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGetContainerRegistry(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && r.URL.Path == "/projects/test-project/providers/Aruba.Container/registries/registry-123" {
				w.WriteHeader(http.StatusOK)
				resp := types.ContainerRegistryResponse{
					Metadata: types.ResourceMetadataResponse{
						Name: ptr.To("test-registry"),
						ID:   ptr.To("registry-123"),
					},
					Properties: types.ContainerRegistryPropertiesResponse{
						VPC: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Network/vpcs/vpc-1"),
						},
						Subnet: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/subnets/subnet-124"),
						},
						SecurityGroup: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-67890"),
						},
						PublicIp: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Network/elasticips/eip-12345"),
						},
						BlockStorage: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Storage/blockStorages/bs-54321"),
						},
						BillingPlanCommon: func() *types.BillingPlanCommon {
							v := types.BillingPeriodHour
							return &types.BillingPlanCommon{BillingPeriod: &v}
						}(),
						AdminUser: &types.UserCredentialCommon{
							Username: "admin",
						},
						ConcurrentUsers: ptr.To("100"),
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
		svc := NewContainerRegistryClientImpl(c)

		resp, err := svc.Get(context.Background(), "test-project", "registry-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "test-registry" {
			t.Errorf("expected name 'test-registry'")
		}
		if resp.Data.Properties.PublicIp.URI != "/projects/test-project/providers/Aruba.Network/elasticips/eip-12345" {
			t.Errorf("expected PublicIp URI")
		}
		if resp.Data.Properties.VPC.URI != "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1" {
			t.Errorf("expected VPC URI")
		}
		if resp.Data.Properties.Subnet.URI != "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/subnets/subnet-124" {
			t.Errorf("expected Subnet URI")
		}
		if resp.Data.Properties.SecurityGroup.URI != "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-67890" {
			t.Errorf("expected SecurityGroup URI")
		}
		if resp.Data.Properties.BlockStorage.URI != "/projects/test-project/providers/Aruba.Storage/blockStorages/bs-54321" {
			t.Errorf("expected BlockStorage URI")
		}
		if resp.Data.Properties.BillingPlanCommon == nil || resp.Data.Properties.BillingPlanCommon.BillingPeriod == nil || *resp.Data.Properties.BillingPlanCommon.BillingPeriod != "Hour" {
			t.Errorf("expected BillingPlanCommon.BillingPeriod Hour")
		}
		if resp.Data.Properties.AdminUser == nil || resp.Data.Properties.AdminUser.Username != "admin" {
			t.Errorf("expected AdminUser username")
		}
		if resp.Data.Properties.ConcurrentUsers == nil || *resp.Data.Properties.ConcurrentUsers != "100" {
			t.Errorf("expected ConcurrentUsers 100")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewContainerRegistryClientImpl(c)

		resp, err := svc.Get(context.Background(), "test-project", "registry-123", nil)
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
		svc := NewContainerRegistryClientImpl(c)

		resp, err := svc.Get(context.Background(), "test-project", "registry-123", nil)
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
		svc := NewContainerRegistryClientImpl(c)

		_, err := svc.Get(context.Background(), "test-project", "registry-123", nil)
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
		svc := NewContainerRegistryClientImpl(c)

		if _, err := svc.Get(context.Background(), "test-project", "registry-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCreateContainerRegistry(t *testing.T) {
	t.Run("successful create", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && r.URL.Path == "/projects/test-project/providers/Aruba.Container/registries" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				resp := types.ContainerRegistryResponse{
					Metadata: types.ResourceMetadataResponse{
						Name: ptr.To("new-registry"),
						ID:   ptr.To("registry-456"),
						URI:  ptr.To("/projects/test-project/providers/Aruba.Container/registries/registry-456"),
					},
					Properties: types.ContainerRegistryPropertiesResponse{
						VPC: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Network/vpcs/vpc-1"),
						},
						Subnet: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/subnets/subnet-124"),
						},
						SecurityGroup: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-67890"),
						},
						PublicIp: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Network/elasticips/eip-12345"),
						},
						BlockStorage: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Storage/blockStorages/bs-54321"),
						},
						BillingPlanCommon: func() *types.BillingPlanCommon {
							v := types.BillingPeriodHour
							return &types.BillingPlanCommon{BillingPeriod: &v}
						}(),
						AdminUser: &types.UserCredentialCommon{
							Username: "admin",
						},
						ConcurrentUsers: ptr.To("100"),
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
		svc := NewContainerRegistryClientImpl(c)

		body := types.ContainerRegistryRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: "new-registry",
				},
			},
			Properties: types.ContainerRegistryPropertiesRequest{
				PublicIp:      types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Network/elasticips/eip-12345"},
				VPC:           types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1"},
				Subnet:        types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/subnets/subnet-124"},
				SecurityGroup: types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-67890"},
				BlockStorage:  types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Storage/blockStorages/bs-54321"},
				BillingPlanCommon: func() *types.BillingPlanCommon {
					v := types.BillingPeriod("Hour")
					return &types.BillingPlanCommon{BillingPeriod: &v}
				}(),
				AdminUser:       &types.UserCredentialCommon{Username: "admin"},
				ConcurrentUsers: ptr.To("100"),
			},
		}

		resp, err := svc.Create(context.Background(), "test-project", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "new-registry" {
			t.Errorf("expected name 'new-registry'")
		}
		if resp.Data.Properties.PublicIp.URI != "/projects/test-project/providers/Aruba.Network/elasticips/eip-12345" {
			t.Errorf("expected PublicIp URI")
		}
		if resp.Data.Properties.VPC.URI != "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1" {
			t.Errorf("expected VPC URI")
		}
		if resp.Data.Properties.Subnet.URI != "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/subnets/subnet-124" {
			t.Errorf("expected Subnet URI")
		}
		if resp.Data.Properties.SecurityGroup.URI != "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-67890" {
			t.Errorf("expected SecurityGroup URI")
		}
		if resp.Data.Properties.BlockStorage.URI != "/projects/test-project/providers/Aruba.Storage/blockStorages/bs-54321" {
			t.Errorf("expected BlockStorage URI")
		}
		if resp.Data.Properties.BillingPlanCommon == nil || resp.Data.Properties.BillingPlanCommon.BillingPeriod == nil || *resp.Data.Properties.BillingPlanCommon.BillingPeriod != "Hour" {
			t.Errorf("expected BillingPlanCommon.BillingPeriod Hour")
		}
		if resp.Data.Properties.AdminUser == nil || resp.Data.Properties.AdminUser.Username != "admin" {
			t.Errorf("expected AdminUser username")
		}
		if resp.Data.Properties.ConcurrentUsers == nil || *resp.Data.Properties.ConcurrentUsers != "100" {
			t.Errorf("expected ConcurrentUsers 100")
		}
		if resp.Data.Metadata.ID == nil || *resp.Data.Metadata.ID != "registry-456" {
			t.Errorf("expected ID 'registry-456', got %v", resp.Data.Metadata.ID)
		}
		if resp.Data.Metadata.URI == nil || *resp.Data.Metadata.URI != "/projects/test-project/providers/Aruba.Container/registries/registry-456" {
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
		svc := NewContainerRegistryClientImpl(c)

		body := types.ContainerRegistryRequest{}
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
		svc := NewContainerRegistryClientImpl(c)

		body := types.ContainerRegistryRequest{}
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
		svc := NewContainerRegistryClientImpl(c)

		body := types.ContainerRegistryRequest{}
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
		svc := NewContainerRegistryClientImpl(c)

		body := types.ContainerRegistryRequest{}
		if _, err := svc.Create(context.Background(), "test-project", body, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("successful create missing id", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"metadata":{"uri":"/projects/test-project/providers/Aruba.Container/registries/res-123","name":"test-name"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewContainerRegistryClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.ContainerRegistryRequest{}, nil)
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
			fmt.Fprint(w, `{"metadata":{"id":"res-123","uri":"/projects/test-project/providers/Aruba.Container/registries/res-123"}}`)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewContainerRegistryClientImpl(c)
		resp, err := svc.Create(context.Background(), "test-project", types.ContainerRegistryRequest{}, nil)
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

func TestUpdateContainerRegistry(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "PUT" && r.URL.Path == "/projects/test-project/providers/Aruba.Container/registries/registry-123" {
				w.WriteHeader(http.StatusOK)
				resp := types.ContainerRegistryResponse{
					Metadata: types.ResourceMetadataResponse{
						Name: ptr.To("updated-registry"),
						ID:   ptr.To("registry-123"),
					},
					Properties: types.ContainerRegistryPropertiesResponse{
						VPC: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Network/vpcs/vpc-1"),
						},
						Subnet: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/subnets/subnet-124"),
						},
						SecurityGroup: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-67890"),
						},
						PublicIp: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Network/elasticips/eip-12345"),
						},
						BlockStorage: types.ReferenceResourceCommon{
							URI: *ptr.To("/projects/test-project/providers/Aruba.Storage/blockStorages/bs-54321"),
						},
						BillingPlanCommon: func() *types.BillingPlanCommon {
							v := types.BillingPeriodHour
							return &types.BillingPlanCommon{BillingPeriod: &v}
						}(),
						AdminUser: &types.UserCredentialCommon{
							Username: "admin",
						},
						ConcurrentUsers: ptr.To("100"),
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
		svc := NewContainerRegistryClientImpl(c)

		body := types.ContainerRegistryRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: "updated-registry",
				},
			},
			Properties: types.ContainerRegistryPropertiesRequest{
				PublicIp:      types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Network/elasticips/eip-12345"},
				VPC:           types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1"},
				Subnet:        types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/subnets/subnet-124"},
				SecurityGroup: types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-67890"},
				BlockStorage:  types.ReferenceResourceCommon{URI: "/projects/test-project/providers/Aruba.Storage/blockStorages/bs-54321"},
				BillingPlanCommon: func() *types.BillingPlanCommon {
					v := types.BillingPeriod("Hour")
					return &types.BillingPlanCommon{BillingPeriod: &v}
				}(),
				AdminUser:       &types.UserCredentialCommon{Username: "admin"},
				ConcurrentUsers: ptr.To("100"),
			},
		}

		resp, err := svc.Update(context.Background(), "test-project", "registry-123", body, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.Data == nil {
			t.Fatalf("expected response data")
		}
		if resp.Data.Metadata.Name == nil || *resp.Data.Metadata.Name != "updated-registry" {
			t.Errorf("expected name 'updated-registry'")
		}
		if resp.Data.Properties.PublicIp.URI != "/projects/test-project/providers/Aruba.Network/elasticips/eip-12345" {
			t.Errorf("expected PublicIp URI")
		}
		if resp.Data.Properties.VPC.URI != "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1" {
			t.Errorf("expected VPC URI")
		}
		if resp.Data.Properties.Subnet.URI != "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/subnets/subnet-124" {
			t.Errorf("expected Subnet URI")
		}
		if resp.Data.Properties.SecurityGroup.URI != "/projects/test-project/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-67890" {
			t.Errorf("expected SecurityGroup URI")
		}
		if resp.Data.Properties.BlockStorage.URI != "/projects/test-project/providers/Aruba.Storage/blockStorages/bs-54321" {
			t.Errorf("expected BlockStorage URI")
		}
		if resp.Data.Properties.BillingPlanCommon == nil || resp.Data.Properties.BillingPlanCommon.BillingPeriod == nil || *resp.Data.Properties.BillingPlanCommon.BillingPeriod != "Hour" {
			t.Errorf("expected BillingPlanCommon.BillingPeriod Hour")
		}
		if resp.Data.Properties.AdminUser == nil || resp.Data.Properties.AdminUser.Username != "admin" {
			t.Errorf("expected AdminUser username")
		}
		if resp.Data.Properties.ConcurrentUsers == nil || *resp.Data.Properties.ConcurrentUsers != "100" {
			t.Errorf("expected ConcurrentUsers 100")
		}
		if resp.Data.Status.State == nil || *resp.Data.Status.State != "updating" {
			t.Errorf("expected state 'updating'")
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "resource not found", 404))
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewContainerRegistryClientImpl(c)

		body := types.ContainerRegistryRequest{}
		resp, err := svc.Update(context.Background(), "test-project", "registry-123", body, nil)
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
		svc := NewContainerRegistryClientImpl(c)

		body := types.ContainerRegistryRequest{}
		resp, err := svc.Update(context.Background(), "test-project", "registry-123", body, nil)
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
		svc := NewContainerRegistryClientImpl(c)

		body := types.ContainerRegistryRequest{}
		_, err := svc.Update(context.Background(), "test-project", "registry-123", body, nil)
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
		svc := NewContainerRegistryClientImpl(c)

		body := types.ContainerRegistryRequest{}
		if _, err := svc.Update(context.Background(), "test-project", "registry-123", body, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestDeleteContainerRegistry(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "DELETE" && r.URL.Path == "/projects/test-project/providers/Aruba.Container/registries/registry-123" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			http.NotFound(w, r)
		})
		c := testutil.NewClient(t, server.URL)
		svc := NewContainerRegistryClientImpl(c)

		_, err := svc.Delete(context.Background(), "test-project", "registry-123", nil)
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
		svc := NewContainerRegistryClientImpl(c)

		resp, err := svc.Delete(context.Background(), "test-project", "registry-123", nil)
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
		svc := NewContainerRegistryClientImpl(c)

		resp, err := svc.Delete(context.Background(), "test-project", "registry-123", nil)
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
		svc := NewContainerRegistryClientImpl(c)

		_, err := svc.Delete(context.Background(), "test-project", "registry-123", nil)
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
		svc := NewContainerRegistryClientImpl(c)

		if _, err := svc.Delete(context.Background(), "test-project", "registry-123", nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
