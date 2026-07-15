package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/network"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// SecurityRuleRef returns a Ref that points to the SecurityRule nested under a SecurityGroup.
func SecurityRuleRef(projectID, vpcID, sgID, ruleID string) Ref {
	return URI(fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpcs/%s/securityGroups/%s/securityRules/%s", projectID, vpcID, sgID, ruleID))
}

// ---- Wrapper ----

// SecurityRule is the wrapper for an Aruba Cloud Security Rule (a child of a SecurityGroup).
// Construct with aruba.NewSecurityRule() and bind it via IntoSecurityGroup(sg).
//
// Wraps types.SecurityRuleResponse / types.SecurityRuleRequest. The wrapper carries
// pointer-typed private fields so unset values round-trip through
// the JSON layer correctly.
type SecurityRule struct {
	errMixin
	metadataMixin
	regionalMixin
	securityGroupScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	direction *types.RuleDirection
	protocol  *RuleProtocol
	port      *string
	target    *types.RuleTargetCommon
	response  *types.SecurityRuleResponse
}

// NewSecurityRule returns a fresh *SecurityRule ready for fluent setters and a Create call.
// Binds securityGroupScopedMixin's error sink so IntoSecurityGroup failures surface via Err().
func NewSecurityRule() *SecurityRule {
	r := &SecurityRule{}
	r.securityGroupScopedMixin = bindSecurityGroupScoped(&r.errMixin)
	return r
}

// Setters — chainable, general → specific

// InSecurityGroup binds this SecurityRule to its parent SecurityGroup. Required before Create.
func (r *SecurityRule) InSecurityGroup(sg Ref) *SecurityRule { r.intoSecurityGroup(sg); return r }

// Named sets the resource name. Required by the API.
func (r *SecurityRule) Named(n string) *SecurityRule { r.named(n); return r }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (r *SecurityRule) Tagged(ts ...string) *SecurityRule {
	for _, t := range ts {
		r.addTag(t)
	}
	return r
}

// Untagged removes each listed tag. No-op for tags not present.
func (r *SecurityRule) Untagged(ts ...string) *SecurityRule {
	for _, t := range ts {
		r.removeTag(t)
	}
	return r
}

// RetaggedAs replaces the entire tag set with the given values.
func (r *SecurityRule) RetaggedAs(ts ...string) *SecurityRule { r.replaceTags(ts...); return r }

// InRegion sets the region for this resource.
func (r *SecurityRule) InRegion(region Region) *SecurityRule { r.inRegion(region); return r }

// WithDirection sets the rule direction.
func (r *SecurityRule) WithDirection(dir types.RuleDirection) *SecurityRule {
	r.direction = &dir
	return r
}

// WithProtocol sets the L4 protocol.
func (r *SecurityRule) WithProtocol(proto RuleProtocol) *SecurityRule {
	r.protocol = &proto
	return r
}

// WithPort sets the port specifier — single (e.g., "22"), range ("80-100"), or wildcard ("*").
func (r *SecurityRule) WithPort(port string) *SecurityRule {
	r.port = &port
	return r
}

// TargetingCIDR sets the target as an IP/CIDR endpoint.
// Mutually exclusive with TargetingSecurityGroup — setting both records a setter-time error.
func (r *SecurityRule) TargetingCIDR(cidr string) *SecurityRule {
	if r.target != nil && r.target.Kind == types.EndpointTypeSecurityGroup {
		r.addErr(fmt.Errorf("TargetingCIDR: target already set to SecurityGroup; pick one"))
		return r
	}
	r.target = &types.RuleTargetCommon{Kind: types.EndpointTypeIP, Value: cidr}
	return r
}

// TargetingSecurityGroup sets the target as another SecurityGroup endpoint.
// Mutually exclusive with TargetingCIDR — setting both records a setter-time error.
func (r *SecurityRule) TargetingSecurityGroup(sg Ref) *SecurityRule {
	if r.target != nil && r.target.Kind == types.EndpointTypeIP {
		r.addErr(fmt.Errorf("TargetingSecurityGroup: target already set to CIDR; pick one"))
		return r
	}
	uri := sg.URI()
	if uri == "" {
		r.addErr(fmt.Errorf("TargetingSecurityGroup: target SecurityGroup Ref has empty URI"))
		return r
	}
	r.target = &types.RuleTargetCommon{Kind: types.EndpointTypeSecurityGroup, Value: uri}
	return r
}

// Getters — general → specific

// URI satisfies Ref.
func (r *SecurityRule) URI() string { return r.RespURI() }

// SecurityRuleID satisfies withSecurityRuleID.
func (r *SecurityRule) SecurityRuleID() string { return r.ID() }

// Raw shadows responseMetadataMixin.Raw() with the full SecurityRule response.
func (r *SecurityRule) Raw() *types.SecurityRuleResponse { return r.response }
func (r *SecurityRule) RawJSON() []byte                  { return marshalRawJSON(r.response) }
func (r *SecurityRule) RawYAML() []byte                  { return marshalRawYAML(r.response) }

// RawRequest returns what toRequest() would emit right now.
func (r *SecurityRule) RawRequest() types.SecurityRuleRequest { return r.toRequest() }

// Direction returns the configured rule direction (zero value if unset).
func (r *SecurityRule) Direction() types.RuleDirection {
	if r.direction == nil {
		return ""
	}
	return *r.direction
}

// Protocol returns the configured protocol ("" if unset).
func (r *SecurityRule) Protocol() RuleProtocol {
	if r.protocol == nil {
		return ""
	}
	return *r.protocol
}

// Port returns the configured port specifier ("" if unset).
func (r *SecurityRule) Port() string {
	if r.port == nil {
		return ""
	}
	return *r.port
}

// TargetKind returns the configured target endpoint kind ("" if unset).
func (r *SecurityRule) TargetKind() types.EndpointTypeDto {
	if r.target == nil {
		return ""
	}
	return r.target.Kind
}

// TargetValue returns the configured target endpoint value ("" if unset).
func (r *SecurityRule) TargetValue() string {
	if r.target == nil {
		return ""
	}
	return r.target.Value
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (r *SecurityRule) toRequest() types.SecurityRuleRequest {
	props := types.SecurityRulePropertiesRequest{}
	if r.direction != nil {
		props.Direction = *r.direction
	}
	if r.protocol != nil {
		props.Protocol = *r.protocol
	}
	if r.port != nil {
		props.Port = *r.port
	}
	if r.target != nil {
		props.Target = r.target
	}
	return types.SecurityRuleRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: r.toMetadata(),
			Location:                r.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (r *SecurityRule) fromResponse(resp *types.SecurityRuleResponse) {
	if resp == nil {
		return
	}
	r.response = resp
	r.setMeta(&resp.Metadata)
	r.named(securityRuleDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		r.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		r.inRegion(resp.Metadata.LocationResponse.Value)
	}
	r.setStatus(&resp.Status)

	r.setLinked(resp.Properties.LinkedResources)

	if resp.Properties.Direction != "" {
		d := resp.Properties.Direction
		r.direction = &d
	}
	if resp.Properties.Protocol != "" {
		p := resp.Properties.Protocol
		r.protocol = &p
	}
	if p := resp.Properties.Port.String(); p != "" {
		r.port = &p
	}
	if resp.Properties.Target != nil {
		t := *resp.Properties.Target
		r.target = &t
	}

	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		r.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if (r.vpcID == "" || r.projectID == "" || r.securityGroupID == "") && r.RespURI() != "" {
		ids := parseURIIDs(r.RespURI())
		if r.vpcID == "" {
			r.vpcID = ids["vpcs"]
		}
		if r.projectID == "" {
			r.projectID = ids["projects"]
		}
		if r.securityGroupID == "" {
			r.securityGroupID = ids["securityGroups"]
		}
	}
}

func securityRuleDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// securityRuleLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type securityRuleLowLevelClient interface {
	List(ctx context.Context, projectID, vpcID, securityGroupID string, params *types.RequestParameters) (*types.Response[types.SecurityRuleListResponse], error)
	Get(ctx context.Context, projectID, vpcID, securityGroupID, securityRuleID string, params *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error)
	Create(ctx context.Context, projectID, vpcID, securityGroupID string, body types.SecurityRuleRequest, params *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error)
	Update(ctx context.Context, projectID, vpcID, securityGroupID, securityRuleID string, body types.SecurityRuleRequest, params *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error)
	Delete(ctx context.Context, projectID, vpcID, securityGroupID, securityRuleID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// securityRulesClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates SecurityRule ↔ types.SecurityRuleRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type securityRulesClientAdapter struct {
	low  securityRuleLowLevelClient
	rest *restclient.Client
}

var _ SecurityGroupRulesClient = (*securityRulesClientAdapter)(nil)

func newSecurityRulesClientAdapter(rest *restclient.Client) *securityRulesClientAdapter {
	if rest == nil {
		return &securityRulesClientAdapter{}
	}
	return &securityRulesClientAdapter{
		low: network.NewSecurityGroupRulesClientImpl(
			rest,
			network.NewSecurityGroupsClientImpl(rest, network.NewVPCsClientImpl(rest)),
		),
		rest: rest,
	}
}

// Create posts a new SecurityRule to the API and hydrates the wrapper from the response.
func (a *securityRulesClientAdapter) Create(ctx context.Context, rule *SecurityRule, opts ...CallOption) (*SecurityRule, error) {
	if err := rule.Err(); err != nil {
		return rule, err
	}
	if rule.SecurityGroupID() == "" || rule.VPCID() == "" || rule.ProjectID() == "" {
		return rule, fmt.Errorf("Create: security rule has no SecurityGroup — call InSecurityGroup first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, rule.ProjectID(), rule.VPCID(), rule.SecurityGroupID(), rule.toRequest(), rp)
	populateHTTPEnvelope(&rule.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		rule.fromResponse(resp.Data)
		rule.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, rule)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				rule.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return rule, err
	}
	if resp != nil && !resp.IsSuccess() {
		return rule, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return rule, nil
}

// Get fetches a SecurityRule by Ref and returns a freshly hydrated wrapper.
func (a *securityRulesClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*SecurityRule, error) {
	projectID, vpcID, securityGroupID, securityRuleID, err := securityRuleIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, vpcID, securityGroupID, securityRuleID, rp)
	out := &SecurityRule{}
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
	if out.securityGroupID == "" {
		out.securityGroupID = securityGroupID
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
func (a *securityRulesClientAdapter) Update(ctx context.Context, rule *SecurityRule, opts ...CallOption) (*SecurityRule, error) {
	if err := rule.Err(); err != nil {
		return rule, err
	}
	if rule.ID() == "" {
		return rule, fmt.Errorf("Update: security rule has no ID — call Get first or seed from response metadata")
	}
	if rule.SecurityGroupID() == "" || rule.VPCID() == "" || rule.ProjectID() == "" {
		return rule, fmt.Errorf("Update: security rule has no SecurityGroup — call InSecurityGroup first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, rule.ProjectID(), rule.VPCID(), rule.SecurityGroupID(), rule.ID(), rule.toRequest(), rp)
	populateHTTPEnvelope(&rule.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		rule.fromResponse(resp.Data)
		rule.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, rule)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				rule.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return rule, err
	}
	if resp != nil && !resp.IsSuccess() {
		return rule, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return rule, nil
}

// Delete removes the SecurityRule identified by Ref.
func (a *securityRulesClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, vpcID, securityGroupID, securityRuleID, err := securityRuleIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, vpcID, securityGroupID, securityRuleID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of SecurityRule in the given parent scope.
func (a *securityRulesClientAdapter) List(ctx context.Context, sg Ref, opts ...CallOption) (*List[*SecurityRule], error) {
	projectID, vpcID, securityGroupID, err := securityGroupIDsFromRef(sg)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.List(ctx, projectID, vpcID, securityGroupID, rp)
	if err != nil {
		return nil, err
	}
	if resp != nil && !resp.IsSuccess() {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	var items []*SecurityRule
	if resp != nil && resp.Data != nil {
		items = make([]*SecurityRule, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			rule := &SecurityRule{}
			rule.fromResponse(&resp.Data.Values[i])
			rule.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, rule)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					rule.fromResponse(fresh.Raw())
				}
				return nil
			})
			if rule.securityGroupID == "" {
				rule.securityGroupID = securityGroupID
			}
			if rule.vpcID == "" {
				rule.vpcID = vpcID
			}
			if rule.projectID == "" {
				rule.projectID = projectID
			}
			items = append(items, rule)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*SecurityRule], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*SecurityRule], error) {
		fetch := listPageFetch[types.SecurityRuleListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*SecurityRule
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*SecurityRule, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				item := &SecurityRule{}
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
				if item.securityGroupID == "" {
					item.securityGroupID = securityGroupID
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

// securityRuleIDsFromRef extracts (projectID, vpcID, securityGroupID, securityRuleID) from a Ref.
// Tries typed assertions first, then falls back to URI path parsing.
func securityRuleIDsFromRef(ref Ref) (projectID, vpcID, securityGroupID, securityRuleID string, err error) {
	rid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withSecurityRuleID); ok {
			return w.SecurityRuleID(), true
		}
		return "", false
	}, "securityRules")
	if !ok || rid == "" {
		return "", "", "", "", fmt.Errorf("cannot determine security rule ID from Ref %q", ref.URI())
	}
	sgid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withSecurityGroupID); ok {
			return w.SecurityGroupID(), true
		}
		return "", false
	}, "securityGroups")
	if !ok || sgid == "" {
		return "", "", "", "", fmt.Errorf("cannot determine security group ID from Ref %q", ref.URI())
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
	pid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || pid == "" {
		return "", "", "", "", fmt.Errorf("cannot determine project ID from Ref %q", ref.URI())
	}
	return pid, vid, sgid, rid, nil
}
