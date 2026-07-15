package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/compute"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// KeyPair is the wrapper for an Aruba Cloud Compute SSH Key Pair (a direct
// child of a Project). Construct with aruba.NewKeyPair() and bind it via
// IntoProject(project) and WithPublicKey(key).
type KeyPair struct {
	errMixin
	metadataMixin
	regionalMixin
	projectScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	publicKey *string

	response *types.KeyPairResponse
}

// NewKeyPair returns a fresh *KeyPair ready for fluent setters and a Create call.
// Binds projectScopedMixin's error sink so IntoProject failures surface via Err().
func NewKeyPair() *KeyPair {
	k := &KeyPair{}
	k.projectScopedMixin = bindProjectScoped(&k.errMixin)
	return k
}

// Setters — chainable, general → specific

// InProject binds this KeyPair to its parent project. Required before Create.
func (k *KeyPair) InProject(p Ref) *KeyPair { k.intoProject(p); return k }

// Named sets the resource name. Required by the API.
func (k *KeyPair) Named(n string) *KeyPair { k.named(n); return k }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (k *KeyPair) Tagged(ts ...string) *KeyPair {
	for _, t := range ts {
		k.addTag(t)
	}
	return k
}

// Untagged removes each listed tag. No-op for tags not present.
func (k *KeyPair) Untagged(ts ...string) *KeyPair {
	for _, t := range ts {
		k.removeTag(t)
	}
	return k
}

// RetaggedAs replaces the entire tag set with the given values.
func (k *KeyPair) RetaggedAs(ts ...string) *KeyPair { k.replaceTags(ts...); return k }

// InRegion sets the region for this resource.
func (k *KeyPair) InRegion(region Region) *KeyPair { k.inRegion(region); return k }

// WithPublicKey sets the SSH public key (mapped to wire field "value").
func (k *KeyPair) WithPublicKey(key string) *KeyPair {
	k.publicKey = &key
	return k
}

// Getters — general → specific

// URI satisfies Ref.
func (k *KeyPair) URI() string { return k.RespURI() }

// KeyPairID satisfies withKeyPairID.
func (k *KeyPair) KeyPairID() string { return k.ID() }

// Raw shadows responseMetadataMixin.Raw() with the typed key-pair response.
func (k *KeyPair) Raw() *types.KeyPairResponse { return k.response }
func (k *KeyPair) RawJSON() []byte             { return marshalRawJSON(k.response) }
func (k *KeyPair) RawYAML() []byte             { return marshalRawYAML(k.response) }

// RawRequest returns what toRequest() would emit right now.
func (k *KeyPair) RawRequest() types.KeyPairRequest { return k.toRequest() }

// PublicKey returns the SSH public key value ("" if unset). On a hydrated
// response wrapper this surfaces the response's Properties.Value.
func (k *KeyPair) PublicKey() string { return keyPairDerefString(k.publicKey) }

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (k *KeyPair) toRequest() types.KeyPairRequest {
	props := types.KeyPairPropertiesRequest{}
	if k.publicKey != nil {
		props.Value = *k.publicKey
	}
	return types.KeyPairRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: k.toMetadata(),
			Location:                k.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (k *KeyPair) fromResponse(resp *types.KeyPairResponse) {
	if resp == nil {
		return
	}
	k.response = resp
	k.setMeta(&resp.Metadata)
	k.named(keyPairDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		k.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		k.inRegion(resp.Metadata.LocationResponse.Value)
	}
	k.setLinked(resp.Properties.LinkedResources)
	k.setStatus(&resp.Status)

	if resp.Properties.Value != "" {
		v := resp.Properties.Value
		k.publicKey = &v
	}

	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		k.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if k.projectID == "" && k.RespURI() != "" {
		ids := parseURIIDs(k.RespURI())
		k.projectID = ids["projects"]
	}
}

func keyPairDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// keyPairLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type keyPairLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.KeyPairListResponse], error)
	Get(ctx context.Context, projectID, keyPairID string, params *types.RequestParameters) (*types.Response[types.KeyPairResponse], error)
	Create(ctx context.Context, projectID string, body types.KeyPairRequest, params *types.RequestParameters) (*types.Response[types.KeyPairResponse], error)
	Delete(ctx context.Context, projectID, keyPairID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// keyPairsClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates KeyPair ↔ types.KeyPairRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type keyPairsClientAdapter struct {
	low  keyPairLowLevelClient
	rest *restclient.Client
}

var _ KeyPairsClient = (*keyPairsClientAdapter)(nil)

func newKeyPairsClientAdapter(rest *restclient.Client) *keyPairsClientAdapter {
	if rest == nil {
		return &keyPairsClientAdapter{}
	}
	return &keyPairsClientAdapter{low: compute.NewKeyPairsClientImpl(rest), rest: rest}
}

// Create posts a new KeyPair to the API and hydrates the wrapper from the response.
func (a *keyPairsClientAdapter) Create(ctx context.Context, kp *KeyPair, opts ...CallOption) (*KeyPair, error) {
	if err := kp.Err(); err != nil {
		return kp, err
	}
	if kp.ProjectID() == "" {
		return kp, fmt.Errorf("Create: KeyPair has no parent project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, kp.ProjectID(), kp.toRequest(), rp)
	populateHTTPEnvelope(&kp.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		kp.fromResponse(resp.Data)
		kp.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, kp)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				kp.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return kp, err
	}
	if resp != nil && !resp.IsSuccess() {
		return kp, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return kp, nil
}

// Get fetches a KeyPair by Ref and returns a freshly hydrated wrapper.
func (a *keyPairsClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*KeyPair, error) {
	projectID, keyPairID, err := keyPairIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, keyPairID, rp)
	out := &KeyPair{}
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

// Delete removes the KeyPair identified by Ref.
func (a *keyPairsClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, keyPairID, err := keyPairIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, keyPairID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of KeyPair in the given parent scope.
func (a *keyPairsClientAdapter) List(ctx context.Context, project Ref, opts ...CallOption) (*List[*KeyPair], error) {
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
	var items []*KeyPair
	if resp != nil && resp.Data != nil {
		items = make([]*KeyPair, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			kp := &KeyPair{}
			kp.fromResponse(&resp.Data.Values[i])
			if kp.projectID == "" {
				kp.projectID = projectID
			}
			items = append(items, kp)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*KeyPair], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*KeyPair], error) {
		fetch := listPageFetch[types.KeyPairListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*KeyPair
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*KeyPair, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				kp := &KeyPair{}
				kp.fromResponse(&pageResp.Data.Values[i])
				if kp.projectID == "" {
					kp.projectID = projectID
				}
				pageItems = append(pageItems, kp)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}

// keyPairIDsFromRef extracts (projectID, keyPairID) from a Ref.
func keyPairIDsFromRef(ref Ref) (projectID, keyPairID string, err error) {
	kid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withKeyPairID); ok {
			return w.KeyPairID(), true
		}
		return "", false
	}, "keyPairs")
	if !ok || kid == "" {
		return "", "", fmt.Errorf("cannot determine KeyPair ID from Ref %q", ref.URI())
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
