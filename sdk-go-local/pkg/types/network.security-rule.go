package types

import (
	"encoding/json"
	"fmt"
)

// FlexPort unmarshals the security-rule port field from either a plain JSON
// string ("80") or an object ({"value":"80"}) — the Update endpoint returns
// the object form while Get returns a plain string.
type FlexPort string

func (fp *FlexPort) UnmarshalJSON(data []byte) error {
	var s string
	if json.Unmarshal(data, &s) == nil {
		*fp = FlexPort(s)
		return nil
	}
	var obj struct {
		Value string `json:"value"`
	}
	if json.Unmarshal(data, &obj) == nil {
		*fp = FlexPort(obj.Value)
		return nil
	}
	return fmt.Errorf("FlexPort: cannot unmarshal %s: expected a JSON string or {\"value\":\"...\"} object", data)
}

func (fp FlexPort) MarshalJSON() ([]byte, error) { return json.Marshal(string(fp)) }

// String returns the underlying port string value.
func (fp FlexPort) String() string { return string(fp) }

// RuleProtocol identifies the L4 protocol for a security rule.
//
// Authoritative list: ANY, TCP, UDP, ICMP (fully enumerated in the API docs).
type RuleProtocol string

const (
	RuleProtocolANY  RuleProtocol = "ANY"
	RuleProtocolTCP  RuleProtocol = "TCP"
	RuleProtocolUDP  RuleProtocol = "UDP"
	RuleProtocolICMP RuleProtocol = "ICMP"
)

// RuleDirection represents the direction of a security rule
type RuleDirection string

const (
	RuleDirectionIngress RuleDirection = "Ingress"
	RuleDirectionEgress  RuleDirection = "Egress"
)

// EndpointTypeDto represents the type of target endpoint
type EndpointTypeDto string

const (
	EndpointTypeIP            EndpointTypeDto = "Ip"
	EndpointTypeSecurityGroup EndpointTypeDto = "SecurityGroup"
)

// RuleTargetCommon represents the target of the rule (source or destination according to the direction)
type RuleTargetCommon struct {
	// Kind Type of the target. Admissible values: Ip, SecurityGroup
	Kind EndpointTypeDto `json:"kind,omitempty"`

	// Value of the target.
	// If kind = "Ip", the value must be a valid network address in CIDR notation (included 0.0.0.0/0)
	// If kind = "SecurityGroup", the value must be a valid uri of any security group within the same vpc
	Value string `json:"value,omitempty"`
}

// SecurityRuleProperties contains the properties of a security rule
type SecurityRulePropertiesRequest struct {
	// Direction of the rule. Admissible values: Ingress, Egress
	Direction RuleDirection `json:"direction,omitempty"`

	// Protocol Name of the protocol. Admissible values: ANY, TCP, UDP, ICMP
	Protocol RuleProtocol `json:"protocol,omitempty"`

	// Port can be set with different values, according to the protocol.
	// - ANY and ICMP must not have a port
	// - TCP and UDP can have:
	//   - a single numeric port. For instance "80", "443" etc.
	//   - a port range. For instance "80-100"
	//   - the "*" value indicating any ports
	Port string `json:"port,omitempty"`

	// Target The target of the rule (source or destination according to the direction)
	Target *RuleTargetCommon `json:"target,omitempty"`
}

type SecurityRulePropertiesResponse struct {
	LinkedResources []LinkedResourceCommon `json:"linkedResources,omitempty"`

	// Direction of the rule. Admissible values: Ingress, Egress
	Direction RuleDirection `json:"direction,omitempty"`

	// Protocol Name of the protocol. Admissible values: ANY, TCP, UDP, ICMP
	Protocol RuleProtocol `json:"protocol,omitempty"`

	// Port can be set with different values, according to the protocol.
	// - ANY and ICMP must not have a port
	// - TCP and UDP can have:
	//   - a single numeric port. For instance "80", "443" etc.
	//   - a port range. For instance "80-100"
	//   - the "*" value indicating any ports
	// FlexPort handles both the plain-string (Get) and object (Update) response formats.
	Port FlexPort `json:"port,omitempty"`

	// Target The target of the rule (source or destination according to the direction)
	Target *RuleTargetCommon `json:"target,omitempty"`
}

type SecurityRuleRequest struct {
	Metadata RegionalResourceMetadataRequest `json:"metadata"`
	// Properties of the security rule (nullable object)
	Properties SecurityRulePropertiesRequest `json:"properties"`
}

type SecurityRuleResponse struct {
	Metadata   ResourceMetadataResponse       `json:"metadata"`
	Status     ResourceStatusResponse         `json:"status"`
	Properties SecurityRulePropertiesResponse `json:"properties"`
}

type SecurityRuleListResponse struct {
	ListResponse
	Values []SecurityRuleResponse `json:"values"`
}
