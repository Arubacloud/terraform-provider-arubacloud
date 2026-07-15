package aruba

import (
	"context"
	"fmt"
	"time"

	"github.com/Arubacloud/sdk-go/internal/clients/container"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// ContainerRegistry is the wrapper for an Aruba Cloud Container Registry
// (a direct child of a Project). Construct with aruba.NewContainerRegistry()
// and bind it via InProject(project), WithVPC(vpc), etc.
//
// Family A: regional, Metadata/Properties envelope, location-aware.
// Supports full CRUD including Update (same request shape as Create).
//
// Path: /projects/{projectID}/providers/Aruba.Container/registries[/{registryID}]
type ContainerRegistry struct {
	errMixin
	metadataMixin
	regionalMixin
	projectScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	// Body-refs (single).
	elasticIPRef     *string
	vpcRef           *string
	subnetRef        *string
	securityGroupRef *string
	blockStorageRef  *string

	// Registry-specific scalars.
	adminUsername   *string
	concurrentUsers *string // wire "size" — flavor enum string ("Small", "Medium", "HighPerf")
	billingPeriod   *BillingPeriod

	response *types.ContainerRegistryResponse
}

// NewContainerRegistry returns a fresh *ContainerRegistry ready for fluent setters and a Create call.
// Binds projectScopedMixin's error sink so InProject failures surface via Err().
func NewContainerRegistry() *ContainerRegistry {
	r := &ContainerRegistry{}
	r.projectScopedMixin = bindProjectScoped(&r.errMixin)
	return r
}

// Setters — chainable, general → specific

// Standard setters.

// InProject binds this ContainerRegistry to its parent project. Required before Create.
func (r *ContainerRegistry) InProject(p Ref) *ContainerRegistry { r.intoProject(p); return r }

// Named sets the resource name. Required by the API.
func (r *ContainerRegistry) Named(n string) *ContainerRegistry { r.named(n); return r }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (r *ContainerRegistry) Tagged(ts ...string) *ContainerRegistry {
	for _, t := range ts {
		r.addTag(t)
	}
	return r
}

// Untagged removes each listed tag. No-op for tags not present.
func (r *ContainerRegistry) Untagged(ts ...string) *ContainerRegistry {
	for _, t := range ts {
		r.removeTag(t)
	}
	return r
}

// RetaggedAs replaces the entire tag set with the given values.
func (r *ContainerRegistry) RetaggedAs(ts ...string) *ContainerRegistry {
	r.replaceTags(ts...)
	return r
}

// InRegion sets the region for this resource.
func (r *ContainerRegistry) InRegion(region Region) *ContainerRegistry {
	r.inRegion(region)
	return r
}

// Body-ref setters. Empty URIs are recorded on the error sink and the field
// remains unset.

// WithElasticIP binds the elastic IP to the registry. Errors if the URI is empty.
func (r *ContainerRegistry) WithElasticIP(eip Ref) *ContainerRegistry {
	return r.setSingleRef("WithElasticIP", eip, &r.elasticIPRef)
}

// WithVPC binds the registry to the given VPC. Errors if the URI is empty.
func (r *ContainerRegistry) WithVPC(v Ref) *ContainerRegistry {
	return r.setSingleRef("WithVPC", v, &r.vpcRef)
}

// WithSubnet binds the registry to the given subnet. Errors if the URI is empty.
func (r *ContainerRegistry) WithSubnet(s Ref) *ContainerRegistry {
	return r.setSingleRef("WithSubnet", s, &r.subnetRef)
}

// WithSecurityGroup binds the registry to the given security group. Errors if the URI is empty.
func (r *ContainerRegistry) WithSecurityGroup(sg Ref) *ContainerRegistry {
	return r.setSingleRef("WithSecurityGroup", sg, &r.securityGroupRef)
}

// WithBlockStorage binds a block storage volume for registry data. Errors if the URI is empty.
func (r *ContainerRegistry) WithBlockStorage(vol Ref) *ContainerRegistry {
	return r.setSingleRef("WithBlockStorage", vol, &r.blockStorageRef)
}

func (r *ContainerRegistry) setSingleRef(label string, ref Ref, dst **string) *ContainerRegistry {
	uri := ref.URI()
	if uri == "" {
		r.addErr(fmt.Errorf("%s: empty URI", label))
		return r
	}
	*dst = &uri
	return r
}

// Registry-specific scalar setters.

// WithAdminUsername sets the admin username for the registry.
func (r *ContainerRegistry) WithAdminUsername(u string) *ContainerRegistry {
	r.adminUsername = &u
	return r
}

// OfSize sets the concurrent-users tier for the registry.
// Accepted values per the platform: "Small", "Medium", "HighPerf".
// Use the ContainerRegistrySizeFlavor* constants.
func (r *ContainerRegistry) OfSize(flavor types.ContainerRegistrySizeFlavor) *ContainerRegistry {
	s := string(flavor)
	r.concurrentUsers = &s
	return r
}

// BilledBy sets the billing cadence. Accepted periods are resource-specific; check the API reference.
func (r *ContainerRegistry) BilledBy(period BillingPeriod) *ContainerRegistry {
	r.billingPeriod = &period
	return r
}

// Getters — general → specific

// Ref + ID accessors.

// URI satisfies Ref by returning the server-assigned canonical URI, or "" if Create hasn't run yet.
func (r *ContainerRegistry) URI() string { return r.RespURI() }

// ContainerRegistryID satisfies withContainerRegistryID so child wrappers can extract this ID by typed assertion.
func (r *ContainerRegistry) ContainerRegistryID() string { return r.ID() }

// Raw accessors.

// Raw shadows responseMetadataMixin.Raw() with the typed ContainerRegistry response.
func (r *ContainerRegistry) Raw() *types.ContainerRegistryResponse { return r.response }
func (r *ContainerRegistry) RawJSON() []byte                       { return marshalRawJSON(r.response) }
func (r *ContainerRegistry) RawYAML() []byte                       { return marshalRawYAML(r.response) }

// RawRequest returns what toRequest() would emit right now.
func (r *ContainerRegistry) RawRequest() types.ContainerRegistryRequest { return r.toRequest() }

// Response-preferring accessors (fall back to request-side field when not hydrated).

// ElasticIP returns the public endpoint URI for the registry (wire field: properties.publicIp.uri).
func (r *ContainerRegistry) ElasticIP() string {
	return r.responseURIField(func() string { return r.response.Properties.PublicIp.URI }, r.elasticIPRef)
}

// VPC returns the VPC URI for the registry, or "" if unset.
func (r *ContainerRegistry) VPC() string {
	return r.responseURIField(func() string { return r.response.Properties.VPC.URI }, r.vpcRef)
}

// Subnet returns the subnet URI for the registry, or "" if unset.
func (r *ContainerRegistry) Subnet() string {
	return r.responseURIField(func() string { return r.response.Properties.Subnet.URI }, r.subnetRef)
}

// SecurityGroup returns the security group URI for the registry, or "" if unset.
func (r *ContainerRegistry) SecurityGroup() string {
	return r.responseURIField(func() string { return r.response.Properties.SecurityGroup.URI }, r.securityGroupRef)
}

// BlockStorage returns the block storage URI attached to the registry, or "" if unset.
func (r *ContainerRegistry) BlockStorage() string {
	return r.responseURIField(func() string { return r.response.Properties.BlockStorage.URI }, r.blockStorageRef)
}
func (r *ContainerRegistry) responseURIField(fromResp func() string, fallback *string) string {
	if r.response != nil {
		if u := fromResp(); u != "" {
			return u
		}
	}
	return containerRegistryDeref(fallback)
}

// AdminUsername returns the admin username for the registry, or "" if unset.
func (r *ContainerRegistry) AdminUsername() string {
	if r.response != nil && r.response.Properties.AdminUser != nil {
		return r.response.Properties.AdminUser.Username
	}
	return containerRegistryDeref(r.adminUsername)
}

// IsPasswordSet reports whether the Aruba provisioner has generated the admin
// password for the registry. The password itself is never returned over the wire.
func (r *ContainerRegistry) IsPasswordSet() bool {
	if r.response != nil && r.response.Data != nil && r.response.Data.Private != nil && r.response.Data.Private.PasswordSet != nil {
		return *r.response.Data.Private.PasswordSet
	}
	return false
}

// AdminPasswordLastSetAt returns when the Aruba provisioner last generated the
// admin password, or the zero time if that information is unavailable.
func (r *ContainerRegistry) AdminPasswordLastSetAt() time.Time {
	if r.response != nil && r.response.Data != nil && r.response.Data.Private != nil && r.response.Data.Private.PasswordLastSetAt != nil {
		return *r.response.Data.Private.PasswordLastSetAt
	}
	return time.Time{}
}

// FQDN returns the fully-qualified domain name for the registry, or "" if unset.
func (r *ContainerRegistry) FQDN() string {
	if r.response != nil && r.response.Data != nil && r.response.Data.Info != nil {
		return containerRegistryDeref(r.response.Data.Info.FQDN)
	}
	return ""
}

// PublicBaseURL returns the public base URL for the registry, or "" if unset.
func (r *ContainerRegistry) PublicBaseURL() string {
	if r.response != nil && r.response.Data != nil && r.response.Data.Info != nil {
		return containerRegistryDeref(r.response.Data.Info.PublicBaseURL)
	}
	return ""
}

// PrivateBaseURL returns the private (VPC-internal) base URL for the registry, or "" if unset.
func (r *ContainerRegistry) PrivateBaseURL() string {
	if r.response != nil && r.response.Data != nil && r.response.Data.Info != nil {
		return containerRegistryDeref(r.response.Data.Info.PrivateBaseURL)
	}
	return ""
}

// Version returns the registry software version, or "" if unset.
func (r *ContainerRegistry) Version() string {
	if r.response != nil && r.response.Data != nil && r.response.Data.Info != nil {
		return containerRegistryDeref(r.response.Data.Info.Version)
	}
	return ""
}

// SizeFlavor returns the registry's concurrent-users tier as the typed enum.
// Returns "" if the wire field has not been populated.
func (r *ContainerRegistry) SizeFlavor() types.ContainerRegistrySizeFlavor {
	if r.response != nil && r.response.Properties.ConcurrentUsers != nil {
		return types.ContainerRegistrySizeFlavor(*r.response.Properties.ConcurrentUsers)
	}
	if r.concurrentUsers != nil {
		return types.ContainerRegistrySizeFlavor(*r.concurrentUsers)
	}
	return ""
}

// BillingPeriod returns the billing period for the registry, or "" if unset.
func (r *ContainerRegistry) BillingPeriod() BillingPeriod {
	if r.response != nil && r.response.Properties.BillingPlanCommon != nil && r.response.Properties.BillingPlanCommon.BillingPeriod != nil {
		return *r.response.Properties.BillingPlanCommon.BillingPeriod
	}
	if r.billingPeriod == nil {
		return ""
	}
	return *r.billingPeriod
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (r *ContainerRegistry) toRequest() types.ContainerRegistryRequest {
	props := types.ContainerRegistryPropertiesRequest{}
	if r.elasticIPRef != nil {
		props.PublicIp = types.ReferenceResourceCommon{URI: *r.elasticIPRef}
	}
	if r.vpcRef != nil {
		props.VPC = types.ReferenceResourceCommon{URI: *r.vpcRef}
	}
	if r.subnetRef != nil {
		props.Subnet = types.ReferenceResourceCommon{URI: *r.subnetRef}
	}
	if r.securityGroupRef != nil {
		props.SecurityGroup = types.ReferenceResourceCommon{URI: *r.securityGroupRef}
	}
	if r.blockStorageRef != nil {
		props.BlockStorage = types.ReferenceResourceCommon{URI: *r.blockStorageRef}
	}
	if r.adminUsername != nil {
		props.AdminUser = &types.UserCredentialCommon{Username: *r.adminUsername}
	}
	if r.concurrentUsers != nil {
		props.ConcurrentUsers = r.concurrentUsers
	}
	props.BillingPlanCommon = &types.BillingPlanCommon{BillingPeriod: defaultBillingPeriod(r.billingPeriod)}
	return types.ContainerRegistryRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: r.toMetadata(),
			Location:                r.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (r *ContainerRegistry) fromResponse(resp *types.ContainerRegistryResponse) {
	if resp == nil {
		return
	}
	r.response = resp
	r.setMeta(&resp.Metadata)
	r.named(containerRegistryDeref(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		r.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		r.inRegion(resp.Metadata.LocationResponse.Value)
	}
	r.setStatus(&resp.Status)

	if resp.Properties.PublicIp.URI != "" {
		v := resp.Properties.PublicIp.URI
		r.elasticIPRef = &v
	}
	if resp.Properties.VPC.URI != "" {
		v := resp.Properties.VPC.URI
		r.vpcRef = &v
	}
	if resp.Properties.Subnet.URI != "" {
		v := resp.Properties.Subnet.URI
		r.subnetRef = &v
	}
	if resp.Properties.SecurityGroup.URI != "" {
		v := resp.Properties.SecurityGroup.URI
		r.securityGroupRef = &v
	}
	if resp.Properties.BlockStorage.URI != "" {
		v := resp.Properties.BlockStorage.URI
		r.blockStorageRef = &v
	}
	if resp.Properties.AdminUser != nil && resp.Properties.AdminUser.Username != "" {
		v := resp.Properties.AdminUser.Username
		r.adminUsername = &v
	}
	if resp.Properties.ConcurrentUsers != nil && *resp.Properties.ConcurrentUsers != "" {
		v := *resp.Properties.ConcurrentUsers
		r.concurrentUsers = &v
	}
	if resp.Properties.BillingPlanCommon != nil && resp.Properties.BillingPlanCommon.BillingPeriod != nil {
		r.billingPeriod = resp.Properties.BillingPlanCommon.BillingPeriod
	}

	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		r.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if r.projectID == "" && r.RespURI() != "" {
		if pid := parseURIIDs(r.RespURI())["projects"]; pid != "" {
			r.projectID = pid
		}
	}
}

func containerRegistryDeref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// containerRegistryIDsFromRef extracts (projectID, registryID) from a Ref.
// Uses URI segment fallback on "registries" — no typed ancestor interface needed
// since ContainerRegistry has no descendant resource types.
func containerRegistryIDsFromRef(ref Ref) (projectID, registryID string, err error) {
	rid, ok := extractID(ref, func(r Ref) (string, bool) {
		return "", false // no typed interface — URI-only path
	}, "registries")
	if !ok || rid == "" {
		return "", "", fmt.Errorf("cannot determine registry ID from Ref %q", ref.URI())
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
	return pid, rid, nil
}

// containerRegistriesLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type containerRegistriesLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.ContainerRegistryListResponse], error)
	Get(ctx context.Context, projectID, registryID string, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error)
	Create(ctx context.Context, projectID string, body types.ContainerRegistryRequest, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error)
	Update(ctx context.Context, projectID, registryID string, body types.ContainerRegistryRequest, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error)
	Delete(ctx context.Context, projectID, registryID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// containerRegistriesClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates ContainerRegistry ↔ types.ContainerRegistryRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type containerRegistriesClientAdapter struct {
	low  containerRegistriesLowLevelClient
	rest *restclient.Client
}

func newContainerRegistriesClientAdapter(rest *restclient.Client) *containerRegistriesClientAdapter {
	if rest == nil {
		return &containerRegistriesClientAdapter{}
	}
	return &containerRegistriesClientAdapter{low: container.NewContainerRegistryClientImpl(rest), rest: rest}
}

// Create posts a new ContainerRegistry to the API and hydrates the wrapper from the response.
func (a *containerRegistriesClientAdapter) Create(ctx context.Context, r *ContainerRegistry, opts ...CallOption) (*ContainerRegistry, error) {
	if err := r.Err(); err != nil {
		return r, err
	}
	if r.ProjectID() == "" {
		return r, fmt.Errorf("Create: ContainerRegistry has no parent project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, r.ProjectID(), r.toRequest(), rp)
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

// Update sends a PUT for the current wrapper state. Requires ID and parent.
func (a *containerRegistriesClientAdapter) Update(ctx context.Context, r *ContainerRegistry, opts ...CallOption) (*ContainerRegistry, error) {
	if err := r.Err(); err != nil {
		return r, err
	}
	if r.ContainerRegistryID() == "" {
		return r, fmt.Errorf("Update: ContainerRegistry has no ID")
	}
	if r.ProjectID() == "" {
		return r, fmt.Errorf("Update: ContainerRegistry has no parent project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, r.ProjectID(), r.ContainerRegistryID(), r.toRequest(), rp)
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

// Get fetches a ContainerRegistry by Ref and returns a freshly hydrated wrapper.
func (a *containerRegistriesClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*ContainerRegistry, error) {
	projectID, registryID, err := containerRegistryIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, registryID, rp)
	out := &ContainerRegistry{}
	out.projectID = projectID
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

// Delete removes the ContainerRegistry identified by Ref.
func (a *containerRegistriesClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, registryID, err := containerRegistryIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, registryID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of ContainerRegistry in the given parent scope.
func (a *containerRegistriesClientAdapter) List(ctx context.Context, parent Ref, opts ...CallOption) (*List[*ContainerRegistry], error) {
	projectID, err := projectIDFromRef(parent)
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
	var items []*ContainerRegistry
	if resp != nil && resp.Data != nil {
		items = make([]*ContainerRegistry, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			cr := &ContainerRegistry{}
			cr.projectID = projectID
			cr.fromResponse(&resp.Data.Values[i])
			cr.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, cr)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					cr.fromResponse(fresh.Raw())
				}
				return nil
			})
			if cr.projectID == "" {
				cr.projectID = projectID
			}
			items = append(items, cr)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*ContainerRegistry], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*ContainerRegistry], error) {
		fetch := listPageFetch[types.ContainerRegistryListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*ContainerRegistry
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*ContainerRegistry, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				cr := &ContainerRegistry{}
				cr.projectID = projectID
				cr.fromResponse(&pageResp.Data.Values[i])
				cr.setRefresh(func(ctx context.Context) error {
					fresh, err := a.Get(ctx, cr)
					if err != nil {
						return err
					}
					if fresh != nil && fresh.Raw() != nil {
						cr.fromResponse(fresh.Raw())
					}
					return nil
				})
				if cr.projectID == "" {
					cr.projectID = projectID
				}
				pageItems = append(pageItems, cr)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}
