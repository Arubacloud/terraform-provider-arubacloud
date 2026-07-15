package aruba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/testutil"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// Compile-time Ref satisfaction
// --------------------------------------------------------------------------

var _ Ref = (*VPNTunnel)(nil)

// --------------------------------------------------------------------------
// Fluent setters
// --------------------------------------------------------------------------

func TestVPNTunnel_FluentSetters(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p1", "my-project", "/projects/p1"))

	tun := NewVPNTunnel().
		InProject(proj).
		Named("my-tunnel").
		Tagged("vpn").
		Tagged("ipsec").
		Tagged("vpn"). // dedupe
		InRegion(RegionITBGBergamo).
		OfType(VPNTypeSiteToSite).
		WithVPNClientProtocol(VPNClientProtocolIKEv2).
		BilledBy(BillingPeriodHour).
		WithPeerClientPublicIP("1.2.3.4")

	if tun.Name() != "my-tunnel" {
		t.Errorf("Name() = %q", tun.Name())
	}
	if tags := tun.Tags(); len(tags) != 2 || tags[0] != "vpn" || tags[1] != "ipsec" {
		t.Errorf("Tags() = %v", tags)
	}
	if tun.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", tun.Region())
	}
	if tun.VPNType() != VPNTypeSiteToSite {
		t.Errorf("VPNType() = %q", tun.VPNType())
	}
	if tun.VPNClientProtocol() != VPNClientProtocolIKEv2 {
		t.Errorf("VPNClientProtocol() = %q", tun.VPNClientProtocol())
	}
	if tun.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", tun.BillingPeriod())
	}
	if tun.PeerClientPublicIP() != "1.2.3.4" {
		t.Errorf("PeerClientPublicIP() = %q", tun.PeerClientPublicIP())
	}
	if tun.ProjectID() != "p1" {
		t.Errorf("ProjectID() = %q", tun.ProjectID())
	}
	if tun.Err() != nil {
		t.Errorf("Err() = %v", tun.Err())
	}

	tun.Untagged("vpn")
	if tags := tun.Tags(); len(tags) != 1 || tags[0] != "ipsec" {
		t.Errorf("after RemoveTag Tags() = %v", tags)
	}

	tun.RetaggedAs("x", "y")
	if tags := tun.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("after ReplaceTags Tags() = %v", tags)
	}
}

// --------------------------------------------------------------------------
// IntoProject — typed Ref
// --------------------------------------------------------------------------

func TestVPNTunnel_IntoProject_TypedRef(t *testing.T) {
	proj := &Project{}
	proj.fromResponse(projectTestResponse("p1", "my-project", "/projects/p1"))

	tun := NewVPNTunnel().InProject(proj)

	if tun.ProjectID() != "p1" {
		t.Errorf("ProjectID() = %q", tun.ProjectID())
	}
	if tun.Err() != nil {
		t.Errorf("Err() = %v", tun.Err())
	}
}

func TestVPNTunnel_IntoProject_URIRef(t *testing.T) {
	tun := NewVPNTunnel().InProject(URI("/projects/p"))

	if tun.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", tun.ProjectID())
	}
	if tun.Err() != nil {
		t.Errorf("Err() = %v", tun.Err())
	}
}

func TestVPNTunnel_IntoProject_BadRef(t *testing.T) {
	tun := NewVPNTunnel().InProject(URI("/garbage"))
	if tun.Err() == nil {
		t.Error("expected Err() != nil for unresolvable Ref, got nil")
	}
}

// --------------------------------------------------------------------------
// VPNIPConfig sub-builder
// --------------------------------------------------------------------------

func TestVPNIPConfig_FluentSetters(t *testing.T) {
	cfg := NewVPNIPConfig().
		WithVPC(URI("/projects/p/providers/Aruba.Network/vpcs/v")).
		WithElasticIP(URI("/projects/p/providers/Aruba.Network/elasticIps/eip-1")).
		WithSubnet("my-subnet", "10.0.0.0/24")

	if cfg.Err() != nil {
		t.Errorf("Err() = %v", cfg.Err())
	}
	built := cfg.build()
	if built == nil {
		t.Fatal("build() returned nil")
	}
	if built.VPC == nil || built.VPC.URI != "/projects/p/providers/Aruba.Network/vpcs/v" {
		t.Errorf("VPC.URI = %q", func() string {
			if built.VPC != nil {
				return built.VPC.URI
			}
			return "<nil>"
		}())
	}
	if built.PublicIP == nil || built.PublicIP.URI != "/projects/p/providers/Aruba.Network/elasticIps/eip-1" {
		t.Errorf("PublicIP.URI = %q", func() string {
			if built.PublicIP != nil {
				return built.PublicIP.URI
			}
			return "<nil>"
		}())
	}
	if built.Subnet == nil || built.Subnet.Name != "my-subnet" || built.Subnet.CIDR != "10.0.0.0/24" {
		t.Errorf("Subnet = %+v", built.Subnet)
	}
}

func TestVPNIPConfig_WithVPC_EmptyURI(t *testing.T) {
	cfg := NewVPNIPConfig().WithVPC(URI(""))
	if cfg.Err() == nil {
		t.Fatal("expected Err() != nil for empty URI VPC")
	}
	if !strings.Contains(cfg.Err().Error(), "empty URI") {
		t.Errorf("error = %q, expected 'empty URI'", cfg.Err().Error())
	}
	built := cfg.build()
	if built != nil && built.VPC != nil {
		t.Error("VPC must remain nil when URI was empty")
	}
}

func TestVPNIPConfig_WithElasticIP_EmptyURI(t *testing.T) {
	cfg := NewVPNIPConfig().WithElasticIP(URI(""))
	if cfg.Err() == nil {
		t.Fatal("expected Err() != nil for empty URI PublicIP")
	}
	built := cfg.build()
	if built != nil && built.PublicIP != nil {
		t.Error("PublicIP must remain nil when URI was empty")
	}
}

