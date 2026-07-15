package memory

import (
	"context"
	"errors"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/Arubacloud/sdk-go/internal/ports/auth"
)

//go:generate mockgen -package memory -destination=zz_mock_auth_test.go github.com/Arubacloud/sdk-go/internal/ports/auth TokenRepository

// Common parameters
var (
	accessToken = "this is a valid token"
	expiry      = time.Now().Add(24 * time.Hour)
)

func TestTokenRepository_FetchToken(t *testing.T) {
	t.Run("should report a token not found error when it has not a token", func(t *testing.T) {
		// Given a fresh new TokenRepository which contains no token
		tokenRepository := NewTokenRepository()

		// When we try to fetch the token
		token, err := tokenRepository.FetchToken(t.Context())

		// Then a token not found error should be reported
		require.ErrorIs(t, err, auth.ErrTokenNotFound)

		// And no token should be returned
		require.Nil(t, token)
	})

	t.Run("should return a preloaded access token with no expiry", func(t *testing.T) {
		// Given a TokenRepository created with a preloaded access token
		tokenRepository := NewTokenRepositoryWithAccessToken(accessToken)

		// When we try to fetch the token
		token, err := tokenRepository.FetchToken(t.Context())

		// Then no error should be reported
		require.NoError(t, err)

		// And a token containing the preloaded access token should be returned
		require.NotNil(t, token)
		require.Equal(t, accessToken, token.AccessToken)

		// And the token should have zero expiry (never expires)
		require.True(t, token.Expiry.IsZero())

		// And the token should be valid (zero expiry means never expires)
		require.True(t, token.IsValid())

		// And the returned token should be a copy, not the same pointer
		require.NotSame(t, tokenRepository.token, token)
	})

	t.Run("should return an expired token with no error", func(t *testing.T) {
		// Given a fresh new TokenRepository which contains an expired token
		tokenRepository := NewTokenRepository()

		passedExpiry := time.Now().Add(-24 * time.Hour)

		tokenRepository.token = &auth.Token{
			AccessToken: accessToken,
			Expiry:      passedExpiry,
		}

		//
		// When we try to fetch the token
		token, err := tokenRepository.FetchToken(t.Context())

		// Then no error shoudt be reported
		require.NoError(t, err)

		// And a token containing the same data should be returned
		require.NotNil(t, token)
		require.Equal(t, accessToken, token.AccessToken)
		require.Equal(t, passedExpiry, token.Expiry)

		// And the token should not be valid
		require.False(t, token.IsValid())

		// And the tokens should not be the same
		require.NotSame(t, tokenRepository.token, token)
	})

	t.Run("should return a valid token", func(t *testing.T) {
		// Given a fresh new TokenRepository which contains an non expired token
		tokenRepository := NewTokenRepository()

		tokenRepository.token = &auth.Token{
			AccessToken: accessToken,
			Expiry:      expiry,
		}

		//
		// When we try to fetch the token
		token, err := tokenRepository.FetchToken(t.Context())

		// Then no error shoudt be reported
		require.NoError(t, err)

		// And a token containing the same data should be returned
		require.NotNil(t, token)
		require.Equal(t, accessToken, token.AccessToken)
		require.Equal(t, expiry, token.Expiry)

		// And the token should be valid
		require.True(t, token.IsValid())

		// And the tokens should not be the same
		require.NotSame(t, tokenRepository.token, token)
	})
}

