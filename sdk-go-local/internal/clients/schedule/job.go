package schedule

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

type jobsClientImpl struct {
	client *restclient.Client
}

// NewJobsClientImpl creates a new unified Schedule service
func NewJobsClientImpl(client *restclient.Client) *jobsClientImpl {
	return &jobsClientImpl{
		client: client,
	}
}

// List retrieves all schedule jobs for a project
func (c *jobsClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.JobListResponse], error) {
	c.client.Logger().Debugf("Listing schedule jobs for project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(JobsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ScheduleJobListAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ScheduleJobListAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.JobListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific schedule job by ID
func (c *jobsClientImpl) Get(ctx context.Context, projectID string, scheduleJobID string, params *types.RequestParameters) (*types.Response[types.JobResponse], error) {
	c.client.Logger().Debugf("Getting schedule job: %s in project: %s", scheduleJobID, projectID)

	if err := types.ValidateProjectAndResource(projectID, scheduleJobID, "job ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(JobPath, projectID, scheduleJobID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ScheduleJobGetAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ScheduleJobGetAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.JobResponse](httpResp, c.client.Logger())
}

// Create creates a new schedule job
func (c *jobsClientImpl) Create(ctx context.Context, projectID string, body types.JobRequest, params *types.RequestParameters) (*types.Response[types.JobResponse], error) {
	c.client.Logger().Debugf("Creating schedule job in project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(JobsPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ScheduleJobCreateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ScheduleJobCreateAPIVersion
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

	response := &types.Response[types.JobResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.JobResponse
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

// Update updates an existing schedule job
func (c *jobsClientImpl) Update(ctx context.Context, projectID string, scheduleJobID string, body types.JobRequest, params *types.RequestParameters) (*types.Response[types.JobResponse], error) {
	c.client.Logger().Debugf("Updating schedule job: %s in project: %s", scheduleJobID, projectID)

	if err := types.ValidateProjectAndResource(projectID, scheduleJobID, "job ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(JobPath, projectID, scheduleJobID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ScheduleJobUpdateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ScheduleJobUpdateAPIVersion
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

	response := &types.Response[types.JobResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.JobResponse
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

// Delete deletes a schedule job by ID
func (c *jobsClientImpl) Delete(ctx context.Context, projectID string, scheduleJobID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting schedule job: %s in project: %s", scheduleJobID, projectID)

	if err := types.ValidateProjectAndResource(projectID, scheduleJobID, "job ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(JobPath, projectID, scheduleJobID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ScheduleJobDeleteAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ScheduleJobDeleteAPIVersion
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