func TestVPNIPConfig_Build_SubnetOmittedWhenUnset(t *testing.T) {
	cfg := NewVPNIPConfig().WithVPC(URI("/projects/p/providers/Aruba.Network/vpcs/v"))
	built := cfg.build()
	if built.Subnet != nil {
		t.Error("Subnet should be nil when WithSubnet was not called")
	}
}

func TestVPNIPConfig_Build_NilReceiver(t *testing.T) {
	var cfg *VPNIPConfig
	if cfg.build() != nil {
		t.Error("nil receiver build() must return nil")
	}
}

// --------------------------------------------------------------------------
// VPNIKE sub-builder
// --------------------------------------------------------------------------

func TestVPNIKE_FluentSetters(t *testing.T) {
	ike := NewVPNIKE().
		WithLifetimeSeconds(28800).
		WithEncryption(IKEEncryptionAES256).
		WithHash(IKEHashSHA256).
		WithDHGroup(IKEDHGroup14).
		WithDPDAction(IKEDPDActionRestart).
		WithDPDIntervalSeconds(30).
		WithDPDTimeoutSeconds(120)

	built := ike.build()
	if built == nil {
		t.Fatal("build() returned nil")
	}
	if built.Lifetime != 28800 {
		t.Errorf("Lifetime = %d", built.Lifetime)
	}
	if built.Encryption == nil || *built.Encryption != IKEEncryptionAES256 {
		t.Errorf("Encryption = %v", built.Encryption)
	}
	if built.Hash == nil || *built.Hash != IKEHashSHA256 {
		t.Errorf("Hash = %v", built.Hash)
	}
	if built.DHGroup == nil || *built.DHGroup != IKEDHGroup14 {
		t.Errorf("DHGroup = %v", built.DHGroup)
	}
	if built.DPDAction == nil || *built.DPDAction != IKEDPDActionRestart {
		t.Errorf("DPDAction = %v", built.DPDAction)
	}
	if built.DPDInterval != 30 {
		t.Errorf("DPDInterval = %d", built.DPDInterval)
	}
	if built.DPDTimeout != 120 {
		t.Errorf("DPDTimeout = %d", built.DPDTimeout)
	}
}

func TestVPNIKE_Build_PartialFields(t *testing.T) {
	ike := NewVPNIKE().WithLifetimeSeconds(3600)
	built := ike.build()
	if built.Encryption != nil {
		t.Error("Encryption should be nil when not set")
	}
	if built.Hash != nil {
		t.Error("Hash should be nil when not set")
	}
	if built.DHGroup != nil {
		t.Error("DHGroup should be nil when not set")
	}
}

func TestVPNIKE_Build_NilReceiver(t *testing.T) {
	var k *VPNIKE
	if k.build() != nil {
		t.Error("nil receiver build() must return nil")
	}
}

// --------------------------------------------------------------------------
// VPNESP sub-builder
// --------------------------------------------------------------------------

func TestVPNESP_FluentSetters(t *testing.T) {
	esp := NewVPNESP().
		WithLifetimeSeconds(3600).
		WithEncryption(ESPEncryptionAES128).
		WithHash(ESPHashSHA1).
		WithPFS(ESPPFSGroupDHGroup14)

	built := esp.build()
	if built == nil {
		t.Fatal("build() returned nil")
	}
	if built.Lifetime != 3600 {
		t.Errorf("Lifetime = %d", built.Lifetime)
	}
	if built.Encryption == nil || *built.Encryption != ESPEncryptionAES128 {
		t.Errorf("Encryption = %v", built.Encryption)
	}
	if built.Hash == nil || *built.Hash != ESPHashSHA1 {
		t.Errorf("Hash = %v", built.Hash)
	}
	if built.PFS == nil || *built.PFS != ESPPFSGroupDHGroup14 {
		t.Errorf("PFS = %v", built.PFS)
	}
}

func TestVPNESP_Build_PartialFields(t *testing.T) {
	esp := NewVPNESP()
	built := esp.build()
	if built.Encryption != nil || built.Hash != nil || built.PFS != nil {
		t.Error("unset pointer fields should be nil")
	}
	if built.Lifetime != 0 {
		t.Errorf("Lifetime = %d", built.Lifetime)
	}
}

func TestVPNESP_Build_NilReceiver(t *testing.T) {
	var e *VPNESP
	if e.build() != nil {
		t.Error("nil receiver build() must return nil")
	}
}

// --------------------------------------------------------------------------
// VPNPSK sub-builder
// --------------------------------------------------------------------------

func TestVPNPSK_FluentSetters(t *testing.T) {
	psk := NewVPNPSK().
		WithCloudSite("cloud-site-A").
		WithOnPremSite("on-prem-site-B").
		WithKey("s3cr3t")

	built := psk.build()
	if built == nil {
		t.Fatal("build() returned nil")
	}
	if built.CloudSite == nil || *built.CloudSite != "cloud-site-A" {
		t.Errorf("CloudSite = %v", built.CloudSite)
	}
	if built.OnPremSite == nil || *built.OnPremSite != "on-prem-site-B" {
		t.Errorf("OnPremSite = %v", built.OnPremSite)
	}
	if built.Secret == nil || *built.Secret != "s3cr3t" {
		t.Errorf("Secret = %v", built.Secret)
	}
}

func TestVPNPSK_Build_PartialFields(t *testing.T) {
	psk := NewVPNPSK()
	built := psk.build()
	if built.CloudSite != nil || built.OnPremSite != nil || built.Secret != nil {
		t.Error("unset pointer fields should be nil")
	}
}

func TestVPNPSK_Build_NilReceiver(t *testing.T) {
	var p *VPNPSK
	if p.build() != nil {
		t.Error("nil receiver build() must return nil")
	}
}

// --------------------------------------------------------------------------
// toRequest round-trip
// --------------------------------------------------------------------------

