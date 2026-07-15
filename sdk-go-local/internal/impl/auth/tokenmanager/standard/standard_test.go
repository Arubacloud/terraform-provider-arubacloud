package standard

import (
	"context"
	"errors"
	"math/rand/v2"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/Arubacloud/sdk-go/internal/ports/auth"
	"github.com/Arubacloud/sdk-go/internal/ports/interceptor"
)

//go:generate mockgen -package standard -destination=zz_mock_auth_test.go github.com/Arubacloud/sdk-go/internal/ports/auth TokenRepository,ProviderConnector
//go:generate mockgen -package standard -destination=zz_mock_interceptor_test.go github.com/Arubacloud/sdk-go/internal/ports/interceptor Interceptable

// Common parameters
var (
	tokenKey    = "Authorization"
	tokenPrefix = "Bearer "
	accessToken = "this is a valid token"
	expiry      = time.Now().Add(24 * time.Hour)
)

func TestTokenManager_BindTo(t *testing.T) {
	t.Run("should handle nil interceptable gracefully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a fresh token manager
		tokenManager := NewTokenManager(NewMockProviderConnector(ctrl), NewMockTokenRepository(ctrl))

		// When we try to bind to a nill intercaptable
		err := tokenManager.BindTo(nil)

		// Then an invalid interceptable eror is reported
		require.ErrorIs(t, err, auth.ErrInvalidInterceptable)

		// And the message informs that a nil interceptable was given as parameter
		require.ErrorContains(t, err, "not possible to bind to a nil interceptable")
	})

	t.Run("should bind to a valid interceptable", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a fresh token manager
		tokenManager := NewTokenManager(NewMockProviderConnector(ctrl), NewMockTokenRepository(ctrl))

		// And a valid interceptable
		interceptable := NewMockInterceptable(ctrl)

		interceptable.EXPECT().Bind(gomock.AssignableToTypeOf(tokenManager.InjectToken)).DoAndReturn(
			func(interceptFunc interceptor.InterceptFunc) error {
				if reflect.ValueOf(interceptFunc).Pointer() == reflect.ValueOf(tokenManager.InjectToken).Pointer() {
					return nil
				}

				return errors.New("interceptFuncs does not contains tokenManager.InjectToken")
			}).Times(1)

		//
		// When we try to bind them
		err := tokenManager.BindTo(interceptable)

		// Then no error should be reported
		require.NoError(t, err)
	})
}

