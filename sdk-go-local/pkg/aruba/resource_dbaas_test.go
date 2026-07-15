package aruba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/clients/database"
	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Compile-time Ref satisfaction
// --------------------------------------------------------------------------

var _ Ref = (*DBaaS)(nil)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestDBaaS_FluentSetters(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-1", "my-project", "/projects/p-1"))

	d := NewDBaaS().
		InProject(proj).
		Named("my-dbaas").
		Tagged("db").
		Tagged("prod").
		InRegion(RegionITBGBergamo).
		InZone(ZoneITBG1).
		OfEngine(DatabaseEngineMySQL80).
		OfFlavor(DBaaSFlavorDBO2A4).
		SizedGB(20).
		BilledBy(BillingPeriodHour).
		WithAutoscaling(50, 10)

	if d.Name() != "my-dbaas" {
		t.Errorf("Name() = %q", d.Name())
	}
	if tags := d.Tags(); len(tags) != 2 || tags[0] != "db" || tags[1] != "prod" {
		t.Errorf("Tags() = %v", tags)
	}
	if d.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", d.Region())
	}
	if d.Zone() != ZoneITBG1 {
		t.Errorf("Zone() = %q", d.Zone())
	}
	if d.Engine() != DatabaseEngineMySQL80 {
		t.Errorf("Engine() = %q", d.Engine())
	}
	if d.Flavor() != DBaaSFlavorDBO2A4 {
		t.Errorf("Flavor() = %q", d.Flavor())
	}
	if d.SizeGB() != 20 {
		t.Errorf("Storage() = %d", d.SizeGB())
	}
	if d.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", d.BillingPeriod())
	}
	if !d.AutoscalingEnabled() {
		t.Error("AutoscalingEnabled() should be true")
	}
	if d.AutoscalingAvailableSpaceGB() != 50 {
		t.Errorf("AutoscalingAvailableSpaceGB() = %d", d.AutoscalingAvailableSpaceGB())
	}
	if d.AutoscalingStepSizeGB() != 10 {
		t.Errorf("AutoscalingStepSizeGB() = %d", d.AutoscalingStepSizeGB())
	}
	if d.ProjectID() != "p-1" {
		t.Errorf("ProjectID() = %q", d.ProjectID())
	}
	if d.Err() != nil {
		t.Errorf("Err() = %v", d.Err())
	}

	d.Untagged("db")
	if tags := d.Tags(); len(tags) != 1 || tags[0] != "prod" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	d.RetaggedAs("x", "y")
	if tags := d.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

// --------------------------------------------------------------------------
// IntoProject — typed / URI / bad Ref
// --------------------------------------------------------------------------

func TestDBaaS_IntoProject_TypedRef(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p-42", "n", "/projects/p-42"))
	d := NewDBaaS().InProject(proj)
	if d.ProjectID() != "p-42" {
		t.Errorf("ProjectID() = %q", d.ProjectID())
	}
	if d.Err() != nil {
		t.Errorf("Err() = %v", d.Err())
	}
}

func TestDBaaS_IntoProject_URIRef(t *testing.T) {
	d := NewDBaaS().InProject(URI("/projects/p-uri"))
	if d.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() = %q", d.ProjectID())
	}
	if d.Err() != nil {
		t.Errorf("Err() = %v", d.Err())
	}
}

func TestDBaaS_IntoProject_BadRef(t *testing.T) {
	d := NewDBaaS().InProject(URI("/something/else"))
	if d.Err() == nil {
		t.Error("expected Err() to be set for unresolvable parent")
	}
}

// --------------------------------------------------------------------------
// Per-ref networking setters
// --------------------------------------------------------------------------

func TestDBaaS_WithVPC_Subnet_SecurityGroup_ElasticIP_URIRef(t *testing.T) {
	d := NewDBaaS().
		WithVPC(URI("/vpcs/v")).
		WithSubnet(URI("/subnets/s")).
		WithSecurityGroup(URI("/sgs/sg")).
		WithElasticIP(URI("/eips/e"))
	if d.VPC() != "/vpcs/v" {
		t.Errorf("VPC() = %q", d.VPC())
	}
	if d.Subnet() != "/subnets/s" {
		t.Errorf("Subnet() = %q", d.Subnet())
	}
	if d.SecurityGroup() != "/sgs/sg" {
		t.Errorf("SecurityGroup() = %q", d.SecurityGroup())
	}
	if d.ElasticIP() != "/eips/e" {
		t.Errorf("ElasticIP() = %q", d.ElasticIP())
	}
	if d.Err() != nil {
		t.Errorf("Err() = %v", d.Err())
	}
}

func TestDBaaS_WithVPC_Subnet_SecurityGroup_ElasticIP_TypedRef(t *testing.T) {
	vpc := URI("/vpcs/v1")
	subnet := URI("/subnets/s1")
	sg := URI("/sgs/sg1")
	eip := URI("/eips/e1")

	d := NewDBaaS().
		WithVPC(vpc).
		WithSubnet(subnet).
		WithSecurityGroup(sg).
		WithElasticIP(eip)
	if d.Err() != nil {
		t.Errorf("Err() = %v", d.Err())
	}
	if d.vpcRef == nil || *d.vpcRef != "/vpcs/v1" {
		t.Errorf("vpcRef = %v", d.vpcRef)
	}
}

