package services

import (
	"context"
	"testing"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockArtifactRepository for artifact service tests
type MockArtifactRepositoryForArtifact struct {
	mock.Mock
}

func (m *MockArtifactRepositoryForArtifact) Create(ctx context.Context, artifact *models.Artifact) error {
	args := m.Called(ctx, artifact)
	return args.Error(0)
}

func (m *MockArtifactRepositoryForArtifact) GetByID(ctx context.Context, id string) (*models.Artifact, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Artifact), args.Error(1)
}

func (m *MockArtifactRepositoryForArtifact) GetByUser(ctx context.Context, userID string, search string, limit, offset int) ([]*models.Artifact, int64, error) {
	args := m.Called(ctx, userID, search, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Artifact), args.Get(1).(int64), args.Error(2)
}

func (m *MockArtifactRepositoryForArtifact) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockArtifactRepositoryForArtifact) GetStats(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func TestCreateArtifact(t *testing.T) {
	repo := new(MockArtifactRepositoryForArtifact)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Artifact")).Return(nil)

	service := NewArtifactService(repo)
	req := &models.CreateArtifactRequest{
		Title:    "Test Artifact",
		Content:  "Test content",
		FileType: "text",
	}
	artifact, err := service.CreateArtifact(context.Background(), "user-123", req)

	assert.NoError(t, err)
	assert.NotNil(t, artifact)
	assert.Equal(t, "user-123", artifact.UserID)
	assert.Equal(t, "Test Artifact", artifact.Title)

	repo.AssertExpectations(t)
}

func TestCreateArtifactWithSessionID(t *testing.T) {
	repo := new(MockArtifactRepositoryForArtifact)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(a *models.Artifact) bool {
		return a.SessionID != nil && *a.SessionID == "session-123"
	})).Return(nil)

	service := NewArtifactService(repo)
	req := &models.CreateArtifactRequest{
		Title:     "Test Artifact",
		Content:   "Test content",
		FileType:   "text",
		SessionID:  "session-123",
	}
	artifact, err := service.CreateArtifact(context.Background(), "user-123", req)

	assert.NoError(t, err)
	assert.NotNil(t, artifact)
	assert.Equal(t, "session-123", *artifact.SessionID)

	repo.AssertExpectations(t)
}

func TestGetArtifactNotFound(t *testing.T) {
	repo := new(MockArtifactRepositoryForArtifact)

	repo.On("GetByID", mock.Anything, "artifact-404").Return(nil, repository.ErrNotFound)

	service := NewArtifactService(repo)
	artifact, err := service.GetArtifact(context.Background(), "artifact-404")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrArtifactNotFound)
	assert.Nil(t, artifact)

	repo.AssertExpectations(t)
}

func TestGetUserArtifactsSearchFilter(t *testing.T) {
	tests := []struct {
		name         string
		search       string
		expectSearch string
	}{
		{"with search term", "test", "test"},
		{"empty search", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockArtifactRepositoryForArtifact)

			repo.On("GetByUser", mock.Anything, "user-123", tt.expectSearch, 10, 0).
				Return([]*models.Artifact{}, int64(0), nil)

			service := NewArtifactService(repo)
			resp, err := service.GetUserArtifacts(context.Background(), "user-123", tt.search, 1, 10)

			assert.NoError(t, err)
			assert.NotNil(t, resp)

			repo.AssertExpectations(t)
		})
	}
}

func TestGetUserArtifactsPagination(t *testing.T) {
	tests := []struct {
		name        string
		page        int
		pageSize    int
		wantPage    int
		wantPageSize int
	}{
		{"valid pagination", 2, 20, 2, 20},
		{"page < 1", 0, 20, 1, 20},
		{"pageSize < 1", 1, 0, 1, 10},
		{"pageSize > 100", 1, 150, 1, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockArtifactRepositoryForArtifact)

			expectedLimit := tt.wantPageSize
			expectedOffset := (tt.wantPage - 1) * tt.wantPageSize

			repo.On("GetByUser", mock.Anything, "user-123", "", expectedLimit, expectedOffset).
				Return([]*models.Artifact{}, int64(0), nil)

			service := NewArtifactService(repo)
			resp, err := service.GetUserArtifacts(context.Background(), "user-123", "", tt.page, tt.pageSize)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantPage, resp.Page)
			assert.Equal(t, tt.wantPageSize, resp.PageSize)

			repo.AssertExpectations(t)
		})
	}
}

func TestGetArtifactStats(t *testing.T) {
	repo := new(MockArtifactRepositoryForArtifact)

	repo.On("GetStats", mock.Anything, "user-123").Return(int64(42), nil)

	service := NewArtifactService(repo)
	stats, err := service.GetArtifactStats(context.Background(), "user-123")

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(42), stats["total"])

	repo.AssertExpectations(t)
}

func TestDeleteArtifact(t *testing.T) {
	repo := new(MockArtifactRepositoryForArtifact)

	repo.On("Delete", mock.Anything, "artifact-123").Return(nil)

	service := NewArtifactService(repo)
	err := service.DeleteArtifact(context.Background(), "artifact-123")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetArtifact(t *testing.T) {
	repo := new(MockArtifactRepositoryForArtifact)

	expectedArtifact := &models.Artifact{
		ID:       "artifact-123",
		UserID:   "user-123",
		Title:    "Test Artifact",
		Content:  "Test content",
		FileType: "text",
	}

	repo.On("GetByID", mock.Anything, "artifact-123").Return(expectedArtifact, nil)

	service := NewArtifactService(repo)
	artifact, err := service.GetArtifact(context.Background(), "artifact-123")

	assert.NoError(t, err)
	assert.NotNil(t, artifact)
	assert.Equal(t, "artifact-123", artifact.ID)

	repo.AssertExpectations(t)
}