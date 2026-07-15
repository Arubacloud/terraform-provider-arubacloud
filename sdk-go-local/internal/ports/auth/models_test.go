package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestToken_IsVAlid(t *testing.T) {
	t.Run("shoul be valid when expiry is zero", func(t *testing.T) {
		// Given a token which have its expiry as "zero"
		token := &Token{}

		// When we check its validity
		valid := token.IsValid()

		// Then it should be true
		require.True(t, valid)
	})

	t.Run("shoul be valid when expiry is in the future", func(t *testing.T) {
		// Given a token which have its expiry is in the future
		token := &Token{Expiry: time.Now().Add(1 * time.Hour)}

		// When we check its validity
		valid := token.IsValid()

		// Then it should be true
		require.True(t, valid)
	})

	t.Run("shoul not be valid when expiry is in the past", func(t *testing.T) {
		// Given a token which have its expiry in the past
		token := &Token{Expiry: time.Now().Add(-1 * time.Hour)}

		// When we check its validity
		valid := token.IsValid()

		// Then it should be false
		require.False(t, valid)
	})
}