func TestVPNTunnel_ToRequestRoundTrip(t *testing.T) {
	tun := NewVPNTunnel().Named(
		"my-tunnel").
		Tagged("t1").
		InRegion(RegionITBGBergamo).
		OfType(VPNTypeSiteToSite).
		WithVPNClientProtocol(VPNClientProtocolIKEv2).
		BilledBy(BillingPeriodHour).
		WithPeerClientPublicIP("1.2.3.4").
		WithIPConfig(
			NewVPNIPConfig().
				WithVPC(URI("/projects/p/providers/Aruba.Network/vpcs/v")).
				WithSubnet("subnet-1", "10.0.0.0/24"),
		).
		WithIKESettings(
			NewVPNIKE().
				WithLifetimeSeconds(28800).
				WithEncryption(IKEEncryptionAES256).
				WithHash(IKEHashSHA256).
				WithDHGroup(IKEDHGroup14),
		).
		WithESPSettings(
			NewVPNESP().
				WithLifetimeSeconds(3600).
				WithPFS(ESPPFSGroupEnable),
		).
		WithPSKSettings(
			NewVPNPSK().
				WithCloudSite("cloud").
				WithKey("secret"),
		)

	req := tun.RawRequest()

	if req.Metadata.Name != "my-tunnel" {
		t.Errorf("Metadata.Name = %q", req.Metadata.Name)
	}
	if len(req.Metadata.Tags) != 1 {
		t.Errorf("Metadata.Tags = %v", req.Metadata.Tags)
	}
	if req.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("Metadata.Location.Value = %q", req.Metadata.Location.Value)
	}
	if req.Properties.VPNType == nil || *req.Properties.VPNType != VPNTypeSiteToSite {
		t.Errorf("Properties.VPNType = %v", req.Properties.VPNType)
	}
	if req.Properties.VPNClientProtocol == nil || *req.Properties.VPNClientProtocol != VPNClientProtocolIKEv2 {
		t.Errorf("Properties.VPNClientProtocol = %v", req.Properties.VPNClientProtocol)
	}
	if req.Properties.BillingPlanCommon == nil || req.Properties.BillingPlanCommon.BillingPeriod == nil || *req.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodHour {
		t.Errorf("BillingPlanCommon.BillingPeriod = %v", req.Properties.BillingPlanCommon)
	}
	if cs := req.Properties.VPNClientSettingsCommon; cs == nil {
		t.Fatal("VPNClientSettingsCommon must be set")
	} else {
		if cs.PeerClientPublicIP == nil || *cs.PeerClientPublicIP != "1.2.3.4" {
			t.Errorf("PeerClientPublicIP = %v", cs.PeerClientPublicIP)
		}
		if cs.IKE == nil {
			t.Fatal("IKE must be set")
		}
		if cs.IKE.Encryption == nil || *cs.IKE.Encryption != IKEEncryptionAES256 {
			t.Errorf("IKE.Encryption = %v", cs.IKE.Encryption)
		}
		if cs.ESP == nil {
			t.Fatal("ESP must be set")
		}
		if cs.PSK == nil {
			t.Fatal("PSK must be set")
		}
	}
	if ip := req.Properties.IPConfigurationsCommon; ip == nil {
		t.Fatal("IPConfigurationsCommon must be set")
	} else {
		if ip.VPC == nil || ip.VPC.URI != "/projects/p/providers/Aruba.Network/vpcs/v" {
			t.Errorf("IPConfig.VPC.URI = %q", func() string {
				if ip.VPC != nil {
					return ip.VPC.URI
				}
				return "<nil>"
			}())
		}
		if ip.Subnet == nil || ip.Subnet.Name != "subnet-1" {
			t.Errorf("IPConfig.Subnet = %v", ip.Subnet)
		}
	}
}

func TestVPNTunnel_ToRequest_NoBillingPeriod_DefaultsToHour(t *testing.T) {
	tun := NewVPNTunnel().
		Named("bare")
	req := tun.RawRequest()
	if req.Properties.BillingPlanCommon == nil || req.Properties.BillingPlanCommon.BillingPeriod == nil || *req.Properties.BillingPlanCommon.BillingPeriod != BillingPeriodHour {
		t.Errorf("BillingPlanCommon.BillingPeriod should default to Hour when not set, got %+v", req.Properties.BillingPlanCommon)
	}
}

func TestVPNTunnel_ToRequest_NoVPNClientSettings_OmitsObject(t *testing.T) {
	tun := NewVPNTunnel().
		Named("bare")
	req := tun.RawRequest()
	if req.Properties.VPNClientSettingsCommon != nil {
		t.Errorf("VPNClientSettingsCommon should be nil when IKE/ESP/PSK/PeerIP all unset")
	}
}

func TestVPNTunnel_ToRequest_PeerClientPublicIPOnly_EmitsClientSettings(t *testing.T) {
	tun := NewVPNTunnel().WithPeerClientPublicIP("5.6.7.8")
	req := tun.RawRequest()
	if req.Properties.VPNClientSettingsCommon == nil {
		t.Fatal("VPNClientSettingsCommon must be non-nil when PeerClientPublicIP is set")
	}
	if req.Properties.VPNClientSettingsCommon.IKE != nil {
		t.Error("IKE should be nil")
	}
	if req.Properties.VPNClientSettingsCommon.ESP != nil {
		t.Error("ESP should be nil")
	}
	if req.Properties.VPNClientSettingsCommon.PSK != nil {
		t.Error("PSK should be nil")
	}
}

// --------------------------------------------------------------------------
// Sub-builder error absorption
// --------------------------------------------------------------------------

func TestVPNTunnel_AbsorbsSubBuilderErrors(t *testing.T) {
	tun := NewVPNTunnel().
		WithIPConfig(NewVPNIPConfig().WithVPC(URI("")))

	if tun.Err() == nil {
		t.Fatal("tunnel.Err() must be non-nil when sub-builder has errors")
	}
}

// --------------------------------------------------------------------------
// fromResponse hydration
// --------------------------------------------------------------------------

