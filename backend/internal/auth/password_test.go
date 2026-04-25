package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	password := "mysecretpassword"

	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestHashPassword_SamePasswordDifferentHash(t *testing.T) {
	password := "mysecretpassword"

	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEmpty(t, hash1)
	assert.NotEmpty(t, hash2)
	assert.NotEqual(t, hash1, hash2) // Hashes should be different due to salt
}

func TestHashPassword_DifferentPasswords(t *testing.T) {
	password1 := "password1"
	password2 := "password2"

	hash1, err1 := HashPassword(password1)
	hash2, err2 := HashPassword(password2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEmpty(t, hash1)
	assert.NotEmpty(t, hash2)
	assert.NotEqual(t, hash1, hash2)
}

func TestVerifyPassword_CorrectPassword(t *testing.T) {
	password := "mysecretpassword"

	hash, err := HashPassword(password)
	assert.NoError(t, err)

	isValid := VerifyPassword(password, hash)
	assert.True(t, isValid)
}

func TestVerifyPassword_IncorrectPassword(t *testing.T) {
	password := "mysecretpassword"
	wrongPassword := "wrongpassword"

	hash, err := HashPassword(password)
	assert.NoError(t, err)

	isValid := VerifyPassword(wrongPassword, hash)
	assert.False(t, isValid)
}

func TestVerifyPassword_EmptyPassword(t *testing.T) {
	password := "mysecretpassword"
	emptyPassword := ""

	hash, err := HashPassword(password)
	assert.NoError(t, err)

	isValid := VerifyPassword(emptyPassword, hash)
	assert.False(t, isValid)
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	password := "mysecretpassword"
	invalidHash := "invalid-hash"

	isValid := VerifyPassword(password, invalidHash)
	assert.False(t, isValid)
}

func TestPasswordComplexity(t *testing.T) {
	passwords := []string{
		"simple",
		"password123",
		"Complex!Password#2024",
		"very-long-password-with-many-characters-and-numbers-12345",
	}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			hash, err := HashPassword(password)
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)

			isValid := VerifyPassword(password, hash)
			assert.True(t, isValid)
		})
	}
}

func TestHashPassword_CompatibilityWithBcrypt(t *testing.T) {
	password := "mysecretpassword"

	hash, err := HashPassword(password)
	assert.NoError(t, err)

	// Verify using bcrypt directly to ensure compatibility
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	assert.NoError(t, err)

	// Verify wrong password fails
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("wrongpassword"))
	assert.Error(t, err)
}