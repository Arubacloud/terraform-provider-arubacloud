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

// kaasClientImpl implements the ContainerAPI interface for all Container operations
type kaasClientImpl struct {
	client *restclient.Client
}

// NewKaaSClientImpl creates a new unified Container service
func NewKaaSClientImpl(client *restclient.Client) *kaasClientImpl {
	return &kaasClientImpl{
		client: client,
	}
}

// List retrieves all KaaS clusters for a project
func (c *kaasClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.KaaSListResponse], error) {
	c.client.Logger().Debugf("Listing KaaS clusters for project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KaaSPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ContainerKaaSListVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ContainerKaaSListVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.KaaSListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific KaaS cluster by ID
func (c *kaasClientImpl) Get(ctx context.Context, projectID string, kaasID string, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
	c.client.Logger().Debugf("Getting KaaS cluster: %s in project: %s", kaasID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kaasID, "KaaS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KaaSItemPath, projectID, kaasID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ContainerKaaSGetVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ContainerKaaSGetVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.KaaSResponse](httpResp, c.client.Logger())
}

// Create creates a new KaaS cluster
func (c *kaasClientImpl) Create(ctx context.Context, projectID string, body types.KaaSRequest, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
	c.client.Logger().Debugf("Creating KaaS cluster in project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KaaSPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ContainerKaaSCreateVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ContainerKaaSCreateVersion
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

	// Read the response body
	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Create the response wrapper
	response := &types.Response[types.KaaSResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	// Parse the response body if successful
	if response.IsSuccess() {
		var data types.KaaSResponse
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

// Update updates an existing KaaS cluster
func (c *kaasClientImpl) Update(ctx context.Context, projectID string, kaasID string, body types.KaaSUpdateRequest, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
	c.client.Logger().Debugf("Updating KaaS cluster: %s in project: %s", kaasID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kaasID, "KaaS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KaaSItemPath, projectID, kaasID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ContainerKaaSUpdateVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ContainerKaaSUpdateVersion
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

	// Read the response body
	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Create the response wrapper
	response := &types.Response[types.KaaSResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	// Parse the response body if successful
	if response.IsSuccess() {
		var data types.KaaSResponse
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

// Delete deletes a KaaS cluster by ID
func (c *kaasClientImpl) Delete(ctx context.Context, projectID string, kaasID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting KaaS cluster: %s in project: %s", kaasID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kaasID, "KaaS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KaaSItemPath, projectID, kaasID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ContainerKaaSDeleteVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ContainerKaaSDeleteVersion
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

// DownloadKubeconfig downloads the kubeconfig file for a KaaS cluster
func (c *kaasClientImpl) DownloadKubeconfig(ctx context.Context, projectID string, kaasID string, params *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error) {
	c.client.Logger().Debugf("Downloading kubeconfig for KaaS cluster: %s in project: %s", kaasID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kaasID, "KaaS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KaaSKubeconfigPath, projectID, kaasID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ContainerKaaSKubeconfigVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ContainerKaaSKubeconfigVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.KaaSKubeconfigResponse](httpResp, c.client.Logger())
}
