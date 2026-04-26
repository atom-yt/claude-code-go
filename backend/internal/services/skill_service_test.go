package services

import (
	"context"
	"testing"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSkillRepository for skill service tests
type MockSkillRepositoryForSkill struct {
	mock.Mock
}

func (m *MockSkillRepositoryForSkill) Create(ctx context.Context, skill *models.Skill) error {
	args := m.Called(ctx, skill)
	return args.Error(0)
}

func (m *MockSkillRepositoryForSkill) GetByID(ctx context.Context, id string) (*models.Skill, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Skill), args.Error(1)
}

func (m *MockSkillRepositoryForSkill) GetByUser(ctx context.Context, userID string, category string, limit, offset int) ([]*models.Skill, int64, error) {
	args := m.Called(ctx, userID, category, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Skill), args.Get(1).(int64), args.Error(2)
}

func (m *MockSkillRepositoryForSkill) Update(ctx context.Context, skill *models.Skill) error {
	args := m.Called(ctx, skill)
	return args.Error(0)
}

func (m *MockSkillRepositoryForSkill) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSkillRepositoryForSkill) ToggleEnabled(ctx context.Context, id string, enabled bool) error {
	args := m.Called(ctx, id, enabled)
	return args.Error(0)
}

func TestCreateSkill(t *testing.T) {
	repo := new(MockSkillRepositoryForSkill)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Skill")).Return(nil)

	service := NewSkillService(repo)
	req := &models.CreateSkillRequest{
		Name:        "Test Skill",
		Description: "A test skill",
		Category:    "personal",
	}
	skill, err := service.CreateSkill(context.Background(), "user-123", req)

	assert.NoError(t, err)
	assert.NotNil(t, skill)
	assert.Equal(t, "user-123", *skill.UserID)
	assert.Equal(t, "Test Skill", skill.Name)

	repo.AssertExpectations(t)
}

func TestToggleSkill(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enable skill", true},
		{"disable skill", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockSkillRepositoryForSkill)

			repo.On("ToggleEnabled", mock.Anything, "skill-123", tt.enabled).Return(nil)

			service := NewSkillService(repo)
			err := service.ToggleSkill(context.Background(), "skill-123", tt.enabled)

			assert.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestGetUserSkillsByCategory(t *testing.T) {
	tests := []struct {
		name     string
		category string
	}{
		{"personal category", "personal"},
		{"team category", "team"},
		{"builtin category", "builtin"},
		{"no category filter", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockSkillRepositoryForSkill)

			skills := []*models.Skill{
				{ID: "skill-1", Name: "Skill 1", Category: tt.category},
			}

			repo.On("GetByUser", mock.Anything, "user-123", tt.category, 10, 0).
				Return(skills, int64(1), nil)

			service := NewSkillService(repo)
			resp, err := service.GetUserSkills(context.Background(), "user-123", tt.category, 1, 10)

			assert.NoError(t, err)
			assert.NotNil(t, resp)

			repo.AssertExpectations(t)
		})
	}
}

func TestGetSkillNotFound(t *testing.T) {
	repo := new(MockSkillRepositoryForSkill)

	repo.On("GetByID", mock.Anything, "skill-404").Return(nil, repository.ErrNotFound)

	service := NewSkillService(repo)
	skill, err := service.GetSkill(context.Background(), "skill-404")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSkillNotFound)
	assert.Nil(t, skill)

	repo.AssertExpectations(t)
}

func TestUpdateSkill(t *testing.T) {
	repo := new(MockSkillRepositoryForSkill)

	existingSkill := &models.Skill{
		ID:          "skill-123",
		UserID:      func() *string { s := "user-123"; return &s }(),
		Name:        "Old Name",
		Description: "Old Description",
		Enabled:     true,
	}

	repo.On("GetByID", mock.Anything, "skill-123").Return(existingSkill, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(s *models.Skill) bool {
		return s.Name == "New Name"
	})).Return(nil)

	service := NewSkillService(repo)
	newName := "New Name"
	req := &models.UpdateSkillRequest{Name: &newName}
	skill, err := service.UpdateSkill(context.Background(), "skill-123", req)

	assert.NoError(t, err)
	assert.NotNil(t, skill)
	assert.Equal(t, "New Name", skill.Name)

	repo.AssertExpectations(t)
}

func TestUpdateSkillToggle(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enable skill", true},
		{"disable skill", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockSkillRepositoryForSkill)

			existingSkill := &models.Skill{
				ID:      "skill-123",
				Enabled: !tt.enabled,
			}

			repo.On("GetByID", mock.Anything, "skill-123").Return(existingSkill, nil)
			repo.On("Update", mock.Anything, mock.MatchedBy(func(s *models.Skill) bool {
				return s.Enabled == tt.enabled
			})).Return(nil)

			service := NewSkillService(repo)
			req := &models.UpdateSkillRequest{Enabled: &tt.enabled}
			skill, err := service.UpdateSkill(context.Background(), "skill-123", req)

			assert.NoError(t, err)
			assert.Equal(t, tt.enabled, skill.Enabled)

			repo.AssertExpectations(t)
		})
	}
}

func TestDeleteSkill(t *testing.T) {
	repo := new(MockSkillRepositoryForSkill)

	repo.On("Delete", mock.Anything, "skill-123").Return(nil)

	service := NewSkillService(repo)
	err := service.DeleteSkill(context.Background(), "skill-123")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetUserSkillsPagination(t *testing.T) {
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
			repo := new(MockSkillRepositoryForSkill)

			expectedLimit := tt.wantPageSize
			expectedOffset := (tt.wantPage - 1) * tt.wantPageSize

			repo.On("GetByUser", mock.Anything, "user-123", "", expectedLimit, expectedOffset).
				Return([]*models.Skill{}, int64(0), nil)

			service := NewSkillService(repo)
			resp, err := service.GetUserSkills(context.Background(), "user-123", "", tt.page, tt.pageSize)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantPage, resp.Page)
			assert.Equal(t, tt.wantPageSize, resp.PageSize)

			repo.AssertExpectations(t)
		})
	}
}