package aruba

import "github.com/Arubacloud/sdk-go/pkg/types"

// ---------------------------------------------------------------------------
// Network
// ---------------------------------------------------------------------------

// RuleDirection is the direction of a security group rule.
type RuleDirection = types.RuleDirection

const (
	RuleDirectionIngress = types.RuleDirectionIngress // traffic flowing INTO the resource
	RuleDirectionEgress  = types.RuleDirectionEgress  // traffic flowing OUT OF the resource
)

// RuleProtocol identifies the L4 protocol for a security group rule.
// The authoritative set is fully enumerated: ANY, TCP, UDP, ICMP.
type RuleProtocol = types.RuleProtocol

const (
	RuleProtocolANY  = types.RuleProtocolANY  // match any protocol (wildcard)
	RuleProtocolTCP  = types.RuleProtocolTCP  // TCP
	RuleProtocolUDP  = types.RuleProtocolUDP  // UDP
	RuleProtocolICMP = types.RuleProtocolICMP // ICMP
)

// EndpointTypeDto represents the type of target endpoint in a security rule.
type EndpointTypeDto = types.EndpointTypeDto

const (
	EndpointTypeIP            = types.EndpointTypeIP            // endpoint identified by IP address or CIDR
	EndpointTypeSecurityGroup = types.EndpointTypeSecurityGroup // endpoint identified by reference to a Security Group
)

// SubnetType identifies whether a subnet is basic (L2) or advanced (L3).
type SubnetType = types.SubnetType

const (
	SubnetTypeBasic    = types.SubnetTypeBasic    // L2 subnet without routing or DHCP
	SubnetTypeAdvanced = types.SubnetTypeAdvanced // L3 subnet with routing and DHCP
)

// ---------------------------------------------------------------------------
// Storage
// ---------------------------------------------------------------------------

// BlockStorageType represents the performance tier of a block storage volume.
type BlockStorageType = types.BlockStorageType

const (
	BlockStorageTypeStandard    = types.BlockStorageTypeStandard    // HDD-backed, lower IOPS
	BlockStorageTypePerformance = types.BlockStorageTypePerformance // SSD-backed, higher IOPS
)

// StorageBackupType indicates whether a storage backup is full or incremental.
type StorageBackupType = types.StorageBackupType

const (
	StorageBackupTypeFull        = types.StorageBackupTypeFull        // captures all data in the volume
	StorageBackupTypeIncremental = types.StorageBackupTypeIncremental // captures only blocks changed since the last backup
)

// VolumeImage identifies a stock OS template (and any bundled software)
// used to provision a bootable BlockStorage volume.
//
// The constants below mirror the official catalog at
// https://api.arubacloud.com/docs/metadata/ — values not in the catalog
// will be rejected by the API.

const (
	VolumeImageWS22001 = types.VolumeImageWS22001 // Windows Server 2022 64-bit
	VolumeImageWS19001 = types.VolumeImageWS19001 // Windows Server 2019 64-bit
	VolumeImageLU24001 = types.VolumeImageLU24001 // Ubuntu Server 24.04
	VolumeImageLU22001 = types.VolumeImageLU22001 // Ubuntu Server 22.04 LTS 64-bit
	VolumeImageLU20001 = types.VolumeImageLU20001 // Ubuntu Server 20.04 LTS 64-bit
	VolumeImageDE12001 = types.VolumeImageDE12001 // Debian 12
	VolumeImageDE11001 = types.VolumeImageDE11001 // Debian 11 64-bit
	VolumeImageAL90001 = types.VolumeImageAL90001 // AlmaLinux 9.x 64-bit
	VolumeImageAL85001 = types.VolumeImageAL85001 // AlmaLinux 8.x 64-bit
	VolumeImageLO15001 = types.VolumeImageLO15001 // openSUSE 15.2 64-bit
)

// ---------------------------------------------------------------------------
// Container Registry
// ---------------------------------------------------------------------------

// ContainerRegistrySizeFlavor is the concurrent-users tier for a container registry.
// Wire-encoded into the "size" JSON field of the request.
// Accepted values per the platform: "Small", "Medium", "HighPerf".
type ContainerRegistrySizeFlavor = types.ContainerRegistrySizeFlavor

const (
	ContainerRegistrySizeFlavorSmall    = types.ContainerRegistrySizeFlavorSmall    // small concurrent-users tier
	ContainerRegistrySizeFlavorMedium   = types.ContainerRegistrySizeFlavorMedium   // medium concurrent-users tier
	ContainerRegistrySizeFlavorHighPerf = types.ContainerRegistrySizeFlavorHighPerf // high-performance concurrent-users tier
)

// ---------------------------------------------------------------------------
// Security / KMS
// ---------------------------------------------------------------------------

// KeyAlgorithm identifies the cryptographic algorithm of a KMS key.
type KeyAlgorithm = types.KeyAlgorithm

const (
	KeyAlgorithmAes = types.KeyAlgorithmAes // AES symmetric key
	KeyAlgorithmRsa = types.KeyAlgorithmRsa // RSA asymmetric key pair
)

// KeyCreationSource identifies how a KMS key was provisioned.
type KeyCreationSource = types.KeyCreationSource

const (
	KeyCreationSourceCmp   = types.KeyCreationSourceCmp   // created in-platform via the Cloud Management Platform (inferred)
	KeyCreationSourceOther = types.KeyCreationSourceOther // imported or created via an external mechanism
)

// KeyType distinguishes symmetric from asymmetric KMS keys.
type KeyType = types.KeyType

const (
	KeyTypeSymmetric  = types.KeyTypeSymmetric  // single shared secret (e.g., AES)
	KeyTypeAsymmetric = types.KeyTypeAsymmetric // public/private key pair (e.g., RSA)
)

// KeyStatus tracks the lifecycle state of a KMS key.
type KeyStatus = types.KeyStatus

