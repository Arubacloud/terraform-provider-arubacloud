package aruba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Compile-time interface satisfaction
// --------------------------------------------------------------------------

var (
	_ Ref     = (*ContainerRegistry)(nil)
	_ Wrapper = (*ContainerRegistry)(nil)
)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestContainerRegistry_FluentSetters(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-1", "my-proj", "/projects/p-1"))

	vpcURI := URI("/projects/p-1/providers/Aruba.Network/vpcs/vpc-1")
	subnetURI := URI("/projects/p-1/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1")
	sgURI := URI("/projects/p-1/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-1")
	eipURI := URI("/projects/p-1/providers/Aruba.Network/elasticips/eip-1")
	bsURI := URI("/projects/p-1/providers/Aruba.Storage/blockStorages/bs-1")

	cr := NewContainerRegistry().
		InProject(proj).
		Named("my-registry").
		Tagged("env:prod").
		Tagged("registry").
		Tagged("env:prod"). // dedupe
		InRegion(RegionITBGBergamo).
		WithVPC(vpcURI).
		WithSubnet(subnetURI).
		WithSecurityGroup(sgURI).
		WithElasticIP(eipURI).
		WithBlockStorage(bsURI).
		WithAdminUsername("admin").
		OfSize(ContainerRegistrySizeFlavorSmall).
		BilledBy(BillingPeriodHour)

	if cr.Name() != "my-registry" {
		t.Errorf("Name() = %q", cr.Name())
	}
	if tags := cr.Tags(); len(tags) != 2 || tags[0] != "env:prod" || tags[1] != "registry" {
		t.Errorf("Tags() = %v", tags)
	}
	if cr.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", cr.Region())
	}
	if cr.VPC() != vpcURI.URI() {
		t.Errorf("VPC() = %q", cr.VPC())
	}
	if cr.Subnet() != subnetURI.URI() {
		t.Errorf("Subnet() = %q", cr.Subnet())
	}
	if cr.SecurityGroup() != sgURI.URI() {
		t.Errorf("SecurityGroup() = %q", cr.SecurityGroup())
	}
	if cr.ElasticIP() != eipURI.URI() {
		t.Errorf("PublicIP() = %q", cr.ElasticIP())
	}
	if cr.BlockStorage() != bsURI.URI() {
		t.Errorf("BlockStorage() = %q", cr.BlockStorage())
	}
	if cr.AdminUsername() != "admin" {
		t.Errorf("AdminUsername() = %q", cr.AdminUsername())
	}
	if cr.SizeFlavor() != ContainerRegistrySizeFlavorSmall {
		t.Errorf("SizeFlavor() = %q", cr.SizeFlavor())
	}
	if cr.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", cr.BillingPeriod())
	}
	if cr.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", cr.ProjectID())
	}
	if cr.Err() != nil {
		t.Errorf("Err() = %v", cr.Err())
	}

	cr.Untagged("env:prod")
	if tags := cr.Tags(); len(tags) != 1 || tags[0] != "registry" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	cr.RetaggedAs("x", "y")
	if tags := cr.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

// --------------------------------------------------------------------------
// IntoProject
// --------------------------------------------------------------------------

func TestContainerRegistry_IntoProject_TypedRef(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-42", "n", "/projects/p-42"))
	cr := NewContainerRegistry().InProject(proj)
	if cr.ProjectID() != "p-42" {
		t.Errorf("ProjectID() = %q", cr.ProjectID())
	}
	if cr.Err() != nil {
		t.Errorf("Err() = %v", cr.Err())
	}
}

func TestContainerRegistry_IntoProject_URIRef(t *testing.T) {
	cr := NewContainerRegistry().InProject(URI("/projects/p-uri"))
	if cr.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", cr.ProjectID())
	}
	if cr.Err() != nil {
		t.Errorf("Err() = %v", cr.Err())
	}
}

func TestContainerRegistry_IntoProject_BadRef(t *testing.T) {
	cr := NewContainerRegistry().InProject(URI("/garbage"))
	if cr.Err() == nil {
		t.Error("expected Err() != nil for unresolvable Ref")
	}
}

// --------------------------------------------------------------------------
// WithElasticIP body-ref setter
// --------------------------------------------------------------------------

func TestContainerRegistry_WithElasticIP_URIRef(t *testing.T) {
	uri := "/projects/p-1/providers/Aruba.Network/elasticips/eip-1"
	cr := NewContainerRegistry().WithElasticIP(URI(uri))
	if cr.ElasticIP() != uri {
		t.Errorf("PublicIP() = %q", cr.ElasticIP())
	}
	if cr.Err() != nil {
		t.Errorf("Err() = %v", cr.Err())
	}
}

func TestContainerRegistry_WithElasticIP_TypedRef(t *testing.T) {
	ref := URI("/projects/p-1/providers/Aruba.Network/elasticips/eip-1")
	cr := NewContainerRegistry().WithElasticIP(ref)
	if cr.ElasticIP() != ref.URI() {
		t.Errorf("PublicIP() = %q, want %q", cr.ElasticIP(), ref.URI())
	}
	if cr.Err() != nil {
		t.Errorf("Err() = %v", cr.Err())
	}
}

func TestContainerRegistry_WithElasticIP_EmptyURI(t *testing.T) {
	cr := NewContainerRegistry().WithElasticIP(URI(""))
	if cr.Err() == nil {
		t.Error("expected Err() != nil for empty PublicIP URI")
	}
	if cr.ElasticIP() != "" {
		t.Errorf("PublicIP() should remain empty, got %q", cr.ElasticIP())
	}
}

// --------------------------------------------------------------------------
// WithVPC body-ref setter
// --------------------------------------------------------------------------

func TestContainerRegistry_WithVPC_URIRef(t *testing.T) {
	uri := "/projects/p-1/providers/Aruba.Network/vpcs/vpc-1"
	cr := NewContainerRegistry().WithVPC(URI(uri))
	if cr.VPC() != uri {
		t.Errorf("VPC() = %q", cr.VPC())
	}
	if cr.Err() != nil {
		t.Errorf("Err() = %v", cr.Err())
	}
}

