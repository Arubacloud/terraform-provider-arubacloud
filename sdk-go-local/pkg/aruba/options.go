package aruba

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Arubacloud/sdk-go/internal/ports/interceptor"
	"github.com/Arubacloud/sdk-go/internal/ports/logger"
)

// Options is the configuration builder for the Aruba Cloud Client.
// It uses a fluent API pattern to chain configuration settings.
type Options struct {
	// baseURL is the root URL for Aruba REST API calls.
	baseURL string

	// loggerType indicates the logging strategy.
	loggerType LoggerType

	// tokenManager contains authentication-specific settings.
	tokenManager tokenManagerOptions

	// userDefinedDependencies contains injected components.
	userDefinedDependencies userDefinedDependenciesOptions

	// userAgent overrides the default User-Agent header value.
	// When empty, defaultUserAgent is used.
	userAgent string
}

func (o *Options) validate() error {
	var errs []error

	if err := validateURL(o.baseURL, "base API URL"); err != nil {
		errs = append(errs, err)
	}

	if err := o.loggerType.validate(); err != nil {
		errs = append(errs, err)
	}
	if o.loggerType == loggerCustom && o.userDefinedDependencies.logger == nil {
		errs = append(
			errs,
			errors.New(
				"logger type is set to 'Custom' but no custom logger implementation was provided via WithCustomLogger()",
			),
		)
	}

	if err := o.tokenManager.validate(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

//
// Logger Options

// LoggerType defines the supported logging strategies.
type LoggerType int

const (
	// LoggerNoLog disables all SDK logging.
	LoggerNoLog LoggerType = iota
	// LoggerNative uses the SDK's built-in standard library logger.
	LoggerNative
	// loggerCustom indicates a user-provided logger implementation is in use.
	// This is set automatically when WithCustomLogger is called.
	loggerCustom
)

func (l LoggerType) validate() error {
	if l < LoggerNoLog || l > loggerCustom {
		return fmt.Errorf("unsupported logger type: %d", l)
	}

	return nil
}

//
// Token Manager (Authentication) Options

// tokenManagerOptions holds internal configuration for the authentication subsystem.
type tokenManagerOptions struct {
	// token holds a string-encoded JWT OAuth2 access token.
	// This token is not aimed to be renewed: once expired, the current client
	// should be discarded and a new one should be set-up.
	// Mutually exclusive with tokenIssuerOptions.
	token *string

	// tokenIssuerOptions contains the configuration to manage the tokens
	// obtained from an OAuth2 token issuer. It also include the nenewal
	// process.
	// Mutually exclusive with token.
	tokenIssuerOptions *tokenIssuerOptions
}

func (tm *tokenManagerOptions) useTokenIssuer() {
	if tm.tokenIssuerOptions == nil {
		tm.token = nil

		tm.tokenIssuerOptions = &tokenIssuerOptions{}
	}
}

func (tm *tokenManagerOptions) validate() error {
	if tm.token != nil && tm.tokenIssuerOptions != nil {
		return errors.New("configuration conflict: cannot have both Token and Token Issuer set; please choose one")
	}

	if tm.token == nil && tm.tokenIssuerOptions == nil {
		return errors.New("missing token source: must provide either a Token or Token Issuer configuration")
	}

	if tm.token != nil {
		if strings.TrimSpace(*tm.token) == "" {
			return errors.New("invalid token")
		}
	}

	if tm.tokenIssuerOptions != nil {
		return tm.tokenIssuerOptions.validate()
	}

	return nil
}

type tokenIssuerOptions struct {
	// issuerURL is the Aruba OAuth2 token endpoint URL.
	issuerURL string

	// expirationDriftSeconds defines the safety buffer subtracted from
	// a token's expiration time to prevent race conditions.
	// Ignored if no persistent repository proxy is configured.
	expirationDriftSeconds uint32

	// scopes is a list of security scopes to be claimed
	scopes []string

	// clientCredentialOptions contains configuration for direct OAuth2 client
	// credentials authentication.
	// Mutually exclusive with vaultCredentialsRepositoryOptions.
	clientCredentialOptions *clientCredentialOptions

	// vaultCredentialsRepositoryOptions contains configuration for HashiCorp Vault.
	// Mutually exclusive with clientCredentialOptions.
	vaultCredentialsRepositoryOptions *vaultCredentialsRepositoryOptions

	// redisTokenRepositoryOptions contains configuration for a Redis token cache.
	// Mutually exclusive with fileTokenRepositoryOptions.
	redisTokenRepositoryOptions *redisTokenRepositoryOptions

	// fileTokenRepositoryOptions contains configuration for a file-system token cache.
	// Mutually exclusive with redisTokenRepositoryOptions.
	fileTokenRepositoryOptions *fileTokenRepositoryOptions
}

func (ti *tokenIssuerOptions) validate() error {
	var errs []error

	//
	// Basic Fields

	if err := validateURL(ti.issuerURL, "token issuer URL"); err != nil {
		errs = append(errs, err)
	}

	//
	// Credentials Mutual Exclusion & Validity

	hasClientCredentials := ti.clientCredentialOptions != nil
	hasVault := ti.vaultCredentialsRepositoryOptions != nil

	if hasClientCredentials && hasVault {
		errs = append(
			errs,
			errors.New(
				"configuration conflict: cannot use both Client Credentials and Vault Repository for credentials; please choose one",
			),
		)

	} else if !hasClientCredentials && !hasVault {
		errs = append(
			errs,
			errors.New(
				"missing credentials: must provide either a Client Credentials or Vault Repository configuration",
			),
		)

	} else if hasClientCredentials {
		if err := ti.clientCredentialOptions.validate(); err != nil {
			errs = append(errs, fmt.Errorf("client credentials configuration error: %w", err))
		}
	} else if hasVault {
		if err := ti.vaultCredentialsRepositoryOptions.validate(); err != nil {
			errs = append(errs, fmt.Errorf("vault configuration error: %w", err))
		}
	}

	//
	// Token Cache Mutual Exclusion & Validity

	// Note: It is Valid for both Redis and File to be nil: implies no
	// persistence/caching.
	hasRedis := ti.redisTokenRepositoryOptions != nil
	hasFile := ti.fileTokenRepositoryOptions != nil

	if hasRedis && hasFile {
		errs = append(
			errs,
			errors.New(
				"configuration conflict: cannot use both Redis and File System for token caching; please choose one",
			),
		)
	}

	if hasRedis {
		if err := ti.redisTokenRepositoryOptions.validate(); err != nil {
			errs = append(errs, fmt.Errorf("redis configuration error: %w", err))
		}
	}

	if hasFile {
		if err := ti.fileTokenRepositoryOptions.validate(); err != nil {
			errs = append(errs, fmt.Errorf("file repository configuration error: %w", err))
		}
	}

	return errors.Join(errs...)
}

// clientCredentialOptions configures direct OAuth2 Client Credentials
// authentication.
type clientCredentialOptions struct {
	// clientID is the OAuth2 client ID.
	clientID string

	// clientSecret is the OAuth2 client secret.
	clientSecret string
}

func (c *clientCredentialOptions) validate() error {
	var errs []error

	if strings.TrimSpace(c.clientID) == "" {
		errs = append(errs, errors.New("client ID is required"))
	}

	if strings.TrimSpace(c.clientSecret) == "" {
		errs = append(errs, errors.New("client Secret is required"))
	}

	return errors.Join(errs...)
}

// vaultCredentialsRepositoryOptions configures the Vault connection.
type vaultCredentialsRepositoryOptions struct {
	// vaultURI is the address of the Vault server (e.g., "https://vault.example.com:8200").
	vaultURI  string
	kvMount   string
	kvPath    string
	namespace string
	rolePath  string
	roleID    string
	secretID  string
}

func (v *vaultCredentialsRepositoryOptions) validate() error {
	var errs []error

	if err := validateURL(v.vaultURI, "vault URI"); err != nil {
		errs = append(errs, err)
	}

	if strings.TrimSpace(v.roleID) == "" {
		errs = append(errs, errors.New("vault Role ID is required"))
	}

	if strings.TrimSpace(v.secretID) == "" {
		errs = append(errs, errors.New("vault Secret ID is required"))
	}

	if strings.TrimSpace(v.kvMount) == "" {
		errs = append(errs, errors.New("vault KV Mount path is required"))
	}

	if strings.TrimSpace(v.kvPath) == "" {
		errs = append(errs, errors.New("vault KV Secret path is required"))
	}

	return errors.Join(errs...)
}

// redisTokenRepositoryOptions configures the Redis connection.
type redisTokenRepositoryOptions struct {
	// redisURI is the connection string for the Redis cluster.
	// Format: "redis://<user>:<pass>@localhost:6379/<db>"
	redisURI string
}

func (r *redisTokenRepositoryOptions) validate() error {
	u, err := url.ParseRequestURI(r.redisURI)
	if err != nil {
		return fmt.Errorf("invalid redis URI format: %w", err)
	}

	if u.Scheme != "redis" && u.Scheme != "rediss" {
		return fmt.Errorf("invalid redis URI scheme '%s': must be 'redis://' or 'rediss://'", u.Scheme)
	}

	if u.Host == "" {
		return errors.New("invalid redis URI: missing host address")
	}

	return nil
}

// fileTokenRepositoryOptions configures local file storage for tokens.
type fileTokenRepositoryOptions struct {
	// baseDir is the directory path where JSON token files will be stored.
	baseDir string
}

func (f *fileTokenRepositoryOptions) validate() error {
	// We rely on string length.
	// Note: We do not check if the directory exists here (os.Stat) because
	// the application might have permissions to create it later.
	// We only validate that the configuration string is sensible.
	path := strings.TrimSpace(f.baseDir)
	if path == "" {
		return errors.New("base directory path cannot be empty")
	}

	// Simple check for potentially dangerous or invalid paths (optional)
	// Example: prevents root directory usage if desired, though usually ignored in SDKs.
	// if path == "/" { return errors.New("cannot use root directory") }

	return nil
}

//
// User-Defined Dependencies Options

// userDefinedDependenciesOptions holds dependencies injected by the user.
type userDefinedDependenciesOptions struct {
	httpClient *http.Client
	logger     logger.Logger
	middleware interceptor.Interceptor
}

// NewOptions creates a new, empty configuration builder.
func NewOptions() *Options {
	return &Options{}
}

// DeepCopy returns a fully independent copy of the Options.
// Injected dependencies (HTTPClient, Logger, Middleware) are shallow-copied
// because they represent external resources meant to be shared.
func (o *Options) DeepCopy() *Options {
	if o == nil {
		return nil
	}

	cp := &Options{
		baseURL:    o.baseURL,
		loggerType: o.loggerType,
		userAgent:  o.userAgent,
		userDefinedDependencies: userDefinedDependenciesOptions{
			httpClient: o.userDefinedDependencies.httpClient,
			logger:     o.userDefinedDependencies.logger,
			middleware: o.userDefinedDependencies.middleware,
		},
	}

	if o.tokenManager.token != nil {
		t := *o.tokenManager.token
		cp.tokenManager.token = &t
	}

	if o.tokenManager.tokenIssuerOptions != nil {
		ti := o.tokenManager.tokenIssuerOptions
		tiCp := &tokenIssuerOptions{
			issuerURL:              ti.issuerURL,
			expirationDriftSeconds: ti.expirationDriftSeconds,
		}

		if len(ti.scopes) > 0 {
			tiCp.scopes = make([]string, len(ti.scopes))
			copy(tiCp.scopes, ti.scopes)
		}

		if ti.clientCredentialOptions != nil {
			cc := *ti.clientCredentialOptions
			tiCp.clientCredentialOptions = &cc
		}

		if ti.vaultCredentialsRepositoryOptions != nil {
			v := *ti.vaultCredentialsRepositoryOptions
			tiCp.vaultCredentialsRepositoryOptions = &v
		}

		if ti.redisTokenRepositoryOptions != nil {
			r := *ti.redisTokenRepositoryOptions
			tiCp.redisTokenRepositoryOptions = &r
		}

		if ti.fileTokenRepositoryOptions != nil {
			f := *ti.fileTokenRepositoryOptions
			tiCp.fileTokenRepositoryOptions = &f
		}

		cp.tokenManager.tokenIssuerOptions = tiCp
	}

	return cp
}

//
// Basic Options Helpers

// WithBaseURL overrides the default Aruba Cloud API URL.
func (o *Options) WithBaseURL(baseURL string) *Options {
	o.baseURL = baseURL
	return o
}

// WithToken set the OAuth2 JWT Access Token to be used in order to get access
// to the Aruba REST API. This token is not aimed to be renewed by the client,
// so the client needs to be discarded when the token expires.
// Usage of that option is recommended only for atomic short-period scenarios.
// Side Effect: Removes any token issuer configuration previously set.
func (o *Options) WithToken(token string) *Options {
	o.tokenManager.tokenIssuerOptions = nil

	o.tokenManager.token = &token

	return o
}

// WithTokenIssuerURL overrides the default OAuth2 token endpoint.
// Side Effect: Removes the token if previously set.
func (o *Options) WithTokenIssuerURL(tokenIssuerURL string) *Options {
	o.tokenManager.useTokenIssuer()

	o.tokenManager.tokenIssuerOptions.issuerURL = tokenIssuerURL
	return o
}

// WithClientCredentials is a helper to set both Client ID and Secret.
// Side Effect: Removes the token if previously set.
// Side Effect: Disable Vault credentials repository.
func (o *Options) WithClientCredentials(clientID string, clientSecret string) *Options {
	o.tokenManager.useTokenIssuer()

	o.tokenManager.tokenIssuerOptions.vaultCredentialsRepositoryOptions = nil

	o.tokenManager.tokenIssuerOptions.clientCredentialOptions = &clientCredentialOptions{
		clientID:     clientID,
		clientSecret: clientSecret,
	}

	return o
}

// WithLoggerType sets the logging strategy.
// Side Effect: Removes any custom logger previously set.
func (o *Options) WithLoggerType(loggerType LoggerType) *Options {
	o.loggerType = loggerType
	o.userDefinedDependencies.logger = nil
	return o
}

//
// Default Options Values and Helpers

const (
	defaultBaseURL        = "https://api.arubacloud.com"
	defaultLoggerType     = LoggerNoLog
	defaultTokenIssuerURL = "https://mylogin.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token"
)

// DefaultOptions creates a ready-to-use configuration for the production environment
// using Client Credentials.
// Side Effect: Removes the token if previously set.
// Side Effect: Disable the File token repository if previously set.
// Side Effect: Disable the Redis token repository if previously set.
func DefaultOptions(clientID string, clientSecret string) *Options {
	return NewOptions().
		WithDefaultBaseURL().
		WithDefaultLogger().
		WithDefaultTokenManagerSchema(clientID, clientSecret)
}

// WithDefaultBaseURL sets the URL to the production Aruba Cloud API.
func (o *Options) WithDefaultBaseURL() *Options {
	o.baseURL = defaultBaseURL
	return o
}

// WithDefaultTokenManagerSchema configures standard Client Credentials auth
// without any persistent caching (Redis/File).
// Side Effect: Removes the token if previously set.
// Side Effect: Disable the File token repository if previously set.
// Side Effect: Disable the Redis token repository if previously set.
func (o *Options) WithDefaultTokenManagerSchema(clientID string, clientSecret string) *Options {
	o.tokenManager.useTokenIssuer()

	o.tokenManager.tokenIssuerOptions.fileTokenRepositoryOptions = nil
	o.tokenManager.tokenIssuerOptions.redisTokenRepositoryOptions = nil

	return o.WithDefaultTokenIssuerURL().WithClientCredentials(clientID, clientSecret)
}

// WithDefaultTokenIssuerURL sets the URL to the production IDP.
// Side Effect: Removes the token if previously set.
func (o *Options) WithDefaultTokenIssuerURL() *Options {
	o.tokenManager.useTokenIssuer()

	o.tokenManager.tokenIssuerOptions.issuerURL = defaultTokenIssuerURL

	return o
}

// WithDefaultLogger sets the logger type to "NoLog".
func (o *Options) WithDefaultLogger() *Options {
	o.loggerType = defaultLoggerType
	o.userDefinedDependencies.logger = nil
	return o
}

//
// Logger Options Helpers

// WithNativeLogger enables the standard library logger.
func (o *Options) WithNativeLogger() *Options {
	o.loggerType = LoggerNative
	return o
}

// WithNoLogs disables logging.
func (o *Options) WithNoLogs() *Options {
	o.loggerType = LoggerNoLog
	return o
}

//
// Token Manager Options Helpers

const (
	stdRedisURI                        = "redis://admin:admin@localhost:6379/0"
	stdFileBaseDir                     = "/tmp/sdk-go"
	stdTokenExpirationDriftSeconds int = 300
)

// WithSecurityScopes set the security scopes to be claimed during the
// authentication.
// Side Effect: Removes the token if previously set.
// Side Effect: All previous defined scopes will be erased.
func (o *Options) WithSecurityScopes(scopes ...string) *Options {
	o.tokenManager.useTokenIssuer()

	o.tokenManager.tokenIssuerOptions.scopes = scopes

	return o
}

// WithAdditionalSecurityScopes append the list security scopes to be claimed
// during the authentication.
// Side Effect: Removes the token if previously set.
func (o *Options) WithAdditionalSecurityScopes(scopes ...string) *Options {
	o.tokenManager.useTokenIssuer()

	o.tokenManager.tokenIssuerOptions.scopes = append(o.tokenManager.tokenIssuerOptions.scopes, scopes...)

	return o
}

// WithVaultCredentialsRepository configures the SDK to fetch secrets from HashiCorp Vault.
// Side Effect: Removes the token if previously set.
// Side Effect: Clears any manually set Client Secret.
func (o *Options) WithVaultCredentialsRepository(
	vaultURI string,
	kvMount string,
	kvPath string,
	namespace string,
	rolePath string,
	roleID string,
	secretID string,
) *Options {
	o.tokenManager.useTokenIssuer()

	o.tokenManager.tokenIssuerOptions.clientCredentialOptions = nil

	o.tokenManager.tokenIssuerOptions.vaultCredentialsRepositoryOptions = &vaultCredentialsRepositoryOptions{
		vaultURI:  vaultURI,
		kvMount:   kvMount,
		kvPath:    kvPath,
		namespace: namespace,
		rolePath:  rolePath,
		roleID:    roleID,
		secretID:  secretID,
	}

	return o
}

// WithTokenExpirationDriftSeconds sets the safety buffer for token expiration.
// Side Effect: Removes the token if previously set.
func (o *Options) WithTokenExpirationDriftSeconds(tokenExpirationDriftSeconds int) *Options {
	o.tokenManager.useTokenIssuer()

	o.tokenManager.tokenIssuerOptions.expirationDriftSeconds = uint32(tokenExpirationDriftSeconds)

	return o
}

// WithStandardTokenExpirationDriftSeconds sets the drift to 300 seconds (5 minutes).
func (o *Options) WithStandardTokenExpirationDriftSeconds() *Options {
	return o.WithTokenExpirationDriftSeconds(stdTokenExpirationDriftSeconds)
}

// WithRedisTokenRepositoryFromURI configures a Redis cluster for token caching.
// Side Effect: Removes the token if previously set.
// Side Effect: Disables File Token Repository.
func (o *Options) WithRedisTokenRepositoryFromURI(redisURI string) *Options {
	o.tokenManager.useTokenIssuer()

	o.tokenManager.tokenIssuerOptions.redisTokenRepositoryOptions = &redisTokenRepositoryOptions{
		redisURI: redisURI,
	}

	o.tokenManager.tokenIssuerOptions.fileTokenRepositoryOptions = nil

	return o
}

// WithRedisTokenRepositoryFromStandardURI configures Redis using localhost defaults.
// Side Effect: Removes the token if previously set.
// Side Effect: Disables File Token Repository.
func (o *Options) WithRedisTokenRepositoryFromStandardURI() *Options {
	return o.WithRedisTokenRepositoryFromURI(stdRedisURI)
}

// WithStandardRedisTokenRepository configures localhost Redis with standard drift settings.
// Side Effect: Removes the token if previously set.
// Side Effect: Disables File Token Repository.
func (o *Options) WithStandardRedisTokenRepository() *Options {
	return o.WithRedisTokenRepositoryFromStandardURI().WithStandardTokenExpirationDriftSeconds()
}

// WithFileTokenRepositoryFromBaseDir configures a directory for storing token files.
// Side Effect: Removes the token if previously set.
// Side Effect: Disables Redis Token Repository.
func (o *Options) WithFileTokenRepositoryFromBaseDir(baseDir string) *Options {
	o.tokenManager.useTokenIssuer()

	o.tokenManager.tokenIssuerOptions.fileTokenRepositoryOptions = &fileTokenRepositoryOptions{
		baseDir: baseDir,
	}

	o.tokenManager.tokenIssuerOptions.redisTokenRepositoryOptions = nil

	return o
}

// WithFileTokenRepositoryFromStandardBaseDir configures file storage in /tmp/sdk-go.
// Side Effect: Removes the token if previously set.
// Side Effect: Disables Redis Token Repository.
func (o *Options) WithFileTokenRepositoryFromStandardBaseDir() *Options {
	return o.WithFileTokenRepositoryFromBaseDir(stdFileBaseDir)
}

// WithStandardFileTokenRepository configures /tmp storage with standard drift settings.
// Side Effect: Removes the token if previously set.
// Side Effect: Disables Redis Token Repository.
func (o *Options) WithStandardFileTokenRepository() *Options {
	return o.WithFileTokenRepositoryFromStandardBaseDir().WithStandardTokenExpirationDriftSeconds()
}

//
// User-Defined Dependency Options Helpers

// WithCustomHTTPClient allows injecting a pre-configured *http.Client.
func (o *Options) WithCustomHTTPClient(client *http.Client) *Options {
	o.userDefinedDependencies.httpClient = client
	return o
}

// WithCustomLogger allows injecting a custom logger.Logger implementation.
func (o *Options) WithCustomLogger(logger logger.Logger) *Options {
	o.loggerType = loggerCustom
	o.userDefinedDependencies.logger = logger
	return o
}

// WithCustomMiddleware allows injecting a custom interceptor.Interceptor.
func (o *Options) WithCustomMiddleware(middleware interceptor.Interceptor) *Options {
	o.userDefinedDependencies.middleware = middleware
	return o
}

// WithUserAgent overrides the default User-Agent header sent with every request.
// The default is "sdk-go@<version>" (see pkg/aruba.Version). Pass an empty string
// to restore the default.
func (o *Options) WithUserAgent(ua string) *Options {
	o.userAgent = ua
	return o
}

//
// Helper Functions

// validateURL parses a string to ensure it is a valid absolute URL (HTTP/HTTPS).
func validateURL(rawURL, fieldName string) error {
	if strings.TrimSpace(rawURL) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}

	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("%s is malformed: %w", fieldName, err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%s has invalid scheme '%s': must be http or https", fieldName, u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("%s is missing a host", fieldName)
	}

	return nil
}
