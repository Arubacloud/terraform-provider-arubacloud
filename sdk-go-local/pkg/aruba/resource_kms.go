package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/security"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// KMS is the wrapper for an Aruba Cloud Key Management Service instance
// (a direct child of a Project). Construct with aruba.NewKMS() and bind it
// via IntoProject(project).
//
// Family A: regional, Metadata/Properties envelope, location-aware.
// Supports full CRUD. Create and Update share the same request type (KmsRequest).
//
// The KMS client also exposes Keys() and Kmips() to access nested raw clients
// for cryptographic keys and KMIP services.
//
// Path: /projects/{projectID}/providers/Aruba.Security/kms[/{kmsID}]
type KMS struct {
	errMixin
	metadataMixin
	regionalMixin
	projectScopedMixin
	responseMetadataMixin
	statusMixin
	httpEnvelopeMixin

	billingPeriod *BillingPeriod

	response *types.KmsResponse
}

// NewKMS returns a fresh *KMS ready for fluent setters and a Create call.
// Binds projectScopedMixin's error sink so IntoProject failures surface via Err().
func NewKMS() *KMS {
	k := &KMS{}
	k.projectScopedMixin = bindProjectScoped(&k.errMixin)
	return k
}

// Setters — chainable, general → specific

// InProject binds this KMS to its parent project. Required before Create.
func (k *KMS) InProject(p Ref) *KMS { k.intoProject(p); return k }

// Named sets the resource name. Required by the API.
func (k *KMS) Named(n string) *KMS { k.named(n); return k }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (k *KMS) Tagged(ts ...string) *KMS {
	for _, t := range ts {
		k.addTag(t)
	}
	return k
}

// Untagged removes each listed tag. No-op for tags not present.
func (k *KMS) Untagged(ts ...string) *KMS {
	for _, t := range ts {
		k.removeTag(t)
	}
	return k
}

// RetaggedAs replaces the entire tag set with the given values.
func (k *KMS) RetaggedAs(ts ...string) *KMS { k.replaceTags(ts...); return k }

// InRegion sets the region for this resource.
func (k *KMS) InRegion(region Region) *KMS { k.inRegion(region); return k }

// BilledBy sets the billing cadence. Accepted periods are resource-specific; check the API reference.
func (k *KMS) BilledBy(period BillingPeriod) *KMS { k.billingPeriod = &period; return k }

// Getters — general → specific

// URI satisfies Ref by returning the server-assigned canonical URI, or "" if Create hasn't run yet.
func (k *KMS) URI() string { return k.RespURI() }

// KMSID satisfies withKMSID so child wrappers can extract this ID by typed assertion.
func (k *KMS) KMSID() string { return k.ID() }

// Raw shadows responseMetadataMixin.Raw() with the typed KMS response.
func (k *KMS) Raw() *types.KmsResponse { return k.response }
func (k *KMS) RawJSON() []byte         { return marshalRawJSON(k.response) }
func (k *KMS) RawYAML() []byte         { return marshalRawYAML(k.response) }

// RawRequest returns what toRequest() would emit right now.
func (k *KMS) RawRequest() types.KmsRequest { return k.toRequest() }