func vpnTunnelTestResponse(id, name, uri, projectID string) *types.VPNTunnelResponse {
	state := types.State("Active")
	vpnType := VPNTypeSiteToSite
	proto := VPNClientProtocolIKEv2
	loc := &types.LocationResponse{Value: RegionITBGBergamo}
	peerIP := "1.2.3.4"
	bp := BillingPeriodHour
	return &types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			URI:              &uri,
			Name:             &name,
			Tags:             []string{"vpn-tag"},
			LocationResponse: loc,
			ProjectMetadataResponse: &types.ProjectMetadataResponse{
				ID: projectID,
			},
		},
		Properties: types.VPNTunnelPropertiesResponse{
			VPNType:           &vpnType,
			VPNClientProtocol: &proto,
			BillingPlanCommon: &types.BillingPlanCommon{BillingPeriod: &bp},
			VPNClientSettingsCommon: &types.VPNClientSettingsCommon{
				PeerClientPublicIP: &peerIP,
			},
		},
		Status: types.ResourceStatusResponse{
			State: &state,
		},
	}
}

func TestVPNTunnel_FromResponseHydration(t *testing.T) {
	tun := &VPNTunnel{}
	resp := vpnTunnelTestResponse("t-1", "my-tunnel",
		"/projects/p1/providers/Aruba.Network/vpnTunnels/t-1", "p1")
	tun.fromResponse(resp)

	if tun.ID() != "t-1" {
		t.Errorf("ID() = %q", tun.ID())
	}
	if tun.URI() != "/projects/p1/providers/Aruba.Network/vpnTunnels/t-1" {
		t.Errorf("URI() = %q", tun.URI())
	}
	if tun.VPNTunnelID() != "t-1" {
		t.Errorf("VPNTunnelID() = %q", tun.VPNTunnelID())
	}
	if tun.Name() != "my-tunnel" {
		t.Errorf("Name() = %q", tun.Name())
	}
	if tags := tun.Tags(); len(tags) != 1 || tags[0] != "vpn-tag" {
		t.Errorf("Tags() = %v", tags)
	}
	if tun.Region() != RegionITBGBergamo {
		t.Errorf("Region() = %q", tun.Region())
	}
	if tun.State() != "Active" {
		t.Errorf("State() = %q", tun.State())
	}
	if tun.VPNType() != VPNTypeSiteToSite {
		t.Errorf("VPNType() = %q", tun.VPNType())
	}
	if tun.VPNClientProtocol() != VPNClientProtocolIKEv2 {
		t.Errorf("VPNClientProtocol() = %q", tun.VPNClientProtocol())
	}
	if tun.BillingPeriod() != BillingPeriodHour {
		t.Errorf("BillingPeriod() = %q", tun.BillingPeriod())
	}
	if tun.PeerClientPublicIP() != "1.2.3.4" {
		t.Errorf("PeerClientPublicIP() = %q", tun.PeerClientPublicIP())
	}
	if tun.ProjectID() != "p1" {
		t.Errorf("ProjectID() = %q", tun.ProjectID())
	}
	if tun.Raw() != resp {
		t.Error("Raw() should return the hydrated response pointer")
	}
}

func TestVPNTunnel_FromResponsePartial(t *testing.T) {
	tun := &VPNTunnel{}
	tun.fromResponse(nil)
	if tun.ID() != "" || tun.URI() != "" || tun.Name() != "" {
		t.Error("fromResponse(nil) should be a no-op")
	}
	if tun.Raw() != nil {
		t.Error("Raw() should be nil before hydration")
	}

	tun2 := &VPNTunnel{}
	tun2.fromResponse(&types.VPNTunnelResponse{})
	if tun2.ID() != "" || tun2.URI() != "" || tun2.State() != "" {
		t.Error("empty response should yield zero accessor values")
	}
	if tun2.VPNType() != "" || tun2.BillingPeriod() != "" {
		t.Error("empty Properties should yield empty strings")
	}
}

func TestVPNTunnel_FromResponseURIBackfill(t *testing.T) {
	uri := "/projects/p2/providers/Aruba.Network/vpnTunnels/t-2"
	id := "t-2"
	name := "uri-tunnel"
	resp := &types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:   &id,
			URI:  &uri,
			Name: &name,
			// ProjectMetadataResponse intentionally nil
		},
	}
	tun := &VPNTunnel{}
	tun.fromResponse(resp)

	if tun.ProjectID() != "p2" {
		t.Errorf("ProjectID() via URI fallback = %q", tun.ProjectID())
	}
}

// --------------------------------------------------------------------------
// Ref + ancestor ID satisfaction (runtime)
// --------------------------------------------------------------------------

