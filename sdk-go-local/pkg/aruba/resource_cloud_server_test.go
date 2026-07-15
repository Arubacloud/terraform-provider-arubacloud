package aruba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Compile-time Ref satisfaction
// --------------------------------------------------------------------------

var _ Ref = (*CloudServer)(nil)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestCloudServer_FluentSetters(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-1", "my-project", "/projects/p-1"))

	cs := NewCloudServer().
		InProject(proj).
		Named("my-server").
		Tagged("compute").
		Tagged("prod").
		InRegion(RegionITBGBergamo).
		InZone(ZoneITBG1).
		OfFlavor(CloudServerFlavorCSO2A4).
		WithUserData("dGVzdA==").
		WithVPCPreset()

	if cs.Name() != "my-server" {
		t.Errorf("Name() = %q", cs.Name())
	}
	if tags := cs.Tags(); len(tags) != 2 || tags[0] != "compute" || tags[1] != "prod" {
		t.Errorf("Tags() = %v", tags)
	}
	if cs.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", cs.Region())
	}
	if cs.Zone() != ZoneITBG1 {
		t.Errorf("Zone() = %q", cs.Zone())
	}
	if cs.Flavor() != CloudServerFlavorCSO2A4 {
		t.Errorf("Flavor() = %q", cs.Flavor())
	}
	if cs.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", cs.ProjectID())
	}
	if cs.Err() != nil {
		t.Errorf("Err() = %v", cs.Err())
	}

	cs.Untagged("compute")
	if tags := cs.Tags(); len(tags) != 1 || tags[0] != "prod" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	cs.RetaggedAs("x", "y")
	if tags := cs.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

// --------------------------------------------------------------------------
// IntoProject — typed / URI / bad Ref
// --------------------------------------------------------------------------

func TestCloudServer_IntoProject_TypedRef(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-42", "n", "/projects/p-42"))

	cs := NewCloudServer().InProject(proj)
	if cs.ProjectID() != "p-42" {
		t.Errorf("ProjectID() = %q", cs.ProjectID())
	}
	if cs.Err() != nil {
		t.Errorf("Err() = %v", cs.Err())
	}
}

func TestCloudServer_IntoProject_URIRef(t *testing.T) {
	cs := NewCloudServer().InProject(URI("/projects/p-uri"))
	if cs.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", cs.ProjectID())
	}
	if cs.Err() != nil {
		t.Errorf("Err() = %v", cs.Err())
	}
}

func TestCloudServer_IntoProject_BadRef(t *testing.T) {
	cs := NewCloudServer().InProject(URI("/something/else"))
	if cs.Err() == nil {
		t.Error("expected Err() to be set for unresolvable parent")
	}
}

// --------------------------------------------------------------------------
// Single body-ref setters
// --------------------------------------------------------------------------

func TestCloudServer_WithVPC_URIRef(t *testing.T) {
	cs := NewCloudServer().WithVPC(URI("/projects/p/network/vpcs/v-1"))
	if cs.VPC() != "/projects/p/network/vpcs/v-1" {
		t.Errorf("VPC() = %q", cs.VPC())
	}
	if cs.Err() != nil {
		t.Errorf("Err() = %v", cs.Err())
	}
}

func TestCloudServer_WithVPC_EmptyURI(t *testing.T) {
	cs := NewCloudServer().WithVPC(URI(""))
	if cs.Err() == nil {
		t.Error("expected Err() to be set for empty VPC URI")
	}
	if cs.vpcRef != nil {
		t.Error("vpcRef should remain nil on error")
	}
}

func TestCloudServer_WithBootVolume_URIRef(t *testing.T) {
	cs := NewCloudServer().BootingFrom(URI("/projects/p/storage/volumes/vol-1"))
	if cs.BootVolume() != "/projects/p/storage/volumes/vol-1" {
		t.Errorf("BootVolume() = %q", cs.BootVolume())
	}
	if cs.Err() != nil {
		t.Errorf("Err() = %v", cs.Err())
	}
}

func TestCloudServer_WithBootVolume_EmptyURI(t *testing.T) {
	cs := NewCloudServer().BootingFrom(URI(""))
	if cs.Err() == nil {
		t.Error("expected Err() for empty BootVolume URI")
	}
}

func TestCloudServer_WithKeyPair_URIRef(t *testing.T) {
	cs := NewCloudServer().UsingKeyPair(URI("/projects/p/providers/Aruba.Compute/keyPairs/kp-1"))
	if cs.KeyPair() != "/projects/p/providers/Aruba.Compute/keyPairs/kp-1" {
		t.Errorf("KeyPair() = %q", cs.KeyPair())
	}
	if cs.Err() != nil {
		t.Errorf("Err() = %v", cs.Err())
	}
}

func TestCloudServer_WithKeyPair_EmptyURI(t *testing.T) {
	cs := NewCloudServer().UsingKeyPair(URI(""))
	if cs.Err() == nil {
		t.Error("expected Err() for empty KeyPair URI")
	}
}

func TestCloudServer_WithElasticIP_URIRef(t *testing.T) {
	cs := NewCloudServer().WithElasticIP(URI("/projects/p/network/elasticips/eip-1"))
	if cs.Err() != nil {
		t.Errorf("Err() = %v", cs.Err())
	}
}

func TestCloudServer_WithElasticIP_EmptyURI(t *testing.T) {
	cs := NewCloudServer().WithElasticIP(URI(""))
	if cs.Err() == nil {
		t.Error("expected Err() for empty ElasticIP URI")
	}
}

// --------------------------------------------------------------------------
// Multi-ref slice setters
// --------------------------------------------------------------------------

func TestCloudServer_AddSubnet_AppendsTwo(t *testing.T) {
	cs := NewCloudServer().
		OnSubnets(URI("/subnets/s-1")).
		OnSubnets(URI("/subnets/s-2"))
	if cs.Err() != nil {
		t.Errorf("Err() = %v", cs.Err())
	}
	req := cs.toRequest()
	if len(req.Properties.Subnets) != 2 {
		t.Fatalf("Subnets len = %d", len(req.Properties.Subnets))
	}
	if req.Properties.Subnets[0].URI != "/subnets/s-1" {
		t.Errorf("Subnets[0].URI = %q", req.Properties.Subnets[0].URI)
	}
	if req.Properties.Subnets[1].URI != "/subnets/s-2" {
		t.Errorf("Subnets[1].URI = %q", req.Properties.Subnets[1].URI)
	}
}

