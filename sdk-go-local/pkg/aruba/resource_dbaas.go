package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/database"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// DBaaS is the wrapper for an Aruba Cloud Database-as-a-Service instance
// (a direct child of a Project). Construct with aruba.NewDBaaS() and bind
// it via IntoProject(project), OfEngine, OfFlavor, WithSizeGB,
// WithVPC/WithSubnet/WithSecurityGroup/WithElasticIP, etc.
//
// Schema asymmetries:
//   - Engine: request emits Engine.ID; response returns a full
//     DBaaSEngineResponse{Type,Name,Version,...}. Engine() reads .Type
//     (the human-meaningful identifier on the response side).
//   - Flavor: request emits Flavor.Name; response returns the full
//     DBaaSFlavorResponse struct.
//   - Networking: request emits 4 raw URI strings (VPCURI/SubnetURI/
//     SecurityGroupURI/ElasticIPURI); response returns 4 *ReferenceResourceCommon
//     objects. Read-back getters (VPC/Subnet/SecurityGroup/ElasticIP)
//     prefer the response side, falling back to the locally-set URI.
//   - Zone: Go field is "Zone" but the wire JSON tag is "dataCenter".
//   - Autoscaling: request emits {Enabled,AvailableSpace,StepSize};
//     response returns {Status,AvailableSpace,StepSize,RuleID}.
//     AutoscalingEnabled() reads only the locally-set value;
//     AutoscalingStatus() / AutoscalingRuleID() read only the response;
//     AutoscalingAvailableSpace() / AutoscalingStepSize() prefer the
//     response and fall back to the locally-set value.
//     fromResponse back-populates autoscaling fields (AvailableSpace, StepSize,
//     and Enabled inferred from Status) so that a Get→Update round-trip preserves
//     the configured autoscaling block without the caller re-asserting it.
type DBaaS struct {
	errMixin
	metadataMixin
	zonalMixin
	projectScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	// Request-side scalars.
	engine                    *DatabaseEngine // wire: Engine.ID
	flavor                    *DBaaSFlavor    // wire: Flavor.Name
	sizeGB                    *int32          // wire: Storage.SizeGB
	autoscalingEnabled        *bool           // wire: Autoscaling.Enabled
	autoscalingAvailableSpace *int32          // wire: Autoscaling.AvailableSpace
	autoscalingStepSize       *int32          // wire: Autoscaling.StepSize
	billingPeriod             *BillingPeriod  // wire: BillingPlanCommon.BillingPeriod

	// Networking refs.
	vpcRef           *string
	subnetRef        *string
	securityGroupRef *string
	elasticIPRef     *string

	// Hydrated response.
	response *types.DBaaSResponse
}

// NewDBaaS returns a fresh *DBaaS ready for fluent setters and a Create call.
// Binds projectScopedMixin's error sink so IntoProject failures surface via Err().
func NewDBaaS() *DBaaS {
	d := &DBaaS{}
	d.projectScopedMixin = bindProjectScoped(&d.errMixin)
	return d
}

// Setters — chainable, general → specific

// InProject binds this DBaaS to its parent project. Required before Create.
func (d *DBaaS) InProject(p Ref) *DBaaS { d.intoProject(p); return d }

// Named sets the resource name. Required by the API.
func (d *DBaaS) Named(n string) *DBaaS { d.named(n); return d }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (d *DBaaS) Tagged(ts ...string) *DBaaS {
	for _, t := range ts {
		d.addTag(t)
	}
	return d
}

// Untagged removes each listed tag. No-op for tags not present.
func (d *DBaaS) Untagged(ts ...string) *DBaaS {
	for _, t := range ts {
		d.removeTag(t)
	}
	return d
}

// RetaggedAs replaces the entire tag set with the given values.
func (d *DBaaS) RetaggedAs(ts ...string) *DBaaS { d.replaceTags(ts...); return d }

// InRegion sets the region for this resource.
func (d *DBaaS) InRegion(region Region) *DBaaS { d.inRegion(region); return d }

// InZone sets the availability zone (wire field: dataCenter) for this resource.
func (d *DBaaS) InZone(zone Zone) *DBaaS { d.inZone(zone); return d }

// OfEngine sets the database engine. The wire request emits Engine.ID.
func (d *DBaaS) OfEngine(engine DatabaseEngine) *DBaaS { d.engine = &engine; return d }

