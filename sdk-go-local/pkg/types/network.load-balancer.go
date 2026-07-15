package types

type LoadBalancerPropertiesResponse struct {
	// LinkedResources array of resources linked to the Load Balancer (nullable)
	LinkedResources []LinkedResourceCommon   `json:"linkedResources,omitempty"`
	Address         *string                  `json:"address,omitempty"`
	VPC             *ReferenceResourceCommon `json:"vpc,omitempty"`
}

type LoadBalancerResponse struct {
	// Metadata of the Load Balancer
	Metadata ResourceMetadataResponse `json:"metadata"`
	// Spec contains the Load Balancer specification
	Properties LoadBalancerPropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type LoadBalancerListResponse struct {
	ListResponse
	Values []LoadBalancerResponse `json:"values"`
}