const (
	KeyStatusActive     = types.KeyStatusActive     // key is available for cryptographic operations
	KeyStatusInCreation = types.KeyStatusInCreation // key is being provisioned
	KeyStatusDeleting   = types.KeyStatusDeleting   // deletion in progress
	KeyStatusDeleted    = types.KeyStatusDeleted    // permanently removed
	KeyStatusFailed     = types.KeyStatusFailed     // provisioning or deletion failed
)

// ServiceStatus tracks the lifecycle state of a KMIP service endpoint.
type ServiceStatus = types.ServiceStatus

const (
	ServiceStatusInCreation           = types.ServiceStatusInCreation           // service is being provisioned
	ServiceStatusActive               = types.ServiceStatusActive               // service is operational
	ServiceStatusUpdating             = types.ServiceStatusUpdating             // configuration update in progress
	ServiceStatusDeleting             = types.ServiceStatusDeleting             // deletion in progress
	ServiceStatusDeleted              = types.ServiceStatusDeleted              // permanently removed
	ServiceStatusFailed               = types.ServiceStatusFailed               // provisioning or deletion failed
	ServiceStatusCertificateAvailable = types.ServiceStatusCertificateAvailable // KMIP client certificate is ready for download
)

// ---------------------------------------------------------------------------
// Schedule
// ---------------------------------------------------------------------------

// JobType represents whether a job fires once or on a recurring schedule.
type JobType = types.JobType

const (
	JobTypeOneShot   = types.JobTypeOneShot   // fires exactly once at the configured time
	JobTypeRecurring = types.JobTypeRecurring // fires repeatedly according to the recurrence pattern
)

// RecurrenceType represents the recurrence pattern of a recurring job.
type RecurrenceType = types.RecurrenceType

const (
	RecurrenceTypeHourly  = types.RecurrenceTypeHourly  // fires every hour
	RecurrenceTypeDaily   = types.RecurrenceTypeDaily   // fires once per day
	RecurrenceTypeWeekly  = types.RecurrenceTypeWeekly  // fires once per week
	RecurrenceTypeMonthly = types.RecurrenceTypeMonthly // fires once per month
	RecurrenceTypeCustom  = types.RecurrenceTypeCustom  // fires on a custom cron-style schedule
)

// DeactiveReasonDto represents the reason a job was deactivated.
type DeactiveReasonDto = types.DeactiveReasonDto

const (
	DeactiveReasonNone            = types.DeactiveReasonNone            // job has not been deactivated
	DeactiveReasonManual          = types.DeactiveReasonManual          // deactivated explicitly by the user
	DeactiveReasonResourceDeleted = types.DeactiveReasonResourceDeleted // deactivated because the target resource was deleted
)

// HTTPVerb is the HTTP method used in a job step action.
type HTTPVerb = types.HTTPVerb

const (
	HTTPVerbGET    = types.HTTPVerbGET    // HTTP GET
	HTTPVerbPOST   = types.HTTPVerbPOST   // HTTP POST
	HTTPVerbPUT    = types.HTTPVerbPUT    // HTTP PUT
	HTTPVerbDELETE = types.HTTPVerbDELETE // HTTP DELETE
	HTTPVerbPATCH  = types.HTTPVerbPATCH  // HTTP PATCH
)

// ---------------------------------------------------------------------------
// Metrics / Alerts
// ---------------------------------------------------------------------------

// ActionType represents the type of action triggered by an alert.
// Note: the wire values for SendSMS and AutoscalingDBaaS use lowercase
// contractions ("SendSms", "AutoscalingDbaas") that differ from the
// constant names.
type ActionType = types.ActionType

const (
	ActionTypeNotificationPanel = types.ActionTypeNotificationPanel // display an in-platform notification (wire: "NotificationPanel")
	ActionTypeSendEmail         = types.ActionTypeSendEmail         // send an email notification (wire: "SendEmail")
	ActionTypeSendSMS           = types.ActionTypeSendSMS           // send an SMS notification (wire: "SendSms")
	ActionTypeAutoscalingDBaaS  = types.ActionTypeAutoscalingDBaaS  // trigger DBaaS autoscaling (wire: "AutoscalingDbaas")
)

// ---------------------------------------------------------------------------
// Location / billing
// ---------------------------------------------------------------------------

// Region identifies an Aruba Cloud datacenter by its v2 API location code.
//
// The pattern observed in SDK fixtures is <COUNTRY><CITY>-<City>
// (e.g., "ITBG-Bergamo"). The complete set of codes is not publicly
// enumerated by the v2 API; add entries as they are confirmed.
type Region = types.Region

const (
	RegionITBGBergamo = types.RegionITBGBergamo // Bergamo (Italy) datacenter
)

// Zone identifies an availability zone within an Aruba Cloud region.
type Zone = types.Zone

const (
	ZoneITBG1 = types.ZoneITBG1 // Bergamo availability zone 1
	ZoneITBG2 = types.ZoneITBG2 // Bergamo availability zone 2
	ZoneITBG3 = types.ZoneITBG3 // Bergamo availability zone 3
)

// ---------------------------------------------------------------------------
// Resource lifecycle state
// ---------------------------------------------------------------------------

// State is the lifecycle state of an Aruba Cloud resource.
// Use the State* constants below instead of raw strings.
type State = types.State

