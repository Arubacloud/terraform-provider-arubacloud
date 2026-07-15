package types

import "time"

// ContainerRegistrySizeFlavor is the concurrent-users tier for a container
// registry. Wire-encoded into the "size" JSON field of the request.
// Accepted values per the platform: "Small", "Medium", "HighPerf".
type ContainerRegistrySizeFlavor string

const (
	ContainerRegistrySizeFlavorSmall    ContainerRegistrySizeFlavor = "Small"
	ContainerRegistrySizeFlavorMedium   ContainerRegistrySizeFlavor = "Medium"
	ContainerRegistrySizeFlavorHighPerf ContainerRegistrySizeFlavor = "HighPerf"
)

type UserCredentialCommon struct {
	// Username is the administrator username for the container registry.
	Username string `json:"username"`
}

type ContainerRegistryPropertiesRequest struct {

	// PublicIp is the public IP associated with the container registry
	PublicIp ReferenceResourceCommon `json:"publicIp"`

	VPC ReferenceResourceCommon `json:"vpc"`

	// Subnet is the subnet associated with the container registry
	Subnet ReferenceResourceCommon `json:"subnet"`

	// SecurityGroup is the security group associated with the container registry
	SecurityGroup ReferenceResourceCommon `json:"securityGroup"`

	// BlockStorage is the block storage associated with the container registry
	BlockStorage ReferenceResourceCommon `json:"blockStorage"`

	// BillingPlanCommon is the billing plan for the container registry
	BillingPlanCommon *BillingPlanCommon `json:"billingPlan,omitempty"`

	// AdminUser is the administrator user for the container registry
	AdminUser *UserCredentialCommon `json:"adminUser,omitempty"`

	// Size is the number of concurrent users allowed for the container registry
	ConcurrentUsers *string `json:"size,omitempty"`
}

type ContainerRegistryPropertiesResponse struct {

	// PublicIp is the public IP associated with the container registry
	PublicIp ReferenceResourceCommon `json:"publicIp"`

	// VPC is the VPC associated with the container registry
	VPC ReferenceResourceCommon `json:"vpc"`

	// Subnet is the subnet associated with the container registry
	Subnet ReferenceResourceCommon `json:"subnet"`

	// SecurityGroup is the security group associated with the container registry
	SecurityGroup ReferenceResourceCommon `json:"securityGroup"`

	// BlockStorage is the block storage associated with the container registry
	BlockStorage ReferenceResourceCommon `json:"blockStorage"`

	// BillingPlanCommon is the billing plan for the container registry
	BillingPlanCommon *BillingPlanCommon `json:"billingPlan,omitempty"`

	// AdminUser is the administrator user for the container registry
	AdminUser *UserCredentialCommon `json:"adminUser,omitempty"`

	// Size is the number of concurrent users allowed for the container registry
	ConcurrentUsers *string `json:"size,omitempty"`
}

// ContainerRegistryDataPrivateResponse holds credential-state information returned by
// the platform after the Aruba provisioner has processed the registry. The
// admin password itself is never returned over the wire; only its readiness
// state is exposed here.
type ContainerRegistryDataPrivateResponse struct {
	// PasswordSet reports whether the provisioner has generated the admin password.
	PasswordSet       *bool      `json:"passwordSet,omitempty"`
	PasswordLastSetAt *time.Time `json:"passwordLastSetAt,omitempty"`
}

// ContainerRegistryDataInfoResponse holds operational endpoint information for the registry.
type ContainerRegistryDataInfoResponse struct {
	FQDN           *string `json:"fqdn,omitempty"`
	PublicBaseURL  *string `json:"publicBaseUrl,omitempty"`
	PrivateBaseURL *string `json:"privateBaseUrl,omitempty"`
	Version        *string `json:"version,omitempty"`
}

// ContainerRegistryDataResponse is the top-level data block returned alongside
// metadata/properties/status on Create and Get responses.
type ContainerRegistryDataResponse struct {
	Private *ContainerRegistryDataPrivateResponse `json:"private,omitempty"`
	Info    *ContainerRegistryDataInfoResponse    `json:"info,omitempty"`
}

type ContainerRegistryRequest struct {
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	Properties ContainerRegistryPropertiesRequest `json:"properties"`
}

type ContainerRegistryResponse struct {
	Metadata   ResourceMetadataResponse            `json:"metadata"`
	Properties ContainerRegistryPropertiesResponse `json:"properties"`
	Data       *ContainerRegistryDataResponse      `json:"data,omitempty"`
	Status     ResourceStatusResponse              `json:"status,omitempty"`
}

type ContainerRegistryListResponse struct {
	ListResponse
	Values []ContainerRegistryResponse `json:"values"`
}
