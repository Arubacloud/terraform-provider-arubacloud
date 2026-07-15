package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/storage"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// StorageBackup is the wrapper for an Aruba Cloud Storage Backup (a direct
// child of a Project, derived from a BlockStorage volume). Construct with
// aruba.NewStorageBackup() and bind it via IntoProject(project) and
// FromVolume(bs).
type StorageBackup struct {
	errMixin
	metadataMixin
	regionalMixin
	projectScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	backupType    *types.StorageBackupType
	originRef     *string // body URI
	retentionDays *int
	billingPeriod *BillingPeriod

	response *types.StorageBackupResponse
}

// NewStorageBackup returns a fresh *StorageBackup ready for fluent setters and a Create call.
// Binds projectScopedMixin's error sink so IntoProject failures surface via Err().
func NewStorageBackup() *StorageBackup {
	b := &StorageBackup{}
	b.projectScopedMixin = bindProjectScoped(&b.errMixin)
	return b
}

// Setters — chainable, general → specific

// InProject binds this StorageBackup to its parent project. Required before Create.
func (b *StorageBackup) InProject(p Ref) *StorageBackup { b.intoProject(p); return b }

// Named sets the resource name. Required by the API.
func (b *StorageBackup) Named(n string) *StorageBackup { b.named(n); return b }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (b *StorageBackup) Tagged(ts ...string) *StorageBackup {
	for _, t := range ts {
		b.addTag(t)
	}
	return b
}

// Untagged removes each listed tag. No-op for tags not present.
func (b *StorageBackup) Untagged(ts ...string) *StorageBackup {
	for _, t := range ts {
		b.removeTag(t)
	}
	return b
}

// RetaggedAs replaces the entire tag set with the given values.
func (b *StorageBackup) RetaggedAs(ts ...string) *StorageBackup { b.replaceTags(ts...); return b }

// InRegion sets the region for this resource.
func (b *StorageBackup) InRegion(region Region) *StorageBackup { b.inRegion(region); return b }

// OfType sets the backup type (Full or Incremental).
func (b *StorageBackup) OfType(t types.StorageBackupType) *StorageBackup {
	v := t
	b.backupType = &v
	return b
}

// RetainedForDays sets the number of days the backup is retained.
func (b *StorageBackup) RetainedForDays(days int) *StorageBackup {
	v := days
	b.retentionDays = &v
	return b
}

// BilledBy sets the billing cadence. Accepted periods are resource-specific; check the API reference.
func (b *StorageBackup) BilledBy(period BillingPeriod) *StorageBackup {
	b.billingPeriod = &period
	return b
}

// FromVolume binds the source BlockStorage via its URI. Pass any Ref (typed
// or aruba.URI(...)). Empty URIs are recorded on the error sink and the field
// remains unset.
func (b *StorageBackup) FromVolume(vol Ref) *StorageBackup {
	uri := vol.URI()
	if uri == "" {
		b.addErr(fmt.Errorf("FromVolume: empty URI"))
		return b
	}
	b.originRef = &uri
	return b
}

// Getters — general → specific

// URI satisfies Ref.
func (b *StorageBackup) URI() string { return b.RespURI() }

// BackupID satisfies withBackupID.
func (b *StorageBackup) BackupID() string { return b.ID() }

// Raw shadows responseMetadataMixin.Raw() with the typed response.
func (b *StorageBackup) Raw() *types.StorageBackupResponse { return b.response }
func (b *StorageBackup) RawJSON() []byte                   { return marshalRawJSON(b.response) }
func (b *StorageBackup) RawYAML() []byte                   { return marshalRawYAML(b.response) }

// RawRequest returns what toRequest() would emit right now.
func (b *StorageBackup) RawRequest() types.StorageBackupRequest { return b.toRequest() }

// Type returns the backup type (Full or Incremental) set via OfType, or "" if unset.
func (b *StorageBackup) Type() types.StorageBackupType {
	if b.backupType == nil {
		return ""
	}
	return *b.backupType
}

// OriginURI returns the source volume URI bound via FromVolume, or "" if unset.
func (b *StorageBackup) OriginURI() string { return storageBackupDerefString(b.originRef) }

// BillingPeriod returns the billing period set on this backup, or "" if unset.
func (b *StorageBackup) BillingPeriod() BillingPeriod {
	if b.billingPeriod == nil {
		return ""
	}
	return *b.billingPeriod
}

