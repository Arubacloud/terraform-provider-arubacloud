package types

// VolumeImage identifies a stock OS template (and any bundled software)
// used to provision a bootable BlockStorage volume.
//
// The constants below mirror the official catalog at
// https://api.arubacloud.com/docs/metadata/ — values not in the catalog
// will be rejected by the API. Personal/custom templates use a different
// identifier scheme and are not enumerated here.

const (
	// VolumeImageWS22001 — Windows Server 2022 64-bit.
	VolumeImageWS22001 string = "WS22-001_W2K22_1_0"
	// VolumeImageWS19001 — Windows Server 2019 64-bit.
	VolumeImageWS19001 string = "WS19-001_W2K19_1_0"
	// VolumeImageLU24001 — Ubuntu Server 24.04.
	VolumeImageLU24001 string = "LU24-001"
	// VolumeImageLU22001 — Ubuntu Server 22.04 LTS 64-bit.
	VolumeImageLU22001 string = "LU22-001"
	// VolumeImageLU20001 — Ubuntu Server 20.04 LTS 64-bit.
	VolumeImageLU20001 string = "LU20-001"
	// VolumeImageDE12001 — Debian 12.
	VolumeImageDE12001 string = "DE12-001"
	// VolumeImageDE11001 — Debian 11 64-bit.
	VolumeImageDE11001 string = "DE11-001"
	// VolumeImageAL90001 — AlmaLinux 9.x 64-bit.
	VolumeImageAL90001 string = "alma9"
	// VolumeImageAL85001 — AlmaLinux 8.x 64-bit.
	VolumeImageAL85001 string = "alma8"
	// VolumeImageLO15001 — openSUSE 15.2 64-bit.
	VolumeImageLO15001 string = "osuse15_2_x64_1_0"
)

// BlockStorageType represents the type of block storage
type BlockStorageType string

const (
	BlockStorageTypeStandard    BlockStorageType = "Standard"
	BlockStorageTypePerformance BlockStorageType = "Performance"
)

type BlockStoragePropertiesRequest struct {

	// SizeGB Size of the block storage in GB
	SizeGB int `json:"sizeGb"`

	// BillingPeriod of the block storage
	BillingPeriod *BillingPeriod `json:"billingPeriod,omitempty"`

	// Zone where blockstorage will be created (optional).
	// If specified, the resource is zonal; otherwise, it is regional.
	Zone *Zone `json:"dataCenter,omitempty"`

	// Type of block storage. Admissible values: Standard, Performance
	Type BlockStorageType `json:"type"`

	Snapshot *ReferenceResourceCommon `json:"snapshot,omitempty"`

	Bootable *bool `json:"bootable,omitempty"`

	Image *string `json:"image,omitempty"`
}

type BlockStoragePropertiesResponse struct {
	LinkedResources []LinkedResourceCommon `json:"linkedResources,omitempty"`

	// SizeGB Size of the block storage in GB
	SizeGB int `json:"sizeGb"`

	// BillingPeriod Billing plan of the block storage
	BillingPeriod *BillingPeriod `json:"billingPeriod,omitempty"`

	//Zone where blockstorage will be created
	Zone Zone `json:"dataCenter"`

	// Type of block storage. Admissible values: Standard, Performance
	Type BlockStorageType `json:"type"`

	Snapshot *ReferenceResourceCommon `json:"snapshot,omitempty"`

	Bootable *bool `json:"bootable,omitempty"`

	Image *string `json:"image,omitempty"`
}

type BlockStorageRequest struct {
	// Metadata of the Block Storage
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	// Spec contains the Block Storage specification
	Properties BlockStoragePropertiesRequest `json:"properties"`
}

type BlockStorageResponse struct {

	// Metadata of the Block Storage
	Metadata ResourceMetadataResponse `json:"metadata"`

	// Spec contains the Block Storage specification
	Properties BlockStoragePropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type BlockStorageListResponse struct {
	ListResponse
	Values []BlockStorageResponse `json:"values"`
}
