package services

import (
	"context"
	"testing"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockScheduleRepository for schedule service tests
type MockScheduleRepositoryForSchedule struct {
	mock.Mock
}

func (m *MockScheduleRepositoryForSchedule) Create(ctx context.Context, task *models.ScheduledTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockScheduleRepositoryForSchedule) GetByID(ctx context.Context, id string) (*models.ScheduledTask, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ScheduledTask), args.Error(1)
}

func (m *MockScheduleRepositoryForSchedule) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.ScheduledTask, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.ScheduledTask), args.Get(1).(int64), args.Error(2)
}

func (m *MockScheduleRepositoryForSchedule) Update(ctx context.Context, task *models.ScheduledTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockScheduleRepositoryForSchedule) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockScheduleRepositoryForSchedule) ToggleEnabled(ctx context.Context, id string, enabled bool) error {
	args := m.Called(ctx, id, enabled)
	return args.Error(0)
}

func TestCreateSchedule(t *testing.T) {
	tests := []struct {
		name       string
		req        *models.CreateScheduledTaskRequest
		setupMocks func(*MockScheduleRepositoryForSchedule)
		wantModel  string
		wantEnabled bool
	}{
		{
			name: "with explicit model",
			req: &models.CreateScheduledTaskRequest{
				Title:        "Test Task",
				Prompt:       "Test prompt",
				ScheduleType: "daily",
				ScheduleTime: "09:00",
				Model:        "gpt-4",
			},
			setupMocks: func(r *MockScheduleRepositoryForSchedule) {
				// Repository will set defaults: enabled=true, model=gpt-4
				r.On("Create", mock.Anything, mock.AnythingOfType("*models.ScheduledTask")).Return(nil).Run(func(args mock.Arguments) {
					task := args.Get(1).(*models.ScheduledTask)
					task.Enabled = true
					if task.Model == "" {
						task.Model = "auto"
					}
				})
			},
			wantModel:  "gpt-4",
			wantEnabled: true,
		},
		{
			name: "without model (default to auto)",
			req: &models.CreateScheduledTaskRequest{
				Title:        "Test Task",
				Prompt:       "Test prompt",
				ScheduleType: "daily",
				ScheduleTime: "09:00",
			},
			setupMocks: func(r *MockScheduleRepositoryForSchedule) {
				// Repository will set defaults: enabled=true, model=auto
				r.On("Create", mock.Anything, mock.MatchedBy(func(t *models.ScheduledTask) bool {
					return t.Model == "" && t.Enabled == false
				})).Return(nil).Run(func(args mock.Arguments) {
					task := args.Get(1).(*models.ScheduledTask)
					task.Model = "auto"
					task.Enabled = true
				})
			},
			wantModel:  "auto",
			wantEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockScheduleRepositoryForSchedule)
			tt.setupMocks(repo)

			service := NewScheduleService(repo)
			task, err := service.CreateSchedule(context.Background(), "user-123", tt.req)

			assert.NoError(t, err)
			assert.NotNil(t, task)
			assert.Equal(t, "user-123", task.UserID)
			assert.Equal(t, tt.wantModel, task.Model)
			assert.Equal(t, tt.wantEnabled, task.Enabled)

			repo.AssertExpectations(t)
		})
	}
}

func TestToggleSchedule(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enable schedule", true},
		{"disable schedule", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockScheduleRepositoryForSchedule)

			repo.On("ToggleEnabled", mock.Anything, "schedule-123", tt.enabled).Return(nil)

			service := NewScheduleService(repo)
			err := service.ToggleSchedule(context.Background(), "schedule-123", tt.enabled)

			assert.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestUpdateSchedule(t *testing.T) {
	repo := new(MockScheduleRepositoryForSchedule)

	existingTask := &models.ScheduledTask{
		ID:             "schedule-123",
		UserID:         "user-123",
		Title:          "Old Title",
		Prompt:         "Old Prompt",
		ScheduleType:   "daily",
		ScheduleTime:   "09:00",
		Model:          "gpt-4",
		Enabled:        true,
		NotifyOnDone:   true,
		ExecutionCount: 0,
	}

	repo.On("GetByID", mock.Anything, "schedule-123").Return(existingTask, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(t *models.ScheduledTask) bool {
		return t.Title == "New Title"
	})).Return(nil)

	service := NewScheduleService(repo)
	newTitle := "New Title"
	req := &models.UpdateScheduledTaskRequest{Title: &newTitle}
	task, err := service.UpdateSchedule(context.Background(), "schedule-123", req)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "New Title", task.Title)

	repo.AssertExpectations(t)
}

