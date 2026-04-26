package services

import (
	"context"
	"testing"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMessageRepository is a mock implementation of MessageRepository
type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(ctx context.Context, message *models.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) GetByID(ctx context.Context, id string) (*models.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockMessageRepository) GetBySession(ctx context.Context, sessionID string, limit, offset int) ([]*models.Message, int64, error) {
	args := m.Called(ctx, sessionID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Message), args.Get(1).(int64), args.Error(2)
}

func (m *MockMessageRepository) GetRecentBySession(ctx context.Context, sessionID string, limit int) ([]*models.Message, error) {
	args := m.Called(ctx, sessionID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockMessageRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMessageRepository) DeleteBySession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockMessageRepository) Count(ctx context.Context, sessionID string) (int64, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(int64), args.Error(1)
}

func TestCreateMessage(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockMessageRepository, *MockSessionRepository)
		sessionID   string
		req         *models.CreateMessageRequest
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful creation",
			setupMocks: func(mr *MockMessageRepository, sr *MockSessionRepository) {
				sr.On("GetByID", mock.Anything, "session-123").Return(&models.Session{ID: "session-123"}, nil)
				mr.On("Create", mock.Anything, mock.AnythingOfType("*models.Message")).Return(nil)
			},
			sessionID: "session-123",
			req: &models.CreateMessageRequest{
				Role:    "user",
				Content: map[string]interface{}{"text": "Hello"},
			},
			wantErr: false,
		},
		{
			name: "session not found",
			setupMocks: func(mr *MockMessageRepository, sr *MockSessionRepository) {
				sr.On("GetByID", mock.Anything, "session-404").Return(nil, repository.ErrNotFound)
			},
			sessionID:   "session-404",
			req:         &models.CreateMessageRequest{Role: "user", Content: map[string]interface{}{}},
			wantErr:     true,
			expectedErr: ErrSessionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messageRepo := new(MockMessageRepository)
			sessionRepo := new(MockSessionRepository)
			tt.setupMocks(messageRepo, sessionRepo)

			service := NewMessageService(messageRepo, sessionRepo)
			message, err := service.CreateMessage(context.Background(), tt.sessionID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Nil(t, message)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, message)
				assert.Equal(t, tt.sessionID, message.SessionID)
				assert.Equal(t, tt.req.Role, message.Role)
			}

			messageRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
		})
	}
}

func TestGetSessionMessagesPagination(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{"valid pagination", 2, 20, 2, 20},
		{"page < 1", 0, 20, 1, 20},
		{"pageSize < 1", 1, 0, 1, 10},
		{"pageSize > 100", 1, 150, 1, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messageRepo := new(MockMessageRepository)
			sessionRepo := new(MockSessionRepository)

			expectedLimit := tt.wantPageSize
			expectedOffset := (tt.wantPage - 1) * tt.wantPageSize

			messageRepo.On("GetBySession", mock.Anything, "session-123", expectedLimit, expectedOffset).
				Return([]*models.Message{}, int64(0), nil)

			service := NewMessageService(messageRepo, sessionRepo)
			resp, err := service.GetSessionMessages(context.Background(), "session-123", tt.page, tt.pageSize)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantPage, resp.Page)
			assert.Equal(t, tt.wantPageSize, resp.PageSize)

			messageRepo.AssertExpectations(t)
		})
	}
}

func TestGetRecentMessagesLimit(t *testing.T) {
	tests := []struct {
		name        string
		inputLimit  int
		expectedLimit int
	}{
		{"valid limit", 50, 50},
		{"limit < 1", 0, 10},
		{"limit > 100", 150, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messageRepo := new(MockMessageRepository)
			sessionRepo := new(MockSessionRepository)

			messageRepo.On("GetRecentBySession", mock.Anything, "session-123", tt.expectedLimit).
				Return([]*models.Message{}, nil)

			service := NewMessageService(messageRepo, sessionRepo)
			messages, err := service.GetRecentMessages(context.Background(), "session-123", tt.inputLimit)

			assert.NoError(t, err)
			assert.NotNil(t, messages)

			messageRepo.AssertExpectations(t)
		})
	}
}

func TestGetMessage(t *testing.T) {
	messageRepo := new(MockMessageRepository)
	sessionRepo := new(MockSessionRepository)

	expectedMessage := &models.Message{
		ID:        "message-123",
		SessionID: "session-123",
		Role:      "user",
		Content:   map[string]interface{}{"text": "Hello"},
	}

	messageRepo.On("GetByID", mock.Anything, "message-123").Return(expectedMessage, nil)

	service := NewMessageService(messageRepo, sessionRepo)
	message, err := service.GetMessage(context.Background(), "message-123")

	assert.NoError(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, "message-123", message.ID)

	messageRepo.AssertExpectations(t)
}

func TestGetMessageNotFound(t *testing.T) {
	messageRepo := new(MockMessageRepository)
	sessionRepo := new(MockSessionRepository)

	messageRepo.On("GetByID", mock.Anything, "message-404").Return(nil, repository.ErrNotFound)

	service := NewMessageService(messageRepo, sessionRepo)
	message, err := service.GetMessage(context.Background(), "message-404")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMessageNotFound)
	assert.Nil(t, message)

	messageRepo.AssertExpectations(t)
}

func TestDeleteMessage(t *testing.T) {
	messageRepo := new(MockMessageRepository)
	sessionRepo := new(MockSessionRepository)

	messageRepo.On("Delete", mock.Anything, "message-123").Return(nil)

	service := NewMessageService(messageRepo, sessionRepo)
	err := service.DeleteMessage(context.Background(), "message-123")

	assert.NoError(t, err)
	messageRepo.AssertExpectations(t)
}

func TestGetSessionMessagesWithTotalPages(t *testing.T) {
	messageRepo := new(MockMessageRepository)
	sessionRepo := new(MockSessionRepository)

	// Test total pages calculation
	// 25 messages with page size 10 should result in 3 pages
	messageRepo.On("GetBySession", mock.Anything, "session-123", 10, 0).
		Return([]*models.Message{}, int64(25), nil)

	service := NewMessageService(messageRepo, sessionRepo)
	resp, err := service.GetSessionMessages(context.Background(), "session-123", 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 3, resp.TotalPages)

	messageRepo.AssertExpectations(t)
}