func TestTokenRepository_SaveToken(t *testing.T) {
	t.Run("should save a token when the repository does not have a token", func(t *testing.T) {
		// Given a fresh new TokenRepository which contains no token
		tokenRepository := NewTokenRepository()

		// And a valid token
		token := &auth.Token{AccessToken: accessToken, Expiry: expiry}

		// When we try to save the token
		err := tokenRepository.SaveToken(t.Context(), token)

		// Then no error should be reported
		require.NoError(t, err)

		// And the repository should contain a token with the same data of the given one
		require.NotNil(t, tokenRepository.token)
		require.Equal(t, accessToken, tokenRepository.token.AccessToken)
		require.Equal(t, expiry, tokenRepository.token.Expiry)

		// And the tokens should not be the same
		require.NotSame(t, token, tokenRepository.token)
	})

	t.Run("should replace a token when the repository has an expired token", func(t *testing.T) {
		// Given a fresh new TokenRepository which contains an expired token
		tokenRepository := NewTokenRepository()

		passedExpiry := time.Now().Add(-24 * time.Hour)

		tokenRepository.token = &auth.Token{
			AccessToken: accessToken,
			Expiry:      passedExpiry,
		}

		// And a valid token
		token := &auth.Token{AccessToken: accessToken, Expiry: expiry}

		// When we try to save the token
		err := tokenRepository.SaveToken(t.Context(), token)

		// Then no error should be reported
		require.NoError(t, err)

		// And the repository should contain a token with the same data of the given one
		require.NotNil(t, tokenRepository.token)
		require.Equal(t, accessToken, tokenRepository.token.AccessToken)
		require.Equal(t, expiry, tokenRepository.token.Expiry)

		// And the tokens should not be the same
		require.NotSame(t, token, tokenRepository.token)
	})

	t.Run("should replace a token when the repository has a valid token", func(t *testing.T) {
		// Given a fresh new TokenRepository which contains a valid token
		tokenRepository := NewTokenRepository()

		differentAcessToken := "different access token"
		differentExpiry := time.Now().Add(1 * time.Hour)

		tokenRepository.token = &auth.Token{
			AccessToken: differentAcessToken,
			Expiry:      differentExpiry,
		}

		// And a valid token
		token := &auth.Token{AccessToken: accessToken, Expiry: expiry}

		// When we try to save the token
		err := tokenRepository.SaveToken(t.Context(), token)

		// Then no error should be reported
		require.NoError(t, err)

		// And the repository should contain a token with the same data of the given one
		require.NotNil(t, tokenRepository.token)
		require.Equal(t, accessToken, tokenRepository.token.AccessToken)
		require.Equal(t, expiry, tokenRepository.token.Expiry)

		// And the tokens should not be the same
		require.NotSame(t, token, tokenRepository.token)
	})
}

