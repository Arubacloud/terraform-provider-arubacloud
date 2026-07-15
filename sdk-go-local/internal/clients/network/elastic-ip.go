package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

type elasticIPsClientImpl struct {
	client *restclient.Client
}

// NewService creates a new unified Network service
func NewElasticIPsClientImpl(client *restclient.Client) *elasticIPsClientImpl {
	return &elasticIPsClientImpl{
		client: client,
	}
}

// List retrieves all elastic IPs for a project
func (c *elasticIPsClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.ElasticIPListResponse], error) {
	c.client.Logger().Debugf("Listing elastic IPs for project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(ElasticIPsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ElasticIPListAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ElasticIPListAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.ElasticIPListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific elastic IP by ID
func (c *elasticIPsClientImpl) Get(ctx context.Context, projectID string, elasticIPID string, params *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error) {
	c.client.Logger().Debugf("Getting elastic IP: %s in project: %s", elasticIPID, projectID)

	if err := types.ValidateProjectAndResource(projectID, elasticIPID, "elastic IP ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(ElasticIPPath, projectID, elasticIPID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ElasticIPGetAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ElasticIPGetAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.ElasticIPResponse](httpResp, c.client.Logger())
}

// Create creates a new elastic IP
func (c *elasticIPsClientImpl) Create(ctx context.Context, projectID string, body types.ElasticIPRequest, params *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error) {
	c.client.Logger().Debugf("Creating elastic IP in project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(ElasticIPsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ElasticIPCreateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ElasticIPCreateAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	// Marshal the request body to JSON
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpResp, err := c.client.DoRequest(ctx, http.MethodPost, path, bytes.NewReader(bodyBytes), queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	resp, err := types.ParseResponseBody[types.ElasticIPResponse](httpResp, c.client.Logger())
	if err != nil {
		return nil, err
	}
	if resp.IsSuccess() && resp.Data != nil {
		if err := resp.Data.Metadata.Validate(); err != nil {
			return resp, fmt.Errorf("invalid create response: %w", err)
		}
	}
	return resp, nil
}

// Update updates an existing elastic IP
func (c *elasticIPsClientImpl) Update(ctx context.Context, projectID string, elasticIPID string, body types.ElasticIPRequest, params *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error) {
	c.client.Logger().Debugf("Updating elastic IP: %s in project: %s", elasticIPID, projectID)

	if err := types.ValidateProjectAndResource(projectID, elasticIPID, "elastic IP ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(ElasticIPPath, projectID, elasticIPID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ElasticIPUpdateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ElasticIPUpdateAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	// Marshal the request body to JSON
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpResp, err := c.client.DoRequest(ctx, http.MethodPut, path, bytes.NewReader(bodyBytes), queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.ElasticIPResponse](httpResp, c.client.Logger())
}

// Delete deletes an elastic IP by ID
func (c *elasticIPsClientImpl) Delete(ctx context.Context, projectID string, elasticIPID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting elastic IP: %s in project: %s", elasticIPID, projectID)

	if err := types.ValidateProjectAndResource(projectID, elasticIPID, "elastic IP ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(ElasticIPPath, projectID, elasticIPID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ElasticIPDeleteAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ElasticIPDeleteAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodDelete, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[any](httpResp, c.client.Logger())
}
