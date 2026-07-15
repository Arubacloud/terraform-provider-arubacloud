package aruba

import (
	"context"
	"time"

	"github.com/Arubacloud/sdk-go/internal/clients/audit"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// AuditEvent is the wrapper for an Aruba Cloud audit event.
// Instances are read-only and can only be obtained via Client.FromAudit().Events().List.
// There is no factory, no setters, and no individual-fetch endpoint.
type AuditEvent struct {
	projectScopedMixin    // ProjectID() — back-filled from parent Ref at List time
	responseMetadataMixin // shadowed: ID(), URI(), CreatedAt(), Raw()
	httpEnvelopeMixin     // RawHTTP(), StatusCode(), Headers(), RawError()

	response *types.AuditEventResponse // backs Raw() and all field accessors
}

// Getters — general → specific

// ID returns the event identifier (from Event.ID), or "" when unset.
// Shadows responseMetadataMixin.ID().
func (e *AuditEvent) ID() string {
	if e.response == nil {
		return ""
	}
	return e.response.Event.ID
}

// URI returns "" — audit events have no individual fetch endpoint.
func (e *AuditEvent) URI() string { return "" }

// CreatedAt returns the event timestamp. Shadows responseMetadataMixin.CreatedAt().
func (e *AuditEvent) CreatedAt() time.Time {
	if e.response == nil {
		return time.Time{}
	}
	return e.response.Timestamp
}

// Raw returns the underlying wire payload. Shadows responseMetadataMixin.Raw().
func (e *AuditEvent) Raw() *types.AuditEventResponse { return e.response }
func (e *AuditEvent) RawJSON() []byte                { return marshalRawJSON(e.response) }
func (e *AuditEvent) RawYAML() []byte                { return marshalRawYAML(e.response) }

// SeverityLevel returns the event severity level, or "" when unset.
func (e *AuditEvent) SeverityLevel() string {
	if e.response == nil {
		return ""
	}
	return e.response.SeverityLevel
}

// Origin returns the event origin, or "" when unset.
func (e *AuditEvent) Origin() string {
	if e.response == nil {
		return ""
	}
	return e.response.Origin
}

// Channel returns the event channel, or "" when unset.
func (e *AuditEvent) Channel() string {
	if e.response == nil {
		return ""
	}
	return e.response.Channel
}

// LogFormat returns the log format version.
func (e *AuditEvent) LogFormat() types.EventLogFormatVersionResponse {
	if e.response == nil {
		return types.EventLogFormatVersionResponse{}
	}
	return e.response.LogFormat
}

// Operation returns the operation associated with this event.
func (e *AuditEvent) Operation() types.EventOperationResponse {
	if e.response == nil {
		return types.EventOperationResponse{}
	}
	return e.response.Operation
}

// Event returns the event information (ID, type, value).
func (e *AuditEvent) Event() types.EventInfoResponse {
	if e.response == nil {
		return types.EventInfoResponse{}
	}
	return e.response.Event
}

// Category returns the event category.
func (e *AuditEvent) Category() types.EventCategoryResponse {
	if e.response == nil {
		return types.EventCategoryResponse{}
	}
	return e.response.Category
}

// Region returns the region this event was emitted in, or "" when absent.
func (e *AuditEvent) Region() Region {
	if e.response == nil || e.response.Region == nil || e.response.Region.Name == nil {
		return ""
	}
	return Region(*e.response.Region.Name)
}

// Zone returns the availability zone this event was emitted in, or "" when absent.
func (e *AuditEvent) Zone() Zone {
	if e.response == nil || e.response.Region == nil || e.response.Region.AvailabilityZone == nil {
		return ""
	}
	return Zone(*e.response.Region.AvailabilityZone)
}

// Status returns the event status.
func (e *AuditEvent) Status() types.EventStatusResponse {
	if e.response == nil {
		return types.EventStatusResponse{}
	}
	return e.response.Status
}

// SubStatus returns the optional sub-status, or nil when absent.
func (e *AuditEvent) SubStatus() *types.EventSubStatusResponse {
	if e.response == nil {
		return nil
	}
	return e.response.SubStatus
}

// Identity returns the caller identity for this event.
func (e *AuditEvent) Identity() types.EventIdentityResponse {
	if e.response == nil {
		return types.EventIdentityResponse{}
	}
	return e.response.Identity
}

// Properties returns the arbitrary properties map, or nil when absent.
func (e *AuditEvent) Properties() map[string]interface{} {
	if e.response == nil {
		return nil
	}
	return e.response.Properties
}

// Actions returns the available actions for this event, or nil when absent.
func (e *AuditEvent) Actions() []types.EventActionResponse {
	if e.response == nil {
		return nil
	}
	return e.response.Actions
}

// CategoryID returns the category ID, or "" when absent.
func (e *AuditEvent) CategoryID() string {
	if e.response == nil || e.response.CategoryID == nil {
		return ""
	}
	return *e.response.CategoryID
}

// TypologyID returns the typology ID, or "" when absent.
func (e *AuditEvent) TypologyID() string {
	if e.response == nil || e.response.TypologyID == nil {
		return ""
	}
	return *e.response.TypologyID
}

// Title returns the event title, or "" when absent.
func (e *AuditEvent) Title() string {
	if e.response == nil || e.response.Title == nil {
		return ""
	}
	return *e.response.Title
}

// EventTypeName returns the event type string from Event.Type, or "" when absent.
func (e *AuditEvent) EventTypeName() string {
	if e.response == nil {
		return ""
	}
	return e.response.Event.Type
}

// IdentityName returns the caller username, or "" when absent.
func (e *AuditEvent) IdentityName() string {
	if e.response == nil || e.response.Identity.Caller.Username == nil {
		return ""
	}
	return *e.response.Identity.Caller.Username
}

// OperationName returns the operation value string, or "" when absent.
func (e *AuditEvent) OperationName() string {
	if e.response == nil || e.response.Operation.Value == nil {
		return ""
	}
	return *e.response.Operation.Value
}

func (e *AuditEvent) fromResponse(resp *types.AuditEventResponse) {
	if resp == nil {
		return
	}
	e.response = resp
}

// ---- Low-level client interface ----

// auditEventsLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type auditEventsLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.AuditEventListResponse], error)
}

