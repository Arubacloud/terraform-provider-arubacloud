// Package name for the Redis implementation of the token repository.
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Arubacloud/sdk-go/internal/ports/auth"
	"github.com/redis/go-redis/v9"
)

// TokenRepository implements auth.TokenRepository using Redis as backend.
// It stores and retrieves tokens associated with a specific client ID.
type TokenRepository struct {
	redisClient RedisClient
	clientID    string
}

// RedisAdapter is a thin wrapper over Redis that exposes a simple,
// domain-friendly API (Get/Set returning plain values and errors).
//
// It does NOT expose go-redis command types to the outside world.
type RedisAdapter struct {
	client RedisCmdClient
}

// RedisCmdClient defines the minimal Redis operations used by RedisAdapter.
// This allows us to:
//   - decouple the adapter from *redis.Client
//   - inject a fake or mock implementation during tests
//   - simplify testing and improve maintainability
type RedisCmdClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
}

var _ auth.TokenRepository = (*TokenRepository)(nil)
var _ RedisClient = (*RedisAdapter)(nil)

// RedisClient defines minimal Redis operations used by TokenRepository.
// This interface uses basic Go types (string, error) and completely abstracts the underlying implementation.
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
}

// NewRedisTokenRepository is the constructor for TokenRepository.
// It accepts the client ID and the decoupled RedisClient interface.
func NewRedisTokenRepository(clientID string, client RedisClient) *TokenRepository {
	return &TokenRepository{
		redisClient: client,
		clientID:    clientID,
	}
}

// NewRedisAdapter creates a new RedisAdapter wrapping the given RedisCmdClient.
// This is the factory function for the Adapter.
func NewRedisAdapter(client RedisCmdClient) *RedisAdapter {
	return &RedisAdapter{client: client}
}

// FetchToken retrieves the token from Redis for the given client ID.
// Returns ErrTokenNotFound if no token exists, or JSON decoding error if unmarshaling fails.
func (tr *TokenRepository) FetchToken(ctx context.Context) (*auth.Token, error) {
	// Retrieve raw string value from Redis
	val, err := tr.redisClient.Get(ctx, tr.clientID)

	// If the specific domain error is returned, don't wrap it further.
	if errors.Is(err, auth.ErrTokenNotFound) {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve value from redis: %w", err)
	}

	// Decode JSON into auth.Token
	var token auth.Token
	if err := json.Unmarshal([]byte(val), &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token JSON: %w", err)
	}

	return &token, nil
}

// SaveToken stores a token in Redis with TTL equal to the remaining time until token expiry.
func (tr *TokenRepository) SaveToken(ctx context.Context, token *auth.Token) error {
	if token == nil {
		return fmt.Errorf("token cannot be nil")
	}

	// Encode token as JSON
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return err
	}

	// Calculate TTL based on token expiration; ensure non-negative
	// set a minimum of 1 second to avoid zero or negative TTL
	ttl := max(time.Until(token.Expiry), 1*time.Second)

	// Save to Redis
	err = tr.redisClient.Set(ctx, tr.clientID, tokenJSON, ttl)

	return err
}

// Get retrieves a key from Redis and returns a decoded string.
//
// Behavior:
//   - redis.Nil  → return ErrTokenNotFound (domain consistent)
//   - any other error → return the error
//   - success → return the string value
func (a *RedisAdapter) Get(ctx context.Context, key string) (string, error) {
	val, err := a.client.Get(ctx, key).Result()

	// 1. Handle the specific "key not found" error from go-redis
	if err == redis.Nil {
		// Return an empty string and nil error, or a specific domain error
		// like "key not found" depending on how you want the repository to handle it.
		// For the TokenRepository, this error check is now shifted back to the repository logic.
		return "", auth.ErrTokenNotFound // Return the redis.Nil error so the FetchToken method can check for it.
	}

	// 2. Handle all other errors (connection, server issues)
	if err != nil {
		return "", err
	}

	// 3. Return the successful string value
	return val, nil
}

// Set stores a value in Redis using the provided TTL.
// Returns any error returned by Redis.
//
// If expiration == 0, Redis will store the key without TTL.
func (a *RedisAdapter) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	cmd := a.client.Set(ctx, key, value, expiration)
	// Simply return any error from the command execution
	return cmd.Err()
}
