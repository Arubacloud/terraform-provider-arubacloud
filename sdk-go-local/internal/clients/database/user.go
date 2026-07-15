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

type usersClientImpl struct {
	client *restclient.Client
}

// NewService creates a new unified Database service
func NewUsersClientImpl(client *restclient.Client) *usersClientImpl {
	return &usersClientImpl{
		client: client,
	}
}

// List retrieves all users for a DBaaS instance
func (c *usersClientImpl) List(ctx context.Context, projectID string, dbaasID string, params *types.RequestParameters) (*types.Response[types.DatabaseUserListResponse], error) {
	c.client.Logger().Debugf("Listing users for DBaaS: %s in project: %s", dbaasID, projectID)

	if err := types.ValidateProjectAndResource(projectID, dbaasID, "DBaaS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(UsersPath, projectID, dbaasID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseUserListVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseUserListVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.DatabaseUserListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific user by ID
func (c *usersClientImpl) Get(ctx context.Context, projectID string, dbaasID string, userID string, params *types.RequestParameters) (*types.Response[types.UserResponse], error) {
	c.client.Logger().Debugf("Getting user: %s from DBaaS: %s in project: %s", userID, dbaasID, projectID)

	if err := types.ValidateDBaaSResource(projectID, dbaasID, userID, "user ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(UserItemPath, projectID, dbaasID, userID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseUserGetVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseUserGetVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.UserResponse](httpResp, c.client.Logger())
}

// Create creates a new user in a DBaaS instance
func (c *usersClientImpl) Create(ctx context.Context, projectID string, dbaasID string, body types.UserRequest, params *types.RequestParameters) (*types.Response[types.UserResponse], error) {
	c.client.Logger().Debugf("Creating user in DBaaS: %s in project: %s", dbaasID, projectID)

	if err := types.ValidateProjectAndResource(projectID, dbaasID, "DBaaS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(UsersPath, projectID, dbaasID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseUserCreateVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseUserCreateVersion
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
	response := &types.Response[types.UserResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	// Parse the response body if successful
	if response.IsSuccess() {
		var data types.UserResponse
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

// Update updates an existing user
func (c *usersClientImpl) Update(ctx context.Context, projectID string, dbaasID string, userID string, body types.UserRequest, params *types.RequestParameters) (*types.Response[types.UserResponse], error) {
	c.client.Logger().Debugf("Updating user: %s in DBaaS: %s in project: %s", userID, dbaasID, projectID)

	if err := types.ValidateDBaaSResource(projectID, dbaasID, userID, "user ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(UserItemPath, projectID, dbaasID, userID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseUserUpdateVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseUserUpdateVersion
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
	response := &types.Response[types.UserResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	// Parse the response body if successful
	if response.IsSuccess() {
		var data types.UserResponse
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

// Delete deletes a user by ID
func (c *usersClientImpl) Delete(ctx context.Context, projectID string, dbaasID string, userID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting user: %s from DBaaS: %s in project: %s", userID, dbaasID, projectID)

	if err := types.ValidateDBaaSResource(projectID, dbaasID, userID, "user ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(UserItemPath, projectID, dbaasID, userID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &DatabaseUserDeleteVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &DatabaseUserDeleteVersion
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
