package network

// Path constants follow the server-canonical lowerCamelCase rule:
//   - Single-word / acronym collections stay lowercase: vpcs, subnets.
//   - Compound collections use lowerCamelCase: securityGroups, securityRules,
//     elasticIps, loadBalancers, vpcPeerings, vpcPeeringRoutes, vpnTunnels, vpnRoutes.
//
// Do not flatten these to all-lowercase. Downstream provisioners store and re-emit
// request URIs verbatim, so a casing change causes silent provisioning failures.
// Verified via examples/all-resources/ create.log (2026-05-28, commit f548a4f alignment).
const (
	// VPC Network paths
	VPCNetworksPath = "/projects/%s/providers/Aruba.Network/vpcs"
	VPCNetworkPath  = "/projects/%s/providers/Aruba.Network/vpcs/%s"

	// Subnet paths (nested under VPC)
	SubnetsPath = "/projects/%s/providers/Aruba.Network/vpcs/%s/subnets"
	SubnetPath  = "/projects/%s/providers/Aruba.Network/vpcs/%s/subnets/%s"

	// Security Group paths (nested under VPC)
	SecurityGroupsPath = "/projects/%s/providers/Aruba.Network/vpcs/%s/securityGroups"
	SecurityGroupPath  = "/projects/%s/providers/Aruba.Network/vpcs/%s/securityGroups/%s"

	// Security Group Rule paths (nested under VPC and Security Group)
	SecurityGroupRulesPath = "/projects/%s/providers/Aruba.Network/vpcs/%s/securityGroups/%s/securityRules"
	SecurityGroupRulePath  = "/projects/%s/providers/Aruba.Network/vpcs/%s/securityGroups/%s/securityRules/%s"

	// Elastic IP paths
	ElasticIPsPath = "/projects/%s/providers/Aruba.Network/elasticIps"
	ElasticIPPath  = "/projects/%s/providers/Aruba.Network/elasticIps/%s"

	// Load Balancer paths
	LoadBalancersPath = "/projects/%s/providers/Aruba.Network/loadBalancers"
	LoadBalancerPath  = "/projects/%s/providers/Aruba.Network/loadBalancers/%s"

	// VPC Peering Connection paths
	VPCPeeringsPath = "/projects/%s/providers/Aruba.Network/vpcs/%s/vpcPeerings"
	VPCPeeringPath  = "/projects/%s/providers/Aruba.Network/vpcs/%s/vpcPeerings/%s"

	// VPC Peering Route paths
	VPCPeeringRoutesPath = "/projects/%s/providers/Aruba.Network/vpcs/%s/vpcPeerings/%s/vpcPeeringRoutes"
	VPCPeeringRoutePath  = "/projects/%s/providers/Aruba.Network/vpcs/%s/vpcPeerings/%s/vpcPeeringRoutes/%s"

	// VPN Tunnel paths
	VPNTunnelsPath = "/projects/%s/providers/Aruba.Network/vpnTunnels"
	VPNTunnelPath  = "/projects/%s/providers/Aruba.Network/vpnTunnels/%s"

	// VPN Route paths
	VPNRoutesPath = "/projects/%s/providers/Aruba.Network/vpnTunnels/%s/vpnRoutes"
	VPNRoutePath  = "/projects/%s/providers/Aruba.Network/vpnTunnels/%s/vpnRoutes/%s"
)