func TestContainerRegistry_WithVPC_TypedRef(t *testing.T) {
	ref := URI("/projects/p-1/providers/Aruba.Network/vpcs/vpc-1")
	cr := NewContainerRegistry().WithVPC(ref)
	if cr.VPC() != ref.URI() {
		t.Errorf("VPC() = %q, want %q", cr.VPC(), ref.URI())
	}
}

func TestContainerRegistry_WithVPC_EmptyURI(t *testing.T) {
	cr := NewContainerRegistry().WithVPC(URI(""))
	if cr.Err() == nil {
		t.Error("expected Err() != nil for empty VPC URI")
	}
	if cr.VPC() != "" {
		t.Errorf("VPC() should remain empty, got %q", cr.VPC())
	}
}

// --------------------------------------------------------------------------
// WithSubnet body-ref setter
// --------------------------------------------------------------------------

func TestContainerRegistry_WithSubnet_URIRef(t *testing.T) {
	uri := "/projects/p-1/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1"
	cr := NewContainerRegistry().WithSubnet(URI(uri))
	if cr.Subnet() != uri {
		t.Errorf("Subnet() = %q", cr.Subnet())
	}
	if cr.Err() != nil {
		t.Errorf("Err() = %v", cr.Err())
	}
}

func TestContainerRegistry_WithSubnet_TypedRef(t *testing.T) {
	ref := URI("/projects/p-1/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1")
	cr := NewContainerRegistry().WithSubnet(ref)
	if cr.Subnet() != ref.URI() {
		t.Errorf("Subnet() = %q, want %q", cr.Subnet(), ref.URI())
	}
}

func TestContainerRegistry_WithSubnet_EmptyURI(t *testing.T) {
	cr := NewContainerRegistry().WithSubnet(URI(""))
	if cr.Err() == nil {
		t.Error("expected Err() != nil for empty Subnet URI")
	}
	if cr.Subnet() != "" {
		t.Errorf("Subnet() should remain empty, got %q", cr.Subnet())
	}
}

// --------------------------------------------------------------------------
// WithSecurityGroup body-ref setter
// --------------------------------------------------------------------------

func TestContainerRegistry_WithSecurityGroup_URIRef(t *testing.T) {
	uri := "/projects/p-1/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-1"
	cr := NewContainerRegistry().WithSecurityGroup(URI(uri))
	if cr.SecurityGroup() != uri {
		t.Errorf("SecurityGroup() = %q", cr.SecurityGroup())
	}
	if cr.Err() != nil {
		t.Errorf("Err() = %v", cr.Err())
	}
}

func TestContainerRegistry_WithSecurityGroup_TypedRef(t *testing.T) {
	ref := URI("/projects/p-1/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-1")
	cr := NewContainerRegistry().WithSecurityGroup(ref)
	if cr.SecurityGroup() != ref.URI() {
		t.Errorf("SecurityGroup() = %q, want %q", cr.SecurityGroup(), ref.URI())
	}
}

func TestContainerRegistry_WithSecurityGroup_EmptyURI(t *testing.T) {
	cr := NewContainerRegistry().WithSecurityGroup(URI(""))
	if cr.Err() == nil {
		t.Error("expected Err() != nil for empty SecurityGroup URI")
	}
	if cr.SecurityGroup() != "" {
		t.Errorf("SecurityGroup() should remain empty, got %q", cr.SecurityGroup())
	}
}

// --------------------------------------------------------------------------
// WithBlockStorage body-ref setter
// --------------------------------------------------------------------------

func TestContainerRegistry_WithBlockStorage_URIRef(t *testing.T) {
	uri := "/projects/p-1/providers/Aruba.Storage/blockStorages/bs-1"
	cr := NewContainerRegistry().WithBlockStorage(URI(uri))
	if cr.BlockStorage() != uri {
		t.Errorf("BlockStorage() = %q", cr.BlockStorage())
	}
	if cr.Err() != nil {
		t.Errorf("Err() = %v", cr.Err())
	}
}

func TestContainerRegistry_WithBlockStorage_TypedRef(t *testing.T) {
	ref := URI("/projects/p-1/providers/Aruba.Storage/blockStorages/bs-1")
	cr := NewContainerRegistry().WithBlockStorage(ref)
	if cr.BlockStorage() != ref.URI() {
		t.Errorf("BlockStorage() = %q, want %q", cr.BlockStorage(), ref.URI())
	}
}

func TestContainerRegistry_WithBlockStorage_EmptyURI(t *testing.T) {
	cr := NewContainerRegistry().WithBlockStorage(URI(""))
	if cr.Err() == nil {
		t.Error("expected Err() != nil for empty BlockStorage URI")
	}
	if cr.BlockStorage() != "" {
		t.Errorf("BlockStorage() should remain empty, got %q", cr.BlockStorage())
	}
}

// --------------------------------------------------------------------------
// Registry scalars
// --------------------------------------------------------------------------

func TestContainerRegistry_WithAdminUsername(t *testing.T) {
	cr := NewContainerRegistry().WithAdminUsername("myuser")
	if cr.AdminUsername() != "myuser" {
		t.Errorf("AdminUsername() = %q", cr.AdminUsername())
	}
	if cr.Err() != nil {
		t.Errorf("Err() = %v", cr.Err())
	}
}

func TestContainerRegistry_OfSize(t *testing.T) {
	cr := NewContainerRegistry().OfSize(ContainerRegistrySizeFlavorSmall)
	if cr.SizeFlavor() != ContainerRegistrySizeFlavorSmall {
		t.Errorf("SizeFlavor() = %q", cr.SizeFlavor())
	}
	req := cr.RawRequest()
	if req.Properties.ConcurrentUsers == nil || *req.Properties.ConcurrentUsers != "Small" {
		t.Errorf("wire ConcurrentUsers = %v", req.Properties.ConcurrentUsers)
	}
}

