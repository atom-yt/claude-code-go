package services

import (
	"context"
	"errors"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
)

var ErrScheduleNotFound = errors.New("scheduled task not found")

// ScheduleService handles scheduled task business logic
type ScheduleService struct {
	repo repository.ScheduleRepositoryI
}

// NewScheduleService creates a new schedule service
func NewScheduleService(repo repository.ScheduleRepositoryI) *ScheduleService {
	return &ScheduleService{repo: repo}
}

// CreateSchedule creates a new scheduled task
func (s *ScheduleService) CreateSchedule(ctx context.Context, userID string, req *models.CreateScheduledTaskRequest) (*models.ScheduledTask, error) {
	task := &models.ScheduledTask{
		UserID:       userID,
		Title:        req.Title,
		Prompt:       req.Prompt,
		ScheduleType: req.ScheduleType,
		ScheduleTime: req.ScheduleTime,
		Model:        req.Model,
	}

	if req.NotifyOnDone != nil {
		task.NotifyOnDone = *req.NotifyOnDone
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

// GetSchedule retrieves a scheduled task by ID
func (s *ScheduleService) GetSchedule(ctx context.Context, id string) (*models.ScheduledTask, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}
	return task, nil
}

// GetUserSchedules retrieves scheduled tasks for a user
func (s *ScheduleService) GetUserSchedules(ctx context.Context, userID string, page, pageSize int) (*models.ListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	tasks, total, err := s.repo.GetByUser(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.ListResponse{
		Items:      tasks,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateSchedule updates a scheduled task
func (s *ScheduleService) UpdateSchedule(ctx context.Context, id string, req *models.UpdateScheduledTaskRequest) (*models.ScheduledTask, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Prompt != nil {
		task.Prompt = *req.Prompt
	}
	if req.ScheduleType != nil {
		task.ScheduleType = *req.ScheduleType
	}
	if req.ScheduleTime != nil {
		task.ScheduleTime = *req.ScheduleTime
	}
	if req.Model != nil {
		task.Model = *req.Model
	}
	if req.Enabled != nil {
		task.Enabled = *req.Enabled
	}
	if req.NotifyOnDone != nil {
		task.NotifyOnDone = *req.NotifyOnDone
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

// DeleteSchedule deletes a scheduled task
func (s *ScheduleService) DeleteSchedule(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ToggleSchedule toggles the enabled status of a scheduled task
func (s *ScheduleService) ToggleSchedule(ctx context.Context, id string, enabled bool) error {
	return s.repo.ToggleEnabled(ctx, id, enabled)
}
