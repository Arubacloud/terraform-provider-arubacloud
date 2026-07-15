package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Arubacloud/sdk-go/internal/impl/interceptor/standard"
	"github.com/Arubacloud/sdk-go/internal/impl/logger/noop"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

const cannedToken = `{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`

// NewMockServer starts an httptest.Server that handles /token (returning a canned bearer token)
// and delegates all other paths to handler. Registers t.Cleanup(s.Close).
func NewMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, err := fmt.Fprint(w, cannedToken); err != nil {
				t.Errorf("mock server: failed to write token: %v", err)
			}
			return
		}
		handler(w, r)
	}))
	t.Cleanup(s.Close)
	return s
}

// NewClient builds a restclient.Client pointed at baseURL with standard interceptor and noop logger.
func NewClient(t *testing.T, baseURL string) *restclient.Client {
	t.Helper()
	return restclient.NewClient(baseURL, http.DefaultClient, standard.NewInterceptor(), &noop.NoOpLogger{})
}

// ErrRoundTripper is an http.RoundTripper that returns Err on every call.
type ErrRoundTripper struct{ Err error }

func (e ErrRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, e.Err
}

// NewBrokenClient returns a client whose transport always returns an error, allowing
// tests to verify network-error handling without racing a server.Close().
func NewBrokenClient(t *testing.T, baseURL string) *restclient.Client {
	t.Helper()
	broken := &http.Client{Transport: ErrRoundTripper{Err: fmt.Errorf("connection refused")}}
	return restclient.NewClient(baseURL, broken, standard.NewInterceptor(), &noop.NoOpLogger{})
}

// ErrorBodyJSON returns a well-formed types.ErrorResponse JSON string fixture.
func ErrorBodyJSON(title, detail string, status int) string {
	s := int32(status)
	b, err := json.Marshal(types.ErrorResponse{
		Title:  &title,
		Detail: &detail,
		Status: &s,
	})
	if err != nil {
		panic(fmt.Sprintf("testutil.ErrorBodyJSON: marshal failed: %v", err))
	}
	return string(b)
}
