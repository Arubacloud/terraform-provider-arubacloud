//nolint:unparam // TODO: better error handling
package aruba

import (
	"errors"
	"fmt"
	"net/http"

	vaultapi "github.com/hashicorp/vault/api"
	redis_client "github.com/redis/go-redis/v9"

	memory_creds_repo "github.com/Arubacloud/sdk-go/internal/impl/auth/credentialsrepository/memory"
	vault_creds_repo "github.com/Arubacloud/sdk-go/internal/impl/auth/credentialsrepository/vault"
	oauth2_connector "github.com/Arubacloud/sdk-go/internal/impl/auth/providerconnector/oauth2"
	std_token_manager "github.com/Arubacloud/sdk-go/internal/impl/auth/tokenmanager/standard"
	file_token_repo "github.com/Arubacloud/sdk-go/internal/impl/auth/tokenrepository/file"
	memory_token_repo "github.com/Arubacloud/sdk-go/internal/impl/auth/tokenrepository/memory"
	redis_token_repo "github.com/Arubacloud/sdk-go/internal/impl/auth/tokenrepository/redis"
	std_interceptor "github.com/Arubacloud/sdk-go/internal/impl/interceptor/standard"
	native_logger "github.com/Arubacloud/sdk-go/internal/impl/logger/native"
	noop_logger "github.com/Arubacloud/sdk-go/internal/impl/logger/noop"
	"github.com/Arubacloud/sdk-go/internal/ports/auth"
	"github.com/Arubacloud/sdk-go/internal/ports/interceptor"
	"github.com/Arubacloud/sdk-go/internal/ports/logger"
	"github.com/Arubacloud/sdk-go/internal/restclient"
	middleware_util "github.com/Arubacloud/sdk-go/pkg/util/middleware"
)

// Client

func buildClient(options *Options) (Client, error) {
	err := options.validate()
	if err != nil {
		return nil, err // TODO: better error handling
	}

	restClient, err := buildRESTClient(options)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	auditClient, err := buildAuditClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	computeClient, err := buildComputeClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	containerClient, err := buildContainerClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	databaseClient, err := buildDatabaseClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	metricClient, err := buildMetricClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	networkClient, err := buildNetworkClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	projectClient, err := buildProjectClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	scheduleClient, err := buildScheduleClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	securityClient, err := buildSecurityClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	storageClient, err := buildStorageClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return &clientImpl{
		auditClient:     auditClient,
		computeClient:   computeClient,
		containerClient: containerClient,
		databaseClient:  databaseClient,
		metricsClient:   metricClient,
		networkClient:   networkClient,
		projectClient:   projectClient,
		scheduleClient:  scheduleClient,
		securityClient:  securityClient,
		storageClient:   storageClient,
	}, nil
}

//
// Dependencies

func buildRESTClient(options *Options) (*restclient.Client, error) {
	httpClient, err := buildHTTPClient(options)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	logger, err := buildLogger(options)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	middleware, err := buildMiddleware(options)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return restclient.NewClient(options.baseURL, httpClient, middleware, logger), nil
}

func buildHTTPClient(options *Options) (*http.Client, error) {
	if options.userDefinedDependencies.httpClient != nil {
		return options.userDefinedDependencies.httpClient, nil
	}

	return http.DefaultClient, nil
}

func buildLogger(options *Options) (logger.Logger, error) {
	switch options.loggerType {
	case LoggerNoLog:
		return &noop_logger.NoOpLogger{}, nil

	case LoggerNative:
		return native_logger.NewDefaultLogger(), nil

	case loggerCustom:
		return options.userDefinedDependencies.logger, nil
	}

	return nil, fmt.Errorf("unknown logging type: %d", options.loggerType)
}

func buildMiddleware(options *Options) (interceptor.Interceptor, error) {
	// The token manager must be always the last to be bound
	tokenManager, err := buildTokenManager(&options.tokenManager)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	ua := options.userAgent
	if ua == "" {
		ua = defaultUserAgent
	}

	if options.userDefinedDependencies.middleware != nil {
		// Bind UA first so user middleware (and the token manager last) can still override it.
		interceptable, ok := options.userDefinedDependencies.middleware.(interceptor.Interceptable)
		if !ok {
			return nil, errors.New("failed to bind the token manager to the user-defined middleware")
		}
		if err := interceptable.Bind(middleware_util.WithUserAgent(ua)); err != nil {
			return nil, err
		}
		middleware, err := buildUserDefinedMiddleware(options.userDefinedDependencies.middleware, tokenManager)
		if err != nil {
			return nil, err // TODO: better error handling
		}
		return middleware, nil
	}

	middleware := std_interceptor.NewInterceptor()
	if err := middleware.Bind(middleware_util.WithUserAgent(ua)); err != nil {
		return nil, err
	}
	err = tokenManager.BindTo(middleware)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return middleware, nil
}

