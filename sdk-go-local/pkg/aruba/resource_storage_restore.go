package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/storage"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// StorageRestore is the wrapper for an Aruba Cloud Storage Restore (a direct
// child of a StorageBackup, grandchild of a Project). Construct with
// aruba.NewStorageRestore() and bind it via FromBackup(backup) and ToVolume(volume).
type StorageRestore struct {
	errMixin
	metadataMixin
	regionalMixin
	backupScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	targetRef *string // body URI (request "Target" / response "Destination")

	response *types.StorageRestoreResponse
}

// NewStorageRestore returns a fresh *StorageRestore ready for fluent setters and a Create call.
// Binds backupScopedMixin's error sink so IntoBackup failures surface via Err().
func NewStorageRestore() *StorageRestore {
	r := &StorageRestore{}
	r.backupScopedMixin = bindBackupScoped(&r.errMixin)
	return r
}

// Setters — chainable, general → specific

// FromBackup binds this StorageRestore to its parent StorageBackup. Required before Create.
func (r *StorageRestore) FromBackup(b Ref) *StorageRestore { r.intoBackup(b); return r }

// Named sets the resource name. Required by the API.
func (r *StorageRestore) Named(n string) *StorageRestore { r.named(n); return r }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (r *StorageRestore) Tagged(ts ...string) *StorageRestore {
	for _, t := range ts {
		r.addTag(t)
	}
	return r
}

// Untagged removes each listed tag. No-op for tags not present.
func (r *StorageRestore) Untagged(ts ...string) *StorageRestore {
	for _, t := range ts {
		r.removeTag(t)
	}
	return r
}

// RetaggedAs replaces the entire tag set with the given values.
func (r *StorageRestore) RetaggedAs(ts ...string) *StorageRestore { r.replaceTags(ts...); return r }

// InRegion sets the region for this resource.
func (r *StorageRestore) InRegion(region Region) *StorageRestore { r.inRegion(region); return r }

// ToVolume binds the destination volume (where the backup will be restored to)
// via its URI. Pass any Ref (typed or aruba.URI(...)). Empty URIs are recorded
// on the error sink and the field remains unset.
func (r *StorageRestore) ToVolume(vol Ref) *StorageRestore {
	uri := vol.URI()
	if uri == "" {
		r.addErr(fmt.Errorf("ToVolume: empty URI"))
		return r
	}
	r.targetRef = &uri
	return r
}

// Getters — general → specific

// URI satisfies Ref.
func (r *StorageRestore) URI() string { return r.RespURI() }

// RestoreID satisfies withRestoreID.
func (r *StorageRestore) RestoreID() string { return r.ID() }

// Raw shadows responseMetadataMixin.Raw() with the typed restore response.
func (r *StorageRestore) Raw() *types.StorageRestoreResponse { return r.response }
func (r *StorageRestore) RawJSON() []byte                    { return marshalRawJSON(r.response) }
func (r *StorageRestore) RawYAML() []byte                    { return marshalRawYAML(r.response) }

// RawRequest returns what toRequest() would emit right now.
func (r *StorageRestore) RawRequest() types.StorageRestoreRequest { return r.toRequest() }

