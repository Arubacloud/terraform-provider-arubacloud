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

// Database is the wrapper for an Aruba Cloud database (a child of a DBaaS
// instance). Construct with aruba.NewDatabase() and bind via InDBaaS(parent).
//
// Family B: flat request (no Metadata/Properties boxing, no metadataMixin,
// no tags, no location).
//
// Identity: DatabaseResponse carries no `id` field; the Name IS the path
// identifier. ID() and DatabaseID() return the locally-set name, and URI()
// is constructed from (projectID, dbaasID, name).
type Database struct {
	errMixin
	refreshMixin
	dbaasScopedMixin
	responseMetadataMixin
	httpEnvelopeMixin

	name     *string
	response *types.DatabaseResponse
}

// NewDatabase returns a fresh *Database ready for fluent setters and a Create call.
// Binds dbaasScopedMixin's error sink so InDBaaS failures surface via Err().
func NewDatabase() *Database {
	d := &Database{}
	d.dbaasScopedMixin = bindDBaaSScoped(&d.errMixin)
	return d
}

// Setters — chainable, general → specific

// InDBaaS binds this Database to its parent DBaaS instance. Required before Create.
func (d *Database) InDBaaS(parent Ref) *Database { d.intoDBaaS(parent); return d }

// Named sets the resource name. Required by the API.
func (d *Database) Named(name string) *Database { d.name = &name; return d }

// Getters — general → specific

// Ref + ID accessors.

// ID returns the database's name (which serves as its path identifier).
// Shadows responseMetadataMixin.ID() since the response has no separate id field.
func (d *Database) ID() string { return dbDerefString(d.name) }

// DatabaseID is an alias for ID() and satisfies withDatabaseID for child wrappers.
func (d *Database) DatabaseID() string { return d.ID() }

// URI constructs the canonical URI from (projectID, dbaasID, name).
// Returns "" if any component is missing.
func (d *Database) URI() string {
	pid, did, name := d.ProjectID(), d.DBaaSID(), d.ID()
	if pid == "" || did == "" || name == "" {
		return ""
	}
	return fmt.Sprintf("/projects/%s/providers/Aruba.Database/dbaas/%s/databases/%s", pid, did, name)
}

// Read accessors.

// Name returns the database name from the response if available, else from the
// locally-set value.
func (d *Database) Name() string {
	if d.response != nil && d.response.Name != "" {
		return d.response.Name
	}
	return dbDerefString(d.name)
}

// CreatedAt returns the database creation time from the response.
func (d *Database) CreatedAt() time.Time {
	if d.response != nil && d.response.CreationDate != nil {
		return *d.response.CreationDate
	}
	return time.Time{}
}

// CreatedBy returns the identity that created this database.
func (d *Database) CreatedBy() string {
	if d.response != nil && d.response.CreatedBy != nil {
		return *d.response.CreatedBy
	}
	return ""
}

// Raw shadows responseMetadataMixin.Raw() with the typed Database response.
func (d *Database) Raw() *types.DatabaseResponse { return d.response }
func (d *Database) RawJSON() []byte              { return marshalRawJSON(d.response) }
func (d *Database) RawYAML() []byte              { return marshalRawYAML(d.response) }

// RawRequest returns what toRequest() would emit right now.
func (d *Database) RawRequest() types.DatabaseRequest { return d.toRequest() }

// Wire converters