const (
	StateInCreation   = types.StateInCreation   // operation in progress: resource being provisioned
	StateCreating     = types.StateCreating     // operation in progress: resource being created
	StateUpdating     = types.StateUpdating     // operation in progress: configuration update in progress
	StateProvisioning = types.StateProvisioning // operation in progress: platform-level provisioning
	StateDeleting     = types.StateDeleting     // operation in progress: deletion in progress
	StateDisabling    = types.StateDisabling    // operation in progress: disabling in progress
	StateEnabling     = types.StateEnabling     // operation in progress: enabling in progress

	StateActive   = types.StateActive   // settled: resource is active and serving
	StateRunning  = types.StateRunning  // settled: resource is running
	StateStopped  = types.StateStopped  // settled: resource is stopped
	StateNotUsed  = types.StateNotUsed  // settled: resource is free to be bound
	StateReserved = types.StateReserved // settled + bound: reserved as a dependency, not actively in use
	StateInUse    = types.StateInUse    // settled + bound + operational: actively attached or consumed
	StateUsed     = types.StateUsed     // settled + bound + operational: in use (alias variant)
	StateDeleted  = types.StateDeleted  // settled: resource has been deleted

	StateFailed   = types.StateFailed   // failure: provisioning or operational fault
	StateError    = types.StateError    // failure: resource is in an error state
	StateDisabled = types.StateDisabled // failure: administratively disabled, requires manual re-enablement
)

// BillingPeriod identifies the billing cadence for a resource. Not every
// wrapper accepts every period — consult the individual resource documentation
// for the authoritative list of accepted values.
type BillingPeriod = types.BillingPeriod

const (
	BillingPeriodHour  = types.BillingPeriodHour  // hourly billing
	BillingPeriodMonth = types.BillingPeriodMonth // monthly billing (resource-specific)
	BillingPeriodYear  = types.BillingPeriodYear  // yearly billing (resource-specific)
)

// ---------------------------------------------------------------------------
// Compute
// ---------------------------------------------------------------------------

// CloudServerFlavor identifies a Cloud Server SKU.
//
// Pattern: CSO<vCPU>A<RAM>. "CSO" is the SKU family prefix
// ("Cloud Server Optimized"); the number after CSO is the vCPU count
// and the number after A is the RAM in GB. All CSO flavors use a
// balanced compute-to-memory ratio — there is no per-SKU optimization
// suffix for this family.
//
// The constants below cover SKUs referenced in SDK fixtures. The
// authoritative live catalog is at:
//
//	GET /providers/Aruba.Compute/flavors
type CloudServerFlavor = types.CloudServerFlavor

const (
	CloudServerFlavorCSO1A2   = types.CloudServerFlavorCSO1A2   //  1 vCPU,  2 GB RAM
	CloudServerFlavorCSO1A4   = types.CloudServerFlavorCSO1A4   //  1 vCPU,  4 GB RAM
	CloudServerFlavorCSO2A4   = types.CloudServerFlavorCSO2A4   //  2 vCPU,  4 GB RAM
	CloudServerFlavorCSO2A8   = types.CloudServerFlavorCSO2A8   //  2 vCPU,  8 GB RAM
	CloudServerFlavorCSO4A8   = types.CloudServerFlavorCSO4A8   //  4 vCPU,  8 GB RAM
	CloudServerFlavorCSO4A16  = types.CloudServerFlavorCSO4A16  //  4 vCPU, 16 GB RAM
	CloudServerFlavorCSO8A16  = types.CloudServerFlavorCSO8A16  //  8 vCPU, 16 GB RAM
	CloudServerFlavorCSO8A32  = types.CloudServerFlavorCSO8A32  //  8 vCPU, 32 GB RAM
	CloudServerFlavorCSO12A24 = types.CloudServerFlavorCSO12A24 // 12 vCPU, 24 GB RAM
	CloudServerFlavorCSO16A32 = types.CloudServerFlavorCSO16A32 // 16 vCPU, 32 GB RAM
	CloudServerFlavorCSO16A64 = types.CloudServerFlavorCSO16A64 // 16 vCPU, 64 GB RAM
	CloudServerFlavorCSO24A48 = types.CloudServerFlavorCSO24A48 // 24 vCPU, 48 GB RAM
	CloudServerFlavorCSO32A64 = types.CloudServerFlavorCSO32A64 // 32 vCPU, 64 GB RAM
)

// ---------------------------------------------------------------------------
// Database
// ---------------------------------------------------------------------------

// DatabaseEngine identifies the RDBMS engine and version for a DBaaS instance.
type DatabaseEngine = types.DatabaseEngine

const (
	DatabaseEngineMySQL80             = types.DatabaseEngineMySQL80             // MySQL 8.0
	DatabaseEngineMSSQL2022Web        = types.DatabaseEngineMSSQL2022Web        // SQL Server 2022 Web
	DatabaseEngineMSSQL2022Standard   = types.DatabaseEngineMSSQL2022Standard   // SQL Server 2022 Standard
	DatabaseEngineMSSQL2022Enterprise = types.DatabaseEngineMSSQL2022Enterprise // SQL Server 2022 Enterprise
)

// DBaaSFlavor identifies a DBaaS instance SKU.
//
// Pattern: DBO<vCPU>A<RAM>. "DBO" is the SKU family prefix
// ("DataBase Optimized"); the number after DBO is the vCPU count
// and the number after A is the RAM in GB. All DBO flavors use a
// balanced compute-to-memory ratio — there is no per-SKU optimization
// suffix for this family.
//
// The constants below cover SKUs referenced in SDK fixtures. The
// authoritative live catalog is at:
//
//	GET /providers/Aruba.Database/flavors
type DBaaSFlavor = types.DBaaSFlavor