func TestContainerRegistry_WithBillingPeriod(t *testing.T) {
	cr := NewContainerRegistry().BilledBy(BillingPeriodHour)
	if cr.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", cr.BillingPeriod())
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestContainerRegistry_ToRequest(t *testing.T) {
	vpcURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1"
	subnetURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1"
	sgURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-1"
	eipURI := "/projects/p/providers/Aruba.Network/elasticips/eip-1"
	bsURI := "/projects/p/providers/Aruba.Storage/blockStorages/bs-1"

	cr := NewContainerRegistry().Named(
		"reg-rt").
		Tagged("t1").Tagged("t2").
		InRegion(RegionITBGBergamo).
		WithVPC(URI(vpcURI)).
		WithSubnet(URI(subnetURI)).
		WithSecurityGroup(URI(sgURI)).
		WithElasticIP(URI(eipURI)).
		WithBlockStorage(URI(bsURI)).
		WithAdminUsername("admin").
		OfSize(ContainerRegistrySizeFlavorHighPerf).
		BilledBy(BillingPeriodHour)

	req := cr.RawRequest()

	if req.Metadata.Name != "reg-rt" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if len(req.Metadata.Tags) != 2 {
		t.Errorf("Metadata.Tags = %v", req.Metadata.Tags)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Location.Value = %q", req.Metadata.Location.Value)
	}
	if req.Properties.VPC.URI != vpcURI {
		t.Errorf("Properties.VPC.URI = %q", req.Properties.VPC.URI)
	}
	if req.Properties.Subnet.URI != subnetURI {
		t.Errorf("Properties.Subnet.URI = %q", req.Properties.Subnet.URI)
	}
	if req.Properties.SecurityGroup.URI != sgURI {
		t.Errorf("Properties.SecurityGroup.URI = %q", req.Properties.SecurityGroup.URI)
	}
	if req.Properties.PublicIp.URI != eipURI {
		t.Errorf("Properties.PublicIp.URI = %q", req.Properties.PublicIp.URI)
	}
	if req.Properties.BlockStorage.URI != bsURI {
		t.Errorf("Properties.BlockStorage.URI = %q", req.Properties.BlockStorage.URI)
	}
	if req.Properties.AdminUser == nil || req.Properties.AdminUser.Username != "admin" {
		t.Errorf("Properties.AdminUser = %v", req.Properties.AdminUser)
	}
	if req.Properties.ConcurrentUsers == nil || *req.Properties.ConcurrentUsers != "HighPerf" {
		t.Errorf("Properties.ConcurrentUsers = %v", req.Properties.ConcurrentUsers)
	}
	if req.Properties.BillingPlanCommon == nil || req.Properties.BillingPlanCommon.BillingPeriod == nil || *req.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodHour {
		t.Errorf("Properties.BillingPlanCommon.BillingPeriod = %v", req.Properties.BillingPlanCommon)
	}
}

func TestContainerRegistry_ToRequest_Empty(t *testing.T) {
	cr := NewContainerRegistry()
	req := cr.RawRequest() // must not panic

	// All optional pointer fields should be nil when not set.
	if req.Properties.AdminUser != nil {
		t.Errorf("AdminUser should be nil, got %v", req.Properties.AdminUser)
	}
	if req.Properties.ConcurrentUsers != nil {
		t.Errorf("ConcurrentUsers should be nil, got %v", req.Properties.ConcurrentUsers)
	}
	if req.Properties.BillingPlanCommon == nil || req.Properties.BillingPlanCommon.BillingPeriod == nil || *req.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodHour {
		t.Errorf("BillingPlanCommon.BillingPeriod should default to Hour, got %v", req.Properties.BillingPlanCommon)
	}
	// Body-ref ReferenceResourceCommon fields should have empty URIs.
	if req.Properties.VPC.URI != "" {
		t.Errorf("VPC.URI should be empty, got %q", req.Properties.VPC.URI)
	}
}

func TestContainerRegistry_ToRequest_OmitsPassword(t *testing.T) {
	cr := NewContainerRegistry().WithAdminUsername("admin")
	req := cr.RawRequest()
	if req.Properties.AdminUser == nil {
		t.Fatal("AdminUser should not be nil")
	}
	if req.Properties.AdminUser.Username != "admin" {
		t.Errorf("Username = %q", req.Properties.AdminUser.Username)
	}
	b, _ := json.Marshal(req.Properties.AdminUser)
	if strings.Contains(string(b), `"password"`) {
		t.Errorf("marshalled AdminUser must not contain password field: %s", b)
	}
}

func TestContainerRegistry_ToRequest_OmitsAdminUserWhenNeither(t *testing.T) {
	cr := NewContainerRegistry()
	req := cr.RawRequest()
	if req.Properties.AdminUser != nil {
		t.Errorf("AdminUser should be nil when neither username nor password is set, got %+v", req.Properties.AdminUser)
	}
	b, _ := json.Marshal(req.Properties)
	if strings.Contains(string(b), `"adminUser"`) {
		t.Errorf("marshalled properties should not contain adminUser key: %s", b)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration helpers
// --------------------------------------------------------------------------

func containerRegistryTestResponse(name string) *types.ContainerRegistryResponse {
	id := "cr-1"
	uri := "/projects/p/providers/Aruba.Container/registries/cr-1"
	state := types.State("Active")
	size := "Small"
	vpcURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1"
	subnetURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1"
	sgURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-1"
	eipURI := "/projects/p/providers/Aruba.Network/elasticips/eip-1"
	bsURI := "/projects/p/providers/Aruba.Storage/blockStorages/bs-1"
	return &types.ContainerRegistryResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			URI:              &uri,
			Name:             func() *string { s := name; return &s }(),
			Tags:             []string{"tag1"},
			LocationResponse: &types.LocationResponse{Value: RegionITBGBergamo},
			ProjectMetadataResponse: &types.ProjectMetadataResponse{
				ID: "p",
			},
		},
		Properties: types.ContainerRegistryPropertiesResponse{
			VPC:             types.ReferenceResourceCommon{URI: vpcURI},
			Subnet:          types.ReferenceResourceCommon{URI: subnetURI},
			SecurityGroup:   types.ReferenceResourceCommon{URI: sgURI},
			PublicIp:        types.ReferenceResourceCommon{URI: eipURI},
			BlockStorage:    types.ReferenceResourceCommon{URI: bsURI},
			AdminUser:       &types.UserCredentialCommon{Username: "admin"},
			ConcurrentUsers: &size,
			BillingPlanCommon: func() *types.BillingPlanCommon {
				v := BillingPeriodHour
				return &types.BillingPlanCommon{BillingPeriod: &v}
			}(),
		},
		Status: types.ResourceStatusResponse{
			State: &state,
		},
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration tests
// --------------------------------------------------------------------------

func TestContainerRegistry_FromResponseHydration(t *testing.T) {
	cr := &ContainerRegistry{}
	resp := containerRegistryTestResponse("my-registry")
	cr.fromResponse(resp)

	if cr.ID() != "cr-1" {
		t.Errorf("ID() = %q", cr.ID())
	}
	if cr.ContainerRegistryID() != "cr-1" {
		t.Errorf("ContainerRegistryID() = %q", cr.ContainerRegistryID())
	}
	if cr.URI() != "/projects/p/providers/Aruba.Container/registries/cr-1" {
		t.Errorf("URI() = %q", cr.URI())
	}
	if cr.Name() != "my-registry" {
		t.Errorf("Name() = %q", cr.Name())
	}
	if tags := cr.Tags(); len(tags) != 1 || tags[0] != "tag1" {
		t.Errorf("Tags() = %v", tags)
	}
	if cr.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", cr.Region())
	}
	if cr.State() != "Active" {
		t.Errorf("State() = %q", cr.State())
	}
	if cr.VPC() != "/projects/p/providers/Aruba.Network/vpcs/vpc-1" {
		t.Errorf("VPC() = %q", cr.VPC())
	}
	if cr.Subnet() != "/projects/p/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1" {
		t.Errorf("Subnet() = %q", cr.Subnet())
	}
	if cr.SecurityGroup() != "/projects/p/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-1" {
		t.Errorf("SecurityGroup() = %q", cr.SecurityGroup())
	}
	if cr.ElasticIP() != "/projects/p/providers/Aruba.Network/elasticips/eip-1" {
		t.Errorf("PublicIP() = %q", cr.ElasticIP())
	}
	if cr.BlockStorage() != "/projects/p/providers/Aruba.Storage/blockStorages/bs-1" {
		t.Errorf("BlockStorage() = %q", cr.BlockStorage())
	}
	if cr.AdminUsername() != "admin" {
		t.Errorf("AdminUsername() = %q", cr.AdminUsername())
	}
	if cr.SizeFlavor() != ContainerRegistrySizeFlavorSmall {
		t.Errorf("SizeFlavor() = %q", cr.SizeFlavor())
	}
	if cr.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", cr.BillingPeriod())
	}
	if cr.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", cr.ProjectID())
	}
	if cr.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestContainerRegistry_FromResponse_NilSafe(t *testing.T) {
	cr := &ContainerRegistry{}
	cr.fromResponse(nil) // must not panic
	if cr.ID() != "" || cr.URI() != "" || cr.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}
}

