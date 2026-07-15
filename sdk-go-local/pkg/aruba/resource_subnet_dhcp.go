package aruba

import "github.com/Arubacloud/sdk-go/pkg/types"

// ---- Sub-builder ----

// SubnetDHCPCommon is the fluent sub-builder for Properties.DHCP on a Subnet.
// Construct with aruba.NewSubnetDHCP() and attach via (*Subnet).WithDHCP.
type SubnetDHCPCommon struct {
	errMixin
	inner types.SubnetDHCPCommon
}

// NewSubnetDHCP returns an empty (disabled) DHCP block.
func NewSubnetDHCP() *SubnetDHCPCommon { return &SubnetDHCPCommon{} }

// Enabled marks DHCP as enabled.
func (d *SubnetDHCPCommon) Enabled() *SubnetDHCPCommon { d.inner.Enabled = true; return d }

// WithRange sets a single DHCP allocation range.
func (d *SubnetDHCPCommon) WithRange(start string, count int) *SubnetDHCPCommon {
	d.inner.Range = &types.SubnetDHCPRangeCommon{Start: start, Count: count}
	return d
}

// WithRoutes appends static routes advertised via DHCP. Repeated calls append.
func (d *SubnetDHCPCommon) WithRoutes(routes ...SubnetDHCPRouteCommon) *SubnetDHCPCommon {
	for _, r := range routes {
		d.inner.Routes = append(d.inner.Routes, types.SubnetDHCPRouteCommon{Address: r.Address, Gateway: r.Gateway})
	}
	return d
}

// WithDNSServers appends DNS servers advertised via DHCP. Repeated calls append.
func (d *SubnetDHCPCommon) WithDNSServers(ips ...string) *SubnetDHCPCommon {
	d.inner.DNS = append(d.inner.DNS, ips...)
	return d
}

// IsEnabled reports whether DHCP is enabled on this subnet.
func (d *SubnetDHCPCommon) IsEnabled() bool { return d.inner.Enabled }

// RangeStart returns the start address of the DHCP allocation range, or "" if unset.
func (d *SubnetDHCPCommon) RangeStart() string {
	if d.inner.Range == nil {
		return ""
	}
	return d.inner.Range.Start
}

// RangeCount returns the number of addresses in the DHCP allocation range, or 0 if unset.
func (d *SubnetDHCPCommon) RangeCount() int {
	if d.inner.Range == nil {
		return 0
	}
	return d.inner.Range.Count
}

// Routes returns a copy of the static routes advertised via DHCP, or nil if none.
func (d *SubnetDHCPCommon) Routes() []SubnetDHCPRouteCommon {
	if len(d.inner.Routes) == 0 {
		return nil
	}
	out := make([]SubnetDHCPRouteCommon, len(d.inner.Routes))
	for i, r := range d.inner.Routes {
		out[i] = SubnetDHCPRouteCommon{Address: r.Address, Gateway: r.Gateway}
	}
	return out
}

// DNS returns a copy of the DNS server list advertised via DHCP, or nil if none.
func (d *SubnetDHCPCommon) DNS() []string {
	if len(d.inner.DNS) == 0 {
		return nil
	}
	return append([]string(nil), d.inner.DNS...)
}

// SubnetDHCPRouteCommon mirrors types.SubnetDHCPRouteCommon at the wrapper boundary.
type SubnetDHCPRouteCommon struct {
	Address string
	Gateway string
}

func (d *SubnetDHCPCommon) build() *types.SubnetDHCPCommon {
	if d == nil {
		return nil
	}
	cp := d.inner
	if len(d.inner.Routes) > 0 {
		cp.Routes = append([]types.SubnetDHCPRouteCommon(nil), d.inner.Routes...)
	}
	if len(d.inner.DNS) > 0 {
		cp.DNS = append([]string(nil), d.inner.DNS...)
	}
	if d.inner.Range != nil {
		r := *d.inner.Range
		cp.Range = &r
	}
	return &cp
}

func dhcpFromType(t *types.SubnetDHCPCommon) *SubnetDHCPCommon {
	if t == nil {
		return nil
	}
	d := &SubnetDHCPCommon{inner: types.SubnetDHCPCommon{Enabled: t.Enabled}}
	if t.Range != nil {
		r := *t.Range
		d.inner.Range = &r
	}
	if len(t.Routes) > 0 {
		d.inner.Routes = append([]types.SubnetDHCPRouteCommon(nil), t.Routes...)
	}
	if len(t.DNS) > 0 {
		d.inner.DNS = append([]string(nil), t.DNS...)
	}
	return d
}