func TestVPNTunnel_RefSatisfaction(t *testing.T) {
	tun := &VPNTunnel{}
	tun.fromResponse(vpnTunnelTestResponse("t-99", "n",
		"/projects/p99/providers/Aruba.Network/vpnTunnels/t-99", "p99"))

	// withVPNTunnelID typed path
	tid, ok := extractID(tun, func(r Ref) (string, bool) {
		if w, ok := r.(withVPNTunnelID); ok {
			return w.VPNTunnelID(), true
		}
		return "", false
	}, "vpnTunnels")
	if !ok || tid != "t-99" {
		t.Errorf("extractID via withVPNTunnelID = (%q, %v)", tid, ok)
	}

	// withProjectID typed path
	projID, ok := extractID(tun, func(r Ref) (string, bool) {
		if w, ok := r.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || projID != "p99" {
		t.Errorf("extractID via withProjectID = (%q, %v)", projID, ok)
	}
}

// --------------------------------------------------------------------------
// vpnTunnelIDsFromRef helper
// --------------------------------------------------------------------------

func TestVPNTunnelIDsFromRef_TypedRef(t *testing.T) {
	tun := &VPNTunnel{}
	tun.fromResponse(vpnTunnelTestResponse("t-1", "n",
		"/projects/p/providers/Aruba.Network/vpnTunnels/t-1", "p"))
	pid, tid, err := vpnTunnelIDsFromRef(tun)
	if err != nil || pid != "p" || tid != "t-1" {
		t.Errorf("vpnTunnelIDsFromRef typed = (%q, %q, %v)", pid, tid, err)
	}
}

func TestVPNTunnelIDsFromRef_URIRef_CamelCase(t *testing.T) {
	ref := URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1")
	pid, tid, err := vpnTunnelIDsFromRef(ref)
	if err != nil || pid != "p" || tid != "t-1" {
		t.Errorf("vpnTunnelIDsFromRef camelCase = (%q, %q, %v)", pid, tid, err)
	}
}

func TestVPNTunnelIDsFromRef_BadURI_MissingTunnel(t *testing.T) {
	_, _, err := vpnTunnelIDsFromRef(URI("/projects/p/providers/Aruba.Network"))
	if err == nil {
		t.Error("expected error for URI without tunnel segment")
	}
}

func TestVPNTunnelIDsFromRef_BadURI_MissingProject(t *testing.T) {
	_, _, err := vpnTunnelIDsFromRef(URI("/providers/Aruba.Network/vpnTunnels/t-1"))
	if err == nil {
		t.Error("expected error for URI without project segment")
	}
}

func TestVPNTunnelIDsFromRef_BadURI_MissingAll(t *testing.T) {
	_, _, err := vpnTunnelIDsFromRef(URI("/something/else"))
	if err == nil {
		t.Error("expected error for totally invalid URI")
	}
}

// --------------------------------------------------------------------------
// vpnTunnelsClientAdapter — CRUD integration tests
// --------------------------------------------------------------------------

func buildVPNTunnelTestAdapter(t *testing.T, handler http.HandlerFunc) *vpnTunnelsClientAdapter {
	t.Helper()
	server := testutil.NewMockServer(t, handler)
	return newVPNTunnelsClientAdapter(testutil.NewClient(t, server.URL))
}

const vpnTunnelSuccessBody = `{` +
	`"metadata":{` +
	`"id":"t-1","name":"my-tunnel",` +
	`"uri":"/projects/p/providers/Aruba.Network/vpnTunnels/t-1",` +
	`"project":{"id":"p"}` +
	`},` +
	`"properties":{` +
	`"vpnType":"Site-To-Site","vpnClientProtocol":"ikev2"` +
	`},` +
	`"status":{"state":"Active"}}`

func TestVPNTunnelsClientAdapter_Create_Success(t *testing.T) {
	var gotBody types.VPNTunnelRequest
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, vpnTunnelSuccessBody)
	})

	tun := NewVPNTunnel().
		InProject(URI("/projects/p")).
		Named("my-tunnel").
		InRegion(RegionITBGBergamo).
		OfType(VPNTypeSiteToSite).
		WithVPNClientProtocol(VPNClientProtocolIKEv2).
		WithIKESettings(NewVPNIKE().WithEncryption(IKEEncryptionAES256))

	result, err := adapter.Create(context.Background(), tun)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if result.ID() != "t-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.Name() != "my-tunnel" {
		t.Errorf("Name() = %q", result.Name())
	}
	if result.StatusCode() != http.StatusCreated {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if gotBody.Metadata.Name != "my-tunnel" {
		t.Errorf("request Name = %q", gotBody.Metadata.Name)
	}
	if gotBody.Metadata.Location.Value != RegionITBGBergamo {
		t.Errorf("request Location = %q", gotBody.Metadata.Location.Value)
	}
	if gotBody.Properties.VPNType == nil || *gotBody.Properties.VPNType != VPNTypeSiteToSite {
		t.Errorf("request VPNType = %v", gotBody.Properties.VPNType)
	}
	if gotBody.Properties.VPNClientSettingsCommon == nil || gotBody.Properties.VPNClientSettingsCommon.IKE == nil {
		t.Error("request IKE must be present")
	}
}

func TestVPNTunnelsClientAdapter_Create_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})

	_, err := adapter.Create(context.Background(), NewVPNTunnel().
		Named("x"))
	if err == nil {
		t.Fatal("expected error when tunnel has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestVPNTunnelsClientAdapter_Create_MetadataValidationError(t *testing.T) {
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		// Missing "id" field — triggers MetadataValidationError
		fmt.Fprint(w, `{"metadata":{"name":"tunnel","uri":"/projects/p/providers/Aruba.Network/vpnTunnels/x"},"properties":{},"status":{}}`)
	})

	tun := NewVPNTunnel().InProject(URI("/projects/p")).
		Named("tunnel")
	result, err := adapter.Create(context.Background(), tun)
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

func TestVPNTunnelsClientAdapter_Create_NonTwoXX(t *testing.T) {
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Validation Failed", "name is required", 422))
	})

	tun := NewVPNTunnel().InProject(URI("/projects/p"))
	result, err := adapter.Create(context.Background(), tun)
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

func TestVPNTunnelsClientAdapter_Get_URIRef(t *testing.T) {
	var capturedPath string
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, vpnTunnelSuccessBody)
	})

	ref := URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1")
	result, err := adapter.Get(context.Background(), ref)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "t-1" {
		t.Errorf("ID() = %q", result.ID())
	}
	if result.ProjectID() != "p" {
		t.Errorf("ProjectID() = %q", result.ProjectID())
	}
	if result.StatusCode() != http.StatusOK {
		t.Errorf("StatusCode() = %d", result.StatusCode())
	}
	if !strings.Contains(capturedPath, "vpnTunnels") {
		t.Errorf("path = %q, expected vpnTunnels segment", capturedPath)
	}
}

func TestVPNTunnelsClientAdapter_Get_TypedRef(t *testing.T) {
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, vpnTunnelSuccessBody)
	})

	existing := &VPNTunnel{}
	existing.fromResponse(vpnTunnelTestResponse("t-1", "n",
		"/projects/p/providers/Aruba.Network/vpnTunnels/t-1", "p"))

	result, err := adapter.Get(context.Background(), existing)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result.ID() != "t-1" {
		t.Errorf("ID() = %q", result.ID())
	}
}

