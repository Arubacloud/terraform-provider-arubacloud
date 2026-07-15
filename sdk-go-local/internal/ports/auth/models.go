package auth

import (
	"time"
)

// Token represents the credentials used to authorize requests to access
// protected resources on the Aruba backend.
type Token struct {
	// AccessToken is the raw string token (usually a JWT or opaque string)
	// that authorizes and authenticates the requests.
	AccessToken string `json:"access_token"`

	// Expiry is the optional expiration time of the access token.
	// If zero, the token is assumed to never expire (or expiration is not tracked).
	Expiry time.Time `json:"expiry,omitempty"`
}

// IsValid checks if the token is usable.
// A token is considered valid if:
// 1. It has no expiration time set (IsZero).
// 2. The expiration time is strictly in the future.
func (t *Token) IsValid() bool {
	if t.Expiry.IsZero() {
		return true
	}

	if t.Expiry.After(time.Now()) {
		return true
	}

	return false
}

// Copy creates a copy of the Token and returns its reference.
func (t *Token) Copy() *Token {
	return &Token{
		AccessToken: t.AccessToken,
		Expiry:      t.Expiry,
	}
}

// Credentials holds the static authentication details required to obtain a Token.
type Credentials struct {
	// ClientID is the application's public identifier.
	ClientID string `json:"client_id"`

	// ClientSecret is the application's private secret.
	ClientSecret string `json:"client_secret"`
}

// Copy creates a copy of the Credentials and returns its reference.
func (c *Credentials) Copy() *Credentials {
	return &Credentials{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
	}
}
