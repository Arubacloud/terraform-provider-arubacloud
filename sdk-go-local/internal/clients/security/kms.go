package security

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

// KMS Client
type kmsClientImpl struct {
	client *restclient.Client
}

// NewKMSClientImpl creates a new KMS client
func NewKMSClientImpl(client *restclient.Client) *kmsClientImpl {
	return &kmsClientImpl{
		client: client,
	}
}

// List retrieves all KMS instances for a project
func (c *kmsClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.KmsListResponse], error) {
	c.client.Logger().Debugf("Listing KMS instances for project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KMSsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &KMSListAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &KMSListAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.KmsListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific KMS instance by ID
func (c *kmsClientImpl) Get(ctx context.Context, projectID string, kmsID string, params *types.RequestParameters) (*types.Response[types.KmsResponse], error) {
	c.client.Logger().Debugf("Getting KMS instance: %s in project: %s", kmsID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kmsID, "KMS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KMSPath, projectID, kmsID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &KMSReadAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &KMSReadAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.KmsResponse](httpResp, c.client.Logger())
}

// Create creates a new KMS instance
func (c *kmsClientImpl) Create(ctx context.Context, projectID string, body types.KmsRequest, params *types.RequestParameters) (*types.Response[types.KmsResponse], error) {
	c.client.Logger().Debugf("Creating KMS instance in project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KMSsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &KMSCreateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &KMSCreateAPIVersion
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

	response := &types.Response[types.KmsResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.KmsResponse
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

// Update updates an existing KMS instance
func (c *kmsClientImpl) Update(ctx context.Context, projectID string, kmsID string, body types.KmsRequest, params *types.RequestParameters) (*types.Response[types.KmsResponse], error) {
	c.client.Logger().Debugf("Updating KMS instance: %s in project: %s", kmsID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kmsID, "KMS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KMSPath, projectID, kmsID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &KMSUpdateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &KMSUpdateAPIVersion
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

	response := &types.Response[types.KmsResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.KmsResponse
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

// Delete deletes a KMS instance by ID
func (c *kmsClientImpl) Delete(ctx context.Context, projectID string, kmsID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting KMS instance: %s in project: %s", kmsID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kmsID, "KMS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KMSPath, projectID, kmsID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &KMSDeleteAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &KMSDeleteAPIVersion
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
