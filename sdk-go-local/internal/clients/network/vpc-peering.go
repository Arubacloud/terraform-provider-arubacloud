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

type vpcPeeringsClientImpl struct {
	client *restclient.Client
}

// NewService creates a new unified Network service
func NewVPCPeeringsClientImpl(client *restclient.Client) *vpcPeeringsClientImpl {
	return &vpcPeeringsClientImpl{
		client: client,
	}
}

// List retrieves all VPC peerings for a VPC
func (c *vpcPeeringsClientImpl) List(ctx context.Context, projectID string, vpcID string, params *types.RequestParameters) (*types.Response[types.VPCPeeringListResponse], error) {
	c.client.Logger().Debugf("Listing VPC peerings for VPC: %s in project: %s", vpcID, projectID)

	if err := types.ValidateProjectAndResource(projectID, vpcID, "VPC ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCPeeringsPath, projectID, vpcID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCPeeringListAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCPeeringListAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.VPCPeeringListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific VPC peering by ID
func (c *vpcPeeringsClientImpl) Get(ctx context.Context, projectID string, vpcID string, vpcPeeringID string, params *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error) {
	c.client.Logger().Debugf("Getting VPC peering: %s from VPC: %s in project: %s", vpcPeeringID, vpcID, projectID)

	if err := types.ValidateVPCResource(projectID, vpcID, vpcPeeringID, "VPC peering ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCPeeringPath, projectID, vpcID, vpcPeeringID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCPeeringGetAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCPeeringGetAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.VPCPeeringResponse](httpResp, c.client.Logger())
}

// Create creates a new VPC peering
func (c *vpcPeeringsClientImpl) Create(ctx context.Context, projectID string, vpcID string, body types.VPCPeeringRequest, params *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error) {
	c.client.Logger().Debugf("Creating VPC peering in VPC: %s in project: %s", vpcID, projectID)

	if err := types.ValidateProjectAndResource(projectID, vpcID, "VPC ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCPeeringsPath, projectID, vpcID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCPeeringCreateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCPeeringCreateAPIVersion
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

	response := &types.Response[types.VPCPeeringResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.VPCPeeringResponse
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

// Update updates an existing VPC peering
func (c *vpcPeeringsClientImpl) Update(ctx context.Context, projectID string, vpcID string, vpcPeeringID string, body types.VPCPeeringRequest, params *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error) {
	c.client.Logger().Debugf("Updating VPC peering: %s in VPC: %s in project: %s", vpcPeeringID, vpcID, projectID)

	if err := types.ValidateVPCResource(projectID, vpcID, vpcPeeringID, "VPC peering ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCPeeringPath, projectID, vpcID, vpcPeeringID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCPeeringUpdateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCPeeringUpdateAPIVersion
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

	response := &types.Response[types.VPCPeeringResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.VPCPeeringResponse
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

// Delete deletes a VPC peering by ID
func (c *vpcPeeringsClientImpl) Delete(ctx context.Context, projectID string, vpcID string, vpcPeeringID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting VPC peering: %s from VPC: %s in project: %s", vpcPeeringID, vpcID, projectID)

	if err := types.ValidateVPCResource(projectID, vpcID, vpcPeeringID, "VPC peering ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPCPeeringPath, projectID, vpcID, vpcPeeringID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPCPeeringDeleteAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPCPeeringDeleteAPIVersion
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
