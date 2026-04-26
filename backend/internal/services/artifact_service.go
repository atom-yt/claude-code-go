package services

import (
	"context"
	"errors"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
)

var ErrArtifactNotFound = errors.New("artifact not found")

// ArtifactService handles artifact business logic
type ArtifactService struct {
	repo *repository.ArtifactRepository
}

// NewArtifactService creates a new artifact service
func NewArtifactService(repo *repository.ArtifactRepository) *ArtifactService {
	return &ArtifactService{repo: repo}
}

// CreateArtifact creates a new artifact
func (s *ArtifactService) CreateArtifact(ctx context.Context, userID string, req *models.CreateArtifactRequest) (*models.Artifact, error) {
	artifact := &models.Artifact{
		UserID:   userID,
		Title:    req.Title,
		Content:  req.Content,
		FileType: req.FileType,
		Tags:     req.Tags,
	}

	if req.SessionID != "" {
		artifact.SessionID = &req.SessionID
	}

	if err := s.repo.Create(ctx, artifact); err != nil {
		return nil, err
	}

	return artifact, nil
}

// GetArtifact retrieves an artifact by ID
func (s *ArtifactService) GetArtifact(ctx context.Context, id string) (*models.Artifact, error) {
	artifact, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrArtifactNotFound
		}
		return nil, err
	}
	return artifact, nil
}

// GetUserArtifacts retrieves artifacts for a user with optional search and pagination
func (s *ArtifactService) GetUserArtifacts(ctx context.Context, userID string, search string, page, pageSize int) (*models.ListResponse, error) {
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
	artifacts, total, err := s.repo.GetByUser(ctx, userID, search, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.ListResponse{
		Items:      artifacts,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// DeleteArtifact deletes an artifact
func (s *ArtifactService) DeleteArtifact(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// GetArtifactStats returns artifact statistics for a user
func (s *ArtifactService) GetArtifactStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	total, err := s.repo.GetStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total": total,
	}, nil
}
