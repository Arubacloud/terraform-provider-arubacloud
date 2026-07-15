package metric

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

type alertsClientImpl struct {
	client *restclient.Client
}

// NewAlertsClientImpl creates a new unified Metric service
func NewAlertsClientImpl(client *restclient.Client) *alertsClientImpl {
	return &alertsClientImpl{
		client: client,
	}
}

// List retrieves all alerts for a project
func (c *alertsClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.AlertsListResponse], error) {
	c.client.Logger().Debugf("Listing alerts for project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(AlertsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &AlertListVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &AlertListVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.AlertsListResponse](httpResp, c.client.Logger())
}
