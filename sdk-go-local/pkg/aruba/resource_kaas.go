package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/container"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// KaaS is the wrapper for an Aruba Cloud Kubernetes-as-a-Service cluster
// (a direct child of a Project). Construct with aruba.NewKaaS() and bind it
// via InProject(project), WithVPC(vpc), WithSubnet(subnet), etc.
//
// Family A: regional, Metadata/Properties envelope, location-aware.
// Supports full CRUD. Update emits KaaSUpdateRequest (narrower than KaaSRequest):
// only KubernetesVersion, NodePools, HA, Storage, and BillingPlanCommon are mutable.
//
// Schema asymmetry (request vs. response):
//   - "nodePools" (request) vs. "nodesPool" (response)
//   - "nodeCidr" (request) vs. "nodecidr" (response)
//   - "podCidr" (request) vs. "podcidr" (response)
//   - NodePoolPropertiesRequest.Instance string (request) vs. KaaSNodePoolInstanceResponse{ID,Name} (response)
//   - NodePoolPropertiesRequest.Zone string (request JSON "dataCenter") vs. KaaSNodePoolDataCenterResponse{Code,Name} (response)
//
// Path: /projects/{projectID}/providers/Aruba.Container/kaas[/{kaasID}]
type KaaS struct {
	errMixin
	metadataMixin
	regionalMixin
	projectScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	// Body-refs (single).
	vpcRef    *string
	subnetRef *string

	// Plain-string / scalar body fields.
	securityGroupName    *string
	nodeCIDRAddress      *string
	nodeCIDRName         *string
	podCIDR              *string
	kubernetesVersion    *KubernetesVersion
	ha                   *bool
	storageMaxCumulative *int32 // wire: storage.maxCumulativeVolumeSize
	billingPeriod        *BillingPeriod
	identityClientID     *string
	identityClientSecret *string

	// API server access profile — stored as scalars so callers need not import pkg/types.
	apiServerPrivateCluster     *bool
	apiServerAuthorizedIPRanges *[]string

	// Sub-builders.
	nodePools []*NodePool

	// Action executor — set by adapter; nil on locally-built wrappers.
	actions kaasActions

	response *types.KaaSResponse
}

// NewKaaS returns a fresh *KaaS ready for fluent setters and a Create call.
// Binds projectScopedMixin's error sink so IntoProject failures surface via Err().
func NewKaaS() *KaaS {
	k := &KaaS{}
	k.projectScopedMixin = bindProjectScoped(&k.errMixin)
	return k
}

// Setters — chainable, general → specific

// InProject binds this KaaS to its parent project. Required before Create.
func (k *KaaS) InProject(p Ref) *KaaS { k.intoProject(p); return k }

// Named sets the resource name. Required by the API.
func (k *KaaS) Named(n string) *KaaS { k.named(n); return k }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (k *KaaS) Tagged(ts ...string) *KaaS {
	for _, t := range ts {
		k.addTag(t)
	}
	return k
}

// Untagged removes each listed tag. No-op for tags not present.
func (k *KaaS) Untagged(ts ...string) *KaaS {
	for _, t := range ts {
		k.removeTag(t)
	}
	return k
}

// RetaggedAs replaces the entire tag set with the given values.
func (k *KaaS) RetaggedAs(ts ...string) *KaaS { k.replaceTags(ts...); return k }

// InRegion sets the region for this resource.
func (k *KaaS) InRegion(region Region) *KaaS { k.inRegion(region); return k }

// WithKubernetesVersion sets the Kubernetes version for the cluster.
func (k *KaaS) WithKubernetesVersion(v KubernetesVersion) *KaaS {
	k.kubernetesVersion = &v
	return k
}

// WithPodCIDR sets the pod CIDR block for the cluster network.
func (k *KaaS) WithPodCIDR(cidr string) *KaaS { k.podCIDR = &cidr; return k }

// HighlyAvailable enables high-availability mode for the control plane.
func (k *KaaS) HighlyAvailable() *KaaS { v := true; k.ha = &v; return k }

// BilledBy sets the billing cadence. Accepted periods are resource-specific; check the API reference.
func (k *KaaS) BilledBy(period BillingPeriod) *KaaS { k.billingPeriod = &period; return k }