func TestTokenProxy_FetchToken(t *testing.T) {
	t.Run("should forward errors from the persistent repository", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a persistent repository which experience some connection error
		persistentRepository := NewMockTokenRepository(ctrl)

		errConnection := errors.New("connection error")

		persistentRepository.EXPECT().FetchToken(
			gomock.AssignableToTypeOf(t.Context()),
		).Return(nil, errConnection).Times(1)

		//
		// And a fresh new proxy using that last
		proxy := NewTokenProxy(persistentRepository)

		// When we try to fetch the token from the proxy
		token, err := proxy.FetchToken(t.Context())

		// Then the same error obtained from the persistent repository should be reported
		require.ErrorIs(t, err, errConnection)

		// And no token should be returned
		require.Nil(t, token)
	})

	t.Run("should fetch the token from the persistent repository when it has no token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a persistent repository which already has a token
		persistentRepository := NewMockTokenRepository(ctrl)

		persistentRepository.EXPECT().FetchToken(gomock.AssignableToTypeOf(t.Context())).Return(
			&auth.Token{
				AccessToken: accessToken,
				Expiry:      expiry,
			}, nil).Times(1)

		//
		// And a fresh new proxy using that last
		proxy := NewTokenProxy(persistentRepository)

		// When we try to fetch the token from the proxy
		token, err := proxy.FetchToken(t.Context())

		// Then no error should be reported
		require.NoError(t, err)

		// And the token should be stored on memory by the proxy
		require.NotNil(t, proxy.token)

		// And the token should match the one returned from the persistent repository
		require.Equal(t, accessToken, token.AccessToken)
		require.Equal(t, expiry, token.Expiry)

		// And the tokens should not be the same
		require.NotSame(t, token, proxy.token)
	})

	t.Run("should fetch the token from the persistent repository when its own is not valid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a persistent repository which has a valid token
		persistentRepository := NewMockTokenRepository(ctrl)

		persistentRepository.EXPECT().FetchToken(gomock.AssignableToTypeOf(t.Context())).Return(
			&auth.Token{
				AccessToken: accessToken,
				Expiry:      expiry,
			}, nil).Times(1)

		//
		// And a proxy using that last which contains an expired token
		proxy := NewTokenProxy(persistentRepository)

		proxy.token = &auth.Token{
			AccessToken: "this is an expired access token",
			Expiry:      time.Now().Add(-1 * time.Hour),
		}

		//
		// When we try to fetch the token from the proxy
		token, err := proxy.FetchToken(t.Context())

		// Then no error should be reported
		require.NoError(t, err)

		// And the token should be stored on memory by the proxy
		require.NotNil(t, proxy.token)

		// And the token should match the one returned from the persistent repository
		require.Equal(t, accessToken, token.AccessToken)
		require.Equal(t, expiry, token.Expiry)

		// And the tokens should not be the same
		require.NotSame(t, token, proxy.token)
	})

	t.Run("should not overlap persistent repository calls", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a persistent repository which has a valid token
		persistentRepository := NewMockTokenRepository(ctrl)

		persistentRepository.EXPECT().FetchToken(gomock.AssignableToTypeOf(t.Context())).DoAndReturn(
			func(ctx context.Context) (*auth.Token, error) {
				time.Sleep(time.Duration(100+rand.IntN(100)) * time.Microsecond)

				return &auth.Token{
					AccessToken: accessToken,
					Expiry:      expiry,
				}, nil
			}).Times(1)

		//
		// And a proxy using that last which contains an expired token
		proxy := NewTokenProxy(persistentRepository)

		proxy.token = &auth.Token{
			AccessToken: "this is an expired access token",
			Expiry:      time.Now().Add(-1 * time.Hour),
		}

		//
		// When we try to fetch the token 100 times simultaneously
		type result struct { // to carry the results of each simultaneous call
			token *auth.Token
			err   error
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

				token, err := proxy.FetchToken(t.Context())

				resultChan <- &result{token, err}
			}()
		}

		bell.Unlock() // unblock all calls at the same time

		wg.Wait() // wait all calls to be finished

		close(resultChan)

		for returned := range resultChan {
			//
			// Then no error should be reported
			require.NoError(t, returned.err)

			// And the token data should match the one from the persistent repository
			require.Equal(t, accessToken, returned.token.AccessToken)
			require.Equal(t, expiry, returned.token.Expiry)
		}
	})

	t.Run("should return its own token when it is valid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a persistent repository which has a valid token
		persistentRepository := NewMockTokenRepository(ctrl)

		persistentRepository.EXPECT().FetchToken(gomock.AssignableToTypeOf(t.Context())).Return(
			&auth.Token{
				AccessToken: accessToken,
				Expiry:      expiry,
			}, nil).Times(0)

		//
		// And a proxy using that last which contains a different valid token
		proxy := NewTokenProxy(persistentRepository)

		proxy.token = &auth.Token{
			AccessToken: "this is a different but valid access token",
			Expiry:      time.Now().Add(1 * time.Hour),
		}

		//
		// When we try to fetch the token from the proxy
		token, err := proxy.FetchToken(t.Context())

		// Then no error should be reported
		require.NoError(t, err)

		// And the token should match the one already on memory
		require.Equal(t, proxy.token.AccessToken, token.AccessToken)
		require.Equal(t, proxy.token.Expiry, token.Expiry)

		// And the tokens should not be the same
		require.NotSame(t, token, proxy.token)
	})
}