// OfFlavor sets the database flavor (size/performance tier). The wire request emits Flavor.Name.
func (d *DBaaS) OfFlavor(flavor DBaaSFlavor) *DBaaS { d.flavor = &flavor; return d }

// SizedGB sets the storage size in GB. The wire request emits Storage.SizeGB.
func (d *DBaaS) SizedGB(gb int) *DBaaS { v := int32(gb); d.sizeGB = &v; return d }

// WithAutoscaling enables autoscaling and pins the available-space threshold and
// step size in GB. Mirrors NodePool.WithAutoscaling(min, max) from resource_kaas_nodepool.go.
func (d *DBaaS) WithAutoscaling(availableSpaceGB, stepSizeGB int) *DBaaS {
	t := true
	av := int32(availableSpaceGB)
	ss := int32(stepSizeGB)
	d.autoscalingEnabled = &t
	d.autoscalingAvailableSpace = &av
	d.autoscalingStepSize = &ss
	return d
}

// WithoutAutoscaling explicitly disables autoscaling and clears the bounds.
func (d *DBaaS) WithoutAutoscaling() *DBaaS {
	f := false
	d.autoscalingEnabled = &f
	d.autoscalingAvailableSpace = nil
	d.autoscalingStepSize = nil
	return d
}

// BilledBy sets the billing cadence. Accepted periods are resource-specific; check the API reference.
func (d *DBaaS) BilledBy(period BillingPeriod) *DBaaS { d.billingPeriod = &period; return d }

// WithVPC sets the VPC for this DBaaS via its URI. Wire field: VPCURI.
func (d *DBaaS) WithVPC(v Ref) *DBaaS { return d.setSingleRef("WithVPC", v, &d.vpcRef) }

// WithSubnet sets the subnet for this DBaaS via its URI. Wire field: SubnetURI.
func (d *DBaaS) WithSubnet(s Ref) *DBaaS { return d.setSingleRef("WithSubnet", s, &d.subnetRef) }

// WithSecurityGroup sets the security group for this DBaaS via its URI. Wire field: SecurityGroupURI.
func (d *DBaaS) WithSecurityGroup(sg Ref) *DBaaS {
	return d.setSingleRef("WithSecurityGroup", sg, &d.securityGroupRef)
}

// WithElasticIP sets the elastic IP for this DBaaS via its URI. Wire field: ElasticIPURI.
func (d *DBaaS) WithElasticIP(eip Ref) *DBaaS {
	return d.setSingleRef("WithElasticIP", eip, &d.elasticIPRef)
}

func (d *DBaaS) setSingleRef(label string, r Ref, dst **string) *DBaaS {
	uri := r.URI()
	if uri == "" {
		d.addErr(fmt.Errorf("%s: empty URI", label))
		return d
	}
	*dst = &uri
	return d
}

// Getters — general → specific

// Ref + ID accessors.

// URI satisfies Ref by returning the server-assigned canonical URI, or "" if Create hasn't run yet.
func (d *DBaaS) URI() string { return d.RespURI() }

// DBaaSID satisfies withDBaaSID so child wrappers can extract this ID by typed assertion.
func (d *DBaaS) DBaaSID() string { return d.ID() }

// Accessors.

// Raw shadows responseMetadataMixin.Raw() with the typed DBaaS response.
func (d *DBaaS) Raw() *types.DBaaSResponse { return d.response }
func (d *DBaaS) RawJSON() []byte           { return marshalRawJSON(d.response) }
func (d *DBaaS) RawYAML() []byte           { return marshalRawYAML(d.response) }

// RawRequest returns what toRequest() would emit right now.
func (d *DBaaS) RawRequest() types.DBaaSRequest { return d.toRequest() }

// Engine returns the engine identifier. On a hydrated response the value comes
// from Engine.Type; before hydration it returns what was passed to OfEngine.
func (d *DBaaS) Engine() DatabaseEngine {
	if d.response != nil && d.response.Properties.Engine != nil && d.response.Properties.Engine.Type != nil {
		return DatabaseEngine(*d.response.Properties.Engine.Type)
	}
	if d.engine == nil {
		return ""
	}
	return *d.engine
}

// EngineRaw returns the full engine struct from the last response, or nil.
func (d *DBaaS) EngineRaw() *types.DBaaSEngineResponse {
	if d.response == nil {
		return nil
	}
	return d.response.Properties.Engine
}

