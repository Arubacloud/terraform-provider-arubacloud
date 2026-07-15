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

type cloudServersClientImpl struct {
	client *restclient.Client
}

func NewCloudServersClientImpl(client *restclient.Client) *cloudServersClientImpl {
	return &cloudServersClientImpl{
		client: client,
	}
}

// List retrieves all cloud servers for a project
func (c *cloudServersClientImpl) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.CloudServerListResponse], error) {
	c.client.Logger().Debugf("Listing cloud servers for project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(CloudServersPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeCloudServerList,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeCloudServerList
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.CloudServerListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific cloud server by ID
func (c *cloudServersClientImpl) Get(ctx context.Context, projectID string, cloudServerID string, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	c.client.Logger().Debugf("Getting cloud server: %s in project: %s", cloudServerID, projectID)

	if err := types.ValidateProjectAndResource(projectID, cloudServerID, "cloud server ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(CloudServerPath, projectID, cloudServerID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeCloudServerGet,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeCloudServerGet
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.CloudServerResponse](httpResp, c.client.Logger())
}

// Create creates a new cloud server
func (c *cloudServersClientImpl) Create(ctx context.Context, projectID string, body types.CloudServerRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	c.client.Logger().Debugf("Creating cloud server in project: %s", projectID)

	if err := types.ValidateProject(projectID); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(CloudServersPath, projectID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeCloudServerCreate,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeCloudServerCreate
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
	response := &types.Response[types.CloudServerResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	// Parse the response body if successful
	if response.IsSuccess() {
		var data types.CloudServerResponse
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

// Update updates an existing cloud server
func (c *cloudServersClientImpl) Update(ctx context.Context, projectID string, cloudServerID string, body types.CloudServerRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	c.client.Logger().Debugf("Updating cloud server: %s in project: %s", cloudServerID, projectID)

	if err := types.ValidateProjectAndResource(projectID, cloudServerID, "cloud server ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(CloudServerPath, projectID, cloudServerID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeCloudServerUpdate,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeCloudServerUpdate
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
	response := &types.Response[types.CloudServerResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	// Parse the response body if successful
	if response.IsSuccess() {
		var data types.CloudServerResponse
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

// Delete deletes a cloud server by ID
func (c *cloudServersClientImpl) Delete(ctx context.Context, projectID string, cloudServerID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting cloud server: %s in project: %s", cloudServerID, projectID)

	if err := types.ValidateProjectAndResource(projectID, cloudServerID, "cloud server ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(CloudServerPath, projectID, cloudServerID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeCloudServerDelete,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeCloudServerDelete
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

// PowerOn powers on a cloud server
func (c *cloudServersClientImpl) PowerOn(ctx context.Context, projectID string, cloudServerID string, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	c.client.Logger().Debugf("Powering on cloud server: %s in project: %s", cloudServerID, projectID)

	if err := types.ValidateProjectAndResource(projectID, cloudServerID, "cloud server ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(CloudServerPowerOnPath, projectID, cloudServerID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeCloudServerPowerOn,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeCloudServerPowerOn
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodPost, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.CloudServerResponse](httpResp, c.client.Logger())
}

// PowerOff powers off a cloud server
func (c *cloudServersClientImpl) PowerOff(ctx context.Context, projectID string, cloudServerID string, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	c.client.Logger().Debugf("Powering off cloud server: %s in project: %s", cloudServerID, projectID)

	if err := types.ValidateProjectAndResource(projectID, cloudServerID, "cloud server ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(CloudServerPowerOffPath, projectID, cloudServerID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeCloudServerPowerOff,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeCloudServerPowerOff
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodPost, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.CloudServerResponse](httpResp, c.client.Logger())
}

// AssociateSubnets associates and/or disassociates subnets on a cloud server.
func (c *cloudServersClientImpl) AssociateSubnets(ctx context.Context, projectID string, cloudServerID string, body types.CloudServerAssociateSubnetsRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	c.client.Logger().Debugf("Managing subnets for cloud server: %s in project: %s", cloudServerID, projectID)

	if err := types.ValidateProjectAndResource(projectID, cloudServerID, "cloud server ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(CloudServerAssociateSubnetsPath, projectID, cloudServerID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeCloudServerAssociateSubnets,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeCloudServerAssociateSubnets
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

	return types.ParseResponseBody[types.CloudServerResponse](httpResp, c.client.Logger())
}

// AssociateSecurityGroups associates and/or disassociates security groups on a cloud server.
func (c *cloudServersClientImpl) AssociateSecurityGroups(ctx context.Context, projectID string, cloudServerID string, body types.CloudServerAssociateSecurityGroupsRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	c.client.Logger().Debugf("Managing security groups for cloud server: %s in project: %s", cloudServerID, projectID)

	if err := types.ValidateProjectAndResource(projectID, cloudServerID, "cloud server ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(CloudServerAssociateSecurityGroupsPath, projectID, cloudServerID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeCloudServerAssociateSecurityGroups,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeCloudServerAssociateSecurityGroups
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

	return types.ParseResponseBody[types.CloudServerResponse](httpResp, c.client.Logger())
}

// AssociateElasticIPs associates and/or disassociates elastic IPs on a cloud server.
func (c *cloudServersClientImpl) AssociateElasticIPs(ctx context.Context, projectID string, cloudServerID string, body types.CloudServerAssociateElasticIPsRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	c.client.Logger().Debugf("Managing elastic IPs for cloud server: %s in project: %s", cloudServerID, projectID)

	if err := types.ValidateProjectAndResource(projectID, cloudServerID, "cloud server ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(CloudServerAssociateElasticIPsPath, projectID, cloudServerID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeCloudServerAssociateElasticIPs,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeCloudServerAssociateElasticIPs
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

	return types.ParseResponseBody[types.CloudServerResponse](httpResp, c.client.Logger())
}

// AttachDetachDataVolumes attaches and/or detaches data volumes on a cloud server.
func (c *cloudServersClientImpl) AttachDetachDataVolumes(ctx context.Context, projectID string, cloudServerID string, body types.CloudServerAttachDetachDataVolumesRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	c.client.Logger().Debugf("Managing data volumes for cloud server: %s in project: %s", cloudServerID, projectID)

	if err := types.ValidateProjectAndResource(projectID, cloudServerID, "cloud server ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(CloudServerAttachDetachDataVolumesPath, projectID, cloudServerID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeCloudServerAttachDetachDataVolumes,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeCloudServerAttachDetachDataVolumes
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

	return types.ParseResponseBody[types.CloudServerResponse](httpResp, c.client.Logger())
}

// SetPassword sets or changes the password for a cloud server
func (c *cloudServersClientImpl) SetPassword(ctx context.Context, projectID string, cloudServerID string, body types.CloudServerPasswordRequest, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Setting password for cloud server: %s in project: %s", cloudServerID, projectID)

	if err := types.ValidateProjectAndResource(projectID, cloudServerID, "cloud server ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(CloudServerPasswordPath, projectID, cloudServerID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &ComputeCloudServerPassword,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &ComputeCloudServerPassword
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

	return types.ParseResponseBody[any](httpResp, c.client.Logger())
}