// WithSecurityGroup attaches a SecurityGroup to the cluster. The KaaS API
// stores only the SG's name (not its URI), so the supplied Ref must be a
// *SecurityGroup whose Name() is non-empty. Bare URI refs are rejected
// because the name cannot be recovered from a URI alone.
//
// If you only have the SG name (e.g. from a CLI flag), use
// WithSecurityGroupName instead.
func (k *KaaS) WithSecurityGroup(sg Ref) *KaaS {
	if sg == nil {
		k.addErr(fmt.Errorf("WithSecurityGroup: nil Ref"))
		return k
	}
	typed, ok := sg.(*SecurityGroup)
	if !ok {
		k.addErr(fmt.Errorf("WithSecurityGroup: requires *SecurityGroup, got %T; for name-only callers use WithSecurityGroupName", sg))
		return k
	}
	name := typed.Name()
	if name == "" {
		k.addErr(fmt.Errorf("WithSecurityGroup: SecurityGroup has empty Name"))
		return k
	}
	k.securityGroupName = &name
	return k
}

// WithSecurityGroupName attaches a Security Group to the cluster by name. Use
// this when only the SG name is known (e.g. from a CLI flag) and a typed
// *SecurityGroup is not available. The wire API stores only the name.
func (k *KaaS) WithSecurityGroupName(name string) *KaaS {
	if name == "" {
		k.addErr(fmt.Errorf("WithSecurityGroupName: name cannot be empty"))
		return k
	}
	k.securityGroupName = &name
	return k
}

// WithNodeCIDR sets the node CIDR block (address and name).
// The wire type is NodeCIDRPropertiesRequest{Address, Name}.
func (k *KaaS) WithNodeCIDR(address, name string) *KaaS {
	k.nodeCIDRAddress = &address
	k.nodeCIDRName = &name
	return k
}

// WithMaxStorageQuotaGB sets the maximum cumulative volume size in GB.
func (k *KaaS) WithMaxStorageQuotaGB(gb int) *KaaS {
	v := int32(gb)
	k.storageMaxCumulative = &v
	return k
}

// WithIdentity sets the managed identity credentials.
func (k *KaaS) WithIdentity(clientID, clientSecret string) *KaaS {
	k.identityClientID = &clientID
	k.identityClientSecret = &clientSecret
	return k
}

// WithPrivateCluster enables private cluster mode for the API server, restricting
// access to the control plane to the cluster's internal network only.
func (k *KaaS) WithPrivateCluster() *KaaS {
	v := true
	k.apiServerPrivateCluster = &v
	return k
}

// EnablePrivateCluster enables private cluster mode for the API server.
// Preferred alias for WithPrivateCluster, matching the HighlyAvailable naming convention.
func (k *KaaS) EnablePrivateCluster() *KaaS { return k.WithPrivateCluster() }

// WithAuthorizedIPRanges restricts API server access to the given CIDR ranges.
// Pass zero arguments to clear any previously-set ranges.
func (k *KaaS) WithAuthorizedIPRanges(ranges ...string) *KaaS {
	if len(ranges) == 0 {
		k.apiServerAuthorizedIPRanges = nil
		return k
	}
	cp := make([]string, len(ranges))
	copy(cp, ranges)
	k.apiServerAuthorizedIPRanges = &cp
	return k
}

// WithAPIServerAccessProfile sets the API server access profile from a pre-built struct.
// Deprecated: use EnablePrivateCluster and WithAuthorizedIPRanges instead to avoid
// importing pkg/types. Passing nil clears any previously-set profile fields.
func (k *KaaS) WithAPIServerAccessProfile(p *types.KaaSAPIServerAccessProfilePropertiesRequest) *KaaS {
	if p == nil {
		k.apiServerPrivateCluster = nil
		k.apiServerAuthorizedIPRanges = nil
		return k
	}
	v := p.EnablePrivateCluster
	k.apiServerPrivateCluster = &v
	k.apiServerAuthorizedIPRanges = p.AuthorizedIPRanges
	return k
}

// WithVPC binds the KaaS cluster to the given VPC by URI. Required before Create.
func (k *KaaS) WithVPC(v Ref) *KaaS { return k.setSingleRef("WithVPC", v, &k.vpcRef) }