func TestTokenProxy_SaveToken(t *testing.T) {
	t.Run("should save the token on the persistent repository", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a persistent repository which does not have any token
		persistentRepository := NewMockTokenRepository(ctrl)

		var tokenPtr **auth.Token

		persistentRepository.EXPECT().SaveToken(gomock.AssignableToTypeOf(t.Context()), gomock.AssignableToTypeOf(&auth.Token{})).DoAndReturn(
			func(ctx context.Context, token *auth.Token) error {
				tokenPtr = &token
				return nil
			}).Times(1)

		//
		// And a proxy using that last and that also contains no token yet
		proxy := NewTokenProxy(persistentRepository)

		// When we try to save a token
		err := proxy.SaveToken(t.Context(), &auth.Token{AccessToken: accessToken, Expiry: expiry})

		// Then no error should be reported
		require.NoError(t, err)

		// And the token on the proxy should match the given data
		require.Equal(t, accessToken, proxy.token.AccessToken)
		require.Equal(t, expiry, proxy.token.Expiry)

		// And the token on the persistent repository should match the given data
		require.Equal(t, accessToken, (*tokenPtr).AccessToken)
		require.Equal(t, expiry, (*tokenPtr).Expiry)

		// And the tokens on both repositories should not be the same
		require.NotSame(t, proxy.token, *tokenPtr)
	})

	t.Run("should not update its token when the persistent repository reports some error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a persistent repository which experience some connection error
		persistentRepository := NewMockTokenRepository(ctrl)

		errConnection := errors.New("connection error")

		persistentRepository.EXPECT().SaveToken(
			gomock.AssignableToTypeOf(t.Context()),
			gomock.AssignableToTypeOf(&auth.Token{}),
		).Return(errConnection).Times(1)

		//
		// And a proxy using that last and that also contains no token yet
		proxy := NewTokenProxy(persistentRepository)

		// When we try to save a token
		err := proxy.SaveToken(t.Context(), &auth.Token{AccessToken: accessToken, Expiry: expiry})

		// Then the same error should be reported
		require.ErrorIs(t, err, errConnection)

		// And the token on the proxy should not be set
		require.Nil(t, proxy.token)

		// And the saveTicket must not have been incremented — a failed persistent
		// write must leave the double-checked-locking sentinel unchanged so that
		// concurrent FetchToken calls are not misled into skipping a re-fetch.
		require.Equal(t, uint64(0), proxy.saveTicket)
	})

	t.Run("should not mark the cache as refreshed when the persistent save fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a persistent repository that fails on save but succeeds on fetch
		persistentRepository := NewMockTokenRepository(ctrl)

		errConnection := errors.New("connection error")

		persistentRepository.EXPECT().SaveToken(
			gomock.AssignableToTypeOf(t.Context()),
			gomock.AssignableToTypeOf(&auth.Token{}),
		).Return(errConnection).Times(1)

		persistentRepository.EXPECT().FetchToken(
			gomock.AssignableToTypeOf(t.Context()),
		).Return(&auth.Token{AccessToken: accessToken, Expiry: expiry}, nil).Times(1)

		//
		// And a proxy with an expired in-memory token
		proxy := NewTokenProxy(persistentRepository)

		proxy.token = &auth.Token{
			AccessToken: "expired token",
			Expiry:      time.Now().Add(-1 * time.Hour),
		}

		// When the save fails
		err := proxy.SaveToken(t.Context(), &auth.Token{AccessToken: "new token", Expiry: expiry})

		// Then the connection error should be reported
		require.ErrorIs(t, err, errConnection)

		// When a subsequent FetchToken is called
		token, err := proxy.FetchToken(t.Context())

		// Then no error should be reported — the fetch must have gone to the
		// persistent store, not served the stale in-memory token, because the
		// failed save must not have bumped saveTicket.
		require.NoError(t, err)

		// And the returned token should match the one from the persistent repository
		require.Equal(t, accessToken, token.AccessToken)
		require.Equal(t, expiry, token.Expiry)
	})

	t.Run("should not be overwriten by fetch calls", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a persistent repository which has a valid token
		persistentRepository := NewMockTokenRepository(ctrl)

		persistentToken := &auth.Token{
			AccessToken: "another access token",
			Expiry:      time.Now().Add(1 * time.Hour),
		}

		tokenPtr := &persistentToken

		persistentRepository.EXPECT().FetchToken(gomock.AssignableToTypeOf(t.Context())).DoAndReturn(
			func(ctx context.Context) (*auth.Token, error) {
				time.Sleep(time.Duration(100+rand.IntN(100)) * time.Microsecond)

				return *tokenPtr, nil
			}).Times(1)

		persistentRepository.EXPECT().SaveToken(gomock.AssignableToTypeOf(t.Context()), gomock.AssignableToTypeOf(&auth.Token{})).DoAndReturn(
			func(ctx context.Context, token *auth.Token) error {
				time.Sleep(time.Duration(100+rand.IntN(100)) * time.Microsecond)

				tokenPtr = &token
				return nil
			}).Times(1)

		//
		// And a proxy using that last and that also contains no token yet
		proxy := NewTokenProxy(persistentRepository)

		// When we try to save a token during a bunch of simultaneous call to fetch the token
		errChan := make(chan error, 1000) // to carry the results of all simultaneous calls

		var wg sync.WaitGroup // to wait all calls to finish

		var bell sync.RWMutex // to make the calls to run closed as possible to be really simultaneous

		bell.Lock() // make sure to block all calls

		wg.Add(1)

		go func() {
			defer wg.Done()

			bell.RLock()
			defer bell.RUnlock()

			err := proxy.SaveToken(t.Context(), &auth.Token{AccessToken: accessToken, Expiry: expiry})

			errChan <- err
		}()

		for range 999 { // launch the simultaneous go routines
			wg.Add(1)

			go func() {
				defer wg.Done()

				bell.RLock()
				defer bell.RUnlock()

				_, err := proxy.FetchToken(t.Context())

				errChan <- err
			}()
		}

		bell.Unlock() // unblock all calls at the same time

		wg.Wait() // wait all calls to be finished

		close(errChan)

		//
		// Then no error should be reported
		for err := range errChan {
			require.NoError(t, err)
		}

		// And the token on the proxy should match the given data
		require.Equal(t, accessToken, proxy.token.AccessToken)
		require.Equal(t, expiry, proxy.token.Expiry)

		// And the token on the persistent repository should match the given data
		require.Equal(t, accessToken, (*tokenPtr).AccessToken)
		require.Equal(t, expiry, (*tokenPtr).Expiry)

		// And the tokens on both repositories should not be the same
		require.NotSame(t, proxy.token, *tokenPtr)
	})
}

