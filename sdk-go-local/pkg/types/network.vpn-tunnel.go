package types

// IKEEncryption is the encryption algorithm for IKESettingsCommon.Encryption.
//
// GET /providers/Aruba.Network/vpn-tunnels for the live catalog.
type IKEEncryption string

const (
	IKEEncryptionAES128            IKEEncryption = "aes128"
	IKEEncryptionAES192            IKEEncryption = "aes192"
	IKEEncryptionAES256            IKEEncryption = "aes256"
	IKEEncryptionAES128CTR         IKEEncryption = "aes128ctr"
	IKEEncryptionAES192CTR         IKEEncryption = "aes192ctr"
	IKEEncryptionAES256CTR         IKEEncryption = "aes256ctr"
	IKEEncryptionAES128CCM64       IKEEncryption = "aes128ccm64"
	IKEEncryptionAES128CCM96       IKEEncryption = "aes128ccm96"
	IKEEncryptionAES128CCM128      IKEEncryption = "aes128ccm128"
	IKEEncryptionAES192CCM64       IKEEncryption = "aes192ccm64"
	IKEEncryptionAES192CCM96       IKEEncryption = "aes192ccm96"
	IKEEncryptionAES192CCM128      IKEEncryption = "aes192ccm128"
	IKEEncryptionAES256CCM64       IKEEncryption = "aes256ccm64"
	IKEEncryptionAES256CCM96       IKEEncryption = "aes256ccm96"
	IKEEncryptionAES256CCM128      IKEEncryption = "aes256ccm128"
	IKEEncryptionAES128GCM64       IKEEncryption = "aes128gcm64"
	IKEEncryptionAES128GCM96       IKEEncryption = "aes128gcm96"
	IKEEncryptionAES128GCM128      IKEEncryption = "aes128gcm128"
	IKEEncryptionAES192GCM64       IKEEncryption = "aes192gcm64"
	IKEEncryptionAES192GCM96       IKEEncryption = "aes192gcm96"
	IKEEncryptionAES192GCM128      IKEEncryption = "aes192gcm128"
	IKEEncryptionAES256GCM64       IKEEncryption = "aes256gcm64"
	IKEEncryptionAES256GCM96       IKEEncryption = "aes256gcm96"
	IKEEncryptionAES256GCM128      IKEEncryption = "aes256gcm128"
	IKEEncryptionAES128GMAC        IKEEncryption = "aes128gmac"
	IKEEncryptionAES192GMAC        IKEEncryption = "aes192gmac"
	IKEEncryptionAES256GMAC        IKEEncryption = "aes256gmac"
	IKEEncryption3DES              IKEEncryption = "3des"
	IKEEncryptionBlowfish128       IKEEncryption = "blowfish128"
	IKEEncryptionBlowfish192       IKEEncryption = "blowfish192"
	IKEEncryptionBlowfish256       IKEEncryption = "blowfish256"
	IKEEncryptionCamellia128       IKEEncryption = "camellia128"
	IKEEncryptionCamellia192       IKEEncryption = "camellia192"
	IKEEncryptionCamellia256       IKEEncryption = "camellia256"
	IKEEncryptionCamellia128CTR    IKEEncryption = "camellia128ctr"
	IKEEncryptionCamellia192CTR    IKEEncryption = "camellia192ctr"
	IKEEncryptionCamellia256CTR    IKEEncryption = "camellia256ctr"
	IKEEncryptionCamellia128CCM64  IKEEncryption = "camellia128ccm64"
	IKEEncryptionCamellia128CCM96  IKEEncryption = "camellia128ccm96"
	IKEEncryptionCamellia128CCM128 IKEEncryption = "camellia128ccm128"
	IKEEncryptionCamellia192CCM64  IKEEncryption = "camellia192ccm64"
	IKEEncryptionCamellia192CCM96  IKEEncryption = "camellia192ccm96"
	IKEEncryptionCamellia192CCM128 IKEEncryption = "camellia192ccm128"
	IKEEncryptionCamellia256CCM64  IKEEncryption = "camellia256ccm64"
	IKEEncryptionCamellia256CCM96  IKEEncryption = "camellia256ccm96"
	IKEEncryptionCamellia256CCM128 IKEEncryption = "camellia256ccm128"
	IKEEncryptionSerpent128        IKEEncryption = "serpent128"
	IKEEncryptionSerpent192        IKEEncryption = "serpent192"
	IKEEncryptionSerpent256        IKEEncryption = "serpent256"
	IKEEncryptionTwofish128        IKEEncryption = "twofish128"
	IKEEncryptionTwofish192        IKEEncryption = "twofish192"
	IKEEncryptionTwofish256        IKEEncryption = "twofish256"
	IKEEncryptionCAST128           IKEEncryption = "cast128"
	IKEEncryptionChaCha20Poly1305  IKEEncryption = "chacha20poly1305"
)

