package aruba

import (
	"context"
	"time"

	"github.com/Arubacloud/sdk-go/internal/clients/metric"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ---- Wrapper ----

// Alert is the wrapper for an Aruba Cloud metric alert.
// Instances are read-only and can only be obtained via Client.FromMetric().Alerts().List.
// There is no factory, no setters, and no individual-fetch endpoint.
type Alert struct {
	projectScopedMixin    // ProjectID() — back-filled from parent Ref at List time
	responseMetadataMixin // shadowed: ID(), Raw(); no metadata envelope on the wire
	httpEnvelopeMixin     // RawHTTP(), StatusCode(), Headers(), RawError()

	response *types.AlertResponse
}

// Getters — general → specific

// ID returns the alert ID, or "" when unset. Shadows responseMetadataMixin.ID().
func (a *Alert) ID() string {
	if a.response == nil {
		return ""
	}
	return a.response.ID
}

// URI returns "" — alerts have no individual fetch endpoint.
func (a *Alert) URI() string { return "" }

// Raw returns the underlying wire payload. Shadows responseMetadataMixin.Raw().
func (a *Alert) Raw() *types.AlertResponse { return a.response }
func (a *Alert) RawJSON() []byte           { return marshalRawJSON(a.response) }
func (a *Alert) RawYAML() []byte           { return marshalRawYAML(a.response) }

// EventID returns the event ID associated with this alert, or "" when unset.
func (a *Alert) EventID() string {
	if a.response == nil {
		return ""
	}
	return a.response.EventID
}

// EventName returns the event name, or "" when unset.
func (a *Alert) EventName() string {
	if a.response == nil {
		return ""
	}
	return a.response.EventName
}

// Username returns the username associated with this alert, or "" when unset.
func (a *Alert) Username() string {
	if a.response == nil {
		return ""
	}
	return a.response.Username
}

// ServiceCategory returns the service category, or "" when unset.
func (a *Alert) ServiceCategory() string {
	if a.response == nil {
		return ""
	}
	return a.response.ServiceCategory
}

// ServiceTypology returns the service typology, or "" when unset.
func (a *Alert) ServiceTypology() string {
	if a.response == nil {
		return ""
	}
	return a.response.ServiceTypology
}

// ServiceName returns the service name, or "" when unset.
func (a *Alert) ServiceName() string {
	if a.response == nil {
		return ""
	}
	return a.response.ServiceName
}

// ResourceID returns the resource ID, or "" when unset.
func (a *Alert) ResourceID() string {
	if a.response == nil {
		return ""
	}
	return a.response.ResourceID
}

// ResourceTypology returns the resource typology, or "" when unset.
func (a *Alert) ResourceTypology() string {
	if a.response == nil {
		return ""
	}
	return a.response.ResourceTypology
}

// Metric returns the metric name, or "" when unset.
func (a *Alert) Metric() string {
	if a.response == nil {
		return ""
	}
	return a.response.Metric
}

// LastReception returns the timestamp of the last alert reception.
func (a *Alert) LastReception() time.Time {
	if a.response == nil {
		return time.Time{}
	}
	return a.response.LastReception
}

// Rule returns the rule name, or "" when unset.
func (a *Alert) Rule() string {
	if a.response == nil {
		return ""
	}
	return a.response.Rule
}

// Threshold returns the alert threshold value.
// Note: the upstream field is misspelled as "Theshold" — this accessor normalizes the name.
func (a *Alert) Threshold() int {
	if a.response == nil {
		return 0
	}
	return int(a.response.Theshold)
}

// UM returns the unit of measure, or "" when unset.
func (a *Alert) UM() string {
	if a.response == nil {
		return ""
	}
	return a.response.UM
}

// Duration returns the duration string, or "" when unset.
func (a *Alert) Duration() string {
	if a.response == nil {
		return ""
	}
	return a.response.Duration
}

// ThresholdExceedance returns whether the threshold was exceeded.
// Note: the upstream field is misspelled as "ThesholdExceedence" — this accessor normalizes the name.
func (a *Alert) ThresholdExceedance() string {
	if a.response == nil {
		return ""
	}
	return a.response.ThesholdExceedence
}

// Component returns the component name, or "" when unset.
func (a *Alert) Component() string {
	if a.response == nil {
		return ""
	}
	return a.response.Component
}

// ClusterTypology returns the cluster typology, or "" when unset.
func (a *Alert) ClusterTypology() string {
	if a.response == nil {
		return ""
	}
	return a.response.ClusterTypology
}

// Cluster returns the cluster identifier, or "" when unset.
func (a *Alert) Cluster() string {
	if a.response == nil {
		return ""
	}
	return a.response.Cluster
}

// Clustername returns the cluster name, or "" when unset.
func (a *Alert) Clustername() string {
	if a.response == nil {
		return ""
	}
	return a.response.Clustername
}

// NodePool returns the node pool name, or "" when unset.
func (a *Alert) NodePool() string {
	if a.response == nil {
		return ""
	}
	return a.response.NodePool
}

// SMS reports whether SMS notification is enabled for this alert.
func (a *Alert) SMS() bool {
	if a.response == nil {
		return false
	}
	return a.response.SMS
}

// Email reports whether email notification is enabled for this alert.
func (a *Alert) Email() bool {
	if a.response == nil {
		return false
	}
	return a.response.Email
}

// Panel reports whether panel notification is enabled for this alert.
func (a *Alert) Panel() bool {
	if a.response == nil {
		return false
	}
	return a.response.Panel
}

// Hidden reports whether this alert is hidden.
func (a *Alert) Hidden() bool {
	if a.response == nil {
		return false
	}
	return a.response.Hidden
}

// ExecutedAlertActions returns the list of executed alert actions, or nil when absent.
func (a *Alert) ExecutedAlertActions() []types.ExecutedAlertActionResponse {
	if a.response == nil {
		return nil
	}
	return a.response.ExecutedAlertActions
}

// Actions returns the list of available alert actions, or nil when absent.
func (a *Alert) Actions() []types.AlertActionResponse {
	if a.response == nil {
		return nil
	}
	return a.response.Actions
}

func (a *Alert) fromResponse(resp *types.AlertResponse) {
	if resp == nil {
		return
	}
	a.response = resp
}

// ---- Low-level client interface ----

// alertsLowLevelClient is the contract the wrapper depends on. Returning
// *types.Response[T] preserves HTTP envelope details (status code, headers,
// raw body) for the wrapper's diagnostics.
type alertsLowLevelClient interface {
	List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.AlertsListResponse], error)
}

