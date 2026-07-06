package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-jwt-secret-that-is-long-enough"
const wrongSecret = "different-secret-key"

func TestGenerateToken_ValidatesSuccessfully(t *testing.T) {
	userID := uint(42)

	token, err := GenerateToken(userID, testSecret)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// The token must round-trip back to the same userID
	gotID, err := ValidateToken(token, testSecret)

	require.NoError(t, err)
	assert.Equal(t, userID, gotID, "validated userID must match original")
}

func TestValidateToken_WrongSecretIsRejected(t *testing.T) {
	// Phase 21A security test: a token signed with secret A must be
	// rejected when validated with secret B.
	token, err := GenerateToken(uint(10), testSecret)
	require.NoError(t, err)

	_, err = ValidateToken(token, wrongSecret)

	assert.Error(t, err, "token signed with a different secret must be rejected")
}

func TestValidateToken_TamperedTokenIsRejected(t *testing.T) {
	token, err := GenerateToken(uint(7), testSecret)
	require.NoError(t, err)

	// Flip the last character to simulate tampering
	tampered := token[:len(token)-1] + "X"

	_, err = ValidateToken(tampered, testSecret)

	assert.Error(t, err, "tampered token signature must be rejected")
}

func TestValidateToken_EmptyTokenIsRejected(t *testing.T) {
	_, err := ValidateToken("", testSecret)
	assert.Error(t, err, "empty token must be rejected")
}

func TestValidateToken_InvalidUserIDClaimIsRejected(t *testing.T) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_id": "hello",
			"exp": time.Now().
				Add(24 * time.Hour).
				Unix(),
		},
	)

	tokenString, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	_, err = ValidateToken(tokenString, testSecret)

	assert.ErrorIs(t, err, jwt.ErrTokenInvalidClaims)
}
