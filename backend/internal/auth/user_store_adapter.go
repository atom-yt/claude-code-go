package auth

import (
	"context"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
)

// UserStoreAdapter adapts repository.UserRepository to auth.UserStore interface
type UserStoreAdapter struct {
	repo *repository.UserRepository
}

// NewUserStoreAdapter creates a new user store adapter
func NewUserStoreAdapter(repo *repository.UserRepository) *UserStoreAdapter {
	return &UserStoreAdapter{repo: repo}
}

// CreateUser creates a new user
func (a *UserStoreAdapter) CreateUser(ctx context.Context, user *User) error {
	repoUser := &models.User{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		DisplayName:  user.DisplayName,
		Role:         string(user.Role),
	}
	return a.repo.Create(ctx, repoUser)
}

// GetUserByEmail retrieves a user by email
func (a *UserStoreAdapter) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	repoUser, err := a.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:           repoUser.ID,
		Email:        repoUser.Email,
		PasswordHash: repoUser.PasswordHash,
		DisplayName:  repoUser.DisplayName,
		Role:         Role(repoUser.Role),
		CreatedAt:    repoUser.CreatedAt,
		UpdatedAt:    repoUser.UpdatedAt,
	}, nil
}

// GetUserByID retrieves a user by ID
func (a *UserStoreAdapter) GetUserByID(ctx context.Context, id string) (*User, error) {
	repoUser, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:           repoUser.ID,
		Email:        repoUser.Email,
		PasswordHash: repoUser.PasswordHash,
		DisplayName:  repoUser.DisplayName,
		Role:         Role(repoUser.Role),
		CreatedAt:    repoUser.CreatedAt,
		UpdatedAt:    repoUser.UpdatedAt,
	}, nil
}