func buildUserDefinedMiddleware(middleware interceptor.Interceptor, tokenManager auth.TokenManager) (interceptor.Interceptor, error) {
	interceptable, ok := middleware.(interceptor.Interceptable)
	if !ok {
		// TODO: better error handling
		return nil, errors.New("failed to bind the token manager to the user-defined middleware")
	}

	err := tokenManager.BindTo(interceptable)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return middleware, nil
}

//
// Token Manager

func buildTokenManager(options *tokenManagerOptions) (*std_token_manager.TokenManager, error) {
	if options.token != nil {
		return std_token_manager.NewStaticTokenManager(
			memory_token_repo.NewTokenRepositoryWithAccessToken(*options.token),
		), nil
	}

	providerConnector, err := buildProviderConnector(options.tokenIssuerOptions)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	tokenRepository, err := buildTokenRepository(options.tokenIssuerOptions)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	tokenManager := std_token_manager.NewTokenManager(providerConnector, tokenRepository)

	return tokenManager, nil
}

func buildProviderConnector(options *tokenIssuerOptions) (*oauth2_connector.ProviderConnector, error) {
	credentialsRepository, err := buildCredentialsRepository(options)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return oauth2_connector.NewProviderConnector(credentialsRepository, options.issuerURL, options.scopes), nil
}

func buildCredentialsRepository(options *tokenIssuerOptions) (auth.CredentialsRepository, error) {
	if options.clientCredentialOptions != nil {
		return memory_creds_repo.NewCredentialsRepository(
			options.clientCredentialOptions.clientID,
			options.clientCredentialOptions.clientSecret,
		), nil
	}

	if options.vaultCredentialsRepositoryOptions != nil {
		vaultCredentialsRepository, err := buildVaultCredentialsRepository(options.vaultCredentialsRepositoryOptions)
		if err != nil {
			return nil, err // TODO: better error handling
		}

		return memory_creds_repo.NewCredentialsProxy(vaultCredentialsRepository), nil
	}

	return nil, errors.New("no credentials repository defined")
}

func buildVaultCredentialsRepository(options *vaultCredentialsRepositoryOptions) (*vault_creds_repo.CredentialsRepository, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = options.vaultURI

	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return vault_creds_repo.NewCredentialsRepository(
		vault_creds_repo.NewVaultClientAdapter(client),
		options.kvMount,
		options.kvPath,
		options.namespace,
		options.rolePath,
		options.roleID,
		options.secretID,
	), nil
}

func buildTokenRepository(options *tokenIssuerOptions) (auth.TokenRepository, error) {
	var keyIDComponent string
	if options.clientCredentialOptions != nil {
		keyIDComponent = options.clientCredentialOptions.clientID
	} else if options.vaultCredentialsRepositoryOptions != nil {
		keyIDComponent = options.vaultCredentialsRepositoryOptions.secretID
	}

	var persistentTokenRepository auth.TokenRepository
	var err error

	if options.redisTokenRepositoryOptions != nil {
		persistentTokenRepository, err = buildRedisTokenRepository(
			keyIDComponent,
			options.redisTokenRepositoryOptions,
		)
		if err != nil {
			return nil, err // TODO: better error handling
		}

	} else if options.fileTokenRepositoryOptions != nil {
		persistentTokenRepository, err = buildFileTokenRepository(
			keyIDComponent,
			options.fileTokenRepositoryOptions,
		)
		if err != nil {
			return nil, err // TODO: better error handling
		}
	}

	if persistentTokenRepository != nil {
		return memory_token_repo.NewTokenProxy(persistentTokenRepository), nil
	}

	return memory_token_repo.NewTokenRepository(), nil
}

func buildRedisTokenRepository(clientID string, options *redisTokenRepositoryOptions) (*redis_token_repo.TokenRepository, error) {
	opt, err := redis_client.ParseURL(options.redisURI)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	rdb := redis_client.NewClient(opt)
	adapter := redis_token_repo.NewRedisAdapter(rdb)

	return redis_token_repo.NewRedisTokenRepository(clientID, adapter), nil
}

func buildFileTokenRepository(clientID string, options *fileTokenRepositoryOptions) (*file_token_repo.TokenRepository, error) {
	return file_token_repo.NewFileTokenRepository(clientID, options.baseDir), nil
}

