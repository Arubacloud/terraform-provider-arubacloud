package aruba

import "context"

type NetworkClient interface {
	ElasticIPs() ElasticIPsClient
	LoadBalancers() LoadBalancersClient
	SecurityGroupRules() SecurityGroupRulesClient
	SecurityGroups() SecurityGroupsClient
	Subnets() SubnetsClient
	VPCPeeringRoutes() VPCPeeringRoutesClient
	VPCPeerings() VPCPeeringsClient
	VPCs() VPCsClient
	VPNRoutes() VPNRoutesClient
	VPNTunnels() VPNTunnelsClient
}

type networkClientImpl struct {
	elasticIPsClient         ElasticIPsClient
	loadBalancersClient      LoadBalancersClient
	securityGroupRulesClient SecurityGroupRulesClient
	securityGroupsClient     SecurityGroupsClient
	subnetsClient            SubnetsClient
	vpcPeeringRoutesClient   VPCPeeringRoutesClient
	vpcPeeringsClient        VPCPeeringsClient
	vpcsClient               VPCsClient
	vpnRoutesClient          VPNRoutesClient
	vpnTunnelsClient         VPNTunnelsClient
}

var _ NetworkClient = (*networkClientImpl)(nil)

func (c *networkClientImpl) ElasticIPs() ElasticIPsClient {
	return c.elasticIPsClient
}
func (c *networkClientImpl) LoadBalancers() LoadBalancersClient {
	return c.loadBalancersClient
}
func (c *networkClientImpl) SecurityGroupRules() SecurityGroupRulesClient {
	return c.securityGroupRulesClient
}
func (c *networkClientImpl) SecurityGroups() SecurityGroupsClient {
	return c.securityGroupsClient
}
func (c *networkClientImpl) Subnets() SubnetsClient {
	return c.subnetsClient
}
func (c *networkClientImpl) VPCPeeringRoutes() VPCPeeringRoutesClient {
	return c.vpcPeeringRoutesClient
}
func (c *networkClientImpl) VPCPeerings() VPCPeeringsClient {
	return c.vpcPeeringsClient
}
func (c *networkClientImpl) VPCs() VPCsClient {
	return c.vpcsClient
}
func (c *networkClientImpl) VPNRoutes() VPNRoutesClient {
	return c.vpnRoutesClient
}
func (c *networkClientImpl) VPNTunnels() VPNTunnelsClient {
	return c.vpnTunnelsClient
}

type ElasticIPsClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*ElasticIP], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*ElasticIP, error)
	Create(ctx context.Context, eip *ElasticIP, opts ...CallOption) (*ElasticIP, error)
	Update(ctx context.Context, eip *ElasticIP, opts ...CallOption) (*ElasticIP, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type LoadBalancersClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*LoadBalancer], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*LoadBalancer, error)
}

type SecurityGroupRulesClient interface {
	List(ctx context.Context, securityGroup Ref, opts ...CallOption) (*List[*SecurityRule], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*SecurityRule, error)
	Create(ctx context.Context, rule *SecurityRule, opts ...CallOption) (*SecurityRule, error)
	Update(ctx context.Context, rule *SecurityRule, opts ...CallOption) (*SecurityRule, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type SecurityGroupsClient interface {
	List(ctx context.Context, vpc Ref, opts ...CallOption) (*List[*SecurityGroup], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*SecurityGroup, error)
	Create(ctx context.Context, sg *SecurityGroup, opts ...CallOption) (*SecurityGroup, error)
	Update(ctx context.Context, sg *SecurityGroup, opts ...CallOption) (*SecurityGroup, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type SubnetsClient interface {
	List(ctx context.Context, vpc Ref, opts ...CallOption) (*List[*Subnet], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*Subnet, error)
	Create(ctx context.Context, subnet *Subnet, opts ...CallOption) (*Subnet, error)
	Update(ctx context.Context, subnet *Subnet, opts ...CallOption) (*Subnet, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type VPCPeeringRoutesClient interface {
	List(ctx context.Context, peering Ref, opts ...CallOption) (*List[*VPCPeeringRoute], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*VPCPeeringRoute, error)
	Create(ctx context.Context, route *VPCPeeringRoute, opts ...CallOption) (*VPCPeeringRoute, error)
	Update(ctx context.Context, route *VPCPeeringRoute, opts ...CallOption) (*VPCPeeringRoute, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type VPCPeeringsClient interface {
	List(ctx context.Context, vpc Ref, opts ...CallOption) (*List[*VPCPeering], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*VPCPeering, error)
	Create(ctx context.Context, peering *VPCPeering, opts ...CallOption) (*VPCPeering, error)
	Update(ctx context.Context, peering *VPCPeering, opts ...CallOption) (*VPCPeering, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type VPCsClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*VPC], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*VPC, error)
	Create(ctx context.Context, vpc *VPC, opts ...CallOption) (*VPC, error)
	Update(ctx context.Context, vpc *VPC, opts ...CallOption) (*VPC, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type VPNRoutesClient interface {
	List(ctx context.Context, tunnel Ref, opts ...CallOption) (*List[*VPNRoute], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*VPNRoute, error)
	Create(ctx context.Context, r *VPNRoute, opts ...CallOption) (*VPNRoute, error)
	Update(ctx context.Context, r *VPNRoute, opts ...CallOption) (*VPNRoute, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type VPNTunnelsClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*VPNTunnel], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*VPNTunnel, error)
	Create(ctx context.Context, t *VPNTunnel, opts ...CallOption) (*VPNTunnel, error)
	Update(ctx context.Context, t *VPNTunnel, opts ...CallOption) (*VPNTunnel, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}
