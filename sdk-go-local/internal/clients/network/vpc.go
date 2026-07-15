package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

type vpcsClientImpl struct {
	client *restclient.Client
}

// NewService creates a new unified Network service
func NewVPCsClientImpl(client *restclient.Client) *vpcsClientImpl {
	return &vpcsClientImpl{
		client: client,
	}
}

// List retrieves all VPCs for a project
func (c *vpcsClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.VPCListResponse], error) {
	c.client.Logger().Debugf("Listing VPCs for project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCNetworksPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCListAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCListAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.VPCListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific VPC by ID
func (c *vpcsClientImpl) Get(ctx context.Context, projectID string, vpcID string, params *types.RequestParameters) (*types.Response[types.VPCResponse], error) {
	c.client.Logger().Debugf("Getting VPC: %s in project: %s", vpcID, projectID)

	if err := types.ValidateProjectAndResource(projectID, vpcID, "VPC ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCNetworkPath, projectID, vpcID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCGetAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCGetAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.VPCResponse](httpResp, c.client.Logger())
}

// Create creates a new VPC
func (c *vpcsClientImpl) Create(ctx context.Context, projectID string, body types.VPCRequest, params *types.RequestParameters) (*types.Response[types.VPCResponse], error) {
	c.client.Logger().Debugf("Creating VPC in project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCNetworksPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCCreateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCCreateAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpResp, err := c.client.DoRequest(ctx, http.MethodPost, path, bytes.NewReader(bodyBytes), queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	response := &types.Response[types.VPCResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.VPCResponse
		if err := json.Unmarshal(respBytes, &data); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		response.Data = &data
		if err := response.Data.Metadata.Validate(); err != nil {
			return response, fmt.Errorf("invalid create response: %w", err)
		}
	} else if response.IsError() && len(respBytes) > 0 {
		var errorResp types.ErrorResponse
		if err := json.Unmarshal(respBytes, &errorResp); err == nil {
			response.Error = &errorResp
		}
	}

	return response, nil
}

// Update updates an existing VPC
func (c *vpcsClientImpl) Update(ctx context.Context, projectID string, vpcID string, body types.VPCRequest, params *types.RequestParameters) (*types.Response[types.VPCResponse], error) {
	c.client.Logger().Debugf("Updating VPC: %s in project: %s", vpcID, projectID)

	if err := types.ValidateProjectAndResource(projectID, vpcID, "VPC ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCNetworkPath, projectID, vpcID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCUpdateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCUpdateAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpResp, err := c.client.DoRequest(ctx, http.MethodPut, path, bytes.NewReader(bodyBytes), queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	response := &types.Response[types.VPCResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.VPCResponse
		if err := json.Unmarshal(respBytes, &data); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		response.Data = &data
	} else if response.IsError() && len(respBytes) > 0 {
		var errorResp types.ErrorResponse
		if err := json.Unmarshal(respBytes, &errorResp); err == nil {
			response.Error = &errorResp
		}
	}

	return response, nil
}

// Delete deletes a VPC by ID
func (c *vpcsClientImpl) Delete(ctx context.Context, projectID string, vpcID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting VPC: %s in project: %s", vpcID, projectID)

	if err := types.ValidateProjectAndResource(projectID, vpcID, "VPC ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCNetworkPath, projectID, vpcID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCDeleteAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCDeleteAPIVersion
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