// ---- Adapter ----

// alertsClientAdapter bridges the wrapper API (chainable, error-accumulating,
// wire-shape-hidden) to the low-level client (parameter-explicit, returning
// typed wire structs). Translates Alert ↔ types.AlertResponse and surfaces HTTP
// errors as *aruba.HTTPError.
type alertsClientAdapter struct {
	low  alertsLowLevelClient
	rest *restclient.Client
}

var _ AlertsClient = (*alertsClientAdapter)(nil)

func newAlertsClientAdapter(rest *restclient.Client) *alertsClientAdapter {
	if rest == nil {
		return &alertsClientAdapter{}
	}
	return &alertsClientAdapter{low: metric.NewAlertsClientImpl(rest), rest: rest}
}

// List returns a paginated list of Alert in the given parent scope.
func (a *alertsClientAdapter) List(ctx context.Context, project Ref, opts ...CallOption) (*List[*Alert], error) {
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
	var items []*Alert
	if resp != nil && resp.Data != nil {
		items = make([]*Alert, 0, len(resp.Data.Values))
		for i := range resp.Data.Values {
			al := &Alert{}
			al.fromResponse(&resp.Data.Values[i])
			al.projectID = projectID
			populateHTTPEnvelope(&al.httpEnvelopeMixin, resp)
			items = append(items, al)
		}
	}
	var refetch func(ctx context.Context, pageURL string) (*List[*Alert], error)
	refetch = func(ctx context.Context, pageURL string) (*List[*Alert], error) {
		fetch := listPageFetch[types.AlertsListResponse](a.rest, opts)
		pageResp, fetchErr := fetch(ctx, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}
		if pageResp != nil && !pageResp.IsSuccess() {
			return nil, &HTTPError{StatusCode: pageResp.StatusCode, Body: pageResp.RawBody, ErrResp: pageResp.Error}
		}
		var pageItems []*Alert
		if pageResp != nil && pageResp.Data != nil {
			pageItems = make([]*Alert, 0, len(pageResp.Data.Values))
			for i := range pageResp.Data.Values {
				al := &Alert{}
				al.fromResponse(&pageResp.Data.Values[i])
				al.projectID = projectID
				populateHTTPEnvelope(&al.httpEnvelopeMixin, pageResp)
				pageItems = append(pageItems, al)
			}
		}
		return newListFromResponse(pageItems, pageResp, opts, refetch), nil
	}
	return newListFromResponse(items, resp, opts, refetch), nil
}
