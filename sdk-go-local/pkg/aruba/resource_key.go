package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/security"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// Key is the wrapper for a cryptographic key nested inside a KMS instance.
// Construct with aruba.NewKey() and bind via InKMS(parent).
//
// Family B: flat request (no Metadata/Properties boxing, no metadataMixin,
// no tags, no location).
//
// No Update operation. CRUD: Create / Get / Delete / List.
//
// Identity: KeyResponse carries no ResourceMetadataResponse; ID() and KeyID()
// read from KeyResponse.KeyID, and URI() is constructed from (projectID, kmsID, keyID).
type Key struct {
	errMixin
	refreshMixin
	kmsScopedMixin
	responseMetadataMixin // present but never populated; ID/URI shadowed below
	httpEnvelopeMixin

	name      *string
	algorithm *types.KeyAlgorithm
	response  *types.KeyResponse
}

// NewKey returns a fresh *Key ready for fluent setters and a Create call.
// Binds kmsScopedMixin's error sink so InKMS failures surface via Err().
func NewKey() *Key {
	k := &Key{}
	k.kmsScopedMixin = bindKMSScoped(&k.errMixin)
	return k
}

// Setters — chainable, general → specific

// InKMS binds this Key to its parent KMS instance. Required before Create.
func (k *Key) InKMS(parent Ref) *Key { k.intoKMS(parent); return k }

// Named sets the key name. Required by the API.
func (k *Key) Named(n string) *Key { k.name = &n; return k }

// OfAlgorithm sets the cryptographic algorithm for this key.
func (k *Key) OfAlgorithm(a types.KeyAlgorithm) *Key { k.algorithm = &a; return k }

// Getters — general → specific

// ID returns the key's unique ID from the response, or "" before a Create/Get.
func (k *Key) ID() string {
	if k.response != nil && k.response.KeyID != nil {
		return *k.response.KeyID
	}
	return ""
}

// KeyID is an alias for ID() and satisfies withKeyID for future child wrappers.
func (k *Key) KeyID() string { return k.ID() }

// URI constructs the canonical path for this key.
// Returns "" if any of projectID, kmsID, or keyID is missing.
func (k *Key) URI() string {
	pid, kid, keyID := k.ProjectID(), k.KMSID(), k.ID()
	if pid == "" || kid == "" || keyID == "" {
		return ""
	}
	return fmt.Sprintf("/projects/%s/providers/Aruba.Security/kms/%s/keys/%s", pid, kid, keyID)
}

// Raw shadows responseMetadataMixin.Raw() with the typed Key response.
func (k *Key) Raw() *types.KeyResponse { return k.response }
func (k *Key) RawJSON() []byte         { return marshalRawJSON(k.response) }
func (k *Key) RawYAML() []byte         { return marshalRawYAML(k.response) }

// RawRequest returns what toRequest() would emit right now.
func (k *Key) RawRequest() types.KeyRequest { return k.toRequest() }

// Name returns the key name from the response, or the locally-set name if unhydrated.
func (k *Key) Name() string {
	if k.response != nil && k.response.Name != nil {
		return *k.response.Name
	}
	return keyDeref(k.name)
}

// Algorithm returns the cryptographic algorithm from the response, or the locally-set value if unhydrated.
func (k *Key) Algorithm() types.KeyAlgorithm {
	if k.response != nil && k.response.Algorithm != nil {
		return *k.response.Algorithm
	}
	if k.algorithm != nil {
		return *k.algorithm
	}
	return ""
}

// Type returns the key type from the response, or "" if unhydrated.
func (k *Key) Type() string {
	if k.response != nil && k.response.Type != nil {
		return string(*k.response.Type)
	}
	return ""
}

// KeyStatus returns the lifecycle status of the key from the response, or "" if unhydrated.
func (k *Key) KeyStatus() string {
	if k.response != nil && k.response.Status != nil {
		return string(*k.response.Status)
	}
	return ""
}

// CreationSource returns the creation source of the key from the response, or "" if unhydrated.
func (k *Key) CreationSource() string {
	if k.response != nil && k.response.CreationSource != nil {
		return string(*k.response.CreationSource)
	}
	return ""
}

// PrivateKeyID returns the associated private key ID from the response, or "" if unhydrated.
func (k *Key) PrivateKeyID() string {
	if k.response != nil && k.response.PrivateKeyID != nil {
		return *k.response.PrivateKeyID
	}
	return ""
}

