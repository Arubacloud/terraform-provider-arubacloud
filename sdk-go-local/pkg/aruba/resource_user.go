package aruba

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Arubacloud/sdk-go/internal/clients/database"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// User is the wrapper for an Aruba Cloud DBaaS user (a child of a DBaaS
// instance). Construct with aruba.NewUser() and bind via IntoDBaaS(parent).
//
// Family B: flat request (no Metadata/Properties boxing, no metadataMixin,
// no tags, no location).
//
// Identity: UserResponse carries no `id` field; the Username IS the path
// identifier. ID() returns the locally-set username, and URI() is
// constructed from (projectID, dbaasID, username).
//
// Password is a write-only field: WithPassword stores it locally for use in
// Create/Update wire bodies, but the wrapper deliberately exposes no
// Password() accessor. The response struct (UserResponse) contains no
// Password field, so hydration cannot accidentally surface it either.
type User struct {
	errMixin
	refreshMixin
	dbaasScopedMixin
	responseMetadataMixin
	httpEnvelopeMixin

	username *string
	password *string
	response *types.UserResponse
}

// NewUser returns a fresh *User ready for fluent setters and a Create call.
// Binds dbaasScopedMixin's error sink so IntoDBaaS failures surface via Err().
func NewUser() *User {
	u := &User{}
	u.dbaasScopedMixin = bindDBaaSScoped(&u.errMixin)
	return u
}

// Setters — chainable, general → specific

// InDBaaS binds this User to its parent DBaaS instance. Required before Create.
func (u *User) InDBaaS(parent Ref) *User { u.intoDBaaS(parent); return u }

// WithUsername sets the username. Used as the path identifier; required before Create.
func (u *User) WithUsername(name string) *User { u.username = &name; return u }

// WithPassword sets the clear-text password for Create/Update. The wrapper base64-encodes it
// at the wire boundary; callers always pass plain text. Write-only: no Password() getter is exposed.
func (u *User) WithPassword(pw string) *User { u.password = &pw; return u }

// Getters — general → specific

// ID returns the user's username (which serves as its path identifier).
// Shadows responseMetadataMixin.ID() since the response has no id field.
func (u *User) ID() string { return userDerefString(u.username) }

// URI constructs the canonical URI from (projectID, dbaasID, username).
// Returns "" if any component is missing.
func (u *User) URI() string {
	pid, did, name := u.ProjectID(), u.DBaaSID(), u.ID()
	if pid == "" || did == "" || name == "" {
		return ""
	}
	return fmt.Sprintf("/projects/%s/providers/Aruba.Database/dbaas/%s/users/%s", pid, did, name)
}

// Read accessors.

// Username returns the username from the response if available, else from the
// locally-set value.
func (u *User) Username() string {
	if u.response != nil && u.response.Username != "" {
		return u.response.Username
	}
	return userDerefString(u.username)
}

// CreatedAt returns the user creation time from the response.
func (u *User) CreatedAt() time.Time {
	if u.response != nil && u.response.CreationDate != nil {
		return *u.response.CreationDate
	}
	return time.Time{}
}

// CreatedBy returns the identity that created this user.
func (u *User) CreatedBy() string {
	if u.response != nil && u.response.CreatedBy != nil {
		return *u.response.CreatedBy
	}
	return ""
}

// Raw shadows responseMetadataMixin.Raw() with the typed User response.
func (u *User) Raw() *types.UserResponse { return u.response }
func (u *User) RawJSON() []byte          { return marshalRawJSON(u.response) }
func (u *User) RawYAML() []byte          { return marshalRawYAML(u.response) }

// RawRequest returns the wire body that would be sent on Create/Update. It
// includes the base64-encoded password if WithPassword was called — by design,
// for parity with other wrappers' RawRequest debugging surface. There is no
// Password() accessor on *User; the password is intentionally not exposed
// through any read-only path other than this wire mirror.
func (u *User) RawRequest() types.UserRequest { return u.toRequest() }

// Wire converters

// toRequest assembles the Create/Update body from current setter state.
// The API expects the password base64-encoded, so the clear-text value is encoded here at the wire boundary.
func (u *User) toRequest() types.UserRequest {
	return types.UserRequest{
		Username: userDerefString(u.username),
		Password: base64.StdEncoding.EncodeToString([]byte(userDerefString(u.password))),
	}
}

// fromResponse hydrates the wrapper from a server reply. Nil-safe.
func (u *User) fromResponse(resp *types.UserResponse) {
	if resp == nil {
		return
	}
	u.response = resp
	if resp.Username != "" {
		v := resp.Username
		u.username = &v
	}
	// Do not touch u.password — UserResponse has no Password field, and
	// the locally-set password must survive hydration so a subsequent
	// Update can still send it on the wire.
}

func userDerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ---- Low-level client interface ----