// BillingPeriod returns the billing period for this KMS instance, or "" if unset.
func (k *KMS) BillingPeriod() BillingPeriod {
	if k.response != nil && k.response.Properties.BillingPeriod != nil {
		return *k.response.Properties.BillingPeriod
	}
	if k.billingPeriod != nil {
		return *k.billingPeriod
	}
	return ""
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (k *KMS) toRequest() types.KmsRequest {
	props := types.KmsPropertiesRequest{
		BillingPeriod: defaultBillingPeriod(k.billingPeriod),
	}
	return types.KmsRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: k.toMetadata(),
			Location:                k.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (k *KMS) fromResponse(resp *types.KmsResponse) {
	if resp == nil {
		return
	}
	k.response = resp
	k.setMeta(&resp.Metadata)
	k.named(kmsDeref(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		k.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		k.inRegion(resp.Metadata.LocationResponse.Value)
	}
	k.setStatus(&resp.Status)

	if resp.Properties.BillingPeriod != nil {
		k.billingPeriod = resp.Properties.BillingPeriod
	}

	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		k.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if k.projectID == "" && k.RespURI() != "" {
		if pid := parseURIIDs(k.RespURI())["projects"]; pid != "" {
			k.projectID = pid
		}
	}
}

func kmsDeref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func kmsIDsFromRef(ref Ref) (projectID, kmsID string, err error) {
	kid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withKMSID); ok {
			return w.KMSID(), true
		}
		return "", false
	}, "kms")
	if !ok || kid == "" {
		return "", "", fmt.Errorf("cannot determine KMS ID from Ref %q", ref.URI())
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

// ---- Low-level client interface ----

// kmsLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type kmsLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.KmsListResponse], error)
	Get(ctx context.Context, projectID, kmsID string, params *types.RequestParameters) (*types.Response[types.KmsResponse], error)
	Create(ctx context.Context, projectID string, body types.KmsRequest, params *types.RequestParameters) (*types.Response[types.KmsResponse], error)
	Update(ctx context.Context, projectID, kmsID string, body types.KmsRequest, params *types.RequestParameters) (*types.Response[types.KmsResponse], error)
	Delete(ctx context.Context, projectID, kmsID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// kmsClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates KMS ↔ types.KmsRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type kmsClientAdapter struct {
	low  kmsLowLevelClient
	rest *restclient.Client
}

func newKMSClientAdapter(rest *restclient.Client) *kmsClientAdapter {
	if rest == nil {
		return &kmsClientAdapter{}
	}
	return &kmsClientAdapter{
		low:  security.NewKMSClientImpl(rest),
		rest: rest,
	}
}

// Create posts a new KMS to the API and hydrates the wrapper from the response.
func (a *kmsClientAdapter) Create(ctx context.Context, k *KMS, opts ...CallOption) (*KMS, error) {
	if err := k.Err(); err != nil {
		return k, err
	}
	if k.ProjectID() == "" {
		return k, fmt.Errorf("Create: KMS has no parent project — call InProject first")
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
	if err != nil {
		return k, err
	}
	if resp != nil && !resp.IsSuccess() {
		return k, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return k, nil
}

// Update sends a PUT for the current wrapper state. Requires ID and parent.
func (a *kmsClientAdapter) Update(ctx context.Context, k *KMS, opts ...CallOption) (*KMS, error) {
	if err := k.Err(); err != nil {
		return k, err
	}
	if k.KMSID() == "" {
		return k, fmt.Errorf("Update: KMS has no ID")
	}
	if k.ProjectID() == "" {
		return k, fmt.Errorf("Update: KMS has no parent project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, k.ProjectID(), k.KMSID(), k.toRequest(), rp)
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
	if err != nil {
		return k, err
	}
	if resp != nil && !resp.IsSuccess() {
		return k, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return k, nil
}

// Get fetches a KMS by Ref and returns a freshly hydrated wrapper.
func (a *kmsClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*KMS, error) {
	projectID, kmsID, err := kmsIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, kmsID, rp)
	out := &KMS{}
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

// Delete removes the KMS identified by Ref.
func (a *kmsClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, kmsID, err := kmsIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, kmsID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of KMS in the given parent scope.
func (a *kmsClientAdapter) List(ctx context.Context, parent Ref, opts ...CallOption) (*List[*KMS], error) {
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
	var items []*KMS
	if resp != nil && resp.Data != nil {
		items = make([]*KMS, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			k := &KMS{}
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
			items = append(items, k)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*KMS], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*KMS], error) {
		fetch := listPageFetch[types.KmsListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*KMS
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*KMS, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				k := &KMS{}
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
				pageItems = append(pageItems, k)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}