const (
	DBaaSFlavorDBO1A2   = types.DBaaSFlavorDBO1A2   //  1 vCPU,  2 GB RAM
	DBaaSFlavorDBO1A4   = types.DBaaSFlavorDBO1A4   //  1 vCPU,  4 GB RAM
	DBaaSFlavorDBO2A4   = types.DBaaSFlavorDBO2A4   //  2 vCPU,  4 GB RAM
	DBaaSFlavorDBO2A8   = types.DBaaSFlavorDBO2A8   //  2 vCPU,  8 GB RAM
	DBaaSFlavorDBO4A8   = types.DBaaSFlavorDBO4A8   //  4 vCPU,  8 GB RAM
	DBaaSFlavorDBO4A16  = types.DBaaSFlavorDBO4A16  //  4 vCPU, 16 GB RAM
	DBaaSFlavorDBO8A16  = types.DBaaSFlavorDBO8A16  //  8 vCPU, 16 GB RAM
	DBaaSFlavorDBO8A32  = types.DBaaSFlavorDBO8A32  //  8 vCPU, 32 GB RAM
	DBaaSFlavorDBO12A24 = types.DBaaSFlavorDBO12A24 // 12 vCPU, 24 GB RAM
	DBaaSFlavorDBO16A32 = types.DBaaSFlavorDBO16A32 // 16 vCPU, 32 GB RAM
	DBaaSFlavorDBO16A64 = types.DBaaSFlavorDBO16A64 // 16 vCPU, 64 GB RAM
	DBaaSFlavorDBO24A48 = types.DBaaSFlavorDBO24A48 // 24 vCPU, 48 GB RAM
	DBaaSFlavorDBO32A64 = types.DBaaSFlavorDBO32A64 // 32 vCPU, 64 GB RAM
)

// ---------------------------------------------------------------------------
// Container / KaaS
// ---------------------------------------------------------------------------

// KubernetesVersion identifies the Kubernetes version for a KaaS cluster.
//
// The constants below mirror the official catalog at
// https://api.arubacloud.com/docs/metadata/ — values not in the catalog
// will be rejected by the API. The authoritative live catalog is at:
//
//	GET /providers/Aruba.Container/versions
type KubernetesVersion = types.KubernetesVersion

const (
	KubernetesVersion1323 = types.KubernetesVersion1323 // Kubernetes 1.32.3
	KubernetesVersion1332 = types.KubernetesVersion1332 // Kubernetes 1.33.2
	KubernetesVersion1341 = types.KubernetesVersion1341 // Kubernetes 1.34.1

	// Deprecated: Kubernetes 1.31.3 is no longer offered by the platform.
	// Use KubernetesVersion1323 or a newer constant instead.
	// This alias will be removed in v0.3.0.
	KubernetesVersion1313 = KubernetesVersion1323
)

// NodePoolInstance identifies a KaaS node pool instance type.
//
// Pattern: K<vCPU>A<RAM>[R]. "K" is the KaaS SKU prefix. The number after K
// is the vCPU count and the number after A is the RAM in GB. The optional
// trailing "R" denotes a RAM-optimized SKU (RAM = 4 × vCPU); SKUs without
// the suffix are balanced (RAM = 2 × vCPU).
//
// The constants below cover SKUs referenced in SDK fixtures. The
// authoritative live catalog is at:
//
//	GET /providers/Aruba.Container/instances
type NodePoolInstance = types.NodePoolInstance

const (
	NodePoolInstanceK1A2   = types.NodePoolInstanceK1A2   //  1 vCPU,  2 GB RAM (balanced)
	NodePoolInstanceK1A4R  = types.NodePoolInstanceK1A4R  //  1 vCPU,  4 GB RAM (RAM-optimized)
	NodePoolInstanceK2A4   = types.NodePoolInstanceK2A4   //  2 vCPU,  4 GB RAM (balanced)
	NodePoolInstanceK2A8R  = types.NodePoolInstanceK2A8R  //  2 vCPU,  8 GB RAM (RAM-optimized)
	NodePoolInstanceK4A8   = types.NodePoolInstanceK4A8   //  4 vCPU,  8 GB RAM (balanced)
	NodePoolInstanceK4A16R = types.NodePoolInstanceK4A16R //  4 vCPU, 16 GB RAM (RAM-optimized)
	NodePoolInstanceK8A16  = types.NodePoolInstanceK8A16  //  8 vCPU, 16 GB RAM (balanced)
	NodePoolInstanceK8A32R = types.NodePoolInstanceK8A32R //  8 vCPU, 32 GB RAM (RAM-optimized)
	NodePoolInstanceK12A24 = types.NodePoolInstanceK12A24 // 12 vCPU, 24 GB RAM (balanced)
	NodePoolInstanceK16A32 = types.NodePoolInstanceK16A32 // 16 vCPU, 32 GB RAM (balanced)
	NodePoolInstanceK24A48 = types.NodePoolInstanceK24A48 // 24 vCPU, 48 GB RAM (balanced)
	NodePoolInstanceK32A64 = types.NodePoolInstanceK32A64 // 32 vCPU, 64 GB RAM (balanced)
)

// ---------------------------------------------------------------------------
// Parameters
// ---------------------------------------------------------------------------

// AcceptHeader is the MIME type sent in the HTTP Accept header when
// downloading binary artifacts (e.g., KMIP certificates).
type AcceptHeader = types.AcceptHeader

// ---------------------------------------------------------------------------
// VPN IKE crypto
// ---------------------------------------------------------------------------

// IKEEncryption is the encryption algorithm for IKESettingsCommon.Encryption.
// Values are StrongSwan wire identifiers. The authoritative list is at:
//
//	GET /providers/Aruba.Network/vpnTunnels
type IKEEncryption = types.IKEEncryption