// Flavor returns the flavor name. On a hydrated response the value comes from
// Flavor.Name; before hydration it returns what was passed to OfFlavor.
func (d *DBaaS) Flavor() DBaaSFlavor {
	if d.response != nil && d.response.Properties.Flavor != nil && d.response.Properties.Flavor.Name != nil {
		return DBaaSFlavor(*d.response.Properties.Flavor.Name)
	}
	if d.flavor == nil {
		return ""
	}
	return *d.flavor
}

// FlavorRaw returns the full flavor struct from the last response, or nil.
func (d *DBaaS) FlavorRaw() *types.DBaaSFlavorResponse {
	if d.response == nil {
		return nil
	}
	return d.response.Properties.Flavor
}

// SizeGB returns the storage size in GB. On a hydrated response the value comes
// from Storage.SizeGB; before hydration it returns what was passed to WithSizeGB.
func (d *DBaaS) SizeGB() int {
	if d.response != nil && d.response.Properties.Storage != nil && d.response.Properties.Storage.SizeGB != nil {
		return int(*d.response.Properties.Storage.SizeGB)
	}
	if d.sizeGB != nil {
		return int(*d.sizeGB)
	}
	return 0
}

// BillingPeriod returns the billing period. On a hydrated response the value comes
// from BillingPlanCommon.BillingPeriod; before hydration it returns what was passed to
// WithBillingPeriod.
func (d *DBaaS) BillingPeriod() BillingPeriod {
	if d.response != nil && d.response.Properties.BillingPlanCommon != nil && d.response.Properties.BillingPlanCommon.BillingPeriod != nil {
		return *d.response.Properties.BillingPlanCommon.BillingPeriod
	}
	if d.billingPeriod == nil {
		return ""
	}
	return *d.billingPeriod
}

// AutoscalingEnabled returns the locally-set Enabled flag (request-side intent).
// The response side carries no Enabled field — see AutoscalingStatus() for the
// platform-reported state.
func (d *DBaaS) AutoscalingEnabled() bool {
	if d.autoscalingEnabled != nil {
		return *d.autoscalingEnabled
	}
	return false
}

// AutoscalingStatus returns the response-side autoscaling status string.
// Empty before hydration.
func (d *DBaaS) AutoscalingStatus() string {
	if d.response != nil && d.response.Properties.Autoscaling != nil &&
		d.response.Properties.Autoscaling.Status != nil {
		return *d.response.Properties.Autoscaling.Status
	}
	return ""
}

// AutoscalingAvailableSpaceGB returns the available-space threshold in GB.
// Hydrated response wins; otherwise returns the locally-set value, else 0.
func (d *DBaaS) AutoscalingAvailableSpaceGB() int {
	if d.response != nil && d.response.Properties.Autoscaling != nil &&
		d.response.Properties.Autoscaling.AvailableSpace != nil {
		return int(*d.response.Properties.Autoscaling.AvailableSpace)
	}
	if d.autoscalingAvailableSpace != nil {
		return int(*d.autoscalingAvailableSpace)
	}
	return 0
}

// AutoscalingStepSizeGB returns the step size in GB.
// Hydrated response wins; otherwise returns the locally-set value, else 0.
func (d *DBaaS) AutoscalingStepSizeGB() int {
	if d.response != nil && d.response.Properties.Autoscaling != nil &&
		d.response.Properties.Autoscaling.StepSize != nil {
		return int(*d.response.Properties.Autoscaling.StepSize)
	}
	if d.autoscalingStepSize != nil {
		return int(*d.autoscalingStepSize)
	}
	return 0
}

// AutoscalingRuleID returns the response-side rule identifier.
// Empty before hydration.
func (d *DBaaS) AutoscalingRuleID() string {
	if d.response != nil && d.response.Properties.Autoscaling != nil &&
		d.response.Properties.Autoscaling.RuleID != nil {
		return *d.response.Properties.Autoscaling.RuleID
	}
	return ""
}

// AutoscalingRaw returns the full autoscaling response struct, or nil before hydration.
func (d *DBaaS) AutoscalingRaw() *types.DBaaSAutoscalingResponse {
	if d.response == nil {
		return nil
	}
	return d.response.Properties.Autoscaling
}

