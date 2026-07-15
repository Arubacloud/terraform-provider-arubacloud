package aruba

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/internal/clients/storage"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// BlockStorage is the wrapper for an Aruba Cloud BlockStorage volume
// (a direct child of a Project). Construct with aruba.NewBlockStorage()
// and bind it via InProject(project).
type BlockStorage struct {
	errMixin
	metadataMixin
	zonalMixin
	projectScopedMixin
	responseMetadataMixin
	statusMixin
	linkedMixin
	httpEnvelopeMixin

	sizeGB        *int32
	storageType   *types.BlockStorageType
	billingPeriod *BillingPeriod
	snapshotRef   *string // body URI
	image         *string
	bootable      *bool

	response *types.BlockStorageResponse
}

// NewBlockStorage returns a fresh *BlockStorage ready for fluent setters and a Create call.
// Binds projectScopedMixin's error sink so IntoProject failures surface via Err().
func NewBlockStorage() *BlockStorage {
	b := &BlockStorage{}
	b.projectScopedMixin = bindProjectScoped(&b.errMixin)
	return b
}

// Setters — chainable, general → specific

// InProject binds this BlockStorage to its parent project. Required before Create.
func (b *BlockStorage) InProject(p Ref) *BlockStorage { b.intoProject(p); return b }

// Named sets the resource name. Required by the API.
func (b *BlockStorage) Named(n string) *BlockStorage { b.named(n); return b }

// Tagged appends tags for filtering and accounting. Repeated calls append.
func (b *BlockStorage) Tagged(ts ...string) *BlockStorage {
	for _, t := range ts {
		b.addTag(t)
	}
	return b
}

// Untagged removes each listed tag. No-op for tags not present.
func (b *BlockStorage) Untagged(ts ...string) *BlockStorage {
	for _, t := range ts {
		b.removeTag(t)
	}
	return b
}

// RetaggedAs replaces the entire tag set with the given values.
func (b *BlockStorage) RetaggedAs(ts ...string) *BlockStorage { b.replaceTags(ts...); return b }

// InRegion sets the region for this resource.
func (b *BlockStorage) InRegion(region Region) *BlockStorage { b.inRegion(region); return b }

// InZone sets the availability zone. More specific than InRegion.
func (b *BlockStorage) InZone(zone Zone) *BlockStorage { b.inZone(zone); return b }

// SizedGB sets the volume size in GiB.
func (b *BlockStorage) SizedGB(gb int) *BlockStorage { v := int32(gb); b.sizeGB = &v; return b }

// OfType sets the storage type (e.g. SSD, HDD).
func (b *BlockStorage) OfType(t types.BlockStorageType) *BlockStorage {
	b.storageType = &t
	return b
}

// BilledBy sets the billing cadence. Accepted periods are resource-specific; check the API reference.
func (b *BlockStorage) BilledBy(period BillingPeriod) *BlockStorage {
	b.billingPeriod = &period
	return b
}

// FromImage sets the source image name for bootable volumes.
func (b *BlockStorage) FromImage(img string) *BlockStorage { b.image = &img; return b }

// AsBootable marks the volume as bootable.
func (b *BlockStorage) AsBootable() *BlockStorage { v := true; b.bootable = &v; return b }

// NotBootable explicitly marks the volume as non-bootable.
func (b *BlockStorage) NotBootable() *BlockStorage { v := false; b.bootable = &v; return b }

// FromSnapshot binds the source snapshot via its URI. Pass any Ref (typed or
// aruba.URI(...)). Empty URIs are recorded on the error sink and the field
// remains unset.
func (b *BlockStorage) FromSnapshot(snap Ref) *BlockStorage {
	uri := snap.URI()
	if uri == "" {
		b.addErr(fmt.Errorf("FromSnapshot: empty URI"))
		return b
	}
	b.snapshotRef = &uri
	return b
}

// Getters — general → specific

// URI satisfies Ref.
func (b *BlockStorage) URI() string { return b.RespURI() }

// BlockStorageID satisfies withBlockStorageID.
func (b *BlockStorage) BlockStorageID() string { return b.ID() }

// Raw shadows responseMetadataMixin.Raw() with the typed BlockStorage response.
func (b *BlockStorage) Raw() *types.BlockStorageResponse { return b.response }
func (b *BlockStorage) RawJSON() []byte                  { return marshalRawJSON(b.response) }
func (b *BlockStorage) RawYAML() []byte                  { return marshalRawYAML(b.response) }

// RawRequest returns what toCreateRequest() would emit right now.
func (b *BlockStorage) RawRequest() types.BlockStorageRequest { return b.toCreateRequest() }

// SizeGB returns the volume size in GiB, or 0 if unset.
func (b *BlockStorage) SizeGB() int {
	if b.sizeGB == nil {
		return 0
	}
	return int(*b.sizeGB)
}

// Type returns the storage type (e.g. SSD, HDD), or "" if unset.
func (b *BlockStorage) Type() types.BlockStorageType {
	if b.storageType == nil {
		return ""
	}
	return *b.storageType
}

