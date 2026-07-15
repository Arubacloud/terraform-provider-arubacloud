package aruba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Compile-time interface satisfaction
// --------------------------------------------------------------------------

var (
	_ Ref     = (*KaaS)(nil)
	_ Wrapper = (*KaaS)(nil)
)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestKaaS_FluentSetters(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-1", "my-proj", "/projects/p-1"))

	vpcURI := URI("/projects/p-1/providers/Aruba.Network/vpcs/vpc-1")
	subnetURI := URI("/projects/p-1/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1")
	sgFixture := NewSecurityGroup().
		Named("sg-name")

	k := NewKaaS().
		InProject(proj).
		Named("my-cluster").
		Tagged("env:prod").
		Tagged("k8s").
		Tagged("env:prod"). // dedupe
		InRegion(RegionITBGBergamo).
		WithVPC(vpcURI).
		WithSubnet(subnetURI).
		WithSecurityGroup(sgFixture).
		WithNodeCIDR("10.100.0.0/16", "node-cidr").
		WithPodCIDR("10.200.0.0/16").
		WithKubernetesVersion("1.32.3").
		HighlyAvailable().
		WithMaxStorageQuotaGB(100).
		BilledBy(BillingPeriodHour).
		WithIdentity("cid", "csecret")

	if k.Name() != "my-cluster" {
		t.Errorf("Name() = %q", k.Name())
	}
	if tags := k.Tags(); len(tags) != 2 || tags[0] != "env:prod" || tags[1] != "k8s" {
		t.Errorf("Tags() = %v", tags)
	}
	if k.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", k.Region())
	}
	if k.VPC() != vpcURI.URI() {
		t.Errorf("VPC() = %q", k.VPC())
	}
	if k.Subnet() != subnetURI.URI() {
		t.Errorf("Subnet() = %q", k.Subnet())
	}
	if k.SecurityGroupName() != "sg-name" {
		t.Errorf("SecurityGroupName() = %q", k.SecurityGroupName())
	}
	if k.KubernetesVersion() != "1.32.3" {
		t.Errorf("KubernetesVersion() = %q", k.KubernetesVersion())
	}
	if k.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", k.BillingPeriod())
	}
	if k.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", k.ProjectID())
	}
	if k.Err() != nil {
		t.Errorf("Err() = %v", k.Err())
	}
}

// --------------------------------------------------------------------------
// IntoProject
// --------------------------------------------------------------------------

func TestKaaS_IntoProject_TypedRef(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-42", "proj", "/projects/p-42"))
	k := NewKaaS().InProject(proj)
	if k.ProjectID() != "p-42" {
		t.Errorf("ProjectID() = %q", k.ProjectID())
	}
	if k.Err() != nil {
		t.Errorf("Err() = %v", k.Err())
	}
}

func TestKaaS_IntoProject_URIRef(t *testing.T) {
	k := NewKaaS().InProject(URI("/projects/p-uri"))
	if k.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", k.ProjectID())
	}
}

func TestKaaS_IntoProject_BadRef(t *testing.T) {
	k := NewKaaS().InProject(URI("not-a-project-uri"))
	if k.Err() == nil {
		t.Error("expected Err() != nil for non-project URI")
	}
}

// --------------------------------------------------------------------------
// Body-ref setters — WithVPC, WithSubnet
// --------------------------------------------------------------------------

func TestKaaS_WithVPC_URIRef(t *testing.T) {
	uri := "/projects/p/providers/Aruba.Network/vpcs/vpc-1"
	k := NewKaaS().WithVPC(URI(uri))
	if k.VPC() != uri {
		t.Errorf("VPC() = %q", k.VPC())
	}
	if k.Err() != nil {
		t.Errorf("Err() = %v", k.Err())
	}
}

func TestKaaS_WithVPC_TypedRef(t *testing.T) {
	vpc := &VPC{}
	vpc.fromResponse(vpcTestResponse("vpc-1", "v", "/projects/p/network/vpcs/vpc-1", "p"))
	k := NewKaaS().WithVPC(vpc)
	if k.VPC() != vpc.URI() {
		t.Errorf("VPC() = %q", k.VPC())
	}
}

func TestKaaS_WithVPC_EmptyURI(t *testing.T) {
	k := NewKaaS().WithVPC(URI(""))
	if k.Err() == nil {
		t.Error("expected Err() != nil for empty VPC URI")
	}
	if k.VPC() != "" {
		t.Errorf("VPC() should remain empty, got %q", k.VPC())
	}
}

func TestKaaS_WithSubnet_URIRef(t *testing.T) {
	uri := "/projects/p/providers/Aruba.Network/vpcs/v/subnets/sn-1"
	k := NewKaaS().WithSubnet(URI(uri))
	if k.Subnet() != uri {
		t.Errorf("Subnet() = %q", k.Subnet())
	}
}

func TestKaaS_WithSubnet_EmptyURI(t *testing.T) {
	k := NewKaaS().WithSubnet(URI(""))
	if k.Err() == nil {
		t.Error("expected Err() != nil for empty Subnet URI")
	}
}

// --------------------------------------------------------------------------
// Scalar setters
// --------------------------------------------------------------------------

func TestKaaS_WithSecurityGroup(t *testing.T) {
	sg := NewSecurityGroup().
		Named("my-sg")
	k := NewKaaS().WithSecurityGroup(sg)
	if k.SecurityGroupName() != "my-sg" {
		t.Errorf("SecurityGroupName() = %q", k.SecurityGroupName())
	}
}

func TestKaaS_WithSecurityGroup_RejectsURIRef(t *testing.T) {
	k := NewKaaS().WithSecurityGroup(URI("/sgs/x"))
	if k.Err() == nil {
		t.Error("expected Err() != nil when passing a URI ref instead of *SecurityGroup")
	}
}

func TestKaaS_WithSecurityGroup_RejectsEmptyName(t *testing.T) {
	sg := NewSecurityGroup() // Name() == ""
	k := NewKaaS().WithSecurityGroup(sg)
	if k.Err() == nil {
		t.Error("expected Err() != nil when SecurityGroup has empty Name")
	}
}

func TestKaaS_WithSecurityGroupName_SetsName(t *testing.T) {
	k := NewKaaS().WithSecurityGroupName("my-sg")
	if k.SecurityGroupName() != "my-sg" {
		t.Errorf("SecurityGroupName() = %q", k.SecurityGroupName())
	}
	if k.Err() != nil {
		t.Errorf("unexpected error: %v", k.Err())
	}
}

func TestKaaS_WithSecurityGroupName_RejectsEmpty(t *testing.T) {
	k := NewKaaS().WithSecurityGroupName("")
	if k.Err() == nil {
		t.Error("expected Err() != nil when name is empty")
	}
}

func TestKaaS_WithNodeCIDR(t *testing.T) {
	k := NewKaaS().WithNodeCIDR("10.100.0.0/16", "node-cidr")
	req := k.RawRequest()
	if req.Properties.NodeCIDR.Address != "10.100.0.0/16" {
		t.Errorf("NodeCIDR.Address = %q", req.Properties.NodeCIDR.Address)
	}
	if req.Properties.NodeCIDR.Name != "node-cidr" {
		t.Errorf("NodeCIDR.Name = %q", req.Properties.NodeCIDR.Name)
	}
}

func TestKaaS_WithPodCIDR(t *testing.T) {
	k := NewKaaS().WithPodCIDR("10.200.0.0/16")
	req := k.RawRequest()
	if req.Properties.PodCIDR == nil || *req.Properties.PodCIDR != "10.200.0.0/16" {
		t.Errorf("PodCIDR = %v", req.Properties.PodCIDR)
	}
}