//
// Audit domain clients

func buildAuditClient(restClient *restclient.Client) (AuditClient, error) {
	eventsClient, err := buildEventsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return &auditClientImpl{eventsClient: eventsClient}, nil
}

func buildEventsClient(restClient *restclient.Client) (EventsClient, error) {
	return newAuditEventsClientAdapter(restClient), nil
}

//
// Compute domain clients

func buildComputeClient(restClient *restclient.Client) (ComputeClient, error) {
	cloudServerClient, err := buildCloudServersClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	keyPairClient, err := buildKeyPairsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return &computeClientImpl{
		cloudServerClient: cloudServerClient,
		keyPairClient:     keyPairClient,
	}, nil
}

func buildCloudServersClient(restClient *restclient.Client) (CloudServersClient, error) {
	return newCloudServersClientAdapter(restClient), nil
}

func buildKeyPairsClient(restClient *restclient.Client) (KeyPairsClient, error) {
	return newKeyPairsClientAdapter(restClient), nil
}

//
// Container domain clients

func buildContainerClient(restClient *restclient.Client) (ContainerClient, error) {
	kaasClient, err := buildKaaSClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	containerRegistryClient, err := buildContainerRegistryClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return &containerClientImpl{
		kaasClient:              kaasClient,
		containerRegistryClient: containerRegistryClient,
	}, nil
}

func buildKaaSClient(restClient *restclient.Client) (KaaSClient, error) {
	return newKaaSClientAdapter(restClient), nil
}

func buildContainerRegistryClient(restClient *restclient.Client) (ContainerRegistryClient, error) {
	return newContainerRegistriesClientAdapter(restClient), nil
}

//
// Database domain clients

func buildDatabaseClient(restClient *restclient.Client) (DatabaseClient, error) {
	dbaasClient, err := buildDBaaSClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	databasesClient, err := buildDatabasesClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	backupsClient, err := buildBackupsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	usersClient, err := buildUsersClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	grantsClient, err := buildGrantsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return &databaseClientImpl{
		dbaasClient:     dbaasClient,
		databasesClient: databasesClient,
		backupsClient:   backupsClient,
		usersClient:     usersClient,
		grantsClient:    grantsClient,
	}, nil
}

func buildDBaaSClient(restClient *restclient.Client) (DBaaSClient, error) {
	return newDBaaSClientAdapter(restClient), nil
}

func buildDatabasesClient(restClient *restclient.Client) (DatabasesClient, error) {
	return newDatabasesClientAdapter(restClient), nil
}

func buildBackupsClient(restClient *restclient.Client) (BackupsClient, error) {
	return newDBaaSBackupsClientAdapter(restClient), nil
}

func buildUsersClient(restClient *restclient.Client) (UsersClient, error) {
	return newUsersClientAdapter(restClient), nil
}

func buildGrantsClient(restClient *restclient.Client) (GrantsClient, error) {
	return newGrantsClientAdapter(restClient), nil
}

//
// Metric domain clients

func buildMetricClient(restClient *restclient.Client) (MetricClient, error) {
	alertsClient, err := buildAlertsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	metricsClient, err := buildMetricsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return &metricClientImpl{
		alertsClient:  alertsClient,
		metricsClient: metricsClient,
	}, nil
}

func buildAlertsClient(restClient *restclient.Client) (AlertsClient, error) {
	return newAlertsClientAdapter(restClient), nil
}

func buildMetricsClient(restClient *restclient.Client) (MetricsClient, error) {
	return newMetricsClientAdapter(restClient), nil
}

//
// Network domain clients

func buildNetworkClient(restClient *restclient.Client) (NetworkClient, error) {
	elasticIPsClient, err := buildElasticIPsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	loadBalancersClient, err := buildLoadBalancersClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	securityGroupRulesClient, err := buildSecurityGroupRulesClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	securityGroupsClient, err := buildSecurityGroupsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	subnetsClient, err := buildSubnetsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	vpcPeeringRoutesClient, err := buildVPCPeeringRoutesClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	vpcPeeringsClient, err := buildVPCPeeringsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	vpcsClient, err := buildVPCsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	vpnRoutesClient, err := buildVPNRoutesClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	vpnTunnelsClient, err := buildVPNTunnelsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return &networkClientImpl{
		elasticIPsClient:         elasticIPsClient,
		loadBalancersClient:      loadBalancersClient,
		securityGroupRulesClient: securityGroupRulesClient,
		securityGroupsClient:     securityGroupsClient,
		subnetsClient:            subnetsClient,
		vpcPeeringRoutesClient:   vpcPeeringRoutesClient,
		vpcPeeringsClient:        vpcPeeringsClient,
		vpcsClient:               vpcsClient,
		vpnRoutesClient:          vpnRoutesClient,
		vpnTunnelsClient:         vpnTunnelsClient,
	}, nil

}