// WithSubnet binds the KaaS cluster to the given subnet by URI. Required before Create.
func (k *KaaS) WithSubnet(s Ref) *KaaS { return k.setSingleRef("WithSubnet", s, &k.subnetRef) }

func (k *KaaS) setSingleRef(label string, ref Ref, dst **string) *KaaS {
	uri := ref.URI()
	if uri == "" {
		k.addErr(fmt.Errorf("%s: empty URI", label))
		return k
	}
	*dst = &uri
	return k
}

// WithNodePools appends node pools to the cluster's pool list.
// Errors accumulated on each pool are drained into k at attachment time.
func (k *KaaS) WithNodePools(nps ...*NodePool) *KaaS {
	for _, np := range nps {
		if np == nil {
			continue
		}
		for _, e := range np.errs {
			k.addErr(e)
		}
		k.nodePools = append(k.nodePools, np)
	}
	return k
}

// WithoutNodePools removes all previously-added node pools from the cluster.
func (k *KaaS) WithoutNodePools() *KaaS {
	k.nodePools = nil
	return k
}

// ReplaceNodePools replaces the entire node pool list with the given pools.
// Equivalent to WithoutNodePools followed by WithNodePools for each entry.
func (k *KaaS) ReplaceNodePools(pools ...*NodePool) *KaaS {
	k.nodePools = nil
	for _, np := range pools {
		k.WithNodePools(np)
	}
	return k
}

