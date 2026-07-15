package aruba

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/internal/clients/database"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// DBaaSBackup is the wrapper for an Aruba Cloud DBaaS Backup (a direct child
// of a Project; the source DBaaS and Database are body-refs, not path-parents).
// Construct with aruba.NewDBaaSBackup() and bind it via InProject(project),
// FromDBaaS(d), and FromDatabase(db).
//
// Family A: regional, Metadata/Properties envelope, location-aware. No Update
// operation — the underlying API exposes only Create / Get / List / Delete.
//
// Path: /projects/{projectID}/providers/Aruba.Database/backups[/{backupID}]
type DBaaSBackup struct {
	errMixin
	metadataMixin
	regionalMixin
	projectScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	billingPeriod *BillingPeriod
	zone          *Zone   // explicit zone; when nil, Zone is derived from Region
	dbaasRef      *string // body URI
	databaseRef   *string // body URI

	response *types.BackupResponse
}

// NewDBaaSBackup returns a fresh *DBaaSBackup ready for fluent setters and a Create call.
// Binds projectScopedMixin's error sink so IntoProject failures surface via Err().
func NewDBaaSBackup() *DBaaSBackup {
	b := &DBaaSBackup{}
	b.projectScopedMixin = bindProjectScoped(&b.errMixin)
	return b
}

// Setters — chainable, general → specific

// InProject binds this DBaaSBackup to its parent project. Required before Create.
func (b *DBaaSBackup) InProject(p Ref) *DBaaSBackup { b.intoProject(p); return b }

// Named sets the resource name. Required by the API.
func (b *DBaaSBackup) Named(n string) *DBaaSBackup { b.named(n); return b }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (b *DBaaSBackup) Tagged(ts ...string) *DBaaSBackup {
	for _, t := range ts {
		b.addTag(t)
	}
	return b
}

// Untagged removes each listed tag. No-op for tags not present.
func (b *DBaaSBackup) Untagged(ts ...string) *DBaaSBackup {
	for _, t := range ts {
		b.removeTag(t)
	}
	return b
}

// RetaggedAs replaces the entire tag set with the given values.
func (b *DBaaSBackup) RetaggedAs(ts ...string) *DBaaSBackup { b.replaceTags(ts...); return b }

// InRegion sets the region for this resource.
func (b *DBaaSBackup) InRegion(region Region) *DBaaSBackup { b.inRegion(region); return b }

// InZone sets an explicit availability zone. When set it overrides the default
// behaviour of deriving the zone value from InRegion.
func (b *DBaaSBackup) InZone(z Zone) *DBaaSBackup { b.zone = &z; return b }

// FromDBaaS binds the source DBaaS via its URI. Empty URIs are recorded on the
// error sink and the field remains unset.
func (b *DBaaSBackup) FromDBaaS(d Ref) *DBaaSBackup {
	uri := d.URI()
	if uri == "" {
		b.addErr(fmt.Errorf("FromDBaaS: empty URI"))
		return b
	}
	b.dbaasRef = &uri
	return b
}

// FromDatabase binds the source Database via its URI. Empty URIs are recorded
// on the error sink and the field remains unset.
func (b *DBaaSBackup) FromDatabase(db Ref) *DBaaSBackup {
	uri := db.URI()
	if uri == "" {
		b.addErr(fmt.Errorf("FromDatabase: empty URI"))
		return b
	}
	b.databaseRef = &uri
	return b
}

// BilledBy sets the billing cadence. Accepted periods are resource-specific; check the API reference.
func (b *DBaaSBackup) BilledBy(period BillingPeriod) *DBaaSBackup {
	b.billingPeriod = &period
	return b
}

// Getters — general → specific

// URI satisfies Ref.
func (b *DBaaSBackup) URI() string { return b.RespURI() }

// DBaaSBackupID satisfies withDBaaSBackupID.
func (b *DBaaSBackup) DBaaSBackupID() string { return b.ID() }