// BillingPeriod returns the billing period, or "" if unset.
func (b *BlockStorage) BillingPeriod() BillingPeriod {
	if b.billingPeriod == nil {
		return ""
	}
	return *b.billingPeriod
}

// Image returns the source image name, or "" if unset.
func (b *BlockStorage) Image() string { return blockStorageDerefString(b.image) }

// IsBootable returns whether the volume is bootable, or false if unset.
func (b *BlockStorage) IsBootable() bool {
	if b.bootable == nil {
		return false
	}
	return *b.bootable
}

// SnapshotURI returns the source snapshot URI, or "" if unset.
func (b *BlockStorage) SnapshotURI() string { return blockStorageDerefString(b.snapshotRef) }

// Wire converters

// toCreateRequest assembles the Create body. bootable defaults to false when unset (API wire contract).
func (b *BlockStorage) toCreateRequest() types.BlockStorageRequest {
	var t types.BlockStorageType
	if b.storageType != nil {
		t = *b.storageType
	}
	var sizeGB int
	if b.sizeGB != nil {
		sizeGB = int(*b.sizeGB)
	}
	props := types.BlockStoragePropertiesRequest{
		SizeGB:        sizeGB,
		BillingPeriod: defaultBillingPeriod(b.billingPeriod),
		Zone:          b.zonePtr(),
		Type:          t,
		Bootable:      blockStorageBootable(b.bootable),
		Image:         b.image,
	}
	if b.snapshotRef != nil {
		props.Snapshot = &types.ReferenceResourceCommon{URI: *b.snapshotRef}
	}
	return types.BlockStorageRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: b.toMetadata(),
			Location:                b.toLocation(),
		},
		Properties: props,
	}
}