// DownloadKubeconfig downloads the kubeconfig for this cluster and returns it
// as raw bytes (the YAML content). Requires the wrapper to have been obtained
// via a client call (Get/Create/Update/List); locally-built wrappers return a
// clear error.
func (k *KaaS) DownloadKubeconfig(ctx context.Context, opts ...CallOption) ([]byte, error) {
	if err := k.preActionCheck("DownloadKubeconfig"); err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := k.actions.downloadKubeconfig(ctx, k.ProjectID(), k.KaaSID(), rp)
	populateHTTPEnvelope(&k.httpEnvelopeMixin, resp)
	if err != nil {
		return nil, err
	}
	if resp != nil && !resp.IsSuccess() {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	if resp == nil || resp.Data == nil {
		return nil, nil
	}
	return []byte(resp.Data.Content), nil
}

func (k *KaaS) preActionCheck(label string) error {
	if k.actions == nil {
		return fmt.Errorf("%s: this *KaaS was not obtained via a client call (no action executor) — fetch via Get/Create/Update/List first", label)
	}
	if k.KaaSID() == "" {
		return fmt.Errorf("%s: missing KaaS ID", label)
	}
	if k.ProjectID() == "" {
		return fmt.Errorf("%s: missing project ID", label)
	}
	return nil
}

// Getters — general → specific

// URI satisfies Ref by returning the server-assigned canonical URI, or "" if Create hasn't run yet.
func (k *KaaS) URI() string { return k.RespURI() }

// KaaSID satisfies withKaaSID so child wrappers can extract this ID by typed assertion.
func (k *KaaS) KaaSID() string { return k.ID() }

// Raw shadows responseMetadataMixin.Raw() with the typed KaaS response.
func (k *KaaS) Raw() *types.KaaSResponse { return k.response }
func (k *KaaS) RawJSON() []byte          { return marshalRawJSON(k.response) }
func (k *KaaS) RawYAML() []byte          { return marshalRawYAML(k.response) }

// RawRequest returns what toRequest() would emit right now.
func (k *KaaS) RawRequest() types.KaaSRequest { return k.toRequest() }

// VPC returns the VPC URI for this cluster, or "" if unset.
func (k *KaaS) VPC() string {
	if k.response != nil && k.response.Properties.VPC.URI != nil {
		return *k.response.Properties.VPC.URI
	}
	return kaasDeref(k.vpcRef)
}

// Subnet returns the subnet URI for this cluster, or "" if unset.
func (k *KaaS) Subnet() string {
	if k.response != nil && k.response.Properties.Subnet.URI != nil {
		return *k.response.Properties.Subnet.URI
	}
	return kaasDeref(k.subnetRef)
}

// SecurityGroupName returns the associated security group name, or "" if unset.
func (k *KaaS) SecurityGroupName() string {
	if k.response != nil && k.response.Properties.SecurityGroup.Name != nil {
		return *k.response.Properties.SecurityGroup.Name
	}
	return kaasDeref(k.securityGroupName)
}

// KubernetesVersion returns the Kubernetes version configured for this cluster, or "" if unset.
func (k *KaaS) KubernetesVersion() KubernetesVersion {
	if k.response != nil && k.response.Properties.KubernetesVersion.Value != nil {
		return KubernetesVersion(*k.response.Properties.KubernetesVersion.Value)
	}
	if k.kubernetesVersion == nil {
		return ""
	}
	return *k.kubernetesVersion
}

// BillingPeriod returns the billing period for this cluster, or "" if unset.
func (k *KaaS) BillingPeriod() BillingPeriod {
	if k.response != nil && k.response.Properties.BillingPlanCommon != nil && k.response.Properties.BillingPlanCommon.BillingPeriod != nil {
		return *k.response.Properties.BillingPlanCommon.BillingPeriod
	}
	if k.billingPeriod == nil {
		return ""
	}
	return *k.billingPeriod
}

// MaxStorageQuotaGB returns the maximum cumulative volume size in GB.
// Returns 0 if not set.
func (k *KaaS) MaxStorageQuotaGB() int {
	if k.response != nil && k.response.Properties.Storage != nil && k.response.Properties.Storage.MaxCumulativeVolumeSize != nil {
		return int(*k.response.Properties.Storage.MaxCumulativeVolumeSize)
	}
	if k.storageMaxCumulative == nil {
		return 0
	}
	return int(*k.storageMaxCumulative)
}

// PodCIDR returns the pod CIDR block for the cluster, or "" if unset.
// On a hydrated response the value comes from the response; otherwise returns what was passed to WithPodCIDR.
func (k *KaaS) PodCIDR() string {
	return kaasDeref(k.podCIDR)
}

// NodeCIDR returns the node CIDR block address for the cluster, or "" if unset.
// On a hydrated response the value comes from the response; otherwise returns what was passed to WithNodeCIDR.
func (k *KaaS) NodeCIDR() string {
	return kaasDeref(k.nodeCIDRAddress)
}

// IdentityClientID returns the managed identity client ID, or "" if unset.
// On a hydrated response the value comes from the response; otherwise returns what was passed to WithIdentity.
func (k *KaaS) IdentityClientID() string {
	return kaasDeref(k.identityClientID)
}

// APIServerPrivateCluster reports whether private cluster mode is enabled for the API server.
func (k *KaaS) APIServerPrivateCluster() bool {
	if k.response != nil && k.response.Properties.APIServerAccessProfile != nil {
		return k.response.Properties.APIServerAccessProfile.EnablePrivateCluster
	}
	if k.apiServerPrivateCluster != nil {
		return *k.apiServerPrivateCluster
	}
	return false
}

// APIServerAuthorizedIPRanges returns a copy of the authorized CIDR ranges for the API server,
// or nil if unset. Returns a copy to prevent callers from mutating internal state.
func (k *KaaS) APIServerAuthorizedIPRanges() []string {
	var src *[]string
	if k.response != nil && k.response.Properties.APIServerAccessProfile != nil && k.response.Properties.APIServerAccessProfile.AuthorizedIPRanges != nil {
		src = k.response.Properties.APIServerAccessProfile.AuthorizedIPRanges
	} else if k.apiServerAuthorizedIPRanges != nil {
		src = k.apiServerAuthorizedIPRanges
	}
	if src == nil {
		return nil
	}
	cp := make([]string, len(*src))
	copy(cp, *src)
	return cp
}

// NodePools returns the node pools attached to this cluster.
// Returns nil if no node pools have been configured.
func (k *KaaS) NodePools() []*NodePool {
	return k.nodePools
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (k *KaaS) toRequest() types.KaaSRequest {
	props := types.KaaSPropertiesRequest{
		VPC:    types.ReferenceResourceCommon{URI: kaasDeref(k.vpcRef)},
		Subnet: types.ReferenceResourceCommon{URI: kaasDeref(k.subnetRef)},
		SecurityGroup: types.KaaSSecurityGroupPropertiesRequest{
			Name: kaasDeref(k.securityGroupName),
		},
		NodeCIDR: types.NodeCIDRPropertiesRequest{
			Address: kaasDeref(k.nodeCIDRAddress),
			Name:    kaasDeref(k.nodeCIDRName),
		},
		PodCIDR: k.podCIDR,
		KubernetesVersion: types.KubernetesVersionInfoRequest{Value: func() KubernetesVersion {
			if k.kubernetesVersion != nil {
				return *k.kubernetesVersion
			}
			return ""
		}()},
		HA:                k.ha,
		BillingPlanCommon: &types.BillingPlanCommon{BillingPeriod: defaultBillingPeriod(k.billingPeriod)},
	}
	if k.storageMaxCumulative != nil {
		props.Storage = types.StorageKubernetesCommon{MaxCumulativeVolumeSize: k.storageMaxCumulative}
	}
	if k.identityClientID != nil || k.identityClientSecret != nil {
		props.Identity = &types.KaaSIdentityPropertiesRequest{
			ClientID:     k.identityClientID,
			ClientSecret: k.identityClientSecret,
		}
	}
	if k.apiServerPrivateCluster != nil || k.apiServerAuthorizedIPRanges != nil {
		privateCluster := false
		if k.apiServerPrivateCluster != nil {
			privateCluster = *k.apiServerPrivateCluster
		}
		props.APIServerAccessProfile = &types.KaaSAPIServerAccessProfilePropertiesRequest{
			EnablePrivateCluster: privateCluster,
			AuthorizedIPRanges:   k.apiServerAuthorizedIPRanges,
		}
	}
	if len(k.nodePools) > 0 {
		props.NodePools = make([]types.NodePoolPropertiesRequest, 0, len(k.nodePools))
		for _, np := range k.nodePools {
			props.NodePools = append(props.NodePools, np.build())
		}
	}
	return types.KaaSRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: k.toMetadata(),
			Location:                k.toLocation(),
		},
		Properties: props,
	}
}

