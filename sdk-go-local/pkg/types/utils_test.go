package types

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// captureLogger records Debugf calls for assertion in tests.
type captureLogger struct {
	msgs []string
}

func (l *captureLogger) Debugf(format string, args ...interface{}) {
	l.msgs = append(l.msgs, fmt.Sprintf(format, args...))
}

func makeHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{},
	}
}

func TestParseResponseBody(t *testing.T) {
	t.Run("2xx with valid JSON sets Data, no debug message", func(t *testing.T) {
		logger := &captureLogger{}
		resp, err := ParseResponseBody[map[string]string](
			makeHTTPResponse(200, `{"key":"value"}`),
			logger,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Data == nil || (*resp.Data)["key"] != "value" {
			t.Errorf("expected Data to be set with parsed JSON")
		}
		if resp.Error != nil {
			t.Errorf("expected Error to be nil for 2xx response")
		}
		if len(logger.msgs) != 0 {
			t.Errorf("expected no debug messages, got %d: %v", len(logger.msgs), logger.msgs)
		}
	})

	t.Run("4xx with valid JSON error body sets Error, no debug message", func(t *testing.T) {
		logger := &captureLogger{}
		body := `{"code":"NOT_FOUND","message":"resource not found"}`
		resp, err := ParseResponseBody[map[string]string](
			makeHTTPResponse(404, body),
			logger,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Error == nil {
			t.Errorf("expected Error to be set for 4xx response with valid JSON")
		}
		if resp.Data != nil {
			t.Errorf("expected Data to be nil for 4xx response")
		}
		if len(logger.msgs) != 0 {
			t.Errorf("expected no debug messages, got %d: %v", len(logger.msgs), logger.msgs)
		}
	})

	t.Run("4xx with non-JSON body leaves Error nil and emits one debug message", func(t *testing.T) {
		logger := &captureLogger{}
		resp, err := ParseResponseBody[map[string]string](
			makeHTTPResponse(502, "<html><body>Bad Gateway</body></html>"),
			logger,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Error != nil {
			t.Errorf("expected Error to be nil when body is not JSON")
		}
		if len(resp.RawBody) == 0 {
			t.Errorf("expected RawBody to be set even when JSON parse fails")
		}
		if len(logger.msgs) != 1 {
			t.Fatalf("expected exactly 1 debug message, got %d: %v", len(logger.msgs), logger.msgs)
		}
		if !strings.Contains(logger.msgs[0], "502") {
			t.Errorf("expected debug message to mention status code 502, got: %s", logger.msgs[0])
		}
	})

	t.Run("nil httpResp returns error without panic", func(t *testing.T) {
		logger := &captureLogger{}
		resp, err := ParseResponseBody[map[string]string](nil, logger)
		if err == nil {
			t.Fatal("expected error for nil httpResp")
		}
		if resp != nil {
			t.Errorf("expected nil response for nil httpResp")
		}
	})
}
