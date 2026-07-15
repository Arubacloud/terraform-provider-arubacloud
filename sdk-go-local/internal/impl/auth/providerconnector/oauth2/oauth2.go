// Package oauth2 provides an implementation of the auth.ProviderConnector interface
// using the Client Credentials Flow (RFC 6749).
// It acts as a bridge between the SDK's internal authentication ports and the
// standard golang.org/x/oauth2 library.
package oauth2

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/Arubacloud/sdk-go/internal/ports/auth"
)

// ProviderConnector implements auth.ProviderConnector.
// It uses the OAuth2 Client Credentials grant type to exchange a Client ID
// and Client Secret for an access token.
type ProviderConnector struct {
	credentialsRepository auth.CredentialsRepository
	tokenURL              string
	scopes                []string
}

var _ auth.ProviderConnector = (*ProviderConnector)(nil)

// NewProviderConnector creates a new connector instance.
// tokenURL is the specific endpoint of the Identity Provider (IdP).
// scopes are the permissions requested for the token.
func NewProviderConnector(credentialsRepository auth.CredentialsRepository, tokenURL string, scopes []string) *ProviderConnector {
	return &ProviderConnector{
		credentialsRepository: credentialsRepository,
		tokenURL:              tokenURL,
		scopes:                scopes,
	}
}

// RequestToken retrieves the credentials from the repository and exchanges
// them for a new OAuth2 token using the configured IdP.
func (c *ProviderConnector) RequestToken(ctx context.Context) (*auth.Token, error) {
	credentials, err := c.credentialsRepository.FetchCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credentials for oauth2 exchange: %w", err)
	}

	credentials = credentials.Copy()

	oauth2Config := clientcredentials.Config{
		ClientID:     credentials.ClientID,
		ClientSecret: credentials.ClientSecret,
		TokenURL:     c.tokenURL,
		Scopes:       c.scopes,
	}

	oauth2Token, err := oauth2Config.Token(ctx)
	if err != nil {
		return nil, wrapOAuth2Error(err)
	}

	return &auth.Token{
		AccessToken: oauth2Token.AccessToken,
		Expiry:      oauth2Token.Expiry,
	}, nil
}

// wrapOAuth2Error inspects errors returned by the oauth2 library.
// It attempts to unwrap HTTP errors (like 401/403) and map them to
// domain-specific errors defined in the auth package.
func wrapOAuth2Error(err error) error {
	retrieveError := &oauth2.RetrieveError{}

	if errors.As(err, &retrieveError) {
		if retrieveError.Response != nil {
			switch retrieveError.Response.StatusCode {
			case http.StatusUnauthorized:
				// Map 401 Unauthorized -> ErrAuthenticationFailed
				return fmt.Errorf("%w: %w", auth.ErrAuthenticationFailed, err)

			case http.StatusForbidden:
				// Map 403 Forbidden -> ErrInsufficientPrivileges
				return fmt.Errorf("%w: %w", auth.ErrInsufficientPrivileges, err)
			}
		}
	}

	return err
}
