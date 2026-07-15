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

// Key Client (nested under KMS)
type KeyClientImpl struct {
	client *restclient.Client
}

// NewKeyClientImpl creates a new Key client
func NewKeyClientImpl(client *restclient.Client) *KeyClientImpl {
	return &KeyClientImpl{
		client: client,
	}
}

// List retrieves all Keys for a specific KMS instance
func (c *KeyClientImpl) List(ctx context.Context, projectID string, kmsID string, params *types.RequestParameters) (*types.Response[types.KeyListResponse], error) {
	c.client.Logger().Debugf("Listing Keys for KMS: %s in project: %s", kmsID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kmsID, "KMS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KeysPath, projectID, kmsID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &KeyListAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &KeyListAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.KeyListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific Key by ID
func (c *KeyClientImpl) Get(ctx context.Context, projectID string, kmsID string, keyID string, params *types.RequestParameters) (*types.Response[types.KeyResponse], error) {
	c.client.Logger().Debugf("Getting Key: %s for KMS: %s in project: %s", keyID, kmsID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kmsID, "KMS ID"); err != nil {
		return nil, err
	}

	if keyID == "" {
		return nil, fmt.Errorf("key ID cannot be empty")
	}

	path := fmt.Sprintf(KeyPath, projectID, kmsID, keyID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &KeyReadAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &KeyReadAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.KeyResponse](httpResp, c.client.Logger())
}

// Create creates a new Key for a KMS instance
func (c *KeyClientImpl) Create(ctx context.Context, projectID string, kmsID string, body types.KeyRequest, params *types.RequestParameters) (*types.Response[types.KeyResponse], error) {
	c.client.Logger().Debugf("Creating Key for KMS: %s in project: %s", kmsID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kmsID, "KMS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KeysPath, projectID, kmsID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &KeyCreateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &KeyCreateAPIVersion
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

	response := &types.Response[types.KeyResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.KeyResponse
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

// Delete deletes a Key by ID
func (c *KeyClientImpl) Delete(ctx context.Context, projectID string, kmsID string, keyID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting Key: %s for KMS: %s in project: %s", keyID, kmsID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kmsID, "KMS ID"); err != nil {
		return nil, err
	}

	if keyID == "" {
		return nil, fmt.Errorf("key ID cannot be empty")
	}

	path := fmt.Sprintf(KeyPath, projectID, kmsID, keyID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &KeyDeleteAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &KeyDeleteAPIVersion
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
