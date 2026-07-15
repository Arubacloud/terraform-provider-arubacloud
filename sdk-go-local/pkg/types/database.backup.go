package types

// DatabaseNameRef holds the name of a logical database for backup operations.
// The backup API identifies databases by name (not by URI).
type DatabaseNameRef struct {
	Name string `json:"name"`
}

type BackupPropertiesRequest struct {
	Zone Zone `json:"datacenter"`

	DBaaS ReferenceResourceCommon `json:"dbaas"`

	// Database identifies the logical database to back up by name.
	Database DatabaseNameRef `json:"database"`

	BillingPlanCommon *BillingPlanCommon `json:"billingPlan,omitempty"`
}

type BackupStorageResponse struct {
	Size int32 `json:"size"`
}

type BackupPropertiesResponse struct {
	LinkedResources []LinkedResourceCommon `json:"linkedResources,omitempty"`

	Zone Zone `json:"datacenter"`

	DBaaS ReferenceResourceCommon `json:"dbaas"`

	// Database holds the name of the logical database that was backed up.
	Database DatabaseNameRef `json:"database"`

	BillingPlanCommon *BillingPlanCommon `json:"billingPlan,omitempty"`

	Storage BackupStorageResponse `json:"storage"`
}

type BackupRequest struct {
	Metadata   RegionalResourceMetadataRequest `json:"metadata"`
	Properties BackupPropertiesRequest         `json:"properties"`
}

type BackupResponse struct {
	Metadata   ResourceMetadataResponse `json:"metadata"`
	Properties BackupPropertiesResponse `json:"properties"`
	Status     ResourceStatusResponse   `json:"status,omitempty"`
}

type DBaaSBackupListResponse struct {
	ListResponse
	Values []BackupResponse `json:"values"`
}
