package types

type KeyPairPropertiesRequest struct {
	Value string `json:"value"`
}

type KeyPairPropertiesResponse struct {
	LinkedResources []LinkedResourceCommon `json:"linkedResources,omitempty"`

	Value string `json:"value"`
}

type KeyPairRequest struct {
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	Properties KeyPairPropertiesRequest `json:"properties"`
}

type KeyPairResponse struct {
	Metadata   ResourceMetadataResponse  `json:"metadata"`
	Properties KeyPairPropertiesResponse `json:"properties"`
	Status     ResourceStatusResponse    `json:"status"`
}

type KeyPairListResponse struct {
	ListResponse
	Values []KeyPairResponse `json:"values"`
}