const (
	IKEEncryptionAES128            = types.IKEEncryptionAES128            // AES-128 CBC
	IKEEncryptionAES192            = types.IKEEncryptionAES192            // AES-192 CBC
	IKEEncryptionAES256            = types.IKEEncryptionAES256            // AES-256 CBC
	IKEEncryptionAES128CTR         = types.IKEEncryptionAES128CTR         // AES-128 CTR (counter mode, no built-in authentication)
	IKEEncryptionAES192CTR         = types.IKEEncryptionAES192CTR         // AES-192 CTR
	IKEEncryptionAES256CTR         = types.IKEEncryptionAES256CTR         // AES-256 CTR
	IKEEncryptionAES128CCM64       = types.IKEEncryptionAES128CCM64       // AES-128 CCM with 64-bit authentication tag
	IKEEncryptionAES128CCM96       = types.IKEEncryptionAES128CCM96       // AES-128 CCM with 96-bit authentication tag
	IKEEncryptionAES128CCM128      = types.IKEEncryptionAES128CCM128      // AES-128 CCM with 128-bit authentication tag
	IKEEncryptionAES192CCM64       = types.IKEEncryptionAES192CCM64       // AES-192 CCM with 64-bit authentication tag
	IKEEncryptionAES192CCM96       = types.IKEEncryptionAES192CCM96       // AES-192 CCM with 96-bit authentication tag
	IKEEncryptionAES192CCM128      = types.IKEEncryptionAES192CCM128      // AES-192 CCM with 128-bit authentication tag
	IKEEncryptionAES256CCM64       = types.IKEEncryptionAES256CCM64       // AES-256 CCM with 64-bit authentication tag
	IKEEncryptionAES256CCM96       = types.IKEEncryptionAES256CCM96       // AES-256 CCM with 96-bit authentication tag
	IKEEncryptionAES256CCM128      = types.IKEEncryptionAES256CCM128      // AES-256 CCM with 128-bit authentication tag
	IKEEncryptionAES128GCM64       = types.IKEEncryptionAES128GCM64       // AES-128 GCM with 64-bit authentication tag
	IKEEncryptionAES128GCM96       = types.IKEEncryptionAES128GCM96       // AES-128 GCM with 96-bit authentication tag
	IKEEncryptionAES128GCM128      = types.IKEEncryptionAES128GCM128      // AES-128 GCM with 128-bit authentication tag
	IKEEncryptionAES192GCM64       = types.IKEEncryptionAES192GCM64       // AES-192 GCM with 64-bit authentication tag
	IKEEncryptionAES192GCM96       = types.IKEEncryptionAES192GCM96       // AES-192 GCM with 96-bit authentication tag
	IKEEncryptionAES192GCM128      = types.IKEEncryptionAES192GCM128      // AES-192 GCM with 128-bit authentication tag
	IKEEncryptionAES256GCM64       = types.IKEEncryptionAES256GCM64       // AES-256 GCM with 64-bit authentication tag
	IKEEncryptionAES256GCM96       = types.IKEEncryptionAES256GCM96       // AES-256 GCM with 96-bit authentication tag
	IKEEncryptionAES256GCM128      = types.IKEEncryptionAES256GCM128      // AES-256 GCM with 128-bit authentication tag
	IKEEncryptionAES128GMAC        = types.IKEEncryptionAES128GMAC        // AES-128 GMAC (authentication only, no confidentiality)
	IKEEncryptionAES192GMAC        = types.IKEEncryptionAES192GMAC        // AES-192 GMAC (authentication only)
	IKEEncryptionAES256GMAC        = types.IKEEncryptionAES256GMAC        // AES-256 GMAC (authentication only)
	IKEEncryption3DES              = types.IKEEncryption3DES              // Triple-DES (legacy, not recommended)
	IKEEncryptionBlowfish128       = types.IKEEncryptionBlowfish128       // Blowfish-128
	IKEEncryptionBlowfish192       = types.IKEEncryptionBlowfish192       // Blowfish-192
	IKEEncryptionBlowfish256       = types.IKEEncryptionBlowfish256       // Blowfish-256
	IKEEncryptionCamellia128       = types.IKEEncryptionCamellia128       // Camellia-128 CBC
	IKEEncryptionCamellia192       = types.IKEEncryptionCamellia192       // Camellia-192 CBC
	IKEEncryptionCamellia256       = types.IKEEncryptionCamellia256       // Camellia-256 CBC
	IKEEncryptionCamellia128CTR    = types.IKEEncryptionCamellia128CTR    // Camellia-128 CTR
	IKEEncryptionCamellia192CTR    = types.IKEEncryptionCamellia192CTR    // Camellia-192 CTR
	IKEEncryptionCamellia256CTR    = types.IKEEncryptionCamellia256CTR    // Camellia-256 CTR
	IKEEncryptionCamellia128CCM64  = types.IKEEncryptionCamellia128CCM64  // Camellia-128 CCM with 64-bit authentication tag
	IKEEncryptionCamellia128CCM96  = types.IKEEncryptionCamellia128CCM96  // Camellia-128 CCM with 96-bit authentication tag
	IKEEncryptionCamellia128CCM128 = types.IKEEncryptionCamellia128CCM128 // Camellia-128 CCM with 128-bit authentication tag
	IKEEncryptionCamellia192CCM64  = types.IKEEncryptionCamellia192CCM64  // Camellia-192 CCM with 64-bit authentication tag
	IKEEncryptionCamellia192CCM96  = types.IKEEncryptionCamellia192CCM96  // Camellia-192 CCM with 96-bit authentication tag
	IKEEncryptionCamellia192CCM128 = types.IKEEncryptionCamellia192CCM128 // Camellia-192 CCM with 128-bit authentication tag
	IKEEncryptionCamellia256CCM64  = types.IKEEncryptionCamellia256CCM64  // Camellia-256 CCM with 64-bit authentication tag
	IKEEncryptionCamellia256CCM96  = types.IKEEncryptionCamellia256CCM96  // Camellia-256 CCM with 96-bit authentication tag
	IKEEncryptionCamellia256CCM128 = types.IKEEncryptionCamellia256CCM128 // Camellia-256 CCM with 128-bit authentication tag
	IKEEncryptionSerpent128        = types.IKEEncryptionSerpent128        // Serpent-128
	IKEEncryptionSerpent192        = types.IKEEncryptionSerpent192        // Serpent-192
	IKEEncryptionSerpent256        = types.IKEEncryptionSerpent256        // Serpent-256
	IKEEncryptionTwofish128        = types.IKEEncryptionTwofish128        // Twofish-128
	IKEEncryptionTwofish192        = types.IKEEncryptionTwofish192        // Twofish-192
	IKEEncryptionTwofish256        = types.IKEEncryptionTwofish256        // Twofish-256
	IKEEncryptionCAST128           = types.IKEEncryptionCAST128           // CAST-128 (CAST5)
	IKEEncryptionChaCha20Poly1305  = types.IKEEncryptionChaCha20Poly1305  // ChaCha20-Poly1305 AEAD
)

