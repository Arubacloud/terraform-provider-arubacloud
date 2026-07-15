package aruba

import "github.com/Arubacloud/sdk-go/pkg/types"

// ---- Sub-builder ----

// VPNIKE is a fluent builder for the IKESettingsCommon block of a VPNTunnel.
// Construct with NewVPNIKE() and attach via VPNTunnel.WithIKESettings.
type VPNIKE struct {
	errMixin
	lifetime    int32
	encryption  *IKEEncryption
	hash        *IKEHash
	dhGroup     *IKEDHGroup
	dpdAction   *IKEDPDAction
	dpdInterval int32
	dpdTimeout  int32
}

// NewVPNIKE returns a fresh *VPNIKE sub-builder for configuring IKE settings.
func NewVPNIKE() *VPNIKE { return &VPNIKE{} }

// WithLifetimeSeconds sets the IKE SA lifetime in seconds.
func (k *VPNIKE) WithLifetimeSeconds(s int) *VPNIKE { k.lifetime = int32(s); return k }

// WithEncryption sets the IKE encryption algorithm.
func (k *VPNIKE) WithEncryption(v IKEEncryption) *VPNIKE { k.encryption = &v; return k }

// WithHash sets the IKE integrity hash algorithm.
func (k *VPNIKE) WithHash(v IKEHash) *VPNIKE { k.hash = &v; return k }

// WithDHGroup sets the IKE Diffie-Hellman group.
func (k *VPNIKE) WithDHGroup(v IKEDHGroup) *VPNIKE { k.dhGroup = &v; return k }

// WithDPDAction sets the IKE Dead Peer Detection action.
func (k *VPNIKE) WithDPDAction(v IKEDPDAction) *VPNIKE { k.dpdAction = &v; return k }

// WithDPDIntervalSeconds sets the DPD keepalive interval in seconds.
func (k *VPNIKE) WithDPDIntervalSeconds(s int) *VPNIKE { k.dpdInterval = int32(s); return k }

// WithDPDTimeoutSeconds sets the DPD timeout in seconds.
func (k *VPNIKE) WithDPDTimeoutSeconds(s int) *VPNIKE { k.dpdTimeout = int32(s); return k }

func (k *VPNIKE) build() *types.IKESettingsCommon {
	if k == nil {
		return nil
	}
	return &types.IKESettingsCommon{
		Lifetime:    k.lifetime,
		Encryption:  k.encryption,
		Hash:        k.hash,
		DHGroup:     k.dhGroup,
		DPDAction:   k.dpdAction,
		DPDInterval: k.dpdInterval,
		DPDTimeout:  k.dpdTimeout,
	}
}