// TargetURI returns the destination volume URI ("" if unset).
// On a hydrated response wrapper this surfaces the response's "Destination" field.
func (r *StorageRestore) TargetURI() string { return storageRestoreDerefString(r.targetRef) }

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (r *StorageRestore) toRequest() types.StorageRestoreRequest {
	props := types.StorageRestorePropertiesRequest{}
	if r.targetRef != nil {
		props.Target = types.ReferenceResourceCommon{URI: *r.targetRef}
	}
	return types.StorageRestoreRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: r.toMetadata(),
			Location:                r.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (r *StorageRestore) fromResponse(resp *types.StorageRestoreResponse) {
	if resp == nil {
		return
	}
	r.response = resp
	r.setMeta(&resp.Metadata)
	r.named(storageRestoreDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		r.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		r.inRegion(resp.Metadata.LocationResponse.Value)
	}
	r.setStatus(&resp.Status)

	// Response shape uses Destination (not Target).
	if resp.Properties.Destination.URI != "" {
		v := resp.Properties.Destination.URI
		r.targetRef = &v
	}

	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		r.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if (r.projectID == "" || r.backupID == "") && r.RespURI() != "" {
		ids := parseURIIDs(r.RespURI())
		if r.projectID == "" {
			r.projectID = ids["projects"]
		}
		if r.backupID == "" {
			r.backupID = ids["backups"]
		}
	}
}

func storageRestoreDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// storageRestoreLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type storageRestoreLowLevelClient interface {
	List(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[types.StorageRestoreListResponse], error)
	Get(ctx context.Context, projectID, backupID, restoreID string, params *types.RequestParameters) (*types.Response[types.StorageRestoreResponse], error)
	Create(ctx context.Context, projectID, backupID string, body types.StorageRestoreRequest, params *types.RequestParameters) (*types.Response[types.StorageRestoreResponse], error)
	Update(ctx context.Context, projectID, backupID, restoreID string, body types.StorageRestoreRequest, params *types.RequestParameters) (*types.Response[types.StorageRestoreResponse], error)
	Delete(ctx context.Context, projectID, backupID, restoreID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// storageRestoresClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates StorageRestore ↔ types.StorageRestoreRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type storageRestoresClientAdapter struct {
	low  storageRestoreLowLevelClient
	rest *restclient.Client
}

var _ StorageRestoreClient = (*storageRestoresClientAdapter)(nil)

func newStorageRestoresClientAdapter(rest *restclient.Client) *storageRestoresClientAdapter {
	if rest == nil {
		return &storageRestoresClientAdapter{}
	}
	// NewRestoreClientImpl panics if backupClient is nil — instantiate one internally
	// so the public adapter constructor stays single-arg.
	return &storageRestoresClientAdapter{
		low:  storage.NewRestoreClientImpl(rest, storage.NewBackupClientImpl(rest)),
		rest: rest,
	}
}

// Create posts a new StorageRestore to the API and hydrates the wrapper from the response.
func (a *storageRestoresClientAdapter) Create(ctx context.Context, r *StorageRestore, opts ...CallOption) (*StorageRestore, error) {
	if err := r.Err(); err != nil {
		return r, err
	}
	if r.BackupID() == "" || r.ProjectID() == "" {
		return r, fmt.Errorf("Create: StorageRestore has no parent backup — call FromBackup first")
	}
	if r.targetRef == nil {
		return r, fmt.Errorf("Create: StorageRestore has no target — call ToVolume first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, r.ProjectID(), r.BackupID(), r.toRequest(), rp)
	populateHTTPEnvelope(&r.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		r.fromResponse(resp.Data)
		r.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, r)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				r.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return r, err
	}
	if resp != nil && !resp.IsSuccess() {
		return r, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return r, nil
}

// Get fetches a StorageRestore by Ref and returns a freshly hydrated wrapper.
func (a *storageRestoresClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*StorageRestore, error) {
	projectID, backupID, restoreID, err := restoreIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, backupID, restoreID, rp)
	out := &StorageRestore{}
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
	if out.backupID == "" {
		out.backupID = backupID
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
// NOTE: platform support for PUT on restore resources is not currently documented;
// callers may receive a 4xx response. Prefer Create+Delete workflows where possible.
func (a *storageRestoresClientAdapter) Update(ctx context.Context, r *StorageRestore, opts ...CallOption) (*StorageRestore, error) {
	if err := r.Err(); err != nil {
		return r, err
	}
	if r.ID() == "" {
		return r, fmt.Errorf("Update: StorageRestore has no ID — call Get first or seed from response metadata")
	}
	if r.BackupID() == "" || r.ProjectID() == "" {
		return r, fmt.Errorf("Update: StorageRestore has no parent backup — call FromBackup first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, r.ProjectID(), r.BackupID(), r.ID(), r.toRequest(), rp)
	populateHTTPEnvelope(&r.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		r.fromResponse(resp.Data)
		r.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, r)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				r.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return r, err
	}
	if resp != nil && !resp.IsSuccess() {
		return r, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return r, nil
}

// Delete removes the StorageRestore identified by Ref.
func (a *storageRestoresClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, backupID, restoreID, err := restoreIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, backupID, restoreID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of StorageRestore entries for the given backup.
func (a *storageRestoresClientAdapter) List(ctx context.Context, backup Ref, opts ...CallOption) (*List[*StorageRestore], error) {
	projectID, backupID, err := backupIDsFromRef(backup)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.List(ctx, projectID, backupID, rp)
	if err != nil {
		return nil, err
	}
	if resp != nil && !resp.IsSuccess() {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	var items []*StorageRestore
	if resp != nil && resp.Data != nil {
		items = make([]*StorageRestore, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			v := &StorageRestore{}
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
			if v.projectID == "" {
				v.projectID = projectID
			}
			if v.backupID == "" {
				v.backupID = backupID
			}
			items = append(items, v)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*StorageRestore], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*StorageRestore], error) {
		fetch := listPageFetch[types.StorageRestoreListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*StorageRestore
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*StorageRestore, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				item := &StorageRestore{}
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
				if item.projectID == "" {
					item.projectID = projectID
				}
				if item.backupID == "" {
					item.backupID = backupID
				}
				pageItems = append(pageItems, item)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}

// restoreIDsFromRef extracts (projectID, backupID, restoreID) from a Ref.
func restoreIDsFromRef(ref Ref) (projectID, backupID, restoreID string, err error) {
	rid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withRestoreID); ok {
			return w.RestoreID(), true
		}
		return "", false
	}, "restores")
	if !ok || rid == "" {
		return "", "", "", fmt.Errorf("cannot determine StorageRestore ID from Ref %q", ref.URI())
	}
	bid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withBackupID); ok {
			return w.BackupID(), true
		}
		return "", false
	}, "backups")
	if !ok || bid == "" {
		return "", "", "", fmt.Errorf("cannot determine backup ID from Ref %q", ref.URI())
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
	return pid, bid, rid, nil
}