func TestKaaS_WithHA(t *testing.T) {
	k := NewKaaS().HighlyAvailable()
	req := k.RawRequest()
	if req.Properties.HA == nil || !*req.Properties.HA {
		t.Errorf("HA = %v", req.Properties.HA)
	}
}

func TestKaaS_WithStorage(t *testing.T) {
	k := NewKaaS().WithMaxStorageQuotaGB(200)
	req := k.RawRequest()
	if req.Properties.Storage.MaxCumulativeVolumeSize == nil || *req.Properties.Storage.MaxCumulativeVolumeSize != 200 {
		t.Errorf("Storage.MaxCumulativeVolumeSize = %v", req.Properties.Storage.MaxCumulativeVolumeSize)
	}
}

func TestKaaS_WithKubernetesVersion(t *testing.T) {
	k := NewKaaS().WithKubernetesVersion("1.32.3")
	if k.KubernetesVersion() != "1.32.3" {
		t.Errorf("KubernetesVersion() = %q", k.KubernetesVersion())
	}
}

func TestKaaS_WithBillingPeriod(t *testing.T) {
	k := NewKaaS().BilledBy(BillingPeriodHour)
	if k.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", k.BillingPeriod())
	}
}

func TestKaaS_WithIdentity(t *testing.T) {
	k := NewKaaS().WithIdentity("cid", "csecret")
	req := k.RawRequest()
	if req.Properties.Identity == nil {
		t.Fatal("Identity is nil")
	}
	if req.Properties.Identity.ClientID == nil || *req.Properties.Identity.ClientID != "cid" {
		t.Errorf("Identity.ClientID = %v", req.Properties.Identity.ClientID)
	}
	if req.Properties.Identity.ClientSecret == nil || *req.Properties.Identity.ClientSecret != "csecret" {
		t.Errorf("Identity.ClientSecret = %v", req.Properties.Identity.ClientSecret)
	}
}

// --------------------------------------------------------------------------
// NodePool sub-builder
// --------------------------------------------------------------------------

func TestNodePool_Build_Basic(t *testing.T) {
	np := NewNodePool().
		Named("pool-1").OfInstance(NodePoolInstanceK4A8).InZone(ZoneITBG1).WithCount(3)
	p := np.build()
	if p.Name != "pool-1" {
		t.Errorf("Name = %q", p.Name)
	}
	if p.Instance != NodePoolInstanceK4A8 {
		t.Errorf("Instance = %q", p.Instance)
	}
	if p.Zone != ZoneITBG1 {
		t.Errorf("Zone (dataCenter) = %q", p.Zone)
	}
	if p.Nodes != 3 {
		t.Errorf("Nodes = %d", p.Nodes)
	}
	if p.Autoscaling {
		t.Error("Autoscaling should be false by default")
	}
}

func TestNodePool_Build_Autoscaling(t *testing.T) {
	np := NewNodePool().
		Named("pool-auto").WithCount(3).WithAutoscaling(2, 10)
	p := np.build()
	if !p.Autoscaling {
		t.Error("Autoscaling should be true")
	}
	if p.MinCount == nil || *p.MinCount != 2 {
		t.Errorf("MinCount = %v", p.MinCount)
	}
	if p.MaxCount == nil || *p.MaxCount != 10 {
		t.Errorf("MaxCount = %v", p.MaxCount)
	}
}

func TestKaaS_AddNodePool_DrainErrors(t *testing.T) {
	// An error on the sub-builder should propagate to the parent via AddNodePool.
	k := NewKaaS()
	np := NewNodePool()
	np.addErr(fmt.Errorf("test sub-builder error"))
	k.WithNodePools(np)
	if k.Err() == nil {
		t.Error("expected Err() != nil after AddNodePool with errored sub-builder")
	}
}

