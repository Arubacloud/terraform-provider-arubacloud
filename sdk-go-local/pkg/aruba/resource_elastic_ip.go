package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/network"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ElasticIPRef returns a Ref that points to the ElasticIP with the given IDs.
func ElasticIPRef(projectID, eipID string) Ref {
	return URI(fmt.Sprintf("/projects/%s/providers/Aruba.Network/elasticIps/%s", projectID, eipID))
}

// ---- Wrapper ----

// ElasticIP is the wrapper for an Aruba Cloud Elastic IP (a child of a Project).
// Construct with aruba.NewElasticIP() and bind it via IntoProject(project).
//
// Wraps types.ElasticIPResponse / types.ElasticIPRequest. The wrapper carries
// pointer-typed private fields so unset values round-trip through
// the JSON layer correctly.
type ElasticIP struct {
	errMixin
	metadataMixin
	regionalMixin
	projectScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	billingPeriod *BillingPeriod           // Properties.BillingPlanCommon.BillingPeriod
	address       *string                  // Properties.Address (read-only from response)
	response      *types.ElasticIPResponse // backs Raw()
}

// NewElasticIP returns a fresh *ElasticIP ready for fluent setters and a Create call.
// Binds projectScopedMixin's error sink so IntoProject failures surface via Err().
func NewElasticIP() *ElasticIP {
	e := &ElasticIP{}
	e.projectScopedMixin = bindProjectScoped(&e.errMixin)
	return e
}

// Setters — chainable, general → specific

// InProject binds this ElasticIP to its parent project. Required before Create.
func (e *ElasticIP) InProject(p Ref) *ElasticIP { e.intoProject(p); return e }

// Named sets the resource name. Required by the API.
func (e *ElasticIP) Named(n string) *ElasticIP { e.named(n); return e }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (e *ElasticIP) Tagged(ts ...string) *ElasticIP {
	for _, t := range ts {
		e.addTag(t)
	}
	return e
}

// Untagged removes each listed tag. No-op for tags not present.
func (e *ElasticIP) Untagged(ts ...string) *ElasticIP {
	for _, t := range ts {
		e.removeTag(t)
	}
	return e
}

// RetaggedAs replaces the entire tag set with the given values.
func (e *ElasticIP) RetaggedAs(ts ...string) *ElasticIP { e.replaceTags(ts...); return e }

// InRegion sets the region for this resource.
func (e *ElasticIP) InRegion(region Region) *ElasticIP { e.inRegion(region); return e }

// BilledBy sets the billing cadence. Accepted periods are resource-specific; check the API reference.
func (e *ElasticIP) BilledBy(period BillingPeriod) *ElasticIP { e.billingPeriod = &period; return e }

// Getters — general → specific

// URI satisfies Ref.
func (e *ElasticIP) URI() string { return e.RespURI() }

// ElasticIPID satisfies withElasticIPID so adapters can extract this ID typed.
func (e *ElasticIP) ElasticIPID() string { return e.ID() }

// Raw shadows responseMetadataMixin.Raw() with the full ElasticIP response.
func (e *ElasticIP) Raw() *types.ElasticIPResponse { return e.response }
func (e *ElasticIP) RawJSON() []byte               { return marshalRawJSON(e.response) }
func (e *ElasticIP) RawYAML() []byte               { return marshalRawYAML(e.response) }

// RawRequest returns what toRequest() would emit right now.
func (e *ElasticIP) RawRequest() types.ElasticIPRequest { return e.toRequest() }

// BillingPeriod returns the configured billing period ("" if unset).
func (e *ElasticIP) BillingPeriod() BillingPeriod {
	if e.billingPeriod == nil {
		return ""
	}
	return *e.billingPeriod
}

// Address returns the server-assigned public IP address ("" if unassigned).
func (e *ElasticIP) Address() string {
	if e.address == nil {
		return ""
	}
	return *e.address
}

