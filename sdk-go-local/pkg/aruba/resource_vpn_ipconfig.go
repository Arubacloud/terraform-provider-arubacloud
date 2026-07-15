package aruba

import (
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Sub-builder ----

// VPNIPConfig is a fluent builder for the IPConfigurationsCommon block of a VPNTunnel.
// Construct with NewVPNIPConfig() and attach via VPNTunnel.WithIPConfig.
type VPNIPConfig struct {
	errMixin
	vpc        *types.ReferenceResourceCommon
	publicIP   *types.ReferenceResourceCommon
	subnetName string
	subnetCIDR string
	hasSubnet  bool
}

// NewVPNIPConfig returns a fresh *VPNIPConfig sub-builder for configuring IP settings.
func NewVPNIPConfig() *VPNIPConfig { return &VPNIPConfig{} }

// WithVPC sets the VPC reference for this IP configuration. Errors if v's URI is empty.
func (c *VPNIPConfig) WithVPC(v Ref) *VPNIPConfig {
	if v == nil || v.URI() == "" {
		c.addErr(fmt.Errorf("WithVPC: VPC Ref has empty URI"))
		return c
	}
	c.vpc = &types.ReferenceResourceCommon{URI: v.URI()}
	return c
}

// WithElasticIP sets the public elastic IP reference. Errors if v's URI is empty.
func (c *VPNIPConfig) WithElasticIP(v Ref) *VPNIPConfig {
	if v == nil || v.URI() == "" {
		c.addErr(fmt.Errorf("WithElasticIP: PublicIP Ref has empty URI"))
		return c
	}
	c.publicIP = &types.ReferenceResourceCommon{URI: v.URI()}
	return c
}

// WithSubnet sets the name and CIDR of the existing cloud-side subnet to associate
// with this VPN tunnel. The CIDR identifies which cloud subnet routes traffic to/from
// the remote peer; it must not already be used by another VPN tunnel configuration —
// doing so returns a 400 "ipConfigurations.subnet.cidr overlaps" error.
func (c *VPNIPConfig) WithSubnet(name, cidr string) *VPNIPConfig {
	c.subnetName, c.subnetCIDR, c.hasSubnet = name, cidr, true
	return c
}

func (c *VPNIPConfig) build() *types.IPConfigurationsCommon {
	if c == nil {
		return nil
	}
	out := &types.IPConfigurationsCommon{VPC: c.vpc, PublicIP: c.publicIP}
	if c.hasSubnet {
		out.Subnet = &types.SubnetInfoCommon{Name: c.subnetName, CIDR: c.subnetCIDR}
	}
	return out
}