// IKEHash is the hash algorithm for IKESettingsCommon.Hash.
type IKEHash string

const (
	IKEHashMD5        IKEHash = "md5"
	IKEHashMD5128     IKEHash = "md5_128"
	IKEHashSHA1       IKEHash = "sha1"
	IKEHashSHA1160    IKEHash = "sha1_160"
	IKEHashSHA256     IKEHash = "sha256"
	IKEHashSHA25696   IKEHash = "sha256_96"
	IKEHashSHA384     IKEHash = "sha384"
	IKEHashSHA512     IKEHash = "sha512"
	IKEHashAESXCBC    IKEHash = "aesxcbc"
	IKEHashAESCMAC    IKEHash = "aescmac"
	IKEHashAES128GMAC IKEHash = "aes128gmac"
	IKEHashAES192GMAC IKEHash = "aes192gmac"
	IKEHashAES256GMAC IKEHash = "aes256gmac"
)

// IKEDHGroup is the Diffie-Hellman group for IKESettingsCommon.DHGroup.
type IKEDHGroup string

const (
	IKEDHGroup1  IKEDHGroup = "1"
	IKEDHGroup2  IKEDHGroup = "2"
	IKEDHGroup5  IKEDHGroup = "5"
	IKEDHGroup14 IKEDHGroup = "14"
	IKEDHGroup15 IKEDHGroup = "15"
	IKEDHGroup16 IKEDHGroup = "16"
	IKEDHGroup17 IKEDHGroup = "17"
	IKEDHGroup18 IKEDHGroup = "18"
	IKEDHGroup19 IKEDHGroup = "19"
	IKEDHGroup20 IKEDHGroup = "20"
	IKEDHGroup21 IKEDHGroup = "21"
	IKEDHGroup22 IKEDHGroup = "22"
	IKEDHGroup23 IKEDHGroup = "23"
	IKEDHGroup24 IKEDHGroup = "24"
	IKEDHGroup25 IKEDHGroup = "25"
	IKEDHGroup26 IKEDHGroup = "26"
	IKEDHGroup27 IKEDHGroup = "27"
	IKEDHGroup28 IKEDHGroup = "28"
	IKEDHGroup29 IKEDHGroup = "29"
	IKEDHGroup30 IKEDHGroup = "30"
	IKEDHGroup31 IKEDHGroup = "31"
	IKEDHGroup32 IKEDHGroup = "32"
)

// IKEDPDAction is the Dead Peer Detection action for IKESettingsCommon.DPDAction.
type IKEDPDAction string

const (
	IKEDPDActionTrap    IKEDPDAction = "trap"
	IKEDPDActionClear   IKEDPDAction = "clear"
	IKEDPDActionRestart IKEDPDAction = "restart"
)

// ESPEncryption is the encryption algorithm for ESPSettingsCommon.Encryption.
//
// GET /providers/Aruba.Network/vpn-tunnels for the live catalog.
type ESPEncryption string