// RetentionDays returns the retention period in days, or 0 if unset.
func (b *StorageBackup) RetentionDays() int {
	if b.retentionDays == nil {
		return 0
	}
	return *b.retentionDays
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (b *StorageBackup) toRequest() types.StorageBackupRequest {
	props := types.StorageBackupPropertiesRequest{
		RetentionDays: b.retentionDays,
		BillingPeriod: storageBackupBillingPeriodWire().Out(defaultBillingPeriod(b.billingPeriod)),
	}
	if b.backupType != nil {
		props.StorageBackupType = *b.backupType
	}
	if b.originRef != nil {
		props.Origin = types.ReferenceResourceCommon{URI: *b.originRef}
	}
	return types.StorageBackupRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: b.toMetadata(),
			Location:                b.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (b *StorageBackup) fromResponse(resp *types.StorageBackupResponse) {
	if resp == nil {
		return
	}
	b.response = resp
	b.setMeta(&resp.Metadata)
	b.named(storageBackupDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		b.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		b.inRegion(resp.Metadata.LocationResponse.Value)
	}
	b.setStatus(&resp.Status)

	if resp.Properties.Type != "" {
		v := resp.Properties.Type
		b.backupType = &v
	}
	if resp.Properties.Origin.URI != "" {
		v := resp.Properties.Origin.URI
		b.originRef = &v
	}
	if resp.Properties.RetentionDays != nil {
		v := *resp.Properties.RetentionDays
		b.retentionDays = &v
	}
	if resp.Properties.BillingPeriod != nil && *resp.Properties.BillingPeriod != "" {
		b.billingPeriod = storageBackupBillingPeriodWire().In(resp.Properties.BillingPeriod)
	}

	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		b.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if b.projectID == "" && b.RespURI() != "" {
		if pid := parseURIIDs(b.RespURI())["projects"]; pid != "" {
			b.projectID = pid
		}
	}
}

func storageBackupDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// storageBackupBillingPeriodWire returns a translator for StorageBackup's
// TitleCase wire billing-period values (e.g. "Monthly") vs. the standard SDK constants.
func storageBackupBillingPeriodWire() *billingPeriodTranslator {
	return newBillingPeriodTranslator(map[BillingPeriod]string{
		BillingPeriodHour:  "Hourly",
		BillingPeriodMonth: "Monthly",
		BillingPeriodYear:  "Yearly",
	})
}

// ---- Low-level client interface ----

// storageBackupLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type storageBackupLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.StorageBackupListResponse], error)
	Get(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error)
	Create(ctx context.Context, projectID string, body types.StorageBackupRequest, params *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error)
	Update(ctx context.Context, projectID, backupID string, body types.StorageBackupRequest, params *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error)
	Delete(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// storageBackupsClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates StorageBackup ↔ types.StorageBackupRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type storageBackupsClientAdapter struct {
	low  storageBackupLowLevelClient
	rest *restclient.Client
}

var _ StorageBackupsClient = (*storageBackupsClientAdapter)(nil)

func newStorageBackupsClientAdapter(rest *restclient.Client) *storageBackupsClientAdapter {
	if rest == nil {
		return &storageBackupsClientAdapter{}
	}
	return &storageBackupsClientAdapter{low: storage.NewBackupClientImpl(rest), rest: rest}
}

// Create posts a new StorageBackup to the API and hydrates the wrapper from the response.
func (a *storageBackupsClientAdapter) Create(ctx context.Context, b *StorageBackup, opts ...CallOption) (*StorageBackup, error) {
	if err := b.Err(); err != nil {
		return b, err
	}
	if b.ProjectID() == "" {
		return b, fmt.Errorf("Create: StorageBackup has no project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, b.ProjectID(), b.toRequest(), rp)
	populateHTTPEnvelope(&b.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		b.fromResponse(resp.Data)
		b.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, b)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				b.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return b, err
	}
	if resp != nil && !resp.IsSuccess() {
		return b, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return b, nil
}

// Get fetches a StorageBackup by Ref and returns a freshly hydrated wrapper.
func (a *storageBackupsClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*StorageBackup, error) {
	projectID, backupID, err := backupIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, backupID, rp)
	out := &StorageBackup{}
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

// Update sends a PUT for the current wrapper state. Requires ID and parent.
func (a *storageBackupsClientAdapter) Update(ctx context.Context, b *StorageBackup, opts ...CallOption) (*StorageBackup, error) {
	if err := b.Err(); err != nil {
		return b, err
	}
	if b.ID() == "" {
		return b, fmt.Errorf("Update: StorageBackup has no ID — call Get first or seed from response metadata")
	}
	if b.ProjectID() == "" {
		return b, fmt.Errorf("Update: StorageBackup has no project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, b.ProjectID(), b.ID(), b.toRequest(), rp)
	populateHTTPEnvelope(&b.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		b.fromResponse(resp.Data)
		b.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, b)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				b.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return b, err
	}
	if resp != nil && !resp.IsSuccess() {
		return b, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return b, nil
}

// Delete removes the StorageBackup identified by Ref.
func (a *storageBackupsClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, backupID, err := backupIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, backupID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of StorageBackup entries in the given parent scope.
func (a *storageBackupsClientAdapter) List(ctx context.Context, project Ref, opts ...CallOption) (*List[*StorageBackup], error) {
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
	var items []*StorageBackup
	if resp != nil && resp.Data != nil {
		items = make([]*StorageBackup, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			bkp := &StorageBackup{}
			bkp.fromResponse(&resp.Data.Values[i])
			bkp.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, bkp)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					bkp.fromResponse(fresh.Raw())
				}
				return nil
			})
			if bkp.projectID == "" {
				bkp.projectID = projectID
			}
			items = append(items, bkp)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*StorageBackup], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*StorageBackup], error) {
		fetch := listPageFetch[types.StorageBackupListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*StorageBackup
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*StorageBackup, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				item := &StorageBackup{}
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
				pageItems = append(pageItems, item)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}

// backupIDsFromRef extracts (projectID, backupID) from a Ref.
func backupIDsFromRef(ref Ref) (projectID, backupID string, err error) {
	bid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withBackupID); ok {
			return w.BackupID(), true
		}
		return "", false
	}, "backups")
	if !ok || bid == "" {
		return "", "", fmt.Errorf("cannot determine StorageBackup ID from Ref %q", ref.URI())
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
	return pid, bid, nil
}
