package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// snapshotsClientImpl implements the StorageAPI interface for all Storage operations
type snapshotsClientImpl struct {
	client        *restclient.Client
	volumesClient *volumesClientImpl
}

// Update updates an existing snapshot
func (c *snapshotsClientImpl) Update(ctx context.Context, projectID string, snapshotID string, body types.SnapshotRequest, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
	c.client.Logger().Debugf("Updating snapshot: %s in project: %s", snapshotID, projectID)

	if err := types.ValidateProjectAndResource(projectID, snapshotID, "snapshot ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(SnapshotPath, projectID, snapshotID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &SnapshotUpdateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &SnapshotUpdateAPIVersion
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

	return types.ParseResponseBody[types.SnapshotResponse](httpResp, c.client.Logger())
}

// NewSnapshotsClientImpl creates a new unified Storage service
func NewSnapshotsClientImpl(client *restclient.Client, volumesClient *volumesClientImpl) *snapshotsClientImpl {
	if volumesClient == nil {
		panic("volumesClient is required and cannot be nil")
	}
	return &snapshotsClientImpl{
		client:        client,
		volumesClient: volumesClient,
	}
}

// List retrieves all snapshots for a project
func (c *snapshotsClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.SnapshotListResponse], error) {
	c.client.Logger().Debugf("Listing snapshots for project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(SnapshotsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &SnapshotListAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &SnapshotListAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.SnapshotListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific snapshot by ID
func (c *snapshotsClientImpl) Get(ctx context.Context, projectID string, snapshotID string, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
	c.client.Logger().Debugf("Getting snapshot: %s in project: %s", snapshotID, projectID)

	if err := types.ValidateProjectAndResource(projectID, snapshotID, "snapshot ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(SnapshotPath, projectID, snapshotID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &SnapshotGetAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &SnapshotGetAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.SnapshotResponse](httpResp, c.client.Logger())
}

// Create creates a new snapshot
func (c *snapshotsClientImpl) Create(ctx context.Context, projectID string, body types.SnapshotRequest, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
	c.client.Logger().Debugf("Creating snapshot in project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(SnapshotsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &SnapshotCreateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &SnapshotCreateAPIVersion
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

	resp, err := types.ParseResponseBody[types.SnapshotResponse](httpResp, c.client.Logger())
	if err != nil {
		return nil, err
	}
	if resp.IsSuccess() && resp.Data != nil {
		if err := resp.Data.Metadata.Validate(); err != nil {
			return resp, fmt.Errorf("invalid create response: %w", err)
		}
	}
	return resp, nil
}

// Delete deletes a snapshot by ID
func (c *snapshotsClientImpl) Delete(ctx context.Context, projectID string, snapshotID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting snapshot: %s in project: %s", snapshotID, projectID)

	if err := types.ValidateProjectAndResource(projectID, snapshotID, "snapshot ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(SnapshotPath, projectID, snapshotID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &SnapshotDeleteAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &SnapshotDeleteAPIVersion
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