const (
	ESPEncryptionAES128            ESPEncryption = "aes128"
	ESPEncryptionAES192            ESPEncryption = "aes192"
	ESPEncryptionAES256            ESPEncryption = "aes256"
	ESPEncryptionAES128CTR         ESPEncryption = "aes128ctr"
	ESPEncryptionAES192CTR         ESPEncryption = "aes192ctr"
	ESPEncryptionAES256CTR         ESPEncryption = "aes256ctr"
	ESPEncryptionAES128CCM64       ESPEncryption = "aes128ccm64"
	ESPEncryptionAES128CCM96       ESPEncryption = "aes128ccm96"
	ESPEncryptionAES128CCM128      ESPEncryption = "aes128ccm128"
	ESPEncryptionAES192CCM64       ESPEncryption = "aes192ccm64"
	ESPEncryptionAES192CCM96       ESPEncryption = "aes192ccm96"
	ESPEncryptionAES192CCM128      ESPEncryption = "aes192ccm128"
	ESPEncryptionAES256CCM64       ESPEncryption = "aes256ccm64"
	ESPEncryptionAES256CCM96       ESPEncryption = "aes256ccm96"
	ESPEncryptionAES256CCM128      ESPEncryption = "aes256ccm128"
	ESPEncryptionAES128GCM64       ESPEncryption = "aes128gcm64"
	ESPEncryptionAES128GCM96       ESPEncryption = "aes128gcm96"
	ESPEncryptionAES128GCM128      ESPEncryption = "aes128gcm128"
	ESPEncryptionAES192GCM64       ESPEncryption = "aes192gcm64"
	ESPEncryptionAES192GCM96       ESPEncryption = "aes192gcm96"
	ESPEncryptionAES192GCM128      ESPEncryption = "aes192gcm128"
	ESPEncryptionAES256GCM64       ESPEncryption = "aes256gcm64"
	ESPEncryptionAES256GCM96       ESPEncryption = "aes256gcm96"
	ESPEncryptionAES256GCM128      ESPEncryption = "aes256gcm128"
	ESPEncryptionAES128GMAC        ESPEncryption = "aes128gmac"
	ESPEncryptionAES192GMAC        ESPEncryption = "aes192gmac"
	ESPEncryptionAES256GMAC        ESPEncryption = "aes256gmac"
	ESPEncryption3DES              ESPEncryption = "3des"
	ESPEncryptionBlowfish128       ESPEncryption = "blowfish128"
	ESPEncryptionBlowfish192       ESPEncryption = "blowfish192"
	ESPEncryptionBlowfish256       ESPEncryption = "blowfish256"
	ESPEncryptionCamellia128       ESPEncryption = "camellia128"
	ESPEncryptionCamellia192       ESPEncryption = "camellia192"
	ESPEncryptionCamellia256       ESPEncryption = "camellia256"
	ESPEncryptionCamellia128CTR    ESPEncryption = "camellia128ctr"
	ESPEncryptionCamellia192CTR    ESPEncryption = "camellia192ctr"
	ESPEncryptionCamellia256CTR    ESPEncryption = "camellia256ctr"
	ESPEncryptionCamellia128CCM64  ESPEncryption = "camellia128ccm64"
	ESPEncryptionCamellia128CCM96  ESPEncryption = "camellia128ccm96"
	ESPEncryptionCamellia128CCM128 ESPEncryption = "camellia128ccm128"
	ESPEncryptionCamellia192CCM64  ESPEncryption = "camellia192ccm64"
	ESPEncryptionCamellia192CCM96  ESPEncryption = "camellia192ccm96"
	ESPEncryptionCamellia192CCM128 ESPEncryption = "camellia192ccm128"
	ESPEncryptionCamellia256CCM64  ESPEncryption = "camellia256ccm64"
	ESPEncryptionCamellia256CCM96  ESPEncryption = "camellia256ccm96"
	ESPEncryptionCamellia256CCM128 ESPEncryption = "camellia256ccm128"
	ESPEncryptionSerpent128        ESPEncryption = "serpent128"
	ESPEncryptionSerpent192        ESPEncryption = "serpent192"
	ESPEncryptionSerpent256        ESPEncryption = "serpent256"
	ESPEncryptionTwofish128        ESPEncryption = "twofish128"
	ESPEncryptionTwofish192        ESPEncryption = "twofish192"
	ESPEncryptionTwofish256        ESPEncryption = "twofish256"
	ESPEncryptionCAST128           ESPEncryption = "cast128"
	ESPEncryptionChaCha20Poly1305  ESPEncryption = "chacha20poly1305"
)

// ESPHash is the hash algorithm for ESPSettingsCommon.Hash.
type ESPHash string

const (
	ESPHashMD5        ESPHash = "md5"
	ESPHashMD5128     ESPHash = "md5_128"
	ESPHashSHA1       ESPHash = "sha1"
	ESPHashSHA1160    ESPHash = "sha1_160"
	ESPHashSHA256     ESPHash = "sha256"
	ESPHashSHA25696   ESPHash = "sha256_96"
	ESPHashSHA384     ESPHash = "sha384"
	ESPHashSHA512     ESPHash = "sha512"
	ESPHashAESXCBC    ESPHash = "aesxcbc"
	ESPHashAESCMAC    ESPHash = "aescmac"
	ESPHashAES128GMAC ESPHash = "aes128gmac"
	ESPHashAES192GMAC ESPHash = "aes192gmac"
	ESPHashAES256GMAC ESPHash = "aes256gmac"
)

