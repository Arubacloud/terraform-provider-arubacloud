package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/network"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// VPNRouteRef returns a Ref that points to the VPNRoute nested under a VPNTunnel.
func VPNRouteRef(projectID, tunnelID, routeID string) Ref {
	return URI(fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpnTunnels/%s/vpnRoutes/%s", projectID, tunnelID, routeID))
}

// ---- Wrapper ----

// VPNRoute is the wrapper for an Aruba Cloud VPN Route (a child of a VPNTunnel).
// Construct with aruba.NewVPNRoute() and bind it via InVPNTunnel(tunnel).
//
// Wraps types.VPNRouteResponse / types.VPNRouteRequest. The wrapper carries
// pointer-typed private fields so unset values round-trip through
// the JSON layer correctly.
type VPNRoute struct {
	errMixin
	metadataMixin
	regionalMixin
	vpnTunnelScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	cloudSubnet  *string
	onPremSubnet *string

	response *types.VPNRouteResponse
}

// NewVPNRoute returns a fresh *VPNRoute ready for fluent setters and a Create call.
// Binds vpnTunnelScopedMixin's error sink so InVPNTunnel failures surface via Err().
func NewVPNRoute() *VPNRoute {
	r := &VPNRoute{}
	r.vpnTunnelScopedMixin = bindVPNTunnelScoped(&r.errMixin)
	return r
}

// Setters — chainable, general → specific

// InVPNTunnel binds this VPNRoute to its parent VPNTunnel. Required before Create.
func (r *VPNRoute) InVPNTunnel(t Ref) *VPNRoute { r.intoVPNTunnel(t); return r }

// Named sets the resource name. Required by the API.
func (r *VPNRoute) Named(n string) *VPNRoute { r.named(n); return r }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (r *VPNRoute) Tagged(ts ...string) *VPNRoute {
	for _, tag := range ts {
		r.addTag(tag)
	}
	return r
}

// Untagged removes each listed tag. No-op for tags not present.
func (r *VPNRoute) Untagged(ts ...string) *VPNRoute {
	for _, tag := range ts {
		r.removeTag(tag)
	}
	return r
}

// RetaggedAs replaces the entire tag set with the given values.
func (r *VPNRoute) RetaggedAs(tags ...string) *VPNRoute { r.replaceTags(tags...); return r }

// InRegion sets the region for this resource.
func (r *VPNRoute) InRegion(region Region) *VPNRoute { r.inRegion(region); return r }

// WithCloudSubnet sets the cloud-side subnet CIDR for this VPN route.
func (r *VPNRoute) WithCloudSubnet(cidr string) *VPNRoute { r.cloudSubnet = &cidr; return r }

// WithOnPremSubnet sets the on-premises subnet CIDR for this VPN route.
func (r *VPNRoute) WithOnPremSubnet(cidr string) *VPNRoute { r.onPremSubnet = &cidr; return r }

// Getters — general → specific

// URI satisfies Ref.
func (r *VPNRoute) URI() string { return r.RespURI() }

// VPNRouteID satisfies withVPNRouteID.
func (r *VPNRoute) VPNRouteID() string { return r.ID() }

// Raw shadows responseMetadataMixin.Raw() with the typed VPN route response.
func (r *VPNRoute) Raw() *types.VPNRouteResponse { return r.response }
func (r *VPNRoute) RawJSON() []byte              { return marshalRawJSON(r.response) }
func (r *VPNRoute) RawYAML() []byte              { return marshalRawYAML(r.response) }

// RawRequest returns what toRequest() would emit right now.
func (r *VPNRoute) RawRequest() types.VPNRouteRequest { return r.toRequest() }

// CloudSubnet returns the cloud-side subnet CIDR ("" if unset).
// Prefers the server response value; falls back to the locally-cached setter value.
func (r *VPNRoute) CloudSubnet() string {
	if r.response != nil && r.response.Properties.CloudSubnet.CIDR != "" {
		return r.response.Properties.CloudSubnet.CIDR
	}
	return vpnRouteDerefString(r.cloudSubnet)
}

// CloudSubnetCIDR is an alias for CloudSubnet — returns the cloud-side subnet CIDR ("" if unset).
func (r *VPNRoute) CloudSubnetCIDR() string { return r.CloudSubnet() }

// VPNTunnelURI returns the URI of the parent VPN tunnel from the response, or "" if absent.
func (r *VPNRoute) VPNTunnelURI() string {
	if r.response == nil {
		return ""
	}
	if r.response.Properties.VPNTunnel == nil {
		return ""
	}
	return r.response.Properties.VPNTunnel.URI
}

// OnPremSubnet returns the configured on-premises subnet CIDR ("" if unset).
func (r *VPNRoute) OnPremSubnet() string { return vpnRouteDerefString(r.onPremSubnet) }

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (r *VPNRoute) toRequest() types.VPNRouteRequest {
	props := types.VPNRoutePropertiesRequest{
		CloudSubnet:  vpnRouteDerefString(r.cloudSubnet),
		OnPremSubnet: vpnRouteDerefString(r.onPremSubnet),
	}
	return types.VPNRouteRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: r.toMetadata(),
			Location:                r.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (r *VPNRoute) fromResponse(resp *types.VPNRouteResponse) {
	if resp == nil {
		return
	}
	r.response = resp
	r.setMeta(&resp.Metadata)
	r.named(vpnRouteDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		r.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		r.inRegion(resp.Metadata.LocationResponse.Value)
	}
	r.setStatus(&resp.Status)

	if len(resp.Properties.LinkedResources) > 0 {
		r.setLinked(resp.Properties.LinkedResources)
	}
	if resp.Properties.CloudSubnet.CIDR != "" {
		v := resp.Properties.CloudSubnet.CIDR
		r.cloudSubnet = &v
	}
	if resp.Properties.OnPremSubnet != "" {
		v := resp.Properties.OnPremSubnet
		r.onPremSubnet = &v
	}
	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		r.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if (r.projectID == "" || r.vpnTunnelID == "") && r.RespURI() != "" {
		ids := parseURIIDs(r.RespURI())
		if r.projectID == "" {
			r.projectID = ids["projects"]
		}
		if r.vpnTunnelID == "" {
			r.vpnTunnelID = ids["vpnTunnels"]
		}
	}
}

func vpnRouteDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// vpnRouteLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type vpnRouteLowLevelClient interface {
	List(ctx context.Context, projectID, vpnTunnelID string, params *types.RequestParameters) (*types.Response[types.VPNRouteListResponse], error)
	Get(ctx context.Context, projectID, vpnTunnelID, vpnRouteID string, params *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error)
	Create(ctx context.Context, projectID, vpnTunnelID string, body types.VPNRouteRequest, params *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error)
	Update(ctx context.Context, projectID, vpnTunnelID, vpnRouteID string, body types.VPNRouteRequest, params *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error)
	Delete(ctx context.Context, projectID, vpnTunnelID, vpnRouteID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// vpnRoutesClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates VPNRoute ↔ types.VPNRouteRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type vpnRoutesClientAdapter struct {
	low  vpnRouteLowLevelClient
	rest *restclient.Client
}

var _ VPNRoutesClient = (*vpnRoutesClientAdapter)(nil)

func newVPNRoutesClientAdapter(rest *restclient.Client) *vpnRoutesClientAdapter {
	if rest == nil {
		return &vpnRoutesClientAdapter{}
	}
	return &vpnRoutesClientAdapter{low: network.NewVPNRoutesClientImpl(rest), rest: rest}
}

// Create posts a new VPNRoute to the API and hydrates the wrapper from the response.
func (a *vpnRoutesClientAdapter) Create(ctx context.Context, r *VPNRoute, opts ...CallOption) (*VPNRoute, error) {
	if err := r.Err(); err != nil {
		return r, err
	}
	if r.VPNTunnelID() == "" || r.ProjectID() == "" {
		return r, fmt.Errorf("Create: VPN route has no parent tunnel — call InVPNTunnel first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, r.ProjectID(), r.VPNTunnelID(), r.toRequest(), rp)
	populateHTTPEnvelope(&r.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		r.fromResponse(resp.Data)
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
	}
	if err != nil {
		return r, err
	}
	if resp != nil && !resp.IsSuccess() {
		return r, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return r, nil
}

// Get fetches a VPNRoute by Ref and returns a freshly hydrated wrapper.
func (a *vpnRoutesClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*VPNRoute, error) {
	projectID, vpnTunnelID, vpnRouteID, err := vpnRouteIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, vpnTunnelID, vpnRouteID, rp)
	out := &VPNRoute{}
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
	if out.vpnTunnelID == "" {
		out.vpnTunnelID = vpnTunnelID
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
func (a *vpnRoutesClientAdapter) Update(ctx context.Context, r *VPNRoute, opts ...CallOption) (*VPNRoute, error) {
	if err := r.Err(); err != nil {
		return r, err
	}
	if r.ID() == "" {
		return r, fmt.Errorf("Update: VPN route has no ID — call Get first or seed from response metadata")
	}
	if r.VPNTunnelID() == "" || r.ProjectID() == "" {
		return r, fmt.Errorf("Update: VPN route has no parent tunnel — call InVPNTunnel first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, r.ProjectID(), r.VPNTunnelID(), r.ID(), r.toRequest(), rp)
	populateHTTPEnvelope(&r.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		r.fromResponse(resp.Data)
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
	}
	if err != nil {
		return r, err
	}
	if resp != nil && !resp.IsSuccess() {
		return r, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return r, nil
}

// Delete removes the VPNRoute identified by Ref.
func (a *vpnRoutesClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, vpnTunnelID, vpnRouteID, err := vpnRouteIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, vpnTunnelID, vpnRouteID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of VPNRoute in the given parent scope.
func (a *vpnRoutesClientAdapter) List(ctx context.Context, tunnel Ref, opts ...CallOption) (*List[*VPNRoute], error) {
	projectID, vpnTunnelID, err := vpnTunnelIDsFromRef(tunnel)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.List(ctx, projectID, vpnTunnelID, rp)
	if err != nil {
		return nil, err
	}
	if resp != nil && !resp.IsSuccess() {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	var items []*VPNRoute
	if resp != nil && resp.Data != nil {
		items = make([]*VPNRoute, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			v := &VPNRoute{}
			v.fromResponse(&resp.Data.Values[i])
			v.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, v)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					v.fromResponse(fresh.Raw())
				}
				return nil
			})
			if v.vpnTunnelID == "" {
				v.vpnTunnelID = vpnTunnelID
			}
			if v.projectID == "" {
				v.projectID = projectID
			}
			items = append(items, v)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*VPNRoute], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*VPNRoute], error) {
		fetch := listPageFetch[types.VPNRouteListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*VPNRoute
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*VPNRoute, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				item := &VPNRoute{}
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
				if item.vpnTunnelID == "" {
					item.vpnTunnelID = vpnTunnelID
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

// vpnRouteIDsFromRef extracts (projectID, vpnTunnelID, vpnRouteID) from a Ref.
func vpnRouteIDsFromRef(ref Ref) (projectID, vpnTunnelID, vpnRouteID string, err error) {
	rid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withVPNRouteID); ok {
			return w.VPNRouteID(), true
		}
		return "", false
	}, "vpnRoutes")
	if !ok || rid == "" {
		return "", "", "", fmt.Errorf("cannot determine VPN route ID from Ref %q", ref.URI())
	}
	tid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withVPNTunnelID); ok {
			return w.VPNTunnelID(), true
		}
		return "", false
	}, "vpnTunnels")
	if !ok || tid == "" {
		return "", "", "", fmt.Errorf("cannot determine VPN tunnel ID from Ref %q", ref.URI())
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
	return pid, tid, rid, nil
}
