package aruba

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/security"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/async"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// Kmip is the wrapper for a KMIP service nested inside a KMS instance.
// Construct with aruba.NewKmip() and bind via InKMS(parent).
//
// Family B: flat request (no Metadata/Properties boxing, no metadataMixin,
// no tags, no location).
//
// No Update operation. CRUD: Create / Get / Delete / List.
// Action: Download (returns the KMIP certificate key+cert pair).
//
// Identity: KmipResponse.ID carries the resource ID; URI() is constructed
// from (projectID, kmsID, kmipID).
type Kmip struct {
	errMixin
	refreshMixin
	kmsScopedMixin
	responseMetadataMixin // present but never populated; ID/URI shadowed below
	httpEnvelopeMixin

	name *string

	response *types.KmipResponse
}

// NewKmip returns a fresh *Kmip ready for fluent setters and a Create call.
// Binds kmsScopedMixin's error sink so InKMS failures surface via Err().
func NewKmip() *Kmip {
	km := &Kmip{}
	km.kmsScopedMixin = bindKMSScoped(&km.errMixin)
	return km
}

var kmipTerminalStates = map[string]bool{
	string(types.ServiceStatusCertificateAvailable): true,
	string(types.ServiceStatusActive):               true,
	string(types.ServiceStatusFailed):               false,
	string(types.ServiceStatusDeleted):              false,
}

func kmipSuccessTargets() []string {
	return []string{
		string(types.ServiceStatusCertificateAvailable),
		string(types.ServiceStatusActive),
	}
}

// WaitUntilReady blocks until the KMIP service reaches "CertificateAvailable"
// or "Active" — either of which means the certificate is downloadable and the
// service is operational. Provided so all polling-aware resources expose a
// uniform "ready" gate; equivalent to WaitUntilCertificateAvailable for the
// KMIP resource type.
func (km *Kmip) WaitUntilReady(ctx context.Context, opts ...WaitOption) error {
	return km.WaitUntilCertificateAvailable(ctx, opts...)
}

// WaitUntilCertificateAvailable blocks until the KMIP service reaches
// "CertificateAvailable" or "Active" — either of which means the certificate
// is downloadable and the service is operational. Returns immediately with an
// error if the service enters a terminal error state ("Failed", "Deleted"), or
// if the wrapper was not produced by an adapter (Create/Get/List).
func (km *Kmip) WaitUntilCertificateAvailable(ctx context.Context, opts ...WaitOption) error {
	if km.refresh == nil {
		return errors.New("WaitUntilCertificateAvailable: refresh callback not set; resource must be produced by an adapter (Create/Get/List) to support polling")
	}
	cfg := applyWaitOptions(opts)
	call := func(ctx context.Context) (*types.Response[any], error) {
		if err := km.refresh(ctx); err != nil {
			return nil, err
		}
		return &types.Response[any]{}, nil
	}
	var terminalErr error
	check := func(_ *types.Response[any]) (bool, error) {
		state := km.KmipStatus()
		if isSuccess, isTerminal := kmipTerminalStates[state]; isTerminal {
			if isSuccess {
				return true, nil
			}
			terminalErr = fmt.Errorf("KMIP entered terminal state %q (targets %q)", state, kmipSuccessTargets())
			return true, terminalErr
		}
		return false, nil
	}
	_, err := async.WaitFor[any](ctx, cfg.retries, cfg.baseDelay, cfg.timeout, call, check).Await(ctx)
	if terminalErr != nil {
		return terminalErr
	}
	return err
}

// Setters — chainable, general → specific

// InKMS binds this Kmip to its parent KMS instance. Required before Create.
func (km *Kmip) InKMS(parent Ref) *Kmip { km.intoKMS(parent); return km }

// Named sets the resource name. Required by the API.
func (km *Kmip) Named(n string) *Kmip { km.name = &n; return km }

// Getters — general → specific

// ID returns the KMIP service's unique ID from the response, or "" before a Create/Get.
func (km *Kmip) ID() string {
	if km.response != nil && km.response.ID != nil {
		return *km.response.ID
	}
	return ""
}

