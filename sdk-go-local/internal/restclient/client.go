package restclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Arubacloud/sdk-go/internal/ports/interceptor"
	"github.com/Arubacloud/sdk-go/internal/ports/logger"
)

// Client is the main SDK client that aggregates all resource providers
type Client struct {
	baseURL    string
	httpClient *http.Client
	middleware interceptor.Interceptor
	logger     logger.Logger
}

// NewClient creates a new SDK client with the given configuration
func NewClient(baseURL string, httpClient *http.Client, middleware interceptor.Interceptor, logger logger.Logger) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		logger:     logger,
		middleware: middleware,
	}
}

// Logger returns the client logger
func (c *Client) Logger() logger.Logger {
	return c.logger
}

// DoRequestAbs performs an HTTP request to an absolute URL (no baseURL prefix). Use this
// for pagination links returned by the server in links.next / links.prev.
func (c *Client) DoRequestAbs(ctx context.Context, method, absURL string, body io.Reader, queryParams map[string]string, headers map[string]string) (*http.Response, error) {
	c.logger.Debugf("Making %s request to absolute URL %s", method, absURL)

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			c.logger.Errorf("Failed to read request body: %v", err)
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		c.logger.Debugf("Request body: %s", string(bodyBytes))
		body = bytes.NewReader(bodyBytes)
	}

	// Resolve relative pagination links against the client base URL.
	if parsed, parseErr := url.Parse(absURL); parseErr == nil && !parsed.IsAbs() {
		if base, baseErr := url.Parse(c.baseURL); baseErr == nil {
			absURL = base.ResolveReference(parsed).String()
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, absURL, body)
	if err != nil {
		c.logger.Errorf("Failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if len(queryParams) > 0 {
		q := req.URL.Query()
		for key, value := range queryParams {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if err := c.middleware.Intercept(ctx, req); err != nil {
		c.logger.Errorf("Failed to prepare request: %v", err)
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Errorf("Request failed: %v", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}

	c.logger.Debugf("Received response with status: %d %s", resp.StatusCode, resp.Status)

	if resp.Body != nil {
		respBodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger.Warnf("Failed to read response body for logging: %v", err)
		} else {
			c.logger.Debugf("Response body: %s", string(respBodyBytes))
			resp.Body = io.NopCloser(bytes.NewReader(respBodyBytes))
		}
	}

	return resp, nil
}

// DoRequest performs an HTTP request with automatic authentication token injection
func (c *Client) DoRequest(ctx context.Context, method, path string, body io.Reader, queryParams map[string]string, headers map[string]string) (*http.Response, error) {
	// Build full URL
	url := c.baseURL + path

	c.logger.Debugf("Making %s request to %s", method, url)

	// Read body for logging if present
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			c.logger.Errorf("Failed to read request body: %v", err)
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		c.logger.Debugf("Request body: %s", string(bodyBytes))
		// Recreate reader for actual request
		body = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		c.logger.Errorf("Failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	if len(queryParams) > 0 {
		q := req.URL.Query()
		for key, value := range queryParams {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()
		c.logger.Debugf("Added query parameters: %v", queryParams)
	}

	// Add caller-supplied headers first so SDK-controlled headers applied below cannot be overridden
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Set content type for requests with body; applied after caller headers to prevent override
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Log request headers (before auth)
	c.logger.Debugf("Request headers (pre-auth): %v", headers)

	// Use middleware
	if err := c.middleware.Intercept(ctx, req); err != nil {
		c.logger.Errorf("Failed to prepare request: %v", err)
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	// Log all headers after auth (excluding Authorization token for security)
	sanitizedHeaders := make(map[string]string)
	for key, values := range req.Header {
		if key == "Authorization" {
			sanitizedHeaders[key] = "Bearer [REDACTED]"
		} else {
			sanitizedHeaders[key] = values[0]
		}
	}
	c.logger.Debugf("Request headers (final): %v", sanitizedHeaders)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Errorf("Request failed: %v", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}

	c.logger.Debugf("Received response with status: %d %s", resp.StatusCode, resp.Status)

	// Log response headers
	c.logger.Debugf("Response headers: %v", resp.Header)

	// Log response body (for debugging)
	if resp.Body != nil {
		respBodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			c.logger.Warnf("Failed to read response body for logging: %v", err)
		} else {
			c.logger.Debugf("Response body: %s", string(respBodyBytes))
			// Recreate the response body so it can be read by the caller
			resp.Body = io.NopCloser(bytes.NewReader(respBodyBytes))
		}
	}

	return resp, nil
}
