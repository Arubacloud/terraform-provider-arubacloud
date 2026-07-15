package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/compute"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// CloudServer is the wrapper for an Aruba Cloud Compute server (a direct child of a Project).
// Construct with aruba.NewCloudServer() and bind it via InProject(project), WithVPC(vpc),
// BootingFrom(volume), etc.
//
// Schema asymmetry: the request side uses FlavorName *string under the "flavorName" wire
// field; the response side returns a full Flavor struct under the "flavor" wire field.
// This wrapper exposes OfFlavor(flavor) for the request leg and Flavor() / FlavorRaw()
// for the response leg.
//
// The response also carries Template ReferenceResourceCommon (no request equivalent); this
// wrapper exposes Template() as a read-only getter.
type CloudServer struct {
	errMixin
	metadataMixin
	zonalMixin
	projectScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	// Request-side scalars.
	flavor        *CloudServerFlavor
	userData      *string
	vpcPreset     *bool
	billingPeriod *BillingPeriod

	// Body-refs (single).
	vpcRef        *string
	bootVolumeRef *string
	keyPairRef    *string
	elasticIPRef  *string

	// Body-refs (multi-slice, Create-time).
	subnetRefs        []string
	securityGroupRefs []string

	// Update-time association/attachment deltas (cleared after a successful Update call).
	subnetsToAssociate    []string
	subnetsToDisassociate []string
	sgsToAssociate        []string
	sgsToDisassociate     []string
	eipsToAssociate       []string
	eipsToDisassociate    []string
	dataVolumesToAttach   []string
	dataVolumesToDetach   []string

	// Action executor — set by the adapter when this wrapper is produced by a real client
	// call. Locally-constructed wrappers (NewCloudServer()) have actions == nil and will
	// return a clear error when PowerOn/PowerOff/SetPassword are called.
	actions cloudServerActions

	response *types.CloudServerResponse
}

// NewCloudServer returns a fresh *CloudServer ready for fluent setters and a Create call.
// Binds projectScopedMixin's error sink so IntoProject failures surface via Err().
//
// Action methods (PowerOn, PowerOff, SetPassword) on the returned wrapper will fail until
// the wrapper has been hydrated by a real client call (Get/Create/Update/List populate
// the internal action executor).
func NewCloudServer() *CloudServer {
	cs := &CloudServer{}
	cs.projectScopedMixin = bindProjectScoped(&cs.errMixin)
	return cs
}

// Setters — chainable, general → specific

// InProject binds this CloudServer to its parent project. Required before Create.
func (cs *CloudServer) InProject(p Ref) *CloudServer { cs.intoProject(p); return cs }

// Named sets the resource name. Required by the API.
func (cs *CloudServer) Named(n string) *CloudServer { cs.named(n); return cs }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (cs *CloudServer) Tagged(ts ...string) *CloudServer {
	for _, t := range ts {
		cs.addTag(t)
	}
	return cs
}

// Untagged removes each listed tag. No-op for tags not present.
func (cs *CloudServer) Untagged(ts ...string) *CloudServer {
	for _, t := range ts {
		cs.removeTag(t)
	}
	return cs
}

// RetaggedAs replaces the entire tag set with the given values.
func (cs *CloudServer) RetaggedAs(ts ...string) *CloudServer { cs.replaceTags(ts...); return cs }

// InRegion sets the region for this resource.
func (cs *CloudServer) InRegion(region Region) *CloudServer { cs.inRegion(region); return cs }

// InZone sets the availability zone. More specific than InRegion.
func (cs *CloudServer) InZone(zone Zone) *CloudServer { cs.inZone(zone); return cs }

// OfFlavor sets the server flavor (instance size). Maps to wire field "flavorName".
func (cs *CloudServer) OfFlavor(flavor CloudServerFlavor) *CloudServer {
	cs.flavor = &flavor
	return cs
}

// WithUserData sets the base64-encoded cloud-init user data.
func (cs *CloudServer) WithUserData(b64 string) *CloudServer { cs.userData = &b64; return cs }

// WithVPCPreset marks the server to use VPC preset networking.
func (cs *CloudServer) WithVPCPreset() *CloudServer { v := true; cs.vpcPreset = &v; return cs }

