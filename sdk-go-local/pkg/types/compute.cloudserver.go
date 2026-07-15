package types

// CloudServerFlavor identifies a Cloud Server SKU.
//
// Pattern: CSO<vCPU>A<RAM>. The constants below cover SKUs referenced in
// SDK fixtures. The authoritative list is available via:
//
//	GET /providers/Aruba.Compute/flavors
type CloudServerFlavor string

const (
	CloudServerFlavorCSO1A2   CloudServerFlavor = "CSO1A2"
	CloudServerFlavorCSO1A4   CloudServerFlavor = "CSO1A4"
	CloudServerFlavorCSO2A4   CloudServerFlavor = "CSO2A4"
	CloudServerFlavorCSO2A8   CloudServerFlavor = "CSO2A8"
	CloudServerFlavorCSO4A8   CloudServerFlavor = "CSO4A8"
	CloudServerFlavorCSO4A16  CloudServerFlavor = "CSO4A16"
	CloudServerFlavorCSO8A16  CloudServerFlavor = "CSO8A16"
	CloudServerFlavorCSO8A32  CloudServerFlavor = "CSO8A32"
	CloudServerFlavorCSO12A24 CloudServerFlavor = "CSO12A24"
	CloudServerFlavorCSO16A32 CloudServerFlavor = "CSO16A32"
	CloudServerFlavorCSO16A64 CloudServerFlavor = "CSO16A64"
	CloudServerFlavorCSO24A48 CloudServerFlavor = "CSO24A48"
	CloudServerFlavorCSO32A64 CloudServerFlavor = "CSO32A64"
)

type CloudServerPropertiesRequest struct {
	Zone Zone `json:"dataCenter"`

	VPC ReferenceResourceCommon `json:"vpc"`

	VPCPreset bool `json:"vpcPreset,omitempty"`

	FlavorName *CloudServerFlavor `json:"flavorName,omitempty"`

	ElasticIP *ReferenceResourceCommon `json:"elasticIp,omitempty"`

	BootVolume *ReferenceResourceCommon `json:"bootVolume,omitempty"`

	KeyPair *ReferenceResourceCommon `json:"keyPair,omitempty"`

	Subnets []ReferenceResourceCommon `json:"subnets,omitempty"`

	SecurityGroups []ReferenceResourceCommon `json:"securityGroups,omitempty"`

	UserData *string `json:"userData,omitempty"`

	BillingPlanCommon *BillingPlanCommon `json:"billingPlan,omitempty"`
}

type CloudServerFlavorResponse struct {
	ID string `json:"id"`

	Name CloudServerFlavor `json:"name"`

	Category string `json:"category"`

	CPU int32 `json:"cpu"`

	RAM int32 `json:"ram"`

	HD int32 `json:"hd"`
}

type CloudServerNetworkInterfaceResponse struct {
	Subnet *string `json:"subnet,omitempty"`

	MacAddress *string `json:"macAddress,omitempty"`

	IPs []string `json:"ips,omitempty"`
}

type CloudServerPropertiesResponse struct {
	LinkedResources []LinkedResourceCommon `json:"linkedResources,omitempty"`

	Zone Zone `json:"dataCenter"`

	VPC ReferenceResourceCommon `json:"vpc"`

	Flavor CloudServerFlavorResponse `json:"flavor,omitempty"`

	Template ReferenceResourceCommon `json:"template"`

	BootVolume ReferenceResourceCommon `json:"bootVolume"`

	KeyPair ReferenceResourceCommon `json:"keyPair"`

	NetworkInterfaces []CloudServerNetworkInterfaceResponse `json:"networkInterfaces,omitempty"`

	BillingPlanCommon *BillingPlanCommon `json:"billingPlan,omitempty"`
}

type CloudServerRequest struct {
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	Properties CloudServerPropertiesRequest `json:"properties"`
}

type CloudServerResponse struct {
	Metadata   ResourceMetadataResponse      `json:"metadata"`
	Properties CloudServerPropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type CloudServerListResponse struct {
	ListResponse
	Values []CloudServerResponse `json:"values"`
}

type CloudServerPasswordRequest struct {
	Password string `json:"password"`
}

type CloudServerAssociateSubnetsRequest struct {
	SubnetsToAssociate    []ReferenceResourceCommon `json:"subnetsToAssociate,omitempty"`
	SubnetsToDisassociate []ReferenceResourceCommon `json:"subnetsToDisassociate,omitempty"`
}

type CloudServerAssociateSecurityGroupsRequest struct {
	SecurityGroupsToAssociate    []ReferenceResourceCommon `json:"securityGroupsToAssociate,omitempty"`
	SecurityGroupsToDisassociate []ReferenceResourceCommon `json:"securityGroupsToDisassociate,omitempty"`
}

type CloudServerAssociateElasticIPsRequest struct {
	ElasticIPsToAssociate    []ReferenceResourceCommon `json:"elasticIPsToAssociate,omitempty"`
	ElasticIPsToDisassociate []ReferenceResourceCommon `json:"elasticIPsToDisassociate,omitempty"`
}

type CloudServerAttachDetachDataVolumesRequest struct {
	VolumesToAttach []ReferenceResourceCommon `json:"volumesToAttach,omitempty"`
	VolumesToDetach []ReferenceResourceCommon `json:"volumesToDetach,omitempty"`
}
