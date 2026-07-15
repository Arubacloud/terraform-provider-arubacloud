package vault

import (
	"fmt"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/internal/ports/auth"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

//go:generate mockgen -package vault -destination=zz_mock_vault_test.go github.com/Arubacloud/sdk-go/internal/impl/auth/credentialsrepository/vault VaultClient,LogicalAPI,KvAPI

func TestCredentialsRepository_ensureAuthenticated(t *testing.T) {
	t.Run("do nothing if token is already set", func(t *testing.T) {

		repo := &CredentialsRepository{
			tokenExist: true,
			expiration: time.Now().Add(24 * time.Hour),
		}

		err := repo.ensureAuthenticated()

		require.NoError(t, err)
		require.True(t, repo.tokenExist)
	})
	t.Run("Error on write secret", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := NewMockVaultClient(ctrl)
		mockLogicalAPI := NewMockLogicalAPI(ctrl)

		mockClient.EXPECT().Logical().Return(mockLogicalAPI)
		repo := newRepo(mockClient, false, 0, time.Time{})
		data := loginData(repo.roleID, repo.secretID)

		mockLogicalAPI.
			EXPECT().Write(gomock.Any(), data).
			Return(nil, fmt.Errorf("mock error"))

		err := repo.ensureAuthenticated()

		require.Error(t, err)
		require.False(t, repo.tokenExist)
	})
	t.Run("successful login sets token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := NewMockVaultClient(ctrl)
		mockLogicalAPI := NewMockLogicalAPI(ctrl)

		// Setup the mock responses

		calls := []string{}

		mockClient.
			EXPECT().
			SetToken("mock-token").
			Do(func(token string) {
				calls = append(calls, token)
			}).Times(1)

		mockClient.EXPECT().Logical().Return(mockLogicalAPI)

		repo := newRepo(mockClient, false, 0, time.Time{})
		data := loginData(repo.roleID, repo.secretID)

		mockLogicalAPI.
			EXPECT().Write(gomock.Any(), data).
			Return(&vaultapi.Secret{
				Auth: &vaultapi.SecretAuth{
					ClientToken:   "mock-token",
					Renewable:     false,
					LeaseDuration: 3600,
				},
			}, nil)

		err := repo.ensureAuthenticated()

		require.NoError(t, err)
		require.True(t, repo.tokenExist)
		require.Equal(t, "mock-token", calls[0])
		require.Equal(t, false, repo.renewable)
		require.Equal(t, 3600*time.Second, repo.ttl)
	})
	t.Run("successful login sets token and namespace", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := NewMockVaultClient(ctrl)
		mockLogicalAPI := NewMockLogicalAPI(ctrl)

		// Setup the mock responses
		repo := newRepo(mockClient, false, 0, time.Time{})
		repo.namespace = "test-namespace"
		mockClient.
			EXPECT().
			SetToken("mock-token").
			Return()

		mockClient.EXPECT().
			SetNamespace(repo.namespace).
			Return()

		mockClient.EXPECT().Logical().Return(mockLogicalAPI)

		data := loginData(repo.roleID, repo.secretID)

		mockLogicalAPI.
			EXPECT().Write(gomock.Any(), data).
			Return(&vaultapi.Secret{
				Auth: &vaultapi.SecretAuth{
					ClientToken:   "mock-token",
					Renewable:     true,
					LeaseDuration: 3600,
				},
			}, nil)

		err := repo.ensureAuthenticated()

		require.NoError(t, err)
		require.True(t, repo.tokenExist)
		require.Equal(t, true, repo.renewable)
		require.Equal(t, 3600*time.Second, repo.ttl)
	})
}

func TestCredentialsRepository_GetCredentialsFromVaultSecret(t *testing.T) {
	t.Run("should return credentials when secret contains client_id and client_secret", func(t *testing.T) {
		secret := &vaultapi.KVSecret{
			Data: map[string]interface{}{
				"client-id":     "test-client-id",
				"client-secret": "test-client-secret",
			},
		}

		creds, err := getCredentialsFromVaultSecret(secret)

		require.NoError(t, err)
		require.Equal(t, "test-client-id", creds.ClientID)
		require.Equal(t, "test-client-secret", creds.ClientSecret)
	})

	t.Run("should return error when client_id is missing", func(t *testing.T) {
		secret := &vaultapi.KVSecret{
			Data: map[string]interface{}{
				"client-secret": "test-client-secret",
			},
		}

		creds, err := getCredentialsFromVaultSecret(secret)

		require.Error(t, err)
		require.ErrorIs(t, auth.ErrCredentialsNotFound, err)
		require.Nil(t, creds)
	})

	t.Run("should return error when client_secret is missing", func(t *testing.T) {
		secret := &vaultapi.KVSecret{
			Data: map[string]interface{}{
				"client-id": "test-client-id",
			},
		}

		creds, err := getCredentialsFromVaultSecret(secret)

		require.Error(t, err)
		require.ErrorIs(t, auth.ErrCredentialsNotFound, err)
		require.Nil(t, creds)
	})
	t.Run("should return error when client_id is not a string", func(t *testing.T) {
		secret := &vaultapi.KVSecret{
			Data: map[string]interface{}{
				"client-id":     12345,
				"client-secret": "test-client-secret",
			},
		}

		creds, err := getCredentialsFromVaultSecret(secret)

		require.Error(t, err)
		require.ErrorIs(t, auth.ErrCredentialsNotFound, err)
		require.Nil(t, creds)
	})
	t.Run("should return error when client-secret is not a string", func(t *testing.T) {
		secret := &vaultapi.KVSecret{
			Data: map[string]interface{}{
				"client-id":     "test-client-id",
				"client-secret": 67890,
			},
		}

		creds, err := getCredentialsFromVaultSecret(secret)

		require.Error(t, err)
		require.ErrorIs(t, auth.ErrCredentialsNotFound, err)
		require.Nil(t, creds)
	})
}

