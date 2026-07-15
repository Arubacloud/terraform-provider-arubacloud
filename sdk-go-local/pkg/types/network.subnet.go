package types

// SubnetType represents the type of subnet
type SubnetType string

const (
	SubnetTypeBasic    SubnetType = "Basic"
	SubnetTypeAdvanced SubnetType = "Advanced"
)

// SubnetNetworkCommon contains the network configuration
type SubnetNetworkCommon struct {
	Address string `json:"address"`
}

// SubnetDHCPRangeCommon contains the DHCP range configuration
type SubnetDHCPRangeCommon struct {
	// Start is the starting IP address of the DHCP range
	Start string `json:"start"`
	// Count is the number of IP addresses in the DHCP range
	Count int `json:"count"`
}

// SubnetDHCPRouteCommon contains the DHCP route configuration
type SubnetDHCPRouteCommon struct {
	// Address is the destination network address
	Address string `json:"address"`
	// Gateway is the gateway IP address for the route
	Gateway string `json:"gateway"`
}

// SubnetDHCPCommon contains the DHCP configuration
type SubnetDHCPCommon struct {
	// Enabled indicates if DHCP is enabled
	Enabled bool `json:"enabled"`
	// Range contains the DHCP IP address range
	Range *SubnetDHCPRangeCommon `json:"range,omitempty"`
	// Routes contains the DHCP routes configuration
	Routes []SubnetDHCPRouteCommon `json:"routes,omitempty"`
	// DNS contains the DNS server addresses
	DNS []string `json:"dns,omitempty"`
}

// SubnetPropertiesRequest contains the specification for creating a Subnet
type SubnetPropertiesRequest struct {
	// Type of subnet (Basic or Advanced)
	Type SubnetType `json:"type,omitempty"`

	// Default indicates if the subnet must be a default subnet
	Default *bool `json:"default,omitempty"`

	// Network configuration
	Network *SubnetNetworkCommon `json:"network,omitempty"`

	// DHCP configuration
	DHCP *SubnetDHCPCommon `json:"dhcp,omitempty"`
}

// SubnetPropertiesResponse contains the specification returned for a Subnet
type SubnetPropertiesResponse struct {
	// LinkedResources array of resources linked to the Subnet (nullable)
	LinkedResources []LinkedResourceCommon `json:"linkedResources,omitempty"`

	// Type of subnet
	Type SubnetType `json:"type,omitempty"`

	// Default indicates if the subnet is the default one within the region
	Default bool `json:"default,omitempty"`

	// Network configuration
	Network *SubnetNetworkCommon `json:"network,omitempty"`

	// DHCP configuration
	DHCP *SubnetDHCPCommon `json:"dhcp,omitempty"`
}

type SubnetRequest struct {
	// Metadata of the Subnet
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	// Spec contains the Subnet specification
	Properties SubnetPropertiesRequest `json:"properties"`
}

type SubnetResponse struct {
	// Metadata of the Subnet
	Metadata ResourceMetadataResponse `json:"metadata"`
	// Spec contains the Subnet specification
	Properties SubnetPropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type SubnetListResponse struct {
	ListResponse
	Values []SubnetResponse `json:"values"`
}