func TestContainerRegistry_FromResponse_BackfillsProjectID_FromMetadata(t *testing.T) {
	resp := containerRegistryTestResponse("n")
	cr := &ContainerRegistry{}
	cr.fromResponse(resp)
	if cr.ProjectID() != "p" {
		t.Errorf("ProjectID() from metadata = %q", cr.ProjectID())
	}
}

func TestContainerRegistry_FromResponse_BackfillsProjectID_FromURI(t *testing.T) {
	id := "cr-99"
	uri := "/projects/p-uri/providers/Aruba.Container/registries/cr-99"
	resp := &types.ContainerRegistryResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  &id,
			URI: &uri,
			// No ProjectMetadataResponse — should backfill from URI.
		},
	}
	cr := &ContainerRegistry{}
	cr.fromResponse(resp)
	if cr.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() via URI backfill = %q", cr.ProjectID())
	}
}

// --------------------------------------------------------------------------
// containerRegistryIDsFromRef helper
// --------------------------------------------------------------------------

func TestContainerRegistryIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Container/registries/cr-1")
	pid, rid, err := containerRegistryIDsFromRef(ref)
	if err != nil || pid != "p" || rid != "cr-1" {
		t.Errorf("containerRegistryIDsFromRef = (%q, %q, %v)", pid, rid, err)
	}
}

func TestContainerRegistryIDsFromRef_BadURI_NoRegistries(t *testing.T) {
	_, _, err := containerRegistryIDsFromRef(URI("/projects/p/providers/Aruba.Container/something/else"))
	if err == nil {
		t.Error("expected error for URI without /registries/<id>")
	}
}

func TestContainerRegistryIDsFromRef_BadURI_NoProject(t *testing.T) {
	_, _, err := containerRegistryIDsFromRef(URI("/providers/Aruba.Container/registries/cr-1"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

// --------------------------------------------------------------------------
// fakeContainerRegistryLowLevel — body-capture tests
// --------------------------------------------------------------------------

type fakeContainerRegistryLowLevel struct {
	createFunc func(ctx context.Context, projectID string, body types.ContainerRegistryRequest, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error)
	updateFunc func(ctx context.Context, projectID, registryID string, body types.ContainerRegistryRequest, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error)
	getFunc    func(ctx context.Context, projectID, registryID string, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error)
	deleteFunc func(ctx context.Context, projectID, registryID string, params *types.RequestParameters) (*types.Response[any], error)
	listFunc   func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.ContainerRegistryListResponse], error)
}

func (f *fakeContainerRegistryLowLevel) Create(ctx context.Context, projectID string, body types.ContainerRegistryRequest, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
	return f.createFunc(ctx, projectID, body, params)
}
func (f *fakeContainerRegistryLowLevel) Update(ctx context.Context, projectID, registryID string, body types.ContainerRegistryRequest, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
	return f.updateFunc(ctx, projectID, registryID, body, params)
}
func (f *fakeContainerRegistryLowLevel) Get(ctx context.Context, projectID, registryID string, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
	return f.getFunc(ctx, projectID, registryID, params)
}
func (f *fakeContainerRegistryLowLevel) Delete(ctx context.Context, projectID, registryID string, params *types.RequestParameters) (*types.Response[any], error) {
	return f.deleteFunc(ctx, projectID, registryID, params)
}
func (f *fakeContainerRegistryLowLevel) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.ContainerRegistryListResponse], error) {
	return f.listFunc(ctx, projectID, params)
}