// toUpdateRequest emits KaaSUpdateRequest, which exposes only the mutable
// fields (KubernetesVersion, NodePools, HA, Storage, BillingPlanCommon).
// VPC, Subnet, SecurityGroup, and CIDRs are immutable after creation.
func (k *KaaS) toUpdateRequest() types.KaaSUpdateRequest {
	props := types.KaaSPropertiesUpdateRequest{
		KubernetesVersion: types.KubernetesVersionInfoUpdateRequest{
			Value: func() KubernetesVersion {
				if k.kubernetesVersion != nil {
					return *k.kubernetesVersion
				}
				return ""
			}(),
		},
		HA: k.ha,
	}
	if len(k.nodePools) > 0 {
		props.NodePools = make([]types.NodePoolPropertiesRequest, 0, len(k.nodePools))
		for _, np := range k.nodePools {
			props.NodePools = append(props.NodePools, np.build())
		}
	}
	if k.storageMaxCumulative != nil {
		props.Storage = &types.StorageKubernetesCommon{MaxCumulativeVolumeSize: k.storageMaxCumulative}
	}
	if k.billingPeriod != nil {
		props.BillingPlanCommon = &types.BillingPlanCommon{BillingPeriod: k.billingPeriod}
	}
	return types.KaaSUpdateRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: k.toMetadata(),
			Location:                k.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (k *KaaS) fromResponse(resp *types.KaaSResponse) {
	if resp == nil {
		return
	}
	k.response = resp
	k.setMeta(&resp.Metadata)
	k.named(kaasDeref(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		k.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		k.inRegion(resp.Metadata.LocationResponse.Value)
	}
	k.setStatus(&resp.Status)

	k.setLinked(resp.Properties.LinkedResources)
	k.kaasHydrateCacheFromProps(resp.Properties)
	k.nodePools = kaasRebuildNodePools(resp.Properties.NodePools)
	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		k.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if k.projectID == "" && k.RespURI() != "" {
		if pid := parseURIIDs(k.RespURI())["projects"]; pid != "" {
			k.projectID = pid
		}
	}
}

func (k *KaaS) kaasHydrateCacheFromProps(props types.KaaSPropertiesResponse) {
	if props.VPC.URI != nil && *props.VPC.URI != "" {
		v := *props.VPC.URI
		k.vpcRef = &v
	}
	if props.Subnet.URI != nil && *props.Subnet.URI != "" {
		v := *props.Subnet.URI
		k.subnetRef = &v
	}
	if props.SecurityGroup.Name != nil && *props.SecurityGroup.Name != "" {
		v := *props.SecurityGroup.Name
		k.securityGroupName = &v
	}
	if props.NodeCIDR.Address != nil && *props.NodeCIDR.Address != "" {
		v := *props.NodeCIDR.Address
		k.nodeCIDRAddress = &v
	}
	if props.NodeCIDR.Name != nil && *props.NodeCIDR.Name != "" {
		v := *props.NodeCIDR.Name
		k.nodeCIDRName = &v
	}
	if props.PodCIDR != nil && props.PodCIDR.Address != nil {
		v := *props.PodCIDR.Address
		k.podCIDR = &v
	}
	if props.KubernetesVersion.Value != nil && *props.KubernetesVersion.Value != "" {
		v := KubernetesVersion(*props.KubernetesVersion.Value)
		k.kubernetesVersion = &v
	}
	k.ha = props.HA
	if props.Storage != nil && props.Storage.MaxCumulativeVolumeSize != nil {
		v := *props.Storage.MaxCumulativeVolumeSize
		k.storageMaxCumulative = &v
	}
	if props.BillingPlanCommon != nil && props.BillingPlanCommon.BillingPeriod != nil {
		k.billingPeriod = props.BillingPlanCommon.BillingPeriod
	}
	if props.Identity != nil && props.Identity.ClientID != nil {
		v := *props.Identity.ClientID
		k.identityClientID = &v
		// ClientSecret is not returned in the response — caller must re-set on Update.
	}
	if props.APIServerAccessProfile != nil {
		v := props.APIServerAccessProfile.EnablePrivateCluster
		k.apiServerPrivateCluster = &v
		k.apiServerAuthorizedIPRanges = props.APIServerAccessProfile.AuthorizedIPRanges
	}
}

// kaasRebuildNodePools flattens response-side object types (KaaSNodePoolInstanceResponse,
// KaaSNodePoolDataCenterResponse) back to plain strings so toUpdateRequest() round-trips correctly.
func kaasRebuildNodePools(pools *[]types.NodePoolPropertiesResponse) []*NodePool {
	if pools == nil {
		return nil
	}
	result := make([]*NodePool, 0, len(*pools))
	for _, rp := range *pools {
		np := &NodePool{}
		if rp.Name != nil {
			v := *rp.Name
			np.name = &v
		}
		if rp.Nodes != nil {
			v := *rp.Nodes
			np.nodes = &v
		}
		if rp.Instance != nil && rp.Instance.Name != nil {
			v := NodePoolInstance(*rp.Instance.Name)
			np.instance = &v
		}
		if rp.DataCenter != nil && rp.DataCenter.Code != nil {
			v := Zone(*rp.DataCenter.Code)
			np.zone = &v
		}
		if rp.MinCount != nil {
			v := *rp.MinCount
			np.minCount = &v
		}
		if rp.MaxCount != nil {
			v := *rp.MaxCount
			np.maxCount = &v
		}
		b := rp.Autoscaling
		np.autoscaling = &b
		result = append(result, np)
	}
	return result
}

func kaasDeref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func kaasIDsFromRef(ref Ref) (projectID, kaasID string, err error) {
	kid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withKaaSID); ok {
			return w.KaaSID(), true
		}
		return "", false
	}, "kaas")
	if !ok || kid == "" {
		return "", "", fmt.Errorf("cannot determine KaaS ID from Ref %q", ref.URI())
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
	return pid, kid, nil
}

