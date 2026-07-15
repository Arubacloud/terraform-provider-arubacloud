package database

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

type grantsClientImpl struct {
	client *restclient.Client
}

// NewService creates a new unified Database service
func NewGrantsClientImpl(client *restclient.Client) *grantsClientImpl {
	return &grantsClientImpl{
		client: client,
	}
}

// List retrieves all grants for a database
func (c *grantsClientImpl) List(ctx context.Context, projectID string, dbaasID string, databaseID string, params *types.RequestParameters) (*types.Response[types.GrantListResponse], error) {
	c.client.Logger().Debugf("Listing grants for database: %s in DBaaS: %s in project: %s", databaseID, dbaasID, projectID)

	if err := types.ValidateDBaaSResource(projectID, dbaasID, databaseID, "database ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(GrantsPath, projectID, dbaasID, databaseID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseGrantListVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseGrantListVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.GrantListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific grant by ID
func (c *grantsClientImpl) Get(ctx context.Context, projectID string, dbaasID string, databaseID string, grantID string, params *types.RequestParameters) (*types.Response[types.GrantResponse], error) {
	c.client.Logger().Debugf("Getting grant: %s from database: %s in DBaaS: %s in project: %s", grantID, databaseID, dbaasID, projectID)

	if err := types.ValidateDatabaseGrant(projectID, dbaasID, databaseID, grantID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(GrantItemPath, projectID, dbaasID, databaseID, grantID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseGrantGetVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseGrantGetVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.GrantResponse](httpResp, c.client.Logger())
}

// Create creates a new grant for a database
func (c *grantsClientImpl) Create(ctx context.Context, projectID string, dbaasID string, databaseID string, body types.GrantRequest, params *types.RequestParameters) (*types.Response[types.GrantResponse], error) {
	c.client.Logger().Debugf("Creating grant in database: %s in DBaaS: %s in project: %s", databaseID, dbaasID, projectID)

	if err := types.ValidateDBaaSResource(projectID, dbaasID, databaseID, "database ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(GrantsPath, projectID, dbaasID, databaseID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseGrantCreateVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseGrantCreateVersion
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
	response := &types.Response[types.GrantResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	// Parse the response body if successful
	if response.IsSuccess() {
		var data types.GrantResponse
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

// Update updates an existing grant
func (c *grantsClientImpl) Update(ctx context.Context, projectID string, dbaasID string, databaseID string, grantID string, body types.GrantRequest, params *types.RequestParameters) (*types.Response[types.GrantResponse], error) {
	c.client.Logger().Debugf("Updating grant: %s in database: %s in DBaaS: %s in project: %s", grantID, databaseID, dbaasID, projectID)

	if err := types.ValidateDatabaseGrant(projectID, dbaasID, databaseID, grantID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(GrantItemPath, projectID, dbaasID, databaseID, grantID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseGrantUpdateVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseGrantUpdateVersion
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
	response := &types.Response[types.GrantResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	// Parse the response body if successful
	if response.IsSuccess() {
		var data types.GrantResponse
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

// Delete deletes a grant by ID
func (c *grantsClientImpl) Delete(ctx context.Context, projectID string, dbaasID string, databaseID string, grantID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting grant: %s from database: %s in DBaaS: %s in project: %s", grantID, databaseID, dbaasID, projectID)

	if err := types.ValidateDatabaseGrant(projectID, dbaasID, databaseID, grantID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(GrantItemPath, projectID, dbaasID, databaseID, grantID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseGrantDeleteVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseGrantDeleteVersion
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