// --------------------------------------------------------------------------
// HTTP-mock adapter helper
// --------------------------------------------------------------------------

func buildContainerRegistryTestAdapter(t *testing.T, handler http.HandlerFunc) *containerRegistriesClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newContainerRegistriesClientAdapter(testutil.NewClient(t, server.URL))
}

const containerRegistrySuccessBody = `{` +
	`"metadata":{"id":"cr-1","name":"my-registry","uri":"/projects/p/providers/Aruba.Container/registries/cr-1","project":{"id":"p"}},` +
	`"properties":{` +
	`"vpc":{"uri":"/projects/p/providers/Aruba.Network/vpcs/vpc-1"},` +
	`"subnet":{"uri":"/projects/p/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1"},` +
	`"securityGroup":{"uri":"/projects/p/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-1"},` +
	`"publicIp":{"uri":"/projects/p/providers/Aruba.Network/elasticips/eip-1"},` +
	`"blockStorage":{"uri":"/projects/p/providers/Aruba.Storage/blockStorages/bs-1"},` +
	`"adminUser":{"username":"admin"},"size":"Small","billingPlan":{"billingPeriod":"Hour"}` +
	`},` +
	`"status":{"state":"Active"}}`

// --------------------------------------------------------------------------
// Create adapter tests
// --------------------------------------------------------------------------

func TestContainerRegistriesClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.ContainerRegistryRequest
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "registries") {
			t.Errorf("path %q should contain 'registries'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, containerRegistrySuccessBody)
	})

	cr := NewContainerRegistry().
		InProject(URI("/projects/p")).
		Named("my-registry").
		InRegion(RegionITBGBergamo).
		WithVPC(URI("/projects/p/providers/Aruba.Network/vpcs/vpc-1")).
		WithSubnet(URI("/projects/p/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1")).
		WithSecurityGroup(URI("/projects/p/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-1")).
		WithElasticIP(URI("/projects/p/providers/Aruba.Network/elasticips/eip-1")).
		WithBlockStorage(URI("/projects/p/providers/Aruba.Storage/blockStorages/bs-1")).
		WithAdminUsername("admin").
		OfSize(ContainerRegistrySizeFlavorSmall).
		BilledBy(BillingPeriodHour)

	result, err := adapter.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "cr-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-registry" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	// Wire body assertions
	if gotBody.Metadata.Name != "my-registry" {
		t.Errorf("request Metadata.Name = %q", gotBody.Metadata.Name)
	}
	if gotBody.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("request Metadata.Location.Value = %q", gotBody.Metadata.Location.Value)
	}
	if gotBody.Properties.VPC.URI != "/projects/p/providers/Aruba.Network/vpcs/vpc-1" {
		t.Errorf("request Properties.VPC.URI = %q", gotBody.Properties.VPC.URI)
	}
}

func TestContainerRegistriesClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	_, err := adapter.Create(context.Background(), NewContainerRegistry().
		Named("x"))
	if err == nil {
		t.Fatal("expected error when ContainerRegistry has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestContainerRegistriesClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" field — triggers MetadataValidationError from low-level Validate()
		fmt.Fprint(w, `{"metadata":{"name":"reg","uri":"/projects/p/providers/Aruba.Container/registries/x"},"properties":{},"status":{}}`)
	})

	cr := NewContainerRegistry().InProject(URI("/projects/p")).
		Named("reg")
	result, err := adapter.Create(context.Background(), cr)
	if err == nil {
		t.Fatal("expected MetadataValidationError, got nil")
	}
	var mvErr *types.MetadataValidationError
	if !errors.As(err, &mvErr) {
		t.Fatalf("expected *types.MetadataValidationError, got %T: %v", err, err)
	}
	if result == nil {
		t.Fatal("result must be non-nil alongside MetadataValidationError")
	}
}

func TestContainerRegistriesClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "name is required", 422))
	})

	cr := NewContainerRegistry().InProject(URI("/projects/p"))
	result, err := adapter.Create(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error on 422")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("HTTPError.StatusCode = %d", httpErr.StatusCode)
	}
	if result == nil {
		t.Fatal("result must be non-nil on non-2xx")
	}
}

func TestContainerRegistriesClientAdapter_Create_WithBodyRefs_ViaFake(t *testing.T) {
	vpcURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1"
	subnetURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1"
	sgURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-1"
	eipURI := "/projects/p/providers/Aruba.Network/elasticips/eip-1"
	bsURI := "/projects/p/providers/Aruba.Storage/blockStorages/bs-1"

	var captured types.ContainerRegistryRequest
	resp := &types.Response[types.ContainerRegistryResponse]{
		StatusCode: http.StatusCreated,
		Data:       containerRegistryTestResponse("reg"),
	}
	fake := &fakeContainerRegistryLowLevel{
		createFunc: func(_ context.Context, _ string, body types.ContainerRegistryRequest, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
			captured = body
			return resp, nil
		},
	}
	adapter := &containerRegistriesClientAdapter{low: fake}

	cr := NewContainerRegistry().
		InProject(URI("/projects/p")).
		InRegion(RegionITBGBergamo).
		WithVPC(URI(vpcURI)).
		WithSubnet(URI(subnetURI)).
		WithSecurityGroup(URI(sgURI)).
		WithElasticIP(URI(eipURI)).
		WithBlockStorage(URI(bsURI)).
		WithAdminUsername("admin").
		OfSize(ContainerRegistrySizeFlavorHighPerf)

	_, err := adapter.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if captured.Properties.VPC.URI != vpcURI {
		t.Errorf("captured VPC.URI = %q", captured.Properties.VPC.URI)
	}
	if captured.Properties.Subnet.URI != subnetURI {
		t.Errorf("captured Subnet.URI = %q", captured.Properties.Subnet.URI)
	}
	if captured.Properties.SecurityGroup.URI != sgURI {
		t.Errorf("captured SecurityGroup.URI = %q", captured.Properties.SecurityGroup.URI)
	}
	if captured.Properties.PublicIp.URI != eipURI {
		t.Errorf("captured PublicIp.URI = %q", captured.Properties.PublicIp.URI)
	}
	if captured.Properties.BlockStorage.URI != bsURI {
		t.Errorf("captured BlockStorage.URI = %q", captured.Properties.BlockStorage.URI)
	}
	if captured.Properties.AdminUser == nil || captured.Properties.AdminUser.Username != "admin" {
		t.Errorf("captured AdminUser = %v", captured.Properties.AdminUser)
	}
	if captured.Properties.ConcurrentUsers == nil || *captured.Properties.ConcurrentUsers != "HighPerf" {
		t.Errorf("captured ConcurrentUsers = %v", captured.Properties.ConcurrentUsers)
	}
}