func TestDBaaS_WithVPC_EmptyURI_AddsErr(t *testing.T) {
	d := NewDBaaS().WithVPC(URI(""))
	if d.Err() == nil {
		t.Error("expected Err() to be set for empty VPC URI")
	}
	if d.vpcRef != nil {
		t.Error("vpcRef should remain nil on error")
	}
}

// --------------------------------------------------------------------------
// Scalar setter individual tests
// --------------------------------------------------------------------------

func TestDBaaS_InZone(t *testing.T) {
	d := NewDBaaS().InZone(ZoneITBG1)
	if d.Zone() != ZoneITBG1 {
		t.Errorf("Zone() = %q", d.Zone())
	}
}

func TestDBaaS_OfEngine(t *testing.T) {
	d := NewDBaaS().OfEngine(DatabaseEngineMySQL80)
	if d.Engine() != DatabaseEngineMySQL80 {
		t.Errorf("Engine() = %q", d.Engine())
	}
}

func TestDBaaS_OfFlavor(t *testing.T) {
	d := NewDBaaS().OfFlavor(DBaaSFlavorDBO2A4)
	if d.Flavor() != DBaaSFlavorDBO2A4 {
		t.Errorf("Flavor() = %q", d.Flavor())
	}
}

func TestDBaaS_WithSizeGB(t *testing.T) {
	d := NewDBaaS().SizedGB(50)
	if d.SizeGB() != 50 {
		t.Errorf("Storage() = %d", d.SizeGB())
	}
}

func TestDBaaS_WithAutoscaling(t *testing.T) {
	d := NewDBaaS().WithAutoscaling(50, 10)
	if !d.AutoscalingEnabled() {
		t.Error("AutoscalingEnabled() should be true")
	}
	if d.AutoscalingAvailableSpaceGB() != 50 {
		t.Errorf("AutoscalingAvailableSpaceGB() = %d", d.AutoscalingAvailableSpaceGB())
	}
	if d.AutoscalingStepSizeGB() != 10 {
		t.Errorf("AutoscalingStepSizeGB() = %d", d.AutoscalingStepSizeGB())
	}
	req := d.RawRequest()
	if req.Properties.Autoscaling == nil {
		t.Fatal("Autoscaling block should be present in request")
	}
	if req.Properties.Autoscaling.Enabled == nil || !*req.Properties.Autoscaling.Enabled {
		t.Error("Autoscaling.Enabled should be true on wire")
	}
	if req.Properties.Autoscaling.AvailableSpace == nil || *req.Properties.Autoscaling.AvailableSpace != 50 {
		t.Errorf("Autoscaling.AvailableSpace = %v", req.Properties.Autoscaling.AvailableSpace)
	}
	if req.Properties.Autoscaling.StepSize == nil || *req.Properties.Autoscaling.StepSize != 10 {
		t.Errorf("Autoscaling.StepSize = %v", req.Properties.Autoscaling.StepSize)
	}
}

func TestDBaaS_WithoutAutoscaling(t *testing.T) {
	d := NewDBaaS().WithoutAutoscaling()
	if d.AutoscalingEnabled() {
		t.Error("AutoscalingEnabled() should be false after WithoutAutoscaling()")
	}
	req := d.RawRequest()
	if req.Properties.Autoscaling == nil {
		t.Fatal("Autoscaling block should be present in request (Enabled=false must be sent explicitly)")
	}
	if req.Properties.Autoscaling.Enabled == nil || *req.Properties.Autoscaling.Enabled {
		t.Error("Autoscaling.Enabled should be false on wire")
	}
	if req.Properties.Autoscaling.AvailableSpace != nil {
		t.Error("Autoscaling.AvailableSpace should be nil when disabled")
	}
	if req.Properties.Autoscaling.StepSize != nil {
		t.Error("Autoscaling.StepSize should be nil when disabled")
	}
}

