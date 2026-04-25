package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewJWTService(t *testing.T) {
	secret := "test-secret-key"
	service := NewJWTService(secret)

	assert.NotNil(t, service)
	assert.Equal(t, []byte(secret), service.secretKey)
	assert.Equal(t, time.Hour, service.accessExpiry)
	assert.Equal(t, time.Hour*24*7, service.refreshExpiry)
}

func TestGenerateToken(t *testing.T) {
	secret := "test-secret-key"
	service := NewJWTService(secret)

	user := &User{
		ID:          "123",
		Email:       "test@example.com",
		DisplayName: "Test User",
		Role:        RoleUser,
	}

	accessToken, refreshToken, err := service.GenerateToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotEqual(t, accessToken, refreshToken)
}

func TestValidateToken_AccessToken(t *testing.T) {
	secret := "test-secret-key"
	service := NewJWTService(secret)

	user := &User{
		ID:          "123",
		Email:       "test@example.com",
		DisplayName: "Test User",
		Role:        RoleUser,
	}

	accessToken, _, err := service.GenerateToken(user)
	assert.NoError(t, err)

	claims, err := service.ValidateToken(accessToken)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
	assert.Equal(t, "access", claims.TokenType)
}

func TestValidateToken_RefreshToken(t *testing.T) {
	secret := "test-secret-key"
	service := NewJWTService(secret)

	user := &User{
		ID:          "123",
		Email:       "test@example.com",
		DisplayName: "Test User",
		Role:        RoleUser,
	}

	_, refreshToken, err := service.GenerateToken(user)
	assert.NoError(t, err)

	claims, err := service.ValidateToken(refreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
	assert.Equal(t, "refresh", claims.TokenType)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	secret := "test-secret-key"
	service := NewJWTService(secret)

	_, err := service.ValidateToken("invalid.token.here")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	secret1 := "test-secret-key-1"
	service1 := NewJWTService(secret1)

	user := &User{
		ID:          "123",
		Email:       "test@example.com",
		DisplayName: "Test User",
		Role:        RoleUser,
	}

	accessToken, _, err := service1.GenerateToken(user)
	assert.NoError(t, err)

	// Try to validate with different secret
	secret2 := "test-secret-key-2"
	service2 := NewJWTService(secret2)

	_, err = service2.ValidateToken(accessToken)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	secret := "test-secret-key"

	// Create a service with very short expiry
	service := &JWTService{
		secretKey:     []byte(secret),
		accessExpiry:  -time.Hour, // Expired
		refreshExpiry: time.Hour * 24 * 7,
	}

	user := &User{
		ID:          "123",
		Email:       "test@example.com",
		DisplayName: "Test User",
		Role:        RoleUser,
	}

	accessToken, _, err := service.GenerateToken(user)
	assert.NoError(t, err)

	_, err = service.ValidateToken(accessToken)
	assert.Error(t, err)
	assert.Equal(t, ErrExpiredToken, err)
}

func TestRefreshToken(t *testing.T) {
	secret := "test-secret-key"
	service := NewJWTService(secret)

	user := &User{
		ID:          "123",
		Email:       "test@example.com",
		DisplayName: "Test User",
		Role:        RoleUser,
	}

	_, refreshToken, err := service.GenerateToken(user)
	assert.NoError(t, err)

	newAccessToken, err := service.RefreshToken(refreshToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)
	assert.NotEqual(t, refreshToken, newAccessToken)

	// Verify the new token is valid
	claims, err := service.ValidateToken(newAccessToken)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, "access", claims.TokenType)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	secret := "test-secret-key"
	service := NewJWTService(secret)

	_, err := service.RefreshToken("invalid.token.here")
	assert.Error(t, err)
}

func TestRefreshToken_AccessToken(t *testing.T) {
	secret := "test-secret-key"
	service := NewJWTService(secret)

	user := &User{
		ID:          "123",
		Email:       "test@example.com",
		DisplayName: "Test User",
		Role:        RoleUser,
	}

	accessToken, _, err := service.GenerateToken(user)
	assert.NoError(t, err)

	// Try to refresh with access token instead of refresh token
	_, err = service.RefreshToken(accessToken)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestJWTClaims(t *testing.T) {
	secret := "test-secret-key"
	service := NewJWTService(secret)

	user := &User{
		ID:          "123",
		Email:       "test@example.com",
		DisplayName: "Test User",
		Role:        RoleUser,
	}

	accessToken, _, err := service.GenerateToken(user)
	assert.NoError(t, err)

	token, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	assert.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(*Claims)
	assert.True(t, ok)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
	assert.Equal(t, "access", claims.TokenType)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.ExpiresAt)
}

func TestDifferentSecrets(t *testing.T) {
	user := &User{
		ID:          "123",
		Email:       "test@example.com",
		DisplayName: "Test User",
		Role:        RoleUser,
	}

	secret1 := "test-secret-key-1"
	secret2 := "test-secret-key-2"

	service1 := NewJWTService(secret1)
	service2 := NewJWTService(secret2)

	token1, _, err := service1.GenerateToken(user)
	assert.NoError(t, err)

	// Validate with same service
	_, err = service1.ValidateToken(token1)
	assert.NoError(t, err)

	// Validate with different service
	_, err = service2.ValidateToken(token1)
	assert.Error(t, err)
}