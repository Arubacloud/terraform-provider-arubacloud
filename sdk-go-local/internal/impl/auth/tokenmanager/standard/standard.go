// Package standard provides concrete implementations of the core SDK interfaces.
package standard

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/Arubacloud/sdk-go/internal/ports/auth"
	"github.com/Arubacloud/sdk-go/internal/ports/interceptor"
)

// TokenManager is the standard implementation of auth.TokenManager.
// It handles token storage, retrieval, and automatic refreshing using a
// thread-safe mechanism to prevent race conditions during updates.
type TokenManager struct {
	connector  auth.ProviderConnector
	repository auth.TokenRepository

	// locker guards access to the token storage logic and the ticket counter.
	locker sync.RWMutex
	// ticket is a counter used to detect if a token refresh has occurred
	// between the time a read lock was released and a write lock was acquired.
	ticket uint64
}

// Verify at compile-time that TokenManager implements auth.TokenManager.
var _ auth.TokenManager = (*TokenManager)(nil)

// NewTokenManager creates a new instance of TokenManager with the provided
// repository (for caching) and connector (for fetching fresh tokens).
func NewTokenManager(connector auth.ProviderConnector, repository auth.TokenRepository) *TokenManager {
	return &TokenManager{
		repository: repository,
		connector:  connector,
	}
}

// NewTokenManager creates a new instance of TokenManager without a provider
// connector.
func NewStaticTokenManager(repository auth.TokenRepository) *TokenManager {
	return &TokenManager{repository: repository}
}

// BindTo registers the InjectToken method as a callback function within the
// provided Interceptable (e.g., an HTTP client middleware chain).
func (m *TokenManager) BindTo(interceptable interceptor.Interceptable) error {
	if interceptable == nil {
		return fmt.Errorf("%w: not possible to bind to a nil interceptable", auth.ErrInvalidInterceptable)
	}

	return interceptable.Bind(m.InjectToken)
}

// InjectToken retrieves a valid token and adds it to the request "Authorization" header.
//
// Logic Flow:
//  1. Optimistically tries to read a valid token from the repository (Read Lock).
//  2. If the token is missing or expired, it upgrades to a Write Lock.
//  3. It uses a "ticket" system to ensure only one goroutine performs the refresh
//     (preventing the "thundering herd" problem), while others simply wait and
//     use the newly refreshed token.
func (m *TokenManager) InjectToken(ctx context.Context, r *http.Request) error {
	// Step 1: Optimistic Read
	m.locker.RLock()

	// Capture the current "version" (ticket) of the token state.
	currentTicket := m.ticket

	token, err := m.repository.FetchToken(ctx)
	if err != nil && !errors.Is(err, auth.ErrTokenNotFound) {
		m.locker.RUnlock()
		return fmt.Errorf("unexpected error: %w", err)
	}

	m.locker.RUnlock()

	// Step 2: Validation & Refresh (if needed)
	// If we have no token, or the token is invalid/expired, we need to refresh.
	if m.connector != nil && (errors.Is(err, auth.ErrTokenNotFound) || !token.IsValid()) {
		m.locker.Lock()
		defer m.locker.Unlock()

		// Double-Checked Locking with Ticket:
		// Check if the ticket hasn't changed since we released the read lock.
		// If currentTicket == m.tokenTicket, it means we are the first to grab the
		// write lock, and we are responsible for the refresh.
		if currentTicket == m.ticket {
			// Increment ticket so pending readers know a change happened.
			m.ticket++

			token, err = m.connector.RequestToken(ctx)
			if err != nil {
				return fmt.Errorf("unexpected error: %w", err)
			}

			if err := m.repository.SaveToken(ctx, token); err != nil {
				return fmt.Errorf("unexpected error: %w", err)
			}
		} else {
			// If the tickets don't match, another goroutine already performed the
			// refresh while we were waiting for the lock.
			// We can simply fetch the new token from the repository.
			token, err = m.repository.FetchToken(ctx)
			if err != nil && !errors.Is(err, auth.ErrTokenNotFound) {
				return fmt.Errorf("unexpected error: %w", err)
			}
		}
	}

	// Step 3: Injection
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	return nil
}
