package restclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/impl/interceptor/standard"
	"github.com/Arubacloud/sdk-go/internal/impl/logger/noop"
)

func newTestRestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	return NewClient(baseURL, http.DefaultClient, standard.NewInterceptor(), &noop.NoOpLogger{})
}

func TestDoRequestAbs_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(server.Close)

	client := newTestRestClient(t, server.URL)
	resp, err := client.DoRequestAbs(context.Background(), http.MethodGet, server.URL+"/resource", nil, nil, nil)
	if err != nil {
		t.Fatalf("DoRequestAbs() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "ok") {
		t.Errorf("body = %q, want to contain %q", body, "ok")
	}
}

func TestDoRequestAbs_RelativeURL(t *testing.T) {
	var capturedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := newTestRestClient(t, server.URL)
	resp, err := client.DoRequestAbs(context.Background(), http.MethodGet, "/v1/resource", nil, nil, nil)
	if err != nil {
		t.Fatalf("DoRequestAbs() with relative URL error = %v", err)
	}
	defer resp.Body.Close()
	if capturedPath != "/v1/resource" {
		t.Errorf("resolved path = %q, want %q", capturedPath, "/v1/resource")
	}
}

func TestDoRequestAbs_NonSuccessReturnsWithoutTransportError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":500}`))
	}))
	t.Cleanup(server.Close)

	client := newTestRestClient(t, server.URL)
	resp, err := client.DoRequestAbs(context.Background(), http.MethodGet, server.URL+"/err", nil, nil, nil)
	if err != nil {
		t.Fatalf("DoRequestAbs() should not return transport error on 500: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestDoRequestAbs_HeadersForwarded(t *testing.T) {
	var capturedHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Custom-Header")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := newTestRestClient(t, server.URL)
	resp, err := client.DoRequestAbs(context.Background(), http.MethodGet, server.URL+"/resource", nil, nil,
		map[string]string{"X-Custom-Header": "test-value"})
	if err != nil {
		t.Fatalf("DoRequestAbs() error = %v", err)
	}
	defer resp.Body.Close()
	if capturedHeader != "test-value" {
		t.Errorf("X-Custom-Header = %q, want %q", capturedHeader, "test-value")
	}
}

func TestDoRequestAbs_TransportError(t *testing.T) {
	// Point client at a server that is not listening.
	client := newTestRestClient(t, "http://localhost:1")
	_, err := client.DoRequestAbs(context.Background(), http.MethodGet, "http://localhost:1/resource", nil, nil, nil)
	if err == nil {
		t.Fatal("DoRequestAbs() should return error when transport fails")
	}
}

func TestDoRequestAbs_WithBody(t *testing.T) {
	var capturedContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := newTestRestClient(t, server.URL)
	body := strings.NewReader(`{"key":"value"}`)
	resp, err := client.DoRequestAbs(context.Background(), http.MethodPost, server.URL+"/resource", body, nil, nil)
	if err != nil {
		t.Fatalf("DoRequestAbs() with body error = %v", err)
	}
	defer resp.Body.Close()
	if capturedContentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", capturedContentType, "application/json")
	}
}