func buildElasticIPsClient(restClient *restclient.Client) (ElasticIPsClient, error) {
	return newElasticIPsClientAdapter(restClient), nil
}

func buildLoadBalancersClient(restClient *restclient.Client) (LoadBalancersClient, error) {
	return newLoadBalancersClientAdapter(restClient), nil
}

func buildSecurityGroupRulesClient(restClient *restclient.Client) (SecurityGroupRulesClient, error) {
	return newSecurityRulesClientAdapter(restClient), nil
}

func buildSecurityGroupsClient(restClient *restclient.Client) (SecurityGroupsClient, error) {
	return newSecurityGroupsClientAdapter(restClient), nil
}

func buildSubnetsClient(restClient *restclient.Client) (SubnetsClient, error) {
	return newSubnetsClientAdapter(restClient), nil
}

func buildVPCPeeringRoutesClient(restClient *restclient.Client) (VPCPeeringRoutesClient, error) {
	return newVPCPeeringRoutesClientAdapter(restClient), nil
}

func buildVPCPeeringsClient(restClient *restclient.Client) (VPCPeeringsClient, error) {
	return newVPCPeeringsClientAdapter(restClient), nil
}

func buildVPCsClient(restClient *restclient.Client) (VPCsClient, error) {
	return newVPCsClientAdapter(restClient), nil
}

func buildVPNRoutesClient(restClient *restclient.Client) (VPNRoutesClient, error) {
	return newVPNRoutesClientAdapter(restClient), nil
}

func buildVPNTunnelsClient(restClient *restclient.Client) (VPNTunnelsClient, error) {
	return newVPNTunnelsClientAdapter(restClient), nil
}

//
// Project domain clients

func buildProjectClient(restClient *restclient.Client) (ProjectClient, error) {
	return newProjectClientAdapter(restClient), nil
}

//
// Schedule domain clients

func buildScheduleClient(restClient *restclient.Client) (ScheduleClient, error) {
	jobsClient, err := buildJobsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return &scheduleClientImpl{
		jobsClient: jobsClient,
	}, nil
}

func buildJobsClient(restClient *restclient.Client) (JobsClient, error) {
	return newJobsClientAdapter(restClient), nil
}

//
// Security domain clients

func buildSecurityClient(restClient *restclient.Client) (SecurityClient, error) {
	kmsClient, err := buildKMSClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	keysClient, err := buildKeysClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	kmipsClient, err := buildKmipsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return &securityClientImpl{
		kmsClient:   kmsClient,
		keysClient:  keysClient,
		kmipsClient: kmipsClient,
	}, nil
}

func buildKMSClient(restClient *restclient.Client) (KMSClient, error) {
	return newKMSClientAdapter(restClient), nil
}

func buildKeysClient(rest *restclient.Client) (KeysClient, error) {
	return newKeysClientAdapter(rest), nil
}

func buildKmipsClient(rest *restclient.Client) (KmipsClient, error) {
	return newKmipsClientAdapter(rest), nil
}

//
// Storage domain clients

func buildStorageClient(restClient *restclient.Client) (StorageClient, error) {
	snapshotsClient, err := buildSnapshotsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	volumesClient, err := buildVolumesClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	restoresClient, err := buildStorageRestoresClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	backupsClient, err := buildStorageBackupsClient(restClient)
	if err != nil {
		return nil, err // TODO: better error handling
	}

	return &storageClientImpl{
		snapshotsClient: snapshotsClient,
		volumesClient:   volumesClient,
		backupsClient:   backupsClient,
		restoresClient:  restoresClient,
	}, nil
}

func buildSnapshotsClient(restClient *restclient.Client) (SnapshotsClient, error) {
	return newSnapshotsClientAdapter(restClient), nil
}

func buildVolumesClient(restClient *restclient.Client) (VolumesClient, error) {
	return newVolumesClientAdapter(restClient), nil
}

func buildStorageBackupsClient(restClient *restclient.Client) (StorageBackupsClient, error) {
	return newStorageBackupsClientAdapter(restClient), nil
}

func buildStorageRestoresClient(restClient *restclient.Client) (StorageRestoreClient, error) {
	return newStorageRestoresClientAdapter(restClient), nil
}
