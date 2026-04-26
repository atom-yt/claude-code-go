package services

import (
	"context"
	"errors"
	"testing"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSessionRepository is a mock implementation of SessionRepository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id string) (*models.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Session, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Session), args.Get(1).(int64), args.Error(2)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) Archive(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) GetActiveSessions(ctx context.Context, userID string) ([]*models.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

// MockAgentRepository is a mock implementation of AgentRepository
type MockAgentRepository struct {
	mock.Mock
}

func (m *MockAgentRepository) Create(ctx context.Context, agent *models.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentRepository) GetByID(ctx context.Context, id string) (*models.Agent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Agent), args.Error(1)
}

func (m *MockAgentRepository) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Agent, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Agent), args.Get(1).(int64), args.Error(2)
}

func (m *MockAgentRepository) Update(ctx context.Context, agent *models.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAgentRepository) List(ctx context.Context, limit, offset int) ([]*models.Agent, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Agent), args.Get(1).(int64), args.Error(2)
}

func (m *MockAgentRepository) GetDefaultAgent(ctx context.Context, userID string) (*models.Agent, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Agent), args.Error(1)
}

func TestCreateSession(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockSessionRepository, *MockAgentRepository)
		req         *models.CreateSessionRequest
		userID      string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful creation",
			setupMocks: func(sr *MockSessionRepository, ar *MockAgentRepository) {
				ar.On("GetByID", mock.Anything, "agent-123").Return(&models.Agent{ID: "agent-123"}, nil)
				sr.On("Create", mock.Anything, mock.AnythingOfType("*models.Session")).Return(nil)
			},
			req:    &models.CreateSessionRequest{AgentID: "agent-123", Title: "Test Session"},
			userID: "user-123",
			wantErr: false,
		},
		{
			name: "agent not found",
			setupMocks: func(sr *MockSessionRepository, ar *MockAgentRepository) {
				ar.On("GetByID", mock.Anything, "agent-404").Return(nil, repository.ErrNotFound)
			},
			req:         &models.CreateSessionRequest{AgentID: "agent-404"},
			userID:      "user-123",
			wantErr:     true,
			expectedErr: ErrAgentNotFound,
		},
		{
			name: "repository error",
			setupMocks: func(sr *MockSessionRepository, ar *MockAgentRepository) {
				ar.On("GetByID", mock.Anything, "agent-123").Return(&models.Agent{ID: "agent-123"}, nil)
				sr.On("Create", mock.Anything, mock.AnythingOfType("*models.Session")).Return(errors.New("db error"))
			},
			req:    &models.CreateSessionRequest{AgentID: "agent-123"},
			userID: "user-123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := new(MockSessionRepository)
			agentRepo := new(MockAgentRepository)
			tt.setupMocks(sessionRepo, agentRepo)

			service := NewSessionService(sessionRepo, agentRepo)
			session, err := service.CreateSession(context.Background(), tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Nil(t, session)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, session)
				assert.Equal(t, tt.userID, session.UserID)
				assert.Equal(t, tt.req.AgentID, session.AgentID)
			}

			sessionRepo.AssertExpectations(t)
			agentRepo.AssertExpectations(t)
		})
	}
}

func TestGetUserSessionsPagination(t *testing.T) {
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
		{"negative values", -5, -10, 1, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := new(MockSessionRepository)
			agentRepo := new(MockAgentRepository)

			expectedLimit := tt.wantPageSize
			expectedOffset := (tt.wantPage - 1) * tt.wantPageSize

			sessionRepo.On("GetByUser", mock.Anything, "user-123", expectedLimit, expectedOffset).
				Return([]*models.Session{}, int64(0), nil)

			service := NewSessionService(sessionRepo, agentRepo)
			resp, err := service.GetUserSessions(context.Background(), "user-123", tt.page, tt.pageSize)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantPage, resp.Page)
			assert.Equal(t, tt.wantPageSize, resp.PageSize)

			sessionRepo.AssertExpectations(t)
		})
	}
}

