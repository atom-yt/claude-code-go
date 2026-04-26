package services

import (
	"context"
	"errors"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
)

var ErrSkillNotFound = errors.New("skill not found")

// SkillService handles skill business logic
type SkillService struct {
	repo *repository.SkillRepository
}

// NewSkillService creates a new skill service
func NewSkillService(repo *repository.SkillRepository) *SkillService {
	return &SkillService{repo: repo}
}

// CreateSkill creates a new skill
func (s *SkillService) CreateSkill(ctx context.Context, userID string, req *models.CreateSkillRequest) (*models.Skill, error) {
	skill := &models.Skill{
		UserID:      &userID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Icon:        req.Icon,
		Config:      req.Config,
	}

	if err := s.repo.Create(ctx, skill); err != nil {
		return nil, err
	}

	return skill, nil
}

// GetSkill retrieves a skill by ID
func (s *SkillService) GetSkill(ctx context.Context, id string) (*models.Skill, error) {
	skill, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrSkillNotFound
		}
		return nil, err
	}
	return skill, nil
}

// GetUserSkills retrieves skills for a user with optional category filter
func (s *SkillService) GetUserSkills(ctx context.Context, userID string, category string, page, pageSize int) (*models.ListResponse, error) {
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
	skills, total, err := s.repo.GetByUser(ctx, userID, category, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.ListResponse{
		Items:      skills,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateSkill updates a skill
func (s *SkillService) UpdateSkill(ctx context.Context, id string, req *models.UpdateSkillRequest) (*models.Skill, error) {
	skill, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrSkillNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		skill.Name = *req.Name
	}
	if req.Description != nil {
		skill.Description = *req.Description
	}
	if req.Icon != nil {
		skill.Icon = *req.Icon
	}
	if req.Enabled != nil {
		skill.Enabled = *req.Enabled
	}
	if req.Config != nil {
		skill.Config = req.Config
	}

	if err := s.repo.Update(ctx, skill); err != nil {
		return nil, err
	}

	return skill, nil
}

// DeleteSkill deletes a skill
func (s *SkillService) DeleteSkill(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ToggleSkill toggles the enabled status of a skill
func (s *SkillService) ToggleSkill(ctx context.Context, id string, enabled bool) error {
	return s.repo.ToggleEnabled(ctx, id, enabled)
}
