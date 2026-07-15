package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/storage"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// Snapshot is the wrapper for an Aruba Cloud Snapshot (a direct child of a
// Project, derived from a BlockStorage volume). Construct with
// aruba.NewSnapshot() and bind it via InProject(project) and FromVolume(bs).
type Snapshot struct {
	errMixin
	metadataMixin
	regionalMixin
	projectScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	billingPeriod *BillingPeriod
	volumeRef     *string // body URI

	// Read-only fields hydrated from response.
	sizeGB      *int32
	storageType *types.BlockStorageType
	zone        *Zone
	bootable    *bool

	response *types.SnapshotResponse
}

// NewSnapshot returns a fresh *Snapshot ready for fluent setters and a Create call.
// Binds projectScopedMixin's error sink so IntoProject failures surface via Err().
func NewSnapshot() *Snapshot {
	s := &Snapshot{}
	s.projectScopedMixin = bindProjectScoped(&s.errMixin)
	return s
}

// Setters — chainable, general → specific

// InProject binds this Snapshot to its parent project. Required before Create.
func (s *Snapshot) InProject(p Ref) *Snapshot { s.intoProject(p); return s }

// Named sets the resource name. Required by the API.
func (s *Snapshot) Named(n string) *Snapshot { s.named(n); return s }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (s *Snapshot) Tagged(ts ...string) *Snapshot {
	for _, t := range ts {
		s.addTag(t)
	}
	return s
}

// Untagged removes each listed tag. No-op for tags not present.
func (s *Snapshot) Untagged(ts ...string) *Snapshot {
	for _, t := range ts {
		s.removeTag(t)
	}
	return s
}

// RetaggedAs replaces the entire tag set with the given values.
func (s *Snapshot) RetaggedAs(ts ...string) *Snapshot { s.replaceTags(ts...); return s }

// InRegion sets the region for this resource.
func (s *Snapshot) InRegion(region Region) *Snapshot { s.inRegion(region); return s }

// BilledBy sets the billing cadence. Accepted periods are resource-specific; check the API reference.
func (s *Snapshot) BilledBy(period BillingPeriod) *Snapshot { s.billingPeriod = &period; return s }

// FromVolume binds the source BlockStorage via its URI. Pass any Ref (typed or
// aruba.URI(...)). Empty URIs are recorded on the error sink and the field
// remains unset.
func (s *Snapshot) FromVolume(vol Ref) *Snapshot {
	uri := vol.URI()
	if uri == "" {
		s.addErr(fmt.Errorf("FromVolume: empty URI"))
		return s
	}
	s.volumeRef = &uri
	return s
}

// Getters — general → specific

// URI satisfies Ref.
func (s *Snapshot) URI() string { return s.RespURI() }

// SnapshotID satisfies withSnapshotID.
func (s *Snapshot) SnapshotID() string { return s.ID() }

// Raw shadows responseMetadataMixin.Raw() with the typed Snapshot response.
func (s *Snapshot) Raw() *types.SnapshotResponse { return s.response }
func (s *Snapshot) RawJSON() []byte              { return marshalRawJSON(s.response) }
func (s *Snapshot) RawYAML() []byte              { return marshalRawYAML(s.response) }

// RawRequest returns what toRequest() would emit right now.
func (s *Snapshot) RawRequest() types.SnapshotRequest { return s.toRequest() }

// BillingPeriod returns the billing period, or "" if unset.
func (s *Snapshot) BillingPeriod() BillingPeriod {
	if s.billingPeriod == nil {
		return ""
	}
	return *s.billingPeriod
}

// VolumeURI returns the source volume URI, or "" if unset.
func (s *Snapshot) VolumeURI() string { return snapshotDerefString(s.volumeRef) }

// Read-only accessors hydrated from response.

// SizeGB returns the snapshot size in GiB, or 0 if unset.
func (s *Snapshot) SizeGB() int {
	if s.sizeGB == nil {
		return 0
	}
	return int(*s.sizeGB)
}

// Type returns the storage type of the source volume, or "" if unset.
func (s *Snapshot) Type() types.BlockStorageType {
	if s.storageType == nil {
		return ""
	}
	return *s.storageType
}

// Zone returns the availability zone of the snapshot, or "" if unset.
func (s *Snapshot) Zone() Zone { return snapshotDerefZone(s.zone) }

