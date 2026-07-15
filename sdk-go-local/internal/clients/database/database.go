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

type databasesClientImpl struct {
	client *restclient.Client
}

// NewService creates a new unified Database service
func NewDatabasesClientImpl(client *restclient.Client) *databasesClientImpl {
	return &databasesClientImpl{
		client: client,
	}
}

// List retrieves all databases for a DBaaS instance
func (c *databasesClientImpl) List(ctx context.Context, projectID string, dbaasID string, params *types.RequestParameters) (*types.Response[types.DatabaseListResponse], error) {
	c.client.Logger().Debugf("Listing databases for DBaaS: %s in project: %s", dbaasID, projectID)

	if err := types.ValidateProjectAndResource(projectID, dbaasID, "DBaaS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(DatabaseInstancesPath, projectID, dbaasID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseInstanceListVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseInstanceListVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.DatabaseListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific database by ID
func (c *databasesClientImpl) Get(ctx context.Context, projectID string, dbaasID string, databaseID string, params *types.RequestParameters) (*types.Response[types.DatabaseResponse], error) {
	c.client.Logger().Debugf("Getting database: %s from DBaaS: %s in project: %s", databaseID, dbaasID, projectID)

	if err := types.ValidateDBaaSResource(projectID, dbaasID, databaseID, "database ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(DatabaseInstancePath, projectID, dbaasID, databaseID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseInstanceGetVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseInstanceGetVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.DatabaseResponse](httpResp, c.client.Logger())
}

// Create creates a new database
func (c *databasesClientImpl) Create(ctx context.Context, projectID string, dbaasID string, body types.DatabaseRequest, params *types.RequestParameters) (*types.Response[types.DatabaseResponse], error) {
	c.client.Logger().Debugf("Creating database in DBaaS: %s in project: %s", dbaasID, projectID)

	if err := types.ValidateProjectAndResource(projectID, dbaasID, "DBaaS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(DatabaseInstancesPath, projectID, dbaasID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseInstanceCreateVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseInstanceCreateVersion
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
	response := &types.Response[types.DatabaseResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	// Parse the response body if successful
	if response.IsSuccess() {
		var data types.DatabaseResponse
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

// Update updates an existing database
func (c *databasesClientImpl) Update(ctx context.Context, projectID string, dbaasID string, databaseID string, body types.DatabaseRequest, params *types.RequestParameters) (*types.Response[types.DatabaseResponse], error) {
	c.client.Logger().Debugf("Updating database: %s in DBaaS: %s in project: %s", databaseID, dbaasID, projectID)

	if err := types.ValidateDBaaSResource(projectID, dbaasID, databaseID, "database ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(DatabaseInstancePath, projectID, dbaasID, databaseID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseInstanceUpdateVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseInstanceUpdateVersion
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
	response := &types.Response[types.DatabaseResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	// Parse the response body if successful
	if response.IsSuccess() {
		var data types.DatabaseResponse
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

// Delete deletes a database by ID
func (c *databasesClientImpl) Delete(ctx context.Context, projectID string, dbaasID string, databaseID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting database: %s from DBaaS: %s in project: %s", databaseID, dbaasID, projectID)

	if err := types.ValidateDBaaSResource(projectID, dbaasID, databaseID, "database ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(DatabaseInstancePath, projectID, dbaasID, databaseID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseInstanceDeleteVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseInstanceDeleteVersion
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
