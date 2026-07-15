package network

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

type loadBalancersClientImpl struct {
	client *restclient.Client
}

// NewService creates a new unified Network service
func NewLoadBalancersClientImpl(client *restclient.Client) *loadBalancersClientImpl {
	return &loadBalancersClientImpl{
		client: client,
	}
}

// List retrieves all load balancers for a project
func (c *loadBalancersClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.LoadBalancerListResponse], error) {
	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(LoadBalancersPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &LoadBalancerListAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &LoadBalancerListAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.LoadBalancerListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific load balancer by ID
func (c *loadBalancersClientImpl) Get(ctx context.Context, projectID string, loadBalancerID string, params *types.RequestParameters) (*types.Response[types.LoadBalancerResponse], error) {
	if err := types.ValidateProjectAndResource(projectID, loadBalancerID, "load balancer ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(LoadBalancerPath, projectID, loadBalancerID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &LoadBalancerGetAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &LoadBalancerGetAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.LoadBalancerResponse](httpResp, c.client.Logger())
}
