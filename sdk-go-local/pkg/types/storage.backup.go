package types

type StorageBackupType string

const (
	StorageBackupTypeFull        StorageBackupType = "Full"
	StorageBackupTypeIncremental StorageBackupType = "Incremental"
)

type StorageBackupPropertiesRequest struct {

	// StorageBackupType indicates whether the StorageBackup is full or incremental
	StorageBackupType StorageBackupType `json:"type"`

	// Origin indicates the source volume
	Origin ReferenceResourceCommon `json:"sourceVolume"`

	// RetentionDays indicates the number of days to retain the backup
	RetentionDays *int `json:"retentionDays,omitempty"`

	// BillingPeriod indicates the billing period
	BillingPeriod *BillingPeriod `json:"billingPeriod,omitempty"`
}

type StorageBackupPropertiesResponse struct {

	// StorageBackupType indicates whether the StorageBackup is full or incremental
	Type StorageBackupType `json:"type"`

	// Origin indicates the source volume
	Origin ReferenceResourceCommon `json:"sourceVolume"`

	// RetentionDays indicates the number of days to retain the backup
	RetentionDays *int `json:"retentionDays,omitempty"`

	// BillingPeriod indicates the billing period
	BillingPeriod *BillingPeriod `json:"billingPeriod,omitempty"`
}

type StorageBackupRequest struct {
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	Properties StorageBackupPropertiesRequest `json:"properties"`
}

type StorageBackupResponse struct {
	Metadata ResourceMetadataResponse `json:"metadata"`

	Properties StorageBackupPropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type StorageBackupListResponse struct {
	ListResponse
	Values []StorageBackupResponse `json:"values"`
}