// IKEHash is the hash/PRF algorithm for IKESettingsCommon.Hash.
type IKEHash = types.IKEHash

const (
	IKEHashMD5        = types.IKEHashMD5        // HMAC-MD5 (legacy, not recommended)
	IKEHashMD5128     = types.IKEHashMD5128     // HMAC-MD5 truncated to 128 bits
	IKEHashSHA1       = types.IKEHashSHA1       // HMAC-SHA-1
	IKEHashSHA1160    = types.IKEHashSHA1160    // HMAC-SHA-1 with full 160-bit output
	IKEHashSHA256     = types.IKEHashSHA256     // HMAC-SHA-256
	IKEHashSHA25696   = types.IKEHashSHA25696   // HMAC-SHA-256 truncated to 96 bits
	IKEHashSHA384     = types.IKEHashSHA384     // HMAC-SHA-384
	IKEHashSHA512     = types.IKEHashSHA512     // HMAC-SHA-512
	IKEHashAESXCBC    = types.IKEHashAESXCBC    // AES-XCBC-MAC-96
	IKEHashAESCMAC    = types.IKEHashAESCMAC    // AES-CMAC-96
	IKEHashAES128GMAC = types.IKEHashAES128GMAC // AES-128-GMAC
	IKEHashAES192GMAC = types.IKEHashAES192GMAC // AES-192-GMAC
	IKEHashAES256GMAC = types.IKEHashAES256GMAC // AES-256-GMAC
)

// IKEDHGroup is the Diffie-Hellman group for IKESettingsCommon.DHGroup.
// Groups 3, 4, and 6–13 are not exposed by the platform.
type IKEDHGroup = types.IKEDHGroup

const (
	IKEDHGroup1  = types.IKEDHGroup1  // MODP-768 (obsolete, do not use)
	IKEDHGroup2  = types.IKEDHGroup2  // MODP-1024 (legacy, do not use)
	IKEDHGroup5  = types.IKEDHGroup5  // MODP-1536 (deprecated)
	IKEDHGroup14 = types.IKEDHGroup14 // MODP-2048
	IKEDHGroup15 = types.IKEDHGroup15 // MODP-3072
	IKEDHGroup16 = types.IKEDHGroup16 // MODP-4096
	IKEDHGroup17 = types.IKEDHGroup17 // MODP-6144
	IKEDHGroup18 = types.IKEDHGroup18 // MODP-8192
	IKEDHGroup19 = types.IKEDHGroup19 // ECP-256 (P-256 / secp256r1)
	IKEDHGroup20 = types.IKEDHGroup20 // ECP-384 (P-384 / secp384r1)
	IKEDHGroup21 = types.IKEDHGroup21 // ECP-521 (P-521 / secp521r1)
	IKEDHGroup22 = types.IKEDHGroup22 // MODP-1024 with 160-bit prime-order subgroup
	IKEDHGroup23 = types.IKEDHGroup23 // MODP-2048 with 224-bit prime-order subgroup
	IKEDHGroup24 = types.IKEDHGroup24 // MODP-2048 with 256-bit prime-order subgroup
	IKEDHGroup25 = types.IKEDHGroup25 // ECP-192 (secp192r1)
	IKEDHGroup26 = types.IKEDHGroup26 // ECP-224 (secp224r1)
	IKEDHGroup27 = types.IKEDHGroup27 // Brainpool P-224r1
	IKEDHGroup28 = types.IKEDHGroup28 // Brainpool P-256r1
	IKEDHGroup29 = types.IKEDHGroup29 // Brainpool P-384r1
	IKEDHGroup30 = types.IKEDHGroup30 // Brainpool P-512r1
	IKEDHGroup31 = types.IKEDHGroup31 // Curve25519 (X25519)
	IKEDHGroup32 = types.IKEDHGroup32 // Curve448 (X448)
)

// IKEDPDAction is the Dead Peer Detection action for IKESettingsCommon.DPDAction.
type IKEDPDAction = types.IKEDPDAction

const (
	IKEDPDActionTrap    = types.IKEDPDActionTrap    // keep the SA and log the dead peer (traffic triggers DPD)
	IKEDPDActionClear   = types.IKEDPDActionClear   // drop the SA and let traffic fail
	IKEDPDActionRestart = types.IKEDPDActionRestart // drop and immediately re-establish the SA
)

// ---------------------------------------------------------------------------
// VPN ESP crypto
// ---------------------------------------------------------------------------

// ESPEncryption is the encryption algorithm for ESPSettingsCommon.Encryption.
// Values are StrongSwan wire identifiers. The authoritative list is at:
//
//	GET /providers/Aruba.Network/vpnTunnels
type ESPEncryption = types.ESPEncryption

