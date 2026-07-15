package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/network"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// VPCPeeringRouteRef returns a Ref that points to the VPCPeeringRoute nested under a VPCPeering.
func VPCPeeringRouteRef(projectID, vpcID, peeringID, routeID string) Ref {
	return URI(fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpcs/%s/vpcPeerings/%s/vpcPeeringRoutes/%s", projectID, vpcID, peeringID, routeID))
}

// ---- Wrapper ----

// VPCPeeringRoute is the wrapper for an Aruba Cloud VPC Peering Route (a child of a VPCPeering).
// Construct with aruba.NewVPCPeeringRoute() and bind it via InVPCPeering(peering).
//
// Wraps types.VPCPeeringRouteResponse / types.VPCPeeringRouteRequest. The wrapper carries
// pointer-typed private fields so unset values round-trip through
// the JSON layer correctly.
type VPCPeeringRoute struct {
	errMixin
	metadataMixin
	regionalMixin
	vpcPeeringScopedMixin
	responseMetadataMixin
	statusMixin
	httpEnvelopeMixin

	localCIDR     *string
	remoteCIDR    *string
	billingPeriod *BillingPeriod
	response      *types.VPCPeeringRouteResponse
}

// NewVPCPeeringRoute returns a fresh *VPCPeeringRoute ready for fluent setters and a Create call.
// Binds vpcPeeringScopedMixin's error sink so InVPCPeering failures surface via Err().
func NewVPCPeeringRoute() *VPCPeeringRoute {
	r := &VPCPeeringRoute{}
	r.vpcPeeringScopedMixin = bindVPCPeeringScoped(&r.errMixin)
	return r
}

// Setters — chainable, general → specific

// InVPCPeering binds this VPCPeeringRoute to its parent VPCPeering. Required before Create.
func (r *VPCPeeringRoute) InVPCPeering(p Ref) *VPCPeeringRoute { r.intoVPCPeering(p); return r }

// Named sets the resource name. Required by the API.
func (r *VPCPeeringRoute) Named(n string) *VPCPeeringRoute { r.named(n); return r }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (r *VPCPeeringRoute) Tagged(ts ...string) *VPCPeeringRoute {
	for _, t := range ts {
		r.addTag(t)
	}
	return r
}

// Untagged removes each listed tag. No-op for tags not present.
func (r *VPCPeeringRoute) Untagged(ts ...string) *VPCPeeringRoute {
	for _, t := range ts {
		r.removeTag(t)
	}
	return r
}

// RetaggedAs replaces the entire tag set with the given values.
func (r *VPCPeeringRoute) RetaggedAs(ts ...string) *VPCPeeringRoute {
	r.replaceTags(ts...)
	return r
}

// InRegion sets the region for this resource.
func (r *VPCPeeringRoute) InRegion(region Region) *VPCPeeringRoute { r.inRegion(region); return r }

// WithLocalCIDR sets the local network CIDR for the peering route.
func (r *VPCPeeringRoute) WithLocalCIDR(cidr string) *VPCPeeringRoute { r.localCIDR = &cidr; return r }

// WithRemoteCIDR sets the remote network CIDR for the peering route.
func (r *VPCPeeringRoute) WithRemoteCIDR(cidr string) *VPCPeeringRoute {
	r.remoteCIDR = &cidr
	return r
}

// BilledBy sets the billing cadence. Accepted periods are resource-specific; check the API reference.
func (r *VPCPeeringRoute) BilledBy(period BillingPeriod) *VPCPeeringRoute {
	r.billingPeriod = &period
	return r
}

// Getters — general → specific

// URI satisfies Ref.
func (r *VPCPeeringRoute) URI() string { return r.RespURI() }

// VPCPeeringRouteID satisfies withVPCPeeringRouteID.
func (r *VPCPeeringRoute) VPCPeeringRouteID() string { return r.ID() }

// Raw shadows responseMetadataMixin.Raw() with the typed VPC peering route response.
func (r *VPCPeeringRoute) Raw() *types.VPCPeeringRouteResponse { return r.response }
func (r *VPCPeeringRoute) RawJSON() []byte                     { return marshalRawJSON(r.response) }
func (r *VPCPeeringRoute) RawYAML() []byte                     { return marshalRawYAML(r.response) }

// RawRequest returns what toRequest() would emit right now.
func (r *VPCPeeringRoute) RawRequest() types.VPCPeeringRouteRequest { return r.toRequest() }

// LocalCIDR returns the configured local network CIDR ("" if unset).
func (r *VPCPeeringRoute) LocalCIDR() string {
	if r.localCIDR == nil {
		return ""
	}
	return *r.localCIDR
}

// RemoteCIDR returns the configured remote network CIDR ("" if unset).
func (r *VPCPeeringRoute) RemoteCIDR() string {
	if r.remoteCIDR == nil {
		return ""
	}
	return *r.remoteCIDR
}

// BillingPeriod returns the configured billing period ("" if unset).
func (r *VPCPeeringRoute) BillingPeriod() BillingPeriod {
	if r.billingPeriod == nil {
		return ""
	}
	return *r.billingPeriod
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (r *VPCPeeringRoute) toRequest() types.VPCPeeringRouteRequest {
	props := types.VPCPeeringRoutePropertiesRequest{
		BillingPlanCommon: &types.BillingPlanCommon{BillingPeriod: defaultBillingPeriod(r.billingPeriod)},
	}
	if r.localCIDR != nil {
		props.LocalNetworkAddress = *r.localCIDR
	}
	if r.remoteCIDR != nil {
		props.RemoteNetworkAddress = *r.remoteCIDR
	}
	return types.VPCPeeringRouteRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: r.toMetadata(),
			Location:                r.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (r *VPCPeeringRoute) fromResponse(resp *types.VPCPeeringRouteResponse) {
	if resp == nil {
		return
	}
	r.response = resp
	r.setMeta(&resp.Metadata)
	r.named(vpcPeeringRouteDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		r.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		r.inRegion(resp.Metadata.LocationResponse.Value)
	}
	r.setStatus(&resp.Status)

	if resp.Properties.LocalNetworkAddress != "" {
		v := resp.Properties.LocalNetworkAddress
		r.localCIDR = &v
	}
	if resp.Properties.RemoteNetworkAddress != "" {
		v := resp.Properties.RemoteNetworkAddress
		r.remoteCIDR = &v
	}
	if resp.Properties.BillingPlanCommon != nil && resp.Properties.BillingPlanCommon.BillingPeriod != nil {
		r.billingPeriod = resp.Properties.BillingPlanCommon.BillingPeriod
	}

	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		r.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if (r.vpcID == "" || r.projectID == "" || r.vpcPeeringID == "") && r.RespURI() != "" {
		ids := parseURIIDs(r.RespURI())
		if r.vpcID == "" {
			r.vpcID = ids["vpcs"]
		}
		if r.projectID == "" {
			r.projectID = ids["projects"]
		}
		if r.vpcPeeringID == "" {
			r.vpcPeeringID = ids["vpcPeerings"]
		}
	}
}

func vpcPeeringRouteDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// vpcPeeringRouteLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type vpcPeeringRouteLowLevelClient interface {
	List(ctx context.Context, projectID, vpcID, vpcPeeringID string, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteListResponse], error)
	Get(ctx context.Context, projectID, vpcID, vpcPeeringID, vpcPeeringRouteID string, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error)
	Create(ctx context.Context, projectID, vpcID, vpcPeeringID string, body types.VPCPeeringRouteRequest, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error)
	Update(ctx context.Context, projectID, vpcID, vpcPeeringID, vpcPeeringRouteID string, body types.VPCPeeringRouteRequest, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error)
	Delete(ctx context.Context, projectID, vpcID, vpcPeeringID, vpcPeeringRouteID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// vpcPeeringRoutesClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates VPCPeeringRoute ↔ types.VPCPeeringRouteRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type vpcPeeringRoutesClientAdapter struct {
	low  vpcPeeringRouteLowLevelClient
	rest *restclient.Client
}

var _ VPCPeeringRoutesClient = (*vpcPeeringRoutesClientAdapter)(nil)

func newVPCPeeringRoutesClientAdapter(rest *restclient.Client) *vpcPeeringRoutesClientAdapter {
	if rest == nil {
		return &vpcPeeringRoutesClientAdapter{}
	}
	return &vpcPeeringRoutesClientAdapter{low: network.NewVPCPeeringRoutesClientImpl(rest), rest: rest}
}

// Create posts a new VPCPeeringRoute to the API and hydrates the wrapper from the response.
func (a *vpcPeeringRoutesClientAdapter) Create(ctx context.Context, route *VPCPeeringRoute, opts ...CallOption) (*VPCPeeringRoute, error) {
	if err := route.Err(); err != nil {
		return route, err
	}
	if route.VPCPeeringID() == "" || route.VPCID() == "" || route.ProjectID() == "" {
		return route, fmt.Errorf("Create: VPC peering route has no parent peering — call InVPCPeering first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, route.ProjectID(), route.VPCID(), route.VPCPeeringID(), route.toRequest(), rp)
	populateHTTPEnvelope(&route.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		route.fromResponse(resp.Data)
		route.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, route)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				route.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return route, err
	}
	if resp != nil && !resp.IsSuccess() {
		return route, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return route, nil
}

// Get fetches a VPCPeeringRoute by Ref and returns a freshly hydrated wrapper.
func (a *vpcPeeringRoutesClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*VPCPeeringRoute, error) {
	projectID, vpcID, vpcPeeringID, routeID, err := vpcPeeringRouteIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, vpcID, vpcPeeringID, routeID, rp)
	out := &VPCPeeringRoute{}
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
	if out.vpcPeeringID == "" {
		out.vpcPeeringID = vpcPeeringID
	}
	if out.vpcID == "" {
		out.vpcID = vpcID
	}
	if out.projectID == "" {
		out.projectID = projectID
	}
	if err != nil {
		return out, err
	}
	if resp != nil && !resp.IsSuccess() {
		return out, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return out, nil
}

// Update sends a PUT for the current wrapper state. Requires ID and parent.
func (a *vpcPeeringRoutesClientAdapter) Update(ctx context.Context, route *VPCPeeringRoute, opts ...CallOption) (*VPCPeeringRoute, error) {
	if err := route.Err(); err != nil {
		return route, err
	}
	if route.ID() == "" {
		return route, fmt.Errorf("Update: VPC peering route has no ID — call Get first or seed from response metadata")
	}
	if route.VPCPeeringID() == "" || route.VPCID() == "" || route.ProjectID() == "" {
		return route, fmt.Errorf("Update: VPC peering route has no parent peering — call InVPCPeering first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, route.ProjectID(), route.VPCID(), route.VPCPeeringID(), route.ID(), route.toRequest(), rp)
	populateHTTPEnvelope(&route.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		route.fromResponse(resp.Data)
		route.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, route)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				route.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return route, err
	}
	if resp != nil && !resp.IsSuccess() {
		return route, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return route, nil
}

// Delete removes the VPCPeeringRoute identified by Ref.
func (a *vpcPeeringRoutesClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, vpcID, vpcPeeringID, routeID, err := vpcPeeringRouteIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, vpcID, vpcPeeringID, routeID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of VPCPeeringRoute in the given parent scope.
func (a *vpcPeeringRoutesClientAdapter) List(ctx context.Context, peering Ref, opts ...CallOption) (*List[*VPCPeeringRoute], error) {
	projectID, vpcID, vpcPeeringID, err := vpcPeeringIDsFromRef(peering)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.List(ctx, projectID, vpcID, vpcPeeringID, rp)
	if err != nil {
		return nil, err
	}
	if resp != nil && !resp.IsSuccess() {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	var items []*VPCPeeringRoute
	if resp != nil && resp.Data != nil {
		items = make([]*VPCPeeringRoute, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			r := &VPCPeeringRoute{}
			r.fromResponse(&resp.Data.Values[i])
			r.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, r)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					r.fromResponse(fresh.Raw())
				}
				return nil
			})
			if r.vpcPeeringID == "" {
				r.vpcPeeringID = vpcPeeringID
			}
			if r.vpcID == "" {
				r.vpcID = vpcID
			}
			if r.projectID == "" {
				r.projectID = projectID
			}
			items = append(items, r)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*VPCPeeringRoute], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*VPCPeeringRoute], error) {
		fetch := listPageFetch[types.VPCPeeringRouteListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*VPCPeeringRoute
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*VPCPeeringRoute, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				item := &VPCPeeringRoute{}
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
				if item.vpcPeeringID == "" {
					item.vpcPeeringID = vpcPeeringID
				}
				if item.vpcID == "" {
					item.vpcID = vpcID
				}
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

// vpcPeeringRouteIDsFromRef extracts (projectID, vpcID, vpcPeeringID, vpcPeeringRouteID) from a Ref.
func vpcPeeringRouteIDsFromRef(ref Ref) (projectID, vpcID, vpcPeeringID, vpcPeeringRouteID string, err error) {
	rid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withVPCPeeringRouteID); ok {
			return w.VPCPeeringRouteID(), true
		}
		return "", false
	}, "vpcPeeringRoutes")
	if !ok {
		return "", "", "", "", fmt.Errorf("cannot determine VPC peering route ID from Ref %q", ref.URI())
	}
	pid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withVPCPeeringID); ok {
			return w.VPCPeeringID(), true
		}
		return "", false
	}, "vpcPeerings")
	if !ok {
		return "", "", "", "", fmt.Errorf("cannot determine VPC peering ID from Ref %q", ref.URI())
	}
	vid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withVPCID); ok {
			return w.VPCID(), true
		}
		return "", false
	}, "vpcs")
	if !ok || vid == "" {
		return "", "", "", "", fmt.Errorf("cannot determine VPC ID from Ref %q", ref.URI())
	}
	projID, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || projID == "" {
		return "", "", "", "", fmt.Errorf("cannot determine project ID from Ref %q", ref.URI())
	}
	return projID, vid, pid, rid, nil
}