// Wire converters

// toRequest assembles the Create body from current setter state. Defaults are applied at the wire boundary.
func (k *Key) toRequest() types.KeyRequest {
	req := types.KeyRequest{}
	if k.name != nil {
		req.Name = *k.name
	}
	if k.algorithm != nil {
		req.Algorithm = *k.algorithm
	}
	return req
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (k *Key) fromResponse(resp *types.KeyResponse) {
	if resp == nil {
		return
	}
	k.response = resp
	if resp.Name != nil {
		v := *resp.Name
		k.name = &v
	}
	if resp.Algorithm != nil {
		v := *resp.Algorithm
		k.algorithm = &v
	}
}

func keyDeref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func keyIDsFromRef(ref Ref) (projectID, kmsID, keyID string, err error) {
	keyID, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withKeyID); ok {
			return w.KeyID(), true
		}
		return "", false
	}, "keys")
	if !ok || keyID == "" {
		return "", "", "", fmt.Errorf("cannot determine key ID from Ref %q", ref.URI())
	}
	kmsID, ok = extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withKMSID); ok {
			return w.KMSID(), true
		}
		return "", false
	}, "kms")
	if !ok || kmsID == "" {
		return "", "", "", fmt.Errorf("cannot determine KMS ID from Ref %q", ref.URI())
	}
	projectID, ok = extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || projectID == "" {
		return "", "", "", fmt.Errorf("cannot determine project ID from Ref %q", ref.URI())
	}
	return projectID, kmsID, keyID, nil
}

// ---- Low-level client interface ----

// keysLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type keysLowLevelClient interface {
	List(ctx context.Context, projectID, kmsID string, params *types.RequestParameters) (*types.Response[types.KeyListResponse], error)
	Get(ctx context.Context, projectID, kmsID, keyID string, params *types.RequestParameters) (*types.Response[types.KeyResponse], error)
	Create(ctx context.Context, projectID, kmsID string, body types.KeyRequest, params *types.RequestParameters) (*types.Response[types.KeyResponse], error)
	Delete(ctx context.Context, projectID, kmsID, keyID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// keysClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates Key ↔ types.KeyRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type keysClientAdapter struct {
	low  keysLowLevelClient
	rest *restclient.Client
}

var _ KeysClient = (*keysClientAdapter)(nil)

func newKeysClientAdapter(rest *restclient.Client) *keysClientAdapter {
	if rest == nil {
		return &keysClientAdapter{}
	}
	return &keysClientAdapter{low: security.NewKeyClientImpl(rest), rest: rest}
}

// Create posts a new Key to the API and hydrates the wrapper from the response.
func (a *keysClientAdapter) Create(ctx context.Context, k *Key, opts ...CallOption) (*Key, error) {
	if err := k.Err(); err != nil {
		return k, err
	}
	if k.ProjectID() == "" {
		return k, fmt.Errorf("Create: Key has no parent project — call InKMS first")
	}
	if k.KMSID() == "" {
		return k, fmt.Errorf("Create: Key has no parent KMS — call InKMS first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, k.ProjectID(), k.KMSID(), k.toRequest(), rp)
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

// Get fetches a Key by Ref and returns a freshly hydrated wrapper.
func (a *keysClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*Key, error) {
	projectID, kmsID, keyID, err := keyIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, kmsID, keyID, rp)
	out := &Key{}
	out.projectID = projectID
	out.kmsID = kmsID
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

// Delete removes the Key identified by Ref.
func (a *keysClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, kmsID, keyID, err := keyIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, kmsID, keyID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of Key in the given KMS parent scope.
func (a *keysClientAdapter) List(ctx context.Context, parent Ref, opts ...CallOption) (*List[*Key], error) {
	projectID, kmsID, err := kmsIDsFromRef(parent)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.List(ctx, projectID, kmsID, rp)
	if err != nil {
		return nil, err
	}
	if resp != nil && !resp.IsSuccess() {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	var items []*Key
	if resp != nil && resp.Data != nil {
		items = make([]*Key, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			k := &Key{}
			k.projectID = projectID
			k.kmsID = kmsID
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
			items = append(items, k)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*Key], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*Key], error) {
		fetch := listPageFetch[types.KeyListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*Key
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*Key, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				k := &Key{}
				k.projectID = projectID
				k.kmsID = kmsID
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
				pageItems = append(pageItems, k)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}
