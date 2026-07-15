package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/network"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"k8s.io/utils/ptr"
)

// VPCRef returns a Ref that points to the VPC with the given IDs.
func VPCRef(projectID, vpcID string) Ref {
	return URI(fmt.Sprintf("/projects/%s/providers/Aruba.Network/vpcs/%s", projectID, vpcID))
}

// ---- Wrapper ----

// VPC is the wrapper for an Aruba Cloud VPC (a child of a Project).
// Construct with aruba.NewVPC() and bind it via IntoProject(parent).
//
// Wraps types.VPCResponse / types.VPCRequest. The wrapper carries
// pointer-typed private fields so unset values round-trip through
// the JSON layer correctly.
type VPC struct {
	errMixin
	metadataMixin
	regionalMixin
	projectScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	defaultVPC *bool
	preset     *bool
	response   *types.VPCResponse
}

// NewVPC returns a fresh *VPC ready for fluent setters and a Create call.
// Binds the projectScopedMixin's error sink to the VPC's errMixin so IntoProject
// failures surface via Err().
func NewVPC() *VPC {
	v := &VPC{defaultVPC: ptr.To(false)}
	v.projectScopedMixin = bindProjectScoped(&v.errMixin)
	return v
}

// Setters — chainable, general → specific

// InProject binds this VPC to its parent project. Required before Create.
func (v *VPC) InProject(p Ref) *VPC { v.intoProject(p); return v }

// Named sets the resource name. Required by the API.
func (v *VPC) Named(n string) *VPC { v.named(n); return v }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (v *VPC) Tagged(ts ...string) *VPC {
	for _, t := range ts {
		v.addTag(t)
	}
	return v
}

// Untagged removes each listed tag. No-op for tags not present.
func (v *VPC) Untagged(ts ...string) *VPC {
	for _, t := range ts {
		v.removeTag(t)
	}
	return v
}

// RetaggedAs replaces the entire tag set with the given values.
func (v *VPC) RetaggedAs(ts ...string) *VPC { v.replaceTags(ts...); return v }

// InRegion sets the region for this resource.
func (v *VPC) InRegion(region Region) *VPC { v.inRegion(region); return v }

// AsDefault marks this VPC as the account-region default.
func (v *VPC) AsDefault() *VPC { t := true; v.defaultVPC = &t; return v }

// NotDefault explicitly unsets the default flag.
func (v *VPC) NotDefault() *VPC { f := false; v.defaultVPC = &f; return v }

// WithPreset marks the VPC to use preset networking.
func (v *VPC) WithPreset() *VPC { t := true; v.preset = &t; return v }

// WithoutPreset disables VPC preset networking.
func (v *VPC) WithoutPreset() *VPC { f := false; v.preset = &f; return v }

// Getters — general → specific

// URI satisfies Ref.
func (v *VPC) URI() string { return v.RespURI() }

// VPCID satisfies withVPCID so children's IntoVPC can extract the parent ID.
func (v *VPC) VPCID() string { return v.ID() }

// Raw shadows the promoted responseMetadataMixin.Raw() returning the full response.
func (v *VPC) Raw() *types.VPCResponse { return v.response }
func (v *VPC) RawJSON() []byte         { return marshalRawJSON(v.response) }
func (v *VPC) RawYAML() []byte         { return marshalRawYAML(v.response) }

// RawRequest returns the wire-level request that toRequest() would emit.
func (v *VPC) RawRequest() types.VPCRequest { return v.toRequest() }

// IsDefault returns true if this VPC is the account-region default.
func (v *VPC) IsDefault() bool {
	if v.defaultVPC == nil {
		return false
	}
	return *v.defaultVPC
}