// ---- Adapter ----

// auditEventsClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates AuditEvent ↔ types.AuditEventListResponse and surfaces HTTP
// errors as *aruba.HTTPError.
type auditEventsClientAdapter struct {
	low  auditEventsLowLevelClient
	rest *restclient.Client
}

var _ EventsClient = (*auditEventsClientAdapter)(nil)

func newAuditEventsClientAdapter(rest *restclient.Client) *auditEventsClientAdapter {
	if rest == nil {
		return &auditEventsClientAdapter{}
	}
	return &auditEventsClientAdapter{low: audit.NewEventsClientImpl(rest), rest: rest}
}

// List returns a paginated list of AuditEvent in the given parent scope.
func (a *auditEventsClientAdapter) List(ctx context.Context, project Ref, opts ...CallOption) (*List[*AuditEvent], error) {
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
	var items []*AuditEvent
	if resp != nil && resp.Data != nil {
		items = make([]*AuditEvent, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			ev := &AuditEvent{}
			ev.fromResponse(&resp.Data.Values[i])
			ev.projectID = projectID
			populateHTTPEnvelope(&ev.httpEnvelopeMixin, resp)
			items = append(items, ev)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*AuditEvent], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*AuditEvent], error) {
		fetch := listPageFetch[types.AuditEventListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*AuditEvent
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*AuditEvent, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				ev := &AuditEvent{}
				ev.fromResponse(&pageResp.Data.Values[i])
				ev.projectID = projectID
				populateHTTPEnvelope(&ev.httpEnvelopeMixin, pageResp)
				pageItems = append(pageItems, ev)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}
