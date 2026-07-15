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

type vpcPeeringRoutesClientImpl struct {
	client *restclient.Client
}

// NewService creates a new unified Network service
func NewVPCPeeringRoutesClientImpl(client *restclient.Client) *vpcPeeringRoutesClientImpl {
	return &vpcPeeringRoutesClientImpl{
		client: client,
	}
}

// List retrieves all VPC peering routes for a VPC peering connection
func (c *vpcPeeringRoutesClientImpl) List(ctx context.Context, projectID string, vpcID string, vpcPeeringID string, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteListResponse], error) {
	c.client.Logger().Debugf("Listing VPC peering routes for VPC peering: %s in VPC: %s in project: %s", vpcPeeringID, vpcID, projectID)

	if err := types.ValidateVPCResource(projectID, vpcID, vpcPeeringID, "VPC peering ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCPeeringRoutesPath, projectID, vpcID, vpcPeeringID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCPeeringRouteListAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCPeeringRouteListAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.VPCPeeringRouteListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific VPC peering route by ID
func (c *vpcPeeringRoutesClientImpl) Get(ctx context.Context, projectID string, vpcID string, vpcPeeringID string, vpcPeeringRouteID string, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
	c.client.Logger().Debugf("Getting VPC peering route: %s from VPC peering: %s in VPC: %s in project: %s", vpcPeeringRouteID, vpcPeeringID, vpcID, projectID)

	if err := types.ValidateVPCPeeringRoute(projectID, vpcID, vpcPeeringID, vpcPeeringRouteID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCPeeringRoutePath, projectID, vpcID, vpcPeeringID, vpcPeeringRouteID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCPeeringRouteGetAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCPeeringRouteGetAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.VPCPeeringRouteResponse](httpResp, c.client.Logger())
}

// Create creates a new VPC peering route
func (c *vpcPeeringRoutesClientImpl) Create(ctx context.Context, projectID string, vpcID string, vpcPeeringID string, body types.VPCPeeringRouteRequest, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
	c.client.Logger().Debugf("Creating VPC peering route in VPC peering: %s in VPC: %s in project: %s", vpcPeeringID, vpcID, projectID)

	if err := types.ValidateVPCResource(projectID, vpcID, vpcPeeringID, "VPC peering ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCPeeringRoutesPath, projectID, vpcID, vpcPeeringID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCPeeringRouteCreateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCPeeringRouteCreateAPIVersion
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

	response := &types.Response[types.VPCPeeringRouteResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.VPCPeeringRouteResponse
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

// Update updates an existing VPC peering route
func (c *vpcPeeringRoutesClientImpl) Update(ctx context.Context, projectID string, vpcID string, vpcPeeringID string, vpcPeeringRouteID string, body types.VPCPeeringRouteRequest, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
	c.client.Logger().Debugf("Updating VPC peering route: %s in VPC peering: %s in VPC: %s in project: %s", vpcPeeringRouteID, vpcPeeringID, vpcID, projectID)

	if err := types.ValidateVPCPeeringRoute(projectID, vpcID, vpcPeeringID, vpcPeeringRouteID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCPeeringRoutePath, projectID, vpcID, vpcPeeringID, vpcPeeringRouteID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCPeeringRouteUpdateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCPeeringRouteUpdateAPIVersion
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

	response := &types.Response[types.VPCPeeringRouteResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.VPCPeeringRouteResponse
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

// Delete deletes a VPC peering route by ID
func (c *vpcPeeringRoutesClientImpl) Delete(ctx context.Context, projectID string, vpcID string, vpcPeeringID string, vpcPeeringRouteID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting VPC peering route: %s from VPC peering: %s in VPC: %s in project: %s", vpcPeeringRouteID, vpcPeeringID, vpcID, projectID)

	if err := types.ValidateVPCPeeringRoute(projectID, vpcID, vpcPeeringID, vpcPeeringRouteID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCPeeringRoutePath, projectID, vpcID, vpcPeeringID, vpcPeeringRouteID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCPeeringRouteDeleteAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCPeeringRouteDeleteAPIVersion
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