func TestKaaS_AddNodePool_Nil(t *testing.T) {
	k := NewKaaS().WithNodePools(nil)
	if k.Err() != nil {
		t.Errorf("WithNodePools(nil) should not error: %v", k.Err())
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestKaaS_ToRequest(t *testing.T) {
	vpcURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1"
	subnetURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1"
	sgFixture := NewSecurityGroup().
		Named("sg-name")

	k := NewKaaS().Named(
		"my-cluster").
		Tagged("t1").Tagged("t2").
		InRegion(RegionITBGBergamo).
		WithVPC(URI(vpcURI)).
		WithSubnet(URI(subnetURI)).
		WithSecurityGroup(sgFixture).
		WithNodeCIDR("10.100.0.0/16", "node-cidr").
		WithPodCIDR("10.200.0.0/16").
		WithKubernetesVersion("1.32.3").
		HighlyAvailable().
		WithMaxStorageQuotaGB(100).
		BilledBy(BillingPeriodHour).
		WithIdentity("cid", "csecret").
		WithNodePools(NewNodePool().
			Named("pool-1").OfInstance(NodePoolInstanceK4A8).InZone(ZoneITBG1).WithCount(3))

	req := k.RawRequest()

	if req.Metadata.Name != "my-cluster" {
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
	if req.Properties.SecurityGroup.Name != "sg-name" {
		t.Errorf("Properties.SecurityGroup.Name = %q", req.Properties.SecurityGroup.Name)
	}
	if req.Properties.NodeCIDR.Address != "10.100.0.0/16" || req.Properties.NodeCIDR.Name != "node-cidr" {
		t.Errorf("NodeCIDR = %+v", req.Properties.NodeCIDR)
	}
	if req.Properties.PodCIDR == nil || *req.Properties.PodCIDR != "10.200.0.0/16" {
		t.Errorf("PodCIDR = %v", req.Properties.PodCIDR)
	}
	if req.Properties.KubernetesVersion.Value != "1.32.3" {
		t.Errorf("KubernetesVersion.Value = %q", req.Properties.KubernetesVersion.Value)
	}
	if req.Properties.HA == nil || !*req.Properties.HA {
		t.Errorf("HA = %v", req.Properties.HA)
	}
	if req.Properties.Storage.MaxCumulativeVolumeSize == nil || *req.Properties.Storage.MaxCumulativeVolumeSize != 100 {
		t.Errorf("Storage = %+v", req.Properties.Storage)
	}
	if req.Properties.BillingPlanCommon == nil || req.Properties.BillingPlanCommon.BillingPeriod == nil || *req.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodHour {
		t.Errorf("BillingPlanCommon.BillingPeriod = %v", req.Properties.BillingPlanCommon)
	}
	if req.Properties.Identity == nil {
		t.Fatal("Identity is nil")
	}
	if req.Properties.Identity.ClientID == nil || *req.Properties.Identity.ClientID != "cid" {
		t.Errorf("Identity.ClientID = %v", req.Properties.Identity.ClientID)
	}
	if len(req.Properties.NodePools) != 1 {
		t.Fatalf("NodePools len = %d", len(req.Properties.NodePools))
	}
	if req.Properties.NodePools[0].Name != "pool-1" {
		t.Errorf("NodePools[0].Name = %q", req.Properties.NodePools[0].Name)
	}
	if req.Properties.NodePools[0].Nodes != 3 {
		t.Errorf("NodePools[0].Nodes = %d", req.Properties.NodePools[0].Nodes)
	}
}

// --------------------------------------------------------------------------
// toUpdateRequest round-trip — only mutable fields
// --------------------------------------------------------------------------

func TestKaaS_ToUpdateRequest_MutableOnly(t *testing.T) {
	vpcURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1"
	k := NewKaaS().Named(
		"updated-cluster").
		InRegion(RegionITBGBergamo).
		WithVPC(URI(vpcURI)). // set but must NOT appear in update request
		WithKubernetesVersion("1.33.0").
		HighlyAvailable().
		WithMaxStorageQuotaGB(200).
		BilledBy(BillingPeriodHour).
		WithNodePools(NewNodePool().
			Named("pool-1").WithCount(5).OfInstance(NodePoolInstanceK4A8).InZone(ZoneITBG1))

	upd := k.toUpdateRequest()

	if upd.Properties.KubernetesVersion.Value != "1.33.0" {
		t.Errorf("KubernetesVersion.Value = %q", upd.Properties.KubernetesVersion.Value)
	}
	if upd.Properties.HA == nil || !*upd.Properties.HA {
		t.Errorf("HA = %v", upd.Properties.HA)
	}
	if upd.Properties.Storage == nil || upd.Properties.Storage.MaxCumulativeVolumeSize == nil || *upd.Properties.Storage.MaxCumulativeVolumeSize != 200 {
		t.Errorf("Storage = %v", upd.Properties.Storage)
	}
	if upd.Properties.BillingPlanCommon == nil || upd.Properties.BillingPlanCommon.BillingPeriod == nil || *upd.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodHour {
		t.Errorf("BillingPlanCommon.BillingPeriod = %v", upd.Properties.BillingPlanCommon)
	}
	if len(upd.Properties.NodePools) != 1 || upd.Properties.NodePools[0].Nodes != 5 {
		t.Errorf("NodePools = %+v", upd.Properties.NodePools)
	}
	// Immutable fields must not be present (VPC is in KaaSPropertiesRequest, not UpdateRequest)
	// — the update type simply doesn't have VPC/Subnet fields, so this is a type-level guarantee.
}

func TestKaaS_ToUpdateRequest_Empty(t *testing.T) {
	k := NewKaaS()
	upd := k.toUpdateRequest() // must not panic
	if upd.Properties.Storage != nil {
		t.Errorf("Storage should be nil when not set, got %v", upd.Properties.Storage)
	}
	if upd.Properties.BillingPlanCommon != nil {
		t.Errorf("BillingPlanCommon should be nil when not set, got %v", upd.Properties.BillingPlanCommon)
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func kaasTestResponse(name string) *types.KaaSResponse {
	id := "kaas-1"
	uri := "/projects/p/providers/Aruba.Container/kaas/kaas-1"
	state := types.State("Active")
	vpcURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1"
	subnetURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1"
	sgName := "sg-name"
	sgURI := "/projects/p/providers/Aruba.Network/vpcs/vpc-1/securityGroups/sg-1"
	nodeCIDRAddr := "10.100.0.0/16"
	nodeCIDRName := "node-cidr"
	podCIDRAddr := "10.200.0.0/16"
	k8sVersion := "1.32.3"
	billingPeriod := BillingPeriodHour
	haTrue := true
	maxVol := int32(100)
	instanceName := string(NodePoolInstanceK4A8)
	dcCode := string(ZoneITBG1)
	poolName := "pool-1"
	poolNodes := int32(3)
	autoFalse := false
	return &types.KaaSResponse{
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
		Properties: types.KaaSPropertiesResponse{
			VPC:    types.ReferenceResourceResponse{URI: &vpcURI},
			Subnet: types.ReferenceResourceResponse{URI: &subnetURI},
			SecurityGroup: types.KaasSecurityGroupPropertiesResponse{
				Name: &sgName,
				URI:  &sgURI,
			},
			NodeCIDR: types.NodeCIDRPropertiesResponse{
				Address: &nodeCIDRAddr,
				Name:    &nodeCIDRName,
			},
			PodCIDR: &types.PodCIDRPropertiesResponse{
				Address: &podCIDRAddr,
			},
			KubernetesVersion: types.KubernetesVersionInfoResponse{
				Value: &k8sVersion,
			},
			HA: &haTrue,
			Storage: &types.StorageKubernetesCommon{
				MaxCumulativeVolumeSize: &maxVol,
			},
			BillingPlanCommon: &types.BillingPlanCommon{BillingPeriod: &billingPeriod},
			NodePools: &[]types.NodePoolPropertiesResponse{
				{
					Name:        &poolName,
					Nodes:       &poolNodes,
					Instance:    &types.KaaSNodePoolInstanceResponse{Name: &instanceName},
					DataCenter:  &types.KaaSNodePoolDataCenterResponse{Code: &dcCode},
					Autoscaling: autoFalse,
				},
			},
		},
		Status: types.ResourceStatusResponse{
			State: &state,
		},
	}
}

func TestKaaS_FromResponseHydration(t *testing.T) {
	k := &KaaS{}
	resp := kaasTestResponse("my-cluster")
	k.fromResponse(resp)

	if k.ID() != "kaas-1" {
		t.Errorf("ID() = %q", k.ID())
	}
	if k.KaaSID() != "kaas-1" {
		t.Errorf("KaaSID() = %q", k.KaaSID())
	}
	if k.URI() != "/projects/p/providers/Aruba.Container/kaas/kaas-1" {
		t.Errorf("URI() = %q", k.URI())
	}
	if k.Name() != "my-cluster" {
		t.Errorf("Name() = %q", k.Name())
	}
	if tags := k.Tags(); len(tags) != 1 || tags[0] != "tag1" {
		t.Errorf("Tags() = %v", tags)
	}
	if k.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", k.Region())
	}
	if k.State() != "Active" {
		t.Errorf("State() = %q", k.State())
	}
	if k.VPC() != "/projects/p/providers/Aruba.Network/vpcs/vpc-1" {
		t.Errorf("VPC() = %q", k.VPC())
	}
	if k.Subnet() != "/projects/p/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1" {
		t.Errorf("Subnet() = %q", k.Subnet())
	}
	if k.SecurityGroupName() != "sg-name" {
		t.Errorf("SecurityGroupName() = %q", k.SecurityGroupName())
	}
	if k.KubernetesVersion() != "1.32.3" {
		t.Errorf("KubernetesVersion() = %q", k.KubernetesVersion())
	}
	if k.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", k.BillingPeriod())
	}
	if k.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", k.ProjectID())
	}
	// NodePools round-trip flattening
	if len(k.nodePools) != 1 {
		t.Fatalf("nodePools len = %d", len(k.nodePools))
	}
	np := k.nodePools[0]
	if np.instance == nil || *np.instance != NodePoolInstanceK4A8 {
		t.Errorf("nodePool.instance = %v", np.instance)
	}
	if np.zone == nil || *np.zone != ZoneITBG1 {
		t.Errorf("nodePool.zone (dataCenter code) = %v", np.zone)
	}
}

func TestKaaS_FromResponse_BackfillsProjectID_FromURI(t *testing.T) {
	id := "kaas-1"
	uri := "/projects/p-uri/providers/Aruba.Container/kaas/kaas-1"
	resp := &types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  &id,
			URI: &uri,
		},
	}
	k := &KaaS{}
	k.fromResponse(resp)
	if k.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() via URI backfill = %q", k.ProjectID())
	}
}

func TestKaaS_FromResponse_Nil(t *testing.T) {
	k := &KaaS{}
	k.fromResponse(nil) // must not panic
	if k.ID() != "" {
		t.Errorf("ID() should be empty after fromResponse(nil)")
	}
}

// --------------------------------------------------------------------------
// kaasIDsFromRef helper
// --------------------------------------------------------------------------

func TestKaaSIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Container/kaas/kaas-1")
	pid, kid, err := kaasIDsFromRef(ref)
	if err != nil || pid != "p" || kid != "kaas-1" {
		t.Errorf("kaasIDsFromRef = (%q, %q, %v)", pid, kid, err)
	}
}

func TestKaaSIDsFromRef_TypedRef(t *testing.T) {
	k := &KaaS{}
	id := "kaas-2"
	uri := "/projects/p2/providers/Aruba.Container/kaas/kaas-2"
	resp := &types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:                      &id,
			URI:                     &uri,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{ID: "p2"},
		},
	}
	k.fromResponse(resp)
	pid, kid, err := kaasIDsFromRef(k)
	if err != nil || pid != "p2" || kid != "kaas-2" {
		t.Errorf("kaasIDsFromRef typed = (%q, %q, %v)", pid, kid, err)
	}
}

