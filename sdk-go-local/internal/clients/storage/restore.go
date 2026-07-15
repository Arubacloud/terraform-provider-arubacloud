package storage

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

// restoreClientImpl implements the RestoreAPI interface for all Restore operations
type restoreClientImpl struct {
	client       *restclient.Client
	backupClient *backupClientImpl
}

// NewRestoreClientImpl creates a new unified Restore service
func NewRestoreClientImpl(client *restclient.Client, backupClient *backupClientImpl) *restoreClientImpl {
	if backupClient == nil {
		panic("backupClient is required and cannot be nil")
	}
	return &restoreClientImpl{
		client:       client,
		backupClient: backupClient,
	}
}

// List retrieves all Restores for a backup in a project
func (c *restoreClientImpl) List(ctx context.Context, projectID string, backupID string, params *types.RequestParameters) (*types.Response[types.StorageRestoreListResponse], error) {
	c.client.Logger().Debugf("Listing Restores for project: %s, backup: %s", projectID, backupID)

	if err := types.ValidateProjectAndResource(projectID, backupID, "Backup ID"); err != nil {
		return nil, err
	}
	path := fmt.Sprintf(RestoresPath, projectID, backupID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &RestoreListVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &RestoreListVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.StorageRestoreListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific Restore by ID
func (c *restoreClientImpl) Get(ctx context.Context, projectID string, backupID string, restoreID string, params *types.RequestParameters) (*types.Response[types.StorageRestoreResponse], error) {
	c.client.Logger().Debugf("Getting Restore: %s in project: %s, backup: %s", restoreID, projectID, backupID)

	if err := types.ValidateProjectAndResource(projectID, backupID, "Backup ID"); err != nil {
		return nil, err
	}
	if restoreID == "" {
		return nil, fmt.Errorf("restore id cannot be empty")
	}
	path := fmt.Sprintf(RestorePath, projectID, backupID, restoreID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &RestoreGetVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &RestoreGetVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.StorageRestoreResponse](httpResp, c.client.Logger())
}

// Create creates a new Restore
func (c *restoreClientImpl) Create(ctx context.Context, projectID string, backupID string, body types.StorageRestoreRequest, params *types.RequestParameters) (*types.Response[types.StorageRestoreResponse], error) {
	c.client.Logger().Debugf("Creating Restore in project: %s, backup: %s", projectID, backupID)

	if err := types.ValidateStorageRestore(projectID, backupID, nil); err != nil {
		return nil, err
	}

	// Wait for destination volume to become ready before creating restore
	if body.Properties.Target.URI == "" {
		return nil, fmt.Errorf("target cannot be empty")
	}

	path := fmt.Sprintf(RestoresPath, projectID, backupID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &RestoreCreateVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &RestoreCreateVersion
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

	response := &types.Response[types.StorageRestoreResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.StorageRestoreResponse
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

// Update updates an existing Restore
func (c *restoreClientImpl) Update(ctx context.Context, projectID string, backupID string, restoreID string, body types.StorageRestoreRequest, params *types.RequestParameters) (*types.Response[types.StorageRestoreResponse], error) {
	c.client.Logger().Debugf("Updating Restore: %s in project: %s, backup: %s", restoreID, projectID, backupID)

	if err := types.ValidateProjectAndResource(projectID, backupID, "Backup ID"); err != nil {
		return nil, err
	}
	if restoreID == "" {
		return nil, fmt.Errorf("restore ID cannot be empty")
	}
	path := fmt.Sprintf(RestorePath, projectID, backupID, restoreID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &RestoreUpdateVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &RestoreUpdateVersion
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

	response := &types.Response[types.StorageRestoreResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.StorageRestoreResponse
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

// Delete deletes a Restore by ID
func (c *restoreClientImpl) Delete(ctx context.Context, projectID string, backupID string, restoreID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting Restore: %s in project: %s, backup: %s", restoreID, projectID, backupID)

	if err := types.ValidateProjectAndResource(projectID, backupID, "Backup ID"); err != nil {
		return nil, err
	}
	if restoreID == "" {
		return nil, fmt.Errorf("restore ID cannot be empty")
	}
	path := fmt.Sprintf(RestorePath, projectID, backupID, restoreID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &RestoreDeleteVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &RestoreDeleteVersion
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
