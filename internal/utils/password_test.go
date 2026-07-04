package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain runs once before all tests in this package.
// Lowering BcryptCost from 14 to 10 keeps tests fast (~2s vs ~90s)
// without changing production behaviour.
func TestMain(m *testing.M) {
	BcryptCost = 10
	os.Exit(m.Run())
}

func TestHashPassword_ProducesValidHash(t *testing.T) {
	hash, err := HashPassword("password123")

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "password123", hash, "hash must not equal the plain text")
}

func TestCheckPassword_CorrectPassword(t *testing.T) {
	hash, err := HashPassword("password123")
	require.NoError(t, err)

	match := CheckPassword("password123", hash)
	assert.True(t, match, "correct password should match its hash")
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	hash, err := HashPassword("password123")
	require.NoError(t, err)

	match := CheckPassword("wrongpassword", hash)
	assert.False(t, match, "wrong password should not match the hash")
}

func TestHashPassword_SameInputProducesDifferentHashes(t *testing.T) {
	hash1, err1 := HashPassword("password123")
	hash2, err2 := HashPassword("password123")

	require.NoError(t, err1)
	require.NoError(t, err2)

	// bcrypt generates a new salt each call — identical inputs must NOT produce identical hashes
	assert.NotEqual(t, hash1, hash2, "bcrypt should produce unique hashes per call (different salts)")
}