// IsBootable returns whether the snapshot was taken from a bootable volume, or false if unset.
func (s *Snapshot) IsBootable() bool {
	if s.bootable == nil {
		return false
	}
	return *s.bootable
}

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (s *Snapshot) toRequest() types.SnapshotRequest {
	props := types.SnapshotPropertiesRequest{
		BillingPeriod: defaultBillingPeriod(s.billingPeriod),
	}
	if s.volumeRef != nil {
		props.Volume = types.ReferenceResourceCommon{URI: *s.volumeRef}
	}
	return types.SnapshotRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: s.toMetadata(),
			Location:                s.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (s *Snapshot) fromResponse(resp *types.SnapshotResponse) {
	if resp == nil {
		return
	}
	s.response = resp
	s.setMeta(&resp.Metadata)
	s.named(snapshotDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		s.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		s.inRegion(resp.Metadata.LocationResponse.Value)
	}
	s.setStatus(&resp.Status)

	if resp.Properties.SizeGB != nil {
		v := *resp.Properties.SizeGB
		s.sizeGB = &v
	}
	if resp.Properties.BillingPeriod != nil && *resp.Properties.BillingPeriod != "" {
		v := *resp.Properties.BillingPeriod
		s.billingPeriod = &v
	}
	if resp.Properties.Type != "" {
		v := resp.Properties.Type
		s.storageType = &v
	}
	if resp.Properties.Zone != "" {
		v := resp.Properties.Zone
		s.zone = &v
	}
	if resp.Properties.Bootable != nil {
		v := *resp.Properties.Bootable
		s.bootable = &v
	}
	if resp.Properties.Volume != nil && resp.Properties.Volume.URI != nil && *resp.Properties.Volume.URI != "" {
		v := *resp.Properties.Volume.URI
		s.volumeRef = &v
	}

	if resp.Metadata.ProjectMetadataResponse != nil && resp.Metadata.ProjectMetadataResponse.ID != "" {
		s.projectID = resp.Metadata.ProjectMetadataResponse.ID
	}
	if s.projectID == "" && s.RespURI() != "" {
		if pid := parseURIIDs(s.RespURI())["projects"]; pid != "" {
			s.projectID = pid
		}
	}
}

func snapshotDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func snapshotDerefZone(p *Zone) Zone {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// snapshotLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type snapshotLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.SnapshotListResponse], error)
	Get(ctx context.Context, projectID, snapshotID string, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error)
	Create(ctx context.Context, projectID string, body types.SnapshotRequest, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error)
	Update(ctx context.Context, projectID, snapshotID string, body types.SnapshotRequest, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error)
	Delete(ctx context.Context, projectID, snapshotID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// snapshotsClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates Snapshot ↔ types.SnapshotRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type snapshotsClientAdapter struct {
	low  snapshotLowLevelClient
	rest *restclient.Client
}

var _ SnapshotsClient = (*snapshotsClientAdapter)(nil)

func newSnapshotsClientAdapter(rest *restclient.Client) *snapshotsClientAdapter {
	if rest == nil {
		return &snapshotsClientAdapter{}
	}
	return &snapshotsClientAdapter{
		low:  storage.NewSnapshotsClientImpl(rest, storage.NewVolumesClientImpl(rest)),
		rest: rest,
	}
}

// Create posts a new Snapshot to the API and hydrates the wrapper from the response.
func (a *snapshotsClientAdapter) Create(ctx context.Context, snap *Snapshot, opts ...CallOption) (*Snapshot, error) {
	if err := snap.Err(); err != nil {
		return snap, err
	}
	if snap.ProjectID() == "" {
		return snap, fmt.Errorf("Create: Snapshot has no project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, snap.ProjectID(), snap.toRequest(), rp)
	populateHTTPEnvelope(&snap.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		snap.fromResponse(resp.Data)
		snap.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, snap)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				snap.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return snap, err
	}
	if resp != nil && !resp.IsSuccess() {
		return snap, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return snap, nil
}

// Get fetches a Snapshot by Ref and returns a freshly hydrated wrapper.
func (a *snapshotsClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*Snapshot, error) {
	projectID, snapshotID, err := snapshotIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, snapshotID, rp)
	out := &Snapshot{}
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
func (a *snapshotsClientAdapter) Update(ctx context.Context, snap *Snapshot, opts ...CallOption) (*Snapshot, error) {
	if err := snap.Err(); err != nil {
		return snap, err
	}
	if snap.ID() == "" {
		return snap, fmt.Errorf("Update: Snapshot has no ID — call Get first or seed from response metadata")
	}
	if snap.ProjectID() == "" {
		return snap, fmt.Errorf("Update: Snapshot has no project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, snap.ProjectID(), snap.ID(), snap.toRequest(), rp)
	populateHTTPEnvelope(&snap.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		snap.fromResponse(resp.Data)
		snap.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, snap)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				snap.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return snap, err
	}
	if resp != nil && !resp.IsSuccess() {
		return snap, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return snap, nil
}

// Delete removes the Snapshot identified by Ref.
func (a *snapshotsClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, snapshotID, err := snapshotIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, snapshotID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of Snapshot in the given parent scope.
func (a *snapshotsClientAdapter) List(ctx context.Context, project Ref, opts ...CallOption) (*List[*Snapshot], error) {
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
	var items []*Snapshot
	if resp != nil && resp.Data != nil {
		items = make([]*Snapshot, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			snap := &Snapshot{}
			snap.fromResponse(&resp.Data.Values[i])
			snap.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, snap)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					snap.fromResponse(fresh.Raw())
				}
				return nil
			})
			if snap.projectID == "" {
				snap.projectID = projectID
			}
			items = append(items, snap)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*Snapshot], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*Snapshot], error) {
		fetch := listPageFetch[types.SnapshotListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*Snapshot
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*Snapshot, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				item := &Snapshot{}
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

// snapshotIDsFromRef extracts (projectID, snapshotID) from a Ref.
func snapshotIDsFromRef(ref Ref) (projectID, snapshotID string, err error) {
	sid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withSnapshotID); ok {
			return w.SnapshotID(), true
		}
		return "", false
	}, "snapshots")
	if !ok || sid == "" {
		return "", "", fmt.Errorf("cannot determine Snapshot ID from Ref %q", ref.URI())
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
	return pid, sid, nil
}
