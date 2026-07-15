package types

// VPCPeeringRoutePropertiesRequest contains properties of a VPC peering route to create
type VPCPeeringPropertiesRequest struct {
	RemoteVPC *ReferenceResourceCommon `json:"remoteVpc,omitempty"`
}

type VPCPeeringPropertiesResponse struct {
	LinkedResources []LinkedResourceCommon   `json:"linkedResources,omitempty"`
	RemoteVPC       *ReferenceResourceCommon `json:"remoteVpc,omitempty"`
}

type VPCPeeringRequest struct {
	// Metadata of the VPC Peering
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	// Spec contains the VPC Peering specification
	Properties VPCPeeringPropertiesRequest `json:"properties"`
}

type VPCPeeringResponse struct {
	// Metadata of the VPC Peering
	Metadata ResourceMetadataResponse `json:"metadata"`
	// Spec contains the VPC Peering specification
	Properties VPCPeeringPropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type VPCPeeringListResponse struct {
	ListResponse
	Values []VPCPeeringResponse `json:"values"`
}
