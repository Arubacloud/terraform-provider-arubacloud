package aruba

import (
	"context"
	"fmt"
	"time"

	"github.com/Arubacloud/sdk-go/internal/clients/database"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// Grant is the wrapper for an Aruba Cloud DBaaS grant (a child of a Database).
// Construct with aruba.NewGrant() and bind via IntoDatabase(parent).
//
// Family B: flat request (no Metadata/Properties boxing, no metadataMixin,
// no tags, no location).
//
// Identity quirk: GrantResponse carries no `id` field, but the path segment
// /grants/<grantID> uses a server-supplied opaque ID. Practical consequences:
//
//   - After Create, ID() stays empty — the wire response cannot reveal it.
//     A subsequent Update/Delete on the same wrapper will fail pre-flight.
//     Discover the new grant via List and use the typed *Grant from there,
//     or call Get with a URI Ref carrying /grants/<id>.
//   - List items have empty ID() and empty URI() for the same reason.
//   - Get with URI(".../grants/<id>") populates ID() from the URL segment.
//   - Update requires the wrapper to already carry ID() — typically from Get.
type Grant struct {
	errMixin
	refreshMixin
	databaseScopedMixin
	responseMetadataMixin
	httpEnvelopeMixin

	id       *string
	username *string
	roleName *string
	response *types.GrantResponse
}

// NewGrant returns a fresh *Grant ready for fluent setters and a Create call.
// Binds databaseScopedMixin's error sink so IntoDatabase failures surface via Err().
func NewGrant() *Grant {
	g := &Grant{}
	g.databaseScopedMixin = bindDatabaseScoped(&g.errMixin)
	return g
}

// Setters — chainable, general → specific

// InDatabase binds this Grant to its parent Database. Required before Create.
func (g *Grant) InDatabase(parent Ref) *Grant { g.intoDatabase(parent); return g }

// ForUser sets the database username this grant applies to. Wire field: User.Username.
func (g *Grant) ForUser(name string) *Grant { g.username = &name; return g }

// OfRole sets the role to grant to the user. Wire field: Role.Name.
func (g *Grant) OfRole(name string) *Grant { g.roleName = &name; return g }

// Getters — general → specific

// ID returns the opaque server-supplied grantID. See type docstring for when
// this is and isn't populated. Shadows responseMetadataMixin.ID() since the
// response has no id field.
func (g *Grant) ID() string { return grantDerefString(g.id) }

// URI constructs the canonical URI from (projectID, dbaasID, databaseID, ID).
// Returns "" if any component is missing.
func (g *Grant) URI() string {
	pid, did, dbid, gid := g.ProjectID(), g.DBaaSID(), g.DatabaseID(), g.ID()
	if pid == "" || did == "" || dbid == "" || gid == "" {
		return ""
	}
	return fmt.Sprintf(
		"/projects/%s/providers/Aruba.Database/dbaas/%s/databases/%s/grants/%s",
		pid, did, dbid, gid,
	)
}

// Read accessors.

// Username returns the username from the response if available, else from the
// locally-set value.
func (g *Grant) Username() string {
	if g.response != nil && g.response.User.Username != "" {
		return g.response.User.Username
	}
	return grantDerefString(g.username)
}

// RoleName returns the role name from the response if available, else from the
// locally-set value.
func (g *Grant) RoleName() string {
	if g.response != nil && g.response.Role.Name != "" {
		return g.response.Role.Name
	}
	return grantDerefString(g.roleName)
}

// DatabaseName returns the database name from the response.
func (g *Grant) DatabaseName() string {
	if g.response != nil {
		return g.response.Database.Name
	}
	return ""
}

// CreatedAt returns the grant creation time from the response.
func (g *Grant) CreatedAt() time.Time {
	if g.response != nil && g.response.CreationDate != nil {
		return *g.response.CreationDate
	}
	return time.Time{}
}

// CreatedBy returns the identity that created this grant.
func (g *Grant) CreatedBy() string {
	if g.response != nil && g.response.CreatedBy != nil {
		return *g.response.CreatedBy
	}
	return ""
}

// Raw shadows responseMetadataMixin.Raw() with the typed Grant response.
func (g *Grant) Raw() *types.GrantResponse { return g.response }
func (g *Grant) RawJSON() []byte           { return marshalRawJSON(g.response) }
func (g *Grant) RawYAML() []byte           { return marshalRawYAML(g.response) }

// RawRequest returns what toRequest() would emit right now.
func (g *Grant) RawRequest() types.GrantRequest { return g.toRequest() }

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (g *Grant) toRequest() types.GrantRequest {
	return types.GrantRequest{
		User: types.GrantUserCommon{Username: grantDerefString(g.username)},
		Role: types.GrantRoleCommon{Name: grantDerefString(g.roleName)},
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (g *Grant) fromResponse(resp *types.GrantResponse) {
	if resp == nil {
		return
	}
	g.response = resp
	if resp.User.Username != "" {
		v := resp.User.Username
		g.username = &v
	}
	if resp.Role.Name != "" {
		v := resp.Role.Name
		g.roleName = &v
	}
	// Do not touch g.id — GrantResponse has no id field. The opaque grantID
	// is set by the adapter Get from a URI Ref's path segment, never here.
}

func grantDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// grantIDsFromRef extracts (projectID, dbaasID, databaseID, grantID) from a Ref.
func grantIDsFromRef(ref Ref) (projectID, dbaasID, databaseID, grantID string, err error) {
	gid, ok := extractID(ref, func(r Ref) (string, bool) {
		return "", false // no withGrantID interface; rely on URI segment fallback
	}, "grants")
	if !ok || gid == "" {
		return "", "", "", "", fmt.Errorf("cannot determine grant ID from Ref %q", ref.URI())
	}
	dbid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withDatabaseID); ok {
			return w.DatabaseID(), true
		}
		return "", false
	}, "databases")
	if !ok || dbid == "" {
		return "", "", "", "", fmt.Errorf("cannot determine database ID from Ref %q", ref.URI())
	}
	did, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withDBaaSID); ok {
			return w.DBaaSID(), true
		}
		return "", false
	}, "dbaas")
	if !ok || did == "" {
		return "", "", "", "", fmt.Errorf("cannot determine DBaaS ID from Ref %q", ref.URI())
	}
	pid, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withProjectID); ok {
			return w.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok || pid == "" {
		return "", "", "", "", fmt.Errorf("cannot determine project ID from Ref %q", ref.URI())
	}
	return pid, did, dbid, gid, nil
}

// grantsLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type grantsLowLevelClient interface {
	List(ctx context.Context, projectID, dbaasID, databaseID string, params *types.RequestParameters) (*types.Response[types.GrantListResponse], error)
	Get(ctx context.Context, projectID, dbaasID, databaseID, grantID string, params *types.RequestParameters) (*types.Response[types.GrantResponse], error)
	Create(ctx context.Context, projectID, dbaasID, databaseID string, body types.GrantRequest, params *types.RequestParameters) (*types.Response[types.GrantResponse], error)
	Update(ctx context.Context, projectID, dbaasID, databaseID, grantID string, body types.GrantRequest, params *types.RequestParameters) (*types.Response[types.GrantResponse], error)
	Delete(ctx context.Context, projectID, dbaasID, databaseID, grantID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// grantsClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates Grant ↔ types.GrantRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type grantsClientAdapter struct {
	low  grantsLowLevelClient
	rest *restclient.Client
}

var _ GrantsClient = (*grantsClientAdapter)(nil)

func newGrantsClientAdapter(rest *restclient.Client) *grantsClientAdapter {
	if rest == nil {
		return &grantsClientAdapter{}
	}
	return &grantsClientAdapter{low: database.NewGrantsClientImpl(rest), rest: rest}
}

// Create posts a new Grant to the API and hydrates the wrapper from the response.
func (a *grantsClientAdapter) Create(ctx context.Context, g *Grant, opts ...CallOption) (*Grant, error) {
	if err := g.Err(); err != nil {
		return g, err
	}
	if g.ProjectID() == "" {
		return g, fmt.Errorf("Create: Grant has no parent project — call InDatabase first")
	}
	if g.DBaaSID() == "" {
		return g, fmt.Errorf("Create: Grant has no parent DBaaS — call InDatabase first")
	}
	if g.DatabaseID() == "" {
		return g, fmt.Errorf("Create: Grant has no parent database — call InDatabase first")
	}
	if g.username == nil {
		return g, fmt.Errorf("Create: Grant has no username — call ForUser first")
	}
	if g.roleName == nil {
		return g, fmt.Errorf("Create: Grant has no role — call OfRole first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, g.ProjectID(), g.DBaaSID(), g.DatabaseID(), g.toRequest(), rp)
	populateHTTPEnvelope(&g.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		g.fromResponse(resp.Data)
		g.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, g)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				g.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return g, err
	}
	if resp != nil && !resp.IsSuccess() {
		return g, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return g, nil
}

// Update sends a PUT for the current wrapper state. Requires ID and parent.
func (a *grantsClientAdapter) Update(ctx context.Context, g *Grant, opts ...CallOption) (*Grant, error) {
	if err := g.Err(); err != nil {
		return g, err
	}
	if g.ID() == "" {
		return g, fmt.Errorf("Update: Grant has no ID — get the grant via Get first to obtain the opaque ID")
	}
	if g.DatabaseID() == "" {
		return g, fmt.Errorf("Update: Grant has no parent database — call InDatabase first")
	}
	if g.DBaaSID() == "" {
		return g, fmt.Errorf("Update: Grant has no parent DBaaS — call InDatabase first")
	}
	if g.ProjectID() == "" {
		return g, fmt.Errorf("Update: Grant has no parent project — call InDatabase first")
	}
	if g.username == nil {
		return g, fmt.Errorf("Update: Grant has no username — call ForUser first")
	}
	if g.roleName == nil {
		return g, fmt.Errorf("Update: Grant has no role — call OfRole first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, g.ProjectID(), g.DBaaSID(), g.DatabaseID(), g.ID(), g.toRequest(), rp)
	populateHTTPEnvelope(&g.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		g.fromResponse(resp.Data)
		g.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, g)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				g.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return g, err
	}
	if resp != nil && !resp.IsSuccess() {
		return g, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return g, nil
}

// Get fetches a Grant by Ref and returns a freshly hydrated wrapper.
func (a *grantsClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*Grant, error) {
	projectID, dbaasID, databaseID, grantID, err := grantIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, dbaasID, databaseID, grantID, rp)
	out := &Grant{}
	out.databaseID = databaseID
	out.dbaasID = dbaasID
	out.projectID = projectID
	out.id = &grantID
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

// Delete removes the Grant identified by Ref.
func (a *grantsClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, dbaasID, databaseID, grantID, err := grantIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, dbaasID, databaseID, grantID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of Grants in the given Database scope.
func (a *grantsClientAdapter) List(ctx context.Context, parent Ref, opts ...CallOption) (*List[*Grant], error) {
	projectID, dbaasID, databaseID, err := databaseIDsFromRef(parent)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.List(ctx, projectID, dbaasID, databaseID, rp)
	if err != nil {
		return nil, err
	}
	if resp != nil && !resp.IsSuccess() {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	var items []*Grant
	if resp != nil && resp.Data != nil {
		items = make([]*Grant, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			g := &Grant{}
			g.databaseID = databaseID
			g.dbaasID = dbaasID
			g.projectID = projectID
			g.fromResponse(&resp.Data.Values[i])
			g.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, g)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					g.fromResponse(fresh.Raw())
				}
				return nil
			})
			items = append(items, g)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*Grant], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*Grant], error) {
		fetch := listPageFetch[types.GrantListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*Grant
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*Grant, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				g := &Grant{}
				g.databaseID = databaseID
				g.dbaasID = dbaasID
				g.projectID = projectID
				g.fromResponse(&pageResp.Data.Values[i])
				g.setRefresh(func(ctx context.Context) error {
					fresh, err := a.Get(ctx, g)
					if err != nil {
						return err
					}
					if fresh != nil && fresh.Raw() != nil {
						g.fromResponse(fresh.Raw())
					}
					return nil
				})
				pageItems = append(pageItems, g)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}