func TestCloudServer_AddSubnet_EmptyURI(t *testing.T) {
	cs := NewCloudServer().OnSubnets(URI(""))
	if cs.Err() == nil {
		t.Error("expected Err() for empty Subnet URI")
	}
	if len(cs.subnetRefs) != 0 {
		t.Error("subnetRefs should remain empty on error")
	}
}

func TestCloudServer_AddSecurityGroup_AppendsTwo(t *testing.T) {
	cs := NewCloudServer().
		WithSecurityGroups(URI("/sgs/sg-1")).
		WithSecurityGroups(URI("/sgs/sg-2"))
	if cs.Err() != nil {
		t.Errorf("Err() = %v", cs.Err())
	}
	req := cs.toRequest()
	if len(req.Properties.SecurityGroups) != 2 {
		t.Fatalf("SecurityGroups len = %d", len(req.Properties.SecurityGroups))
	}
}

func TestCloudServer_AddSecurityGroup_EmptyURI(t *testing.T) {
	cs := NewCloudServer().WithSecurityGroups(URI(""))
	if cs.Err() == nil {
		t.Error("expected Err() for empty SecurityGroup URI")
	}
	if len(cs.securityGroupRefs) != 0 {
		t.Error("securityGroupRefs should remain empty on error")
	}
}

// --------------------------------------------------------------------------
// Scalar setters
// --------------------------------------------------------------------------

func TestCloudServer_OfFlavor(t *testing.T) {
	cs := NewCloudServer().OfFlavor(CloudServerFlavorCSO2A4)
	if cs.Flavor() != CloudServerFlavorCSO2A4 {
		t.Errorf("Flavor() = %q", cs.Flavor())
	}
}

func TestCloudServer_WithUserData(t *testing.T) {
	cs := NewCloudServer().WithUserData("dGVzdA==")
	req := cs.toRequest()
	if req.Properties.UserData == nil || *req.Properties.UserData != "dGVzdA==" {
		t.Error("UserData not emitted correctly")
	}
}

func TestCloudServer_WithVPCPreset(t *testing.T) {
	cs := NewCloudServer().WithVPCPreset()
	req := cs.toRequest()
	if !req.Properties.VPCPreset {
		t.Error("VPCPreset not emitted correctly")
	}
}

func TestCloudServer_InZone(t *testing.T) {
	cs := NewCloudServer().InZone(ZoneITBG1)
	if cs.Zone() != ZoneITBG1 {
		t.Errorf("Zone() = %q", cs.Zone())
	}
	req := cs.toRequest()
	if req.Properties.Zone != ZoneITBG1 {
		t.Errorf("request Zone = %q", req.Properties.Zone)
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestCloudServer_ToRequestRoundTrip(t *testing.T) {
	cs := NewCloudServer().
		InProject(URI("/projects/p")).
		Named("srv").
		Tagged("tag1").
		InRegion(RegionITBGBergamo).
		InZone(ZoneITBG1).
		OfFlavor(CloudServerFlavorCSO2A4).
		WithVPC(URI("/vpcs/v")).
		BootingFrom(URI("/vols/bv")).
		UsingKeyPair(URI("/kps/kp")).
		WithElasticIP(URI("/eips/eip")).
		OnSubnets(URI("/subnets/s")).
		WithSecurityGroups(URI("/sgs/sg")).
		WithUserData("dA==")

	req := cs.RawRequest()

	if req.Metadata.Name != "srv" {
		t.Errorf("request name = %q", req.Metadata.Name)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("request location = %q", req.Metadata.Location.Value)
	}
	if req.Properties.Zone != ZoneITBG1 {
		t.Errorf("Zone = %q", req.Properties.Zone)
	}
	if req.Properties.FlavorName == nil || *req.Properties.FlavorName != CloudServerFlavorCSO2A4 {
		t.Errorf("FlavorName = %v", req.Properties.FlavorName)
	}
	if req.Properties.VPC.URI != "/vpcs/v" {
		t.Errorf("VPC.URI = %q", req.Properties.VPC.URI)
	}
	if req.Properties.BootVolume.URI != "/vols/bv" {
		t.Errorf("BootVolume.URI = %q", req.Properties.BootVolume.URI)
	}
	if req.Properties.KeyPair.URI != "/kps/kp" {
		t.Errorf("KeyPair.URI = %q", req.Properties.KeyPair.URI)
	}
	if req.Properties.ElasticIP.URI != "/eips/eip" {
		t.Errorf("ElasticIP.URI = %q", req.Properties.ElasticIP.URI)
	}
	if len(req.Properties.Subnets) != 1 || req.Properties.Subnets[0].URI != "/subnets/s" {
		t.Errorf("Subnets = %v", req.Properties.Subnets)
	}
	if len(req.Properties.SecurityGroups) != 1 || req.Properties.SecurityGroups[0].URI != "/sgs/sg" {
		t.Errorf("SecurityGroups = %v", req.Properties.SecurityGroups)
	}
	if req.Properties.UserData == nil || *req.Properties.UserData != "dA==" {
		t.Errorf("UserData = %v", req.Properties.UserData)
	}
}

func TestCloudServer_ToRequest_AllUnset(t *testing.T) {
	cs := &CloudServer{}
	req := cs.toRequest() // must not panic
	if req.Properties.FlavorName != nil {
		t.Error("FlavorName should be nil when unset")
	}
	if req.Properties.Zone != "" {
		t.Error("Zone should be empty when unset")
	}
	if len(req.Properties.Subnets) != 0 {
		t.Error("Subnets should be empty when unset")
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func cloudServerTestResponse(id, name, uri string) *types.CloudServerResponse {
	loc := &types.LocationResponse{Value: RegionITBGBergamo}
	state := types.State("Running")
	return &types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			URI:              &uri,
			Name:             &name,
			Tags:             []string{"tag1"},
			LocationResponse: loc,
		},
		Properties: types.CloudServerPropertiesResponse{
			Zone:       ZoneITBG1,
			Flavor:     types.CloudServerFlavorResponse{Name: CloudServerFlavorCSO2A4, CPU: 2, RAM: 4096},
			VPC:        types.ReferenceResourceCommon{URI: "/vpcs/v"},
			BootVolume: types.ReferenceResourceCommon{URI: "/vols/bv"},
			KeyPair:    types.ReferenceResourceCommon{URI: "/kps/kp"},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}
}

func TestCloudServer_FromResponseHydration(t *testing.T) {
	cs := &CloudServer{}
	resp := cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1")
	cs.fromResponse(resp)

	if cs.ID() != "cs-1" {
		t.Errorf("ID() = %q", cs.ID())
	}
	if cs.URI() != "/projects/p/providers/Aruba.Compute/cloudServers/cs-1" {
		t.Errorf("URI() = %q", cs.URI())
	}
	if cs.CloudServerID() != "cs-1" {
		t.Errorf("CloudServerID() = %q", cs.CloudServerID())
	}
	if cs.Name() != "my-server" {
		t.Errorf("Name() = %q", cs.Name())
	}
	if tags := cs.Tags(); len(tags) != 1 || tags[0] != "tag1" {
		t.Errorf("Tags() = %v", tags)
	}
	if cs.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", cs.Region())
	}
	if cs.Zone() != ZoneITBG1 {
		t.Errorf("Zone() = %q", cs.Zone())
	}
	// Flavor reads from response Flavor.Name
	if cs.Flavor() != CloudServerFlavorCSO2A4 {
		t.Errorf("Flavor() = %q", cs.Flavor())
	}
	if cs.FlavorRaw() == nil || cs.FlavorRaw().CPU != 2 {
		t.Error("FlavorRaw() should carry full struct")
	}
	if cs.VPC() != "/vpcs/v" {
		t.Errorf("VPC() = %q", cs.VPC())
	}
	if cs.BootVolume() != "/vols/bv" {
		t.Errorf("BootVolume() = %q", cs.BootVolume())
	}
	if cs.KeyPair() != "/kps/kp" {
		t.Errorf("KeyPair() = %q", cs.KeyPair())
	}
	if cs.State() != "Running" {
		t.Errorf("State() = %q", cs.State())
	}
	if cs.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
	// ProjectID backfilled from URI when ProjectMetadataResponse is nil
	if cs.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", cs.ProjectID())
	}
}

func TestCloudServer_FromResponseURIBackfill(t *testing.T) {
	id := "cs-99"
	uri := "/projects/p-uri/providers/Aruba.Compute/cloudServers/cs-99"
	resp := &types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  &id,
			URI: &uri,
		},
	}
	cs := &CloudServer{}
	cs.fromResponse(resp)

	if cs.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() via URI backfill = %q", cs.ProjectID())
	}
}