func TestDBaaS_BilledMonthly(t *testing.T) {
	d := NewDBaaS().BilledBy(BillingPeriodMonth)
	if d.BillingPeriod() != BillingPeriodMonth {
		t.Errorf("BillingPeriod() = %q", d.BillingPeriod())
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestDBaaS_ToRequestRoundTrip(t *testing.T) {
	d := NewDBaaS().
		InProject(URI("/projects/p")).
		Named("my-dbaas").
		Tagged("tag1").
		InRegion(RegionITBGBergamo).
		InZone(ZoneITBG1).
		OfEngine(DatabaseEngineMySQL80).
		OfFlavor(DBaaSFlavorDBO2A4).
		SizedGB(20).
		BilledBy(BillingPeriodHour).
		WithAutoscaling(50, 10).
		WithVPC(URI("/vpcs/v")).
		WithSubnet(URI("/subnets/s")).
		WithSecurityGroup(URI("/sgs/sg")).
		WithElasticIP(URI("/eips/e"))

	req := d.RawRequest()

	if req.Metadata.Name != "my-dbaas" {
		t.Errorf("request name = %q", req.Metadata.Name)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("request location = %q", req.Metadata.Location.Value)
	}
	if req.Properties.Zone == nil || *req.Properties.Zone != ZoneITBG1 {
		t.Errorf("Zone = %v", req.Properties.Zone)
	}
	if req.Properties.Engine == nil || req.Properties.Engine.ID == nil || *req.Properties.Engine.ID != DatabaseEngineMySQL80 {
		t.Errorf("Engine.ID = %v", req.Properties.Engine)
	}
	if req.Properties.Flavor == nil || req.Properties.Flavor.Name == nil || *req.Properties.Flavor.Name != DBaaSFlavorDBO2A4 {
		t.Errorf("Flavor.Name = %v", req.Properties.Flavor)
	}
	if req.Properties.Storage == nil || req.Properties.Storage.SizeGB == nil || *req.Properties.Storage.SizeGB != 20 {
		t.Errorf("Storage.SizeGB = %v", req.Properties.Storage)
	}
	if req.Properties.BillingPlanCommon == nil || req.Properties.BillingPlanCommon.BillingPeriod == nil || *req.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodHour {
		t.Errorf("BillingPlanCommon.BillingPeriod = %v", req.Properties.BillingPlanCommon)
	}
	if req.Properties.Autoscaling == nil || req.Properties.Autoscaling.Enabled == nil || !*req.Properties.Autoscaling.Enabled {
		t.Errorf("Autoscaling.Enabled = %v", req.Properties.Autoscaling)
	}
	if req.Properties.Autoscaling == nil || req.Properties.Autoscaling.AvailableSpace == nil || *req.Properties.Autoscaling.AvailableSpace != 50 {
		t.Errorf("Autoscaling.AvailableSpace = %v", req.Properties.Autoscaling)
	}
	if req.Properties.Autoscaling == nil || req.Properties.Autoscaling.StepSize == nil || *req.Properties.Autoscaling.StepSize != 10 {
		t.Errorf("Autoscaling.StepSize = %v", req.Properties.Autoscaling)
	}
	if req.Properties.Networking == nil {
		t.Fatal("Networking should not be nil")
	}
	if req.Properties.Networking.VPCURI == nil || *req.Properties.Networking.VPCURI != "/vpcs/v" {
		t.Errorf("Networking.VPCURI = %v", req.Properties.Networking.VPCURI)
	}
	if req.Properties.Networking.SubnetURI == nil || *req.Properties.Networking.SubnetURI != "/subnets/s" {
		t.Errorf("Networking.SubnetURI = %v", req.Properties.Networking.SubnetURI)
	}
	if req.Properties.Networking.SecurityGroupURI == nil || *req.Properties.Networking.SecurityGroupURI != "/sgs/sg" {
		t.Errorf("Networking.SecurityGroupURI = %v", req.Properties.Networking.SecurityGroupURI)
	}
	if req.Properties.Networking.ElasticIPURI == nil || *req.Properties.Networking.ElasticIPURI != "/eips/e" {
		t.Errorf("Networking.ElasticIPURI = %v", req.Properties.Networking.ElasticIPURI)
	}
}

func TestDBaaS_ToRequest_AllUnset(t *testing.T) {
	d := &DBaaS{}
	req := d.toRequest() // must not panic
	if req.Properties.Zone != nil {
		t.Error("Zone should be nil when unset")
	}
	if req.Properties.Engine != nil {
		t.Error("Engine should be nil when unset")
	}
	if req.Properties.Networking != nil {
		t.Error("Networking should be nil when no refs are set")
	}
	if req.Properties.Autoscaling != nil {
		t.Error("Autoscaling should be nil when neither WithAutoscaling nor WithoutAutoscaling was called")
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func dbaasTestResponse(id, name, uri string) *types.DBaaSResponse {
	loc := &types.LocationResponse{Value: RegionITBGBergamo}
	engineType := string(DatabaseEngineMySQL80)
	flavorName := string(DBaaSFlavorDBO2A4)
	sizeGB := int32(20)
	billingPeriod := BillingPeriodHour
	state := types.State("Active")
	vpcURI := "/vpcs/v"
	subnetURI := "/subnets/s"
	sgURI := "/sgs/sg"
	eipURI := "/eips/e"
	return &types.DBaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			URI:              &uri,
			Name:             &name,
			Tags:             []string{"tag1"},
			LocationResponse: loc,
		},
		Properties: types.DBaaSPropertiesResponse{
			Engine:            &types.DBaaSEngineResponse{Type: &engineType},
			Flavor:            &types.DBaaSFlavorResponse{Name: &flavorName},
			Storage:           &types.DBaaSStorageResponse{SizeGB: &sizeGB},
			BillingPlanCommon: &types.BillingPlanCommon{BillingPeriod: &billingPeriod},
			Networking: &types.DBaaSNetworkingResponse{
				VPC:           &types.ReferenceResourceCommon{URI: vpcURI},
				Subnet:        &types.ReferenceResourceCommon{URI: subnetURI},
				SecurityGroup: &types.ReferenceResourceCommon{URI: sgURI},
				ElasticIP:     &types.ReferenceResourceCommon{URI: eipURI},
			},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}
}

func TestDBaaS_FromResponseHydration(t *testing.T) {
	d := &DBaaS{}
	resp := dbaasTestResponse("db-1", "my-dbaas", "/projects/p/providers/Aruba.Database/dbaas/db-1")
	d.fromResponse(resp)

	if d.ID() != "db-1" {
		t.Errorf("ID() = %q", d.ID())
	}
	if d.URI() != "/projects/p/providers/Aruba.Database/dbaas/db-1" {
		t.Errorf("URI() = %q", d.URI())
	}
	if d.DBaaSID() != "db-1" {
		t.Errorf("DBaaSID() = %q", d.DBaaSID())
	}
	if d.Name() != "my-dbaas" {
		t.Errorf("Name() = %q", d.Name())
	}
	if tags := d.Tags(); len(tags) != 1 || tags[0] != "tag1" {
		t.Errorf("Tags() = %v", tags)
	}
	if d.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", d.Region())
	}
	if d.Engine() != DatabaseEngineMySQL80 {
		t.Errorf("Engine() = %q", d.Engine())
	}
	if d.EngineRaw() == nil || d.EngineRaw().Type == nil {
		t.Error("EngineRaw() should carry full struct")
	}
	if d.Flavor() != DBaaSFlavorDBO2A4 {
		t.Errorf("Flavor() = %q", d.Flavor())
	}
	if d.FlavorRaw() == nil || d.FlavorRaw().Name == nil {
		t.Error("FlavorRaw() should carry full struct")
	}
	if d.SizeGB() != 20 {
		t.Errorf("Storage() = %d", d.SizeGB())
	}
	if d.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", d.BillingPeriod())
	}
	if d.VPC() != "/vpcs/v" {
		t.Errorf("VPC() = %q", d.VPC())
	}
	if d.Subnet() != "/subnets/s" {
		t.Errorf("Subnet() = %q", d.Subnet())
	}
	if d.SecurityGroup() != "/sgs/sg" {
		t.Errorf("SecurityGroup() = %q", d.SecurityGroup())
	}
	if d.ElasticIP() != "/eips/e" {
		t.Errorf("ElasticIP() = %q", d.ElasticIP())
	}
	if d.State() != "Active" {
		t.Errorf("State() = %q", d.State())
	}
	if d.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
	if d.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", d.ProjectID())
	}
}

