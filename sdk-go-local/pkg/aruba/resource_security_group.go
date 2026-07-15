package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/network"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"k8s.io/utils/ptr"
)

// SecurityGroupRef returns a Ref that points to the SecurityGroup nested under a VPC.
func SecurityGroupRef(projectID, vpcID, sgID string) Ref {
	return URI(fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpcs/%s/securityGroups/%s", projectID, vpcID, sgID))
}

// ---- Wrapper ----

// SecurityGroup is the wrapper for an Aruba Cloud Security Group (a child of a VPC).
// Construct with aruba.NewSecurityGroup() and bind it via IntoVPC(vpc).
//
// Wraps types.SecurityGroupResponse / types.SecurityGroupRequest. The wrapper carries
// pointer-typed private fields so unset values round-trip through
// the JSON layer correctly.
type SecurityGroup struct {
	errMixin
	metadataMixin
	vpcScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	defaultSG *bool                        // Properties.Default (request: *bool for omitempty; response: plain bool)
	response  *types.SecurityGroupResponse // backs Raw()
}

// NewSecurityGroup returns a fresh *SecurityGroup ready for fluent setters and a Create call.
// Binds vpcScopedMixin's error sink so IntoVPC failures surface via Err().
func NewSecurityGroup() *SecurityGroup {
	sg := &SecurityGroup{defaultSG: ptr.To(false)}
	sg.vpcScopedMixin = bindVPCScoped(&sg.errMixin)
	return sg
}

// Setters — chainable, general → specific

// InVPC binds this SecurityGroup to its parent VPC. Required before Create.
func (sg *SecurityGroup) InVPC(v Ref) *SecurityGroup { sg.intoVPC(v); return sg }

// Named sets the resource name. Required by the API.
func (sg *SecurityGroup) Named(n string) *SecurityGroup { sg.named(n); return sg }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (sg *SecurityGroup) Tagged(ts ...string) *SecurityGroup {
	for _, t := range ts {
		sg.addTag(t)
	}
	return sg
}

// Untagged removes each listed tag. No-op for tags not present.
func (sg *SecurityGroup) Untagged(ts ...string) *SecurityGroup {
	for _, t := range ts {
		sg.removeTag(t)
	}
	return sg
}

// RetaggedAs replaces the entire tag set with the given values.
func (sg *SecurityGroup) RetaggedAs(ts ...string) *SecurityGroup { sg.replaceTags(ts...); return sg }

// AsDefault marks this security group as the VPC default.
func (sg *SecurityGroup) AsDefault() *SecurityGroup { t := true; sg.defaultSG = &t; return sg }

// NotDefault explicitly unsets the default flag.
func (sg *SecurityGroup) NotDefault() *SecurityGroup { f := false; sg.defaultSG = &f; return sg }

// Getters — general → specific

// URI satisfies Ref.
func (sg *SecurityGroup) URI() string { return sg.RespURI() }

// SecurityGroupID satisfies withSecurityGroupID so child wrappers (SecurityGroupRule)
// can extract this ID via typed assertion.
func (sg *SecurityGroup) SecurityGroupID() string { return sg.ID() }

// Raw shadows responseMetadataMixin.Raw() with the full SecurityGroup response.
func (sg *SecurityGroup) Raw() *types.SecurityGroupResponse { return sg.response }
func (sg *SecurityGroup) RawJSON() []byte                   { return marshalRawJSON(sg.response) }
func (sg *SecurityGroup) RawYAML() []byte                   { return marshalRawYAML(sg.response) }

// RawRequest returns what toRequest() would emit right now.
func (sg *SecurityGroup) RawRequest() types.SecurityGroupRequest { return sg.toRequest() }

// IsDefault returns the security group's default flag, or false if unset.
func (sg *SecurityGroup) IsDefault() bool {
	if sg.defaultSG == nil {
		return false
	}
	return *sg.defaultSG
}

// Rules returns the linked security group rules as a slice of LinkedResourceCommon.
// This is an alias for LinkedResources(), exposing the rules under a domain-specific name.
// Returns nil when no rules are linked.
func (sg *SecurityGroup) Rules() []types.LinkedResourceCommon {
	return sg.LinkedResources()
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (sg *SecurityGroup) toRequest() types.SecurityGroupRequest {
	return types.SecurityGroupRequest{
		Metadata: sg.toMetadata(),
		Properties: types.SecurityGroupPropertiesRequest{
			Default: sg.defaultSG,
		},
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (sg *SecurityGroup) fromResponse(resp *types.SecurityGroupResponse) {
	if resp == nil {
		return
	}
	sg.response = resp
	sg.setMeta(&resp.Metadata)
	sg.named(securityGroupDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		sg.replaceTags(resp.Metadata.Tags...)
	}
	sg.setStatus(&resp.Status)

	sg.setLinked(resp.Properties.LinkedResources)

	// Properties.Default is plain bool on the response — backfill into our *bool field.
	d := resp.Properties.Default
	sg.defaultSG = &d

	// Backfill ancestor IDs: prefer ProjectMetadataResponse, then URI parse.
	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		sg.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if (sg.vpcID == "" || sg.projectID == "") && sg.RespURI() != "" {
		ids := parseURIIDs(sg.RespURI())
		if sg.vpcID == "" {
			sg.vpcID = ids["vpcs"]
		}
		if sg.projectID == "" {
			sg.projectID = ids["projects"]
		}
	}
}

func securityGroupDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// securityGroupLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type securityGroupLowLevelClient interface {
	List(ctx context.Context, projectID, vpcID string, params *types.RequestParameters) (*types.Response[types.SecurityGroupListResponse], error)
	Get(ctx context.Context, projectID, vpcID, securityGroupID string, params *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error)
	Create(ctx context.Context, projectID, vpcID string, body types.SecurityGroupRequest, params *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error)
	Update(ctx context.Context, projectID, vpcID, securityGroupID string, body types.SecurityGroupRequest, params *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error)
	Delete(ctx context.Context, projectID, vpcID, securityGroupID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// securityGroupsClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates SecurityGroup ↔ types.SecurityGroupRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type securityGroupsClientAdapter struct {
	low  securityGroupLowLevelClient
	rest *restclient.Client
}

var _ SecurityGroupsClient = (*securityGroupsClientAdapter)(nil)

func newSecurityGroupsClientAdapter(rest *restclient.Client) *securityGroupsClientAdapter {
	if rest == nil {
		return &securityGroupsClientAdapter{}
	}
	return &securityGroupsClientAdapter{
		low:  network.NewSecurityGroupsClientImpl(rest, network.NewVPCsClientImpl(rest)),
		rest: rest,
	}
}

// Create posts a new SecurityGroup to the API and hydrates the wrapper from the response.
func (a *securityGroupsClientAdapter) Create(ctx context.Context, sg *SecurityGroup, opts ...CallOption) (*SecurityGroup, error) {
	if err := sg.Err(); err != nil {
		return sg, err
	}
	if sg.VPCID() == "" || sg.ProjectID() == "" {
		return sg, fmt.Errorf("Create: security group has no VPC — call InVPC first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, sg.ProjectID(), sg.VPCID(), sg.toRequest(), rp)
	populateHTTPEnvelope(&sg.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		sg.fromResponse(resp.Data)
		sg.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, sg)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				sg.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return sg, err
	}
	if resp != nil && !resp.IsSuccess() {
		return sg, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return sg, nil
}

// Get fetches a SecurityGroup by Ref and returns a freshly hydrated wrapper.
func (a *securityGroupsClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*SecurityGroup, error) {
	projectID, vpcID, securityGroupID, err := securityGroupIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, vpcID, securityGroupID, rp)
	out := &SecurityGroup{}
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
func (a *securityGroupsClientAdapter) Update(ctx context.Context, sg *SecurityGroup, opts ...CallOption) (*SecurityGroup, error) {
	if err := sg.Err(); err != nil {
		return sg, err
	}
	if sg.ID() == "" {
		return sg, fmt.Errorf("Update: security group has no ID — call Get first or seed from response metadata")
	}
	if sg.VPCID() == "" || sg.ProjectID() == "" {
		return sg, fmt.Errorf("Update: security group has no VPC — call InVPC first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, sg.ProjectID(), sg.VPCID(), sg.ID(), sg.toRequest(), rp)
	populateHTTPEnvelope(&sg.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		sg.fromResponse(resp.Data)
		sg.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, sg)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				sg.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return sg, err
	}
	if resp != nil && !resp.IsSuccess() {
		return sg, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return sg, nil
}

// Delete removes the SecurityGroup identified by Ref.
func (a *securityGroupsClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, vpcID, securityGroupID, err := securityGroupIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, vpcID, securityGroupID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of SecurityGroup in the given parent scope.
func (a *securityGroupsClientAdapter) List(ctx context.Context, vpc Ref, opts ...CallOption) (*List[*SecurityGroup], error) {
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
	var items []*SecurityGroup
	if resp != nil && resp.Data != nil {
		items = make([]*SecurityGroup, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			sg := &SecurityGroup{}
			sg.fromResponse(&resp.Data.Values[i])
			sg.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, sg)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					sg.fromResponse(fresh.Raw())
				}
				return nil
			})
			if sg.vpcID == "" {
				sg.vpcID = vpcID
			}
			if sg.projectID == "" {
				sg.projectID = projectID
			}
			items = append(items, sg)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*SecurityGroup], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*SecurityGroup], error) {
		fetch := listPageFetch[types.SecurityGroupListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*SecurityGroup
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*SecurityGroup, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				item := &SecurityGroup{}
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

// securityGroupIDsFromRef extracts (projectID, vpcID, securityGroupID) from a Ref.
// Tries typed assertions first, then falls back to URI path parsing.
func securityGroupIDsFromRef(ref Ref) (projectID, vpcID, securityGroupID string, err error) {
	sgid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withSecurityGroupID); ok {
			return w.SecurityGroupID(), true
		}
		return "", false
	}, "securityGroups")
	if !ok || sgid == "" {
		return "", "", "", fmt.Errorf("cannot determine security group ID from Ref %q", ref.URI())
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
	return pid, vid, sgid, nil
}