func TestUpdateSchedulePartial(t *testing.T) {
	repo := new(MockScheduleRepositoryForSchedule)

	existingTask := &models.ScheduledTask{
		ID:             "schedule-123",
		UserID:         "user-123",
		Title:          "Old Title",
		Prompt:         "Old Prompt",
		ScheduleType:   "daily",
		ScheduleTime:   "09:00",
		Model:          "gpt-4",
		Enabled:        true,
		NotifyOnDone:   true,
		ExecutionCount: 0,
	}

	repo.On("GetByID", mock.Anything, "schedule-123").Return(existingTask, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(t *models.ScheduledTask) bool {
		return t.Prompt == "New Prompt" && t.Title == "Old Title" // Only prompt changed
	})).Return(nil)

	service := NewScheduleService(repo)
	newPrompt := "New Prompt"
	req := &models.UpdateScheduledTaskRequest{Prompt: &newPrompt}
	task, err := service.UpdateSchedule(context.Background(), "schedule-123", req)

	assert.NoError(t, err)
	assert.Equal(t, "Old Title", task.Title)
	assert.Equal(t, "New Prompt", task.Prompt)

	repo.AssertExpectations(t)
}

func TestGetScheduleNotFound(t *testing.T) {
	repo := new(MockScheduleRepositoryForSchedule)

	repo.On("GetByID", mock.Anything, "schedule-404").Return(nil, repository.ErrNotFound)

	service := NewScheduleService(repo)
	task, err := service.GetSchedule(context.Background(), "schedule-404")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrScheduleNotFound)
	assert.Nil(t, task)

	repo.AssertExpectations(t)
}

func TestDeleteSchedule(t *testing.T) {
	repo := new(MockScheduleRepositoryForSchedule)

	repo.On("Delete", mock.Anything, "schedule-123").Return(nil)

	service := NewScheduleService(repo)
	err := service.DeleteSchedule(context.Background(), "schedule-123")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestGetUserSchedulesPagination(t *testing.T) {
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
			repo := new(MockScheduleRepositoryForSchedule)

			expectedLimit := tt.wantPageSize
			expectedOffset := (tt.wantPage - 1) * tt.wantPageSize

			repo.On("GetByUser", mock.Anything, "user-123", expectedLimit, expectedOffset).
				Return([]*models.ScheduledTask{}, int64(0), nil)

			service := NewScheduleService(repo)
			resp, err := service.GetUserSchedules(context.Background(), "user-123", tt.page, tt.pageSize)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantPage, resp.Page)
			assert.Equal(t, tt.wantPageSize, resp.PageSize)

			repo.AssertExpectations(t)
		})
	}
}

func TestUpdateScheduleNotFound(t *testing.T) {
	repo := new(MockScheduleRepositoryForSchedule)

	repo.On("GetByID", mock.Anything, "schedule-404").Return(nil, repository.ErrNotFound)

	service := NewScheduleService(repo)
	newTitle := "New Title"
	req := &models.UpdateScheduledTaskRequest{Title: &newTitle}
	task, err := service.UpdateSchedule(context.Background(), "schedule-404", req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrScheduleNotFound)
	assert.Nil(t, task)

	repo.AssertExpectations(t)
}

func TestCreateScheduleWithCronExpression(t *testing.T) {
	repo := new(MockScheduleRepositoryForSchedule)

	req := &models.CreateScheduledTaskRequest{
		Title:        "Test Task",
		Prompt:       "Test prompt",
		ScheduleType: "cron",
		ScheduleTime: "0 9 * * *", // Every day at 9am
		Model:        "gpt-4",
	}

	repo.On("Create", mock.Anything, mock.MatchedBy(func(t *models.ScheduledTask) bool {
		return t.ScheduleType == "cron" && t.ScheduleTime == "0 9 * * *"
	})).Return(nil).Run(func(args mock.Arguments) {
		task := args.Get(1).(*models.ScheduledTask)
		task.Enabled = true
	})

	service := NewScheduleService(repo)
	task, err := service.CreateSchedule(context.Background(), "user-123", req)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "cron", task.ScheduleType)
	assert.Equal(t, "0 9 * * *", task.ScheduleTime)

	repo.AssertExpectations(t)
}