func TestTokenManager_InjectToken(t *testing.T) {
	t.Run("should panic when using a nil repository", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a fresh token manager which received a nil repository
		tokenManager := NewTokenManager(NewMockProviderConnector(ctrl), nil)

		// And a fresh valid http request
		r, _ := http.NewRequest(http.MethodGet, "https://www.aruba.it/", nil)

		// When we try to inject a token into the request
		// Then it should panic
		require.Panics(t, func() {
			tokenManager.InjectToken(t.Context(), r)
		})
	})

	t.Run("should panic when using a nil connector", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a repository which does not contains a token
		repository := NewMockTokenRepository(ctrl)

		repository.EXPECT().FetchToken(
			gomock.AssignableToTypeOf(t.Context()),
		).Return(nil, auth.ErrTokenNotFound).Times(1)

		repository.EXPECT().SaveToken(gomock.Any(), gomock.Any()).Times(0)

		//
		// And a fresh token manager which received a nil connector
		tokenManager := NewTokenManager(nil, repository)

		// And a fresh valid http request
		r, _ := http.NewRequest(http.MethodGet, "https://www.aruba.it/", nil)

		// When we try to inject a token into the request
		// Then it should panic
		require.Panics(t, func() {
			tokenManager.InjectToken(t.Context(), r)
		})
	})

	t.Run("should require a new token and save it when the repository does not have one", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a repository which does not contains a token
		repository := NewMockTokenRepository(ctrl)

		repository.EXPECT().FetchToken(
			gomock.AssignableToTypeOf(t.Context()),
		).Return(nil, auth.ErrTokenNotFound).Times(1)

		var savedToken *auth.Token

		repository.EXPECT().SaveToken(
			gomock.AssignableToTypeOf(t.Context()),
			gomock.AssignableToTypeOf(savedToken),
		).DoAndReturn(
			func(ctx context.Context, token *auth.Token) error {
				savedToken = token

				return nil
			}).Times(1)

		//
		// And a valid connector connector capable to return a token
		connector := NewMockProviderConnector(ctrl)

		connector.EXPECT().RequestToken(gomock.AssignableToTypeOf(t.Context())).Return(
			&auth.Token{
				AccessToken: accessToken,
				Expiry:      expiry,
			}, nil).Times(1)

		//
		// And a fresh token manager using both repository and connector
		tokenManager := NewTokenManager(connector, repository)

		// And a fresh valid http request
		r, _ := http.NewRequest(http.MethodGet, "https://www.aruba.it/", nil)

		// When we try to inject a token into the request
		err := tokenManager.InjectToken(t.Context(), r)

		// Then no error should be reported
		require.NoError(t, err)

		// And the access token into the request should match the token fetch
		// from the repository
		extractAndValidateToken(t, r, accessToken)

		// And the seved token should be the same as returned by the connector
		require.Equal(t, accessToken, savedToken.AccessToken)
		require.Equal(t, expiry, savedToken.Expiry)
	})

	t.Run("should require a new token and save it when the repository has an expired one", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a repository which contains an expired token
		repository := NewMockTokenRepository(ctrl)

		repository.EXPECT().FetchToken(gomock.AssignableToTypeOf(t.Context())).Return(
			&auth.Token{
				AccessToken: accessToken,
				Expiry:      time.Now().Add(-24 * time.Hour),
			}, nil).Times(1)

		var savedToken *auth.Token

		repository.EXPECT().SaveToken(
			gomock.AssignableToTypeOf(t.Context()),
			gomock.AssignableToTypeOf(savedToken),
		).DoAndReturn(
			func(ctx context.Context, token *auth.Token) error {
				savedToken = token

				return nil
			}).Times(1)

		//
		// And a valid connector connector capable to return a token
		connector := NewMockProviderConnector(ctrl)

		connector.EXPECT().RequestToken(gomock.AssignableToTypeOf(t.Context())).Return(
			&auth.Token{
				AccessToken: accessToken,
				Expiry:      expiry,
			}, nil).Times(1)

		//
		// And a fresh token manager using both repository and connector
		tokenManager := NewTokenManager(connector, repository)

		// And a fresh valid http request
		r, _ := http.NewRequest(http.MethodGet, "https://www.aruba.it/", nil)

		// When we try to inject a token into the request
		err := tokenManager.InjectToken(t.Context(), r)

		// Then no error should be reported
		require.NoError(t, err)

		// And the access token into the request should match the token fetch
		// from the repository
		extractAndValidateToken(t, r, accessToken)

		// And the seved token should be the same as returned by the connector
		require.Equal(t, accessToken, savedToken.AccessToken)
		require.Equal(t, expiry, savedToken.Expiry)
	})

	t.Run("should not overlap token renewals", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a repository which contains an expired token
		repository := NewMockTokenRepository(ctrl)

		savedToken := &auth.Token{
			AccessToken: accessToken,
			Expiry:      time.Now().Add(-24 * time.Hour),
		}

		savedTokenPtr := &savedToken

		repository.EXPECT().FetchToken(gomock.AssignableToTypeOf(t.Context())).DoAndReturn(
			func(ctx context.Context) (*auth.Token, error) {
				time.Sleep(time.Duration(100+rand.IntN(100)) * time.Microsecond)

				return *savedTokenPtr, nil
			}).MaxTimes(199)

		repository.EXPECT().SaveToken(
			gomock.AssignableToTypeOf(t.Context()),
			gomock.AssignableToTypeOf(savedToken),
		).DoAndReturn(
			func(ctx context.Context, token *auth.Token) error {
				time.Sleep(time.Duration(100+rand.IntN(100)) * time.Microsecond)

				*savedTokenPtr = token

				return nil
			}).Times(1)

		//
		// And a valid connector connector capable to return a token
		connector := NewMockProviderConnector(ctrl)

		connector.EXPECT().RequestToken(gomock.AssignableToTypeOf(t.Context())).DoAndReturn(
			func(ctx context.Context) (*auth.Token, error) {
				time.Sleep(time.Duration(100+rand.IntN(100)) * time.Microsecond)

				return &auth.Token{
					AccessToken: accessToken,
					Expiry:      expiry,
				}, nil
			}).Times(1)

		//
		// And a fresh token manager using both repository and connector
		tokenManager := NewTokenManager(connector, repository)

		// When we try to inject token 100 times simultaneously
		type result struct { // to carry the results of each simultaneous call
			r   *http.Request
			err error
		}

		resultChan := make(chan *result, 100) // to carry the results of all simultaneous calls

		var wg sync.WaitGroup // to wait all calls to finish

		var bell sync.RWMutex // to make the calls to run closed as possible to be really simultaneous

		bell.Lock() // make sure to block all calls

		for range 100 { // launch the simultaneous go routines
			wg.Add(1)
			go func() {
				defer wg.Done()

				bell.RLock()
				defer bell.RUnlock()

				r, _ := http.NewRequest(http.MethodGet, "https://www.aruba.it/", nil)
				err := tokenManager.InjectToken(t.Context(), r)

				resultChan <- &result{r, err}
			}()
		}

		bell.Unlock() // unblock all calls at the same time

		wg.Wait() // wait all calls to be finished

		close(resultChan)

		for returned := range resultChan {
			//
			// Then no error should be reported
			require.NoError(t, returned.err)

			// And the access token into the request should match the token fetch
			// from the repository
			extractAndValidateToken(t, returned.r, accessToken)
		}

		// And the seved token should be the same as returned by the connector
		require.Equal(t, accessToken, savedToken.AccessToken)
		require.Equal(t, expiry, savedToken.Expiry)

		// And the token ticket must be 1
		require.Equal(t, uint64(1), tokenManager.ticket)
	})

	t.Run("should return an error when fetch token fail in unexpected way", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a repository which a connection error leads fetch token to fail
		repository := NewMockTokenRepository(ctrl)

		repository.EXPECT().FetchToken(gomock.AssignableToTypeOf(t.Context())).Return(
			nil,
			errors.New("connection error"),
		).Times(1)

		repository.EXPECT().SaveToken(gomock.Any(), gomock.Any()).Times(0)

		//
		// And a connector connector which returns an error when the token is request
		connector := NewMockProviderConnector(ctrl)

		connector.EXPECT().RequestToken(gomock.Any()).Times(0)

		//
		// And a fresh token manager using both repository and connector
		tokenManager := NewTokenManager(connector, repository)

		// And a fresh valid http request
		r, _ := http.NewRequest(http.MethodGet, "https://www.aruba.it/", nil)

		// When we try to inject a token into the request
		err := tokenManager.InjectToken(t.Context(), r)

		// Then an error should be reported
		require.Error(t, err)
	})

	t.Run("should return an error when the token request fail", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a repository which contains an expired token
		repository := NewMockTokenRepository(ctrl)

		repository.EXPECT().FetchToken(gomock.AssignableToTypeOf(t.Context())).Return(
			&auth.Token{
				AccessToken: accessToken,
				Expiry:      time.Now().Add(-24 * time.Hour),
			}, nil).Times(1)

		repository.EXPECT().SaveToken(gomock.Any(), gomock.Any()).Times(0)

		//
		// And a connector connector which a connection problem leads token requesto to fail
		connector := NewMockProviderConnector(ctrl)

		connector.EXPECT().RequestToken(
			gomock.AssignableToTypeOf(t.Context()),
		).Return(nil, errors.New("connection error")).Times(1)

		//
		// And a fresh token manager using both repository and connector
		tokenManager := NewTokenManager(connector, repository)

		// And a fresh valid http request
		r, _ := http.NewRequest(http.MethodGet, "https://www.aruba.it/", nil)

		// When we try to inject a token into the request
		err := tokenManager.InjectToken(t.Context(), r)

		// Then an error should be reported
		require.Error(t, err)
	})

	t.Run("should return an error when save token fail", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a repository which a permission problem leads save token to fail
		repository := NewMockTokenRepository(ctrl)

		repository.EXPECT().FetchToken(gomock.AssignableToTypeOf(t.Context())).Return(
			&auth.Token{
				AccessToken: accessToken,
				Expiry:      time.Now().Add(-24 * time.Hour),
			}, nil).Times(1)

		repository.EXPECT().SaveToken(
			gomock.AssignableToTypeOf(t.Context()),
			gomock.AssignableToTypeOf(&auth.Token{}),
		).Return(errors.New("permission error")).Times(1)

		//
		// And a valid connector connector capable to return a token
		connector := NewMockProviderConnector(ctrl)

		connector.EXPECT().RequestToken(gomock.AssignableToTypeOf(t.Context())).Return(
			&auth.Token{
				AccessToken: accessToken,
				Expiry:      expiry,
			}, nil).Times(1)

		//
		// And a fresh token manager using both repository and connector
		tokenManager := NewTokenManager(connector, repository)

		// And a fresh valid http request
		r, _ := http.NewRequest(http.MethodGet, "https://www.aruba.it/", nil)

		// When we try to inject a token into the request
		err := tokenManager.InjectToken(t.Context(), r)

		// Then an error should be reported
		require.Error(t, err)
	})

	t.Run("should inject the token received from the repository if it is valid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a repository which contains a valid token
		repository := NewMockTokenRepository(ctrl)

		repository.EXPECT().FetchToken(gomock.AssignableToTypeOf(t.Context())).Return(
			&auth.Token{
				AccessToken: accessToken,
				Expiry:      expiry,
			}, nil).Times(1)

		repository.EXPECT().SaveToken(gomock.Any(), gomock.Any()).Times(0)

		//
		// And a valid connector connector
		connector := NewMockProviderConnector(ctrl)

		connector.EXPECT().RequestToken(gomock.Any()).Times(0)

		// And a fresh token manager using both repository and connector
		tokenManager := NewTokenManager(connector, repository)

		// And a fresh valid http request
		r, _ := http.NewRequest(http.MethodGet, "https://www.aruba.it/", nil)

		// When we try to inject a token into the request
		err := tokenManager.InjectToken(t.Context(), r)

		// Then no error should be reported
		require.NoError(t, err)

		// And the access token into the request should match the token fetch
		// from the repository
		extractAndValidateToken(t, r, accessToken)
	})
}

func extractAndValidateToken(t *testing.T, r *http.Request, expectedToken string) {
	t.Helper()

	require.Contains(t, r.Header, tokenKey)
	require.Len(t, r.Header[tokenKey], 1)
	require.True(t, strings.HasPrefix(r.Header[tokenKey][0], tokenPrefix))
	require.Equal(t, expectedToken, strings.TrimPrefix(r.Header[tokenKey][0], tokenPrefix))
}