func TestDBaaS_FromResponseURIBackfill(t *testing.T) {
	id := "db-99"
	uri := "/projects/p-uri/providers/Aruba.Database/dbaas/db-99"
	resp := &types.DBaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:  &id,
			URI: &uri,
		},
	}
	d := &DBaaS{}
	d.fromResponse(resp)
	if d.ProjectID() != "p-uri" {
		t.Errorf("ProjectID() via URI backfill = %q", d.ProjectID())
	}
}

func TestDBaaS_FromResponse_NilSafe(t *testing.T) {
	d := &DBaaS{}
	d.fromResponse(nil)
	if d.ID() != "" || d.URI() != "" || d.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}

	d2 := &DBaaS{}
	d2.fromResponse(&types.DBaaSResponse{})
	if d2.ID() != "" || d2.URI() != "" || d2.Engine() != "" || d2.Flavor() != "" {
		t.Error("empty response should yield zero accessor values")
	}
}

// --------------------------------------------------------------------------
// Engine / Flavor asymmetry (request vs response)
// --------------------------------------------------------------------------

func TestDBaaS_Engine_Asymmetry(t *testing.T) {
	engineType := string(DatabaseEngineMySQL80)
	d := NewDBaaS().OfEngine(DatabaseEngineMySQL80)
	if d.Engine() != DatabaseEngineMySQL80 {
		t.Errorf("pre-response Engine() = %q", d.Engine())
	}
	if d.EngineRaw() != nil {
		t.Error("EngineRaw() should be nil before hydration")
	}

	d.fromResponse(&types.DBaaSResponse{
		Properties: types.DBaaSPropertiesResponse{
			Engine: &types.DBaaSEngineResponse{Type: &engineType},
		},
	})
	if d.Engine() != DatabaseEngineMySQL80 {
		t.Errorf("post-response Engine() = %q", d.Engine())
	}
	if d.EngineRaw() == nil {
		t.Error("EngineRaw() should be populated after hydration")
	}
}

func TestDBaaS_Flavor_Asymmetry(t *testing.T) {
	flavorName := string(DBaaSFlavorDBO4A8)
	d := NewDBaaS().OfFlavor(DBaaSFlavorDBO2A4)
	if d.Flavor() != DBaaSFlavorDBO2A4 {
		t.Errorf("pre-response Flavor() = %q", d.Flavor())
	}
	if d.FlavorRaw() != nil {
		t.Error("FlavorRaw() should be nil before hydration")
	}

	d.fromResponse(&types.DBaaSResponse{
		Properties: types.DBaaSPropertiesResponse{
			Flavor: &types.DBaaSFlavorResponse{Name: &flavorName},
		},
	})
	if d.Flavor() != DBaaSFlavorDBO4A8 {
		t.Errorf("post-response Flavor() = %q (expected response value to win)", d.Flavor())
	}
}

// --------------------------------------------------------------------------
// Round-trip Update flow
// --------------------------------------------------------------------------

func TestDBaaS_RoundTripUpdateFlow(t *testing.T) {
	resp := dbaasTestResponse("db-1", "my-dbaas", "/projects/p/providers/Aruba.Database/dbaas/db-1")
	d := &DBaaS{}
	d.fromResponse(resp)

	// Mutate storage; networking URIs should still come from the hydrated response.
	d.SizedGB(40)

	req := d.RawRequest()
	if req.Properties.Storage == nil || req.Properties.Storage.SizeGB == nil || *req.Properties.Storage.SizeGB != 40 {
		t.Errorf("Storage.SizeGB after mutation = %v", req.Properties.Storage)
	}
	if req.Properties.Networking == nil || req.Properties.Networking.VPCURI == nil || *req.Properties.Networking.VPCURI != "/vpcs/v" {
		t.Errorf("Networking.VPCURI from hydrated response = %v", req.Properties.Networking)
	}
}

