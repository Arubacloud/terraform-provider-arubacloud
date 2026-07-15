package audit

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// eventsClientImpl implements the AuditAPI interface for all Audit operations
type eventsClientImpl struct {
	client *restclient.Client
}

// NewEventsClientImpl creates a new unified Audit service
func NewEventsClientImpl(client *restclient.Client) *eventsClientImpl {
	return &eventsClientImpl{
		client: client,
	}
}

// List retrieves all audit events for a project
func (s *eventsClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.AuditEventListResponse], error) {
	s.client.Logger().Debugf("Listing audit events for project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(EventsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &AuditLogListVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &AuditLogListVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := s.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.AuditEventListResponse](httpResp, s.client.Logger())
}