const (
	ESPEncryptionAES128            = types.ESPEncryptionAES128            // AES-128 CBC
	ESPEncryptionAES192            = types.ESPEncryptionAES192            // AES-192 CBC
	ESPEncryptionAES256            = types.ESPEncryptionAES256            // AES-256 CBC
	ESPEncryptionAES128CTR         = types.ESPEncryptionAES128CTR         // AES-128 CTR (counter mode, no built-in authentication)
	ESPEncryptionAES192CTR         = types.ESPEncryptionAES192CTR         // AES-192 CTR
	ESPEncryptionAES256CTR         = types.ESPEncryptionAES256CTR         // AES-256 CTR
	ESPEncryptionAES128CCM64       = types.ESPEncryptionAES128CCM64       // AES-128 CCM with 64-bit authentication tag
	ESPEncryptionAES128CCM96       = types.ESPEncryptionAES128CCM96       // AES-128 CCM with 96-bit authentication tag
	ESPEncryptionAES128CCM128      = types.ESPEncryptionAES128CCM128      // AES-128 CCM with 128-bit authentication tag
	ESPEncryptionAES192CCM64       = types.ESPEncryptionAES192CCM64       // AES-192 CCM with 64-bit authentication tag
	ESPEncryptionAES192CCM96       = types.ESPEncryptionAES192CCM96       // AES-192 CCM with 96-bit authentication tag
	ESPEncryptionAES192CCM128      = types.ESPEncryptionAES192CCM128      // AES-192 CCM with 128-bit authentication tag
	ESPEncryptionAES256CCM64       = types.ESPEncryptionAES256CCM64       // AES-256 CCM with 64-bit authentication tag
	ESPEncryptionAES256CCM96       = types.ESPEncryptionAES256CCM96       // AES-256 CCM with 96-bit authentication tag
	ESPEncryptionAES256CCM128      = types.ESPEncryptionAES256CCM128      // AES-256 CCM with 128-bit authentication tag
	ESPEncryptionAES128GCM64       = types.ESPEncryptionAES128GCM64       // AES-128 GCM with 64-bit authentication tag
	ESPEncryptionAES128GCM96       = types.ESPEncryptionAES128GCM96       // AES-128 GCM with 96-bit authentication tag
	ESPEncryptionAES128GCM128      = types.ESPEncryptionAES128GCM128      // AES-128 GCM with 128-bit authentication tag
	ESPEncryptionAES192GCM64       = types.ESPEncryptionAES192GCM64       // AES-192 GCM with 64-bit authentication tag
	ESPEncryptionAES192GCM96       = types.ESPEncryptionAES192GCM96       // AES-192 GCM with 96-bit authentication tag
	ESPEncryptionAES192GCM128      = types.ESPEncryptionAES192GCM128      // AES-192 GCM with 128-bit authentication tag
	ESPEncryptionAES256GCM64       = types.ESPEncryptionAES256GCM64       // AES-256 GCM with 64-bit authentication tag
	ESPEncryptionAES256GCM96       = types.ESPEncryptionAES256GCM96       // AES-256 GCM with 96-bit authentication tag
	ESPEncryptionAES256GCM128      = types.ESPEncryptionAES256GCM128      // AES-256 GCM with 128-bit authentication tag
	ESPEncryptionAES128GMAC        = types.ESPEncryptionAES128GMAC        // AES-128 GMAC (authentication only, no confidentiality)
	ESPEncryptionAES192GMAC        = types.ESPEncryptionAES192GMAC        // AES-192 GMAC (authentication only)
	ESPEncryptionAES256GMAC        = types.ESPEncryptionAES256GMAC        // AES-256 GMAC (authentication only)
	ESPEncryption3DES              = types.ESPEncryption3DES              // Triple-DES (legacy, not recommended)
	ESPEncryptionBlowfish128       = types.ESPEncryptionBlowfish128       // Blowfish-128
	ESPEncryptionBlowfish192       = types.ESPEncryptionBlowfish192       // Blowfish-192
	ESPEncryptionBlowfish256       = types.ESPEncryptionBlowfish256       // Blowfish-256
	ESPEncryptionCamellia128       = types.ESPEncryptionCamellia128       // Camellia-128 CBC
	ESPEncryptionCamellia192       = types.ESPEncryptionCamellia192       // Camellia-192 CBC
	ESPEncryptionCamellia256       = types.ESPEncryptionCamellia256       // Camellia-256 CBC
	ESPEncryptionCamellia128CTR    = types.ESPEncryptionCamellia128CTR    // Camellia-128 CTR
	ESPEncryptionCamellia192CTR    = types.ESPEncryptionCamellia192CTR    // Camellia-192 CTR
	ESPEncryptionCamellia256CTR    = types.ESPEncryptionCamellia256CTR    // Camellia-256 CTR
	ESPEncryptionCamellia128CCM64  = types.ESPEncryptionCamellia128CCM64  // Camellia-128 CCM with 64-bit authentication tag
	ESPEncryptionCamellia128CCM96  = types.ESPEncryptionCamellia128CCM96  // Camellia-128 CCM with 96-bit authentication tag
	ESPEncryptionCamellia128CCM128 = types.ESPEncryptionCamellia128CCM128 // Camellia-128 CCM with 128-bit authentication tag
	ESPEncryptionCamellia192CCM64  = types.ESPEncryptionCamellia192CCM64  // Camellia-192 CCM with 64-bit authentication tag
	ESPEncryptionCamellia192CCM96  = types.ESPEncryptionCamellia192CCM96  // Camellia-192 CCM with 96-bit authentication tag
	ESPEncryptionCamellia192CCM128 = types.ESPEncryptionCamellia192CCM128 // Camellia-192 CCM with 128-bit authentication tag
	ESPEncryptionCamellia256CCM64  = types.ESPEncryptionCamellia256CCM64  // Camellia-256 CCM with 64-bit authentication tag
	ESPEncryptionCamellia256CCM96  = types.ESPEncryptionCamellia256CCM96  // Camellia-256 CCM with 96-bit authentication tag
	ESPEncryptionCamellia256CCM128 = types.ESPEncryptionCamellia256CCM128 // Camellia-256 CCM with 128-bit authentication tag
	ESPEncryptionSerpent128        = types.ESPEncryptionSerpent128        // Serpent-128
	ESPEncryptionSerpent192        = types.ESPEncryptionSerpent192        // Serpent-192
	ESPEncryptionSerpent256        = types.ESPEncryptionSerpent256        // Serpent-256
	ESPEncryptionTwofish128        = types.ESPEncryptionTwofish128        // Twofish-128
	ESPEncryptionTwofish192        = types.ESPEncryptionTwofish192        // Twofish-192
	ESPEncryptionTwofish256        = types.ESPEncryptionTwofish256        // Twofish-256
	ESPEncryptionCAST128           = types.ESPEncryptionCAST128           // CAST-128 (CAST5)
	ESPEncryptionChaCha20Poly1305  = types.ESPEncryptionChaCha20Poly1305  // ChaCha20-Poly1305 AEAD
)