// --------------------------------------------------------------------------
// Ref + ancestor ID satisfaction (runtime)
// --------------------------------------------------------------------------

func TestDBaaS_RefSatisfaction(t *testing.T) {
	d := &DBaaS{}
	d.fromResponse(dbaasTestResponse("db-99", "n", "/projects/p99/providers/Aruba.Database/dbaas/db-99"))

	// withDBaaSID typed path
	did, ok := extractID(d, func(ref Ref) (string, bool) {
		if w, ok := ref.(withDBaaSID); ok {
			return w.DBaaSID(), true
		}
		return "", false
	}, "dbaas")
	if !ok || did != "db-99" {
		t.Errorf("extractID via withDBaaSID = (%q, %v)", did, ok)
	}

	// withProjectID typed path
	pid, ok := extractID(d, func(ref Ref) (string, bool) {
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
// dbaasIDsFromRef helper
// --------------------------------------------------------------------------

func TestDBaaSIDsFromRef_TypedRef(t *testing.T) {
	d := &DBaaS{}
	d.fromResponse(dbaasTestResponse("db-1", "n", "/projects/p/providers/Aruba.Database/dbaas/db-1"))
	pid, did, err := dbaasIDsFromRef(d)
	if err != nil || pid != "p" || did != "db-1" {
		t.Errorf("dbaasIDsFromRef typed = (%q, %q, %v)", pid, did, err)
	}
}

func TestDBaaSIDsFromRef_URIRef(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Database/dbaas/db-1")
	pid, did, err := dbaasIDsFromRef(ref)
	if err != nil || pid != "p" || did != "db-1" {
		t.Errorf("dbaasIDsFromRef URI = (%q, %q, %v)", pid, did, err)
	}
}

func TestDBaaSIDsFromRef_BadURI_MissingDBaaS(t *testing.T) {
	_, _, err := dbaasIDsFromRef(URI("/projects/p/providers/Aruba.Database"))
	if err == nil {
		t.Error("expected error for URI without /dbaas/<id>")
	}
}

func TestDBaaSIDsFromRef_BadURI_MissingProject(t *testing.T) {
	_, _, err := dbaasIDsFromRef(URI("/providers/Aruba.Database/dbaas/db-1"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

func TestDBaaSIDsFromRef_BadURI_MissingAll(t *testing.T) {
	_, _, err := dbaasIDsFromRef(URI("/something/else"))
	if err == nil {
		t.Error("expected error for completely unrelated URI")
	}
}

// --------------------------------------------------------------------------
// dbaasClientAdapter — HTTP mock tests
// --------------------------------------------------------------------------

func buildDBaaSTestAdapter(t *testing.T, handler http.HandlerFunc) *dbaasClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newDBaaSClientAdapter(testutil.NewClient(t, server.URL))
}

const dbaasSuccessBody = `{` +
	`"metadata":{"id":"db-1","name":"my-dbaas","uri":"/projects/p/providers/Aruba.Database/dbaas/db-1"},` +
	`"properties":{` +
	`"engine":{"type":"mysql-8.0"},` +
	`"flavor":{"name":"DBO2A4"},` +
	`"storage":{"sizeGb":20},` +
	`"billingPlan":{"billingPeriod":"Hour"},` +
	`"networking":{"vpc":{"uri":"/vpcs/v"},"subnet":{"uri":"/subnets/s"},"securityGroup":{"uri":"/sgs/sg"},"elasticIp":{"uri":"/eips/e"}}` +
	`},` +
	`"status":{"state":"Active"}}`

func TestDBaaSClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.DBaaSRequest
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if !containsSubstring(r.URL.Path, "dbaas") {
			t.Errorf("path %q should contain 'dbaas'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, dbaasSuccessBody)
	})

	d := NewDBaaS().
		InProject(URI("/projects/p")).
		Named("my-dbaas").
		InZone(ZoneITBG1).
		OfEngine(DatabaseEngineMySQL80).
		OfFlavor(DBaaSFlavorDBO2A4).
		SizedGB(20).
		BilledBy(BillingPeriodHour).
		WithAutoscaling(50, 10).
		WithVPC(URI("/vpcs/v")).
		WithSubnet(URI("/subnets/s")).
		WithSecurityGroup(URI("/sgs/sg")).
		WithElasticIP(URI("/eips/e"))

	result, err := adapter.Create(context.Background(), d)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "db-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-dbaas" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.Engine() != DatabaseEngineMySQL80 {
		t.Errorf("Engine() = %q", result.Engine())
	}
	if result.State() != "Active" {
		t.Errorf("State() = %q", result.State())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	// Verify wire body.
	if gotBody.Properties.Engine == nil || gotBody.Properties.Engine.ID == nil || *gotBody.Properties.Engine.ID != DatabaseEngineMySQL80 {
		t.Errorf("request Engine.ID = %v", gotBody.Properties.Engine)
	}
	if gotBody.Properties.Networking == nil || gotBody.Properties.Networking.VPCURI == nil || *gotBody.Properties.Networking.VPCURI != "/vpcs/v" {
		t.Errorf("request Networking.VPCURI = %v", gotBody.Properties.Networking)
	}
}

func TestDBaaSClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	_, err := adapter.Create(context.Background(), NewDBaaS().
		Named("x").OfEngine(DatabaseEngineMySQL80))
	if err == nil {
		t.Fatal("expected error when DBaaS has no parent project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without parent project")
	}
}

func TestDBaaSClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" field — triggers MetadataValidationError from the low-level client.
		fmt.Fprint(w, `{"metadata":{"name":"db","uri":"/projects/p/providers/Aruba.Database/dbaas/x"},"properties":{}}`)
	})

	d := NewDBaaS().InProject(URI("/projects/p")).
		Named("db").OfEngine(DatabaseEngineMySQL80)
	result, err := adapter.Create(context.Background(), d)
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

func TestDBaaSClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "engine is required", 422))
	})

	d := NewDBaaS().InProject(URI("/projects/p")).
		Named("db")
	result, err := adapter.Create(context.Background(), d)
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