// EngineVersion returns the database engine version string from the response, or "" before hydration.
func (d *DBaaS) EngineVersion() string {
	if d.response != nil && d.response.Properties.Engine != nil && d.response.Properties.Engine.Version != nil {
		return *d.response.Properties.Engine.Version
	}
	return ""
}

// EngineType returns the database engine type string from the response.
// On a hydrated response it reads Engine.Type; before hydration it returns what was passed to OfEngine.
func (d *DBaaS) EngineType() string {
	return string(d.Engine())
}

// PrivateIPAddress returns the private IP address from the engine response, or "" before hydration.
func (d *DBaaS) PrivateIPAddress() string {
	if d.response != nil && d.response.Properties.Engine != nil && d.response.Properties.Engine.PrivateIPAddress != nil {
		return *d.response.Properties.Engine.PrivateIPAddress
	}
	return ""
}

// FlavorCPU returns the number of CPUs from the flavor response, or 0 before hydration.
func (d *DBaaS) FlavorCPU() int32 {
	if d.response != nil && d.response.Properties.Flavor != nil && d.response.Properties.Flavor.CPU != nil {
		return *d.response.Properties.Flavor.CPU
	}
	return 0
}

// FlavorRAMMB returns the amount of RAM in MB from the flavor response, or 0 before hydration.
func (d *DBaaS) FlavorRAMMB() int32 {
	if d.response != nil && d.response.Properties.Flavor != nil && d.response.Properties.Flavor.RAM != nil {
		return *d.response.Properties.Flavor.RAM
	}
	return 0
}

// VPC returns the VPC URI for this DBaaS instance, or "" if unset.
func (d *DBaaS) VPC() string {
	return dbaasNetworkingURI(d.response, func(n *types.DBaaSNetworkingResponse) *types.ReferenceResourceCommon { return n.VPC }, d.vpcRef)
}

// Subnet returns the subnet URI for this DBaaS instance, or "" if unset.
func (d *DBaaS) Subnet() string {
	return dbaasNetworkingURI(d.response, func(n *types.DBaaSNetworkingResponse) *types.ReferenceResourceCommon { return n.Subnet }, d.subnetRef)
}

// SecurityGroup returns the security group URI for this DBaaS instance, or "" if unset.
func (d *DBaaS) SecurityGroup() string {
	return dbaasNetworkingURI(d.response, func(n *types.DBaaSNetworkingResponse) *types.ReferenceResourceCommon { return n.SecurityGroup }, d.securityGroupRef)
}

// ElasticIP returns the elastic IP URI for this DBaaS instance, or "" if unset.
func (d *DBaaS) ElasticIP() string {
	return dbaasNetworkingURI(d.response, func(n *types.DBaaSNetworkingResponse) *types.ReferenceResourceCommon { return n.ElasticIP }, d.elasticIPRef)
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (d *DBaaS) toRequest() types.DBaaSRequest {
	props := types.DBaaSPropertiesRequest{}
	props.Zone = d.zonePtr()
	if d.engine != nil {
		props.Engine = &types.DBaaSEngineRequest{ID: d.engine}
	}
	if d.flavor != nil {
		props.Flavor = &types.DBaaSFlavorRequest{Name: d.flavor}
	}
	if d.sizeGB != nil {
		v := *d.sizeGB
		props.Storage = &types.DBaaSStorageRequest{SizeGB: &v}
	}
	if d.autoscalingEnabled != nil || d.autoscalingAvailableSpace != nil || d.autoscalingStepSize != nil {
		a := &types.DBaaSAutoscalingRequest{}
		if d.autoscalingEnabled != nil {
			v := *d.autoscalingEnabled
			a.Enabled = &v
		}
		if d.autoscalingAvailableSpace != nil {
			v := *d.autoscalingAvailableSpace
			a.AvailableSpace = &v
		}
		if d.autoscalingStepSize != nil {
			v := *d.autoscalingStepSize
			a.StepSize = &v
		}
		props.Autoscaling = a
	}
	props.BillingPlanCommon = &types.BillingPlanCommon{BillingPeriod: defaultBillingPeriod(d.billingPeriod)}
	if d.vpcRef != nil || d.subnetRef != nil || d.securityGroupRef != nil || d.elasticIPRef != nil {
		net := &types.DBaaSNetworkingRequest{}
		if d.vpcRef != nil {
			net.VPCURI = d.vpcRef
		}
		if d.subnetRef != nil {
			net.SubnetURI = d.subnetRef
		}
		if d.securityGroupRef != nil {
			net.SecurityGroupURI = d.securityGroupRef
		}
		if d.elasticIPRef != nil {
			net.ElasticIPURI = d.elasticIPRef
		}
		props.Networking = net
	}
	return types.DBaaSRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: d.toMetadata(),
			Location:                d.toLocation(),
		},
		Properties: props,
	}
}