// AssociatedResourceURI returns the URI of the first linked resource ("" if none).
// This is the resource the elastic IP is currently attached to (e.g. a cloud server).
func (e *ElasticIP) AssociatedResourceURI() string {
	linked := e.LinkedResources()
	if len(linked) == 0 {
		return ""
	}
	return linked[0].URI
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (e *ElasticIP) toRequest() types.ElasticIPRequest {
	return types.ElasticIPRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: e.toMetadata(),
			Location:                e.toLocation(),
		},
		Properties: types.ElasticIPPropertiesRequest{
			BillingPlanCommon: &types.BillingPlanCommon{BillingPeriod: defaultBillingPeriod(e.billingPeriod)},
		},
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (e *ElasticIP) fromResponse(resp *types.ElasticIPResponse) {
	if resp == nil {
		return
	}
	e.response = resp
	e.setMeta(&resp.Metadata)
	e.named(elasticIPDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		e.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		e.inRegion(resp.Metadata.LocationResponse.Value)
	}
	e.setStatus(&resp.Status)

	e.setLinked(resp.Properties.LinkedResources)

	if resp.Properties.BillingPlanCommon != nil && resp.Properties.BillingPlanCommon.BillingPeriod != nil {
		e.billingPeriod = resp.Properties.BillingPlanCommon.BillingPeriod
	}
	if resp.Properties.Address != nil && *resp.Properties.Address != "" {
		addr := *resp.Properties.Address
		e.address = &addr
	}

	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		e.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if e.projectID == "" && e.RespURI() != "" {
		if pid := parseURIIDs(e.RespURI())["projects"]; pid != "" {
			e.projectID = pid
		}
	}
}

func elasticIPDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// WaitUntilNotUsed blocks until the ElasticIP reaches the "NotUsed" state —
// the steady terminal state for an unattached EIP. Call this after Create and
// before passing the EIP to a CloudServer, ContainerRegistry, or LoadBalancer.
func (e *ElasticIP) WaitUntilNotUsed(ctx context.Context, opts ...WaitOption) error {
	return e.WaitUntilStates(ctx, []types.State{types.StateNotUsed}, opts...)
}

// WaitUntilUsed blocks until the ElasticIP is bound to a consumer resource.
// The platform may emit "InUse", "Used", or "Reserved" (bound as a dependency
// but not actively in use); this method succeeds on whichever arrives first.
func (e *ElasticIP) WaitUntilUsed(ctx context.Context, opts ...WaitOption) error {
	return e.WaitUntilStates(ctx, []types.State{types.StateInUse, types.StateUsed, types.StateReserved}, opts...)
}

// ---- Low-level client interface ----

// elasticIPLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type elasticIPLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.ElasticIPListResponse], error)
	Get(ctx context.Context, projectID, elasticIPID string, params *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error)
	Create(ctx context.Context, projectID string, body types.ElasticIPRequest, params *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error)
	Update(ctx context.Context, projectID, elasticIPID string, body types.ElasticIPRequest, params *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error)
	Delete(ctx context.Context, projectID, elasticIPID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// elasticIPsClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates ElasticIP ↔ types.ElasticIPRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type elasticIPsClientAdapter struct {
	low  elasticIPLowLevelClient
	rest *restclient.Client
}

var _ ElasticIPsClient = (*elasticIPsClientAdapter)(nil)

func newElasticIPsClientAdapter(rest *restclient.Client) *elasticIPsClientAdapter {
	if rest == nil {
		return &elasticIPsClientAdapter{}
	}
	return &elasticIPsClientAdapter{low: network.NewElasticIPsClientImpl(rest), rest: rest}
}

// Create posts a new ElasticIP to the API and hydrates the wrapper from the response.
func (a *elasticIPsClientAdapter) Create(ctx context.Context, e *ElasticIP, opts ...CallOption) (*ElasticIP, error) {
	if err := e.Err(); err != nil {
		return e, err
	}
	if e.ProjectID() == "" {
		return e, fmt.Errorf("Create: elastic IP has no project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, e.ProjectID(), e.toRequest(), rp)
	populateHTTPEnvelope(&e.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		e.fromResponse(resp.Data)
		e.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, e)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				e.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return e, err
	}
	if resp != nil && !resp.IsSuccess() {
		return e, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return e, nil
}

// Get fetches an ElasticIP by Ref and returns a freshly hydrated wrapper.
func (a *elasticIPsClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*ElasticIP, error) {
	projectID, elasticIPID, err := elasticIPIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, elasticIPID, rp)
	out := &ElasticIP{}
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
func (a *elasticIPsClientAdapter) Update(ctx context.Context, e *ElasticIP, opts ...CallOption) (*ElasticIP, error) {
	if err := e.Err(); err != nil {
		return e, err
	}
	if e.ID() == "" {
		return e, fmt.Errorf("Update: elastic IP has no ID — call Get first or seed from response metadata")
	}
	if e.ProjectID() == "" {
		return e, fmt.Errorf("Update: elastic IP has no project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, e.ProjectID(), e.ID(), e.toRequest(), rp)
	populateHTTPEnvelope(&e.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		e.fromResponse(resp.Data)
		e.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, e)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				e.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return e, err
	}
	if resp != nil && !resp.IsSuccess() {
		return e, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return e, nil
}

// Delete removes the ElasticIP identified by Ref.
func (a *elasticIPsClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, elasticIPID, err := elasticIPIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, elasticIPID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of ElasticIP in the given parent scope.
func (a *elasticIPsClientAdapter) List(ctx context.Context, project Ref, opts ...CallOption) (*List[*ElasticIP], error) {
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
	var items []*ElasticIP
	if resp != nil && resp.Data != nil {
		items = make([]*ElasticIP, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			e := &ElasticIP{}
			e.fromResponse(&resp.Data.Values[i])
			e.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, e)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					e.fromResponse(fresh.Raw())
				}
				return nil
			})
			if e.projectID == "" {
				e.projectID = projectID
			}
			items = append(items, e)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*ElasticIP], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*ElasticIP], error) {
		fetch := listPageFetch[types.ElasticIPListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*ElasticIP
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*ElasticIP, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				e := &ElasticIP{}
				e.fromResponse(&pageResp.Data.Values[i])
				e.setRefresh(func(ctx context.Context) error {
					fresh, err := a.Get(ctx, e)
					if err != nil {
						return err
					}
					if fresh != nil && fresh.Raw() != nil {
						e.fromResponse(fresh.Raw())
					}
					return nil
				})
				if e.projectID == "" {
					e.projectID = projectID
				}
				pageItems = append(pageItems, e)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}

// elasticIPIDsFromRef extracts (projectID, elasticIPID) from a Ref. Tries typed
// assertions first, then falls back to URI path parsing.
func elasticIPIDsFromRef(ref Ref) (projectID, elasticIPID string, err error) {
	eid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withElasticIPID); ok {
			return w.ElasticIPID(), true
		}
		return "", false
	}, "elasticIps")
	if !ok || eid == "" {
		return "", "", fmt.Errorf("cannot determine elastic IP ID from Ref %q", ref.URI())
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
	return pid, eid, nil
}