func TestVPNTunnelsClientAdapter_Update_Success(t *testing.T) {
	var capturedBody types.VPNTunnelRequest
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"metadata":{"id":"t-1","name":"renamed","uri":"/projects/p/providers/Aruba.Network/vpnTunnels/t-1","project":{"id":"p"}},"properties":{},"status":{}}`)
	})

	tun := &VPNTunnel{}
	tun.fromResponse(vpnTunnelTestResponse("t-1", "orig",
		"/projects/p/providers/Aruba.Network/vpnTunnels/t-1", "p"))
	tun.Named("renamed")

	result, err := adapter.Update(context.Background(), tun)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if result.Name() != "renamed" {
		t.Errorf("Name() = %q", result.Name())
	}
	if capturedBody.Metadata.Name != "renamed" {
		t.Errorf("request Name = %q", capturedBody.Metadata.Name)
	}
}

func TestVPNTunnelsClientAdapter_Update_NoID(t *testing.T) {
	callCount := 0
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	tun := NewVPNTunnel().InProject(URI("/projects/p")).
		Named("x")
	_, err := adapter.Update(context.Background(), tun)
	if err == nil {
		t.Fatal("expected error when tunnel has no ID")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when ID is missing")
	}
}

func TestVPNTunnelsClientAdapter_Update_NoProject(t *testing.T) {
	callCount := 0
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	tun := &VPNTunnel{}
	id := "t-1"
	tun.fromResponse(&types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{
			ID: &id,
		},
	})

	_, err := adapter.Update(context.Background(), tun)
	if err == nil {
		t.Fatal("expected error when tunnel has no project")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made without project")
	}
}

func TestVPNTunnelsClientAdapter_Delete_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	err := adapter.Delete(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made for bad Ref")
	}
}

func TestVPNTunnelsClientAdapter_Delete_Success(t *testing.T) {
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1"))
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestVPNTunnelsClientAdapter_Delete_NonTwoXX(t *testing.T) {
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "vpn tunnel not found", 404))
	})

	err := adapter.Delete(context.Background(), URI("/projects/p/providers/Aruba.Network/vpnTunnels/missing"))
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

// InRegion exercises the 0% branch.
func TestVPNTunnel_InRegion(t *testing.T) {
	tun := NewVPNTunnel().
		Tagged("a").
		Tagged("b").
		Untagged("a").
		RetaggedAs("x", "y").
		InRegion("ITMI-Milano-1")

	if tun.Region() != "ITMI-Milano-1" {
		t.Errorf("Region() = %q", tun.Region())
	}
	if tags := tun.Tags(); len(tags) != 2 || tags[0] != "x" || tags[1] != "y" {
		t.Errorf("Tags() = %v", tags)
	}
}

// TestVPNTunnel_SubResourceAccessors covers the 0% IPConfig/IKE/ESP/PSK accessors and
// also exercises the nil branch and error-absorbing branches of
// WithIKESettings/WithESPSettings/WithPSKSettings/WithIPConfig.
func TestVPNTunnel_SubResourceAccessors(t *testing.T) {
	// Nil sub-builders — must not panic and must leave fields nil.
	tunNil := NewVPNTunnel().
		WithIKESettings(nil).
		WithESPSettings(nil).
		WithPSKSettings(nil)
	if tunNil.IKE() != nil {
		t.Error("IKE() should be nil when set with nil")
	}
	if tunNil.ESP() != nil {
		t.Error("ESP() should be nil when set with nil")
	}
	if tunNil.PSK() != nil {
		t.Error("PSK() should be nil when set with nil")
	}
	if tunNil.IPConfig() != nil {
		t.Error("IPConfig() should be nil when not set")
	}

	// Non-nil sub-builders — accessors return the set value.
	tun := NewVPNTunnel().
		WithIPConfig(NewVPNIPConfig()).
		WithIKESettings(NewVPNIKE()).
		WithESPSettings(NewVPNESP()).
		WithPSKSettings(NewVPNPSK())

	if tun.IPConfig() == nil {
		t.Error("IPConfig() nil after WithIPConfig")
	}
	if tun.IKE() == nil {
		t.Error("IKE() nil after WithIKESettings")
	}
	if tun.ESP() == nil {
		t.Error("ESP() nil after WithESPSettings")
	}
	if tun.PSK() == nil {
		t.Error("PSK() nil after WithPSKSettings")
	}

	// Sub-builder with errors — error must be absorbed into tunnel.Err().
	ikeWithErr := NewVPNIKE()
	ikeWithErr.errs = []error{fmt.Errorf("ike-error")}
	espWithErr := NewVPNESP()
	espWithErr.errs = []error{fmt.Errorf("esp-error")}
	pskWithErr := NewVPNPSK()
	pskWithErr.errs = []error{fmt.Errorf("psk-error")}

	tunWithErrs := NewVPNTunnel().
		WithIKESettings(ikeWithErr).
		WithESPSettings(espWithErr).
		WithPSKSettings(pskWithErr)
	if tunWithErrs.Err() == nil {
		t.Error("tunnel.Err() must be non-nil when sub-builders have errors")
	}
}

func TestVPNTunnelsClientAdapter_Get_NonTwoXX(t *testing.T) {
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "vpn tunnel not found", 404))
	})

	ref := URI("/projects/p/providers/Aruba.Network/vpnTunnels/missing")
	result, err := adapter.Get(context.Background(), ref)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d", httpErr.StatusCode)
	}
	if result == nil {
		t.Fatal("result must be non-nil on non-2xx")
	}
}

func TestVPNTunnelsClientAdapter_Update_NonTwoXX(t *testing.T) {
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Not Found", "vpn tunnel not found", 404))
	})

	tun := &VPNTunnel{}
	tun.fromResponse(vpnTunnelTestResponse("t-1", "my-tunnel",
		"/projects/p/providers/Aruba.Network/vpnTunnels/t-1", "p"))
	_, err := adapter.Update(context.Background(), tun)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
}

func TestVPNTunnelsClientAdapter_List_NonTwoXX(t *testing.T) {
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, testutil.ErrorBodyJSON("Forbidden", "access denied", 403))
	})

	_, err := adapter.List(context.Background(), URI("/projects/p"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *HTTPError, got %T: %v", err, err)
	}
}

func TestVPNTunnelIDsFromRef_BadURI_MissingProjectID(t *testing.T) {
	// URI has vpnTunnels segment but no projects segment
	_, _, err := vpnTunnelIDsFromRef(URI("/providers/Aruba.Network/vpnTunnels/t"))
	if err == nil {
		t.Error("expected error for URI without /projects/<id>")
	}
}

func TestVPNTunnelsClientAdapter_Create_WithBuilderError(t *testing.T) {
	callCount := 0
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusCreated)
	})
	tun := NewVPNTunnel().InProject(URI("/garbage"))
	_, err := adapter.Create(context.Background(), tun)
	if err == nil {
		t.Fatal("expected error for builder error")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when builder has errors")
	}
}

func TestVPNTunnelsClientAdapter_Get_BadRef(t *testing.T) {
	callCount := 0
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	result, err := adapter.Get(context.Background(), URI("/something/else"))
	if err == nil {
		t.Fatal("expected error for bad Ref")
	}
	if result != nil {
		t.Error("result should be nil on bad Ref")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made for bad Ref")
	}
}

func TestVPNTunnelsClientAdapter_Get_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPNTunnelsClientAdapter(testutil.NewClient(t, server.URL))
	result, err := adapter.Get(context.Background(),
		URI("/projects/p/providers/Aruba.Network/vpnTunnels/t"))
	if err == nil {
		t.Fatal("expected transport error")
	}
	_ = result
}

func TestVPNTunnelsClientAdapter_Update_WithBuilderError(t *testing.T) {
	callCount := 0
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	tun := NewVPNTunnel().InProject(URI("/garbage"))
	_, err := adapter.Update(context.Background(), tun)
	if err == nil {
		t.Fatal("expected error for builder error")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made when builder has errors")
	}
}

func TestVPNTunnelsClientAdapter_Update_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPNTunnelsClientAdapter(testutil.NewClient(t, server.URL))
	tun := &VPNTunnel{}
	tun.fromResponse(vpnTunnelTestResponse("t-1", "tunnel-a",
		"/projects/p/providers/Aruba.Network/vpnTunnels/t-1", "p"))
	_, err := adapter.Update(context.Background(), tun)
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestVPNTunnelsClientAdapter_Delete_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPNTunnelsClientAdapter(testutil.NewClient(t, server.URL))
	err := adapter.Delete(context.Background(),
		URI("/projects/p/providers/Aruba.Network/vpnTunnels/t"))
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestVPNTunnelsClientAdapter_List_BadProjectRef(t *testing.T) {
	callCount := 0
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})
	_, err := adapter.List(context.Background(), URI("/garbage"))
	if err == nil {
		t.Fatal("expected error for bad project Ref")
	}
	if callCount != 0 {
		t.Error("no HTTP call should be made for bad project Ref")
	}
}

func TestVPNTunnelsClientAdapter_List_TransportError(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Error("server doesn't support hijacking")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	})
	adapter := newVPNTunnelsClientAdapter(testutil.NewClient(t, server.URL))
	_, err := adapter.List(context.Background(), URI("/projects/p"))
	if err == nil {
		t.Fatal("expected transport error")
	}
}

func TestVPNTunnelsClientAdapter_List_ProjectIDBackfill(t *testing.T) {
	// Items without projectID in metadata or URI: triggers v.projectID = projectID backfill
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":1,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"t-x","name":"tunnel-x"},"properties":{},"status":{}}`+
			`]}`)
	})

	list, err := adapter.List(context.Background(), URI("/projects/proj-x"))
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	items := list.Items()
	if len(items) != 1 {
		t.Fatalf("Items() len = %d", len(items))
	}
	if items[0].ProjectID() != "proj-x" {
		t.Errorf("ProjectID() after backfill = %q, want %q", items[0].ProjectID(), "proj-x")
	}
}