func TestCredentialsRepository_FetchCredentials(t *testing.T) {
	t.Run("should report credentials not found error when vault returns nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockClient := NewMockVaultClient(ctrl)
		mockKvAPI := NewMockKvAPI(ctrl)

		repo := newRepo(mockClient, true, time.Until(time.Now()), time.Now().Add(24*time.Hour))

		mockClient.EXPECT().
			KVv2(repo.kvMount).Return(mockKvAPI)

		mockKvAPI.EXPECT().
			Get(gomock.Any(), repo.kvPath).
			Return(nil, fmt.Errorf("vault: secret not found"))

		// When we try to fetch the token
		secretData, err := repo.FetchCredentials(t.Context())

		// And no token should be returned
		require.Error(t, err)
		require.ErrorIs(t, err, auth.ErrCredentialsNotFound)
		require.Nil(t, secretData)
	})

	t.Run("should report credentials not found error when vault secret is missing data", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockClient := NewMockVaultClient(ctrl)
		mockKvAPI := NewMockKvAPI(ctrl)

		repo := newRepo(mockClient, true, time.Until(time.Now()), time.Now().Add(24*time.Hour))

		mockClient.EXPECT().
			KVv2(repo.kvMount).Return(mockKvAPI)

		mockKvAPI.EXPECT().
			Get(gomock.Any(), repo.kvPath).
			Return(&vaultapi.KVSecret{
				Data: map[string]any{},
			}, nil)

		// When we try to fetch the token
		secretData, err := repo.FetchCredentials(t.Context())

		// And no token should be returned
		require.Error(t, err)
		require.ErrorIs(t, err, auth.ErrCredentialsNotFound)
		require.Nil(t, secretData)
	})
	t.Run("Run ok when vault secret contains credentials", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockClient := NewMockVaultClient(ctrl)
		mockKvAPI := NewMockKvAPI(ctrl)

		repo := newRepo(mockClient, true, time.Until(time.Now()), time.Now().Add(24*time.Hour))

		mockClient.EXPECT().
			KVv2(repo.kvMount).Return(mockKvAPI)

		mockKvAPI.EXPECT().
			Get(gomock.Any(), repo.kvPath).
			Return(&vaultapi.KVSecret{
				Data: map[string]any{
					"client-id":     "test-client-id",
					"client-secret": "test-client-secret",
				},
			}, nil)

		// When we try to fetch the token
		secretData, err := repo.FetchCredentials(t.Context())

		// And no token should be returned
		require.NoError(t, err)
		require.NotNil(t, secretData)
		require.Equal(t, "test-client-id", secretData.ClientID)
		require.Equal(t, "test-client-secret", secretData.ClientSecret)
	})

	t.Run("should authenticate when no token exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockClient := NewMockVaultClient(ctrl)
		mockKvAPI := NewMockKvAPI(ctrl)
		mockLogicalAPI := NewMockLogicalAPI(ctrl)

		repo := newRepo(mockClient, false, time.Until(time.Now()), time.Time{})

		// Expect loginWithAppRole to be called
		mockClient.EXPECT().
			Logical().
			Return(mockLogicalAPI)

		loginData := loginData(repo.roleID, repo.secretID)

		mockLogicalAPI.
			EXPECT().
			Write(gomock.Any(), loginData).
			Return(&vaultapi.Secret{
				Auth: &vaultapi.SecretAuth{
					ClientToken:   "mock-token",
					Renewable:     false,
					LeaseDuration: 3600,
				},
			}, nil)

		mockClient.
			EXPECT().
			SetToken("mock-token").
			Return()

		mockClient.EXPECT().
			KVv2(repo.kvMount).Return(mockKvAPI)

		mockKvAPI.EXPECT().
			Get(gomock.Any(), repo.kvPath).
			Return(&vaultapi.KVSecret{
				Data: map[string]interface{}{
					"client-id":     "test-client-id",
					"client-secret": "test-client-secret",
				},
			}, nil)

		// When we try to fetch the token
		secretData, err := repo.FetchCredentials(t.Context())

		// And no token should be returned
		require.NoError(t, err)
		require.NotNil(t, secretData)
		require.Equal(t, "test-client-id", secretData.ClientID)
		require.Equal(t, "test-client-secret", secretData.ClientSecret)
	})
	t.Run("should authenticate when token exist but is expired", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockClient := NewMockVaultClient(ctrl)
		mockKvAPI := NewMockKvAPI(ctrl)
		mockLogicalAPI := NewMockLogicalAPI(ctrl)

		repo := newRepo(mockClient, true, time.Until(time.Now().Add(-1*time.Hour)), time.Now().Add(-1*time.Hour))

		// Expect loginWithAppRole to be called
		mockClient.EXPECT().
			Logical().
			Return(mockLogicalAPI)

		loginData := loginData(repo.roleID, repo.secretID)

		mockLogicalAPI.
			EXPECT().
			Write(gomock.Any(), loginData).
			Return(&vaultapi.Secret{
				Auth: &vaultapi.SecretAuth{
					ClientToken:   "mock-token",
					Renewable:     false,
					LeaseDuration: 3600,
				},
			}, nil).Times(1)
		mockClient.
			EXPECT().
			SetToken("mock-token").
			Return().Times(1)

		mockClient.EXPECT().
			KVv2(repo.kvMount).Return(mockKvAPI)

		mockKvAPI.EXPECT().
			Get(gomock.Any(), repo.kvPath).
			Return(&vaultapi.KVSecret{
				Data: map[string]interface{}{
					"client-id":     "test-client-id",
					"client-secret": "test-client-secret",
				},
			}, nil)
		// When we try to fetch the token
		secretData, err := repo.FetchCredentials(t.Context())

		// And no token should be returned
		require.NoError(t, err)
		require.NotNil(t, secretData)
		require.Equal(t, "test-client-id", secretData.ClientID)
		require.Equal(t, "test-client-secret", secretData.ClientSecret)
	})

}

