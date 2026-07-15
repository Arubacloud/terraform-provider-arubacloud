package file

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Arubacloud/sdk-go/internal/ports/auth"
)

var (
	accessToken = "this is a valid token"
	expiry      = time.Now().Add(24 * time.Hour)
)

func TestTokenRepository_FetchToken(t *testing.T) {
	t.Run("should return nil when no token file", func(t *testing.T) {
		dir := t.TempDir()
		repo := NewFileTokenRepository("user-123", dir)

		token, err := repo.FetchToken(t.Context())
		require.Error(t, err)
		require.Nil(t, token)
	})

	t.Run("should return saved token", func(t *testing.T) {
		dir := t.TempDir()
		repo := NewFileTokenRepository("user-123", dir)

		savedToken := &auth.Token{AccessToken: accessToken, Expiry: expiry}
		err := repo.SaveToken(t.Context(), savedToken)
		require.NoError(t, err)

		fetchedToken, err := repo.FetchToken(t.Context())
		require.NoError(t, err)
		require.NotNil(t, fetchedToken)
		require.Equal(t, savedToken.AccessToken, fetchedToken.AccessToken)
		require.NotEmpty(t, fetchedToken.Expiry)
	})

	t.Run("should return token for expired token", func(t *testing.T) {
		dir := t.TempDir()
		repo := NewFileTokenRepository("user-123", dir)

		expiredToken := &auth.Token{AccessToken: accessToken, Expiry: time.Now().Add(-1 * time.Hour)}
		err := repo.SaveToken(t.Context(), expiredToken)
		require.NoError(t, err)

		fetchedToken, err := repo.FetchToken(t.Context())
		require.NoError(t, err)
		require.NotNil(t, fetchedToken)
	})
}

func TestTokenRepository_SaveToken(t *testing.T) {

	t.Run("should save file without problem", func(t *testing.T) {
		dir := t.TempDir()
		repo := NewFileTokenRepository("user-123", dir)

		token := &auth.Token{AccessToken: accessToken, Expiry: expiry}
		// Test fetching a token when none exists	// Test saving a token
		err := repo.SaveToken(t.Context(), token)
		require.NoError(t, err)
	})
	t.Run("should return error when token is nil", func(t *testing.T) {
		dir := t.TempDir()
		repo := NewFileTokenRepository("user-123", dir)

		// Test fetching a token when none exists	// Test saving a token
		err := repo.SaveToken(t.Context(), nil)
		require.Error(t, err)
	})

	t.Run("should return error for not existing dir", func(t *testing.T) {
		dir := "/non/existing\\*path\\*/dir"
		repo := NewFileTokenRepository("user-123", dir)
		token := &auth.Token{AccessToken: accessToken, Expiry: expiry}

		// Test fetching a token when none exists	// Test saving a token
		err := repo.SaveToken(t.Context(), token)
		require.Error(t, err)
	})
}