func TestKaaSIDsFromRef_BadURI_NoKaaS(t *testing.T) {
	_, _, err := kaasIDsFromRef(URI("/projects/p/providers/Aruba.Container/something/else"))
	if err == nil {
		t.Error("expected error for URI without /kaas/<id>")
	}
}

func TestKaaSIDsFromRef_BadURI_NoProject(t *testing.T) {
	_, _, err := kaasIDsFromRef(URI("/providers/Aruba.Container/kaas/kaas-1"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

// --------------------------------------------------------------------------
// fakeKaaSLowLevel — body-capture / action tests
// --------------------------------------------------------------------------

type fakeKaaSLowLevel struct {
	createFunc             func(ctx context.Context, projectID string, body types.KaaSRequest, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error)
	updateFunc             func(ctx context.Context, projectID, kaasID string, body types.KaaSUpdateRequest, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error)
	getFunc                func(ctx context.Context, projectID, kaasID string, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error)
	deleteFunc             func(ctx context.Context, projectID, kaasID string, params *types.RequestParameters) (*types.Response[any], error)
	listFunc               func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.KaaSListResponse], error)
	downloadKubeconfigFunc func(ctx context.Context, projectID, kaasID string, params *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error)
}

func (f *fakeKaaSLowLevel) Create(ctx context.Context, projectID string, body types.KaaSRequest, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
	return f.createFunc(ctx, projectID, body, params)
}
func (f *fakeKaaSLowLevel) Update(ctx context.Context, projectID, kaasID string, body types.KaaSUpdateRequest, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
	return f.updateFunc(ctx, projectID, kaasID, body, params)
}
func (f *fakeKaaSLowLevel) Get(ctx context.Context, projectID, kaasID string, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
	return f.getFunc(ctx, projectID, kaasID, params)
}
func (f *fakeKaaSLowLevel) Delete(ctx context.Context, projectID, kaasID string, params *types.RequestParameters) (*types.Response[any], error) {
	return f.deleteFunc(ctx, projectID, kaasID, params)
}
func (f *fakeKaaSLowLevel) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.KaaSListResponse], error) {
	return f.listFunc(ctx, projectID, params)
}
func (f *fakeKaaSLowLevel) DownloadKubeconfig(ctx context.Context, projectID, kaasID string, params *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error) {
	return f.downloadKubeconfigFunc(ctx, projectID, kaasID, params)
}

// --------------------------------------------------------------------------
// HTTP-mock adapter helper
// --------------------------------------------------------------------------

func buildKaaSTestAdapter(t *testing.T, handler http.HandlerFunc) *kaasClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newKaaSClientAdapter(testutil.NewClient(t, server.URL))
}

const kaasSuccessBody = `{` +
	`"metadata":{"id":"kaas-1","name":"my-cluster","uri":"/projects/p/providers/Aruba.Container/kaas/kaas-1","project":{"id":"p"}},` +
	`"properties":{` +
	`"vpc":{"uri":"/projects/p/providers/Aruba.Network/vpcs/vpc-1"},` +
	`"subnet":{"uri":"/projects/p/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1"},` +
	`"securityGroup":{"name":"sg-name"},` +
	`"nodeCidr":{"address":"10.100.0.0/16","name":"node-cidr"},` +
	`"kubernetesVersion":{"value":"1.32.3"},` +
	`"billingPlan":{"billingPeriod":"Hour"},` +
	`"ha":true` +
	`},` +
	`"status":{"state":"Active"}}`

// --------------------------------------------------------------------------
// Create adapter tests
// --------------------------------------------------------------------------

func TestKaaSClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.KaaSRequest
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "kaas") {
			t.Errorf("path %q should contain 'kaas'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, kaasSuccessBody)
	})

	sgFixture := NewSecurityGroup().
		Named("sg-name")
	k := NewKaaS().
		InProject(URI("/projects/p")).
		Named("my-cluster").
		InRegion(RegionITBGBergamo).
		WithVPC(URI("/projects/p/providers/Aruba.Network/vpcs/vpc-1")).
		WithSubnet(URI("/projects/p/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1")).
		WithSecurityGroup(sgFixture).
		WithNodeCIDR("10.100.0.0/16", "node-cidr").
		WithKubernetesVersion("1.32.3").
		BilledBy(BillingPeriodHour).
		HighlyAvailable().
		WithNodePools(NewNodePool().
			Named("pool-1").WithCount(3).OfInstance(NodePoolInstanceK4A8).InZone(ZoneITBG1))

	result, err := adapter.Create(context.Background(), k)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "kaas-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-cluster" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	// Wire body assertions
	if gotBody.Metadata.Name != "my-cluster" {
		t.Errorf("request Metadata.Name = %q", gotBody.Metadata.Name)
	}
	if gotBody.Properties.VPC.URI != "/projects/p/providers/Aruba.Network/vpcs/vpc-1" {
		t.Errorf("request Properties.VPC.URI = %q", gotBody.Properties.VPC.URI)
	}
	if len(gotBody.Properties.NodePools) != 1 || gotBody.Properties.NodePools[0].Name != "pool-1" {
		t.Errorf("request NodePools = %+v", gotBody.Properties.NodePools)
	}
	// actions must be set so DownloadKubeconfig works after Create
	if result.actions == nil {
		t.Error("actions should be set after Create")
	}
}

func TestKaaSClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})
	_, err := adapter.Create(context.Background(), NewKaaS().
		Named("x"))
	if err == nil {
		t.Fatal("expected error when KaaS has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestKaaSClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" field — triggers MetadataValidationError from low-level Validate()
		fmt.Fprint(w, `{"metadata":{"name":"cluster","uri":"/projects/p/providers/Aruba.Container/kaas/x"},"properties":{},"status":{}}`)
	})

	k := NewKaaS().InProject(URI("/projects/p")).
		Named("cluster")
	result, err := adapter.Create(context.Background(), k)
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

func TestKaaSClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "name is required", 422))
	})

	k := NewKaaS().InProject(URI("/projects/p"))
	result, err := adapter.Create(context.Background(), k)
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

func TestKaaSClientAdapter_Create_WithBodyRefs_ViaFake(t *testing.T) {
	var gotBody types.KaaSRequest
	fake := &fakeKaaSLowLevel{
		createFunc: func(_ context.Context, _ string, body types.KaaSRequest, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
			gotBody = body
			id := "kaas-x"
			uri := "/projects/p/providers/Aruba.Container/kaas/kaas-x"
			return &types.Response[types.KaaSResponse]{
				StatusCode: http.StatusCreated,
				Data: &types.KaaSResponse{
					Metadata: types.ResourceMetadataResponse{
						ID:                      &id,
						URI:                     &uri,
						ProjectMetadataResponse: &types.ProjectMetadataResponse{ID: "p"},
					},
				},
			}, nil
		},
	}
	adapter := &kaasClientAdapter{low: fake}

	k := NewKaaS().
		InProject(URI("/projects/p")).
		WithVPC(URI("/projects/p/providers/Aruba.Network/vpcs/vpc-1")).
		WithSubnet(URI("/projects/p/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1"))

	result, err := adapter.Create(context.Background(), k)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if gotBody.Properties.VPC.URI != "/projects/p/providers/Aruba.Network/vpcs/vpc-1" {
		t.Errorf("VPC.URI = %q", gotBody.Properties.VPC.URI)
	}
	if gotBody.Properties.Subnet.URI != "/projects/p/providers/Aruba.Network/vpcs/vpc-1/subnets/sn-1" {
		t.Errorf("Subnet.URI = %q", gotBody.Properties.Subnet.URI)
	}
}