// toRequest assembles the Create/Update body from current setter state. Defaults are applied at the wire boundary.
func (d *Database) toRequest() types.DatabaseRequest {
	return types.DatabaseRequest{Name: dbDerefString(d.name)}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (d *Database) fromResponse(resp *types.DatabaseResponse) {
	if resp == nil {
		return
	}
	d.response = resp
	if resp.Name != "" {
		v := resp.Name
		d.name = &v
	}
}

func dbDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// databasesLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type databasesLowLevelClient interface {
	List(ctx context.Context, projectID, dbaasID string, params *types.RequestParameters) (*types.Response[types.DatabaseListResponse], error)
	Get(ctx context.Context, projectID, dbaasID, databaseID string, params *types.RequestParameters) (*types.Response[types.DatabaseResponse], error)
	Create(ctx context.Context, projectID, dbaasID string, body types.DatabaseRequest, params *types.RequestParameters) (*types.Response[types.DatabaseResponse], error)
	Update(ctx context.Context, projectID, dbaasID, databaseID string, body types.DatabaseRequest, params *types.RequestParameters) (*types.Response[types.DatabaseResponse], error)
	Delete(ctx context.Context, projectID, dbaasID, databaseID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// databasesClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates Database ↔ types.DatabaseRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type databasesClientAdapter struct {
	low  databasesLowLevelClient
	rest *restclient.Client
}

func newDatabasesClientAdapter(rest *restclient.Client) *databasesClientAdapter {
	if rest == nil {
		return &databasesClientAdapter{}
	}
	return &databasesClientAdapter{low: database.NewDatabasesClientImpl(rest), rest: rest}
}

// Create posts a new Database to the API and hydrates the wrapper from the response.
func (a *databasesClientAdapter) Create(ctx context.Context, db *Database, opts ...CallOption) (*Database, error) {
	if err := db.Err(); err != nil {
		return db, err
	}
	if db.ProjectID() == "" {
		return db, fmt.Errorf("Create: Database has no parent project — call InDBaaS first")
	}
	if db.DBaaSID() == "" {
		return db, fmt.Errorf("Create: Database has no parent DBaaS — call InDBaaS first")
	}
	if db.Name() == "" {
		return db, fmt.Errorf("Create: Database has no name — call Named first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, db.ProjectID(), db.DBaaSID(), db.toRequest(), rp)
	populateHTTPEnvelope(&db.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		db.fromResponse(resp.Data)
		db.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, db)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				db.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return db, err
	}
	if resp != nil && !resp.IsSuccess() {
		return db, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return db, nil
}

// Update sends a PUT for the current wrapper state. Requires ID and parent.
func (a *databasesClientAdapter) Update(ctx context.Context, db *Database, opts ...CallOption) (*Database, error) {
	if err := db.Err(); err != nil {
		return db, err
	}
	if db.DatabaseID() == "" {
		return db, fmt.Errorf("Update: Database has no ID — call Named first")
	}
	if db.DBaaSID() == "" {
		return db, fmt.Errorf("Update: Database has no parent DBaaS — call InDBaaS first")
	}
	if db.ProjectID() == "" {
		return db, fmt.Errorf("Update: Database has no parent project — call InDBaaS first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, db.ProjectID(), db.DBaaSID(), db.DatabaseID(), db.toRequest(), rp)
	populateHTTPEnvelope(&db.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		db.fromResponse(resp.Data)
		db.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, db)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				db.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return db, err
	}
	if resp != nil && !resp.IsSuccess() {
		return db, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return db, nil
}

// Get fetches a Database by Ref and returns a freshly hydrated wrapper.
func (a *databasesClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*Database, error) {
	projectID, dbaasID, databaseID, err := databaseIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, dbaasID, databaseID, rp)
	out := &Database{}
	out.dbaasID = dbaasID
	out.projectID = projectID
	name := databaseID
	out.name = &name
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

// Delete removes the Database identified by Ref.
func (a *databasesClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, dbaasID, databaseID, err := databaseIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, dbaasID, databaseID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of Database in the given parent scope.
func (a *databasesClientAdapter) List(ctx context.Context, dbaas Ref, opts ...CallOption) (*List[*Database], error) {
	projectID, dbaasID, err := dbaasIDsFromRef(dbaas)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.List(ctx, projectID, dbaasID, rp)
	if err != nil {
		return nil, err
	}
	if resp != nil && !resp.IsSuccess() {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	var items []*Database
	if resp != nil && resp.Data != nil {
		items = make([]*Database, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			db := &Database{}
			db.dbaasID = dbaasID
			db.projectID = projectID
			db.fromResponse(&resp.Data.Values[i])
			db.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, db)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					db.fromResponse(fresh.Raw())
				}
				return nil
			})
			items = append(items, db)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*Database], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*Database], error) {
		fetch := listPageFetch[types.DatabaseListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*Database
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*Database, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				db := &Database{}
				db.dbaasID = dbaasID
				db.projectID = projectID
				db.fromResponse(&pageResp.Data.Values[i])
				db.setRefresh(func(ctx context.Context) error {
					fresh, err := a.Get(ctx, db)
					if err != nil {
						return err
					}
					if fresh != nil && fresh.Raw() != nil {
						db.fromResponse(fresh.Raw())
					}
					return nil
				})
				pageItems = append(pageItems, db)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}

// databaseIDsFromRef extracts (projectID, dbaasID, databaseID) from a Ref.
func databaseIDsFromRef(ref Ref) (projectID, dbaasID, databaseID string, err error) {
	name, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withDatabaseID); ok {
			return w.DatabaseID(), true
		}
		return "", false
	}, "databases")
	if !ok || name == "" {
		return "", "", "", fmt.Errorf("cannot determine database ID from Ref %q", ref.URI())
	}
	did, ok := extractID(ref, func(r Ref) (string, bool) {
		if w, ok := r.(withDBaaSID); ok {
			return w.DBaaSID(), true
		}
		return "", false
	}, "dbaas")
	if !ok || did == "" {
		return "", "", "", fmt.Errorf("cannot determine DBaaS ID from Ref %q", ref.URI())
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
	return pid, did, name, nil
}