// WithoutVPCPreset disables VPC preset networking.
func (cs *CloudServer) WithoutVPCPreset() *CloudServer { v := false; cs.vpcPreset = &v; return cs }

// BilledBy sets the billing cadence. Accepted periods are resource-specific; check the API reference.
func (cs *CloudServer) BilledBy(period BillingPeriod) *CloudServer {
	cs.billingPeriod = &period
	return cs
}

// Single body-ref setters.

// WithVPC attaches the server to the given VPC by URI reference.
func (cs *CloudServer) WithVPC(v Ref) *CloudServer { return cs.setSingleRef("WithVPC", v, &cs.vpcRef) }

// BootingFrom sets the boot volume by URI reference.
func (cs *CloudServer) BootingFrom(vol Ref) *CloudServer {
	return cs.setSingleRef("BootingFrom", vol, &cs.bootVolumeRef)
}

// UsingKeyPair attaches an SSH key pair by URI reference.
func (cs *CloudServer) UsingKeyPair(kp Ref) *CloudServer {
	return cs.setSingleRef("UsingKeyPair", kp, &cs.keyPairRef)
}

// WithElasticIP attaches an elastic IP by URI reference.
func (cs *CloudServer) WithElasticIP(eip Ref) *CloudServer {
	return cs.setSingleRef("WithElasticIP", eip, &cs.elasticIPRef)
}

// Multi-ref slice setters.

// OnSubnets appends subnets by URI reference. Repeated calls append.
func (cs *CloudServer) OnSubnets(s ...Ref) *CloudServer {
	for _, ref := range s {
		cs.appendRef("OnSubnets", ref, &cs.subnetRefs)
	}
	return cs
}

// WithSecurityGroups appends security groups by URI reference. Repeated calls append.
func (cs *CloudServer) WithSecurityGroups(sg ...Ref) *CloudServer {
	for _, ref := range sg {
		cs.appendRef("WithSecurityGroups", ref, &cs.securityGroupRefs)
	}
	return cs
}

// Update-time delta setters — these queue association/attachment changes that
// are dispatched by the adapter's Update call. They must NOT be used during Create;
// for Create-time subnet/SG configuration use OnSubnets/WithSecurityGroups instead.

// AssociateSubnets queues subnets to associate on the next Update call. Repeated calls append.
func (cs *CloudServer) AssociateSubnets(refs ...Ref) *CloudServer {
	for _, r := range refs {
		cs.appendRef("AssociateSubnets", r, &cs.subnetsToAssociate)
	}
	return cs
}

// DisassociateSubnets queues subnets to disassociate on the next Update call. Repeated calls append.
func (cs *CloudServer) DisassociateSubnets(refs ...Ref) *CloudServer {
	for _, r := range refs {
		cs.appendRef("DisassociateSubnets", r, &cs.subnetsToDisassociate)
	}
	return cs
}

// AssociateSecurityGroups queues security groups to associate on the next Update call. Repeated calls append.
func (cs *CloudServer) AssociateSecurityGroups(refs ...Ref) *CloudServer {
	for _, r := range refs {
		cs.appendRef("AssociateSecurityGroups", r, &cs.sgsToAssociate)
	}
	return cs
}

// DisassociateSecurityGroups queues security groups to disassociate on the next Update call. Repeated calls append.
func (cs *CloudServer) DisassociateSecurityGroups(refs ...Ref) *CloudServer {
	for _, r := range refs {
		cs.appendRef("DisassociateSecurityGroups", r, &cs.sgsToDisassociate)
	}
	return cs
}

// AssociateElasticIPs queues elastic IPs to associate on the next Update call. Repeated calls append.
func (cs *CloudServer) AssociateElasticIPs(refs ...Ref) *CloudServer {
	for _, r := range refs {
		cs.appendRef("AssociateElasticIPs", r, &cs.eipsToAssociate)
	}
	return cs
}

// DisassociateElasticIPs queues elastic IPs to disassociate on the next Update call. Repeated calls append.
func (cs *CloudServer) DisassociateElasticIPs(refs ...Ref) *CloudServer {
	for _, r := range refs {
		cs.appendRef("DisassociateElasticIPs", r, &cs.eipsToDisassociate)
	}
	return cs
}

