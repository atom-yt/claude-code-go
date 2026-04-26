package services

import (
	"context"
	"errors"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
)

var ErrKnowledgeNotFound = errors.New("knowledge not found")

// KnowledgeService handles knowledge base business logic
type KnowledgeService struct {
	repo repository.KnowledgeRepositoryI
}

// NewKnowledgeService creates a new knowledge service
func NewKnowledgeService(repo repository.KnowledgeRepositoryI) *KnowledgeService {
	return &KnowledgeService{repo: repo}
}

// CreateKnowledge creates a new knowledge base
func (s *KnowledgeService) CreateKnowledge(ctx context.Context, userID string, req *models.CreateKnowledgeBaseRequest) (*models.KnowledgeBase, error) {
	kb := &models.KnowledgeBase{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Source:      req.Source,
	}

	if err := s.repo.Create(ctx, kb); err != nil {
		return nil, err
	}

	return kb, nil
}

// GetKnowledge retrieves a knowledge base by ID
func (s *KnowledgeService) GetKnowledge(ctx context.Context, id string) (*models.KnowledgeBase, error) {
	kb, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrKnowledgeNotFound
		}
		return nil, err
	}
	return kb, nil
}

// GetUserKnowledge retrieves knowledge bases for a user
func (s *KnowledgeService) GetUserKnowledge(ctx context.Context, userID string, page, pageSize int) (*models.ListResponse, error) {
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
	kbs, total, err := s.repo.GetByUser(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.ListResponse{
		Items:      kbs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateKnowledge updates a knowledge base
func (s *KnowledgeService) UpdateKnowledge(ctx context.Context, id string, req *models.UpdateKnowledgeBaseRequest) (*models.KnowledgeBase, error) {
	kb, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrKnowledgeNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		kb.Name = *req.Name
	}
	if req.Description != nil {
		kb.Description = *req.Description
	}

	if err := s.repo.Update(ctx, kb); err != nil {
		return nil, err
	}

	return kb, nil
}

// DeleteKnowledge deletes a knowledge base
func (s *KnowledgeService) DeleteKnowledge(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