// Raw shadows responseMetadataMixin.Raw() with the typed response.
func (b *DBaaSBackup) Raw() *types.BackupResponse { return b.response }
func (b *DBaaSBackup) RawJSON() []byte            { return marshalRawJSON(b.response) }
func (b *DBaaSBackup) RawYAML() []byte            { return marshalRawYAML(b.response) }

// RawRequest returns what toRequest() would emit right now.
func (b *DBaaSBackup) RawRequest() types.BackupRequest { return b.toRequest() }

// BillingPeriod returns the billing period set on this backup, or "" if unset.
func (b *DBaaSBackup) BillingPeriod() BillingPeriod {
	if b.billingPeriod == nil {
		return ""
	}
	return *b.billingPeriod
}

// DBaaSURI returns the source DBaaS URI bound via FromDBaaS, or "" if unset.
func (b *DBaaSBackup) DBaaSURI() string { return dbaasBackupDerefString(b.dbaasRef) }

// DatabaseURI returns the raw value stored in the database reference field.
// When set via FromDatabase it holds the full URI; after fromResponse hydration
// it holds the bare database name returned by the API. Prefer DatabaseName()
// when you only need the database name for display or comparison.
func (b *DBaaSBackup) DatabaseURI() string { return dbaasBackupDerefString(b.databaseRef) }

// DatabaseName returns the database name. If the stored reference is a full URI
// (set via FromDatabase), the name is extracted as the path segment after
// "/databases/". If it is already a bare name (set from a response), it is
// returned as-is.
func (b *DBaaSBackup) DatabaseName() string { return b.databaseName() }

// databaseName is the shared implementation used by both DatabaseName() and toRequest().
func (b *DBaaSBackup) databaseName() string {
	if b.databaseRef == nil {
		return ""
	}
	ref := *b.databaseRef
	if idx := strings.LastIndex(ref, "/databases/"); idx >= 0 {
		// Take only the first path segment after "/databases/" so that trailing
		// slashes or extra segments (e.g. "/databases/mydb/") do not leak through.
		seg := ref[idx+len("/databases/"):]
		if end := strings.IndexByte(seg, '/'); end >= 0 {
			seg = seg[:end]
		}
		return seg
	}
	return ref
}

// SizeGB returns the backup storage size in GB from the response, or 0 before hydration.
func (b *DBaaSBackup) SizeGB() int {
	if b.response == nil {
		return 0
	}
	return int(b.response.Properties.Storage.Size)
}