// AttachDataVolumes queues data volumes to attach on the next Update call. Repeated calls append.
func (cs *CloudServer) AttachDataVolumes(refs ...Ref) *CloudServer {
	for _, r := range refs {
		cs.appendRef("AttachDataVolumes", r, &cs.dataVolumesToAttach)
	}
	return cs
}

// DetachDataVolumes queues data volumes to detach on the next Update call. Repeated calls append.
func (cs *CloudServer) DetachDataVolumes(refs ...Ref) *CloudServer {
	for _, r := range refs {
		cs.appendRef("DetachDataVolumes", r, &cs.dataVolumesToDetach)
	}
	return cs
}

// Internal ref helpers.

func (cs *CloudServer) setSingleRef(label string, r Ref, dst **string) *CloudServer {
	uri := r.URI()
	if uri == "" {
		cs.addErr(fmt.Errorf("%s: empty URI", label))
		return cs
	}
	*dst = &uri
	return cs
}

func (cs *CloudServer) appendRef(label string, r Ref, dst *[]string) {
	uri := r.URI()
	if uri == "" {
		cs.addErr(fmt.Errorf("%s: empty URI", label))
		return
	}
	*dst = append(*dst, uri)
}

// Getters — general → specific

// Ref + ID accessors.

// URI satisfies Ref by returning the server-assigned canonical URI, or "" if Create hasn't run yet.
func (cs *CloudServer) URI() string { return cs.RespURI() }

// CloudServerID satisfies withCloudServerID so child wrappers can extract this ID by typed assertion.
func (cs *CloudServer) CloudServerID() string { return cs.ID() }

// Accessors.

// Raw shadows responseMetadataMixin.Raw() with the typed CloudServer response.
func (cs *CloudServer) Raw() *types.CloudServerResponse { return cs.response }
func (cs *CloudServer) RawJSON() []byte                 { return marshalRawJSON(cs.response) }
func (cs *CloudServer) RawYAML() []byte                 { return marshalRawYAML(cs.response) }

// RawRequest returns what toRequest() would emit right now.
func (cs *CloudServer) RawRequest() types.CloudServerRequest { return cs.toRequest() }

// Flavor returns the flavor name. On a hydrated response the value comes from the
// response's Flavor.Name; before hydration it returns what was passed to OfFlavor.
func (cs *CloudServer) Flavor() CloudServerFlavor {
	if cs.response != nil && cs.response.Properties.Flavor.Name != "" {
		return cs.response.Properties.Flavor.Name
	}
	if cs.flavor == nil {
		return ""
	}
	return *cs.flavor
}

// FlavorRaw returns the full flavor struct from the last response, or nil.
func (cs *CloudServer) FlavorRaw() *types.CloudServerFlavorResponse {
	if cs.response == nil {
		return nil
	}
	return &cs.response.Properties.Flavor
}

// Template returns the template URI from the last response (read-only; no request equivalent).
func (cs *CloudServer) Template() string {
	if cs.response == nil {
		return ""
	}
	return cs.response.Properties.Template.URI
}

// VPC returns the VPC URI from the response, or the locally-set ref URI if unhydrated.
func (cs *CloudServer) VPC() string {
	if cs.response != nil && cs.response.Properties.VPC.URI != "" {
		return cs.response.Properties.VPC.URI
	}
	return cloudServerDerefString(cs.vpcRef)
}

// BootVolume returns the boot volume URI from the response, or the locally-set ref URI if unhydrated.
func (cs *CloudServer) BootVolume() string {
	if cs.response != nil && cs.response.Properties.BootVolume.URI != "" {
		return cs.response.Properties.BootVolume.URI
	}
	return cloudServerDerefString(cs.bootVolumeRef)
}

// KeyPair returns the key pair URI from the response, or the locally-set ref URI if unhydrated.
func (cs *CloudServer) KeyPair() string {
	if cs.response != nil && cs.response.Properties.KeyPair.URI != "" {
		return cs.response.Properties.KeyPair.URI
	}
	return cloudServerDerefString(cs.keyPairRef)
}

