package auth

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDatabaseError    = errors.New("database error")
)

// UserStore defines the interface for user persistence
type UserStore interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
}

// Service provides authentication operations
type Service struct {
	jwtService *JWTService
	store      UserStore
}

// NewService creates a new authentication service
func NewService(jwtSecret string, store UserStore) *Service {
	return &Service{
		jwtService: NewJWTService(jwtSecret),
		store:      store,
	}
}

// Register registers a new user and returns authentication tokens
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	existingUser, err := s.store.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrUserExists
	}

	// Hash password
	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		Email:        req.Email,
		PasswordHash: hash,
		DisplayName:  req.DisplayName,
		Role:         RoleUser,
	}

	if err := s.store.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Generate tokens
	accessToken, refreshToken, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login authenticates a user and returns authentication tokens
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	// Get user by email
	user, err := s.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Verify password
	if !VerifyPassword(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, refreshToken, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Refresh generates a new access token from a refresh token
func (s *Service) Refresh(refreshToken string) (*RefreshResponse, error) {
	accessToken, err := s.jwtService.RefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return &RefreshResponse{
		AccessToken: accessToken,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, userID string) (*User, error) {
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// ValidateToken validates an access token and returns claims
func (s *Service) ValidateToken(token string) (*Claims, error) {
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}