func TestCloudServer_FromResponse_NilSafe(t *testing.T) {
	cs := &CloudServer{}
	cs.fromResponse(nil)
	if cs.ID() != "" || cs.URI() != "" || cs.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}

	cs2 := &CloudServer{}
	cs2.fromResponse(&types.CloudServerResponse{})
	if cs2.ID() != "" || cs2.URI() != "" || cs2.Flavor() != "" {
		t.Error("empty response should yield zero accessor values")
	}
}

// --------------------------------------------------------------------------
// New getters: ElasticIP, Subnets, SecurityGroups, UserData
// --------------------------------------------------------------------------

func TestCloudServer_ElasticIP_Getter(t *testing.T) {
	cs := NewCloudServer().WithElasticIP(URI("/eips/eip-1"))
	if cs.ElasticIP() != "/eips/eip-1" {
		t.Errorf("ElasticIP() = %q", cs.ElasticIP())
	}
	// not set
	cs2 := NewCloudServer()
	if cs2.ElasticIP() != "" {
		t.Errorf("ElasticIP() unset = %q", cs2.ElasticIP())
	}
}

func TestCloudServer_UserData_Getter(t *testing.T) {
	cs := NewCloudServer().WithUserData("dGVzdA==")
	if cs.UserData() != "dGVzdA==" {
		t.Errorf("UserData() = %q", cs.UserData())
	}
	cs2 := NewCloudServer()
	if cs2.UserData() != "" {
		t.Errorf("UserData() unset = %q", cs2.UserData())
	}
}

func TestCloudServer_SecurityGroups_Getter(t *testing.T) {
	cs := NewCloudServer().
		WithSecurityGroups(URI("/sgs/sg-1")).
		WithSecurityGroups(URI("/sgs/sg-2"))
	sgs := cs.SecurityGroups()
	if len(sgs) != 2 || sgs[0] != "/sgs/sg-1" || sgs[1] != "/sgs/sg-2" {
		t.Errorf("SecurityGroups() = %v", sgs)
	}
	cs2 := NewCloudServer()
	if cs2.SecurityGroups() != nil {
		t.Errorf("SecurityGroups() unset = %v", cs2.SecurityGroups())
	}
}

func TestCloudServer_FromResponse_RehydratesSubnetRefs(t *testing.T) {
	sub1 := "/subnets/s-1"
	sub2 := "/subnets/s-2"
	resp := &types.CloudServerResponse{
		Properties: types.CloudServerPropertiesResponse{
			NetworkInterfaces: []types.CloudServerNetworkInterfaceResponse{
				{Subnet: &sub1},
				{Subnet: &sub2},
			},
		},
	}
	cs := &CloudServer{}
	cs.fromResponse(resp)

	subnets := cs.Subnets()
	if len(subnets) != 2 {
		t.Fatalf("Subnets() after fromResponse = %v (want 2)", subnets)
	}
	if subnets[0] != sub1 || subnets[1] != sub2 {
		t.Errorf("Subnets() = %v", subnets)
	}

	// toRequest should include the rehydrated subnets
	req := cs.toRequest()
	if len(req.Properties.Subnets) != 2 {
		t.Fatalf("toRequest().Subnets len = %d (want 2)", len(req.Properties.Subnets))
	}
	if req.Properties.Subnets[0].URI != sub1 || req.Properties.Subnets[1].URI != sub2 {
		t.Errorf("toRequest().Subnets = %v", req.Properties.Subnets)
	}
}

func TestCloudServer_Subnets_FallbackToLocal(t *testing.T) {
	cs := NewCloudServer().OnSubnets(URI("/subnets/local"))
	subnets := cs.Subnets()
	if len(subnets) != 1 || subnets[0] != "/subnets/local" {
		t.Errorf("Subnets() fallback = %v", subnets)
	}
}