// IsPreset returns true if the VPC was created with a preset subnet/SG.
func (v *VPC) IsPreset() bool {
	if v.preset == nil {
		return false
	}
	return *v.preset
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (v *VPC) toRequest() types.VPCRequest {
	var props *types.VPCPropertiesInnerRequest
	if v.defaultVPC != nil || v.preset != nil {
		props = &types.VPCPropertiesInnerRequest{Default: v.defaultVPC, Preset: v.preset}
	}
	return types.VPCRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: v.toMetadata(),
			Location:                v.toLocation(),
		},
		Properties: types.VPCPropertiesRequest{Properties: props},
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (v *VPC) fromResponse(resp *types.VPCResponse) {
	if resp == nil {
		return
	}
	v.response = resp
	v.setMeta(&resp.Metadata)
	v.named(vpcDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		v.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		v.inRegion(resp.Metadata.LocationResponse.Value)
	}
	v.setStatus(&resp.Status)

	v.setLinked(resp.Properties.LinkedResources)
	d := resp.Properties.Default
	v.defaultVPC = &d
	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		v.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
}

func vpcDerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ---- Low-level client interface ----

// vpcLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type vpcLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.VPCListResponse], error)
	Get(ctx context.Context, projectID, vpcID string, params *types.RequestParameters) (*types.Response[types.VPCResponse], error)
	Create(ctx context.Context, projectID string, body types.VPCRequest, params *types.RequestParameters) (*types.Response[types.VPCResponse], error)
	Update(ctx context.Context, projectID, vpcID string, body types.VPCRequest, params *types.RequestParameters) (*types.Response[types.VPCResponse], error)
	Delete(ctx context.Context, projectID, vpcID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// vpcsClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates VPC ↔ types.VPCRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type vpcsClientAdapter struct {
	low  vpcLowLevelClient
	rest *restclient.Client
}

var _ VPCsClient = (*vpcsClientAdapter)(nil)

func newVPCsClientAdapter(rest *restclient.Client) *vpcsClientAdapter {
	if rest == nil {
		return &vpcsClientAdapter{}
	}
	return &vpcsClientAdapter{low: network.NewVPCsClientImpl(rest), rest: rest}
}

// Create posts a new VPC to the API and hydrates the wrapper from the response.
func (a *vpcsClientAdapter) Create(ctx context.Context, v *VPC, opts ...CallOption) (*VPC, error) {
	if err := v.Err(); err != nil {
		return v, err
	}
	if v.ProjectID() == "" {
		return v, fmt.Errorf("Create: VPC has no project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, v.ProjectID(), v.toRequest(), rp)
	populateHTTPEnvelope(&v.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		v.fromResponse(resp.Data)
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
	}
	if err != nil {
		return v, err
	}
	if resp != nil && !resp.IsSuccess() {
		return v, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return v, nil
}

// Get fetches a VPC by Ref and returns a freshly hydrated wrapper.
func (a *vpcsClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*VPC, error) {
	projectID, vpcID, err := vpcIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, vpcID, rp)
	out := &VPC{}
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
	if err != nil {
		return out, err
	}
	if resp != nil && !resp.IsSuccess() {
		return out, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return out, nil
}

// Update sends a PUT for the current wrapper state. Requires ID and parent.
func (a *vpcsClientAdapter) Update(ctx context.Context, v *VPC, opts ...CallOption) (*VPC, error) {
	if err := v.Err(); err != nil {
		return v, err
	}
	if v.ID() == "" {
		return v, fmt.Errorf("Update: VPC has no ID — call Get first or seed from response metadata")
	}
	if v.ProjectID() == "" {
		return v, fmt.Errorf("Update: VPC has no project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, v.ProjectID(), v.ID(), v.toRequest(), rp)
	populateHTTPEnvelope(&v.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		v.fromResponse(resp.Data)
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
	}
	if err != nil {
		return v, err
	}
	if resp != nil && !resp.IsSuccess() {
		return v, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return v, nil
}

// Delete removes the VPC identified by Ref.
func (a *vpcsClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, vpcID, err := vpcIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, vpcID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of VPC in the given parent scope.
func (a *vpcsClientAdapter) List(ctx context.Context, project Ref, opts ...CallOption) (*List[*VPC], error) {
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
	var items []*VPC
	if resp != nil && resp.Data != nil {
		items = make([]*VPC, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			v := &VPC{}
			v.projectID = projectID
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
			items = append(items, v)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*VPC], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*VPC], error) {
		fetch := listPageFetch[types.VPCListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*VPC
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*VPC, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				item := &VPC{}
				item.projectID = projectID
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
				pageItems = append(pageItems, item)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}

// vpcIDsFromRef extracts (projectID, vpcID) from a Ref. Tries typed assertions
// first, then falls back to URI path parsing.
func vpcIDsFromRef(ref Ref) (projectID, vpcID string, err error) {
	vid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withVPCID); ok {
			return w.VPCID(), true
		}
		return "", false
	}, "vpcs")
	if !ok || vid == "" {
		return "", "", fmt.Errorf("cannot determine VPC ID from Ref %q", ref.URI())
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
	return pid, vid, nil
}
