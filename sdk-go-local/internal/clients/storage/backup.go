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

// backupClientImpl implements the StorageBackupAPI interface for all Storage Backup operations
type backupClientImpl struct {
	client *restclient.Client
}

// NewBackupClientImpl creates a new unified Storage Backup service
func NewBackupClientImpl(client *restclient.Client) *backupClientImpl {
	return &backupClientImpl{
		client: client,
	}
}

// List retrieves all Storage Backups for a project
func (c *backupClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.StorageBackupListResponse], error) {
	c.client.Logger().Debugf("Listing Storage Backups for project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(BackupsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &StorageBackupListVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &StorageBackupListVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.StorageBackupListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific Storage Backup by ID
func (c *backupClientImpl) Get(ctx context.Context, projectID string, backupID string, params *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error) {
	c.client.Logger().Debugf("Getting Storage Backup: %s in project: %s", backupID, projectID)

	if err := types.ValidateProjectAndResource(projectID, backupID, "Storage Backup ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(BackupPath, projectID, backupID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &StorageBackupGetVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &StorageBackupGetVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.StorageBackupResponse](httpResp, c.client.Logger())
}

// Create creates a new Storage Backup
func (c *backupClientImpl) Create(ctx context.Context, projectID string, body types.StorageBackupRequest, params *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error) {
	c.client.Logger().Debugf("Creating Storage Backup in project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(BackupsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &StorageBackupCreateVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &StorageBackupCreateVersion
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

	response := &types.Response[types.StorageBackupResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.StorageBackupResponse
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

// Update updates an existing Storage Backup
func (c *backupClientImpl) Update(ctx context.Context, projectID string, backupID string, body types.StorageBackupRequest, params *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error) {
	c.client.Logger().Debugf("Updating Storage Backup: %s in project: %s", backupID, projectID)

	if err := types.ValidateProjectAndResource(projectID, backupID, "Storage Backup ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(BackupPath, projectID, backupID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &StorageBackupUpdateVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &StorageBackupUpdateVersion
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

	response := &types.Response[types.StorageBackupResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.StorageBackupResponse
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

// Delete deletes a Storage Backup by ID
func (c *backupClientImpl) Delete(ctx context.Context, projectID string, backupID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting Storage Backup: %s in project: %s", backupID, projectID)

	if err := types.ValidateProjectAndResource(projectID, backupID, "Storage Backup ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(BackupPath, projectID, backupID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &StorageBackupDeleteVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &StorageBackupDeleteVersion
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
