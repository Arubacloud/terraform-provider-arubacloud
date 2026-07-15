// Package auth defines the core interfaces and types for handling authentication
// within the SDK. It establishes contracts for token management, storage,
// and retrieval without binding to specific implementations.
package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/Arubacloud/sdk-go/internal/ports/interceptor"
)

var (
	ErrInvalidInterceptable   = errors.New("invalid interceptable")
	ErrTokenNotFound          = errors.New("token not found")
	ErrCredentialsNotFound    = errors.New("credentials not found")
	ErrAuthenticationFailed   = errors.New("authentication failed")
	ErrInsufficientPrivileges = errors.New("insufficient privileges")
)

// TokenManager defines the behavior for a component that manages the lifecycle
// of authentication tokens. It is responsible for ensuring a valid token exists
// and injecting it into outgoing HTTP requests.
type TokenManager interface {
	// BindTo registers the TokenManager's injection logic (InjectToken) with
	// a request interceptor. This allows the manager to act as middleware.
	BindTo(interceptable interceptor.Interceptable) error

	// InjectToken is an interceptor function that retrieves a valid token
	// (refreshing it if necessary) and adds it to the HTTP request headers.
	InjectToken(ctx context.Context, r *http.Request) error
}

// TokenRepository defines the contract for persisting and retrieving tokens.
// This allows for caching strategies (memory, disk, Redis, etc.) to be swapped easily.
type TokenRepository interface {
	// FetchToken retrieves the current token from storage.
	// It should return ErrTokenNotFound if no token is currently stored.
	FetchToken(ctx context.Context) (*Token, error)

	// SaveToken persists the provided token to storage.
	SaveToken(ctx context.Context, token *Token) error
}

// ProviderConnector defines the contract for communicating with the external
// identity provider (IdP). Its sole responsibility is fetching a *fresh* token.
type ProviderConnector interface {
	// RequestToken performs the actual remote call to the authentication server
	// to obtain a new token.
	RequestToken(ctx context.Context) (*Token, error)
}

// CredentialsRepository defines the contract for retrieving static application credentials.
type CredentialsRepository interface {
	// FetchCredentials retrieves the Client ID and Secret needed to request a token.
	FetchCredentials(ctx context.Context) (*Credentials, error)
}