// KmipID is an alias for ID() and satisfies withKmipID for future child wrappers.
func (km *Kmip) KmipID() string { return km.ID() }

// URI constructs the canonical path for this KMIP service.
// Returns "" if any of projectID, kmsID, or kmipID is missing.
func (km *Kmip) URI() string {
	pid, kid, kmipID := km.ProjectID(), km.KMSID(), km.ID()
	if pid == "" || kid == "" || kmipID == "" {
		return ""
	}
	return fmt.Sprintf("/projects/%s/providers/Aruba.Security/kms/%s/kmips/%s", pid, kid, kmipID)
}

// Raw shadows responseMetadataMixin.Raw() with the typed Kmip response.
func (km *Kmip) Raw() *types.KmipResponse { return km.response }
func (km *Kmip) RawJSON() []byte          { return marshalRawJSON(km.response) }
func (km *Kmip) RawYAML() []byte          { return marshalRawYAML(km.response) }

// RawRequest returns what toRequest() would emit right now.
func (km *Kmip) RawRequest() types.KmipRequest { return km.toRequest() }

// Name returns the KMIP service name, or "" if unset.
func (km *Kmip) Name() string {
	if km.response != nil && km.response.Name != nil {
		return *km.response.Name
	}
	return kmipDeref(km.name)
}

// Type returns the KMIP service type string, or "" if unset.
func (km *Kmip) Type() string {
	if km.response != nil && km.response.Type != nil {
		return *km.response.Type
	}
	return ""
}

// KmipStatus returns the current KMIP service status string, or "" if unset.
func (km *Kmip) KmipStatus() string {
	if km.response != nil && km.response.Status != nil {
		return string(*km.response.Status)
	}
	return ""
}

// CreationDate returns the ISO-8601 creation timestamp, or "" if unset.
func (km *Kmip) CreationDate() string {
	if km.response != nil && km.response.CreationDate != nil {
		return *km.response.CreationDate
	}
	return ""
}

// DeletionDate returns the ISO-8601 deletion timestamp, or "" if unset.
func (km *Kmip) DeletionDate() string {
	if km.response != nil && km.response.DeletionDate != nil {
		return *km.response.DeletionDate
	}
	return ""
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (km *Kmip) toRequest() types.KmipRequest {
	req := types.KmipRequest{}
	if km.name != nil {
		req.Name = *km.name
	}
	return req
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (km *Kmip) fromResponse(resp *types.KmipResponse) {
	if resp == nil {
		return
	}
	km.response = resp
	if resp.Name != nil {
		v := *resp.Name
		km.name = &v
	}
}

func kmipDeref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func kmipIDsFromRef(ref Ref) (projectID, kmsID, kmipID string, err error) {
	kmipID, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withKmipID); ok {
			return w.KmipID(), true
		}
		return "", false
	}, "kmips")
	if !ok || kmipID == "" {
		return "", "", "", fmt.Errorf("cannot determine KMIP ID from Ref %q", ref.URI())
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
	return projectID, kmsID, kmipID, nil
}

// ---- Low-level client interface ----

// kmipsLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type kmipsLowLevelClient interface {
	List(ctx context.Context, projectID, kmsID string, params *types.RequestParameters) (*types.Response[types.KmipListResponse], error)
	Get(ctx context.Context, projectID, kmsID, kmipID string, params *types.RequestParameters) (*types.Response[types.KmipResponse], error)
	Create(ctx context.Context, projectID, kmsID string, body types.KmipRequest, params *types.RequestParameters) (*types.Response[types.KmipResponse], error)
	Delete(ctx context.Context, projectID, kmsID, kmipID string, params *types.RequestParameters) (*types.Response[any], error)
	Download(ctx context.Context, projectID, kmsID, kmipID string, params *types.RequestParameters) (*types.Response[types.KmipCertificateResponse], error)
}

// ---- Adapter ----

// kmipsClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates Kmip ↔ types.KmipRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type kmipsClientAdapter struct {
	low  kmipsLowLevelClient
	rest *restclient.Client
}