func TestDBaaSClientAdapter_Update_Success(t *testing.T) {
	var capturedMethod string
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		if !containsSubstring(r.URL.Path, "dbaas") {
			t.Errorf("path %q should contain 'dbaas'", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, dbaasSuccessBody)
	})

	d := &DBaaS{}
	d.fromResponse(dbaasTestResponse("db-1", "my-dbaas", "/projects/p/providers/Aruba.Database/dbaas/db-1"))
	d.projectID = "p"

	result, err := adapter.Update(context.Background(), d)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if capturedMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", capturedMethod)
	}
	if result.ID() != "db-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestDBaaSClientAdapter_Update_NoID(t *testing.T) {
	callCount := 0
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	d := NewDBaaS().InProject(URI("/projects/p")).
		Named("x")
	_, err := adapter.Update(context.Background(), d)
	if err == nil {
		t.Fatal("expected error when DBaaS has no ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without ID")
	}
}

func TestDBaaSClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, dbaasSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Database/dbaas/db-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "db-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	if !containsSubstring(capturedPath, "dbaas") {
		t.Errorf("path %q should contain 'dbaas'", capturedPath)
	}
}

func TestDBaaSClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, dbaasSuccessBody)
	})

	existing := &DBaaS{}
	existing.fromResponse(dbaasTestResponse("db-1", "n", "/projects/p/providers/Aruba.Database/dbaas/db-1"))

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "db-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
}

func TestDBaaSClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Database/dbaas/db-1"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestDBaaSClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "dbaas not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Database/dbaas/missing"))
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
// InRegion (0% → covers that branch)
// --------------------------------------------------------------------------

func TestDBaaS_InRegion(t *testing.T) {
	d := NewDBaaS().InRegion(RegionITBGBergamo)
	if d.Region() != RegionITBGBergamo {
		t.Errorf("Region() after InRegion = %q", d.Region())
	}
}

// --------------------------------------------------------------------------
// Zero-value accessors (Shape F — covers the nil-response branch)
// --------------------------------------------------------------------------

func TestDBaaS_Accessors_ZeroValue(t *testing.T) {
	d := &DBaaS{}
	if d.SizeGB() != 0 {
		t.Errorf("Storage() zero = %d", d.SizeGB())
	}
	if d.AutoscalingEnabled() {
		t.Error("AutoscalingEnabled() zero should be false")
	}
	if d.AutoscalingStatus() != "" {
		t.Errorf("AutoscalingStatus() zero = %q", d.AutoscalingStatus())
	}
	if d.AutoscalingAvailableSpaceGB() != 0 {
		t.Errorf("AutoscalingAvailableSpaceGB() zero = %d", d.AutoscalingAvailableSpaceGB())
	}
	if d.AutoscalingStepSizeGB() != 0 {
		t.Errorf("AutoscalingStepSizeGB() zero = %d", d.AutoscalingStepSizeGB())
	}
	if d.AutoscalingRuleID() != "" {
		t.Errorf("AutoscalingRuleID() zero = %q", d.AutoscalingRuleID())
	}
	if d.AutoscalingRaw() != nil {
		t.Error("AutoscalingRaw() zero should be nil")
	}
}

func TestDBaaS_AutoscalingResponse_Accessors(t *testing.T) {
	status := "Enabled"
	avail := int32(50)
	step := int32(10)
	ruleID := "rule-1"
	d := &DBaaS{}
	d.fromResponse(&types.DBaaSResponse{
		Properties: types.DBaaSPropertiesResponse{
			Autoscaling: &types.DBaaSAutoscalingResponse{
				Status:         &status,
				AvailableSpace: &avail,
				StepSize:       &step,
				RuleID:         &ruleID,
			},
		},
	})

	if d.AutoscalingStatus() != "Enabled" {
		t.Errorf("AutoscalingStatus() = %q", d.AutoscalingStatus())
	}
	if d.AutoscalingAvailableSpaceGB() != 50 {
		t.Errorf("AutoscalingAvailableSpaceGB() = %d", d.AutoscalingAvailableSpaceGB())
	}
	if d.AutoscalingStepSizeGB() != 10 {
		t.Errorf("AutoscalingStepSizeGB() = %d", d.AutoscalingStepSizeGB())
	}
	if d.AutoscalingRuleID() != "rule-1" {
		t.Errorf("AutoscalingRuleID() = %q", d.AutoscalingRuleID())
	}
	if d.AutoscalingRaw() == nil {
		t.Error("AutoscalingRaw() should be non-nil after hydration")
	}
	if d.AutoscalingRaw().Status == nil || *d.AutoscalingRaw().Status != "Enabled" {
		t.Error("AutoscalingRaw().Status should match")
	}
}