// ESPPFSGroup is the Perfect Forward Secrecy group for ESPSettingsCommon.PFS.
type ESPPFSGroup string

const (
	ESPPFSGroupEnable    ESPPFSGroup = "enable"
	ESPPFSGroupDisable   ESPPFSGroup = "disable"
	ESPPFSGroupDHGroup1  ESPPFSGroup = "dh-group1"
	ESPPFSGroupDHGroup2  ESPPFSGroup = "dh-group2"
	ESPPFSGroupDHGroup5  ESPPFSGroup = "dh-group5"
	ESPPFSGroupDHGroup14 ESPPFSGroup = "dh-group14"
	ESPPFSGroupDHGroup15 ESPPFSGroup = "dh-group15"
	ESPPFSGroupDHGroup16 ESPPFSGroup = "dh-group16"
	ESPPFSGroupDHGroup17 ESPPFSGroup = "dh-group17"
	ESPPFSGroupDHGroup18 ESPPFSGroup = "dh-group18"
	ESPPFSGroupDHGroup19 ESPPFSGroup = "dh-group19"
	ESPPFSGroupDHGroup20 ESPPFSGroup = "dh-group20"
	ESPPFSGroupDHGroup21 ESPPFSGroup = "dh-group21"
	ESPPFSGroupDHGroup22 ESPPFSGroup = "dh-group22"
	ESPPFSGroupDHGroup23 ESPPFSGroup = "dh-group23"
	ESPPFSGroupDHGroup24 ESPPFSGroup = "dh-group24"
	ESPPFSGroupDHGroup25 ESPPFSGroup = "dh-group25"
	ESPPFSGroupDHGroup26 ESPPFSGroup = "dh-group26"
	ESPPFSGroupDHGroup27 ESPPFSGroup = "dh-group27"
	ESPPFSGroupDHGroup28 ESPPFSGroup = "dh-group28"
	ESPPFSGroupDHGroup29 ESPPFSGroup = "dh-group29"
	ESPPFSGroupDHGroup30 ESPPFSGroup = "dh-group30"
	ESPPFSGroupDHGroup31 ESPPFSGroup = "dh-group31"
	ESPPFSGroupDHGroup32 ESPPFSGroup = "dh-group32"
)

// VPNType is the type of VPN tunnel.
type VPNType string

const (
	VPNTypeSiteToSite VPNType = "Site-To-Site"
)

// VPNClientProtocol is the client protocol for a VPN tunnel.
type VPNClientProtocol string

const (
	VPNClientProtocolIKEv2 VPNClientProtocol = "ikev2"
)

// SubnetInfoCommon identifies an existing cloud-side subnet by name and CIDR for use
// in VPN tunnel IP configuration. The CIDR is a routing reference, not a provisioning
// spec — the subnet must already exist. The 400 "overlaps" error means the same CIDR
// is already associated with another VPN tunnel configuration.
type SubnetInfoCommon struct {
	CIDR string `json:"cidr,omitempty"`
	Name string `json:"name,omitempty"`
}

// IPConfigurationsCommon contains network configuration of the VPN tunnel
type IPConfigurationsCommon struct {
	// VPC reference to the VPC (nullable)
	VPC *ReferenceResourceCommon `json:"vpc,omitempty"`

	// Subnet info (nullable)
	Subnet *SubnetInfoCommon `json:"subnet,omitempty"`

	// PublicIP reference to the public IP (nullable)
	PublicIP *ReferenceResourceCommon `json:"publicIp,omitempty"`
}

// IKESettingsCommon contains IKE settings
type IKESettingsCommon struct {
	// Lifetime Lifetime value
	Lifetime int32 `json:"lifetime,omitempty"`

	// Encryption Encryption algorithm (nullable)
	Encryption *IKEEncryption `json:"encryption,omitempty"`

	// Hash Hash algorithm (nullable)
	Hash *IKEHash `json:"hash,omitempty"`

	// DHGroup Diffie-Hellman group (nullable)
	DHGroup *IKEDHGroup `json:"dhGroup,omitempty"`

	// DPDAction Dead Peer Detection action (nullable)
	DPDAction *IKEDPDAction `json:"dpdAction,omitempty"`

	// DPDInterval Dead Peer Detection interval
	DPDInterval int32 `json:"dpdInterval,omitempty"`

	// DPDTimeout Dead Peer Detection timeout
	DPDTimeout int32 `json:"dpdTimeout,omitempty"`
}

