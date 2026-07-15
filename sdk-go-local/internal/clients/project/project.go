package project

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

// projectsClientImpl implements the ProjectAPI interface for all Project operations
type projectsClientImpl struct {
	client *restclient.Client
}

// NewProjectsClientImpl creates a new unified Project service
func NewProjectsClientImpl(client *restclient.Client) *projectsClientImpl {
	return &projectsClientImpl{
		client: client,
	}
}

// List retrieves all projects
func (c *projectsClientImpl) List(ctx context.Context, params *types.RequestParameters) (*types.Response[types.ProjectListResponse], error) {
	c.client.Logger().Debugf("Listing projects")

	path := ProjectsPath

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ProjectListAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ProjectListAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.ProjectListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific project by ID
func (c *projectsClientImpl) Get(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.ProjectResponse], error) {
	c.client.Logger().Debugf("Getting project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(ProjectPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ProjectGetAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ProjectGetAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.ProjectResponse](httpResp, c.client.Logger())
}

// Create creates a new project
func (c *projectsClientImpl) Create(ctx context.Context, body types.ProjectRequest, params *types.RequestParameters) (*types.Response[types.ProjectResponse], error) {
	c.client.Logger().Debugf("Creating project")

	path := ProjectsPath

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ProjectCreateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ProjectCreateAPIVersion
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

	response := &types.Response[types.ProjectResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.ProjectResponse
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

// Update updates an existing project
func (c *projectsClientImpl) Update(ctx context.Context, projectID string, body types.ProjectRequest, params *types.RequestParameters) (*types.Response[types.ProjectResponse], error) {
	c.client.Logger().Debugf("Updating project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(ProjectPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ProjectUpdateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ProjectUpdateAPIVersion
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

	response := &types.Response[types.ProjectResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.ProjectResponse
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

// Delete deletes a project by ID
func (c *projectsClientImpl) Delete(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(ProjectPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ProjectDeleteAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ProjectDeleteAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodDelete, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	response := &types.Response[any]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      bodyBytes,
	}

	if response.IsError() && len(bodyBytes) > 0 {
		var errorResp types.ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResp); err == nil {
			response.Error = &errorResp
		}
	}

	return response, nil
}