// --------------------------------------------------------------------------
// Update — extra guard paths
// --------------------------------------------------------------------------

func TestDBaaSClientAdapter_Update_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	d := &DBaaS{}
	// Set an ID but no project
	resp := dbaasTestResponse("db-x", "n", "/projects/p/providers/Aruba.Database/dbaas/db-x")
	d.fromResponse(resp)
	d.projectID = "" // explicitly clear
	_, err := adapter.Update(context.Background(), d)
	if err == nil {
		t.Fatal("expected error when DBaaS has no parent project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without parent project")
	}
}

func TestDBaaSClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Conflict", "concurrent update", 409))
	})

	d := &DBaaS{}
	d.fromResponse(dbaasTestResponse("db-1", "n", "/projects/p/providers/Aruba.Database/dbaas/db-1"))
	d.projectID = "p"
	_, err := adapter.Update(context.Background(), d)
	if err == nil {
		t.Fatal("expected error on non-2xx")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusConflict {
		t.Errorf("HTTPError.StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Get — bad ref path
// --------------------------------------------------------------------------

func TestDBaaSClientAdapter_Get_BadRef(t *testing.T) {
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	_, err := adapter.Get(context.Background(), URI("/something/not-a-dbaas"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
}

// --------------------------------------------------------------------------
// Get — non-2xx response
// --------------------------------------------------------------------------

func TestDBaaSClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "dbaas not found", 404))
	})

	ref := URI("/projects/p/providers/Aruba.Database/dbaas/db-1")
	_, err := adapter.Get(context.Background(), ref)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("HTTPError.StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// List — non-2xx response
// --------------------------------------------------------------------------

func TestDBaaSClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Forbidden", "not allowed", 403))
	})

	_, err := adapter.List(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected error on 403")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusForbidden {
		t.Errorf("HTTPError.StatusCode = %d", httpErr.StatusCode)
	}
}

// --------------------------------------------------------------------------
// Update — errMixin path
// --------------------------------------------------------------------------

func TestDBaaSClientAdapter_Update_ErrMixin(t *testing.T) {
	callCount := 0
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	d := NewDBaaS().InProject(URI("/garbage/project")) // triggers errMixin on intoProject
	_, err := adapter.Update(context.Background(), d)
	if err == nil {
		t.Fatal("expected error from errMixin")
	}
	if callCount != 0 {
		t.Errorf("expected 0 HTTP calls, got %d", callCount)
	}
}

// --------------------------------------------------------------------------
// Delete — bad ref and broken client
// --------------------------------------------------------------------------