// userIDsFromRef extracts (projectID, dbaasID, userID) from a Ref.
func userIDsFromRef(ref Ref) (projectID, dbaasID, userID string, err error) {
	name, ok := extractID(ref, func(r Ref) (string, bool) {
		return "", false // no withUserID interface; rely on URI segment fallback
	}, "users")
	if !ok || name == "" {
		return "", "", "", fmt.Errorf("cannot determine user ID from Ref %q", ref.URI())
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

// usersLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type usersLowLevelClient interface {
	List(ctx context.Context, projectID, dbaasID string, params *types.RequestParameters) (*types.Response[types.DatabaseUserListResponse], error)
	Get(ctx context.Context, projectID, dbaasID, userID string, params *types.RequestParameters) (*types.Response[types.UserResponse], error)
	Create(ctx context.Context, projectID, dbaasID string, body types.UserRequest, params *types.RequestParameters) (*types.Response[types.UserResponse], error)
	Update(ctx context.Context, projectID, dbaasID, userID string, body types.UserRequest, params *types.RequestParameters) (*types.Response[types.UserResponse], error)
	Delete(ctx context.Context, projectID, dbaasID, userID string, params *types.RequestParameters) (*types.Response[any], error)
}

// ---- Adapter ----

// usersClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates User ↔ types.UserRequest/Response and
// surfaces HTTP errors as *aruba.HTTPError.
type usersClientAdapter struct {
	low  usersLowLevelClient
	rest *restclient.Client
}

var _ UsersClient = (*usersClientAdapter)(nil)

func newUsersClientAdapter(rest *restclient.Client) *usersClientAdapter {
	if rest == nil {
		return &usersClientAdapter{}
	}
	return &usersClientAdapter{low: database.NewUsersClientImpl(rest), rest: rest}
}

// Create posts a new User to the API and hydrates the wrapper from the response.
func (a *usersClientAdapter) Create(ctx context.Context, u *User, opts ...CallOption) (*User, error) {
	if err := u.Err(); err != nil {
		return u, err
	}
	if u.ProjectID() == "" {
		return u, fmt.Errorf("Create: User has no parent project — call InDBaaS first")
	}
	if u.DBaaSID() == "" {
		return u, fmt.Errorf("Create: User has no parent DBaaS — call InDBaaS first")
	}
	if u.Username() == "" {
		return u, fmt.Errorf("Create: User has no username — call WithUsername first")
	}
	if u.password == nil {
		return u, fmt.Errorf("Create: password is required — call WithPassword first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Create(ctx, u.ProjectID(), u.DBaaSID(), u.toRequest(), rp)
	populateHTTPEnvelope(&u.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		u.fromResponse(resp.Data)
		u.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, u)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				u.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return u, err
	}
	if resp != nil && !resp.IsSuccess() {
		return u, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return u, nil
}

// Update sends a PUT for the current wrapper state. Requires ID and parent.
func (a *usersClientAdapter) Update(ctx context.Context, u *User, opts ...CallOption) (*User, error) {
	if err := u.Err(); err != nil {
		return u, err
	}
	if u.ID() == "" {
		return u, fmt.Errorf("Update: User has no ID — call WithUsername first")
	}
	if u.DBaaSID() == "" {
		return u, fmt.Errorf("Update: User has no parent DBaaS — call InDBaaS first")
	}
	if u.ProjectID() == "" {
		return u, fmt.Errorf("Update: User has no parent project — call InDBaaS first")
	}
	if u.password == nil {
		return u, fmt.Errorf("Update: password is required — call WithPassword first")
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Update(ctx, u.ProjectID(), u.DBaaSID(), u.ID(), u.toRequest(), rp)
	populateHTTPEnvelope(&u.httpEnvelopeMixin, resp)
	if resp != nil && resp.Data != nil {
		u.fromResponse(resp.Data)
		u.setRefresh(func(ctx context.Context) error {
			fresh, err := a.Get(ctx, u)
			if err != nil {
				return err
			}
			if fresh != nil && fresh.Raw() != nil {
				u.fromResponse(fresh.Raw())
			}
			return nil
		})
	}
	if err != nil {
		return u, err
	}
	if resp != nil && !resp.IsSuccess() {
		return u, &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return u, nil
}

// Get fetches a User by Ref and returns a freshly hydrated wrapper.
func (a *usersClientAdapter) Get(ctx context.Context, ref Ref, opts ...CallOption) (*User, error) {
	projectID, dbaasID, userID, err := userIDsFromRef(ref)
	if err != nil {
		return nil, err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Get(ctx, projectID, dbaasID, userID, rp)
	out := &User{}
	out.dbaasID = dbaasID
	out.projectID = projectID
	name := userID
	out.username = &name
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

// Delete removes the User identified by Ref.
func (a *usersClientAdapter) Delete(ctx context.Context, ref Ref, opts ...CallOption) error {
	projectID, dbaasID, userID, err := userIDsFromRef(ref)
	if err != nil {
		return err
	}
	co := applyCallOptions(opts)
	rp := co.toRequestParameters()
	resp, err := a.low.Delete(ctx, projectID, dbaasID, userID, rp)
	if err != nil {
		return err
	}
	if resp != nil && !resp.IsSuccess() {
		return &HTTPError{StatusCode: resp.StatusCode, Body: resp.RawBody, ErrResp: resp.Error}
	}
	return nil
}

// List returns a paginated list of Users in the given DBaaS scope.
func (a *usersClientAdapter) List(ctx context.Context, dbaas Ref, opts ...CallOption) (*List[*User], error) {
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
	var items []*User
	if resp != nil && resp.Data != nil {
		items = make([]*User, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			u := &User{}
			u.dbaasID = dbaasID
			u.projectID = projectID
			u.fromResponse(&resp.Data.Values[i])
			u.setRefresh(func(ctx context.Context) error {
				fresh, err := a.Get(ctx, u)
				if err != nil {
					return err
				}
				if fresh != nil && fresh.Raw() != nil {
					u.fromResponse(fresh.Raw())
				}
				return nil
			})
			items = append(items, u)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*User], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*User], error) {
		fetch := listPageFetch[types.DatabaseUserListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*User
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*User, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				item := &User{}
				item.dbaasID = dbaasID
				item.projectID = projectID
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
				pageItems = append(pageItems, item)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}