// ElasticIP returns the elastic IP URI set via WithElasticIP, or "" if unset.
// The API does not return the elastic IP in the response body; this getter
// returns the locally-cached setter value only.
func (cs *CloudServer) ElasticIP() string { return cloudServerDerefString(cs.elasticIPRef) }

// Subnets returns the subnet URIs. After a Get/Create/Update the values come from
// the server response (NetworkInterfaces[].Subnet); before hydration they reflect
// what was passed to OnSubnets.
func (cs *CloudServer) Subnets() []string {
	if cs.response != nil && len(cs.response.Properties.NetworkInterfaces) > 0 {
		out := make([]string, 0, len(cs.response.Properties.NetworkInterfaces))
		for _, ni := range cs.response.Properties.NetworkInterfaces {
			if ni.Subnet != nil && *ni.Subnet != "" {
				out = append(out, *ni.Subnet)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	if len(cs.subnetRefs) == 0 {
		return nil
	}
	out := make([]string, len(cs.subnetRefs))
	copy(out, cs.subnetRefs)
	return out
}

// SecurityGroups returns the security group URIs set via WithSecurityGroups.
// The API does not return security groups in a distinguishable form in the
// response body; this getter returns the locally-cached setter values only.
func (cs *CloudServer) SecurityGroups() []string {
	if len(cs.securityGroupRefs) == 0 {
		return nil
	}
	out := make([]string, len(cs.securityGroupRefs))
	copy(out, cs.securityGroupRefs)
	return out
}

// UserData returns the base64-encoded cloud-init user data set via WithUserData.
// The API does not return user data in the response body; this getter returns
// the locally-cached setter value only.
func (cs *CloudServer) UserData() string { return cloudServerDerefString(cs.userData) }

// NetworkInterfaces returns the list of network interface details from the last response, or nil.
func (cs *CloudServer) NetworkInterfaces() []types.CloudServerNetworkInterfaceResponse {
	if cs.response == nil {
		return nil
	}
	return cs.response.Properties.NetworkInterfaces
}

// IsVPCPreset returns whether VPC preset networking was requested.
// Returns false if unset.
func (cs *CloudServer) IsVPCPreset() bool {
	if cs.vpcPreset == nil {
		return false
	}
	return *cs.vpcPreset
}

// BillingPeriod returns the billing period, or "" if unset.
func (cs *CloudServer) BillingPeriod() BillingPeriod {
	if cs.billingPeriod == nil {
		return ""
	}
	return *cs.billingPeriod
}

// Action methods (require hydration via a client Get/Create/Update/List call).

// PowerOn powers on the cloud server. Requires the wrapper to be obtained via a client call.
func (cs *CloudServer) PowerOn(ctx context.Context, opts ...CallOption) error {
	if err := cs.preActionCheck("PowerOn"); err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := cs.actions.powerOn(ctx, cs.ProjectID(), cs.CloudServerID(), rp)
	populateHTTPEnvelope(&cs.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		cs.fromResponse(resp.Data)
	}
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// PowerOff powers off the cloud server. Requires the wrapper to be obtained via a client call.
func (cs *CloudServer) PowerOff(ctx context.Context, opts ...CallOption) error {
	if err := cs.preActionCheck("PowerOff"); err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := cs.actions.powerOff(ctx, cs.ProjectID(), cs.CloudServerID(), rp)
	populateHTTPEnvelope(&cs.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		cs.fromResponse(resp.Data)
	}
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// SetPassword sets the administrative password for the cloud server. Requires a client-obtained wrapper.
func (cs *CloudServer) SetPassword(ctx context.Context, password string, opts ...CallOption) error {
	if err := cs.preActionCheck("SetPassword"); err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := cs.actions.setPassword(ctx, cs.ProjectID(), cs.CloudServerID(), password, rp)
	populateHTTPEnvelope(&cs.httpEnvelopeMixin, resp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// hasMetadataChanges reports whether name or tags differ from the last hydrated response,
// meaning a PUT to /cloudServers/:id is required in Update.
func (cs *CloudServer) hasMetadataChanges() bool {
	if cs.response == nil {
		return true // not yet hydrated — always send
	}
	if cs.name != cloudServerDerefString(cs.response.Metadata.Name) {
		return true
	}
	rt := cs.response.Metadata.Tags
	ct := cs.tags
	if len(ct) != len(rt) {
		return true
	}
	for i := range ct {
		if ct[i] != rt[i] {
			return true
		}
	}
	return false
}

// cloudServerStringsToCommon converts URI strings to ReferenceResourceCommon, skipping empty strings.
func cloudServerStringsToCommon(uris []string) []types.ReferenceResourceCommon {
	if len(uris) == 0 {
		return nil
	}
	out := make([]types.ReferenceResourceCommon, 0, len(uris))
	for _, u := range uris {
		if u != "" {
			out = append(out, types.ReferenceResourceCommon{URI: u})
		}
	}
	return out
}

func (cs *CloudServer) preActionCheck(label string) error {
	if cs.actions == nil {
		return fmt.Errorf("%s: this *CloudServer was not obtained via a client call (no action executor) — fetch via Get/Create/Update/List first", label)
	}
	if cs.CloudServerID() == "" {
		return fmt.Errorf("%s: missing cloud-server ID", label)
	}
	if cs.ProjectID() == "" {
		return fmt.Errorf("%s: missing project ID", label)
	}
	return nil
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (cs *CloudServer) toRequest() types.CloudServerRequest {
	props := types.CloudServerPropertiesRequest{}
	props.Zone = cs.Zone()
	if cs.vpcPreset != nil {
		props.VPCPreset = *cs.vpcPreset
	}
	if cs.flavor != nil {
		props.FlavorName = cs.flavor
	}
	if cs.userData != nil {
		v := *cs.userData
		props.UserData = &v
	}
	if cs.vpcRef != nil {
		props.VPC = types.ReferenceResourceCommon{URI: *cs.vpcRef}
	}
	if cs.bootVolumeRef != nil {
		props.BootVolume = &types.ReferenceResourceCommon{URI: *cs.bootVolumeRef}
	}
	if cs.keyPairRef != nil {
		props.KeyPair = &types.ReferenceResourceCommon{URI: *cs.keyPairRef}
	}
	if cs.elasticIPRef != nil {
		props.ElasticIP = &types.ReferenceResourceCommon{URI: *cs.elasticIPRef}
	}
	if len(cs.subnetRefs) > 0 {
		props.Subnets = make([]types.ReferenceResourceCommon, 0, len(cs.subnetRefs))
		for _, u := range cs.subnetRefs {
			props.Subnets = append(props.Subnets, types.ReferenceResourceCommon{URI: u})
		}
	}
	if len(cs.securityGroupRefs) > 0 {
		props.SecurityGroups = make([]types.ReferenceResourceCommon, 0, len(cs.securityGroupRefs))
		for _, u := range cs.securityGroupRefs {
			props.SecurityGroups = append(props.SecurityGroups, types.ReferenceResourceCommon{URI: u})
		}
	}
	props.BillingPlanCommon = &types.BillingPlanCommon{BillingPeriod: defaultBillingPeriod(cs.billingPeriod)}
	return types.CloudServerRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: cs.toMetadata(),
			Location:                cs.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (cs *CloudServer) fromResponse(resp *types.CloudServerResponse) {
	if resp == nil {
		return
	}
	cs.response = resp
	cs.setMeta(&resp.Metadata)
	cs.named(cloudServerDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		cs.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		cs.inRegion(resp.Metadata.LocationResponse.Value)
	}
	cs.setLinked(resp.Properties.LinkedResources)
	cs.setStatus(&resp.Status)

	if resp.Properties.Zone != "" {
		v := resp.Properties.Zone
		cs.zone = &v
	}
	if resp.Properties.Flavor.Name != "" {
		name := resp.Properties.Flavor.Name
		cs.flavor = &name
	}
	if resp.Properties.VPC.URI != "" {
		v := resp.Properties.VPC.URI
		cs.vpcRef = &v
	}
	if resp.Properties.BootVolume.URI != "" {
		v := resp.Properties.BootVolume.URI
		cs.bootVolumeRef = &v
	}
	if resp.Properties.KeyPair.URI != "" {
		v := resp.Properties.KeyPair.URI
		cs.keyPairRef = &v
	}
	if resp.Properties.BillingPlanCommon != nil && resp.Properties.BillingPlanCommon.BillingPeriod != nil {
		cs.billingPeriod = resp.Properties.BillingPlanCommon.BillingPeriod
	}
	if len(resp.Properties.NetworkInterfaces) > 0 {
		cs.subnetRefs = cs.subnetRefs[:0]
		for _, ni := range resp.Properties.NetworkInterfaces {
			if ni.Subnet != nil && *ni.Subnet != "" {
				cs.subnetRefs = append(cs.subnetRefs, *ni.Subnet)
			}
		}
	}

	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		cs.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if cs.projectID == "" && cs.RespURI() != "" {
		ids := parseURIIDs(cs.RespURI())
		cs.projectID = ids["projects"]
	}
}

func cloudServerDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// cloudServerActions is an internal interface satisfied by cloudServersClientAdapter. It
// allows *CloudServer to dispatch PowerOn/PowerOff/SetPassword without leaking the adapter
// into the public API.
type cloudServerActions interface {
	powerOn(ctx context.Context, projectID, cloudServerID string, rp *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	powerOff(ctx context.Context, projectID, cloudServerID string, rp *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	setPassword(ctx context.Context, projectID, cloudServerID, password string, rp *types.RequestParameters) (*types.Response[any], error)
}

// cloudServerLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type cloudServerLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.CloudServerListResponse], error)
	Get(ctx context.Context, projectID, cloudServerID string, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	Create(ctx context.Context, projectID string, body types.CloudServerRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	Update(ctx context.Context, projectID, cloudServerID string, body types.CloudServerRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	Delete(ctx context.Context, projectID, cloudServerID string, params *types.RequestParameters) (*types.Response[any], error)
	PowerOn(ctx context.Context, projectID, cloudServerID string, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	PowerOff(ctx context.Context, projectID, cloudServerID string, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	SetPassword(ctx context.Context, projectID, cloudServerID string, body types.CloudServerPasswordRequest, params *types.RequestParameters) (*types.Response[any], error)
	AssociateSubnets(ctx context.Context, projectID, cloudServerID string, body types.CloudServerAssociateSubnetsRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	AssociateSecurityGroups(ctx context.Context, projectID, cloudServerID string, body types.CloudServerAssociateSecurityGroupsRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	AssociateElasticIPs(ctx context.Context, projectID, cloudServerID string, body types.CloudServerAssociateElasticIPsRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	AttachDetachDataVolumes(ctx context.Context, projectID, cloudServerID string, body types.CloudServerAttachDetachDataVolumesRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
}

// ---- Adapter ----

// cloudServersClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates CloudServer ↔ types.CloudServerRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type cloudServersClientAdapter struct {
	low  cloudServerLowLevelClient
	rest *restclient.Client
}

var _ cloudServerActions = (*cloudServersClientAdapter)(nil)
var _ CloudServersClient = (*cloudServersClientAdapter)(nil)

func newCloudServersClientAdapter(rest *restclient.Client) *cloudServersClientAdapter {
	if rest == nil {
		return &cloudServersClientAdapter{}
	}
	return &cloudServersClientAdapter{low: compute.NewCloudServersClientImpl(rest), rest: rest}
}

// Create posts a new CloudServer to the API and hydrates the wrapper from the response.
func (a *cloudServersClientAdapter) Create(ctx context.Context, cs *CloudServer, opts ...CallOption) (*CloudServer, error) {
	if err := cs.Err(); err != nil {
		return cs, err
	}
	if cs.ProjectID() == "" {
		return cs, fmt.Errorf("Create: CloudServer has no parent project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, cs.ProjectID(), cs.toRequest(), rp)
	populateHTTPEnvelope(&cs.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		cs.fromResponse(resp.Data)
		cs.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, cs)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				cs.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	cs.actions = a
	if err != nil {
		return cs, err
	}
	if resp != nil && !resp.IsSuccess() {
		return cs, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return cs, nil
}

// Update dispatches to one or more API endpoints depending on what changed on the wrapper:
//   - name/tags differ from the hydrated response → PUT /cloudServers/:id
//   - AssociateSubnets/DisassociateSubnets called   → POST …/associateDisassociateSubnets
//   - AssociateSecurityGroups/Disassociate…         → POST …/associateDisassociateSecurityGroups
//   - AssociateElasticIPs/Disassociate…             → POST …/associateDisassociateElasticIPs
//   - AttachDataVolumes/DetachDataVolumes called     → POST …/attachDetachDataVolumes
//
// Calls are sequential. If a sub-call fails, the wrapper reflects the state after the last
// successful call. Delta queues are cleared on success.
func (a *cloudServersClientAdapter) Update(ctx context.Context, cs *CloudServer, opts ...CallOption) (*CloudServer, error) {
	if err := cs.Err(); err != nil {
		return cs, err
	}
	if cs.CloudServerID() == "" {
		return cs, fmt.Errorf("Update: CloudServer has no ID")
	}
	if cs.ProjectID() == "" {
		return cs, fmt.Errorf("Update: CloudServer has no parent project — call InProject first")
	}

	co := applyCallOptions(opts)
	rp := co.toRequestParameters()

	// 1. Metadata (name / tags) via PUT — only when something actually changed.
	if cs.hasMetadataChanges() {
		resp, err := a.low.Update(ctx, cs.ProjectID(), cs.CloudServerID(), cs.toRequest(), rp)
		if applyErr := a.applyUpdateResponse(cs, resp, err); applyErr != nil {
			return cs, applyErr
		}
	}

	// 2. Subnet associations.
	if len(cs.subnetsToAssociate) > 0 || len(cs.subnetsToDisassociate) > 0 {
		body := types.CloudServerAssociateSubnetsRequest{
			SubnetsToAssociate:    cloudServerStringsToCommon(cs.subnetsToAssociate),
			SubnetsToDisassociate: cloudServerStringsToCommon(cs.subnetsToDisassociate),
		}
		resp, err := a.low.AssociateSubnets(ctx, cs.ProjectID(), cs.CloudServerID(), body, rp)
		if applyErr := a.applyUpdateResponse(cs, resp, err); applyErr != nil {
			return cs, applyErr
		}
		cs.subnetsToAssociate = nil
		cs.subnetsToDisassociate = nil
	}

	// 3. Security group associations.
	if len(cs.sgsToAssociate) > 0 || len(cs.sgsToDisassociate) > 0 {
		body := types.CloudServerAssociateSecurityGroupsRequest{
			SecurityGroupsToAssociate:    cloudServerStringsToCommon(cs.sgsToAssociate),
			SecurityGroupsToDisassociate: cloudServerStringsToCommon(cs.sgsToDisassociate),
		}
		resp, err := a.low.AssociateSecurityGroups(ctx, cs.ProjectID(), cs.CloudServerID(), body, rp)
		if applyErr := a.applyUpdateResponse(cs, resp, err); applyErr != nil {
			return cs, applyErr
		}
		cs.sgsToAssociate = nil
		cs.sgsToDisassociate = nil
	}

	// 4. Elastic IP associations.
	if len(cs.eipsToAssociate) > 0 || len(cs.eipsToDisassociate) > 0 {
		body := types.CloudServerAssociateElasticIPsRequest{
			ElasticIPsToAssociate:    cloudServerStringsToCommon(cs.eipsToAssociate),
			ElasticIPsToDisassociate: cloudServerStringsToCommon(cs.eipsToDisassociate),
		}
		resp, err := a.low.AssociateElasticIPs(ctx, cs.ProjectID(), cs.CloudServerID(), body, rp)
		if applyErr := a.applyUpdateResponse(cs, resp, err); applyErr != nil {
			return cs, applyErr
		}
		cs.eipsToAssociate = nil
		cs.eipsToDisassociate = nil
	}

	// 5. Data volume attachments.
	if len(cs.dataVolumesToAttach) > 0 || len(cs.dataVolumesToDetach) > 0 {
		body := types.CloudServerAttachDetachDataVolumesRequest{
			VolumesToAttach: cloudServerStringsToCommon(cs.dataVolumesToAttach),
			VolumesToDetach: cloudServerStringsToCommon(cs.dataVolumesToDetach),
		}
		resp, err := a.low.AttachDetachDataVolumes(ctx, cs.ProjectID(), cs.CloudServerID(), body, rp)
		if applyErr := a.applyUpdateResponse(cs, resp, err); applyErr != nil {
			return cs, applyErr
		}
		cs.dataVolumesToAttach = nil
		cs.dataVolumesToDetach = nil
	}

	cs.actions = a
	if cs.response != nil {
		cs.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, cs)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				cs.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	return cs, nil
}

// applyUpdateResponse populates the HTTP envelope, hydrates the wrapper from a non-nil
// response payload, sets the action executor, and surfaces any transport or HTTP error.
func (a *cloudServersClientAdapter) applyUpdateResponse(cs *CloudServer, resp *types.Response[types.CloudServerResponse], err error) error {
	populateHTTPEnvelope(&cs.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		cs.fromResponse(resp.Data)
	}
	cs.actions = a
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// Get fetches a CloudServer by Ref and returns a freshly hydrated wrapper.
func (a *cloudServersClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*CloudServer, error) {
	projectID, cloudServerID, err := cloudServerIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, cloudServerID, rp)
	out := &CloudServer{}
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
	out.actions = a
	if err != nil {
		return out, err
	}
	if resp != nil && !resp.IsSuccess() {
		return out, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return out, nil
}

// Delete removes the CloudServer identified by Ref.
func (a *cloudServersClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, cloudServerID, err := cloudServerIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, cloudServerID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of CloudServer in the given parent scope.
func (a *cloudServersClientAdapter) List(ctx context.Context, project Ref, opts ...CallOption) (*List[*CloudServer], error) {
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
	var items []*CloudServer
	if resp != nil && resp.Data != nil {
		items = make([]*CloudServer, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			cs := &CloudServer{}
			cs.fromResponse(&resp.Data.Values[i])
			cs.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, cs)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					cs.fromResponse(fresh.Raw())
				}
				return nil
			})
			if cs.projectID == "" {
				cs.projectID = projectID
			}
			cs.actions = a
			items = append(items, cs)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*CloudServer], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*CloudServer], error) {
		fetch := listPageFetch[types.CloudServerListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*CloudServer
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*CloudServer, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				cs := &CloudServer{}
				cs.fromResponse(&pageResp.Data.Values[i])
				cs.setRefresh(func(ctx context.Context) error {
					fresh, err := a.Get(ctx, cs)
					if err != nil {
						return err
					}
					if fresh != nil && fresh.Raw() != nil {
						cs.fromResponse(fresh.Raw())
					}
					return nil
				})
				if cs.projectID == "" {
					cs.projectID = projectID
				}
				cs.actions = a
				pageItems = append(pageItems, cs)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}

// Internal action methods — satisfy cloudServerActions; called by *CloudServer action methods.

// powerOn sends a power-on action to the API for the given server.
func (a *cloudServersClientAdapter) powerOn(ctx context.Context, projectID, cloudServerID string, rp *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	return a.low.PowerOn(ctx, projectID, cloudServerID, rp)
}

// powerOff sends a power-off action to the API for the given server.
func (a *cloudServersClientAdapter) powerOff(ctx context.Context, projectID, cloudServerID string, rp *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	return a.low.PowerOff(ctx, projectID, cloudServerID, rp)
}

// setPassword sends a set-password action to the API for the given server.
func (a *cloudServersClientAdapter) setPassword(ctx context.Context, projectID, cloudServerID, password string, rp *types.RequestParameters) (*types.Response[any], error) {
	return a.low.SetPassword(ctx, projectID, cloudServerID, types.CloudServerPasswordRequest{Password: password}, rp)
}

// cloudServerIDsFromRef extracts (projectID, cloudServerID) from a Ref.
func cloudServerIDsFromRef(ref Ref) (projectID, cloudServerID string, err error) {
	csID, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withCloudServerID); ok {
			return w.CloudServerID(), true
		}
		return "", false
	}, "cloudServers")
	if !ok || csID == "" {
		return "", "", fmt.Errorf("cannot determine CloudServer ID from Ref %q", ref.URI())
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
	return pid, csID, nil
}
