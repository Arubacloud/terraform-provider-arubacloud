package oauth2

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/Arubacloud/sdk-go/internal/ports/auth"
)

//go:generate mockgen -package oauth2 -destination=zz_mock_auth_test.go github.com/Arubacloud/sdk-go/internal/ports/auth CredentialsRepository

var (
	clientID     = "this_is_a_valid_client_id"
	clientSecret = "this_is_a_valid_client_secret"
	accessToken  = "this_is_a_valid_access_token"
	scopes       = []string{"read:compute", "write:compute", "read:storage", "write:storage"}
	expireIn     = 24 * 60 * 20 // 1 day, 24 hour in seconds
	expiry       = time.Now().Add(time.Duration(expireIn) * time.Second)
)

func TestProviderConnector_RequestToken(t *testing.T) {
	t.Run("should return token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a fully functional OAuth2 server
		oauth2Server := SetupConfigurableTokenServer(t, MockServerConfig{
			StatusCode: http.StatusOK,

			AccessToken:  accessToken,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       scopes,
			ExpiresIn:    expireIn,
		})

		defer oauth2Server.Close()

		// And a fully functional credentials repository
		credentialsRepository := NewMockCredentialsRepository(ctrl)

		credentialsRepository.EXPECT().FetchCredentials(
			gomock.AssignableToTypeOf(t.Context()),
		).Return(&auth.Credentials{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}, nil).Times(1)

		//
		// And a Provider connector which use both previous components
		providerConnector := NewProviderConnector(credentialsRepository, oauth2Server.URL, scopes)

		// When we request a token
		token, err := providerConnector.RequestToken(t.Context())

		// Then no error should be reported
		require.NoError(t, err)

		// And a token should be returned
		require.NotNil(t, token)

		// And the token data should be coherent with data sent by the server
		require.Equal(t, accessToken, token.AccessToken)
		require.InDelta(t, expiry.UTC().Unix(), token.Expiry.UTC().Unix(), 5.0) // 5 seconds of toleration
	})

	t.Run("should report a proper error for unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a fully functional OAuth2 server
		oauth2Server := SetupConfigurableTokenServer(t, MockServerConfig{
			StatusCode: http.StatusUnauthorized,
			ErrorBody:  `{"message": "Unauthorized"}`,
		})

		defer oauth2Server.Close()

		// And a fully functional credentials repository which contains invalid credentials
		credentialsRepository := NewMockCredentialsRepository(ctrl)

		credentialsRepository.EXPECT().FetchCredentials(
			gomock.AssignableToTypeOf(t.Context()),
		).Return(&auth.Credentials{
			ClientID:     clientID,
			ClientSecret: "this_is_s_not_valid_client_secret",
		}, nil).Times(1)

		//
		// And a Provider connector which use both previous components
		providerConnector := NewProviderConnector(credentialsRepository, oauth2Server.URL, scopes)

		// When we request a token
		token, err := providerConnector.RequestToken(t.Context())

		// Then an authentication failed error should be reported
		require.ErrorIs(t, err, auth.ErrAuthenticationFailed)

		// And no token should be returned
		require.Nil(t, token)
	})

	t.Run("should report a proper error for insufficient privileges", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a fully functional OAuth2 server
		oauth2Server := SetupConfigurableTokenServer(t, MockServerConfig{
			StatusCode: http.StatusForbidden,
			ErrorBody:  `{"message": "Insufficient Privileges"}`,
		})

		defer oauth2Server.Close()

		// And a fully functional credentials repository which contains invalid credentials
		credentialsRepository := NewMockCredentialsRepository(ctrl)

		credentialsRepository.EXPECT().FetchCredentials(
			gomock.AssignableToTypeOf(t.Context()),
		).Return(&auth.Credentials{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}, nil).Times(1)

		//
		// And a Provider connector which use both previous components
		providerConnector := NewProviderConnector(credentialsRepository, oauth2Server.URL, scopes)

		// When we request a token
		token, err := providerConnector.RequestToken(t.Context())

		// Then an insufficient privileges error should be reported
		require.ErrorIs(t, err, auth.ErrInsufficientPrivileges)

		// And no token should be returned
		require.Nil(t, token)
	})

	t.Run("should forward other auth provider errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a OAuth2 server experiencing some issues
		oauth2Server := SetupConfigurableTokenServer(t, MockServerConfig{
			StatusCode: http.StatusInternalServerError,
			ErrorBody:  `{"message": "Service Temporarelly Unavailable"}`,
		})

		defer oauth2Server.Close()

		// And a fully functional credentials repository which contains valid credentials
		credentialsRepository := NewMockCredentialsRepository(ctrl)

		credentialsRepository.EXPECT().FetchCredentials(
			gomock.AssignableToTypeOf(t.Context()),
		).Return(&auth.Credentials{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}, nil).Times(1)

		//
		// And a Provider connector which use both previous components
		providerConnector := NewProviderConnector(credentialsRepository, oauth2Server.URL, scopes)

		// When we request a token
		token, err := providerConnector.RequestToken(t.Context())

		// Then an error should be reported
		require.Error(t, err)

		// And no token should be returned
		require.Nil(t, token)
	})

	t.Run("should forward credential provider errors", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a fully functional OAuth2 server
		// Given a fully functional OAuth2 server
		oauth2Server := SetupConfigurableTokenServer(t, MockServerConfig{
			StatusCode: http.StatusOK,

			AccessToken:  accessToken,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       scopes,
			ExpiresIn:    expireIn,
		})

		defer oauth2Server.Close()

		// And a credentials repository which does not contain any credentials
		credentialsRepository := NewMockCredentialsRepository(ctrl)

		credentialsRepository.EXPECT().FetchCredentials(
			gomock.AssignableToTypeOf(t.Context()),
		).Return(nil, auth.ErrCredentialsNotFound).Times(1)

		//
		// And a Provider connector which use both previous components
		providerConnector := NewProviderConnector(credentialsRepository, oauth2Server.URL, scopes)

		// When we request a token
		token, err := providerConnector.RequestToken(t.Context())

		// Then a credentials not found error should be reported
		require.ErrorIs(t, err, auth.ErrCredentialsNotFound)

		// And no token should be returned
		require.Nil(t, token)
	})
}