// kaasActions is the contract for KaaS lifecycle action operations (Start, Stop, etc.).
type kaasActions interface {
	downloadKubeconfig(ctx context.Context, projectID, kaasID string, rp *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error)
}

// ---- Low-level client interface ----

// kaasLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type kaasLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.KaaSListResponse], error)
	Get(ctx context.Context, projectID, kaasID string, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error)
	Create(ctx context.Context, projectID string, body types.KaaSRequest, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error)
	Update(ctx context.Context, projectID, kaasID string, body types.KaaSUpdateRequest, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error)
	Delete(ctx context.Context, projectID, kaasID string, params *types.RequestParameters) (*types.Response[any], error)
	DownloadKubeconfig(ctx context.Context, projectID, kaasID string, params *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error)
}

// ---- Adapter ----

// kaasClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates KaaS ↔ types.KaaSRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type kaasClientAdapter struct {
	low  kaasLowLevelClient
	rest *restclient.Client
}

var _ kaasActions = (*kaasClientAdapter)(nil)

func newKaaSClientAdapter(rest *restclient.Client) *kaasClientAdapter {
	if rest == nil {
		return &kaasClientAdapter{}
	}
	return &kaasClientAdapter{low: container.NewKaaSClientImpl(rest), rest: rest}
}

