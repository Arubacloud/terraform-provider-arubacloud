package metric

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

type metricssClientImpl struct {
	client *restclient.Client
}

// NewAlertsClientImpl creates a new unified Metric service
func NewMetricsClientImpl(client *restclient.Client) *metricssClientImpl {
	return &metricssClientImpl{
		client: client,
	}
}

// List retrieves all metrics for a project
func (c *metricssClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.MetricListResponse], error) {
	c.client.Logger().Debugf("Listing metrics for project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(MetricsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &MetricListVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &MetricListVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.MetricListResponse](httpResp, c.client.Logger())
}