func TestDBaaSClientAdapter_Delete_BadRef(t *testing.T) {
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := adapter.Delete(context.Background(), URI("/something/bad"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
}

func TestDBaaSClientAdapter_Delete_BrokenClient(t *testing.T) {
	adapter := &dbaasClientAdapter{low: database.NewDBaaSClientImpl(testutil.NewBrokenClient(t, "http://localhost:9"))}
	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Database/dbaas/db-1"))
	if err == nil {
		t.Fatal("expected network error from broken client")
	}
}

// --------------------------------------------------------------------------
// List — bad parent ref and broken client
// --------------------------------------------------------------------------

func TestDBaaSClientAdapter_List_BadRef(t *testing.T) {
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/something/bad"))
	if err == nil {
		t.Fatal("expected error for bad parent Ref")
	}
}

func TestDBaaSClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildDBaaSTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"db-1","name":"dbaas1","uri":"/projects/p/providers/Aruba.Database/dbaas/db-1"},"properties":{"engine":{"type":"mysql-8.0"},"flavor":{"name":"DBO2A4"}}},`+
			`{"metadata":{"id":"db-2","name":"dbaas2","uri":"/projects/p/providers/Aruba.Database/dbaas/db-2"},"properties":{"engine":{"type":"mssql-2022-web"},"flavor":{"name":"DBO4A8"}}}`+
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
	if items[0].ID() != "db-1" || items[0].Name() != "dbaas1" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[0].Engine() != DatabaseEngineMySQL80 {
		t.Errorf("items[0].Engine() = %q", items[0].Engine())
	}
	if items[1].ID() != "db-2" || items[1].Engine() != DatabaseEngineMSSQL2022Web {
		t.Errorf("items[1] ID=%q Engine=%q", items[1].ID(), items[1].Engine())
	}
	if items[0].ProjectID() != "p" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
}

func TestDBaaS_FromResponse_SetsStatus(t *testing.T) {
	d := &DBaaS{}
	state := types.State("Active")
	d.fromResponse(&types.DBaaSResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if d.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", d.State())
	}
}

func TestDBaaSClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, dbaasSuccessBody)
	})
	adapter := newDBaaSClientAdapter(testutil.NewClient(t, server.URL))
	db, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Database/dbaas/db-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&db.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned DBaaS")
	}
}

func TestDBaaS_FromResponse_BackPopulatesAutoscaling(t *testing.T) {
	avail := int32(20)
	step := int32(10)
	status := "Enabled"

	d := &DBaaS{}
	d.fromResponse(&types.DBaaSResponse{
		Properties: types.DBaaSPropertiesResponse{
			Autoscaling: &types.DBaaSAutoscalingResponse{
				Status:         &status,
				AvailableSpace: &avail,
				StepSize:       &step,
			},
		},
	})

	if !d.AutoscalingEnabled() {
		t.Error("AutoscalingEnabled() should be true when Status=Enabled")
	}
	req := d.toRequest()
	if req.Properties.Autoscaling == nil {
		t.Fatal("Update request must include autoscaling block after fromResponse")
	}
	if req.Properties.Autoscaling.AvailableSpace == nil || *req.Properties.Autoscaling.AvailableSpace != 20 {
		t.Errorf("AvailableSpace = %v, want 20", req.Properties.Autoscaling.AvailableSpace)
	}
	if req.Properties.Autoscaling.StepSize == nil || *req.Properties.Autoscaling.StepSize != 10 {
		t.Errorf("StepSize = %v, want 10", req.Properties.Autoscaling.StepSize)
	}
}

func TestDBaaS_FromResponse_AutoscalingDisabledStatus(t *testing.T) {
	status := "Disabled"
	d := &DBaaS{}
	d.fromResponse(&types.DBaaSResponse{
		Properties: types.DBaaSPropertiesResponse{
			Autoscaling: &types.DBaaSAutoscalingResponse{Status: &status},
		},
	})
	if d.AutoscalingEnabled() {
		t.Error("AutoscalingEnabled() should be false when Status=Disabled")
	}
}

// --------------------------------------------------------------------------
// EngineVersion / EngineType / PrivateIPAddress / FlavorCPU / FlavorRAMMB getters
// --------------------------------------------------------------------------

func TestDBaaS_EngineVersion_NilResponse(t *testing.T) {
	d := &DBaaS{}
	if got := d.EngineVersion(); got != "" {
		t.Errorf("EngineVersion() on nil response = %q, want \"\"", got)
	}
}

func TestDBaaS_EngineVersion_Populated(t *testing.T) {
	ver := "8.0"
	d := &DBaaS{}
	d.fromResponse(&types.DBaaSResponse{
		Properties: types.DBaaSPropertiesResponse{
			Engine: &types.DBaaSEngineResponse{Version: &ver},
		},
	})
	if got := d.EngineVersion(); got != "8.0" {
		t.Errorf("EngineVersion() = %q, want %q", got, "8.0")
	}
}

func TestDBaaS_EngineType_FromResponse(t *testing.T) {
	eng := "mysql-8.0"
	d := &DBaaS{}
	d.fromResponse(&types.DBaaSResponse{
		Properties: types.DBaaSPropertiesResponse{
			Engine: &types.DBaaSEngineResponse{Type: &eng},
		},
	})
	if got := d.EngineType(); got != "mysql-8.0" {
		t.Errorf("EngineType() = %q, want %q", got, "mysql-8.0")
	}
}

func TestDBaaS_EngineType_FromLocalField(t *testing.T) {
	d := NewDBaaS().OfEngine(types.DatabaseEngineMySQL80)
	if got := d.EngineType(); got != "mysql-8.0" {
		t.Errorf("EngineType() from local field = %q, want %q", got, "mysql-8.0")
	}
}

func TestDBaaS_PrivateIPAddress_NilResponse(t *testing.T) {
	d := &DBaaS{}
	if got := d.PrivateIPAddress(); got != "" {
		t.Errorf("PrivateIPAddress() on nil response = %q, want \"\"", got)
	}
}

func TestDBaaS_PrivateIPAddress_Populated(t *testing.T) {
	ip := "10.0.0.5"
	d := &DBaaS{}
	d.fromResponse(&types.DBaaSResponse{
		Properties: types.DBaaSPropertiesResponse{
			Engine: &types.DBaaSEngineResponse{PrivateIPAddress: &ip},
		},
	})
	if got := d.PrivateIPAddress(); got != ip {
		t.Errorf("PrivateIPAddress() = %q, want %q", got, ip)
	}
}

func TestDBaaS_FlavorCPU_NilResponse(t *testing.T) {
	d := &DBaaS{}
	if got := d.FlavorCPU(); got != 0 {
		t.Errorf("FlavorCPU() on nil response = %d, want 0", got)
	}
}

func TestDBaaS_FlavorCPU_Populated(t *testing.T) {
	cpu := int32(4)
	d := &DBaaS{}
	d.fromResponse(&types.DBaaSResponse{
		Properties: types.DBaaSPropertiesResponse{
			Flavor: &types.DBaaSFlavorResponse{CPU: &cpu},
		},
	})
	if got := d.FlavorCPU(); got != 4 {
		t.Errorf("FlavorCPU() = %d, want 4", got)
	}
}

func TestDBaaS_FlavorRAMMB_NilResponse(t *testing.T) {
	d := &DBaaS{}
	if got := d.FlavorRAMMB(); got != 0 {
		t.Errorf("FlavorRAMMB() on nil response = %d, want 0", got)
	}
}

func TestDBaaS_FlavorRAMMB_Populated(t *testing.T) {
	ram := int32(8192)
	d := &DBaaS{}
	d.fromResponse(&types.DBaaSResponse{
		Properties: types.DBaaSPropertiesResponse{
			Flavor: &types.DBaaSFlavorResponse{RAM: &ram},
		},
	})
	if got := d.FlavorRAMMB(); got != 8192 {
		t.Errorf("FlavorRAMMB() = %d, want 8192", got)
	}
}
