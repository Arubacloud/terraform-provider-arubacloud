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

type subnetsClientImpl struct {
	client    *restclient.Client
	vpcClient *vpcsClientImpl
}

// NewService creates a new unified Network service
func NewSubnetsClientImpl(client *restclient.Client, vpcClient *vpcsClientImpl) *subnetsClientImpl {
	if vpcClient == nil {
		panic("vpcClient is required and cannot be nil")
	}
	return &subnetsClientImpl{
		client:    client,
		vpcClient: vpcClient,
	}
}

// List retrieves all subnets for a VPC
func (c *subnetsClientImpl) List(ctx context.Context, projectID string, vpcID string, params *types.RequestParameters) (*types.Response[types.SubnetListResponse], error) {
	c.client.Logger().Debugf("Listing subnets for VPC: %s in project: %s", vpcID, projectID)

	if err := types.ValidateProjectAndResource(projectID, vpcID, "VPC ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(SubnetsPath, projectID, vpcID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &SubnetListAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &SubnetListAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.SubnetListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific subnet by ID
func (c *subnetsClientImpl) Get(ctx context.Context, projectID string, vpcID string, subnetID string, params *types.RequestParameters) (*types.Response[types.SubnetResponse], error) {
	c.client.Logger().Debugf("Getting subnet: %s from VPC: %s in project: %s", subnetID, vpcID, projectID)

	if err := types.ValidateVPCResource(projectID, vpcID, subnetID, "subnet ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(SubnetPath, projectID, vpcID, subnetID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &SubnetGetAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &SubnetGetAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.SubnetResponse](httpResp, c.client.Logger())
}

// Create creates a new subnet in a VPC
func (c *subnetsClientImpl) Create(ctx context.Context, projectID string, vpcID string, body types.SubnetRequest, params *types.RequestParameters) (*types.Response[types.SubnetResponse], error) {
	c.client.Logger().Debugf("Creating subnet in VPC: %s in project: %s", vpcID, projectID)

	if err := types.ValidateProjectAndResource(projectID, vpcID, "VPC ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(SubnetsPath, projectID, vpcID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &SubnetCreateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &SubnetCreateAPIVersion
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

	response := &types.Response[types.SubnetResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.SubnetResponse
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

// Update updates an existing subnet
func (c *subnetsClientImpl) Update(ctx context.Context, projectID string, vpcID string, subnetID string, body types.SubnetRequest, params *types.RequestParameters) (*types.Response[types.SubnetResponse], error) {
	c.client.Logger().Debugf("Updating subnet: %s in VPC: %s in project: %s", subnetID, vpcID, projectID)

	if err := types.ValidateVPCResource(projectID, vpcID, subnetID, "subnet ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(SubnetPath, projectID, vpcID, subnetID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &SubnetUpdateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &SubnetUpdateAPIVersion
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

	response := &types.Response[types.SubnetResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.SubnetResponse
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

// Delete deletes a subnet by ID
func (c *subnetsClientImpl) Delete(ctx context.Context, projectID string, vpcID string, subnetID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting subnet: %s from VPC: %s in project: %s", subnetID, vpcID, projectID)

	if err := types.ValidateVPCResource(projectID, vpcID, subnetID, "subnet ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(SubnetPath, projectID, vpcID, subnetID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &SubnetDeleteAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &SubnetDeleteAPIVersion
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
