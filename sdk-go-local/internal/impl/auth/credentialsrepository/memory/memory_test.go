package memory

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/Arubacloud/sdk-go/internal/ports/auth"
)

//go:generate mockgen -package memory -destination=zz_mock_auth_test.go github.com/Arubacloud/sdk-go/internal/ports/auth CredentialsRepository

const (
	clientID     = "client id"
	clientSecret = "client secret"
)

func TestCredentialsRepository_FetchCredentials(t *testing.T) {
	t.Run("should return the credentials", func(t *testing.T) {
		// Given a fresh new credentials repository which contains valid credentials
		credentialsRepository := NewCredentialsRepository(clientID, clientSecret)

		// When we try to fetch the credentials
		credentials, err := credentialsRepository.FetchCredentials(t.Context())

		// Then no error should be reported
		require.NoError(t, err)

		// And some credentials should be returned
		require.NotNil(t, credentials)

		// And the credentials should match with the ones stored on the repository
		require.Equal(t, clientID, credentials.ClientID)
		require.Equal(t, clientSecret, credentials.ClientSecret)

		// And the credential holders should not be the same
		require.NotSame(t, credentialsRepository.credentials, credentials)
	})
}

func TestCredentialsProxy_FetchCredentials(t *testing.T) {
	t.Run("should forward errors received from the persistent repository", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a persistent repository which experiences some connection issues
		persistentRepository := NewMockCredentialsRepository(ctrl)

		errConnection := errors.New("connection error")

		persistentRepository.EXPECT().FetchCredentials(
			gomock.AssignableToTypeOf(t.Context()),
		).Return(nil, errConnection).Times(1)

		//
		// And a fresh new proxy using that last
		proxy := NewCredentialsProxy(persistentRepository)

		//
		// When we try to fetch the credentials
		credentials, err := proxy.FetchCredentials(t.Context())

		// Then it should report the same error returned by the persistent repository
		require.ErrorIs(t, err, errConnection)

		// And no credentials shoud be returned
		require.Nil(t, credentials)
	})

	t.Run("should fetch the credentials from the persistent repository when is has no credentials yet", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a persistent repository which contains valid credentials
		persistentRepository := NewMockCredentialsRepository(ctrl)

		persistentRepository.EXPECT().FetchCredentials(
			gomock.AssignableToTypeOf(t.Context()),
		).Return(
			&auth.Credentials{
				ClientID:     clientID,
				ClientSecret: clientSecret,
			}, nil).Times(1)

		//
		// And a fresh new proxy using that last
		proxy := NewCredentialsProxy(persistentRepository)

		// When we try to fetch the credentials
		credentials, err := proxy.FetchCredentials(t.Context())

		// Then no error should be reported
		require.NoError(t, err)

		// And some credentials should be returned
		require.NotNil(t, credentials)

		// And the credentials should match with the ones stored on the repository
		require.Equal(t, clientID, credentials.ClientID)
		require.Equal(t, clientSecret, credentials.ClientSecret)

		// And the credential holders should not be the same
		require.NotSame(t, proxy.credentials, credentials)
	})

	t.Run("should not fetch the credentials from the persistent repository when is has already has cached credentials", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a persistent repository which contains valid credentials
		persistentRepository := NewMockCredentialsRepository(ctrl)

		persistentRepository.EXPECT().FetchCredentials(
			gomock.AssignableToTypeOf(t.Context()),
		).Return(
			&auth.Credentials{
				ClientID:     "different client id",
				ClientSecret: "different client secret",
			}, nil).Times(0)

		//
		// And a proxy which already contains valid credentials and uses that last
		proxy := NewCredentialsProxy(persistentRepository)

		proxy.credentials = &auth.Credentials{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}

		//
		// When we try to fetch the credentials
		credentials, err := proxy.FetchCredentials(t.Context())

		// Then no error should be reported
		require.NoError(t, err)

		// And some credentials should be returned
		require.NotNil(t, credentials)

		// And the credentials should not change
		require.Equal(t, clientID, credentials.ClientID)
		require.Equal(t, clientSecret, credentials.ClientSecret)

		// And the credential holders should not be the same
		require.NotSame(t, proxy.credentials, credentials)
	})
}
