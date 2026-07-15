package aruba

import "github.com/Arubacloud/sdk-go/pkg/types"

// ---- Sub-builder ----

// NodePool is a fluent builder for a single KaaS node pool.
// Construct with NewNodePool() and attach via KaaS.AddNodePool.
//
// Schema note: Instance and Zone are plain strings in the Create/Update request
// (wire fields "instance" and "dataCenter"); the response side uses object types
// (KaaSNodePoolInstanceResponse, KaaSNodePoolDataCenterResponse). AddNodePool flattens the response
// representation back to strings so Update round-trips correctly.
type NodePool struct {
	errMixin
	name        *string
	nodes       *int32
	instance    *NodePoolInstance // wire JSON: "instance"
	zone        *Zone             // wire JSON: "dataCenter"
	minCount    *int32
	maxCount    *int32
	autoscaling *bool
}

// NewNodePool returns a fresh *NodePool sub-builder for configuring a KaaS node pool.
func NewNodePool() *NodePool { return &NodePool{} }

// Named sets the node pool name.
func (n *NodePool) Named(name string) *NodePool { n.name = &name; return n }

// Name returns the node pool name, or "" if unset.
func (n *NodePool) Name() string {
	if n.name == nil {
		return ""
	}
	return *n.name
}

// OfInstance sets the machine instance type for this node pool.
func (n *NodePool) OfInstance(instance NodePoolInstance) *NodePool {
	n.instance = &instance
	return n
}

// InZone sets the availability zone (dataCenter) for this node pool.
func (n *NodePool) InZone(zone Zone) *NodePool { n.zone = &zone; return n }

// WithCount sets the fixed number of nodes in this pool. Mutually exclusive with WithAutoscaling.
func (n *NodePool) WithCount(count int) *NodePool { v := int32(count); n.nodes = &v; return n }

// WithAutoscaling enables autoscaling with the given min and max node counts.
func (n *NodePool) WithAutoscaling(min, max int) *NodePool {
	t := true
	mn, mx := int32(min), int32(max)
	n.autoscaling = &t
	n.minCount = &mn
	n.maxCount = &mx
	return n
}

func (n *NodePool) build() types.NodePoolPropertiesRequest {
	out := types.NodePoolPropertiesRequest{}
	if n.name != nil {
		out.Name = *n.name
	}
	if n.nodes != nil {
		out.Nodes = *n.nodes
	}
	if n.instance != nil {
		out.Instance = *n.instance
	}
	if n.zone != nil {
		out.Zone = *n.zone
	}
	if n.minCount != nil {
		out.MinCount = n.minCount
	}
	if n.maxCount != nil {
		out.MaxCount = n.maxCount
	}
	if n.autoscaling != nil {
		out.Autoscaling = *n.autoscaling
	}
	return out
}