// Zone returns the availability zone from the response, or "" before hydration.
func (b *DBaaSBackup) Zone() Zone {
	if b.response == nil {
		return ""
	}
	return b.response.Properties.Zone
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (b *DBaaSBackup) toRequest() types.BackupRequest {
	zone := Zone(b.Region()) // default: derive zone from region
	if b.zone != nil {
		zone = *b.zone // explicit InZone overrides the derived value
	}
	props := types.BackupPropertiesRequest{
		Zone: zone,
	}
	if b.dbaasRef != nil {
		props.DBaaS = types.ReferenceResourceCommon{URI: *b.dbaasRef}
	}
	if b.databaseRef != nil {
		props.Database = types.DatabaseNameRef{Name: b.databaseName()}
	}
	props.BillingPlanCommon = &types.BillingPlanCommon{BillingPeriod: defaultBillingPeriod(b.billingPeriod)}
	return types.BackupRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: b.toMetadata(),
			Location:                b.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (b *DBaaSBackup) fromResponse(resp *types.BackupResponse) {
	if resp == nil {
		return
	}
	b.response = resp
	b.setMeta(&resp.Metadata)
	b.named(dbaasBackupDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		b.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		b.inRegion(resp.Metadata.LocationResponse.Value)
	}
	b.setStatus(&resp.Status)

	if resp.Properties.DBaaS.URI != "" {
		v := resp.Properties.DBaaS.URI
		b.dbaasRef = &v
	}
	if resp.Properties.Database.Name != "" {
		v := resp.Properties.Database.Name
		b.databaseRef = &v
	}
	if resp.Properties.BillingPlanCommon != nil && resp.Properties.BillingPlanCommon.BillingPeriod != nil {
		b.billingPeriod = resp.Properties.BillingPlanCommon.BillingPeriod
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

func dbaasBackupDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// dbaasBackupIDsFromRef extracts (projectID, backupID) from a Ref. Uses the
// dedicated withDBaaSBackupID interface for typed extraction so a typed
// *StorageBackup Ref does not silently route to the DBaaS endpoint. URI
// fallback (segment "backups") remains inherently ambiguous between the two
// backup scopes — callers must pass URIs from the correct domain.
func dbaasBackupIDsFromRef(ref Ref) (projectID, backupID string, err error) {
	bid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withDBaaSBackupID); ok {
			return w.DBaaSBackupID(), true
		}
		return "", false
	}, "backups")
	if !ok || bid == "" {
		return "", "", fmt.Errorf("cannot determine backup ID from Ref %q", ref.URI())
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

// dbaasBackupsLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type dbaasBackupsLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.DBaaSBackupListResponse], error)
	Get(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[types.BackupResponse], error)
	Create(ctx context.Context, projectID string, body types.BackupRequest, params *types.RequestParameters) (*types.Response[types.BackupResponse], error)
	Delete(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// dbaasBackupsClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates DBaaSBackup ↔ types.BackupRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type dbaasBackupsClientAdapter struct {
	low  dbaasBackupsLowLevelClient
	rest *restclient.Client
}

var _ BackupsClient = (*dbaasBackupsClientAdapter)(nil)

func newDBaaSBackupsClientAdapter(rest *restclient.Client) *dbaasBackupsClientAdapter {
	if rest == nil {
		return &dbaasBackupsClientAdapter{}
	}
	return &dbaasBackupsClientAdapter{low: database.NewBackupsClientImpl(rest), rest: rest}
}

// Create posts a new DBaaSBackup to the API and hydrates the wrapper from the response.
func (a *dbaasBackupsClientAdapter) Create(ctx context.Context, b *DBaaSBackup, opts ...CallOption) (*DBaaSBackup, error) {
	if err := b.Err(); err != nil {
		return b, err
	}
	if b.ProjectID() == "" {
		return b, fmt.Errorf("Create: DBaaSBackup has no parent project — call InProject first")
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

// Get fetches a DBaaSBackup by Ref and returns a freshly hydrated wrapper.
func (a *dbaasBackupsClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*DBaaSBackup, error) {
	projectID, backupID, err := dbaasBackupIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, backupID, rp)
	out := &DBaaSBackup{}
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

// Delete removes the DBaaSBackup identified by Ref.
func (a *dbaasBackupsClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, backupID, err := dbaasBackupIDsFromRef(ref)
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

// List returns a paginated list of DBaaSBackup entries in the given parent scope.
func (a *dbaasBackupsClientAdapter) List(ctx context.Context, parent Ref, opts ...CallOption) (*List[*DBaaSBackup], error) {
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
	var items []*DBaaSBackup
	if resp != nil && resp.Data != nil {
		items = make([]*DBaaSBackup, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			b := &DBaaSBackup{}
			b.projectID = projectID
			b.fromResponse(&resp.Data.Values[i])
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
			if b.projectID == "" {
				b.projectID = projectID
			}
			items = append(items, b)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*DBaaSBackup], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*DBaaSBackup], error) {
		fetch := listPageFetch[types.DBaaSBackupListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*DBaaSBackup
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*DBaaSBackup, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				b := &DBaaSBackup{}
				b.projectID = projectID
				b.fromResponse(&pageResp.Data.Values[i])
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
				if b.projectID == "" {
					b.projectID = projectID
				}
				pageItems = append(pageItems, b)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}
