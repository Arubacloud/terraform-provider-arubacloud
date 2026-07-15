package security

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

// KMIP Client (nested under KMS)
type KmipClientImpl struct {
	client *restclient.Client
}

// NewKmipClientImpl creates a new KMIP client
func NewKmipClientImpl(client *restclient.Client) *KmipClientImpl {
	return &KmipClientImpl{
		client: client,
	}
}

// List retrieves all KMIP services for a specific KMS instance
func (c *KmipClientImpl) List(ctx context.Context, projectID string, kmsID string, params *types.RequestParameters) (*types.Response[types.KmipListResponse], error) {
	c.client.Logger().Debugf("Listing KMIP services for KMS: %s in project: %s", kmsID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kmsID, "KMS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KmipsPath, projectID, kmsID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &KmipListAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &KmipListAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.KmipListResponse](httpResp, c.client.Logger())
}

// Get retrieves a specific KMIP service by ID
func (c *KmipClientImpl) Get(ctx context.Context, projectID string, kmsID string, kmipID string, params *types.RequestParameters) (*types.Response[types.KmipResponse], error) {
	c.client.Logger().Debugf("Getting KMIP service: %s for KMS: %s in project: %s", kmipID, kmsID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kmsID, "KMS ID"); err != nil {
		return nil, err
	}

	if kmipID == "" {
		return nil, fmt.Errorf("KMIP ID cannot be empty")
	}

	path := fmt.Sprintf(KmipPath, projectID, kmsID, kmipID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &KmipReadAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &KmipReadAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.KmipResponse](httpResp, c.client.Logger())
}

// Create creates a new KMIP service for a KMS instance
func (c *KmipClientImpl) Create(ctx context.Context, projectID string, kmsID string, body types.KmipRequest, params *types.RequestParameters) (*types.Response[types.KmipResponse], error) {
	c.client.Logger().Debugf("Creating KMIP service for KMS: %s in project: %s", kmsID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kmsID, "KMS ID"); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(KmipsPath, projectID, kmsID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &KmipCreateAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &KmipCreateAPIVersion
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

	response := &types.Response[types.KmipResponse]{
		HTTPResponse: httpResp,
		StatusCode:   httpResp.StatusCode,
		Headers:      httpResp.Header,
		RawBody:      respBytes,
	}

	if response.IsSuccess() {
		var data types.KmipResponse
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

// Delete deletes a KMIP service by ID
func (c *KmipClientImpl) Delete(ctx context.Context, projectID string, kmsID string, kmipID string, params *types.RequestParameters) (*types.Response[any], error) {
	c.client.Logger().Debugf("Deleting KMIP service: %s for KMS: %s in project: %s", kmipID, kmsID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kmsID, "KMS ID"); err != nil {
		return nil, err
	}

	if kmipID == "" {
		return nil, fmt.Errorf("KMIP ID cannot be empty")
	}

	path := fmt.Sprintf(KmipPath, projectID, kmsID, kmipID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &KmipDeleteAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &KmipDeleteAPIVersion
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

// Download downloads the KMIP certificate (key and cert) for a specific KMIP service
func (c *KmipClientImpl) Download(ctx context.Context, projectID string, kmsID string, kmipID string, params *types.RequestParameters) (*types.Response[types.KmipCertificateResponse], error) {
	c.client.Logger().Debugf("Downloading KMIP certificate: %s for KMS: %s in project: %s", kmipID, kmsID, projectID)

	if err := types.ValidateProjectAndResource(projectID, kmsID, "KMS ID"); err != nil {
		return nil, err
	}

	if kmipID == "" {
		return nil, fmt.Errorf("KMIP ID cannot be empty")
	}

	path := fmt.Sprintf(KmipDownloadPath, projectID, kmsID, kmipID)

	if params == nil {
		params = &types.RequestParameters{
			APIVersion: &KmipDownloadAPIVersion,
		}
	} else if params.APIVersion == nil {
		params.APIVersion = &KmipDownloadAPIVersion
	}

	queryParams := params.ToQueryParams()
	headers := params.ToHeaders()

	httpResp, err := c.client.DoRequest(ctx, http.MethodGet, path, nil, queryParams, headers)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return types.ParseResponseBody[types.KmipCertificateResponse](httpResp, c.client.Logger())
}