// ESPHash is the integrity/authentication algorithm for ESPSettingsCommon.Hash.
type ESPHash = types.ESPHash

const (
	ESPHashMD5        = types.ESPHashMD5        // HMAC-MD5 (legacy, not recommended)
	ESPHashMD5128     = types.ESPHashMD5128     // HMAC-MD5 truncated to 128 bits
	ESPHashSHA1       = types.ESPHashSHA1       // HMAC-SHA-1
	ESPHashSHA1160    = types.ESPHashSHA1160    // HMAC-SHA-1 with full 160-bit output
	ESPHashSHA256     = types.ESPHashSHA256     // HMAC-SHA-256
	ESPHashSHA25696   = types.ESPHashSHA25696   // HMAC-SHA-256 truncated to 96 bits
	ESPHashSHA384     = types.ESPHashSHA384     // HMAC-SHA-384
	ESPHashSHA512     = types.ESPHashSHA512     // HMAC-SHA-512
	ESPHashAESXCBC    = types.ESPHashAESXCBC    // AES-XCBC-MAC-96
	ESPHashAESCMAC    = types.ESPHashAESCMAC    // AES-CMAC-96
	ESPHashAES128GMAC = types.ESPHashAES128GMAC // AES-128-GMAC
	ESPHashAES192GMAC = types.ESPHashAES192GMAC // AES-192-GMAC
	ESPHashAES256GMAC = types.ESPHashAES256GMAC // AES-256-GMAC
)

// ESPPFSGroup is the Perfect Forward Secrecy DH group for ESPSettingsCommon.PFS.
// Wire values use a "dh-group<N>" prefix (e.g., "dh-group14"), unlike the
// bare numeric IDs used by IKEDHGroup. Groups 3, 4, and 6–13 are not exposed.
type ESPPFSGroup = types.ESPPFSGroup

const (
	ESPPFSGroupEnable    = types.ESPPFSGroupEnable    // PFS on; DH group negotiated automatically
	ESPPFSGroupDisable   = types.ESPPFSGroupDisable   // PFS off
	ESPPFSGroupDHGroup1  = types.ESPPFSGroupDHGroup1  // MODP-768 (obsolete, do not use)
	ESPPFSGroupDHGroup2  = types.ESPPFSGroupDHGroup2  // MODP-1024 (legacy, do not use)
	ESPPFSGroupDHGroup5  = types.ESPPFSGroupDHGroup5  // MODP-1536 (deprecated)
	ESPPFSGroupDHGroup14 = types.ESPPFSGroupDHGroup14 // MODP-2048
	ESPPFSGroupDHGroup15 = types.ESPPFSGroupDHGroup15 // MODP-3072
	ESPPFSGroupDHGroup16 = types.ESPPFSGroupDHGroup16 // MODP-4096
	ESPPFSGroupDHGroup17 = types.ESPPFSGroupDHGroup17 // MODP-6144
	ESPPFSGroupDHGroup18 = types.ESPPFSGroupDHGroup18 // MODP-8192
	ESPPFSGroupDHGroup19 = types.ESPPFSGroupDHGroup19 // ECP-256 (P-256 / secp256r1)
	ESPPFSGroupDHGroup20 = types.ESPPFSGroupDHGroup20 // ECP-384 (P-384 / secp384r1)
	ESPPFSGroupDHGroup21 = types.ESPPFSGroupDHGroup21 // ECP-521 (P-521 / secp521r1)
	ESPPFSGroupDHGroup22 = types.ESPPFSGroupDHGroup22 // MODP-1024 with 160-bit prime-order subgroup
	ESPPFSGroupDHGroup23 = types.ESPPFSGroupDHGroup23 // MODP-2048 with 224-bit prime-order subgroup
	ESPPFSGroupDHGroup24 = types.ESPPFSGroupDHGroup24 // MODP-2048 with 256-bit prime-order subgroup
	ESPPFSGroupDHGroup25 = types.ESPPFSGroupDHGroup25 // ECP-192 (secp192r1)
	ESPPFSGroupDHGroup26 = types.ESPPFSGroupDHGroup26 // ECP-224 (secp224r1)
	ESPPFSGroupDHGroup27 = types.ESPPFSGroupDHGroup27 // Brainpool P-224r1
	ESPPFSGroupDHGroup28 = types.ESPPFSGroupDHGroup28 // Brainpool P-256r1
	ESPPFSGroupDHGroup29 = types.ESPPFSGroupDHGroup29 // Brainpool P-384r1
	ESPPFSGroupDHGroup30 = types.ESPPFSGroupDHGroup30 // Brainpool P-512r1
	ESPPFSGroupDHGroup31 = types.ESPPFSGroupDHGroup31 // Curve25519 (X25519)
	ESPPFSGroupDHGroup32 = types.ESPPFSGroupDHGroup32 // Curve448 (X448)
)

// ---------------------------------------------------------------------------
// VPN tunnel
// ---------------------------------------------------------------------------

// VPNType identifies the topology of a VPN tunnel.
type VPNType = types.VPNType

const (
	VPNTypeSiteToSite = types.VPNTypeSiteToSite // site-to-site IPsec tunnel between two fixed endpoints
)

// VPNClientProtocol identifies the key-exchange protocol for a VPN tunnel.
type VPNClientProtocol = types.VPNClientProtocol

const (
	VPNClientProtocolIKEv2 = types.VPNClientProtocolIKEv2 // IKEv2 (RFC 7296)
)