// ESPSettingsCommon contains ESP settings
type ESPSettingsCommon struct {
	// Lifetime Lifetime value
	Lifetime int32 `json:"lifetime,omitempty"`

	// Encryption Encryption algorithm (nullable)
	Encryption *ESPEncryption `json:"encryption,omitempty"`

	// Hash Hash algorithm (nullable)
	Hash *ESPHash `json:"hash,omitempty"`

	// PFS Perfect Forward Secrecy (nullable)
	PFS *ESPPFSGroup `json:"pfs,omitempty"`
}

// PSKSettingsCommon contains Pre-Shared Key settings
type PSKSettingsCommon struct {
	// CloudSite Cloud site identifier (nullable)
	CloudSite *string `json:"cloudSite,omitempty"`

	// OnPremSite On-premises site identifier (nullable)
	OnPremSite *string `json:"onPremSite,omitempty"`

	// Secret Pre-shared key secret (nullable)
	Secret *string `json:"secret,omitempty"`
}

// VPNClientSettingsCommon contains client settings of the VPN tunnel
type VPNClientSettingsCommon struct {
	// IKE settings (nullable)
	IKE *IKESettingsCommon `json:"ike,omitempty"`

	// ESP settings (nullable)
	ESP *ESPSettingsCommon `json:"esp,omitempty"`

	// PSK Pre-Shared Key settings (nullable)
	PSK *PSKSettingsCommon `json:"psk,omitempty"`

	// PeerClientPublicIP Peer client public IP address (nullable)
	PeerClientPublicIP *string `json:"peerClientPublicIp,omitempty"`
}

// VPNTunnelPropertiesRequest contains properties of a VPN tunnel
type VPNTunnelPropertiesRequest struct {
	// VPNType Type of VPN tunnel. Admissible values: Site-To-Site (nullable)
	VPNType *VPNType `json:"vpnType,omitempty"`

	// VPNClientProtocol Protocol of the VPN tunnel. Admissible values: ikev2 (nullable)
	VPNClientProtocol *VPNClientProtocol `json:"vpnClientProtocol,omitempty"`

	// IPConfigurationsCommon Network configuration of the VPN tunnel (nullable)
	IPConfigurationsCommon *IPConfigurationsCommon `json:"ipConfigurations,omitempty"`

	// VPNClientSettingsCommon Client settings of the VPN tunnel (nullable)
	VPNClientSettingsCommon *VPNClientSettingsCommon `json:"vpnClientSettings,omitempty"`

	// BillingPlanCommon Billing plan
	BillingPlanCommon *BillingPlanCommon `json:"billingPlan,omitempty"`
}

// VPNTunnelPropertiesResponse contains the response properties of a VPN tunnel
type VPNTunnelPropertiesResponse struct {
	// VPNType Type of the VPN tunnel (nullable)
	VPNType *VPNType `json:"vpnType,omitempty"`

	// VPNClientProtocol Protocol of the VPN tunnel (nullable)
	VPNClientProtocol *VPNClientProtocol `json:"vpnClientProtocol,omitempty"`

	// IPConfigurationsCommon Network configuration of the VPN tunnel (nullable)
	IPConfigurationsCommon *IPConfigurationsCommon `json:"ipConfigurations,omitempty"`

	// VPNClientSettingsCommon Client settings of the VPN tunnel (nullable)
	VPNClientSettingsCommon *VPNClientSettingsCommon `json:"vpnClientSettings,omitempty"`

	// RoutesNumber Number of valid VPN routes of the VPN tunnel
	RoutesNumber int32 `json:"routesNumber,omitempty"`

	// BillingPlanCommon Billing plan (nullable)
	BillingPlanCommon *BillingPlanCommon `json:"billingPlan,omitempty"`
}

type VPNTunnelRequest struct {
	// Metadata of the VPN Tunnel
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	// Spec contains the VPN Tunnel specification
	Properties VPNTunnelPropertiesRequest `json:"properties"`
}

type VPNTunnelResponse struct {
	// Metadata of the VPN Tunnel
	Metadata ResourceMetadataResponse `json:"metadata"`
	// Spec contains the VPN Tunnel specification
	Properties VPNTunnelPropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type VPNTunnelListResponse struct {
	ListResponse
	Values []VPNTunnelResponse `json:"values"`
}
