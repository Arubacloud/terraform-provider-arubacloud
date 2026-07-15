package aruba

import "github.com/Arubacloud/sdk-go/pkg/types"

// ---- Sub-builder ----

// VPNPSK is a fluent builder for the PSKSettingsCommon block of a VPNTunnel.
// Construct with NewVPNPSK() and attach via VPNTunnel.WithPSKSettings.
type VPNPSK struct {
	errMixin
	cloudSite  *string
	onPremSite *string
	secret     *string
}

// NewVPNPSK returns a fresh *VPNPSK sub-builder for configuring PSK settings.
func NewVPNPSK() *VPNPSK { return &VPNPSK{} }

// WithCloudSite sets the cloud-side identifier for the PSK tunnel.
func (p *VPNPSK) WithCloudSite(v string) *VPNPSK { p.cloudSite = &v; return p }

// WithOnPremSite sets the on-premises identifier for the PSK tunnel.
func (p *VPNPSK) WithOnPremSite(v string) *VPNPSK { p.onPremSite = &v; return p }

// WithKey sets the pre-shared secret key.
func (p *VPNPSK) WithKey(v string) *VPNPSK { p.secret = &v; return p }

func (p *VPNPSK) build() *types.PSKSettingsCommon {
	if p == nil {
		return nil
	}
	return &types.PSKSettingsCommon{
		CloudSite:  p.cloudSite,
		OnPremSite: p.onPremSite,
		Secret:     p.secret,
	}
}