func newKmipsClientAdapter(rest *restclient.Client) *kmipsClientAdapter {
	if rest == nil {
		return &kmipsClientAdapter{}
	}
	return &kmipsClientAdapter{low: security.NewKmipClientImpl(rest), rest: rest}
}

// Create posts a new Kmip to the API and hydrates the wrapper from the response.
func (a *kmipsClientAdapter) Create(ctx context.Context, km *Kmip, opts ...CallOption) (*Kmip, error) {
	if err := km.Err(); err != nil {
		return km, err
	}
	if km.ProjectID() == "" {
		return km, fmt.Errorf("Create: Kmip has no parent project — call InKMS first")
	}
	if km.KMSID() == "" {
		return km, fmt.Errorf("Create: Kmip has no parent KMS — call InKMS first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, km.ProjectID(), km.KMSID(), km.toRequest(), rp)
	populateHTTPEnvelope(&km.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		km.fromResponse(resp.Data)
		km.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, km)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				km.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return km, err
	}
	if resp != nil && !resp.IsSuccess() {
		return km, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return km, nil
}

// Get fetches a Kmip by Ref and returns a freshly hydrated wrapper.
func (a *kmipsClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*Kmip, error) {
	projectID, kmsID, kmipID, err := kmipIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, kmsID, kmipID, rp)
	out := &Kmip{}
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

// Delete removes the Kmip identified by Ref.
func (a *kmipsClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, kmsID, kmipID, err := kmipIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, kmsID, kmipID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of Kmip in the given parent scope.
func (a *kmipsClientAdapter) List(ctx context.Context, parent Ref, opts ...CallOption) (*List[*Kmip], error) {
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
	var items []*Kmip
	if resp != nil && resp.Data != nil {
		items = make([]*Kmip, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			km := &Kmip{}
			km.projectID = projectID
			km.kmsID = kmsID
			km.fromResponse(&resp.Data.Values[i])
			km.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, km)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					km.fromResponse(fresh.Raw())
				}
				return nil
			})
			items = append(items, km)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*Kmip], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*Kmip], error) {
		fetch := listPageFetch[types.KmipListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*Kmip
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*Kmip, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				km := &Kmip{}
				km.projectID = projectID
				km.kmsID = kmsID
				km.fromResponse(&pageResp.Data.Values[i])
				km.setRefresh(func(ctx context.Context) error {
					fresh, err := a.Get(ctx, km)
					if err != nil {
						return err
					}
					if fresh != nil && fresh.Raw() != nil {
						km.fromResponse(fresh.Raw())
					}
					return nil
				})
				pageItems = append(pageItems, km)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}

// Download retrieves the KMIP certificate key+cert pair for the given Ref.
func (a *kmipsClientAdapter) Download(ctx context.Context, ref Ref, opts ...CallOption) (*KmipCertificate, error) {
	projectID, kmsID, kmipID, err := kmipIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Download(ctx, projectID, kmsID, kmipID, rp)
	if err != nil {
		return nil, err
	}
	if resp != nil && !resp.IsSuccess() {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	if resp != nil {
		return &KmipCertificate{response: resp.Data}, nil
	}
	return nil, nil
}

// KmipCertificate is the wrapper for a KMIP certificate download response.
// Returned by KmipsClient.Download. Read-only; no factory.
type KmipCertificate struct {
	response *types.KmipCertificateResponse
}

// Cert returns the PEM-encoded certificate string.
func (c *KmipCertificate) Cert() string {
	if c == nil || c.response == nil {
		return ""
	}
	return c.response.Cert
}

// Key returns the PEM-encoded private key string.
func (c *KmipCertificate) Key() string {
	if c == nil || c.response == nil {
		return ""
	}
	return c.response.Key
}

// Raw returns the underlying wire response.
func (c *KmipCertificate) Raw() *types.KmipCertificateResponse {
	if c == nil {
		return nil
	}
	return c.response
}
func (c *KmipCertificate) RawJSON() []byte { return marshalRawJSON(c.response) }
func (c *KmipCertificate) RawYAML() []byte { return marshalRawYAML(c.response) }
