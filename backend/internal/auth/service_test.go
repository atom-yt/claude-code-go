package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockUserStore implements UserStore for testing
type MockUserStore struct {
	users       map[string]*User
	getByIDErr  error
	getByEmailErr error
	createErr   error
}

func NewMockUserStore() *MockUserStore {
	return &MockUserStore{
		users: make(map[string]*User),
	}
}

func (m *MockUserStore) CreateUser(ctx context.Context, user *User) error {
	if m.createErr != nil {
		return m.createErr
	}
	// Set ID if empty
	if user.ID == "" {
		user.ID = "user-" + user.Email
	}
	m.users[user.Email] = user
	return nil
}

func (m *MockUserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}
	user, ok := m.users[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserStore) GetUserByID(ctx context.Context, id string) (*User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func TestNewService(t *testing.T) {
	store := NewMockUserStore()
	secret := "test-secret"

	service := NewService(secret, store)

	assert.NotNil(t, service)
	assert.NotNil(t, service.jwtService)
	assert.Equal(t, store, service.store)
}

func TestService_Register(t *testing.T) {
	store := NewMockUserStore()
	service := NewService("test-secret", store)

	req := &RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}

	response, err := service.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotNil(t, response.User)
	assert.Equal(t, req.Email, response.User.Email)
	assert.Equal(t, req.DisplayName, response.User.DisplayName)
	assert.Equal(t, RoleUser, response.User.Role)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
}

func TestService_Register_UserAlreadyExists(t *testing.T) {
	store := NewMockUserStore()
	service := NewService("test-secret", store)

	req := &RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}

	// Register once
	_, err := service.Register(context.Background(), req)
	assert.NoError(t, err)

	// Try to register again with same email
	_, err = service.Register(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, ErrUserExists, err)
}

func TestService_Login(t *testing.T) {
	store := NewMockUserStore()
	service := NewService("test-secret", store)

	// Register a user first
	regReq := &RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}
	_, err := service.Register(context.Background(), regReq)
	assert.NoError(t, err)

	// Login with correct credentials
	loginReq := &LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	response, err := service.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, regReq.Email, response.User.Email)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
}

func TestService_Login_UserNotFound(t *testing.T) {
	store := NewMockUserStore()
	service := NewService("test-secret", store)

	loginReq := &LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	_, err := service.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestService_Login_WrongPassword(t *testing.T) {
	store := NewMockUserStore()
	service := NewService("test-secret", store)

	// Register a user first
	regReq := &RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}
	_, err := service.Register(context.Background(), regReq)
	assert.NoError(t, err)

	// Login with wrong password
	loginReq := &LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	_, err = service.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
}

func TestService_Refresh(t *testing.T) {
	store := NewMockUserStore()
	service := NewService("test-secret", store)

	// Register a user to get tokens
	regReq := &RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}
	regResponse, err := service.Register(context.Background(), regReq)
	assert.NoError(t, err)

	// Refresh the token
	refreshResponse, err := service.Refresh(regResponse.RefreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, refreshResponse)
	assert.NotEmpty(t, refreshResponse.AccessToken)

	// Verify the new token is valid
	claims, err := service.ValidateToken(refreshResponse.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, regReq.Email, claims.Email)
	assert.Equal(t, "access", claims.TokenType)
}

func TestService_Refresh_InvalidToken(t *testing.T) {
	store := NewMockUserStore()
	service := NewService("test-secret", store)

	_, err := service.Refresh("invalid-token")

	assert.Error(t, err)
}

func TestService_Refresh_AccessToken(t *testing.T) {
	store := NewMockUserStore()
	service := NewService("test-secret", store)

	// Register a user to get tokens
	regReq := &RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}
	regResponse, err := service.Register(context.Background(), regReq)
	assert.NoError(t, err)

	// Try to refresh with access token instead of refresh token
	_, err = service.Refresh(regResponse.AccessToken)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestService_GetUserByID(t *testing.T) {
	store := NewMockUserStore()
	service := NewService("test-secret", store)

	// Register a user first
	regReq := &RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}
	regResponse, err := service.Register(context.Background(), regReq)
	assert.NoError(t, err)

	// Get user by ID
	user, err := service.GetUserByID(context.Background(), regResponse.User.ID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, regReq.Email, user.Email)
}

func TestService_GetUserByID_NotFound(t *testing.T) {
	store := NewMockUserStore()
	service := NewService("test-secret", store)

	_, err := service.GetUserByID(context.Background(), "nonexistent-id")

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestService_ValidateToken(t *testing.T) {
	store := NewMockUserStore()
	service := NewService("test-secret", store)

	// Register a user to get tokens
	regReq := &RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}
	regResponse, err := service.Register(context.Background(), regReq)
	assert.NoError(t, err)

	// Validate the access token
	claims, err := service.ValidateToken(regResponse.AccessToken)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, regResponse.User.ID, claims.UserID)
	assert.Equal(t, regReq.Email, claims.Email)
	assert.Equal(t, "access", claims.TokenType)
}

func TestService_ValidateToken_InvalidToken(t *testing.T) {
	store := NewMockUserStore()
	service := NewService("test-secret", store)

	_, err := service.ValidateToken("invalid-token")

	assert.Error(t, err)
}

func TestService_ValidateToken_RefreshToken(t *testing.T) {
	store := NewMockUserStore()
	service := NewService("test-secret", store)

	// Register a user to get tokens
	regReq := &RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}
	regResponse, err := service.Register(context.Background(), regReq)
	assert.NoError(t, err)

	// Validate refresh token should fail (only access tokens allowed)
	_, err = service.ValidateToken(regResponse.RefreshToken)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestService_DatabaseError(t *testing.T) {
	store := NewMockUserStore()
	store.createErr = errors.New("database error")
	service := NewService("test-secret", store)

	req := &RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
	}

	_, err := service.Register(context.Background(), req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDatabaseError)
}