func TestGetSession(t *testing.T) {
	sessionRepo := new(MockSessionRepository)
	agentRepo := new(MockAgentRepository)

	expectedSession := &models.Session{
		ID:      "session-123",
		UserID:  "user-123",
		AgentID: "agent-123",
		Title:   "Test Session",
	}

	sessionRepo.On("GetByID", mock.Anything, "session-123").Return(expectedSession, nil)

	service := NewSessionService(sessionRepo, agentRepo)
	session, err := service.GetSession(context.Background(), "session-123")

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "session-123", session.ID)

	sessionRepo.AssertExpectations(t)
}

func TestGetSessionNotFound(t *testing.T) {
	sessionRepo := new(MockSessionRepository)
	agentRepo := new(MockAgentRepository)

	sessionRepo.On("GetByID", mock.Anything, "session-404").Return(nil, repository.ErrNotFound)

	service := NewSessionService(sessionRepo, agentRepo)
	session, err := service.GetSession(context.Background(), "session-404")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSessionNotFound)
	assert.Nil(t, session)

	sessionRepo.AssertExpectations(t)
}

func TestUpdateSession(t *testing.T) {
	sessionRepo := new(MockSessionRepository)
	agentRepo := new(MockAgentRepository)

	existingSession := &models.Session{
		ID:      "session-123",
		UserID:  "user-123",
		AgentID: "agent-123",
		Title:   "Old Title",
	}

	sessionRepo.On("GetByID", mock.Anything, "session-123").Return(existingSession, nil)
	sessionRepo.On("Update", mock.Anything, mock.MatchedBy(func(s *models.Session) bool {
		return s.Title == "New Title"
	})).Return(nil)

	service := NewSessionService(sessionRepo, agentRepo)
	newTitle := "New Title"
	req := &models.UpdateSessionRequest{Title: &newTitle}
	session, err := service.UpdateSession(context.Background(), "session-123", req)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "New Title", session.Title)

	sessionRepo.AssertExpectations(t)
}

func TestDeleteSession(t *testing.T) {
	sessionRepo := new(MockSessionRepository)
	agentRepo := new(MockAgentRepository)

	sessionRepo.On("Delete", mock.Anything, "session-123").Return(nil)

	service := NewSessionService(sessionRepo, agentRepo)
	err := service.DeleteSession(context.Background(), "session-123")

	assert.NoError(t, err)
	sessionRepo.AssertExpectations(t)
}

func TestArchiveSession(t *testing.T) {
	sessionRepo := new(MockSessionRepository)
	agentRepo := new(MockAgentRepository)

	sessionRepo.On("Archive", mock.Anything, "session-123").Return(nil)

	service := NewSessionService(sessionRepo, agentRepo)
	err := service.ArchiveSession(context.Background(), "session-123")

	assert.NoError(t, err)
	sessionRepo.AssertExpectations(t)
}

func TestGetActiveSessions(t *testing.T) {
	sessionRepo := new(MockSessionRepository)
	agentRepo := new(MockAgentRepository)

	expectedSessions := []*models.Session{
		{ID: "session-1", Status: "active"},
		{ID: "session-2", Status: "active"},
	}

	sessionRepo.On("GetActiveSessions", mock.Anything, "user-123").Return(expectedSessions, nil)

	service := NewSessionService(sessionRepo, agentRepo)
	sessions, err := service.GetActiveSessions(context.Background(), "user-123")

	assert.NoError(t, err)
	assert.Len(t, sessions, 2)
	assert.Equal(t, "session-1", sessions[0].ID)

	sessionRepo.AssertExpectations(t)
}

func TestGetUserSessionsWithTotalPages(t *testing.T) {
	sessionRepo := new(MockSessionRepository)
	agentRepo := new(MockAgentRepository)

	// Test total pages calculation
	// 25 sessions with page size 10 should result in 3 pages
	sessionRepo.On("GetByUser", mock.Anything, "user-123", 10, 0).
		Return([]*models.Session{}, int64(25), nil)

	service := NewSessionService(sessionRepo, agentRepo)
	resp, err := service.GetUserSessions(context.Background(), "user-123", 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 3, resp.TotalPages)

	sessionRepo.AssertExpectations(t)
}
