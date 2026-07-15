package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/network"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// VPNTunnelRef returns a Ref that points to the VPNTunnel with the given IDs.
func VPNTunnelRef(projectID, tunnelID string) Ref {
	return URI(fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpnTunnels/%s", projectID, tunnelID))
}

// ---- Wrapper ----

// VPNTunnel is the wrapper for an Aruba Cloud VPN Tunnel (a child of a Project).
// Construct with aruba.NewVPNTunnel() and bind it via InProject(project).
//
// Wraps types.VPNTunnelResponse / types.VPNTunnelRequest. The wrapper carries
// pointer-typed private fields so unset values round-trip through
// the JSON layer correctly.
type VPNTunnel struct {
	errMixin
	metadataMixin
	regionalMixin
	projectScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	vpnType            *VPNType
	vpnClientProtocol  *VPNClientProtocol
	billingPeriod      *BillingPeriod
	peerClientPublicIP *string

	ipConfig *VPNIPConfig
	ike      *VPNIKE
	esp      *VPNESP
	psk      *VPNPSK

	response *types.VPNTunnelResponse
}

// NewVPNTunnel returns a fresh *VPNTunnel ready for fluent setters and a Create call.
// Binds projectScopedMixin's error sink so InProject failures surface via Err().
func NewVPNTunnel() *VPNTunnel {
	t := &VPNTunnel{}
	t.projectScopedMixin = bindProjectScoped(&t.errMixin)
	return t
}

// Setters — chainable, general → specific

// InProject binds this VPNTunnel to its parent project. Required before Create.
func (t *VPNTunnel) InProject(p Ref) *VPNTunnel { t.intoProject(p); return t }

// Named sets the resource name. Required by the API.
func (t *VPNTunnel) Named(n string) *VPNTunnel { t.named(n); return t }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (t *VPNTunnel) Tagged(ts ...string) *VPNTunnel {
	for _, tag := range ts {
		t.addTag(tag)
	}
	return t
}

// Untagged removes each listed tag. No-op for tags not present.
func (t *VPNTunnel) Untagged(ts ...string) *VPNTunnel {
	for _, tag := range ts {
		t.removeTag(tag)
	}
	return t
}

// RetaggedAs replaces the entire tag set with the given values.
func (t *VPNTunnel) RetaggedAs(tags ...string) *VPNTunnel { t.replaceTags(tags...); return t }

// InRegion sets the region for this resource.
func (t *VPNTunnel) InRegion(r Region) *VPNTunnel { t.inRegion(r); return t }

// OfType sets the VPN tunnel type (Site-to-Site, Client, etc.).
func (t *VPNTunnel) OfType(s VPNType) *VPNTunnel { t.vpnType = &s; return t }

// WithVPNClientProtocol sets the client VPN protocol (e.g. IKEv2).
func (t *VPNTunnel) WithVPNClientProtocol(s VPNClientProtocol) *VPNTunnel {
	t.vpnClientProtocol = &s
	return t
}

// BilledBy sets the billing cadence. Accepted periods are resource-specific; check the API reference.
func (t *VPNTunnel) BilledBy(period BillingPeriod) *VPNTunnel { t.billingPeriod = &period; return t }

// WithPeerClientPublicIP sets the public IP of the remote VPN peer.
func (t *VPNTunnel) WithPeerClientPublicIP(s string) *VPNTunnel {
	t.peerClientPublicIP = &s
	return t
}

// WithIPConfig attaches a VPNIPConfig sub-builder. Errors from c are drained into the tunnel.
func (t *VPNTunnel) WithIPConfig(c *VPNIPConfig) *VPNTunnel {
	t.ipConfig = c
	if c != nil {
		for _, e := range c.errs {
			t.addErr(e)
		}
	}
	return t
}

// WithIKESettings attaches an IKE settings sub-builder. Errors from k are drained into the tunnel.
func (t *VPNTunnel) WithIKESettings(k *VPNIKE) *VPNTunnel {
	t.ike = k
	if k != nil {
		for _, e := range k.errs {
			t.addErr(e)
		}
	}
	return t
}

// WithESPSettings attaches an ESP settings sub-builder. Errors from e are drained into the tunnel.
func (t *VPNTunnel) WithESPSettings(e *VPNESP) *VPNTunnel {
	t.esp = e
	if e != nil {
		for _, err := range e.errs {
			t.addErr(err)
		}
	}
	return t
}

// WithPSKSettings attaches a PSK settings sub-builder. Errors from p are drained into the tunnel.
func (t *VPNTunnel) WithPSKSettings(p *VPNPSK) *VPNTunnel {
	t.psk = p
	if p != nil {
		for _, e := range p.errs {
			t.addErr(e)
		}
	}
	return t
}

// Getters — general → specific

// URI satisfies Ref.
func (t *VPNTunnel) URI() string { return t.RespURI() }

// VPNTunnelID satisfies withVPNTunnelID.
func (t *VPNTunnel) VPNTunnelID() string { return t.ID() }

// Raw shadows responseMetadataMixin.Raw() with the typed VPN tunnel response.
func (t *VPNTunnel) Raw() *types.VPNTunnelResponse { return t.response }
func (t *VPNTunnel) RawJSON() []byte               { return marshalRawJSON(t.response) }
func (t *VPNTunnel) RawYAML() []byte               { return marshalRawYAML(t.response) }

// RawRequest returns what toRequest() would emit right now.
func (t *VPNTunnel) RawRequest() types.VPNTunnelRequest { return t.toRequest() }

// Read accessors.

// IPConfig returns the current IP configuration sub-builder (nil if unset).
func (t *VPNTunnel) IPConfig() *VPNIPConfig { return t.ipConfig }

// IKE returns the IKE settings sub-builder (nil if unset).
func (t *VPNTunnel) IKE() *VPNIKE { return t.ike }

// ESP returns the ESP settings sub-builder (nil if unset).
func (t *VPNTunnel) ESP() *VPNESP { return t.esp }

// PSK returns the Pre-Shared Key settings sub-builder (nil if unset).
func (t *VPNTunnel) PSK() *VPNPSK { return t.psk }

// VPNType returns the configured VPN tunnel type ("" if unset).
func (t *VPNTunnel) VPNType() VPNType {
	if t.vpnType == nil {
		return ""
	}
	return *t.vpnType
}

// VPNClientProtocol returns the configured client VPN protocol ("" if unset).
func (t *VPNTunnel) VPNClientProtocol() VPNClientProtocol {
	if t.vpnClientProtocol == nil {
		return ""
	}
	return *t.vpnClientProtocol
}

// BillingPeriod returns the configured billing period ("" if unset).
func (t *VPNTunnel) BillingPeriod() BillingPeriod {
	if t.billingPeriod == nil {
		return ""
	}
	return *t.billingPeriod
}

// PeerClientPublicIP returns the peer client public IP address, or "" if unset.
func (t *VPNTunnel) PeerClientPublicIP() string { return vpnTunnelDerefString(t.peerClientPublicIP) }

// RoutesNumber returns the number of valid VPN routes associated with this tunnel,
// as reported by the last server response. Returns 0 before any response.
func (t *VPNTunnel) RoutesNumber() int32 {
	if t.response == nil {
		return 0
	}
	return t.response.Properties.RoutesNumber
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (t *VPNTunnel) toRequest() types.VPNTunnelRequest {
	props := types.VPNTunnelPropertiesRequest{
		VPNType:           t.vpnType,
		VPNClientProtocol: t.vpnClientProtocol,
	}
	if t.ipConfig != nil {
		props.IPConfigurationsCommon = t.ipConfig.build()
	}
	if t.ike != nil || t.esp != nil || t.psk != nil || t.peerClientPublicIP != nil {
		cs := &types.VPNClientSettingsCommon{PeerClientPublicIP: t.peerClientPublicIP}
		if t.ike != nil {
			cs.IKE = t.ike.build()
		}
		if t.esp != nil {
			cs.ESP = t.esp.build()
		}
		if t.psk != nil {
			cs.PSK = t.psk.build()
		}
		props.VPNClientSettingsCommon = cs
	}
	props.BillingPlanCommon = &types.BillingPlanCommon{BillingPeriod: defaultBillingPeriod(t.billingPeriod)}
	return types.VPNTunnelRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: t.toMetadata(),
			Location:                t.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (t *VPNTunnel) fromResponse(resp *types.VPNTunnelResponse) {
	if resp == nil {
		return
	}
	t.response = resp
	t.setMeta(&resp.Metadata)
	t.named(vpnTunnelDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		t.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		t.inRegion(resp.Metadata.LocationResponse.Value)
	}
	t.setStatus(&resp.Status)

	if resp.Properties.VPNType != nil {
		v := *resp.Properties.VPNType
		t.vpnType = &v
	}
	if resp.Properties.VPNClientProtocol != nil {
		p := *resp.Properties.VPNClientProtocol
		t.vpnClientProtocol = &p
	}
	if resp.Properties.BillingPlanCommon != nil && resp.Properties.BillingPlanCommon.BillingPeriod != nil {
		t.billingPeriod = resp.Properties.BillingPlanCommon.BillingPeriod
	}
	if cs := resp.Properties.VPNClientSettingsCommon; cs != nil {
		if cs.PeerClientPublicIP != nil {
			v := *cs.PeerClientPublicIP
			t.peerClientPublicIP = &v
		}
		if cs.IKE != nil {
			k := &VPNIKE{
				lifetime:    cs.IKE.Lifetime,
				encryption:  cs.IKE.Encryption,
				hash:        cs.IKE.Hash,
				dhGroup:     cs.IKE.DHGroup,
				dpdAction:   cs.IKE.DPDAction,
				dpdInterval: cs.IKE.DPDInterval,
				dpdTimeout:  cs.IKE.DPDTimeout,
			}
			t.ike = k
		}
		if cs.ESP != nil {
			e := &VPNESP{
				lifetime:   cs.ESP.Lifetime,
				encryption: cs.ESP.Encryption,
				hash:       cs.ESP.Hash,
				pfs:        cs.ESP.PFS,
			}
			t.esp = e
		}
		if cs.PSK != nil {
			p := &VPNPSK{
				cloudSite:  cs.PSK.CloudSite,
				onPremSite: cs.PSK.OnPremSite,
				secret:     cs.PSK.Secret,
			}
			t.psk = p
		}
	}
	if ipc := resp.Properties.IPConfigurationsCommon; ipc != nil {
		c := &VPNIPConfig{vpc: ipc.VPC, publicIP: ipc.PublicIP}
		if ipc.Subnet != nil {
			c.subnetName = ipc.Subnet.Name
			c.subnetCIDR = ipc.Subnet.CIDR
			c.hasSubnet = true
		}
		t.ipConfig = c
	}

	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		t.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if t.projectID == "" && t.RespURI() != "" {
		if id := parseURIIDs(t.RespURI())["projects"]; id != "" {
			t.projectID = id
		}
	}
}

func vpnTunnelDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// vpnTunnelLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type vpnTunnelLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.VPNTunnelListResponse], error)
	Get(ctx context.Context, projectID, vpnTunnelID string, params *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error)
	Create(ctx context.Context, projectID string, body types.VPNTunnelRequest, params *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error)
	Update(ctx context.Context, projectID, vpnTunnelID string, body types.VPNTunnelRequest, params *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error)
	Delete(ctx context.Context, projectID, vpnTunnelID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// vpnTunnelsClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates VPNTunnel ↔ types.VPNTunnelRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type vpnTunnelsClientAdapter struct {
	low  vpnTunnelLowLevelClient
	rest *restclient.Client
}

var _ VPNTunnelsClient = (*vpnTunnelsClientAdapter)(nil)

func newVPNTunnelsClientAdapter(rest *restclient.Client) *vpnTunnelsClientAdapter {
	if rest == nil {
		return &vpnTunnelsClientAdapter{}
	}
	return &vpnTunnelsClientAdapter{low: network.NewVPNTunnelsClientImpl(rest), rest: rest}
}

// Create posts a new VPNTunnel to the API and hydrates the wrapper from the response.
func (a *vpnTunnelsClientAdapter) Create(ctx context.Context, t *VPNTunnel, opts ...CallOption) (*VPNTunnel, error) {
	if err := t.Err(); err != nil {
		return t, err
	}
	if t.ProjectID() == "" {
		return t, fmt.Errorf("Create: VPN tunnel has no project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, t.ProjectID(), t.toRequest(), rp)
	populateHTTPEnvelope(&t.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		t.fromResponse(resp.Data)
		t.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, t)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				t.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return t, err
	}
	if resp != nil && !resp.IsSuccess() {
		return t, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return t, nil
}

// Get fetches a VPNTunnel by Ref and returns a freshly hydrated wrapper.
func (a *vpnTunnelsClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*VPNTunnel, error) {
	projectID, vpnTunnelID, err := vpnTunnelIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, vpnTunnelID, rp)
	out := &VPNTunnel{}
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
func (a *vpnTunnelsClientAdapter) Update(ctx context.Context, t *VPNTunnel, opts ...CallOption) (*VPNTunnel, error) {
	if err := t.Err(); err != nil {
		return t, err
	}
	if t.ID() == "" {
		return t, fmt.Errorf("Update: VPN tunnel has no ID — call Get first or seed from response metadata")
	}
	if t.ProjectID() == "" {
		return t, fmt.Errorf("Update: VPN tunnel has no project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, t.ProjectID(), t.ID(), t.toRequest(), rp)
	populateHTTPEnvelope(&t.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		t.fromResponse(resp.Data)
		t.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, t)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				t.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return t, err
	}
	if resp != nil && !resp.IsSuccess() {
		return t, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return t, nil
}

// Delete removes the VPNTunnel identified by Ref.
func (a *vpnTunnelsClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, vpnTunnelID, err := vpnTunnelIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, vpnTunnelID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of VPNTunnel in the given parent scope.
func (a *vpnTunnelsClientAdapter) List(ctx context.Context, project Ref, opts ...CallOption) (*List[*VPNTunnel], error) {
	projectID, err := projectIDFromRef(project)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.List(ctx, projectID, rp)
	if err != nil {
		return nil, err
	}
	if resp != nil && !resp.IsSuccess() {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	var items []*VPNTunnel
	if resp != nil && resp.Data != nil {
		items = make([]*VPNTunnel, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			v := &VPNTunnel{}
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
			if v.projectID == "" {
				v.projectID = projectID
			}
			items = append(items, v)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*VPNTunnel], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*VPNTunnel], error) {
		fetch := listPageFetch[types.VPNTunnelListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*VPNTunnel
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*VPNTunnel, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				item := &VPNTunnel{}
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

// vpnTunnelIDsFromRef extracts (projectID, vpnTunnelID) from a Ref.
func vpnTunnelIDsFromRef(ref Ref) (projectID, vpnTunnelID string, err error) {
	tid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withVPNTunnelID); ok {
			return w.VPNTunnelID(), true
		}
		return "", false
	}, "vpnTunnels")
	if !ok || tid == "" {
		return "", "", fmt.Errorf("cannot determine VPN tunnel ID from Ref %q", ref.URI())
	}
	pid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || pid == "" {
		return "", "", fmt.Errorf("cannot determine project ID from Ref %q", ref.URI())
	}
	return pid, tid, nil
}