// Create posts a new KaaS to the API and hydrates the wrapper from the response.
func (a *kaasClientAdapter) Create(ctx context.Context, k *KaaS, opts ...CallOption) (*KaaS, error) {
	if err := k.Err(); err != nil {
		return k, err
	}
	if k.ProjectID() == "" {
		return k, fmt.Errorf("Create: KaaS has no parent project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, k.ProjectID(), k.toRequest(), rp)
	populateHTTPEnvelope(&k.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		k.fromResponse(resp.Data)
		k.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, k)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				k.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	k.actions = a
	if err != nil {
		return k, err
	}
	if resp != nil && !resp.IsSuccess() {
		return k, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return k, nil
}

// Update sends a PUT for the current wrapper state. Requires ID and parent.
func (a *kaasClientAdapter) Update(ctx context.Context, k *KaaS, opts ...CallOption) (*KaaS, error) {
	if err := k.Err(); err != nil {
		return k, err
	}
	if k.KaaSID() == "" {
		return k, fmt.Errorf("Update: KaaS has no ID")
	}
	if k.ProjectID() == "" {
		return k, fmt.Errorf("Update: KaaS has no parent project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, k.ProjectID(), k.KaaSID(), k.toUpdateRequest(), rp)
	populateHTTPEnvelope(&k.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		k.fromResponse(resp.Data)
		k.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, k)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				k.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	k.actions = a
	if err != nil {
		return k, err
	}
	if resp != nil && !resp.IsSuccess() {
		return k, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return k, nil
}

// Get fetches a KaaS by Ref and returns a freshly hydrated wrapper.
func (a *kaasClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*KaaS, error) {
	projectID, kaasID, err := kaasIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, kaasID, rp)
	out := &KaaS{}
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
	out.actions = a
	if err != nil {
		return out, err
	}
	if resp != nil && !resp.IsSuccess() {
		return out, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return out, nil
}

// Delete removes the KaaS identified by Ref.
func (a *kaasClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, kaasID, err := kaasIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, kaasID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of KaaS in the given parent scope.
func (a *kaasClientAdapter) List(ctx context.Context, parent Ref, opts ...CallOption) (*List[*KaaS], error) {
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
	var items []*KaaS
	if resp != nil && resp.Data != nil {
		items = make([]*KaaS, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			k := &KaaS{}
			k.projectID = projectID
			k.fromResponse(&resp.Data.Values[i])
			k.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, k)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					k.fromResponse(fresh.Raw())
				}
				return nil
			})
			if k.projectID == "" {
				k.projectID = projectID
			}
			k.actions = a
			items = append(items, k)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*KaaS], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*KaaS], error) {
		fetch := listPageFetch[types.KaaSListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*KaaS
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*KaaS, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				k := &KaaS{}
				k.projectID = projectID
				k.fromResponse(&pageResp.Data.Values[i])
				k.setRefresh(func(ctx context.Context) error {
					fresh, err := a.Get(ctx, k)
					if err != nil {
						return err
					}
					if fresh != nil && fresh.Raw() != nil {
						k.fromResponse(fresh.Raw())
					}
					return nil
				})
				if k.projectID == "" {
					k.projectID = projectID
				}
				k.actions = a
				pageItems = append(pageItems, k)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}

// downloadKubeconfig satisfies kaasActions (lowercase, internal interface).
func (a *kaasClientAdapter) downloadKubeconfig(ctx context.Context, projectID, kaasID string, rp *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error) {
	return a.low.DownloadKubeconfig(ctx, projectID, kaasID, rp)
}
