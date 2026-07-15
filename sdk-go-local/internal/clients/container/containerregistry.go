package container

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

// containerRegistryClientImpl implements the ContainerRegistryAPI interface for all Container Registry operations
type containerRegistryClientImpl struct {
	client *restclient.Client
}

// NewContainerRegistryClientImpl creates a new unified Container Registry service
func NewContainerRegistryClientImpl(client *restclient.Client) *containerRegistryClientImpl {
	return &containerRegistryClientImpl{
		client: client,
	}
}

// List retrieves all Container Registries for a project
func (c *containerRegistryClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.ContainerRegistryListResponse], error) {
	c.client.Logger().Debugf("Listing Container Registries for project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(ContainerRegistryPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ContainerRegistryListVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ContainerRegistryListVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.ContainerRegistryListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific Container Registry by ID
func (c *containerRegistryClientImpl) Get(ctx context.Context, projectID string, registryID string, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
	c.client.Logger().Debugf("Getting Container Registry: %s in project: %s", registryID, projectID)

	if err := types.ValidateProjectAndResource(projectID, registryID, "Container Registry ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(ContainerRegistryItemPath, projectID, registryID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ContainerRegistryGetVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ContainerRegistryGetVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.ContainerRegistryResponse](httpResp, c.client.Logger())
}

// Create creates a new Container Registry
func (c *containerRegistryClientImpl) Create(ctx context.Context, projectID string, body types.ContainerRegistryRequest, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
	c.client.Logger().Debugf("Creating Container Registry in project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(ContainerRegistryPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ContainerRegistryCreateVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ContainerRegistryCreateVersion
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

	response := &types.Response[types.ContainerRegistryResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.ContainerRegistryResponse
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

// Update updates an existing Container Registry
func (c *containerRegistryClientImpl) Update(ctx context.Context, projectID string, registryID string, body types.ContainerRegistryRequest, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
	c.client.Logger().Debugf("Updating Container Registry: %s in project: %s", registryID, projectID)

	if err := types.ValidateProjectAndResource(projectID, registryID, "Container Registry ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(ContainerRegistryItemPath, projectID, registryID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ContainerRegistryUpdateVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ContainerRegistryUpdateVersion
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

	response := &types.Response[types.ContainerRegistryResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.ContainerRegistryResponse
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

// Delete deletes a Container Registry by ID
func (c *containerRegistryClientImpl) Delete(ctx context.Context, projectID string, registryID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting Container Registry: %s in project: %s", registryID, projectID)

	if err := types.ValidateProjectAndResource(projectID, registryID, "Container Registry ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(ContainerRegistryItemPath, projectID, registryID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ContainerRegistryDeleteVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ContainerRegistryDeleteVersion
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
