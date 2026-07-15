package types

// VPCPropertiesInnerRequest contains the properties of a VPC
type VPCPropertiesInnerRequest struct {
	// Default indicates if the vpc must be a default vpc. Only one default vpc for region is admissible.
	// Default value: true
	Default *bool `json:"default,omitempty"`

	// Preset indicates if a subnet and a securityGroup with default configuration will be created automatically within the vpc
	// Default value: false
	Preset *bool `json:"preset,omitempty"`
}

// VPCPropertiesRequest contains the specification for creating a VPC
type VPCPropertiesRequest struct {
	// Properties of the vpc (nullable object)
	Properties *VPCPropertiesInnerRequest `json:"properties,omitempty"`
}

// VPCPropertiesResponse contains the specification returned for a VPC
type VPCPropertiesResponse struct {
	// LinkedResources array of resources linked to the VPC (nullable)
	LinkedResources []LinkedResourceCommon `json:"linkedResources,omitempty"`

	// Default indicates if the vpc is the default one within the region
	Default bool `json:"default,omitempty"`
}

type VPCRequest struct {
	// Metadata of the VPC
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	// Spec contains the VPC specification
	Properties VPCPropertiesRequest `json:"properties"`
}

type VPCResponse struct {
	// Metadata of the VPC
	Metadata ResourceMetadataResponse `json:"metadata"`
	// Spec contains the VPC specification
	Properties VPCPropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type VPCListResponse struct {
	ListResponse
	Values []VPCResponse `json:"values"`
}
