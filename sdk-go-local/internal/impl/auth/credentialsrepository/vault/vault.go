package vault

import (
	"context"
	"fmt"
	"sync"
	"time"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/Arubacloud/sdk-go/internal/ports/auth"
)

// CredentialsRepository implements auth.CredentialsRepository and is
// responsible for authenticating to Vault (via AppRole) and fetching
// credential secrets from a KVv2 backend.
//
// Thread-safety:
// * loginWithAppRole uses a write-lock to prevent concurrent token refresh.
// * FetchCredentials is safe to call concurrently.
type CredentialsRepository struct {
	// Implementation details would go here
	client    VaultClient
	kvMount   string
	kvPath    string
	namespace string
	rolePath  string
	roleID    string
	secretID  string

	tokenExist bool
	renewable  bool
	expiration time.Time
	ttl        time.Duration

	mu sync.Mutex // protects tokenExist, expiration, ttl, renewable
}

var _ auth.CredentialsRepository = (*CredentialsRepository)(nil)

// Tokens will be renewed when within this duration of expiration.
const renewTokenBeforeExpirationDuration = 20 * time.Second

// VaultClient is the main interface abstracting the *vaultapi.Client.
// This allows the business logic (CredentialsRepository) to be decoupled from
// the specific Vault SDK implementation, enabling dependency injection and testing.
type VaultClient interface {
	Logical() LogicalAPI
	SetToken(token string)
	KVv2(mount string) KvAPI
	SetNamespace(namespace string)
}

// VaultClientAdapter wraps *vaultapi.Client to conform to the VaultClient interface.
// This is the Adapter pattern in action.
type VaultClientAdapter struct {
	c *vaultapi.Client
}

// logicalAPIAdapter adapts *vaultapi.Logical for our logical API interface.
type logicalAPIAdapter struct {
	l *vaultapi.Logical
}

// kvAPIAdapter adapts *vaultapi.KVv2.
type kvAPIAdapter struct {
	kv *vaultapi.KVv2
}

// LogicalAPI exposes only the methods we actually need from Vault's logical API.
type LogicalAPI interface {
	Write(path string, data map[string]any) (*vaultapi.Secret, error)
}

// KvAPI exposes only the methods we actually need from Vault's KVv2 secrets engine.
type KvAPI interface {
	Get(ctx context.Context, path string) (*vaultapi.KVSecret, error)
}

// NewVaultClientAdapter returns a new adapter around *vaultapi.Client.
func NewVaultClientAdapter(c *vaultapi.Client) *VaultClientAdapter {
	return &VaultClientAdapter{c: c}
}

// NewCredentialsRepository creates a new CredentialsRepository that fetches
// credentials from a Vault backend.
func NewCredentialsRepository(
	client VaultClient,
	kvMount string,
	kvPath string,
	namespace string,
	rolePath string,
	roleID string,
	secretID string,
) *CredentialsRepository {
	return &CredentialsRepository{
		client:     client,
		kvMount:    kvMount,
		kvPath:     kvPath,
		namespace:  namespace,
		rolePath:   rolePath,
		roleID:     roleID,
		secretID:   secretID,
		tokenExist: false,
		renewable:  false,
		expiration: time.Time{},
		ttl:        0,
	}
}

// isTokenExpired returns true if the token should be refreshed.
// expiration==zero -> treat as expired.
func isTokenExpired(expiration time.Time) bool {
	// Check if the current time is after the expiration time minus the renew duration
	if expiration.IsZero() {
		return true
	}
	// Renew before the expiration threshold
	checkDate := expiration.Add(-renewTokenBeforeExpirationDuration)
	return time.Now().After(checkDate)
}

// FetchCredentials ensures an authenticated Vault token is available,
// then fetches client credentials from KVv2.
func (r *CredentialsRepository) FetchCredentials(ctx context.Context) (*auth.Credentials, error) {
	// Ensure we have a valid token (AppRole login)
	if err := r.ensureAuthenticated(); err != nil {
		return nil, auth.ErrAuthenticationFailed
	}

	// Read from KV v2
	secret, err := r.client.KVv2(r.kvMount).Get(ctx, r.kvPath)
	if err != nil {
		return nil, auth.ErrCredentialsNotFound
	}

	// Extract credentials from the Vault secret
	return getCredentialsFromVaultSecret(secret)
}

// loginWithAppRole performs AppRole authentication to obtain a Vault token. Only one
// goroutine can perform login at a time.
func (r *CredentialsRepository) ensureAuthenticated() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If we already have a token and it's not expired, do nothing.
	if r.tokenExist && !isTokenExpired(r.expiration) {
		return nil
	}

	// Namespace support (Enterprise Vault)
	token, err := r.performAppRoleLogin()
	if err != nil {
		return auth.ErrAuthenticationFailed
	}

	// Validate token response
	if token == nil || token.Auth == nil || token.Auth.ClientToken == "" {
		return auth.ErrAuthenticationFailed
	}

	clientToken := token.Auth.ClientToken

	r.client.SetToken(clientToken)
	r.tokenExist = true

	r.ttl = time.Duration(token.Auth.LeaseDuration) * time.Second
	// Compute expiration time
	r.expiration = time.Now().Add(r.ttl)
	// Track whether the token is renewable
	r.renewable = token.Auth.Renewable

	return nil
}

// performAppRoleLogin executes the AppRole login against Vault and returns the token secret.
func (r *CredentialsRepository) performAppRoleLogin() (*vaultapi.Secret, error) {
	if r.namespace != "" {
		r.client.SetNamespace(r.namespace)
	}

	payload := map[string]any{
		"role_id":   r.roleID,
		"secret_id": r.secretID,
	}

	// Perform AppRole auth
	appRoleURI := fmt.Sprintf("auth/%v/login", r.rolePath)
	token, err := r.client.Logical().Write(appRoleURI, payload)
	return token, err
}

// getCredentialsFromVaultSecret extracts Client ID and Secret from a Vault KV secret.
// It assumes the secret data contains "client_id" and "client_secret" keys.
func getCredentialsFromVaultSecret(secret *vaultapi.KVSecret) (*auth.Credentials, error) {
	get := func(key string) (string, error) {
		v, ok := secret.Data[key].(string)
		if !ok {
			return "", auth.ErrCredentialsNotFound
		}
		return v, nil
	}

	clientID, err := get("client-id")
	if err != nil {
		return nil, err
	}

	clientSecret, err := get("client-secret")
	if err != nil {
		return nil, err
	}

	creds := &auth.Credentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	return creds, nil
}

// ---------------- Adapter Implementations ----------------

// VaultClientAdapter methods
func (v *VaultClientAdapter) SetNamespace(namespace string) {
	v.c.SetNamespace(namespace)
}

func (v *VaultClientAdapter) SetToken(token string) {
	v.c.SetToken(token)
}

func (v *VaultClientAdapter) Logical() LogicalAPI {
	return &logicalAPIAdapter{l: v.c.Logical()}
}
func (v *VaultClientAdapter) KVv2(mount string) KvAPI {
	return &kvAPIAdapter{kv: v.c.KVv2(mount)}
}

// KvAPIAdapter methods
func (k *kvAPIAdapter) Get(ctx context.Context, path string) (*vaultapi.KVSecret, error) {
	return k.kv.Get(ctx, path)
}

// LogicalAPIAdapter methods
func (la *logicalAPIAdapter) Write(path string, data map[string]any) (*vaultapi.Secret, error) {
	return la.l.Write(path, data)
}