// --------------------------------------------------------------------------
// Update adapter tests
// --------------------------------------------------------------------------

func TestContainerRegistriesClientAdapter_Update_Success(t *testing.T) {
	var capturedPath string
	var gotBody types.ContainerRegistryRequest
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, containerRegistrySuccessBody)
	})

	cr := NewContainerRegistry().
		InProject(URI("/projects/p")).
		Named("my-registry").
		InRegion(RegionITBGBergamo).
		WithVPC(URI("/projects/p/providers/Aruba.Network/vpcs/vpc-1"))

	// Hydrate to get an ID before Update.
	cr.fromResponse(containerRegistryTestResponse("my-registry"))

	result, err := adapter.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if result.ID() != "cr-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if !containsSubstring(capturedPath, "cr-1") {
		t.Errorf("path %q should contain registry ID 'cr-1'", capturedPath)
	}
}

func TestContainerRegistriesClientAdapter_Update_NoID(t *testing.T) {
	callCount := 0
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	cr := NewContainerRegistry().InProject(URI("/projects/p")).
		Named("x")
	_, err := adapter.Update(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error when ContainerRegistry has no ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without ID")
	}
}

func TestContainerRegistriesClientAdapter_Update_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	// Hydrate to get an ID but without a project.
	cr := &ContainerRegistry{}
	cr.fromResponse(containerRegistryTestResponse("n"))
	cr.projectID = "" // strip project

	_, err := adapter.Update(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error when ContainerRegistry has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestContainerRegistriesClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Conflict", "registry already exists", 409))
	})

	cr := &ContainerRegistry{}
	cr.fromResponse(containerRegistryTestResponse("n"))

	result, err := adapter.Update(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error on 409")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusConflict {
		t.Errorf("HTTPError.StatusCode = %d", httpErr.StatusCode)
	}
	if result == nil {
		t.Fatal("result must be non-nil on non-2xx")
	}
}

// --------------------------------------------------------------------------
// Get adapter tests
// --------------------------------------------------------------------------

func TestContainerRegistriesClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, containerRegistrySuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Container/registries/cr-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "cr-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	if !containsSubstring(capturedPath, "registries") {
		t.Errorf("path %q should contain 'registries'", capturedPath)
	}
}

func TestContainerRegistriesClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, containerRegistrySuccessBody)
	})

	existing := &ContainerRegistry{}
	existing.fromResponse(containerRegistryTestResponse("my-registry"))

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "cr-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

// --------------------------------------------------------------------------
// Delete adapter tests
// --------------------------------------------------------------------------

func TestContainerRegistriesClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Container/registries/cr-1"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestContainerRegistriesClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "registry not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Container/registries/missing"))
	if err == nil {
		t.Fatal("expected error on 404")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// List adapter tests
// --------------------------------------------------------------------------

func TestContainerRegistriesClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"cr-1","name":"n1","uri":"/projects/p/providers/Aruba.Container/registries/cr-1","project":{"id":"p"}},"properties":{"vpc":{"uri":"/projects/p/providers/Aruba.Network/vpcs/vpc-1"},"size":"Small","billingPlan":{"billingPeriod":"Hour"}},"status":{}},`+
			`{"metadata":{"id":"cr-2","name":"n2","uri":"/projects/p/providers/Aruba.Container/registries/cr-2","project":{"id":"p"}},"properties":{"vpc":{"uri":"/projects/p/providers/Aruba.Network/vpcs/vpc-1"},"size":"Medium","billingPlan":{"billingPeriod":"Hour"}},"status":{}}`+
			`]}`)
	})

	list, err := adapter.List(context.Background(), URI("/projects/p"))
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if list.Total() != 2 {
		t.Errorf("Total() = %d", list.Total())
	}
	items := list.Items()
	if len(items) != 2 {
		t.Fatalf("Items() len = %d", len(items))
	}
	if items[0].ID() != "cr-1" || items[0].Name() != "n1" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[0].SizeFlavor() != ContainerRegistrySizeFlavorSmall {
		t.Errorf("items[0].SizeFlavor() = %q", items[0].SizeFlavor())
	}
	if items[1].ID() != "cr-2" || items[1].BillingPeriod() != BillingPeriodHour {
		t.Errorf("items[1] ID=%q BillingPeriod=%q", items[1].ID(), items[1].BillingPeriod())
	}
	if items[0].ProjectID() != "p" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
}

// --------------------------------------------------------------------------
// Reflective check: ContainerRegistryClient has Update method
// --------------------------------------------------------------------------

func TestContainerRegistryClient_HasUpdateMethod(t *testing.T) {
	iface := reflect.TypeOf((*ContainerRegistryClient)(nil)).Elem()
	for i := range iface.NumMethod() {
		if iface.Method(i).Name == "Update" {
			return // found — test passes
		}
	}
	t.Fatal("ContainerRegistryClient must have an Update method")
}

// --------------------------------------------------------------------------
// Shape A — InRegion
// --------------------------------------------------------------------------

func TestContainerRegistry_InRegion(t *testing.T) {
	cr := NewContainerRegistry().InRegion("ITMI-Milano-1")
	if cr.Region() != "ITMI-Milano-1" {
		t.Errorf("Region() = %q, want ITMI-Milano-1", cr.Region())
	}
}

// --------------------------------------------------------------------------
// Shape F — SizeFlavor() unknown-string branches
// --------------------------------------------------------------------------

func TestContainerRegistry_SizeFlavor_UnknownResponseString(t *testing.T) {
	cr := &ContainerRegistry{}
	unknown := "not-a-flavor"
	cr.response = &types.ContainerRegistryResponse{
		Properties: types.ContainerRegistryPropertiesResponse{
			ConcurrentUsers: &unknown,
		},
	}
	if cr.SizeFlavor() != ContainerRegistrySizeFlavor("not-a-flavor") {
		t.Errorf("SizeFlavor() = %q for unknown response string, want %q", cr.SizeFlavor(), "not-a-flavor")
	}
}

func TestContainerRegistry_SizeFlavor_UnknownLocalString(t *testing.T) {
	cr := &ContainerRegistry{}
	unknown := "not-a-flavor"
	cr.concurrentUsers = &unknown
	if cr.SizeFlavor() != ContainerRegistrySizeFlavor("not-a-flavor") {
		t.Errorf("SizeFlavor() = %q for unknown local string, want %q", cr.SizeFlavor(), "not-a-flavor")
	}
}

// --------------------------------------------------------------------------
// Shape E — ContainerRegistry adapter additional error paths
// --------------------------------------------------------------------------

// Get_BadRef: URI lacking /registries/<id>.
func TestContainerRegistriesClientAdapter_Get_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.Get(context.Background(), URI("/projects/p/no-registries"))
	if err == nil {
		t.Fatal("expected error for bad ref")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made for bad ref")
	}
}

// Get_NonTwoXX: stub returns 404 → HTTPError.
func TestContainerRegistriesClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "not found", 404))
	})
	_, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Container/registries/cr-1"))
	if err == nil {
		t.Fatal("expected error on 404")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// Delete_BadRef: URI without registries segment → error before HTTP.
func TestContainerRegistriesClientAdapter_Delete_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusNoContent)
	})
	err := adapter.Delete(context.Background(), URI("/projects/p/no-registries"))
	if err == nil {
		t.Fatal("expected error for bad ref")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made for bad ref")
	}
}

// List_BadRef: parent ref without project ID → error before HTTP.
func TestContainerRegistriesClientAdapter_List_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/no-project"))
	if err == nil {
		t.Fatal("expected error for ref without project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made for bad ref")
	}
}

// List_NonTwoXX: stub returns 500 → HTTPError.
func TestContainerRegistriesClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildContainerRegistryTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Server Error", "internal", 500))
	})
	_, err := adapter.List(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected error on 500")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T", err)
	}
}

// --------------------------------------------------------------------------
// Shape F — Accessors on zero-value ContainerRegistry
// --------------------------------------------------------------------------

func TestContainerRegistry_Accessors_ZeroValue(t *testing.T) {
	cr := NewContainerRegistry()

	if cr.ElasticIP() != "" {
		t.Errorf("PublicIP() = %q, want empty", cr.ElasticIP())
	}
	if cr.VPC() != "" {
		t.Errorf("VPC() = %q, want empty", cr.VPC())
	}
	if cr.Subnet() != "" {
		t.Errorf("Subnet() = %q, want empty", cr.Subnet())
	}
	if cr.SecurityGroup() != "" {
		t.Errorf("SecurityGroup() = %q, want empty", cr.SecurityGroup())
	}
	if cr.BlockStorage() != "" {
		t.Errorf("BlockStorage() = %q, want empty", cr.BlockStorage())
	}
	if cr.AdminUsername() != "" {
		t.Errorf("AdminUsername() = %q, want empty", cr.AdminUsername())
	}
	if cr.SizeFlavor() != "" {
		t.Errorf("SizeFlavor() = %q, want empty", cr.SizeFlavor())
	}
	if cr.BillingPeriod() != "" {
		t.Errorf("BillingPeriod() = %q, want empty", cr.BillingPeriod())
	}
	if cr.Raw() != nil {
		t.Errorf("Raw() = %v, want nil", cr.Raw())
	}
	if cr.URI() != "" {
		t.Errorf("URI() = %q, want empty", cr.URI())
	}
	if cr.ContainerRegistryID() != "" {
		t.Errorf("ContainerRegistryID() = %q, want empty", cr.ContainerRegistryID())
	}
	_ = cr.RawRequest() // must not panic
}

// --------------------------------------------------------------------------
// Additional coverage: ContainerRegistry adapter error paths via fake
// --------------------------------------------------------------------------

// Update_WithErr: r.Err() is non-nil → return early.
func TestContainerRegistriesClientAdapter_Update_WithErr(t *testing.T) {
	callCount := 0
	fake := &fakeContainerRegistryLowLevel{
		updateFunc: func(_ context.Context, _, _ string, _ types.ContainerRegistryRequest, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
			callCount++
			return nil, nil
		},
	}
	adapter := &containerRegistriesClientAdapter{low: fake}

	cr := &ContainerRegistry{}
	cr.fromResponse(containerRegistryTestResponse("n"))
	cr.addErr(fmt.Errorf("pre-existing validation error"))

	_, err := adapter.Update(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error when r.Err() is set")
	}
	if callCount != 0 {
		t.Error("no low-level call should be made when r.Err() is set")
	}
}

// Update_LowLevelError: low-level Update returns network error → propagate.
func TestContainerRegistriesClientAdapter_Update_LowLevelError(t *testing.T) {
	fake := &fakeContainerRegistryLowLevel{
		updateFunc: func(_ context.Context, _, _ string, _ types.ContainerRegistryRequest, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
			return nil, fmt.Errorf("connection reset")
		},
	}
	adapter := &containerRegistriesClientAdapter{low: fake}

	cr := &ContainerRegistry{}
	cr.fromResponse(containerRegistryTestResponse("n"))

	_, err := adapter.Update(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from low-level Update")
	}
	if !containsSubstring(err.Error(), "connection reset") {
		t.Errorf("error should mention 'connection reset', got %q", err.Error())
	}
}

// Get_LowLevelError: low-level Get returns error → propagate.
func TestContainerRegistriesClientAdapter_Get_LowLevelError(t *testing.T) {
	fake := &fakeContainerRegistryLowLevel{
		getFunc: func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
			return nil, fmt.Errorf("dns lookup failed")
		},
	}
	adapter := &containerRegistriesClientAdapter{low: fake}

	_, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Container/registries/cr-1"))
	if err == nil {
		t.Fatal("expected error from low-level Get")
	}
	if !containsSubstring(err.Error(), "dns lookup failed") {
		t.Errorf("error should mention 'dns lookup failed', got %q", err.Error())
	}
}

// Delete_LowLevelError: low-level Delete returns error → propagate.
func TestContainerRegistriesClientAdapter_Delete_LowLevelError(t *testing.T) {
	fake := &fakeContainerRegistryLowLevel{
		deleteFunc: func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
			return nil, fmt.Errorf("timeout")
		},
	}
	adapter := &containerRegistriesClientAdapter{low: fake}

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Container/registries/cr-1"))
	if err == nil {
		t.Fatal("expected error from low-level Delete")
	}
	if !containsSubstring(err.Error(), "timeout") {
		t.Errorf("error should mention 'timeout', got %q", err.Error())
	}
}

// List_LowLevelError: low-level List returns error → propagate.
func TestContainerRegistriesClientAdapter_List_LowLevelError(t *testing.T) {
	fake := &fakeContainerRegistryLowLevel{
		listFunc: func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryListResponse], error) {
			return nil, fmt.Errorf("upstream unavailable")
		},
	}
	adapter := &containerRegistriesClientAdapter{low: fake}

	_, err := adapter.List(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected error from low-level List")
	}
	if !containsSubstring(err.Error(), "upstream unavailable") {
		t.Errorf("error should mention 'upstream unavailable', got %q", err.Error())
	}
}

func TestContainerRegistry_FromResponse_SetsStatus(t *testing.T) {
	r := &ContainerRegistry{}
	state := types.State("Active")
	r.fromResponse(&types.ContainerRegistryResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if r.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", r.State())
	}
}

func TestContainerRegistriesClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, containerRegistrySuccessBody)
	})
	adapter := newContainerRegistriesClientAdapter(testutil.NewClient(t, server.URL))
	cr, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Container/registries/cr-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&cr.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned ContainerRegistry")
	}
}

// --------------------------------------------------------------------------
// Data block — getters
// --------------------------------------------------------------------------

func TestContainerRegistry_FromResponse_HydratesData(t *testing.T) {
	passwordSet := true
	lastSet := time.Date(2026, 5, 29, 10, 0, 0, 0, time.UTC)
	fqdn := "registry.example.com"
	pub := "https://registry.example.com"
	priv := "https://private.registry.example.com"
	ver := "2.8"

	cr := &ContainerRegistry{}
	cr.fromResponse(&types.ContainerRegistryResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  func() *string { s := "cr-1"; return &s }(),
			URI: func() *string { s := "/projects/p/providers/Aruba.Container/registries/cr-1"; return &s }(),
		},
		Data: &types.ContainerRegistryDataResponse{
			Private: &types.ContainerRegistryDataPrivateResponse{
				PasswordSet:       &passwordSet,
				PasswordLastSetAt: &lastSet,
			},
			Info: &types.ContainerRegistryDataInfoResponse{
				FQDN:           &fqdn,
				PublicBaseURL:  &pub,
				PrivateBaseURL: &priv,
				Version:        &ver,
			},
		},
	})

	if !cr.IsPasswordSet() {
		t.Error("IsPasswordSet() = false, want true")
	}
	if cr.AdminPasswordLastSetAt() != lastSet {
		t.Errorf("AdminPasswordLastSetAt() = %v, want %v", cr.AdminPasswordLastSetAt(), lastSet)
	}
	if cr.FQDN() != fqdn {
		t.Errorf("FQDN() = %q, want %q", cr.FQDN(), fqdn)
	}
	if cr.PublicBaseURL() != pub {
		t.Errorf("PublicBaseURL() = %q, want %q", cr.PublicBaseURL(), pub)
	}
	if cr.PrivateBaseURL() != priv {
		t.Errorf("PrivateBaseURL() = %q, want %q", cr.PrivateBaseURL(), priv)
	}
	if cr.Version() != ver {
		t.Errorf("Version() = %q, want %q", cr.Version(), ver)
	}
}

func TestContainerRegistry_DataGetters_NilSafe(t *testing.T) {
	cr := &ContainerRegistry{}
	cr.fromResponse(&types.ContainerRegistryResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  func() *string { s := "cr-1"; return &s }(),
			URI: func() *string { s := "/projects/p/providers/Aruba.Container/registries/cr-1"; return &s }(),
		},
		// Data intentionally nil
	})

	if cr.IsPasswordSet() {
		t.Error("IsPasswordSet() = true, want false for nil Data")
	}
	if !cr.AdminPasswordLastSetAt().IsZero() {
		t.Errorf("AdminPasswordLastSetAt() = %v, want zero", cr.AdminPasswordLastSetAt())
	}
	if cr.FQDN() != "" {
		t.Errorf("FQDN() = %q, want empty", cr.FQDN())
	}
	if cr.PublicBaseURL() != "" {
		t.Errorf("PublicBaseURL() = %q, want empty", cr.PublicBaseURL())
	}
	if cr.PrivateBaseURL() != "" {
		t.Errorf("PrivateBaseURL() = %q, want empty", cr.PrivateBaseURL())
	}
	if cr.Version() != "" {
		t.Errorf("Version() = %q, want empty", cr.Version())
	}
}
