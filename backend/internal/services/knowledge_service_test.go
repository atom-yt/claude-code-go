package services

import (
	"context"
	"testing"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKnowledgeRepository for knowledge service tests
type MockKnowledgeRepositoryForKnowledge struct {
	mock.Mock
}

func (m *MockKnowledgeRepositoryForKnowledge) Create(ctx context.Context, kb *models.KnowledgeBase) error {
	args := m.Called(ctx, kb)
	return args.Error(0)
}

func (m *MockKnowledgeRepositoryForKnowledge) GetByID(ctx context.Context, id string) (*models.KnowledgeBase, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.KnowledgeBase), args.Error(1)
}

func (m *MockKnowledgeRepositoryForKnowledge) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.KnowledgeBase, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.KnowledgeBase), args.Get(1).(int64), args.Error(2)
}

func (m *MockKnowledgeRepositoryForKnowledge) Update(ctx context.Context, kb *models.KnowledgeBase) error {
	args := m.Called(ctx, kb)
	return args.Error(0)
}

func (m *MockKnowledgeRepositoryForKnowledge) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreateKnowledge(t *testing.T) {
	repo := new(MockKnowledgeRepositoryForKnowledge)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*models.KnowledgeBase")).Return(nil)

	service := NewKnowledgeService(repo)
	req := &models.CreateKnowledgeBaseRequest{
		Name:        "Test Knowledge",
		Description: "A test knowledge base",
		Type:        "document",
		Source:      "local",
	}
	kb, err := service.CreateKnowledge(context.Background(), "user-123", req)

	assert.NoError(t, err)
	assert.NotNil(t, kb)
	assert.Equal(t, "user-123", kb.UserID)
	assert.Equal(t, "Test Knowledge", kb.Name)

	repo.AssertExpectations(t)
}

func TestGetKnowledge(t *testing.T) {
	repo := new(MockKnowledgeRepositoryForKnowledge)

	expectedKB := &models.KnowledgeBase{
		ID:          "kb-123",
		UserID:      "user-123",
		Name:        "Test Knowledge",
		Description: "A test knowledge base",
		Type:        "document",
		Source:      "local",
	}

	repo.On("GetByID", mock.Anything, "kb-123").Return(expectedKB, nil)

	service := NewKnowledgeService(repo)
	kb, err := service.GetKnowledge(context.Background(), "kb-123")

	assert.NoError(t, err)
	assert.NotNil(t, kb)
	assert.Equal(t, "kb-123", kb.ID)

	repo.AssertExpectations(t)
}

func TestGetKnowledgeNotFound(t *testing.T) {
	repo := new(MockKnowledgeRepositoryForKnowledge)

	repo.On("GetByID", mock.Anything, "kb-404").Return(nil, repository.ErrNotFound)

	service := NewKnowledgeService(repo)
	kb, err := service.GetKnowledge(context.Background(), "kb-404")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrKnowledgeNotFound)
	assert.Nil(t, kb)

	repo.AssertExpectations(t)
}

func TestUpdateKnowledge(t *testing.T) {
	repo := new(MockKnowledgeRepositoryForKnowledge)

	existingKB := &models.KnowledgeBase{
		ID:          "kb-123",
		UserID:      "user-123",
		Name:        "Old Name",
		Description: "Old Description",
		Type:        "document",
		Source:      "local",
	}

	repo.On("GetByID", mock.Anything, "kb-123").Return(existingKB, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(kb *models.KnowledgeBase) bool {
		return kb.Name == "New Name"
	})).Return(nil)

	service := NewKnowledgeService(repo)
	newName := "New Name"
	req := &models.UpdateKnowledgeBaseRequest{Name: &newName}
	kb, err := service.UpdateKnowledge(context.Background(), "kb-123", req)

	assert.NoError(t, err)
	assert.NotNil(t, kb)
	assert.Equal(t, "New Name", kb.Name)

	repo.AssertExpectations(t)
}

func TestUpdateKnowledgeNotFound(t *testing.T) {
	repo := new(MockKnowledgeRepositoryForKnowledge)

	repo.On("GetByID", mock.Anything, "kb-404").Return(nil, repository.ErrNotFound)

	service := NewKnowledgeService(repo)
	newName := "New Name"
	req := &models.UpdateKnowledgeBaseRequest{Name: &newName}
	kb, err := service.UpdateKnowledge(context.Background(), "kb-404", req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrKnowledgeNotFound)
	assert.Nil(t, kb)

	repo.AssertExpectations(t)
}

func TestDeleteKnowledge(t *testing.T) {
	repo := new(MockKnowledgeRepositoryForKnowledge)

	repo.On("Delete", mock.Anything, "kb-123").Return(nil)

	service := NewKnowledgeService(repo)
	err := service.DeleteKnowledge(context.Background(), "kb-123")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetUserKnowledge(t *testing.T) {
	repo := new(MockKnowledgeRepositoryForKnowledge)

	kbs := []*models.KnowledgeBase{
		{ID: "kb-1", Name: "Knowledge 1", UserID: "user-123"},
		{ID: "kb-2", Name: "Knowledge 2", UserID: "user-123"},
	}

	repo.On("GetByUser", mock.Anything, "user-123", 10, 0).Return(kbs, int64(2), nil)

	service := NewKnowledgeService(repo)
	resp, err := service.GetUserKnowledge(context.Background(), "user-123", 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Items.([]*models.KnowledgeBase), 2)

	repo.AssertExpectations(t)
}

func TestGetUserKnowledgePagination(t *testing.T) {
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
			repo := new(MockKnowledgeRepositoryForKnowledge)

			expectedLimit := tt.wantPageSize
			expectedOffset := (tt.wantPage - 1) * tt.wantPageSize

			repo.On("GetByUser", mock.Anything, "user-123", expectedLimit, expectedOffset).
				Return([]*models.KnowledgeBase{}, int64(0), nil)

			service := NewKnowledgeService(repo)
			resp, err := service.GetUserKnowledge(context.Background(), "user-123", tt.page, tt.pageSize)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantPage, resp.Page)
			assert.Equal(t, tt.wantPageSize, resp.PageSize)

			repo.AssertExpectations(t)
		})
	}
}

func TestUpdateKnowledgePartial(t *testing.T) {
	repo := new(MockKnowledgeRepositoryForKnowledge)

	existingKB := &models.KnowledgeBase{
		ID:          "kb-123",
		UserID:      "user-123",
		Name:        "Old Name",
		Description: "Old Description",
		Type:        "document",
		Source:      "local",
	}

	repo.On("GetByID", mock.Anything, "kb-123").Return(existingKB, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(kb *models.KnowledgeBase) bool {
		return kb.Description == "New Description" && kb.Name == "Old Name" // Only description changed
	})).Return(nil)

	service := NewKnowledgeService(repo)
	newDescription := "New Description"
	req := &models.UpdateKnowledgeBaseRequest{Description: &newDescription}
	kb, err := service.UpdateKnowledge(context.Background(), "kb-123", req)

	assert.NoError(t, err)
	assert.Equal(t, "Old Name", kb.Name)
	assert.Equal(t, "New Description", kb.Description)

	repo.AssertExpectations(t)
}