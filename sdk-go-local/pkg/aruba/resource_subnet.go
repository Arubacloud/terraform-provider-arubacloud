package aruba

import (
	"context"
	"fmt"

	"k8s.io/utils/ptr"

	"github.com/Arubacloud/sdk-go/internal/clients/network"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// SubnetRef returns a Ref that points to the Subnet nested under a VPC.
func SubnetRef(projectID, vpcID, subnetID string) Ref {
	return URI(fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpcs/%s/subnets/%s", projectID, vpcID, subnetID))
}

// ---- Wrapper ----

// Subnet is the wrapper for an Aruba Cloud subnet (a child of a VPC).
// Construct with aruba.NewSubnet() and bind it via IntoVPC(vpc).
//
// Wraps types.SubnetResponse / types.SubnetRequest. The wrapper carries
// pointer-typed private fields so unset values round-trip through
// the JSON layer correctly.
type Subnet struct {
	errMixin
	metadataMixin
	regionalMixin
	vpcScopedMixin // direct parent; populates vpcID + projectID
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	subnetType    *SubnetType           // Properties.Type ("Basic" / "Advanced")
	defaultSubnet *bool                 // Properties.Default
	cidr          *string               // Properties.Network.Address
	dhcp          *SubnetDHCPCommon     // Properties.DHCP (sub-builder)
	response      *types.SubnetResponse // backs Raw()
}

// NewSubnet returns a fresh *Subnet ready for fluent setters and a Create call.
// Binds vpcScopedMixin's error sink so IntoVPC failures surface via Err().
func NewSubnet() *Subnet {
	s := &Subnet{defaultSubnet: ptr.To(false)}
	s.vpcScopedMixin = bindVPCScoped(&s.errMixin)
	return s
}

// Setters — chainable, general → specific

// InVPC binds this Subnet to its parent VPC. Required before Create.
func (s *Subnet) InVPC(v Ref) *Subnet { s.intoVPC(v); return s }

// Named sets the resource name. Required by the API.
func (s *Subnet) Named(n string) *Subnet { s.named(n); return s }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (s *Subnet) Tagged(ts ...string) *Subnet {
	for _, t := range ts {
		s.addTag(t)
	}
	return s
}

// Untagged removes each listed tag. No-op for tags not present.
func (s *Subnet) Untagged(ts ...string) *Subnet {
	for _, t := range ts {
		s.removeTag(t)
	}
	return s
}

// RetaggedAs replaces the entire tag set with the given values.
func (s *Subnet) RetaggedAs(ts ...string) *Subnet { s.replaceTags(ts...); return s }

// InRegion sets the region for this resource.
func (s *Subnet) InRegion(region Region) *Subnet { s.inRegion(region); return s }

// OfType sets the subnet type (Basic or Advanced).
func (s *Subnet) OfType(t SubnetType) *Subnet { s.subnetType = &t; return s }

// AsDefault marks this subnet as the VPC default.
func (s *Subnet) AsDefault() *Subnet { t := true; s.defaultSubnet = &t; return s }

// NotDefault explicitly unsets the default flag.
func (s *Subnet) NotDefault() *Subnet { f := false; s.defaultSubnet = &f; return s }

// WithCIDR sets the subnet network address (CIDR notation).
func (s *Subnet) WithCIDR(cidr string) *Subnet { s.cidr = &cidr; return s }

// WithDHCP attaches a DHCP configuration sub-builder to the subnet.
func (s *Subnet) WithDHCP(d *SubnetDHCPCommon) *Subnet { s.dhcp = d; return s }

// Getters — general → specific

// URI satisfies Ref.
func (s *Subnet) URI() string { return s.RespURI() }

// SubnetID satisfies withSubnetID so future grandchildren can extract this ID.
func (s *Subnet) SubnetID() string { return s.ID() }

// Raw shadows responseMetadataMixin.Raw() with the full subnet response.
func (s *Subnet) Raw() *types.SubnetResponse { return s.response }
func (s *Subnet) RawJSON() []byte            { return marshalRawJSON(s.response) }
func (s *Subnet) RawYAML() []byte            { return marshalRawYAML(s.response) }

// RawRequest returns what toRequest() would emit right now.
func (s *Subnet) RawRequest() types.SubnetRequest { return s.toRequest() }

// Type returns the subnet type (Basic or Advanced), or "" if unset.
func (s *Subnet) Type() SubnetType {
	if s.subnetType == nil {
		return ""
	}
	return *s.subnetType
}

// IsDefault reports whether this subnet is marked as the VPC default.
func (s *Subnet) IsDefault() bool {
	if s.defaultSubnet == nil {
		return false
	}
	return *s.defaultSubnet
}

// CIDR returns the subnet network address in CIDR notation, or "" if unset.
func (s *Subnet) CIDR() string {
	if s.cidr == nil {
		return ""
	}
	return *s.cidr
}

// Network returns the subnet network address in CIDR notation.
// It is an alias for CIDR() and returns "" if the address is unset.
func (s *Subnet) Network() string { return s.CIDR() }

// DHCP returns the attached DHCP configuration sub-builder, or nil if not set.
func (s *Subnet) DHCP() *SubnetDHCPCommon { return s.dhcp }

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (s *Subnet) toRequest() types.SubnetRequest {
	props := types.SubnetPropertiesRequest{}
	if s.subnetType != nil {
		props.Type = *s.subnetType
	}
	if s.defaultSubnet != nil {
		props.Default = s.defaultSubnet
	}
	if s.cidr != nil {
		props.Network = &types.SubnetNetworkCommon{Address: *s.cidr}
	}
	if s.dhcp != nil {
		props.DHCP = s.dhcp.build()
	}
	return types.SubnetRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: s.toMetadata(),
			Location:                s.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (s *Subnet) fromResponse(resp *types.SubnetResponse) {
	if resp == nil {
		return
	}
	s.response = resp
	s.setMeta(&resp.Metadata)
	s.named(subnetDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		s.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		s.inRegion(resp.Metadata.LocationResponse.Value)
	}
	s.setStatus(&resp.Status)

	s.setLinked(resp.Properties.LinkedResources)

	if resp.Properties.Type != "" {
		t := resp.Properties.Type
		s.subnetType = &t
	}
	d := resp.Properties.Default
	s.defaultSubnet = &d
	if resp.Properties.Network != nil && resp.Properties.Network.Address != "" {
		addr := resp.Properties.Network.Address
		s.cidr = &addr
	}
	if resp.Properties.DHCP != nil {
		s.dhcp = dhcpFromType(resp.Properties.DHCP)
	}

	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		s.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	// Backfill ancestor IDs from response URI if not already set.
	if (s.vpcID == "" || s.projectID == "") && s.RespURI() != "" {
		ids := parseURIIDs(s.RespURI())
		if s.vpcID == "" {
			s.vpcID = ids["vpcs"]
		}
		if s.projectID == "" {
			s.projectID = ids["projects"]
		}
	}
}

func subnetDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// subnetLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type subnetLowLevelClient interface {
	List(ctx context.Context, projectID, vpcID string, params *types.RequestParameters) (*types.Response[types.SubnetListResponse], error)
	Get(ctx context.Context, projectID, vpcID, subnetID string, params *types.RequestParameters) (*types.Response[types.SubnetResponse], error)
	Create(ctx context.Context, projectID, vpcID string, body types.SubnetRequest, params *types.RequestParameters) (*types.Response[types.SubnetResponse], error)
	Update(ctx context.Context, projectID, vpcID, subnetID string, body types.SubnetRequest, params *types.RequestParameters) (*types.Response[types.SubnetResponse], error)
	Delete(ctx context.Context, projectID, vpcID, subnetID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// subnetsClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates Subnet ↔ types.SubnetRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type subnetsClientAdapter struct {
	low  subnetLowLevelClient
	rest *restclient.Client
}

var _ SubnetsClient = (*subnetsClientAdapter)(nil)

func newSubnetsClientAdapter(rest *restclient.Client) *subnetsClientAdapter {
	if rest == nil {
		return &subnetsClientAdapter{}
	}
	return &subnetsClientAdapter{
		low:  network.NewSubnetsClientImpl(rest, network.NewVPCsClientImpl(rest)),
		rest: rest,
	}
}

// Create posts a new Subnet to the API and hydrates the wrapper from the response.
func (a *subnetsClientAdapter) Create(ctx context.Context, s *Subnet, opts ...CallOption) (*Subnet, error) {
	if err := s.Err(); err != nil {
		return s, err
	}
	if s.VPCID() == "" || s.ProjectID() == "" {
		return s, fmt.Errorf("Create: subnet has no VPC — call InVPC first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, s.ProjectID(), s.VPCID(), s.toRequest(), rp)
	populateHTTPEnvelope(&s.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		s.fromResponse(resp.Data)
		s.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, s)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				s.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return s, err
	}
	if resp != nil && !resp.IsSuccess() {
		return s, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return s, nil
}

// Get fetches a Subnet by Ref and returns a freshly hydrated wrapper.
func (a *subnetsClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*Subnet, error) {
	projectID, vpcID, subnetID, err := subnetIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, vpcID, subnetID, rp)
	out := &Subnet{}
	populateHTTPEnvelope(&out.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		out.fromResponse(resp.Data)
		out.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, out)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				out.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	out.vpcID = vpcID
	out.projectID = projectID
	if err != nil {
		return out, err
	}
	if resp != nil && !resp.IsSuccess() {
		return out, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return out, nil
}

// Update sends a PUT for the current wrapper state. Requires ID and parent.
func (a *subnetsClientAdapter) Update(ctx context.Context, s *Subnet, opts ...CallOption) (*Subnet, error) {
	if err := s.Err(); err != nil {
		return s, err
	}
	if s.ID() == "" {
		return s, fmt.Errorf("Update: subnet has no ID — call Get first or seed from response metadata")
	}
	if s.VPCID() == "" || s.ProjectID() == "" {
		return s, fmt.Errorf("Update: subnet has no VPC — call InVPC first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, s.ProjectID(), s.VPCID(), s.ID(), s.toRequest(), rp)
	populateHTTPEnvelope(&s.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		s.fromResponse(resp.Data)
		s.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, s)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				s.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return s, err
	}
	if resp != nil && !resp.IsSuccess() {
		return s, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return s, nil
}

// Delete removes the Subnet identified by Ref.
func (a *subnetsClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, vpcID, subnetID, err := subnetIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, vpcID, subnetID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of Subnet in the given parent scope.
func (a *subnetsClientAdapter) List(ctx context.Context, vpc Ref, opts ...CallOption) (*List[*Subnet], error) {
	projectID, vpcID, err := vpcIDsFromRef(vpc)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.List(ctx, projectID, vpcID, rp)
	if err != nil {
		return nil, err
	}
	if resp != nil && !resp.IsSuccess() {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	var items []*Subnet
	if resp != nil && resp.Data != nil {
		items = make([]*Subnet, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			s := &Subnet{}
			s.fromResponse(&resp.Data.Values[i])
			s.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, s)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					s.fromResponse(fresh.Raw())
				}
				return nil
			})
			s.vpcID = vpcID
			if s.projectID == "" {
				s.projectID = projectID
			}
			items = append(items, s)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*Subnet], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*Subnet], error) {
		fetch := listPageFetch[types.SubnetListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*Subnet
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*Subnet, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				item := &Subnet{}
				item.fromResponse(&pageResp.Data.Values[i])
				item.setRefresh(func(ctx context.Context) error {
					fresh, err := a.Get(ctx, item)
					if err != nil {
						return err
					}
					if fresh != nil && fresh.Raw() != nil {
						item.fromResponse(fresh.Raw())
					}
					return nil
				})
				item.vpcID = vpcID
				if item.projectID == "" {
					item.projectID = projectID
				}
				pageItems = append(pageItems, item)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}

// subnetIDsFromRef extracts (projectID, vpcID, subnetID) from a Ref. Tries typed
// assertions first, then falls back to URI path parsing.
func subnetIDsFromRef(ref Ref) (projectID, vpcID, subnetID string, err error) {
	sid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withSubnetID); ok {
			return w.SubnetID(), true
		}
		return "", false
	}, "subnets")
	if !ok || sid == "" {
		return "", "", "", fmt.Errorf("cannot determine subnet ID from Ref %q", ref.URI())
	}
	vid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withVPCID); ok {
			return w.VPCID(), true
		}
		return "", false
	}, "vpcs")
	if !ok || vid == "" {
		return "", "", "", fmt.Errorf("cannot determine VPC ID from Ref %q", ref.URI())
	}
	pid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || pid == "" {
		return "", "", "", fmt.Errorf("cannot determine project ID from Ref %q", ref.URI())
	}
	return pid, vid, sid, nil
}
