package oauth2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockServerConfig struct {
	StatusCode int
	ErrorBody  string

	AccessToken  string
	ExpiresIn    int
	ClientID     string
	ClientSecret string
	Scopes       []string
}

func SetupConfigurableTokenServer(t *testing.T, config MockServerConfig) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		var response map[string]any

		if config.StatusCode == http.StatusOK {
			parseAndAssertClientCredentials(t, r, config)

			response = map[string]any{
				"access_token": config.AccessToken,
				"token_type":   "Bearer",
				"expires_in":   config.ExpiresIn,
			}

		} else {
			response = map[string]any{
				"error":             http.StatusText(config.StatusCode),
				"error_description": config.ErrorBody,
			}
		}

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(config.StatusCode)

		json.NewEncoder(w).Encode(response)
	}))
}

func parseAndAssertClientCredentials(t *testing.T, r *http.Request, config MockServerConfig) {
	t.Helper()

	assert.NotPanics(t, func() {
		assert.NoError(t, r.ParseForm())

		assert.True(t, r.Form.Has("grant_type"))
		assert.Equal(t, "client_credentials", r.Form.Get("grant_type"))

		assert.True(t, r.Form.Has("scope"))
		scopeValue := r.Form.Get("scope")
		scopes := strings.Split(scopeValue, " ")
		assert.ElementsMatch(t, scopes, config.Scopes)

		givenClientID, givenClientSecret, hasBasicAuth := r.BasicAuth()
		assert.True(t, hasBasicAuth)
		assert.Equal(t, config.ClientID, givenClientID)
		assert.Equal(t, config.ClientSecret, givenClientSecret)
	})
}
