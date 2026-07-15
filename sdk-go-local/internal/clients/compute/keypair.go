package compute

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

type keyPairsClientImpl struct {
	client *restclient.Client
}

func NewKeyPairsClientImpl(client *restclient.Client) *keyPairsClientImpl {
	return &keyPairsClientImpl{
		client: client,
	}
}

// List retrieves all key pairs for a project
func (c *keyPairsClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.KeyPairListResponse], error) {
	c.client.Logger().Debugf("Listing key pairs for project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KeyPairsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeKeyPairList,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeKeyPairList
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.KeyPairListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific key pair by ID
func (c *keyPairsClientImpl) Get(ctx context.Context, projectID string, keyPairID string, params *types.RequestParameters) (*types.Response[types.KeyPairResponse], error) {
	c.client.Logger().Debugf("Getting key pair: %s in project: %s", keyPairID, projectID)

	if err := types.ValidateProjectAndResource(projectID, keyPairID, "key pair ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KeyPairPath, projectID, keyPairID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeKeyPairGet,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeKeyPairGet
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.KeyPairResponse](httpResp, c.client.Logger())
}

// Create creates a new key pair
func (c *keyPairsClientImpl) Create(ctx context.Context, projectID string, body types.KeyPairRequest, params *types.RequestParameters) (*types.Response[types.KeyPairResponse], error) {
	c.client.Logger().Debugf("Creating key pair in project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KeyPairsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeKeyPairCreate,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeKeyPairCreate
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
	response := &types.Response[types.KeyPairResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	// Parse the response body if successful
	if response.IsSuccess() {
		var data types.KeyPairResponse
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

// Delete deletes a key pair by ID
func (c *keyPairsClientImpl) Delete(ctx context.Context, projectID string, keyPairID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting key pair: %s in project: %s", keyPairID, projectID)

	if err := types.ValidateProjectAndResource(projectID, keyPairID, "key pair ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KeyPairPath, projectID, keyPairID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeKeyPairDelete,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeKeyPairDelete
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
