package types

type StorageRestorePropertiesRequest struct {
	Target ReferenceResourceCommon `json:"destinationVolume"`
}

type StorageRestorePropertiesResponse struct {
	Destination ReferenceResourceCommon `json:"destinationVolume"`
}

type StorageRestoreRequest struct {
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	Properties StorageRestorePropertiesRequest `json:"properties"`
}

type StorageRestoreResponse struct {
	Metadata ResourceMetadataResponse `json:"metadata"`

	Properties StorageRestorePropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type StorageRestoreListResponse struct {
	ListResponse
	Values []StorageRestoreResponse `json:"values"`
}
