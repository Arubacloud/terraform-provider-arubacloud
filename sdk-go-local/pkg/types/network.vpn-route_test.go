package types_test

import (
	"encoding/json"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestSubnetCIDROrRef_UnmarshalJSON_PlainString(t *testing.T) {
	var s types.SubnetCIDROrRef
	if err := json.Unmarshal([]byte(`"10.0.0.0/24"`), &s); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if s.CIDR != "10.0.0.0/24" {
		t.Errorf("CIDR = %q, want %q", s.CIDR, "10.0.0.0/24")
	}
}

func TestSubnetCIDROrRef_UnmarshalJSON_FullObject(t *testing.T) {
	raw := `{"metadata":{"id":"sub-1","name":"my-subnet"},"properties":{"network":{"address":"10.1.0.0/24"}}}`
	var s types.SubnetCIDROrRef
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if s.CIDR != "10.1.0.0/24" {
		t.Errorf("CIDR = %q, want %q", s.CIDR, "10.1.0.0/24")
	}
}

func TestSubnetCIDROrRef_UnmarshalJSON_FlatNetworkObject(t *testing.T) {
	// GET response shape: {network:{address:"..."}} without properties wrapper
	raw := `{"name":"my-subnet","network":{"address":"10.2.0.0/24"}}`
	var s types.SubnetCIDROrRef
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if s.CIDR != "10.2.0.0/24" {
		t.Errorf("CIDR = %q, want %q", s.CIDR, "10.2.0.0/24")
	}
}

func TestSubnetCIDROrRef_UnmarshalJSON_SubnetInfoShape(t *testing.T) {
	// SubnetInfoCommon shape: {"cidr":"...","name":"..."} returned by some GET/List responses
	raw := `{"cidr":"10.3.0.0/24","name":"my-subnet"}`
	var s types.SubnetCIDROrRef
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if s.CIDR != "10.3.0.0/24" {
		t.Errorf("CIDR = %q, want %q", s.CIDR, "10.3.0.0/24")
	}
}

func TestSubnetCIDROrRef_UnmarshalJSON_EmptyString(t *testing.T) {
	var s types.SubnetCIDROrRef
	if err := json.Unmarshal([]byte(`""`), &s); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if s.CIDR != "" {
		t.Errorf("CIDR = %q, want empty string", s.CIDR)
	}
}

func TestSubnetCIDROrRef_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var s types.SubnetCIDROrRef
	if err := json.Unmarshal([]byte(`not-json`), &s); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestSubnetCIDROrRef_MarshalJSON(t *testing.T) {
	s := types.SubnetCIDROrRef{CIDR: "192.168.0.0/16"}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	if string(data) != `"192.168.0.0/16"` {
		t.Errorf("MarshalJSON = %s, want %q", data, `"192.168.0.0/16"`)
	}
}

func TestSubnetCIDROrRef_RoundTrip_ViaPropertiesResponse(t *testing.T) {
	// Verify that a VPNRoutePropertiesResponse with a plain-string cloudSubnet
	// marshals and unmarshals symmetrically.
	original := types.VPNRoutePropertiesResponse{
		CloudSubnet:  types.SubnetCIDROrRef{CIDR: "10.0.0.0/24"},
		OnPremSubnet: "192.168.0.0/24",
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var decoded types.VPNRoutePropertiesResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if decoded.CloudSubnet.CIDR != original.CloudSubnet.CIDR {
		t.Errorf("CloudSubnet.CIDR = %q, want %q", decoded.CloudSubnet.CIDR, original.CloudSubnet.CIDR)
	}
	if decoded.OnPremSubnet != original.OnPremSubnet {
		t.Errorf("OnPremSubnet = %q, want %q", decoded.OnPremSubnet, original.OnPremSubnet)
	}
}