func TestTokenProxyWithRandonExpirationDriftSeconds_FetchToken(t *testing.T) {
	t.Run("should apply the drift to its im-memory token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a persistent repository which already has a token
		persistentRepository := NewMockTokenRepository(ctrl)

		persistentRepository.EXPECT().FetchToken(gomock.AssignableToTypeOf(t.Context())).Times(0)

		//
		// And a proxy using that last which already contains a token
		proxy := NewTokenProxyWithRandomExpirationDriftSeconds(persistentRepository, 300)

		proxy.token = &auth.Token{
			AccessToken: accessToken,
			Expiry:      expiry,
		}

		// When we try to fetch the token from the proxy
		token, err := proxy.FetchToken(t.Context())

		// Then no error should be reported
		require.NoError(t, err)

		// And the token should be stored on memory by the proxy
		require.NotNil(t, proxy.token)

		// And the token should match the one returned from the persistent repository
		require.Equal(t, accessToken, token.AccessToken)

		// And the drift should be correctly applied to the returned token
		require.Equal(t, expiry.Add(-1*time.Duration(proxy.expirationDriftSeconds)*time.Second), token.Expiry)

		// And the drift should not be applied to the token in-memory
		require.Equal(t, expiry, proxy.token.Expiry)

		// And the tokens should not be the same
		require.NotSame(t, token, proxy.token)
	})

	t.Run("should apply the drift to the token fetch from the persistent repository", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Given a persistent repository which already has a token
		persistentRepository := NewMockTokenRepository(ctrl)

		persistentRepository.EXPECT().FetchToken(gomock.AssignableToTypeOf(t.Context())).Return(
			&auth.Token{
				AccessToken: accessToken,
				Expiry:      expiry,
			}, nil).Times(1)

		//
		// And a fresh new proxy using that last
		proxy := NewTokenProxyWithRandomExpirationDriftSeconds(persistentRepository, 300)

		// When we try to fetch the token from the proxy
		token, err := proxy.FetchToken(t.Context())

		// Then no error should be reported
		require.NoError(t, err)

		// And the token should be stored on memory by the proxy
		require.NotNil(t, proxy.token)

		// And the token should match the one returned from the persistent repository
		require.Equal(t, accessToken, token.AccessToken)

		// And the drift should be correctly applied to the returned token
		require.Equal(t, expiry.Add(-1*time.Duration(proxy.expirationDriftSeconds)*time.Second), token.Expiry)

		// And the drift should not be applied to the token in-memory
		require.Equal(t, expiry, proxy.token.Expiry)

		// And the tokens should not be the same
		require.NotSame(t, token, proxy.token)
	})
}
