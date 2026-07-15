package types

type SecurityGroupPropertiesRequest struct {
	// Indicates if the security group must be a default subnet. Only one default security group for vpc is admissible.
	Default *bool `json:"default,omitempty"`
}
type SecurityGroupPropertiesResponse struct {
	LinkedResources []LinkedResourceCommon `json:"linkedResources,omitempty"`

	Default bool `json:"default,omitempty"`
}

type SecurityGroupRequest struct {
	// Metadata of the Security Group
	Metadata ResourceMetadataRequest `json:"metadata"`

	// Spec contains the Security Group specification
	Properties SecurityGroupPropertiesRequest `json:"properties"`
}

type SecurityGroupResponse struct {
	// Metadata of the Security Group
	Metadata ResourceMetadataResponse `json:"metadata"`
	// Spec contains the Security Group specification
	Properties SecurityGroupPropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type SecurityGroupListResponse struct {
	ListResponse
	Values []SecurityGroupResponse `json:"values"`
}
