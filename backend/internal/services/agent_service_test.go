package services

import (
	"context"
	"testing"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAgentRepository for agent service tests
type MockAgentRepositoryForAgent struct {
	mock.Mock
}

func (m *MockAgentRepositoryForAgent) Create(ctx context.Context, agent *models.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentRepositoryForAgent) GetByID(ctx context.Context, id string) (*models.Agent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Agent), args.Error(1)
}

func (m *MockAgentRepositoryForAgent) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Agent, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Agent), args.Get(1).(int64), args.Error(2)
}

func (m *MockAgentRepositoryForAgent) Update(ctx context.Context, agent *models.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentRepositoryForAgent) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAgentRepositoryForAgent) List(ctx context.Context, limit, offset int) ([]*models.Agent, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Agent), args.Get(1).(int64), args.Error(2)
}

func (m *MockAgentRepositoryForAgent) GetDefaultAgent(ctx context.Context, userID string) (*models.Agent, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Agent), args.Error(1)
}

func TestCreateAgentDefaults(t *testing.T) {
	tests := []struct {
		name       string
		req        *models.CreateAgentRequest
		setupMocks func(*MockAgentRepositoryForAgent)
		wantTemp   float64
		wantTokens int
	}{
		{
			name: "with temperature and max tokens",
			req: &models.CreateAgentRequest{
				Name:         "Test Agent",
				SystemPrompt: "You are helpful",
				Model:        "gpt-4",
				Provider:     "openai",
				Temperature:  func() *float64 { v := 0.8; return &v }(),
				MaxTokens:    func() *int { v := 8192; return &v }(),
			},
			setupMocks: func(r *MockAgentRepositoryForAgent) {
				r.On("Create", mock.Anything, mock.AnythingOfType("*models.Agent")).Return(nil)
			},
			wantTemp:   0.8,
			wantTokens: 8192,
		},
		{
			name: "without temperature and max tokens",
			req: &models.CreateAgentRequest{
				Name:         "Test Agent",
				SystemPrompt: "You are helpful",
				Model:        "gpt-4",
				Provider:     "openai",
			},
			setupMocks: func(r *MockAgentRepositoryForAgent) {
				// Repository will set defaults to 0.7 and 4096
				r.On("Create", mock.Anything, mock.MatchedBy(func(a *models.Agent) bool {
					return a.Temperature == 0 && a.MaxTokens == 0
				})).Return(nil).Run(func(args mock.Arguments) {
					agent := args.Get(1).(*models.Agent)
					agent.Temperature = 0.7
					agent.MaxTokens = 4096
				})
			},
			wantTemp:   0.7,
			wantTokens: 4096,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockAgentRepositoryForAgent)
			tt.setupMocks(repo)

			service := NewAgentService(repo)
			agent, err := service.CreateAgent(context.Background(), "user-123", tt.req)

			assert.NoError(t, err)
			assert.NotNil(t, agent)
			assert.Equal(t, "user-123", agent.UserID)
			assert.Equal(t, tt.wantTemp, agent.Temperature)
			assert.Equal(t, tt.wantTokens, agent.MaxTokens)

			repo.AssertExpectations(t)
		})
	}
}

func TestUpdateAgentPartialFields(t *testing.T) {
	repo := new(MockAgentRepositoryForAgent)

	existingAgent := &models.Agent{
		ID:           "agent-123",
		UserID:       "user-123",
		Name:         "Old Name",
		Description:  "Old Description",
		SystemPrompt: "Old Prompt",
		Model:        "gpt-3.5",
		Provider:     "openai",
		Temperature:  0.7,
		MaxTokens:    4096,
	}

	repo.On("GetByID", mock.Anything, "agent-123").Return(existingAgent, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(a *models.Agent) bool {
		return a.Name == "New Name" && a.Description == "Old Description"
	})).Return(nil)

	service := NewAgentService(repo)
	newName := "New Name"
	req := &models.UpdateAgentRequest{Name: &newName}
	agent, err := service.UpdateAgent(context.Background(), "agent-123", req)

	assert.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "New Name", agent.Name)
	assert.Equal(t, "Old Description", agent.Description) // Unchanged

	repo.AssertExpectations(t)
}

func TestGetAgentNotFound(t *testing.T) {
	repo := new(MockAgentRepositoryForAgent)

	repo.On("GetByID", mock.Anything, "agent-404").Return(nil, repository.ErrNotFound)

	service := NewAgentService(repo)
	agent, err := service.GetAgent(context.Background(), "agent-404")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAgentNotFound)
	assert.Nil(t, agent)

	repo.AssertExpectations(t)
}

func TestGetUserAgentsPagination(t *testing.T) {
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
			repo := new(MockAgentRepositoryForAgent)

			expectedLimit := tt.wantPageSize
			expectedOffset := (tt.wantPage - 1) * tt.wantPageSize

			repo.On("GetByUser", mock.Anything, "user-123", expectedLimit, expectedOffset).
				Return([]*models.Agent{}, int64(0), nil)

			service := NewAgentService(repo)
			resp, err := service.GetUserAgents(context.Background(), "user-123", tt.page, tt.pageSize)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantPage, resp.Page)
			assert.Equal(t, tt.wantPageSize, resp.PageSize)

			repo.AssertExpectations(t)
		})
	}
}

func TestDeleteAgent(t *testing.T) {
	repo := new(MockAgentRepositoryForAgent)

	repo.On("Delete", mock.Anything, "agent-123").Return(nil)

	service := NewAgentService(repo)
	err := service.DeleteAgent(context.Background(), "agent-123")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetDefaultAgent(t *testing.T) {
	repo := new(MockAgentRepositoryForAgent)

	expectedAgent := &models.Agent{
		ID:     "agent-123",
		UserID: "user-123",
		Name:   "Default Agent",
	}

	repo.On("GetDefaultAgent", mock.Anything, "user-123").Return(expectedAgent, nil)

	service := NewAgentService(repo)
	agent, err := service.GetDefaultAgent(context.Background(), "user-123")

	assert.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "agent-123", agent.ID)

	repo.AssertExpectations(t)
}

func TestListAgents(t *testing.T) {
	repo := new(MockAgentRepositoryForAgent)

	agents := []*models.Agent{
		{ID: "agent-1", Name: "Agent 1"},
		{ID: "agent-2", Name: "Agent 2"},
	}

	repo.On("List", mock.Anything, 10, 0).Return(agents, int64(2), nil)

	service := NewAgentService(repo)
	resp, err := service.ListAgents(context.Background(), 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Items.([]*models.Agent), 2)

	repo.AssertExpectations(t)
}

func TestUpdateAgentNotFound(t *testing.T) {
	repo := new(MockAgentRepositoryForAgent)

	repo.On("GetByID", mock.Anything, "agent-404").Return(nil, repository.ErrNotFound)

	service := NewAgentService(repo)
	newName := "New Name"
	req := &models.UpdateAgentRequest{Name: &newName}
	agent, err := service.UpdateAgent(context.Background(), "agent-404", req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAgentNotFound)
	assert.Nil(t, agent)

	repo.AssertExpectations(t)
}