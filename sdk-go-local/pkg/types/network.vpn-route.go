package types

import "encoding/json"

type VPNRoutePropertiesRequest struct {

	// CloudSubnet CIDR of the cloud subnet
	CloudSubnet string `json:"cloudSubnet"`

	// OnPremSubnet CIDR of the onPrem subnet
	OnPremSubnet string `json:"onPremSubnet"`
}

// SubnetCIDROrRef decodes cloudSubnet from either a plain CIDR string
// (Get/List responses) or a full subnet resource object (Create response),
// normalising both forms to a plain CIDR string in the CIDR field.
type SubnetCIDROrRef struct {
	CIDR string
}

func (s *SubnetCIDROrRef) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		s.CIDR = str
		return nil
	}
	// Full subnet object shape — try the SubnetInfoCommon shape (cidr), the nested form
	// (properties.network.address), and the flat form (network.address).
	var obj struct {
		CIDR       string `json:"cidr"`
		Properties struct {
			Network struct {
				Address string `json:"address"`
			} `json:"network"`
		} `json:"properties"`
		Network struct {
			Address string `json:"address"`
		} `json:"network"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	if obj.CIDR != "" {
		s.CIDR = obj.CIDR
	} else if obj.Properties.Network.Address != "" {
		s.CIDR = obj.Properties.Network.Address
	} else {
		s.CIDR = obj.Network.Address
	}
	return nil
}

func (s SubnetCIDROrRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.CIDR)
}

type VPNRoutePropertiesResponse struct {
	LinkedResources []LinkedResourceCommon `json:"linkedResources,omitempty"`

	// CloudSubnet CIDR of the cloud subnet (plain string or full subnet object on Create)
	CloudSubnet SubnetCIDROrRef `json:"cloudSubnet"`

	// OnPremSubnet CIDR of the onPrem subnet
	OnPremSubnet string `json:"onPremSubnet"`

	VPNTunnel *ReferenceResourceCommon `json:"vpnTunnel,omitempty"`
}

type VPNRouteRequest struct {
	// Metadata of the VPC Route
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	// Spec contains the VPC Route specification
	Properties VPNRoutePropertiesRequest `json:"properties"`
}

type VPNRouteResponse struct {
	// Metadata of the VPC Route
	Metadata ResourceMetadataResponse `json:"metadata"`
	// Spec contains the VPC Route specification
	Properties VPNRoutePropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type VPNRouteListResponse struct {
	ListResponse
	Values []VPNRouteResponse `json:"values"`
}