// --------------------------------------------------------------------------
// Flavor asymmetry (request vs response)
// --------------------------------------------------------------------------

func TestCloudServer_Flavor_Asymmetry(t *testing.T) {
	// Before any response: Flavor() returns what OfFlavor set.
	cs := NewCloudServer().OfFlavor(CloudServerFlavorCSO2A4)
	if cs.Flavor() != CloudServerFlavorCSO2A4 {
		t.Errorf("pre-response Flavor() = %q", cs.Flavor())
	}
	if cs.FlavorRaw() != nil {
		t.Error("FlavorRaw() should be nil before hydration")
	}

	// After fromResponse: Flavor() returns the response Flavor.Name (response wins).
	cs.fromResponse(&types.CloudServerResponse{
		Properties: types.CloudServerPropertiesResponse{
			Flavor: types.CloudServerFlavorResponse{Name: CloudServerFlavorCSO4A8, CPU: 4, RAM: 8192},
		},
	})
	if cs.Flavor() != CloudServerFlavorCSO4A8 {
		t.Errorf("post-response Flavor() = %q (expected response value to win)", cs.Flavor())
	}
	if cs.FlavorRaw() == nil || cs.FlavorRaw().CPU != 4 {
		t.Error("FlavorRaw() should carry full response struct")
	}
}

// --------------------------------------------------------------------------
// Ref + ancestor ID satisfaction (runtime)
// --------------------------------------------------------------------------

func TestCloudServer_RefSatisfaction(t *testing.T) {
	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-99", "n", "/projects/p99/providers/Aruba.Compute/cloudServers/cs-99"))

	// withCloudServerID typed path
	csID, ok := extractID(cs, func(ref Ref) (string, bool) {
		if w, ok := ref.(withCloudServerID); ok {
			return w.CloudServerID(), true
		}
		return "", false
	}, "cloudServers")
	if !ok || csID != "cs-99" {
		t.Errorf("extractID via withCloudServerID = (%q, %v)", csID, ok)
	}

	// withProjectID typed path
	pid, ok := extractID(cs, func(ref Ref) (string, bool) {
		if w, ok := ref.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || pid != "p99" {
		t.Errorf("extractID via withProjectID = (%q, %v)", pid, ok)
	}
}

// --------------------------------------------------------------------------
// cloudServerIDsFromRef helper
// --------------------------------------------------------------------------

func TestCloudServerIDsFromRef_TypedRef(t *testing.T) {
	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "n", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	pid, csID, err := cloudServerIDsFromRef(cs)
	if err != nil || pid != "p" || csID != "cs-1" {
		t.Errorf("cloudServerIDsFromRef typed = (%q, %q, %v)", pid, csID, err)
	}
}

func TestCloudServerIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Compute/cloudServers/cs-1")
	pid, csID, err := cloudServerIDsFromRef(ref)
	if err != nil || pid != "p" || csID != "cs-1" {
		t.Errorf("cloudServerIDsFromRef URI = (%q, %q, %v)", pid, csID, err)
	}
}

func TestCloudServerIDsFromRef_BadURI_MissingCloudServer(t *testing.T) {
	_, _, err := cloudServerIDsFromRef(URI("/projects/p/providers/Aruba.Compute"))
	if err == nil {
		t.Error("expected error for URI without /cloudServers/<id>")
	}
}

func TestCloudServerIDsFromRef_BadURI_MissingProject(t *testing.T) {
	_, _, err := cloudServerIDsFromRef(URI("/providers/Aruba.Compute/cloudServers/cs-1"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

func TestCloudServerIDsFromRef_BadURI_MissingAll(t *testing.T) {
	_, _, err := cloudServerIDsFromRef(URI("/something/else"))
	if err == nil {
		t.Error("expected error for completely unrelated URI")
	}
}

// --------------------------------------------------------------------------
// cloudServersClientAdapter — HTTP mock tests
// --------------------------------------------------------------------------

func buildCloudServersTestAdapter(t *testing.T, handler http.HandlerFunc) *cloudServersClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newCloudServersClientAdapter(testutil.NewClient(t, server.URL))
}

const cloudServerSuccessBody = `{` +
	`"metadata":{"id":"cs-1","name":"my-server","uri":"/projects/p/providers/Aruba.Compute/cloudServers/cs-1"},` +
	`"properties":{"dataCenter":"ITBG-1","flavor":{"name":"CSO2A4","cpu":2,"ram":4096},"vpc":{"uri":"/vpcs/v"},"bootVolume":{"uri":"/vols/bv"}},` +
	`"status":{"state":"Running"}}`

func TestCloudServersClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.CloudServerRequest
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "cloudServers") {
			t.Errorf("path %q should contain 'cloudServers'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, cloudServerSuccessBody)
	})

	cs := NewCloudServer().
		InProject(URI("/projects/p")).
		Named("my-server").
		InZone(ZoneITBG1).
		OfFlavor(CloudServerFlavorCSO2A4).
		WithVPC(URI("/vpcs/v")).
		BootingFrom(URI("/vols/bv")).
		UsingKeyPair(URI("/kps/kp")).
		OnSubnets(URI("/subnets/s"))

	result, err := adapter.Create(context.Background(), cs)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "cs-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-server" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.Flavor() != CloudServerFlavorCSO2A4 {
		t.Errorf("Flavor() = %q", result.Flavor())
	}
	if result.State() != "Running" {
		t.Errorf("State() = %q", result.State())
	}
	if result.actions == nil {
		t.Error("actions executor should be set after Create")
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	// Verify wire body: flavorName field (wire name, mapped from OfFlavor)
	if gotBody.Properties.FlavorName == nil || *gotBody.Properties.FlavorName != CloudServerFlavorCSO2A4 {
		t.Errorf("request FlavorName = %v", gotBody.Properties.FlavorName)
	}
	if gotBody.Properties.VPC.URI != "/vpcs/v" {
		t.Errorf("request VPC.URI = %q", gotBody.Properties.VPC.URI)
	}
	if len(gotBody.Properties.Subnets) != 1 || gotBody.Properties.Subnets[0].URI != "/subnets/s" {
		t.Errorf("request Subnets = %v", gotBody.Properties.Subnets)
	}
}

func TestCloudServersClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	_, err := adapter.Create(context.Background(), NewCloudServer().
		Named("x").OfFlavor(CloudServerFlavorCSO2A4))
	if err == nil {
		t.Fatal("expected error when CloudServer has no parent project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without parent project")
	}
}

func TestCloudServersClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" field — triggers MetadataValidationError from the low-level client.
		fmt.Fprint(w, `{"metadata":{"name":"srv","uri":"/projects/p/providers/Aruba.Compute/cloudServers/x"},"properties":{}}`)
	})

	cs := NewCloudServer().InProject(URI("/projects/p")).
		Named("srv").OfFlavor(CloudServerFlavorCSO2A4)
	result, err := adapter.Create(context.Background(), cs)
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

func TestCloudServersClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "zone is required", 422))
	})

	cs := NewCloudServer().InProject(URI("/projects/p")).
		Named("srv")
	result, err := adapter.Create(context.Background(), cs)
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

func TestCloudServersClientAdapter_Update_Success(t *testing.T) {
	var capturedMethod string
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		if !containsSubstring(r.URL.Path, "cloudServers") {
			t.Errorf("path %q should contain 'cloudServers'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, cloudServerSuccessBody)
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.Named("renamed-server") // trigger metadata PUT

	result, err := adapter.Update(context.Background(), cs)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if capturedMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", capturedMethod)
	}
	if result.ID() != "cs-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.actions == nil {
		t.Error("actions executor should be set after Update")
	}
}

func TestCloudServersClientAdapter_Update_NoID(t *testing.T) {
	callCount := 0
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	cs := NewCloudServer().InProject(URI("/projects/p")).
		Named("x")
	_, err := adapter.Update(context.Background(), cs)
	if err == nil {
		t.Fatal("expected error when CloudServer has no ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without ID")
	}
}

func TestCloudServersClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, cloudServerSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Compute/cloudServers/cs-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "cs-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	if result.actions == nil {
		t.Error("actions executor should be set after Get")
	}
	if !containsSubstring(capturedPath, "cloudServers") {
		t.Errorf("path %q should contain 'cloudServers'", capturedPath)
	}
}

func TestCloudServersClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, cloudServerSuccessBody)
	})

	existing := &CloudServer{}
	existing.fromResponse(cloudServerTestResponse("cs-1", "n", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "cs-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
}

func TestCloudServersClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestCloudServersClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "cloud server not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Compute/cloudServers/missing"))
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

func TestCloudServersClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"cs-1","name":"srv1","uri":"/projects/p/providers/Aruba.Compute/cloudServers/cs-1"},"properties":{"dataCenter":"ITBG-1","flavor":{"name":"CSO2A4"}}},`+
			`{"metadata":{"id":"cs-2","name":"srv2","uri":"/projects/p/providers/Aruba.Compute/cloudServers/cs-2"},"properties":{"dataCenter":"ITBG-2","flavor":{"name":"CSO4A8"}}}`+
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
	if items[0].ID() != "cs-1" || items[0].Name() != "srv1" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[0].Flavor() != CloudServerFlavorCSO2A4 {
		t.Errorf("items[0].Flavor() = %q", items[0].Flavor())
	}
	if items[1].ID() != "cs-2" || items[1].Flavor() != CloudServerFlavorCSO4A8 {
		t.Errorf("items[1] ID=%q Flavor=%q", items[1].ID(), items[1].Flavor())
	}
	if items[0].ProjectID() != "p" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
	if items[0].actions == nil || items[1].actions == nil {
		t.Error("actions executor should be set on all List items")
	}
}

// --------------------------------------------------------------------------
// Action methods — PowerOn / PowerOff / SetPassword
// --------------------------------------------------------------------------

func TestCloudServer_PowerOn_Success(t *testing.T) {
	var capturedPath string
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		state := "Running"
		body := cloudServerSuccessBody
		_ = state
		fmt.Fprint(w, body)
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.actions = adapter

	if err := cs.PowerOn(context.Background()); err != nil {
		t.Fatalf("PowerOn error: %v", err)
	}
	if !containsSubstring(capturedPath, "poweron") {
		t.Errorf("path %q should contain 'poweron'", capturedPath)
	}
	// Status re-hydrated from response
	if cs.State() != "Running" {
		t.Errorf("State() after PowerOn = %q", cs.State())
	}
}

func TestCloudServer_PowerOn_NoExecutor(t *testing.T) {
	cs := NewCloudServer()
	err := cs.PowerOn(context.Background())
	if err == nil {
		t.Fatal("expected error from PowerOn without action executor")
	}
	if !containsSubstring(err.Error(), "no action executor") {
		t.Errorf("error message = %q", err.Error())
	}
}

func TestCloudServer_PowerOn_MissingID(t *testing.T) {
	callCount := 0
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	cs := &CloudServer{}
	cs.projectID = "p"
	cs.actions = adapter // actions set but ID empty

	err := cs.PowerOn(context.Background())
	if err == nil {
		t.Fatal("expected error when CloudServer ID is empty")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without CloudServer ID")
	}
}

func TestCloudServer_PowerOff_Success(t *testing.T) {
	var capturedPath string
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, cloudServerSuccessBody)
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.actions = adapter

	if err := cs.PowerOff(context.Background()); err != nil {
		t.Fatalf("PowerOff error: %v", err)
	}
	if !containsSubstring(capturedPath, "poweroff") {
		t.Errorf("path %q should contain 'poweroff'", capturedPath)
	}
}

func TestCloudServer_PowerOff_NonTwoXX(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Conflict", "server is already off", 409))
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "n", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.actions = adapter

	err := cs.PowerOff(context.Background())
	if err == nil {
		t.Fatal("expected error on 409")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusConflict {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

func TestCloudServer_SetPassword_Success(t *testing.T) {
	var capturedPath string
	var gotBody types.CloudServerPasswordRequest
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "n", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.actions = adapter

	if err := cs.SetPassword(context.Background(), "s3cr3t"); err != nil {
		t.Fatalf("SetPassword error: %v", err)
	}
	if !containsSubstring(capturedPath, "password") {
		t.Errorf("path %q should contain 'password'", capturedPath)
	}
	if gotBody.Password != "s3cr3t" {
		t.Errorf("request password = %q", gotBody.Password)
	}
}

func TestCloudServer_SetPassword_NonTwoXX(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Bad Request", "password too short", 400))
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "n", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.actions = adapter

	err := cs.SetPassword(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error on 400")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// InRegion setter
// --------------------------------------------------------------------------

func TestCloudServer_InRegion(t *testing.T) {
	cs := NewCloudServer().InRegion("ITMI-Milano-1")
	if cs.Region() != "ITMI-Milano-1" {
		t.Errorf("Region() after InRegion = %q", cs.Region())
	}
	req := cs.toRequest()
	if req.Metadata.Location.Value != "ITMI-Milano-1" {
		t.Errorf("request Location.Value = %q", req.Metadata.Location.Value)
	}
}

// --------------------------------------------------------------------------
// Template and NetworkInterfaces response-only accessors
// --------------------------------------------------------------------------

func TestCloudServer_Template_AndNetworkInterfaces_AfterHydration(t *testing.T) {
	cs := &CloudServer{}
	// Before hydration both return zero values.
	if cs.Template() != "" {
		t.Errorf("Template() before hydration = %q, want empty", cs.Template())
	}
	if cs.NetworkInterfaces() != nil {
		t.Errorf("NetworkInterfaces() before hydration = %v, want nil", cs.NetworkInterfaces())
	}

	// Hydrate with a response that carries template and network-interface data.
	templateURI := "/templates/tmpl-1"
	mac := "aa:bb:cc:dd:ee:ff"
	iface := types.CloudServerNetworkInterfaceResponse{
		MacAddress: &mac,
		IPs:        []string{"10.0.0.1"},
	}
	resp := cloudServerTestResponse("cs-1", "srv", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1")
	resp.Properties.Template = types.ReferenceResourceCommon{URI: templateURI}
	resp.Properties.NetworkInterfaces = []types.CloudServerNetworkInterfaceResponse{iface}
	cs.fromResponse(resp)

	if cs.Template() != templateURI {
		t.Errorf("Template() = %q, want %q", cs.Template(), templateURI)
	}
	if got := cs.NetworkInterfaces(); len(got) != 1 || got[0].MacAddress == nil || *got[0].MacAddress != mac {
		t.Errorf("NetworkInterfaces() = %v", got)
	}
}

// --------------------------------------------------------------------------
// preActionCheck — missing project ID
// --------------------------------------------------------------------------

func TestCloudServer_PowerOff_MissingProjectID(t *testing.T) {
	callCount := 0
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	// CloudServer has an ID but no project ID.
	cs := &CloudServer{}
	id := "cs-1"
	uri := "/providers/Aruba.Compute/cloudServers/cs-1" // no /projects/ segment
	cs.fromResponse(&types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, URI: &uri},
	})
	cs.actions = adapter

	err := cs.PowerOff(context.Background())
	if err == nil {
		t.Fatal("expected error when project ID is empty")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project ID")
	}
}

// --------------------------------------------------------------------------
// Delta setters — AssociateSubnets / DisassociateSubnets / etc.
// --------------------------------------------------------------------------

func TestCloudServer_AssociateSubnets_Appends(t *testing.T) {
	cs := NewCloudServer().
		AssociateSubnets(URI("/subnets/s-1"), URI("/subnets/s-2")).
		DisassociateSubnets(URI("/subnets/s-3"))
	if cs.Err() != nil {
		t.Fatalf("Err() = %v", cs.Err())
	}
	if len(cs.subnetsToAssociate) != 2 || cs.subnetsToAssociate[0] != "/subnets/s-1" {
		t.Errorf("subnetsToAssociate = %v", cs.subnetsToAssociate)
	}
	if len(cs.subnetsToDisassociate) != 1 || cs.subnetsToDisassociate[0] != "/subnets/s-3" {
		t.Errorf("subnetsToDisassociate = %v", cs.subnetsToDisassociate)
	}
}

func TestCloudServer_AssociateSubnets_EmptyURIError(t *testing.T) {
	cs := NewCloudServer().AssociateSubnets(URI(""))
	if cs.Err() == nil {
		t.Error("expected Err() for empty URI in AssociateSubnets")
	}
}

func TestCloudServer_AssociateSecurityGroups_Appends(t *testing.T) {
	cs := NewCloudServer().
		AssociateSecurityGroups(URI("/sgs/sg-1")).
		DisassociateSecurityGroups(URI("/sgs/sg-2"))
	if cs.Err() != nil {
		t.Fatalf("Err() = %v", cs.Err())
	}
	if len(cs.sgsToAssociate) != 1 || cs.sgsToAssociate[0] != "/sgs/sg-1" {
		t.Errorf("sgsToAssociate = %v", cs.sgsToAssociate)
	}
	if len(cs.sgsToDisassociate) != 1 || cs.sgsToDisassociate[0] != "/sgs/sg-2" {
		t.Errorf("sgsToDisassociate = %v", cs.sgsToDisassociate)
	}
}

func TestCloudServer_AssociateElasticIPs_Appends(t *testing.T) {
	cs := NewCloudServer().
		AssociateElasticIPs(URI("/eips/e-1")).
		DisassociateElasticIPs(URI("/eips/e-2"))
	if cs.Err() != nil {
		t.Fatalf("Err() = %v", cs.Err())
	}
	if len(cs.eipsToAssociate) != 1 || cs.eipsToAssociate[0] != "/eips/e-1" {
		t.Errorf("eipsToAssociate = %v", cs.eipsToAssociate)
	}
}

func TestCloudServer_AttachDataVolumes_Appends(t *testing.T) {
	cs := NewCloudServer().
		AttachDataVolumes(URI("/vols/v-1")).
		DetachDataVolumes(URI("/vols/v-2"))
	if cs.Err() != nil {
		t.Fatalf("Err() = %v", cs.Err())
	}
	if len(cs.dataVolumesToAttach) != 1 || cs.dataVolumesToAttach[0] != "/vols/v-1" {
		t.Errorf("dataVolumesToAttach = %v", cs.dataVolumesToAttach)
	}
	if len(cs.dataVolumesToDetach) != 1 || cs.dataVolumesToDetach[0] != "/vols/v-2" {
		t.Errorf("dataVolumesToDetach = %v", cs.dataVolumesToDetach)
	}
}

// --------------------------------------------------------------------------
// hasMetadataChanges
// --------------------------------------------------------------------------

func TestCloudServer_HasMetadataChanges_NilResponse(t *testing.T) {
	cs := NewCloudServer().Named("srv")
	if !cs.hasMetadataChanges() {
		t.Error("expected true when response is nil")
	}
}

func TestCloudServer_HasMetadataChanges_SameName(t *testing.T) {
	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	if cs.hasMetadataChanges() {
		t.Error("expected false when name and tags match the hydrated response")
	}
}

func TestCloudServer_HasMetadataChanges_RenamedServer(t *testing.T) {
	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.Named("new-name")
	if !cs.hasMetadataChanges() {
		t.Error("expected true after Named() changes the name")
	}
}

func TestCloudServer_HasMetadataChanges_TagsChanged(t *testing.T) {
	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.Tagged("new-tag")
	if !cs.hasMetadataChanges() {
		t.Error("expected true after Tagged() changes the tag set")
	}
}

// --------------------------------------------------------------------------
// cloudServerStringsToCommon helper
// --------------------------------------------------------------------------

func TestCloudServerStringsToCommon_NilInput(t *testing.T) {
	if got := cloudServerStringsToCommon(nil); got != nil {
		t.Errorf("expected nil for nil input, got %v", got)
	}
}

func TestCloudServerStringsToCommon_FiltersEmpty(t *testing.T) {
	got := cloudServerStringsToCommon([]string{"", "/valid/uri"})
	if len(got) != 1 || got[0].URI != "/valid/uri" {
		t.Errorf("expected one entry with /valid/uri, got %v", got)
	}
}

// --------------------------------------------------------------------------
// Smart Update dispatch
// --------------------------------------------------------------------------

func TestCloudServersClientAdapter_Update_MetadataOnly(t *testing.T) {
	var capturedMethod, capturedPath string
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, cloudServerSuccessBody)
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.Named("renamed-server")

	result, err := adapter.Update(context.Background(), cs)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if capturedMethod != http.MethodPut {
		t.Errorf("expected PUT, got %s", capturedMethod)
	}
	if !containsSubstring(capturedPath, "cloudServers") {
		t.Errorf("path %q should contain 'cloudServers'", capturedPath)
	}
	if result.actions == nil {
		t.Error("actions should be set after Update")
	}
}

func TestCloudServersClientAdapter_Update_SubnetsOnly(t *testing.T) {
	var capturedPaths []string
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPaths = append(capturedPaths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, cloudServerSuccessBody)
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.AssociateSubnets(URI("/subnets/s-new")).DisassociateSubnets(URI("/subnets/s-old"))

	if _, err := adapter.Update(context.Background(), cs); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if len(capturedPaths) != 1 {
		t.Fatalf("expected 1 API call, got %d: %v", len(capturedPaths), capturedPaths)
	}
	if !containsSubstring(capturedPaths[0], "associateDisassociateSubnets") {
		t.Errorf("path %q should contain 'associateDisassociateSubnets'", capturedPaths[0])
	}
	// Deltas cleared after success.
	if len(cs.subnetsToAssociate) != 0 || len(cs.subnetsToDisassociate) != 0 {
		t.Error("delta slices should be cleared after successful Update")
	}
}

func TestCloudServersClientAdapter_Update_MultipleDispatches(t *testing.T) {
	var capturedPaths []string
	callIdx := 0
	responses := []struct {
		code int
		body string
	}{
		{http.StatusOK, cloudServerSuccessBody},       // PUT
		{http.StatusAccepted, cloudServerSuccessBody}, // subnets
		{http.StatusAccepted, cloudServerSuccessBody}, // volumes
	}
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPaths = append(capturedPaths, r.URL.Path)
		resp := responses[callIdx]
		callIdx++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.code)
		fmt.Fprint(w, resp.body)
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.Named("renamed").
		AssociateSubnets(URI("/subnets/s-new")).
		AttachDataVolumes(URI("/vols/v-new"))

	if _, err := adapter.Update(context.Background(), cs); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if len(capturedPaths) != 3 {
		t.Fatalf("expected 3 API calls, got %d: %v", len(capturedPaths), capturedPaths)
	}
	if !containsSubstring(capturedPaths[0], "cloudServers/cs-1") {
		t.Errorf("first call path %q should target cloudServers/cs-1", capturedPaths[0])
	}
	if !containsSubstring(capturedPaths[1], "associateDisassociateSubnets") {
		t.Errorf("second call should be subnet association, got %s", capturedPaths[1])
	}
	if !containsSubstring(capturedPaths[2], "attachDetachDataVolumes") {
		t.Errorf("third call should be volume attachment, got %s", capturedPaths[2])
	}
}

func TestCloudServersClientAdapter_Update_NoOpWhenNothingChanged(t *testing.T) {
	callCount := 0
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	// name and tags unchanged, no deltas queued → Update should be a no-op

	if _, err := adapter.Update(context.Background(), cs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 0 {
		t.Errorf("expected 0 API calls for no-op Update, got %d", callCount)
	}
}

func TestCloudServersClientAdapter_Update_SubnetNonTwoXX(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Bad Request", "invalid subnet", 400))
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.AssociateSubnets(URI("/subnets/bad"))

	_, err := adapter.Update(context.Background(), cs)
	if err == nil {
		t.Fatal("expected error on 400")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected *HTTPError with 400, got %T: %v", err, err)
	}
}

func TestCloudServersClientAdapter_Update_SecurityGroupsDispatch(t *testing.T) {
	var capturedPath string
	var gotBody types.CloudServerAssociateSecurityGroupsRequest
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, cloudServerSuccessBody)
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.AssociateSecurityGroups(URI("/sgs/sg-new")).DisassociateSecurityGroups(URI("/sgs/sg-old"))

	if _, err := adapter.Update(context.Background(), cs); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if !containsSubstring(capturedPath, "associateDisassociateSecurityGroups") {
		t.Errorf("path %q should contain 'associateDisassociateSecurityGroups'", capturedPath)
	}
	if len(gotBody.SecurityGroupsToAssociate) != 1 || gotBody.SecurityGroupsToAssociate[0].URI != "/sgs/sg-new" {
		t.Errorf("SecurityGroupsToAssociate = %v", gotBody.SecurityGroupsToAssociate)
	}
}

func TestCloudServersClientAdapter_Update_ElasticIPsDispatch(t *testing.T) {
	var capturedPath string
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, cloudServerSuccessBody)
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.DisassociateElasticIPs(URI("/eips/eip-old"))

	if _, err := adapter.Update(context.Background(), cs); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if !containsSubstring(capturedPath, "associateDisassociateElasticIPs") {
		t.Errorf("path %q should contain 'associateDisassociateElasticIPs'", capturedPath)
	}
}

func TestCloudServersClientAdapter_Update_DataVolumesDispatch(t *testing.T) {
	var capturedPath string
	var gotBody types.CloudServerAttachDetachDataVolumesRequest
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, cloudServerSuccessBody)
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.AttachDataVolumes(URI("/vols/v-new")).DetachDataVolumes(URI("/vols/v-old"))

	if _, err := adapter.Update(context.Background(), cs); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if !containsSubstring(capturedPath, "attachDetachDataVolumes") {
		t.Errorf("path %q should contain 'attachDetachDataVolumes'", capturedPath)
	}
	if len(gotBody.VolumesToAttach) != 1 || gotBody.VolumesToAttach[0].URI != "/vols/v-new" {
		t.Errorf("VolumesToAttach = %v", gotBody.VolumesToAttach)
	}
	if len(gotBody.VolumesToDetach) != 1 || gotBody.VolumesToDetach[0].URI != "/vols/v-old" {
		t.Errorf("VolumesToDetach = %v", gotBody.VolumesToDetach)
	}
}

func TestCloudServer_SetPassword_MissingProjectID(t *testing.T) {
	callCount := 0
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	cs := &CloudServer{}
	id := "cs-1"
	uri := "/providers/Aruba.Compute/cloudServers/cs-1"
	cs.fromResponse(&types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, URI: &uri},
	})
	cs.actions = adapter

	err := cs.SetPassword(context.Background(), "pwd")
	if err == nil {
		t.Fatal("expected error when project ID is empty")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project ID")
	}
}

func TestCloudServer_PowerOn_NonTwoXX(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Conflict", "server already running", 409))
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "n", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.actions = adapter

	err := cs.PowerOn(context.Background())
	if err == nil {
		t.Fatal("expected error on 409")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusConflict {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Update — missing project path
// --------------------------------------------------------------------------

func TestCloudServersClientAdapter_Update_SetterError(t *testing.T) {
	callCount := 0
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	cs := NewCloudServer()
	cs.addErr(fmt.Errorf("setter error"))
	// Give it an ID and project so the only failure is the setter error.
	id := "cs-1"
	uri := "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"
	cs.fromResponse(&types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, URI: &uri},
	})

	_, err := adapter.Update(context.Background(), cs)
	if err == nil {
		t.Fatal("expected setter-time error")
	}
	if callCount != 0 {
		t.Error("HTTP should not be called when setter errors present")
	}
}

func TestCloudServersClientAdapter_Update_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	// Give it an ID but no project.
	cs := &CloudServer{}
	id := "cs-1"
	uri := "/providers/Aruba.Compute/cloudServers/cs-1"
	cs.fromResponse(&types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, URI: &uri},
	})

	_, err := adapter.Update(context.Background(), cs)
	if err == nil {
		t.Fatal("expected error when CloudServer has no parent project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without parent project")
	}
}

func TestCloudServersClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "invalid field", 422))
	})

	cs := &CloudServer{}
	cs.fromResponse(cloudServerTestResponse("cs-1", "my-server", "/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	cs.projectID = "p"
	cs.Named("trigger-put") // ensure PUT is dispatched so the 422 is surfaced

	result, err := adapter.Update(context.Background(), cs)
	if err == nil {
		t.Fatal("expected error on 422")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
	if result == nil {
		t.Fatal("result must be non-nil on non-2xx")
	}
}

// --------------------------------------------------------------------------
// Get — bad ref and non-2xx
// --------------------------------------------------------------------------

func TestCloudServersClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := adapter.Get(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad ref")
	}
}

func TestCloudServersClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "cloud server not found", 404))
	})

	ref := URI("/projects/p/providers/Aruba.Compute/cloudServers/cs-missing")
	result, err := adapter.Get(context.Background(), ref)
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
	if result == nil {
		t.Fatal("result must be non-nil on non-2xx")
	}
}

// --------------------------------------------------------------------------
// Delete — bad ref
// --------------------------------------------------------------------------

func TestCloudServersClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad ref")
	}
}

// --------------------------------------------------------------------------
// List — bad ref and non-2xx
// --------------------------------------------------------------------------

func TestCloudServersClientAdapter_List_BadRef(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := adapter.List(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad project ref")
	}
}

func TestCloudServersClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildCloudServersTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Forbidden", "access denied", 403))
	})

	_, err := adapter.List(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected error on 403")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
}

func TestCloudServer_FromResponse_SetsStatus(t *testing.T) {
	cs := &CloudServer{}
	state := types.State("Active")
	cs.fromResponse(&types.CloudServerResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if cs.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", cs.State())
	}
}

func TestCloudServersClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, cloudServerSuccessBody)
	})
	adapter := newCloudServersClientAdapter(testutil.NewClient(t, server.URL))
	cs, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Compute/cloudServers/cs-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&cs.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned CloudServer")
	}
}

// --------------------------------------------------------------------------
// WithBillingPeriod (#267)
// --------------------------------------------------------------------------

func TestCloudServer_WithBillingPeriod_SetsField(t *testing.T) {
	cs := NewCloudServer().BilledBy(BillingPeriodHour)
	if cs.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q, want %q", cs.BillingPeriod(), BillingPeriodHour)
	}
}

func TestCloudServer_WithBillingPeriod_InRequest(t *testing.T) {
	cs := NewCloudServer().BilledBy(BillingPeriodMonth)
	req := cs.RawRequest()
	if req.Properties.BillingPlanCommon == nil || req.Properties.BillingPlanCommon.BillingPeriod == nil || *req.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodMonth {
		t.Errorf("request BillingPlanCommon.BillingPeriod = %v, want %q", req.Properties.BillingPlanCommon, BillingPeriodMonth)
	}
}

func TestCloudServer_BillingPeriod_DefaultHour(t *testing.T) {
	cs := NewCloudServer()
	req := cs.RawRequest()
	if req.Properties.BillingPlanCommon == nil || req.Properties.BillingPlanCommon.BillingPeriod == nil || *req.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodHour {
		t.Errorf("default BillingPlanCommon.BillingPeriod = %v, want %q", req.Properties.BillingPlanCommon, BillingPeriodHour)
	}
}
