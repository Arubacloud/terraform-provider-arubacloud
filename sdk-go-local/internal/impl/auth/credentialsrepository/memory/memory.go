// Package memory provides an in-memory implementation of the auth.CredentialsRepository.
// It is designed to hold static credentials (passed at startup) or act as a
// lazy-loading cache over a persistent storage mechanism (e.g., loading from a file or vault).
package memory

import (
	"context"
	"sync"

	"github.com/Arubacloud/sdk-go/internal/ports/auth"
)

// CredentialsRepository is a thread-safe storage for application credentials.
// It supports two modes:
// 1. Static: Credentials are provided at initialization and held in memory.
// 2. Proxy: Credentials are lazy-loaded from a persistent repository on the first request.
type CredentialsRepository struct {
	// credentials holds the cached Client ID and Secret.
	credentials *auth.Credentials

	// locker guards access to the credentials pointer to ensure thread safety
	// during concurrent reads and lazy-loading updates.
	locker sync.RWMutex

	// persistentRepository is the optional underlying storage.
	// If set, it is used to fetch credentials if they are not yet in memory.
	persistentRepository auth.CredentialsRepository
}

var _ auth.CredentialsRepository = (*CredentialsRepository)(nil)

// NewCredentialsRepository creates a repository with static, pre-defined credentials.
// This is useful when the credentials are known at startup (e.g., from environment variables).
func NewCredentialsRepository(clientID string, clientSecret string) *CredentialsRepository {
	return &CredentialsRepository{
		credentials: &auth.Credentials{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		},
	}
}

// NewCredentialsProxy creates a repository that acts as a caching layer.
// It does not hold credentials initially; it fetches them from the persistentRepository
// only when FetchCredentials is first called (Lazy Loading).
func NewCredentialsProxy(persistentRepository auth.CredentialsRepository) *CredentialsRepository {
	return &CredentialsRepository{
		persistentRepository: persistentRepository,
	}
}

// FetchCredentials retrieves the Client ID and Secret.
// It employs a "Lazy Loading" strategy:
// 1. Checks memory (fast path).
// 2. If missing, locks and fetches from the persistent store (slow path).
func (r *CredentialsRepository) FetchCredentials(ctx context.Context) (*auth.Credentials, error) {
	// Step 1: Fast Path (Read Lock)
	r.locker.RLock()

	var credentialsCopy *auth.Credentials

	// If we have credentials in memory, prepare a copy to return.
	// Assumes auth.Credentials has a .Copy() method to prevent external mutation.
	if r.credentials != nil {
		credentialsCopy = r.credentials.Copy()
	}

	r.locker.RUnlock()

	if credentialsCopy != nil {
		return credentialsCopy, nil
	}

	// Step 2: Lazy Load (Write Lock)
	//
	// If credentials are nil (cache miss) and we have a backing repo, fetch them.
	if r.persistentRepository != nil {
		r.locker.Lock()
		defer r.locker.Unlock()

		credentials, err := r.persistentRepository.FetchCredentials(ctx)
		if err != nil {
			return nil, err
		}

		r.credentials = credentials.Copy()

		credentialsCopy = credentials.Copy()
	}

	return credentialsCopy, nil
}
