package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/network"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// LoadBalancerRef returns a Ref that points to the LoadBalancer with the given IDs.
func LoadBalancerRef(projectID, lbID string) Ref {
	return URI(fmt.Sprintf("/projects/%s/providers/Aruba.Network/loadBalancers/%s", projectID, lbID))
}

// ---- Wrapper ----

// LoadBalancer is the wrapper for an Aruba Cloud Load Balancer (a direct child of a Project).
// LoadBalancer is read-only: instances are obtained via Client.FromNetwork().LoadBalancers().Get/List.
// There is no NewLoadBalancer() factory — the resource cannot be created or mutated through the SDK.
type LoadBalancer struct {
	errMixin
	metadataMixin         // Name(), Tags() — populated from response metadata
	regionalMixin         // Region() — populated from response location
	projectScopedMixin    // ProjectID() — back-filled from response/URI; intoProject() is unexported and unused
	responseMetadataMixin // ID(), RespURI(), CreatedAt(), UpdatedAt(), Version()
	statusMixin           // State(), IsDisabled(), FailureReason(), DisableReasons(), PreviousState()
	linkedMixin           // LinkedResources()
	httpEnvelopeMixin     // RawHTTP(), StatusCode(), Headers(), RawError()

	address  *string                        // Properties.Address (read-only from response)
	vpc      *types.ReferenceResourceCommon // Properties.VPC (linked VPC reference)
	response *types.LoadBalancerResponse    // backs Raw()
}

// Getters — general → specific

// URI satisfies Ref.
func (l *LoadBalancer) URI() string { return l.RespURI() }

// LoadBalancerID satisfies withLoadBalancerID so adapters can extract this ID typed.
func (l *LoadBalancer) LoadBalancerID() string { return l.ID() }

// Raw shadows responseMetadataMixin.Raw() with the full LoadBalancer response.
func (l *LoadBalancer) Raw() *types.LoadBalancerResponse { return l.response }
func (l *LoadBalancer) RawJSON() []byte                  { return marshalRawJSON(l.response) }
func (l *LoadBalancer) RawYAML() []byte                  { return marshalRawYAML(l.response) }

// Address returns the public IP address assigned to this Load Balancer, or "" if absent.
func (l *LoadBalancer) Address() string {
	if l.address == nil {
		return ""
	}
	return *l.address
}

// VPC returns the linked VPC reference URI, or "" if the Load Balancer is not VPC-attached.
func (l *LoadBalancer) VPC() string {
	if l.vpc == nil {
		return ""
	}
	return l.vpc.URI
}

func (l *LoadBalancer) fromResponse(resp *types.LoadBalancerResponse) {
	if resp == nil {
		return
	}
	l.response = resp
	l.setMeta(&resp.Metadata)
	l.named(loadBalancerDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		l.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		l.inRegion(resp.Metadata.LocationResponse.Value)
	}
	l.setStatus(&resp.Status)

	l.setLinked(resp.Properties.LinkedResources)

	if resp.Properties.Address != nil && *resp.Properties.Address != "" {
		addr := *resp.Properties.Address
		l.address = &addr
	}
	if resp.Properties.VPC != nil {
		v := *resp.Properties.VPC
		l.vpc = &v
	}

	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		l.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if l.projectID == "" && l.RespURI() != "" {
		if pid := parseURIIDs(l.RespURI())["projects"]; pid != "" {
			l.projectID = pid
		}
	}
}

func loadBalancerDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// loadBalancerLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type loadBalancerLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.LoadBalancerListResponse], error)
	Get(ctx context.Context, projectID, loadBalancerID string, params *types.RequestParameters) (*types.Response[types.LoadBalancerResponse], error)
}

// ---- Adapter ----

// loadBalancersClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates LoadBalancer ↔ types.LoadBalancerResponse and surfaces HTTP
// errors as *aruba.HTTPError.
type loadBalancersClientAdapter struct {
	low  loadBalancerLowLevelClient
	rest *restclient.Client
}

var _ LoadBalancersClient = (*loadBalancersClientAdapter)(nil)

func newLoadBalancersClientAdapter(rest *restclient.Client) *loadBalancersClientAdapter {
	if rest == nil {
		return &loadBalancersClientAdapter{}
	}
	return &loadBalancersClientAdapter{low: network.NewLoadBalancersClientImpl(rest), rest: rest}
}

// Get fetches a LoadBalancer by Ref and returns a freshly hydrated wrapper.
func (a *loadBalancersClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*LoadBalancer, error) {
	projectID, loadBalancerID, err := loadBalancerIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, loadBalancerID, rp)
	out := &LoadBalancer{}
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

// List returns a paginated list of LoadBalancer in the given parent scope.
func (a *loadBalancersClientAdapter) List(ctx context.Context, project Ref, opts ...CallOption) (*List[*LoadBalancer], error) {
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
	var items []*LoadBalancer
	if resp != nil && resp.Data != nil {
		items = make([]*LoadBalancer, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			lb := &LoadBalancer{}
			lb.fromResponse(&resp.Data.Values[i])
			lb.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, lb)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					lb.fromResponse(fresh.Raw())
				}
				return nil
			})
			if lb.projectID == "" {
				lb.projectID = projectID
			}
			items = append(items, lb)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*LoadBalancer], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*LoadBalancer], error) {
		fetch := listPageFetch[types.LoadBalancerListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*LoadBalancer
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*LoadBalancer, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				lb := &LoadBalancer{}
				lb.fromResponse(&pageResp.Data.Values[i])
				lb.setRefresh(func(ctx context.Context) error {
					fresh, err := a.Get(ctx, lb)
					if err != nil {
						return err
					}
					if fresh != nil && fresh.Raw() != nil {
						lb.fromResponse(fresh.Raw())
					}
					return nil
				})
				if lb.projectID == "" {
					lb.projectID = projectID
				}
				pageItems = append(pageItems, lb)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}

// loadBalancerIDsFromRef extracts (projectID, loadBalancerID) from a Ref. Tries typed
// assertions first, then falls back to URI path parsing.
func loadBalancerIDsFromRef(ref Ref) (projectID, loadBalancerID string, err error) {
	lid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withLoadBalancerID); ok {
			return w.LoadBalancerID(), true
		}
		return "", false
	}, "loadBalancers")
	if !ok || lid == "" {
		return "", "", fmt.Errorf("cannot determine load balancer ID from Ref %q", ref.URI())
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
	return pid, lid, nil
}
