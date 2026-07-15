package redis

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/internal/ports/auth"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	gomock "go.uber.org/mock/gomock"
)

//go:generate mockgen -package redis -destination=zz_mock_redis_test.go github.com/Arubacloud/sdk-go/internal/impl/auth/tokenrepository/redis RedisClient,RedisCmdClient
var (
	accessToken = "this is a valid token"
	expiry      = time.Now().Add(24 * time.Hour)
)

func TestTokenRepository_FetchToken(t *testing.T) {

	t.Run("should report an error when redis connection fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRedis := NewMockRedisClient(ctrl)
		tokenRepository := NewRedisTokenRepository("user-123", mockRedis)

		mockRedis.
			EXPECT().
			Get(gomock.Any(), "user-123").
			Return("", errors.New("redis: connection error"))

		// When we try to fetch the token
		token, err := tokenRepository.FetchToken(t.Context())

		// And no token should be returned
		require.Error(t, err)
		require.Equal(t, err.Error(), "failed to retrieve value from redis: redis: connection error")

		require.Nil(t, token)
	})
	t.Run("should report a token not found error when it has empty token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRedis := NewMockRedisClient(ctrl)
		tokenRepository := NewRedisTokenRepository("user-123", mockRedis)

		// Empty value to simulate no token
		mockRedis.
			EXPECT().
			Get(gomock.Any(), "user-123").
			Return("", auth.ErrTokenNotFound)

		// When we try to fetch the token
		token, err := tokenRepository.FetchToken(t.Context())

		// And no token should be returned
		require.Error(t, err)
		require.ErrorIs(t, err, auth.ErrTokenNotFound)

		require.Nil(t, token)
	})

	t.Run("should report a token not found error when it has no jwt token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRedis := NewMockRedisClient(ctrl)
		tokenRepository := NewRedisTokenRepository("user-123", mockRedis)

		mockRedis.
			EXPECT().
			Get(gomock.Any(), "user-123").
			Return("token", nil)

		// When we try to fetch the token
		token, err := tokenRepository.FetchToken(t.Context())

		// And no token should be returned
		require.Error(t, err)
		require.Nil(t, token)
	})

	t.Run("should report a token when it has a jwt token on redis", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRedis := NewMockRedisClient(ctrl)
		tokenRepository := NewRedisTokenRepository("user-123", mockRedis)

		authToken := &auth.Token{
			AccessToken: accessToken,
			Expiry:      expiry,
		}

		tokenJSON, _ := json.Marshal(authToken)

		mockRedis.
			EXPECT().
			Get(gomock.Any(), "user-123").
			Return(string(tokenJSON), nil)

		// When we try to fetch the token
		token, err := tokenRepository.FetchToken(t.Context())

		// And no token should be returned
		require.NoError(t, err)
		require.NotNil(t, token)
		require.Equal(t, accessToken, token.AccessToken)
		require.Equal(t, expiry.Unix(), token.Expiry.Unix())
	})
}

func TestTokenRepository_SaveToken(t *testing.T) {

	t.Run("should save a token if not nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRedis := NewMockRedisClient(ctrl)

		// And a valid token
		token := &auth.Token{AccessToken: accessToken, Expiry: expiry}
		tokenRepository := NewRedisTokenRepository("user-123", mockRedis)
		tokenJSON, _ := json.Marshal(token)

		mockRedis.
			EXPECT().
			Set(gomock.Any(), "user-123", tokenJSON, gomock.Any()).
			Return(nil)
		// When we try to save the token
		err := tokenRepository.SaveToken(t.Context(), token)

		// Then no error should be reported
		require.NoError(t, err)
	})

	t.Run("should not save token if malformed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRedis := NewMockRedisClient(ctrl)

		// And a valid token
		token := &auth.Token{
			AccessToken: string([]byte{0xff, 0xfe, 0xfd}), // Invalid UTF-8
			Expiry:      expiry,
		}
		tokenRepository := NewRedisTokenRepository("user-123", mockRedis)

		tokenJSON, _ := json.Marshal(token)

		mockRedis.
			EXPECT().
			Set(gomock.Any(), "user-123", tokenJSON, gomock.Any()).
			Return(errors.New("malformed token"))
		// When we try to

		// When we try to save the token
		err := tokenRepository.SaveToken(t.Context(), token)

		// Then no error should be reported
		require.Error(t, err)
		require.Equal(t, "malformed token", err.Error())
	})

}

func TestAdapter_Get(t *testing.T) {

	t.Run("should return error when redis GET fails", func(t *testing.T) {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRedis := NewMockRedisCmdClient(ctrl)

		adapter := NewRedisAdapter(mockRedis)

		cmd := redis.NewStringCmd(t.Context())
		cmd.SetErr(redis.Nil)
		mockRedis.
			EXPECT().
			Get(gomock.Any(), "my-key").
			Return(cmd)

		val, err := adapter.Get(t.Context(), "my-key")

		require.ErrorIs(t, err, auth.ErrTokenNotFound)
		require.Empty(t, val)
	})

	t.Run("should return value when redis GET succeeds", func(t *testing.T) {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRedis := NewMockRedisCmdClient(ctrl)

		adapter := NewRedisAdapter(mockRedis)

		cmd := redis.NewStringCmd(t.Context())
		cmd.SetVal("my-value")
		mockRedis.
			EXPECT().
			Get(gomock.Any(), "my-key").
			Return(cmd)

		val, err := adapter.Get(t.Context(), "my-key")

		require.NoError(t, err)
		require.Equal(t, "my-value", val)
	})

	t.Run("should return error when redis GET returns other error", func(t *testing.T) {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRedis := NewMockRedisCmdClient(ctrl)

		adapter := NewRedisAdapter(mockRedis)

		cmd := redis.NewStringCmd(t.Context())
		cmd.SetErr(errors.New("redis connection error"))
		mockRedis.
			EXPECT().
			Get(gomock.Any(), "my-key").
			Return(cmd)

		val, err := adapter.Get(t.Context(), "my-key")

		require.Error(t, err)
		require.Equal(t, "redis connection error", err.Error())
		require.Empty(t, val)
	})
}

func TestAdapter_Set(t *testing.T) {

	t.Run("should return nil when redis SET succeeds", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRedis := NewMockRedisCmdClient(ctrl)
		adapter := NewRedisAdapter(mockRedis)

		mockRedis.
			EXPECT().
			Set(gomock.Any(), "my-key", "my-value", 10*time.Second).
			Return(redis.NewStatusCmd(t.Context()))

		err := adapter.Set(t.Context(), "my-key", "my-value", 10*time.Second)

		require.NoError(t, err)
	})

	t.Run("should return error when redis SET fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRedis := NewMockRedisCmdClient(ctrl)
		adapter := NewRedisAdapter(mockRedis)

		cmd := redis.NewStatusCmd(t.Context())
		cmd.SetErr(errors.New("redis connection error"))

		mockRedis.
			EXPECT().
			Set(gomock.Any(), "my-key", "my-value", 10*time.Second).
			Return(cmd)

		err := adapter.Set(t.Context(), "my-key", "my-value", 10*time.Second)

		require.Error(t, err)
		require.Equal(t, "redis connection error", err.Error())
	})
}
