package types

// VPCPeeringRoutePropertiesRequest contains properties of a VPC peering route to create
type VPCPeeringRoutePropertiesRequest struct {
	// LocalNetworkAddress Local network address in CIDR notation
	LocalNetworkAddress string `json:"localNetworkAddress"`

	// RemoteNetworkAddress Remote network address in CIDR notation
	RemoteNetworkAddress string `json:"remoteNetworkAddress"`

	BillingPlanCommon *BillingPlanCommon `json:"billingPlan,omitempty"`
}

type VPCPeeringRoutePropertiesResponse struct {
	// LocalNetworkAddress Local network address in CIDR notation
	LocalNetworkAddress string `json:"localNetworkAddress"`

	// RemoteNetworkAddress Remote network address in CIDR notation
	RemoteNetworkAddress string `json:"remoteNetworkAddress"`

	BillingPlanCommon *BillingPlanCommon `json:"billingPlan,omitempty"`
}

type VPCPeeringRouteRequest struct {
	// Metadata of the VPC Peering Route
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	// Spec contains the VPC Peering Route specification
	Properties VPCPeeringRoutePropertiesRequest `json:"properties"`
}

type VPCPeeringRouteResponse struct {
	// Metadata of the VPC Peering Route
	Metadata ResourceMetadataResponse `json:"metadata"`
	// Spec contains the VPC Peering Route specification
	Properties VPCPeeringRoutePropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type VPCPeeringRouteListResponse struct {
	ListResponse
	Values []VPCPeeringRouteResponse `json:"values"`
}