func TestVPNTunnelsClientAdapter_List_TwoItems(t *testing.T) {
	adapter := buildVPNTunnelTestAdapter(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"total":2,"self":"","prev":"","next":"","first":"","last":"","values":[`+
			`{"metadata":{"id":"t-1","name":"tunnel-a","uri":"/projects/p/providers/Aruba.Network/vpnTunnels/t-1","project":{"id":"p"}},"properties":{},"status":{"state":"Active"}},`+
			`{"metadata":{"id":"t-2","name":"tunnel-b","uri":"/projects/p/providers/Aruba.Network/vpnTunnels/t-2","project":{"id":"p"}},"properties":{},"status":{"state":"Inactive"}}`+
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
	if items[0].ID() != "t-1" || items[0].Name() != "tunnel-a" {
		t.Errorf("items[0] = {%q, %q}", items[0].ID(), items[0].Name())
	}
	if items[1].ID() != "t-2" || items[1].State() != "Inactive" {
		t.Errorf("items[1] ID=%q State=%q", items[1].ID(), items[1].State())
	}
	if items[0].ProjectID() != "p" {
		t.Errorf("items[0].ProjectID() = %q", items[0].ProjectID())
	}
}

func TestVPNTunnel_FromResponse_SetsStatus(t *testing.T) {
	tun := &VPNTunnel{}
	state := types.State("Active")
	tun.fromResponse(&types.VPNTunnelResponse{
		Status: types.ResourceStatusResponse{State: &state},
	})
	if tun.State() != types.StateActive {
		t.Errorf("State() = %q after fromResponse, want Active", tun.State())
	}
}

func TestVPNTunnelsClientAdapter_Get_InjectsRefresh(t *testing.T) {
	server := testutil.NewMockServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, vpnTunnelSuccessBody)
	})
	adapter := newVPNTunnelsClientAdapter(testutil.NewClient(t, server.URL))
	tun, err := adapter.Get(context.Background(), URI("/projects/p/providers/Aruba.Network/vpnTunnels/t-1"))
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !refreshIsSet(&tun.statusMixin) {
		t.Error("Get should inject a refresh callback into the returned VPNTunnel")
	}
}

func TestVPNTunnelRef(t *testing.T) {
	ref := VPNTunnelRef("p-1", "tun-1")
	want := "/projects/p-1/providers/Aruba.Network/vpnTunnels/tun-1"
	if ref.URI() != want {
		t.Errorf("VPNTunnelRef URI = %q, want %q", ref.URI(), want)
	}
	ids := parseURIIDs(ref.URI())
	if ids["projects"] != "p-1" || ids["vpnTunnels"] != "tun-1" {
		t.Errorf("parseURIIDs = %v", ids)
	}
}

// --------------------------------------------------------------------------
// fromResponse — rehydration of IKE/ESP/PSK/IPConfig sub-builders + RoutesNumber
// --------------------------------------------------------------------------

func TestVPNTunnel_FromResponse_RehydratesSubBuilders(t *testing.T) {
	enc := IKEEncryption("AES256")
	hash := IKEHash("SHA256")
	dhGrp := IKEDHGroup("MODP2048")
	dpdAct := IKEDPDAction("restart")
	espEnc := ESPEncryption("AES256")
	espHash := ESPHash("SHA256")
	espPFS := ESPPFSGroup("MODP2048")
	cloudSite := "cloud-site"
	onPremSite := "on-prem"
	secret := "s3cr3t"
	peerIP := "1.2.3.4"
	vpcURI := "/vpcs/v-1"
	pubIPURI := "/eips/eip-1"

	resp := &types.VPNTunnelResponse{
		Properties: types.VPNTunnelPropertiesResponse{
			RoutesNumber: 3,
			VPNClientSettingsCommon: &types.VPNClientSettingsCommon{
				PeerClientPublicIP: &peerIP,
				IKE: &types.IKESettingsCommon{
					Lifetime:    3600,
					Encryption:  &enc,
					Hash:        &hash,
					DHGroup:     &dhGrp,
					DPDAction:   &dpdAct,
					DPDInterval: 30,
					DPDTimeout:  120,
				},
				ESP: &types.ESPSettingsCommon{
					Lifetime:   1800,
					Encryption: &espEnc,
					Hash:       &espHash,
					PFS:        &espPFS,
				},
				PSK: &types.PSKSettingsCommon{
					CloudSite:  &cloudSite,
					OnPremSite: &onPremSite,
					Secret:     &secret,
				},
			},
			IPConfigurationsCommon: &types.IPConfigurationsCommon{
				VPC:      &types.ReferenceResourceCommon{URI: vpcURI},
				PublicIP: &types.ReferenceResourceCommon{URI: pubIPURI},
				Subnet:   &types.SubnetInfoCommon{Name: "sn-1", CIDR: "10.0.0.0/24"},
			},
		},
	}

	tun := &VPNTunnel{}
	tun.fromResponse(resp)

	if tun.RoutesNumber() != 3 {
		t.Errorf("RoutesNumber() = %d, want 3", tun.RoutesNumber())
	}
	if tun.PeerClientPublicIP() != peerIP {
		t.Errorf("PeerClientPublicIP() = %q", tun.PeerClientPublicIP())
	}
	if tun.IKE() == nil {
		t.Fatal("IKE() is nil after rehydration")
	}
	if tun.IKE().lifetime != 3600 {
		t.Errorf("IKE.lifetime = %d", tun.IKE().lifetime)
	}
	if tun.ESP() == nil {
		t.Fatal("ESP() is nil after rehydration")
	}
	if tun.ESP().lifetime != 1800 {
		t.Errorf("ESP.lifetime = %d", tun.ESP().lifetime)
	}
	if tun.PSK() == nil {
		t.Fatal("PSK() is nil after rehydration")
	}
	if tun.PSK().cloudSite == nil || *tun.PSK().cloudSite != cloudSite {
		t.Errorf("PSK.cloudSite = %v", tun.PSK().cloudSite)
	}
	if tun.IPConfig() == nil {
		t.Fatal("IPConfig() is nil after rehydration")
	}
	if tun.IPConfig().vpc == nil || tun.IPConfig().vpc.URI != vpcURI {
		t.Errorf("IPConfig.vpc = %v", tun.IPConfig().vpc)
	}
	if tun.IPConfig().subnetCIDR != "10.0.0.0/24" {
		t.Errorf("IPConfig.subnetCIDR = %q", tun.IPConfig().subnetCIDR)
	}

	// Round-trip: toRequest must emit the same VPNClientSettingsCommon shape.
	req := tun.toRequest()
	if req.Properties.VPNClientSettingsCommon == nil {
		t.Fatal("toRequest().Properties.VPNClientSettingsCommon is nil after rehydration")
	}
	if req.Properties.VPNClientSettingsCommon.IKE == nil || req.Properties.VPNClientSettingsCommon.IKE.Lifetime != 3600 {
		t.Errorf("round-trip IKE = %+v", req.Properties.VPNClientSettingsCommon.IKE)
	}
	if req.Properties.VPNClientSettingsCommon.ESP == nil || req.Properties.VPNClientSettingsCommon.ESP.Lifetime != 1800 {
		t.Errorf("round-trip ESP = %+v", req.Properties.VPNClientSettingsCommon.ESP)
	}
	if req.Properties.VPNClientSettingsCommon.PSK == nil || req.Properties.VPNClientSettingsCommon.PSK.Secret == nil {
		t.Errorf("round-trip PSK = %+v", req.Properties.VPNClientSettingsCommon.PSK)
	}
	if req.Properties.IPConfigurationsCommon == nil || req.Properties.IPConfigurationsCommon.Subnet == nil {
		t.Errorf("round-trip IPConfigurationsCommon = %+v", req.Properties.IPConfigurationsCommon)
	}
}

func TestVPNTunnel_RoutesNumber_BeforeHydration(t *testing.T) {
	tun := NewVPNTunnel()
	if tun.RoutesNumber() != 0 {
		t.Errorf("RoutesNumber() before hydration = %d, want 0", tun.RoutesNumber())
	}
}
