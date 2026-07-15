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

type vpnTunnelsClientImpl struct {
	client *restclient.Client
}

// NewService creates a new unified Network service
func NewVPNTunnelsClientImpl(client *restclient.Client) *vpnTunnelsClientImpl {
	return &vpnTunnelsClientImpl{
		client: client,
	}
}

// List retrieves all VPN tunnels for a project
func (c *vpnTunnelsClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.VPNTunnelListResponse], error) {
	c.client.Logger().Debugf("Listing VPN tunnels for project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPNTunnelsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPNTunnelListAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPNTunnelListAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.VPNTunnelListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific VPN tunnel by ID
func (c *vpnTunnelsClientImpl) Get(ctx context.Context, projectID string, vpnTunnelID string, params *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error) {
	c.client.Logger().Debugf("Getting VPN tunnel: %s in project: %s", vpnTunnelID, projectID)

	if err := types.ValidateProjectAndResource(projectID, vpnTunnelID, "VPN tunnel ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPNTunnelPath, projectID, vpnTunnelID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPNTunnelGetAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPNTunnelGetAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.VPNTunnelResponse](httpResp, c.client.Logger())
}

// Create creates a new VPN tunnel
func (c *vpnTunnelsClientImpl) Create(ctx context.Context, projectID string, body types.VPNTunnelRequest, params *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error) {
	c.client.Logger().Debugf("Creating VPN tunnel in project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPNTunnelsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPNTunnelCreateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPNTunnelCreateAPIVersion
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

	response := &types.Response[types.VPNTunnelResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.VPNTunnelResponse
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

// Update updates an existing VPN tunnel
func (c *vpnTunnelsClientImpl) Update(ctx context.Context, projectID string, vpnTunnelID string, body types.VPNTunnelRequest, params *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error) {
	c.client.Logger().Debugf("Updating VPN tunnel: %s in project: %s", vpnTunnelID, projectID)

	if err := types.ValidateProjectAndResource(projectID, vpnTunnelID, "VPN tunnel ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPNTunnelPath, projectID, vpnTunnelID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPNTunnelUpdateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPNTunnelUpdateAPIVersion
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

	response := &types.Response[types.VPNTunnelResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.VPNTunnelResponse
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

// Delete deletes a VPN tunnel by ID
func (c *vpnTunnelsClientImpl) Delete(ctx context.Context, projectID string, vpnTunnelID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting VPN tunnel: %s in project: %s", vpnTunnelID, projectID)

	if err := types.ValidateProjectAndResource(projectID, vpnTunnelID, "VPN tunnel ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(VPNTunnelPath, projectID, vpnTunnelID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &VPNTunnelDeleteAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &VPNTunnelDeleteAPIVersion
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
