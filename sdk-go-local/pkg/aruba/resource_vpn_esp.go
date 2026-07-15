package aruba

import "github.com/Arubacloud/sdk-go/pkg/types"

// ---- Sub-builder ----

// VPNESP is a fluent builder for the ESPSettingsCommon block of a VPNTunnel.
// Construct with NewVPNESP() and attach via VPNTunnel.WithESPSettings.
type VPNESP struct {
	errMixin
	lifetime   int32
	encryption *ESPEncryption
	hash       *ESPHash
	pfs        *ESPPFSGroup
}

// NewVPNESP returns a fresh *VPNESP sub-builder for configuring ESP settings.
func NewVPNESP() *VPNESP { return &VPNESP{} }

// WithLifetimeSeconds sets the ESP SA lifetime in seconds.
func (e *VPNESP) WithLifetimeSeconds(s int) *VPNESP { e.lifetime = int32(s); return e }

// WithEncryption sets the ESP encryption algorithm.
func (e *VPNESP) WithEncryption(v ESPEncryption) *VPNESP { e.encryption = &v; return e }

// WithHash sets the ESP integrity hash algorithm.
func (e *VPNESP) WithHash(v ESPHash) *VPNESP { e.hash = &v; return e }

// WithPFS sets the Perfect Forward Secrecy group for ESP.
func (e *VPNESP) WithPFS(v ESPPFSGroup) *VPNESP { e.pfs = &v; return e }

func (e *VPNESP) build() *types.ESPSettingsCommon {
	if e == nil {
		return nil
	}
	return &types.ESPSettingsCommon{
		Lifetime:   e.lifetime,
		Encryption: e.encryption,
		Hash:       e.hash,
		PFS:        e.pfs,
	}
}
