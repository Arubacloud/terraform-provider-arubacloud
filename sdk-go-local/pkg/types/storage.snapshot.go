package types

type SnapshotPropertiesRequest struct {
	// BillingPeriod The billing period for blockStorage. Only Hour is a valid value (nullable)
	BillingPeriod *BillingPeriod `json:"billingPeriod,omitempty"`

	Volume ReferenceResourceCommon `json:"volume,omitempty"`
}

// VolumeInfoResponse contains information about the original volume
type VolumeInfoResponse struct {
	// URI of the volume
	URI *string `json:"uri,omitempty"`

	// Type of the original volume from which the snapshot was created (nullable)
	Name *string `json:"name,omitempty"`

	CompoundResource *ReferenceResourceCommon `json:"compoundResource,omitempty"`
}

type SnapshotPropertiesResponse struct {
	// SizeGB The blockStorage's size in gigabyte (nullable)
	SizeGB *int32 `json:"sizeGb,omitempty"`

	// BillingPeriod The billing period for blockStorage. Only Hour is a valid value (nullable)
	BillingPeriod *BillingPeriod `json:"billingPeriod,omitempty"`

	// Volume information about the original volume
	Volume *VolumeInfoResponse `json:"volume,omitempty"`

	// Type of block storage. Admissible values: Standard, Performance
	Type BlockStorageType `json:"type"`

	//Zone where blockstorage will be created
	Zone Zone `json:"dataCenter"`

	Bootable *bool `json:"bootable,omitempty"`
}

type SnapshotRequest struct {
	// Metadata of the Snapshot
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	// Spec contains the Snapshot specification
	Properties SnapshotPropertiesRequest `json:"properties"`
}

type SnapshotResponse struct {
	// Metadata of the Snapshot
	Metadata ResourceMetadataResponse `json:"metadata"`

	// Spec contains the Snapshot specification
	Properties SnapshotPropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type SnapshotListResponse struct {
	ListResponse
	Values []SnapshotResponse `json:"values"`
}