// --------------------------------------------------------------------------
// Update adapter tests
// --------------------------------------------------------------------------

func TestKaaSClientAdapter_Update_Success(t *testing.T) {
	var gotBody types.KaaSUpdateRequest
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kaasSuccessBody)
	})

	k := &KaaS{}
	id := "kaas-1"
	uri := "/projects/p/providers/Aruba.Container/kaas/kaas-1"
	k.fromResponse(&types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:                      &id,
			URI:                     &uri,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{ID: "p"},
		},
	})
	k.WithKubernetesVersion("1.33.0").
		WithNodePools(NewNodePool().
			Named("pool-1").WithCount(5).OfInstance(NodePoolInstanceK4A8).InZone(ZoneITBG1))

	result, err := adapter.Update(context.Background(), k)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if result.ID() != "kaas-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	// The update request should use KubernetesVersionInfoUpdateRequest
	if gotBody.Properties.KubernetesVersion.Value != "1.33.0" {
		t.Errorf("update KubernetesVersion.Value = %q", gotBody.Properties.KubernetesVersion.Value)
	}
}

func TestKaaSClientAdapter_Update_NoID(t *testing.T) {
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.Update(context.Background(), NewKaaS().InProject(URI("/projects/p")))
	if err == nil {
		t.Fatal("expected error when KaaS has no ID")
	}
}

func TestKaaSClientAdapter_Update_NoProject(t *testing.T) {
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	k := &KaaS{}
	id := "kaas-1"
	k.meta = nil // ensure projectID is unset but ID is set via fake
	k.errMixin = errMixin{}
	// Manually set ID without project
	uri := "/providers/Aruba.Container/kaas/kaas-1"
	k.fromResponse(&types.KaaSResponse{Metadata: types.ResourceMetadataResponse{ID: &id, URI: &uri}})
	_, err := adapter.Update(context.Background(), k)
	if err == nil {
		t.Fatal("expected error when KaaS has no project")
	}
}

func TestKaaSClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "cluster not found", 404))
	})

	k := &KaaS{}
	id := "kaas-1"
	uri := "/projects/p/providers/Aruba.Container/kaas/kaas-1"
	k.fromResponse(&types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:                      &id,
			URI:                     &uri,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{ID: "p"},
		},
	})
	_, err := adapter.Update(context.Background(), k)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T", err)
	}
}

// --------------------------------------------------------------------------
// Get adapter tests
// --------------------------------------------------------------------------

func TestKaaSClientAdapter_Get_URIRef(t *testing.T) {
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if !containsSubstring(r.URL.Path, "kaas-1") {
			t.Errorf("path %q should contain 'kaas-1'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kaasSuccessBody)
	})

	result, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Container/kaas/kaas-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "kaas-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	// actions must be set so DownloadKubeconfig works after Get
	if result.actions == nil {
		t.Error("actions should be set after Get")
	}
}

func TestKaaSClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kaasSuccessBody)
	})

	existing := &KaaS{}
	id := "kaas-1"
	uri := "/projects/p/providers/Aruba.Container/kaas/kaas-1"
	existing.fromResponse(&types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:                      &id,
			URI:                     &uri,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{ID: "p"},
		},
	})

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "kaas-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

// --------------------------------------------------------------------------
// Delete adapter tests
// --------------------------------------------------------------------------

func TestKaaSClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Container/kaas/kaas-1"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestKaaSClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "cluster not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Container/kaas/missing"))
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

func TestKaaSClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"kaas-1","name":"c1","uri":"/projects/p/providers/Aruba.Container/kaas/kaas-1","project":{"id":"p"}},"properties":{"kubernetesVersion":{"value":"1.32.3"},"billingPlan":{"billingPeriod":"Hour"}},"status":{}},`+
			`{"metadata":{"id":"kaas-2","name":"c2","uri":"/projects/p/providers/Aruba.Container/kaas/kaas-2","project":{"id":"p"}},"properties":{"kubernetesVersion":{"value":"1.31.0"},"billingPlan":{"billingPeriod":"Hour"}},"status":{}}`+
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
	if items[0].ID() != "kaas-1" || items[0].Name() != "c1" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[1].ID() != "kaas-2" || items[1].BillingPeriod() != BillingPeriodHour {
		t.Errorf("items[1] ID=%q BillingPeriod=%q", items[1].ID(), items[1].BillingPeriod())
	}
	if items[0].ProjectID() != "p" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
	// actions must be set on list items for DownloadKubeconfig
	if items[0].actions == nil {
		t.Error("actions should be set on list items")
	}
}

// --------------------------------------------------------------------------
// DownloadKubeconfig action tests
// --------------------------------------------------------------------------

func TestKaaS_DownloadKubeconfig_Success(t *testing.T) {
	fake := &fakeKaaSLowLevel{
		downloadKubeconfigFunc: func(_ context.Context, projectID, kaasID string, _ *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error) {
			if projectID != "p" || kaasID != "kaas-1" {
				return nil, fmt.Errorf("unexpected ids: %s/%s", projectID, kaasID)
			}
			return &types.Response[types.KaaSKubeconfigResponse]{
				StatusCode: http.StatusOK,
				Data: &types.KaaSKubeconfigResponse{
					Name:    "cluster-kubeconfig.yaml",
					Content: "apiVersion: v1\nkind: Config\n",
				},
			}, nil
		},
	}
	adapter := &kaasClientAdapter{low: fake}

	k := &KaaS{}
	id := "kaas-1"
	uri := "/projects/p/providers/Aruba.Container/kaas/kaas-1"
	k.fromResponse(&types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:                      &id,
			URI:                     &uri,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{ID: "p"},
		},
	})
	k.actions = adapter

	data, err := k.DownloadKubeconfig(context.Background())
	if err != nil {
		t.Fatalf("DownloadKubeconfig error: %v", err)
	}
	if string(data) != "apiVersion: v1\nkind: Config\n" {
		t.Errorf("kubeconfig content = %q", string(data))
	}
}

func TestKaaS_DownloadKubeconfig_NoActionExecutor(t *testing.T) {
	k := NewKaaS()
	id := "kaas-1"
	uri := "/projects/p/providers/Aruba.Container/kaas/kaas-1"
	k.fromResponse(&types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:                      &id,
			URI:                     &uri,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{ID: "p"},
		},
	})
	// k.actions is nil — locally-built wrapper

	_, err := k.DownloadKubeconfig(context.Background())
	if err == nil {
		t.Fatal("expected error when actions is nil")
	}
	if !containsSubstring(err.Error(), "no action executor") {
		t.Errorf("error should mention 'no action executor', got %q", err.Error())
	}
}

func TestKaaS_DownloadKubeconfig_NonTwoXX(t *testing.T) {
	fake := &fakeKaaSLowLevel{
		downloadKubeconfigFunc: func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error) {
			return &types.Response[types.KaaSKubeconfigResponse]{
				StatusCode: http.StatusNotFound,
			}, nil
		},
	}
	adapter := &kaasClientAdapter{low: fake}

	k := &KaaS{}
	id := "kaas-1"
	uri := "/projects/p/providers/Aruba.Container/kaas/kaas-1"
	k.fromResponse(&types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:                      &id,
			URI:                     &uri,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{ID: "p"},
		},
	})
	k.actions = adapter

	_, err := k.DownloadKubeconfig(context.Background())
	if err == nil {
		t.Fatal("expected error on 404")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T", err)
	}
}

// --------------------------------------------------------------------------
// Reflective check: KaaSClient has Update method
// --------------------------------------------------------------------------

func TestKaaSClient_HasUpdateMethod(t *testing.T) {
	iface := reflect.TypeOf((*KaaSClient)(nil)).Elem()
	for i := range iface.NumMethod() {
		if iface.Method(i).Name == "Update" {
			return // found — test passes
		}
	}
	t.Fatal("KaaSClient must have an Update method")
}

// --------------------------------------------------------------------------
// Shape A — RemoveTag / ReplaceTags / InRegion
// --------------------------------------------------------------------------

func TestKaaS_RemoveTag_ReplaceTags_InRegion(t *testing.T) {
	k := NewKaaS().
		Tagged("a").
		Tagged("b").
		Untagged("a").
		RetaggedAs("x", "y").
		InRegion("ITMI-Milano-1")

	tags := k.Tags()
	if len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("Tags() = %v, want [x y]", tags)
	}
	if k.Region() != "ITMI-Milano-1" {
		t.Errorf("Region() = %q, want ITMI-Milano-1", k.Region())
	}
}

// --------------------------------------------------------------------------
// Shape B — Raw() after fromResponse
// --------------------------------------------------------------------------

func TestKaaS_Raw_AfterFromResponse(t *testing.T) {
	k := &KaaS{}
	resp := kaasTestResponse("cluster-raw")
	k.fromResponse(resp)

	if k.Raw() == nil {
		t.Fatal("Raw() should be non-nil after fromResponse")
	}
	if k.Raw() != resp {
		t.Error("Raw() should return the exact response pointer passed to fromResponse")
	}
	// Also exercise RawRequest on a locally-built wrapper (already covered, but
	// ensure the call path for an un-hydrated wrapper is not missed)
	_ = NewKaaS().RawRequest()
}

// --------------------------------------------------------------------------
// Shape D — WithAPIServerAccessProfile round-trip
// --------------------------------------------------------------------------

func TestKaaS_WithAPIServerAccessProfile_RoundTrip(t *testing.T) {
	ranges := []string{"10.0.0.0/8", "192.168.0.0/16"}
	profile := &types.KaaSAPIServerAccessProfilePropertiesRequest{
		AuthorizedIPRanges:   &ranges,
		EnablePrivateCluster: true,
	}

	k := NewKaaS().WithAPIServerAccessProfile(profile)
	req := k.RawRequest()

	if req.Properties.APIServerAccessProfile == nil {
		t.Fatal("APIServerAccessProfile should be non-nil in request")
	}
	if !req.Properties.APIServerAccessProfile.EnablePrivateCluster {
		t.Error("EnablePrivateCluster should be true")
	}
	if req.Properties.APIServerAccessProfile.AuthorizedIPRanges == nil ||
		len(*req.Properties.APIServerAccessProfile.AuthorizedIPRanges) != 2 {
		t.Errorf("AuthorizedIPRanges = %v", req.Properties.APIServerAccessProfile.AuthorizedIPRanges)
	}
}

func TestKaaS_WithPrivateCluster_RoundTrip(t *testing.T) {
	k := NewKaaS().WithPrivateCluster()
	req := k.RawRequest()

	if req.Properties.APIServerAccessProfile == nil {
		t.Fatal("APIServerAccessProfile should be non-nil in request")
	}
	if !req.Properties.APIServerAccessProfile.EnablePrivateCluster {
		t.Error("EnablePrivateCluster should be true")
	}
	if !k.APIServerPrivateCluster() {
		t.Error("APIServerPrivateCluster() getter should return true")
	}
}

func TestKaaS_WithAuthorizedIPRanges_RoundTrip(t *testing.T) {
	k := NewKaaS().WithAuthorizedIPRanges("10.0.0.0/8", "192.168.0.0/16")
	req := k.RawRequest()

	if req.Properties.APIServerAccessProfile == nil {
		t.Fatal("APIServerAccessProfile should be non-nil in request")
	}
	if req.Properties.APIServerAccessProfile.AuthorizedIPRanges == nil ||
		len(*req.Properties.APIServerAccessProfile.AuthorizedIPRanges) != 2 {
		t.Errorf("AuthorizedIPRanges = %v", req.Properties.APIServerAccessProfile.AuthorizedIPRanges)
	}
	got := k.APIServerAuthorizedIPRanges()
	if len(got) != 2 || got[0] != "10.0.0.0/8" {
		t.Errorf("APIServerAuthorizedIPRanges() = %v", got)
	}
}

func TestKaaS_WithAuthorizedIPRanges_Clear(t *testing.T) {
	k := NewKaaS().WithAuthorizedIPRanges("10.0.0.0/8").WithAuthorizedIPRanges()
	req := k.RawRequest()
	if req.Properties.APIServerAccessProfile != nil {
		t.Error("clearing ranges should leave no APIServerAccessProfile in request")
	}
}

func TestKaaS_EnablePrivateCluster_RoundTrip(t *testing.T) {
	k := NewKaaS().EnablePrivateCluster()
	req := k.RawRequest()
	if req.Properties.APIServerAccessProfile == nil {
		t.Fatal("APIServerAccessProfile should be non-nil in request")
	}
	if !req.Properties.APIServerAccessProfile.EnablePrivateCluster {
		t.Error("EnablePrivateCluster should be true")
	}
	if !k.APIServerPrivateCluster() {
		t.Error("APIServerPrivateCluster() getter should return true")
	}
}

func TestKaaS_WithAPIServerAccessProfile_NilClears(t *testing.T) {
	k := NewKaaS().EnablePrivateCluster().WithAPIServerAccessProfile(nil)
	if k.RawRequest().Properties.APIServerAccessProfile != nil {
		t.Error("passing nil should clear the profile")
	}
}

func TestKaaS_APIServerAuthorizedIPRanges_ReturnsCopy(t *testing.T) {
	k := NewKaaS().WithAuthorizedIPRanges("10.0.0.0/8")
	got := k.APIServerAuthorizedIPRanges()
	got[0] = "mutated"
	if k.APIServerAuthorizedIPRanges()[0] != "10.0.0.0/8" {
		t.Error("APIServerAuthorizedIPRanges() must return a copy, not the internal slice")
	}
}

// --------------------------------------------------------------------------
// Shape E — KaaS adapter error paths
// --------------------------------------------------------------------------

// Get_BadRef: URI that can't be resolved to KaaS/project IDs.
func TestKaaSClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.Get(context.Background(), URI("/projects/p/no-kaas-segment"))
	if err == nil {
		t.Fatal("expected error for URI that lacks /kaas/<id>")
	}
}

// Get_NonTwoXX: stub returns 404 → HTTPError.
func TestKaaSClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "not found", 404))
	})
	_, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Container/kaas/kaas-1"))
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

// Delete_BadRef: URI without kaas segment → error before HTTP.
func TestKaaSClientAdapter_Delete_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusNoContent)
	})
	err := adapter.Delete(context.Background(), URI("/projects/p/no-kaas-segment"))
	if err == nil {
		t.Fatal("expected error for bad ref")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made for bad ref")
	}
}

// List_BadRef: parent ref without project ID → error before HTTP.
func TestKaaSClientAdapter_List_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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
func TestKaaSClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildKaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
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
// Shape E — preActionCheck missing-ID path
// --------------------------------------------------------------------------

func TestKaaS_DownloadKubeconfig_MissingKaaSID(t *testing.T) {
	fake := &fakeKaaSLowLevel{
		downloadKubeconfigFunc: func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error) {
			return &types.Response[types.KaaSKubeconfigResponse]{StatusCode: 200, Data: &types.KaaSKubeconfigResponse{}}, nil
		},
	}
	adapter := &kaasClientAdapter{low: fake}

	// Has actions and project, but no ID.
	k := NewKaaS()
	k.projectID = "p"
	k.actions = adapter

	_, err := k.DownloadKubeconfig(context.Background())
	if err == nil {
		t.Fatal("expected error when KaaS has no ID")
	}
	if !containsSubstring(err.Error(), "missing KaaS ID") {
		t.Errorf("error should mention 'missing KaaS ID', got %q", err.Error())
	}
}

// DownloadKubeconfig_NilData: success status but nil Data → returns nil bytes.
func TestKaaS_DownloadKubeconfig_NilData(t *testing.T) {
	fake := &fakeKaaSLowLevel{
		downloadKubeconfigFunc: func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error) {
			return &types.Response[types.KaaSKubeconfigResponse]{StatusCode: 200, Data: nil}, nil
		},
	}
	adapter := &kaasClientAdapter{low: fake}

	k := &KaaS{}
	id := "kaas-1"
	uri := "/projects/p/providers/Aruba.Container/kaas/kaas-1"
	k.fromResponse(&types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:                      &id,
			URI:                     &uri,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{ID: "p"},
		},
	})
	k.actions = adapter

	data, err := k.DownloadKubeconfig(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data != nil {
		t.Errorf("expected nil data for nil response Data, got %v", data)
	}
}

// --------------------------------------------------------------------------
// Shape F — Accessors on zero-value KaaS
// --------------------------------------------------------------------------

func TestKaaS_Accessors_ZeroValue(t *testing.T) {
	k := NewKaaS()

	if k.VPC() != "" {
		t.Errorf("VPC() = %q, want empty", k.VPC())
	}
	if k.Subnet() != "" {
		t.Errorf("Subnet() = %q, want empty", k.Subnet())
	}
	if k.SecurityGroupName() != "" {
		t.Errorf("SecurityGroupName() = %q, want empty", k.SecurityGroupName())
	}
	if k.KubernetesVersion() != "" {
		t.Errorf("KubernetesVersion() = %q, want empty", k.KubernetesVersion())
	}
	if k.BillingPeriod() != "" {
		t.Errorf("BillingPeriod() = %q, want empty", k.BillingPeriod())
	}
	if k.Raw() != nil {
		t.Errorf("Raw() = %v, want nil", k.Raw())
	}
	if k.URI() != "" {
		t.Errorf("URI() = %q, want empty", k.URI())
	}
	if k.KaaSID() != "" {
		t.Errorf("KaaSID() = %q, want empty", k.KaaSID())
	}
}

// --------------------------------------------------------------------------
// kaasRebuildNodePools — nil-field branches
// --------------------------------------------------------------------------

func TestKaasRebuildNodePools_NilOptionalFields(t *testing.T) {
	// Construct a response pool where Name, Nodes, Instance, DataCenter,
	// MinCount, and MaxCount are all nil to exercise the nil-guard branches.
	autoTrue := true
	pools := &[]types.NodePoolPropertiesResponse{
		{
			// Name: nil
			// Nodes: nil
			Instance:    nil,
			DataCenter:  nil,
			MinCount:    nil,
			MaxCount:    nil,
			Autoscaling: autoTrue,
		},
	}
	result := kaasRebuildNodePools(pools)
	if len(result) != 1 {
		t.Fatalf("expected 1 pool, got %d", len(result))
	}
	np := result[0]
	if np.name != nil {
		t.Errorf("name should be nil, got %v", np.name)
	}
	if np.instance != nil {
		t.Errorf("instance should be nil, got %v", np.instance)
	}
	if np.zone != nil {
		t.Errorf("zone should be nil, got %v", np.zone)
	}
	if np.autoscaling == nil || !*np.autoscaling {
		t.Errorf("autoscaling should be true")
	}
}

func TestKaasRebuildNodePools_WithMinMaxCount(t *testing.T) {
	// Cover the MinCount and MaxCount non-nil branches.
	min := int32(2)
	max := int32(8)
	autoFalse := false
	pools := &[]types.NodePoolPropertiesResponse{
		{
			MinCount:    &min,
			MaxCount:    &max,
			Autoscaling: autoFalse,
		},
	}
	result := kaasRebuildNodePools(pools)
	if len(result) != 1 {
		t.Fatalf("expected 1 pool, got %d", len(result))
	}
	np := result[0]
	if np.minCount == nil || *np.minCount != 2 {
		t.Errorf("minCount = %v, want 2", np.minCount)
	}
	if np.maxCount == nil || *np.maxCount != 8 {
		t.Errorf("maxCount = %v, want 8", np.maxCount)
	}
}

// --------------------------------------------------------------------------
// Additional coverage: DownloadKubeconfig missing-projectID + low-level error
// --------------------------------------------------------------------------

// MissingProjectID: has actions and ID, but no projectID.
func TestKaaS_DownloadKubeconfig_MissingProjectID(t *testing.T) {
	fake := &fakeKaaSLowLevel{
		downloadKubeconfigFunc: func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error) {
			return &types.Response[types.KaaSKubeconfigResponse]{StatusCode: 200, Data: &types.KaaSKubeconfigResponse{}}, nil
		},
	}
	adapter := &kaasClientAdapter{low: fake}

	// Has ID but no project (URI has no /projects segment).
	k := NewKaaS()
	id := "kaas-1"
	uri := "/providers/Aruba.Container/kaas/kaas-1"
	k.fromResponse(&types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  &id,
			URI: &uri,
		},
	})
	k.actions = adapter

	_, err := k.DownloadKubeconfig(context.Background())
	if err == nil {
		t.Fatal("expected error when KaaS has no projectID")
	}
	if !containsSubstring(err.Error(), "missing project ID") {
		t.Errorf("error should mention 'missing project ID', got %q", err.Error())
	}
}

// LowLevelError: low-level client returns error → propagate.
func TestKaaS_DownloadKubeconfig_LowLevelError(t *testing.T) {
	fake := &fakeKaaSLowLevel{
		downloadKubeconfigFunc: func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error) {
			return nil, fmt.Errorf("network timeout")
		},
	}
	adapter := &kaasClientAdapter{low: fake}

	k := &KaaS{}
	id := "kaas-1"
	uri := "/projects/p/providers/Aruba.Container/kaas/kaas-1"
	k.fromResponse(&types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:                      &id,
			URI:                     &uri,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{ID: "p"},
		},
	})
	k.actions = adapter

	_, err := k.DownloadKubeconfig(context.Background())
	if err == nil {
		t.Fatal("expected error from low-level client")
	}
	if !containsSubstring(err.Error(), "network timeout") {
		t.Errorf("error should mention 'network timeout', got %q", err.Error())
	}
}

// --------------------------------------------------------------------------
// Additional coverage: KaaS adapter Update/Get error paths via fake
// --------------------------------------------------------------------------

// Update_WithErr: k.Err() is non-nil → return early.
func TestKaaSClientAdapter_Update_WithErr(t *testing.T) {
	callCount := 0
	fake := &fakeKaaSLowLevel{
		updateFunc: func(_ context.Context, _, _ string, _ types.KaaSUpdateRequest, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
			callCount++
			return nil, nil
		},
	}
	adapter := &kaasClientAdapter{low: fake}

	k := NewKaaS()
	k.addErr(fmt.Errorf("pre-existing error"))
	// Give it an ID and project so those checks pass.
	id := "kaas-1"
	uri := "/projects/p/providers/Aruba.Container/kaas/kaas-1"
	k.fromResponse(&types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:                      &id,
			URI:                     &uri,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{ID: "p"},
		},
	})
	k.addErr(fmt.Errorf("pre-existing error after hydration"))

	_, err := adapter.Update(context.Background(), k)
	if err == nil {
		t.Fatal("expected error when k.Err() is set")
	}
	if callCount != 0 {
		t.Error("no low-level call should be made when k.Err() is set")
	}
}

// Get_LowLevelError: low-level Get returns error → propagate.
func TestKaaSClientAdapter_Get_LowLevelError(t *testing.T) {
	fake := &fakeKaaSLowLevel{
		getFunc: func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	adapter := &kaasClientAdapter{low: fake}

	_, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Container/kaas/kaas-1"))
	if err == nil {
		t.Fatal("expected error from low-level Get")
	}
	if !containsSubstring(err.Error(), "connection refused") {
		t.Errorf("error should mention 'connection refused', got %q", err.Error())
	}
}

func TestKaaS_FromResponse_SetsStatus(t *testing.T) {
	k := &KaaS{}
	state := types.State("Active")
	k.fromResponse(&types.KaaSResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if k.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", k.State())
	}
}

func TestKaaSClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, kaasSuccessBody)
	})
	adapter := newKaaSClientAdapter(testutil.NewClient(t, server.URL))
	k, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Container/kaas/kaas-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&k.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned KaaS")
	}
}

// --------------------------------------------------------------------------
// ClearNodePools / ReplaceNodePools / SetNodePools (#279)
// --------------------------------------------------------------------------

func TestKaaS_ReplaceNodePools_ReplacesExisting(t *testing.T) {
	np1 := NewNodePool().Named("pool-1").OfInstance("inst-1").WithAutoscaling(1, 1)
	np2 := NewNodePool().Named("pool-2").OfInstance("inst-2").WithAutoscaling(2, 2)
	np3 := NewNodePool().Named("pool-3").OfInstance("inst-3").WithAutoscaling(3, 3)

	k := NewKaaS().Named("c").InRegion(RegionITBGBergamo).
		WithVPC(URI("/v")).WithSubnet(URI("/s")).
		WithNodePools(np1).WithNodePools(np2)

	k.ReplaceNodePools(np3)

	req := k.toRequest()
	if len(req.Properties.NodePools) != 1 {
		t.Fatalf("expected 1 node pool after ReplaceNodePools, got %d", len(req.Properties.NodePools))
	}
	if req.Properties.NodePools[0].Name != "pool-3" {
		t.Errorf("expected pool-3, got %q", req.Properties.NodePools[0].Name)
	}
}

func TestKaaS_ClearNodePools_RemovesAll(t *testing.T) {
	np := NewNodePool().Named("pool-1").OfInstance("inst-1").WithAutoscaling(1, 2)
	k := NewKaaS().Named("c").InRegion(RegionITBGBergamo).
		WithVPC(URI("/v")).WithSubnet(URI("/s")).
		WithNodePools(np)

	k.WithoutNodePools()

	req := k.toRequest()
	if len(req.Properties.NodePools) != 0 {
		t.Errorf("expected 0 node pools after ClearNodePools, got %d", len(req.Properties.NodePools))
	}
}

func TestKaaS_SetNodePools_AliasForReplace(t *testing.T) {
	np1 := NewNodePool().Named("pool-1").OfInstance("inst-1").WithAutoscaling(1, 1)
	np2 := NewNodePool().Named("pool-2").OfInstance("inst-2").WithAutoscaling(1, 1)

	k := NewKaaS().Named("c").InRegion(RegionITBGBergamo).
		WithVPC(URI("/v")).WithSubnet(URI("/s")).
		WithNodePools(np1)

	k.ReplaceNodePools(np2)

	req := k.toRequest()
	if len(req.Properties.NodePools) != 1 {
		t.Fatalf("expected 1 node pool after SetNodePools, got %d", len(req.Properties.NodePools))
	}
	if req.Properties.NodePools[0].Name != "pool-2" {
		t.Errorf("expected pool-2, got %q", req.Properties.NodePools[0].Name)
	}
}

// --------------------------------------------------------------------------
// PodCIDR getter
// --------------------------------------------------------------------------

func TestKaaS_PodCIDR_Unset(t *testing.T) {
	k := &KaaS{}
	if got := k.PodCIDR(); got != "" {
		t.Errorf("PodCIDR() = %q, want empty", got)
	}
}

func TestKaaS_PodCIDR_WithPodCIDR(t *testing.T) {
	k := NewKaaS().WithPodCIDR("10.200.0.0/16")
	if got := k.PodCIDR(); got != "10.200.0.0/16" {
		t.Errorf("PodCIDR() = %q, want 10.200.0.0/16", got)
	}
}

func TestKaaS_PodCIDR_FromResponse(t *testing.T) {
	k := &KaaS{}
	k.fromResponse(kaasTestResponse("cluster"))
	if got := k.PodCIDR(); got != "10.200.0.0/16" {
		t.Errorf("PodCIDR() = %q, want 10.200.0.0/16", got)
	}
}

// --------------------------------------------------------------------------
// NodeCIDR getter
// --------------------------------------------------------------------------

func TestKaaS_NodeCIDR_Unset(t *testing.T) {
	k := &KaaS{}
	if got := k.NodeCIDR(); got != "" {
		t.Errorf("NodeCIDR() = %q, want empty", got)
	}
}

func TestKaaS_NodeCIDR_WithNodeCIDR(t *testing.T) {
	k := NewKaaS().WithNodeCIDR("10.100.0.0/16", "node-cidr")
	if got := k.NodeCIDR(); got != "10.100.0.0/16" {
		t.Errorf("NodeCIDR() = %q, want 10.100.0.0/16", got)
	}
}

func TestKaaS_NodeCIDR_FromResponse(t *testing.T) {
	k := &KaaS{}
	k.fromResponse(kaasTestResponse("cluster"))
	if got := k.NodeCIDR(); got != "10.100.0.0/16" {
		t.Errorf("NodeCIDR() = %q, want 10.100.0.0/16", got)
	}
}

// --------------------------------------------------------------------------
// IdentityClientID getter
// --------------------------------------------------------------------------

func TestKaaS_IdentityClientID_Unset(t *testing.T) {
	k := &KaaS{}
	if got := k.IdentityClientID(); got != "" {
		t.Errorf("IdentityClientID() = %q, want empty", got)
	}
}

func TestKaaS_IdentityClientID_WithIdentity(t *testing.T) {
	k := NewKaaS().WithIdentity("client-abc", "secret-xyz")
	if got := k.IdentityClientID(); got != "client-abc" {
		t.Errorf("IdentityClientID() = %q, want client-abc", got)
	}
}

func TestKaaS_IdentityClientID_FromResponse(t *testing.T) {
	clientID := "resp-client-id"
	resp := kaasTestResponse("cluster")
	resp.Properties.Identity = &types.IdentityPropertiesResponse{ClientID: &clientID}
	k := &KaaS{}
	k.fromResponse(resp)
	if got := k.IdentityClientID(); got != "resp-client-id" {
		t.Errorf("IdentityClientID() = %q, want resp-client-id", got)
	}
}

// --------------------------------------------------------------------------
// NodePools getter
// --------------------------------------------------------------------------

func TestKaaS_NodePools_NilWhenEmpty(t *testing.T) {
	k := &KaaS{}
	if pools := k.NodePools(); pools != nil {
		t.Errorf("NodePools() = %v, want nil", pools)
	}
}

func TestKaaS_NodePools_ReturnsConfigured(t *testing.T) {
	np := NewNodePool().Named("pool-1").OfInstance("inst-1").WithAutoscaling(1, 3)
	k := NewKaaS().WithNodePools(np)
	pools := k.NodePools()
	if len(pools) != 1 {
		t.Fatalf("NodePools() len = %d, want 1", len(pools))
	}
	if pools[0].Name() != "pool-1" {
		t.Errorf("NodePools()[0].Name() = %q, want pool-1", pools[0].Name())
	}
}

func TestKaaS_NodePools_FromResponse(t *testing.T) {
	k := &KaaS{}
	k.fromResponse(kaasTestResponse("cluster"))
	pools := k.NodePools()
	if len(pools) != 1 {
		t.Fatalf("NodePools() len = %d, want 1", len(pools))
	}
	if pools[0].Name() != "pool-1" {
		t.Errorf("NodePools()[0].Name() = %q, want pool-1", pools[0].Name())
	}
}