// toUpdateRequest assembles the Update (PUT) body. bootable is omitted when the caller
// never called SetBootable/UnsetBootable, preventing silent clobbering of server-side state.
func (b *BlockStorage) toUpdateRequest() types.BlockStorageRequest {
	var t types.BlockStorageType
	if b.storageType != nil {
		t = *b.storageType
	}
	var sizeGB int
	if b.sizeGB != nil {
		sizeGB = int(*b.sizeGB)
	}
	props := types.BlockStoragePropertiesRequest{
		SizeGB:        sizeGB,
		BillingPeriod: defaultBillingPeriod(b.billingPeriod),
		Zone:          b.zonePtr(),
		Type:          t,
		Bootable:      b.bootable, // nil → omitted by omitempty; only sent when explicitly set
		Image:         b.image,
	}
	if b.snapshotRef != nil {
		props.Snapshot = &types.ReferenceResourceCommon{URI: *b.snapshotRef}
	}
	return types.BlockStorageRequest{
		Metadata: types.RegionalResourceMetadataRequest{
			ResourceMetadataRequest: b.toMetadata(),
			Location:                b.toLocation(),
		},
		Properties: props,
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (b *BlockStorage) fromResponse(resp *types.BlockStorageResponse) {
	if resp == nil {
		return
	}
	b.response = resp
	b.setMeta(&resp.Metadata)
	b.named(blockStorageDerefString(resp.Metadata.Name))
	if len(resp.Metadata.Tags) > 0 {
		b.replaceTags(resp.Metadata.Tags...)
	}
	if resp.Metadata.LocationResponse != nil {
		b.inRegion(resp.Metadata.LocationResponse.Value)
	}
	b.setStatus(&resp.Status)

	b.setLinked(resp.Properties.LinkedResources)

	if resp.Properties.SizeGB != 0 {
		v := int32(resp.Properties.SizeGB)
		b.sizeGB = &v
	}
	if resp.Properties.Type != "" {
		v := resp.Properties.Type
		b.storageType = &v
	}
	if resp.Properties.Zone != "" {
		v := resp.Properties.Zone
		b.zone = &v
	}
	if resp.Properties.BillingPeriod != nil {
		b.billingPeriod = resp.Properties.BillingPeriod
	}

	if resp.Properties.Image != nil && *resp.Properties.Image != "" {
		v := *resp.Properties.Image
		b.image = &v
	}
	if resp.Properties.Bootable != nil {
		v := *resp.Properties.Bootable
		b.bootable = &v
	}
	if resp.Properties.Snapshot != nil && resp.Properties.Snapshot.URI != "" {
		v := resp.Properties.Snapshot.URI
		b.snapshotRef = &v
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

// blockStorageBootable returns p when non-nil, or a pointer to false as the
// wire default (the API treats unset the same as false, so we always send the field).
func blockStorageBootable(p *bool) *bool {
	if p != nil {
		return p
	}
	f := false
	return &f
}

func blockStorageDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// WaitUntilNotUsed blocks until the BlockStorage reaches the "NotUsed" state —
// the steady terminal state for an unattached volume. Call this after Create
// and before passing the volume to a CloudServer.
func (b *BlockStorage) WaitUntilNotUsed(ctx context.Context, opts ...WaitOption) error {
	return b.WaitUntilStates(ctx, []types.State{types.StateNotUsed}, opts...)
}

// WaitUntilUsed blocks until the BlockStorage is bound to a consumer resource.
// The platform may emit "InUse", "Used", or "Reserved" (bound as a dependency
// but not actively in use); this method succeeds on whichever arrives first.
func (b *BlockStorage) WaitUntilUsed(ctx context.Context, opts ...WaitOption) error {
	return b.WaitUntilStates(ctx, []types.State{types.StateInUse, types.StateUsed, types.StateReserved}, opts...)
}

// ---- Low-level client interface ----

// volumeLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type volumeLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.BlockStorageListResponse], error)
	Get(ctx context.Context, projectID, volumeID string, params *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error)
	Create(ctx context.Context, projectID string, body types.BlockStorageRequest, params *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error)
	Update(ctx context.Context, projectID, volumeID string, body types.BlockStorageRequest, params *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error)
	Delete(ctx context.Context, projectID, volumeID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// volumesClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates BlockStorage ↔ types.BlockStorageRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type volumesClientAdapter struct {
	low  volumeLowLevelClient
	rest *restclient.Client
}

var _ VolumesClient = (*volumesClientAdapter)(nil)

func newVolumesClientAdapter(rest *restclient.Client) *volumesClientAdapter {
	if rest == nil {
		return &volumesClientAdapter{}
	}
	return &volumesClientAdapter{low: storage.NewVolumesClientImpl(rest), rest: rest}
}

// Create posts a new BlockStorage to the API and hydrates the wrapper from the response.
func (a *volumesClientAdapter) Create(ctx context.Context, vol *BlockStorage, opts ...CallOption) (*BlockStorage, error) {
	if err := vol.Err(); err != nil {
		return vol, err
	}
	if vol.ProjectID() == "" {
		return vol, fmt.Errorf("Create: BlockStorage has no project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, vol.ProjectID(), vol.toCreateRequest(), rp)
	populateHTTPEnvelope(&vol.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		vol.fromResponse(resp.Data)
		vol.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, vol)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				vol.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return vol, err
	}
	if resp != nil && !resp.IsSuccess() {
		return vol, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return vol, nil
}

// Get fetches a BlockStorage by Ref and returns a freshly hydrated wrapper.
func (a *volumesClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*BlockStorage, error) {
	projectID, blockStorageID, err := blockStorageIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, blockStorageID, rp)
	out := &BlockStorage{}
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
func (a *volumesClientAdapter) Update(ctx context.Context, vol *BlockStorage, opts ...CallOption) (*BlockStorage, error) {
	if err := vol.Err(); err != nil {
		return vol, err
	}
	if vol.ID() == "" {
		return vol, fmt.Errorf("Update: BlockStorage has no ID — call Get first or seed from response metadata")
	}
	if vol.ProjectID() == "" {
		return vol, fmt.Errorf("Update: BlockStorage has no project — call InProject first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, vol.ProjectID(), vol.ID(), vol.toUpdateRequest(), rp)
	populateHTTPEnvelope(&vol.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		vol.fromResponse(resp.Data)
		vol.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, vol)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				vol.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return vol, err
	}
	if resp != nil && !resp.IsSuccess() {
		return vol, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return vol, nil
}

// Delete removes the BlockStorage identified by Ref.
func (a *volumesClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, blockStorageID, err := blockStorageIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, blockStorageID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of BlockStorage in the given parent scope.
func (a *volumesClientAdapter) List(ctx context.Context, project Ref, opts ...CallOption) (*List[*BlockStorage], error) {
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
	var items []*BlockStorage
	if resp != nil && resp.Data != nil {
		items = make([]*BlockStorage, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			bs := &BlockStorage{}
			bs.fromResponse(&resp.Data.Values[i])
			bs.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, bs)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					bs.fromResponse(fresh.Raw())
				}
				return nil
			})
			if bs.projectID == "" {
				bs.projectID = projectID
			}
			items = append(items, bs)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*BlockStorage], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*BlockStorage], error) {
		fetch := listPageFetch[types.BlockStorageListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*BlockStorage
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*BlockStorage, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				bs := &BlockStorage{}
				bs.fromResponse(&pageResp.Data.Values[i])
				bs.setRefresh(func(ctx context.Context) error {
					fresh, err := a.Get(ctx, bs)
					if err != nil {
						return err
					}
					if fresh != nil && fresh.Raw() != nil {
						bs.fromResponse(fresh.Raw())
					}
					return nil
				})
				if bs.projectID == "" {
					bs.projectID = projectID
				}
				pageItems = append(pageItems, bs)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}

// blockStorageIDsFromRef extracts (projectID, blockStorageID) from a Ref.
func blockStorageIDsFromRef(ref Ref) (projectID, blockStorageID string, err error) {
	bid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withBlockStorageID); ok {
			return w.BlockStorageID(), true
		}
		return "", false
	}, "blockStorages")
	if !ok || bid == "" {
		return "", "", fmt.Errorf("cannot determine BlockStorage ID from Ref %q", ref.URI())
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