// Autoscaling status wire values from the DBaaS Autoscaling response.
// The API reports either "Enabled" or "Active" for an enabled autoscaling
// configuration; both are treated as enabled by the wrapper.
const (
	autoscalingStatusEnabled = "Enabled"
	autoscalingStatusActive  = "Active"
)

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (d *DBaaS) fromResponse(resp *types.DBaaSResponse) {
	if resp == nil {
		return
	}
	d.response = resp
	d.setMeta(&resp.Metadata)
	d.named(dbaasDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		d.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		d.inRegion(resp.Metadata.LocationResponse.Value)
	}
	d.setLinked(resp.Properties.LinkedResources)
	d.setStatus(&resp.Status)

	// Hydrate request-side fields from response for round-trip Update support.
	if resp.Properties.Engine != nil && resp.Properties.Engine.Type != nil {
		e := DatabaseEngine(*resp.Properties.Engine.Type)
		d.engine = &e
	}
	if resp.Properties.Flavor != nil && resp.Properties.Flavor.Name != nil {
		f := DBaaSFlavor(*resp.Properties.Flavor.Name)
		d.flavor = &f
	}
	if resp.Properties.Storage != nil && resp.Properties.Storage.SizeGB != nil {
		v := *resp.Properties.Storage.SizeGB
		d.sizeGB = &v
	}
	if resp.Properties.BillingPlanCommon != nil && resp.Properties.BillingPlanCommon.BillingPeriod != nil {
		d.billingPeriod = resp.Properties.BillingPlanCommon.BillingPeriod
	}
	if resp.Properties.Networking != nil {
		if resp.Properties.Networking.VPC != nil && resp.Properties.Networking.VPC.URI != "" {
			v := resp.Properties.Networking.VPC.URI
			d.vpcRef = &v
		}
		if resp.Properties.Networking.Subnet != nil && resp.Properties.Networking.Subnet.URI != "" {
			v := resp.Properties.Networking.Subnet.URI
			d.subnetRef = &v
		}
		if resp.Properties.Networking.SecurityGroup != nil && resp.Properties.Networking.SecurityGroup.URI != "" {
			v := resp.Properties.Networking.SecurityGroup.URI
			d.securityGroupRef = &v
		}
		if resp.Properties.Networking.ElasticIP != nil && resp.Properties.Networking.ElasticIP.URI != "" {
			v := resp.Properties.Networking.ElasticIP.URI
			d.elasticIPRef = &v
		}
	}

	// Back-populate autoscaling from response so Get→Update round-trips preserve
	// the configuration even when the caller does not call WithAutoscaling again.
	// The response type carries Status (not Enabled), so we infer the bool from it.
	if resp.Properties.Autoscaling != nil {
		as := resp.Properties.Autoscaling
		if as.AvailableSpace != nil {
			v := *as.AvailableSpace
			d.autoscalingAvailableSpace = &v
		}
		if as.StepSize != nil {
			v := *as.StepSize
			d.autoscalingStepSize = &v
		}
		if as.Status != nil {
			enabled := *as.Status == autoscalingStatusEnabled || *as.Status == autoscalingStatusActive
			d.autoscalingEnabled = &enabled
		}
	}

	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		d.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if d.projectID == "" && d.RespURI() != "" {
		ids := parseURIIDs(d.RespURI())
		d.projectID = ids["projects"]
	}
}

func dbaasDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func dbaasNetworkingURI(resp *types.DBaaSResponse, pick func(*types.DBaaSNetworkingResponse) *types.ReferenceResourceCommon, fallback *string) string {
	if resp != nil && resp.Properties.Networking != nil {
		if r := pick(resp.Properties.Networking); r != nil && r.URI != "" {
			return r.URI
		}
	}
	return dbaasDerefString(fallback)
}

// ---- Low-level client interface ----

// dbaasLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type dbaasLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.DBaaSListResponse], error)
	Get(ctx context.Context, projectID, dbaasID string, params *types.RequestParameters) (*types.Response[types.DBaaSResponse], error)
	Create(ctx context.Context, projectID string, body types.DBaaSRequest, params *types.RequestParameters) (*types.Response[types.DBaaSResponse], error)
	Update(ctx context.Context, projectID, dbaasID string, body types.DBaaSRequest, params *types.RequestParameters) (*types.Response[types.DBaaSResponse], error)
	Delete(ctx context.Context, projectID, dbaasID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// dbaasClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates DBaaS ↔ types.DBaaSRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type dbaasClientAdapter struct {
	low  dbaasLowLevelClient
	rest *restclient.Client
}

var _ DBaaSClient = (*dbaasClientAdapter)(nil)

func newDBaaSClientAdapter(rest *restclient.Client) *dbaasClientAdapter {
	if rest == nil {
		return &dbaasClientAdapter{}
	}
	return &dbaasClientAdapter{low: database.NewDBaaSClientImpl(rest), rest: rest}
}

// Create posts a new DBaaS to the API and hydrates the wrapper from the response.
func (a *dbaasClientAdapter) Create(ctx context.Context, d *DBaaS, opts ...CallOption) (*DBaaS, error) {
	if err := d.Err(); err != nil {
		return d, err
	}
	if d.ProjectID() == "" {
		return d, fmt.Errorf("Create: DBaaS has no parent project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, d.ProjectID(), d.toRequest(), rp)
	populateHTTPEnvelope(&d.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		d.fromResponse(resp.Data)
		d.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, d)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				d.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return d, err
	}
	if resp != nil && !resp.IsSuccess() {
		return d, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return d, nil
}

// Update sends a PUT for the current wrapper state. Requires ID and parent.
func (a *dbaasClientAdapter) Update(ctx context.Context, d *DBaaS, opts ...CallOption) (*DBaaS, error) {
	if err := d.Err(); err != nil {
		return d, err
	}
	if d.DBaaSID() == "" {
		return d, fmt.Errorf("Update: DBaaS has no ID")
	}
	if d.ProjectID() == "" {
		return d, fmt.Errorf("Update: DBaaS has no parent project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, d.ProjectID(), d.DBaaSID(), d.toRequest(), rp)
	populateHTTPEnvelope(&d.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		d.fromResponse(resp.Data)
		d.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, d)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				d.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return d, err
	}
	if resp != nil && !resp.IsSuccess() {
		return d, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return d, nil
}

// Get fetches a DBaaS by Ref and returns a freshly hydrated wrapper.
func (a *dbaasClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*DBaaS, error) {
	projectID, dbaasID, err := dbaasIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, dbaasID, rp)
	out := &DBaaS{}
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

// Delete removes the DBaaS identified by Ref.
func (a *dbaasClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, dbaasID, err := dbaasIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, dbaasID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of DBaaS instances in the given parent scope.
func (a *dbaasClientAdapter) List(ctx context.Context, project Ref, opts ...CallOption) (*List[*DBaaS], error) {
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
	var items []*DBaaS
	if resp != nil && resp.Data != nil {
		items = make([]*DBaaS, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			d := &DBaaS{}
			d.fromResponse(&resp.Data.Values[i])
			d.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, d)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					d.fromResponse(fresh.Raw())
				}
				return nil
			})
			if d.projectID == "" {
				d.projectID = projectID
			}
			items = append(items, d)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*DBaaS], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*DBaaS], error) {
		fetch := listPageFetch[types.DBaaSListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*DBaaS
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*DBaaS, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				d := &DBaaS{}
				d.fromResponse(&pageResp.Data.Values[i])
				d.setRefresh(func(ctx context.Context) error {
					fresh, err := a.Get(ctx, d)
					if err != nil {
						return err
					}
					if fresh != nil && fresh.Raw() != nil {
						d.fromResponse(fresh.Raw())
					}
					return nil
				})
				if d.projectID == "" {
					d.projectID = projectID
				}
				pageItems = append(pageItems, d)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}

// dbaasIDsFromRef extracts (projectID, dbaasID) from a Ref.
func dbaasIDsFromRef(ref Ref) (projectID, dbaasID string, err error) {
	did, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withDBaaSID); ok {
			return w.DBaaSID(), true
		}
		return "", false
	}, "dbaas")
	if !ok || did == "" {
		return "", "", fmt.Errorf("cannot determine DBaaS ID from Ref %q", ref.URI())
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
	return pid, did, nil
}
