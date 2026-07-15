// Package memory provides an in-memory implementation of the auth.TokenRepository.
// It serves two primary purposes:
//  1. A standalone in-memory store for tokens (volatile storage).
//  2. A caching proxy over a persistent repository (e.g., Redis/File), reducing
//     access to the slower storage and handling "thundering herd" protection.
package memory

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/Arubacloud/sdk-go/internal/ports/auth"
)

// TokenRepository is a thread-safe, in-memory token storage.
// It supports acting as a proxy to a persistent layer and adds "drift" calculation
// to artificially shorten token lifespan for safety.
type TokenRepository struct {
	// token is the cached in-memory copy.
	token *auth.Token

	// locker guards access to the cached token and ticket counters.
	locker sync.RWMutex

	// fetchTicket and saveTicket are counters used for cache invalidation
	// and "double-checked locking". They ensure that if multiple goroutines
	// try to refresh the cache simultaneously, only one does the work.
	fetchTicket uint64
	saveTicket  uint64

	// persistentRepository is the optional underlying storage.
	// If set, Fetch/Save operations are delegated here on cache misses.
	persistentRepository auth.TokenRepository

	// expirationDriftSeconds is a safety buffer subtracted from the token's
	// expiry time. This helps prevent using a token that might expire
	// while the HTTP request is in flight.
	expirationDriftSeconds uint32
}

var _ auth.TokenRepository = (*TokenRepository)(nil)

// NewTokenRepository creates a standalone in-memory repository.
// Tokens stored here are lost when the application restarts.
func NewTokenRepository() *TokenRepository {
	return &TokenRepository{}
}

// NewTokenRepository creates a standalone in-memory repository with a
// preloaded access token.
// As the internal token expiry time is not set, the token repository will
// never consider the token as expired.
// Tokens stored here are lost when the application restarts.
func NewTokenRepositoryWithAccessToken(accessToken string) *TokenRepository {
	return &TokenRepository{
		token: &auth.Token{AccessToken: accessToken},
	}
}

// NewTokenProxy creates a repository that caches tokens in memory but
// delegates to a persistentRepository on cache misses or saves.
func NewTokenProxy(persistentRepository auth.TokenRepository) *TokenRepository {
	return &TokenRepository{
		persistentRepository: persistentRepository,
	}
}

// NewTokenProxyWithRandomExpirationDriftSeconds creates a proxy repository with
// randomized expiration drift.
//
// The maxExpirationDriftSeconds parameter defines the upper bound for the random jitter.
//
// This is useful to prevent all instances of an application from refreshing tokens
// at the exact same millisecond.
func NewTokenProxyWithRandomExpirationDriftSeconds(persistentRepository auth.TokenRepository, maxExpirationDriftSeconds uint32) *TokenRepository {
	return &TokenRepository{
		persistentRepository: persistentRepository,

		// 1 + rand... ensures at least 1 second of safety buffer.
		expirationDriftSeconds: 1 + rand.Uint32N(maxExpirationDriftSeconds),
	}
}

// FetchToken retrieves the token from memory. If missing or invalid, it attempts
// to fetch from the persistent repository (if configured) and updates the cache.
func (r *TokenRepository) FetchToken(ctx context.Context) (*auth.Token, error) {
	// Step 1: Fast Path (Read Lock)
	r.locker.RLock()

	// Capture the state of the system when we started looking.
	currentFetchTicket := r.fetchTicket
	currentSaveTicket := r.saveTicket

	var tokenCopy *auth.Token

	if r.token != nil {
		// Create a copy with the safety drift applied.
		tokenCopy = r.tokenCopyWithDrift()
	}

	r.locker.RUnlock()

	// If we found a valid token in memory, return it immediately.
	if tokenCopy != nil && tokenCopy.IsValid() {
		return tokenCopy, nil
	}

	// If we don't have a persistent repo to fall back on:
	if r.persistentRepository == nil {
		// Return the expired token if we have one (caller might inspect it),
		// otherwise return the NotFound error.
		if tokenCopy != nil {
			return tokenCopy, nil
		}
		return nil, auth.ErrTokenNotFound
	}

	// Step 2: Slow Path (Write Lock)
	// The token is missing or expired in memory. We need to check the persistent store.
	r.locker.Lock()
	defer r.locker.Unlock()

	// Double-Checked Locking:
	// We check if the tickets match what we saw during the Read Lock.
	// If they differ, it means another goroutine already fetched or saved a new
	// token while we were waiting for the Lock. In that case, we skip the fetch.
	if currentFetchTicket == r.fetchTicket && currentSaveTicket == r.saveTicket {
		token, err := r.persistentRepository.FetchToken(ctx)
		if err != nil {
			return nil, err
		}

		// Increment the fetch ticket to signal other waiting readers that
		// the cache has been refreshed.
		r.fetchTicket++

		// Update the in-memory cache.
		// Assumes auth.Token has a .Copy() method to prevent reference sharing.
		r.token = token.Copy()
	}

	// Return the newly cached token (with drift applied).
	return r.tokenCopyWithDrift(), nil
}

// SaveToken persists the token. It writes through to the persistent repository
// (if one exists) and updates the in-memory cache.
func (r *TokenRepository) SaveToken(ctx context.Context, token *auth.Token) error {
	r.locker.Lock()
	defer r.locker.Unlock()

	// Write-Through: Save to persistent storage first. If it fails, the
	// in-memory cache and saveTicket must remain unchanged so that concurrent
	// FetchToken calls continue to observe the prior-consistent state and do
	// not skip the persistent re-fetch due to a spurious ticket change.
	if r.persistentRepository != nil {
		if err := r.persistentRepository.SaveToken(ctx, token.Copy()); err != nil {
			return err
		}
	}

	// Persistent write succeeded (or there is no persistent layer).
	// Bump saveTicket to invalidate any in-flight FetchToken double-checks
	// and update the in-memory cache.
	r.saveTicket++
	r.token = token.Copy()

	return nil
}

func (r *TokenRepository) tokenCopyWithDrift() *auth.Token {
	tokenCopy := r.token.Copy()

	if r.expirationDriftSeconds > 0 {
		tokenCopy.Expiry = tokenCopy.Expiry.Add(-1 * time.Duration(r.expirationDriftSeconds) * time.Second)
	}

	return tokenCopy
}