func TestCredentialsRepository_performLoginAppRole(t *testing.T) {
	t.Run("should perform AppRole login and return token secret", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := NewMockVaultClient(ctrl)
		mockLogicalAPI := NewMockLogicalAPI(ctrl)

		mockClient.EXPECT().Logical().Return(mockLogicalAPI)
		repo := newRepo(mockClient, false, 0, time.Time{})
		data := loginData(repo.roleID, repo.secretID)

		expectedSecret := &vaultapi.Secret{
			Auth: &vaultapi.SecretAuth{
				ClientToken:   "mock-token",
				Renewable:     true,
				LeaseDuration: 7200,
			},
		}

		mockLogicalAPI.
			EXPECT().Write(gomock.Any(), data).
			Return(expectedSecret, nil)

		secret, err := repo.performAppRoleLogin()

		require.NoError(t, err)
		require.Equal(t, expectedSecret, secret)
	})
	t.Run("setting namespace also", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := NewMockVaultClient(ctrl)
		mockLogicalAPI := NewMockLogicalAPI(ctrl)
		mockClient.EXPECT().
			SetNamespace("test-namespace").
			Return().Times(1)

		mockClient.EXPECT().Logical().Return(mockLogicalAPI)
		repo := newRepo(mockClient, false, 0, time.Time{})

		data := loginData(repo.roleID, repo.secretID)
		expectedSecret := &vaultapi.Secret{
			Auth: &vaultapi.SecretAuth{
				ClientToken:   "mock-token",
				Renewable:     true,
				LeaseDuration: 7200,
			},
		}
		mockLogicalAPI.
			EXPECT().Write(gomock.Any(), data).
			Return(expectedSecret, nil)

		repo.namespace = "test-namespace"
		secret, err := repo.performAppRoleLogin()

		require.NoError(t, err)
		require.Equal(t, expectedSecret, secret)
	})
}

func loginData(role, secret string) map[string]interface{} {
	return map[string]interface{}{
		"role_id":   role,
		"secret_id": secret,
	}
}

func newRepo(mock VaultClient, token bool, ttl time.Duration, exp time.Time) *CredentialsRepository {
	return &CredentialsRepository{
		client:     mock,
		kvMount:    "test-kv",
		kvPath:     "test-path",
		rolePath:   "test-role-path",
		roleID:     "test-role-id",
		secretID:   "test-secret-id",
		tokenExist: token,
		ttl:        ttl,
		expiration: exp,
	}
}
