package services

import (
	"context"
	"errors"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
)

var (
	ErrMessageNotFound = errors.New("message not found")
)

// MessageService handles message business logic
type MessageService struct {
	messageRepo  repository.MessageRepositoryI
	sessionRepo  repository.SessionRepositoryI
}

// NewMessageService creates a new message service
func NewMessageService(
	messageRepo repository.MessageRepositoryI,
	sessionRepo repository.SessionRepositoryI,
) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		sessionRepo: sessionRepo,
	}
}

// CreateMessage creates a new message
func (s *MessageService) CreateMessage(ctx context.Context, sessionID string, req *models.CreateMessageRequest) (*models.Message, error) {
	// Verify session exists
	_, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	message := &models.Message{
		SessionID: sessionID,
		Role:      req.Role,
		Content:   req.Content,
		ToolCalls: req.ToolCalls,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	return message, nil
}

// GetMessage retrieves a message by ID
func (s *MessageService) GetMessage(ctx context.Context, id string) (*models.Message, error) {
	message, err := s.messageRepo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrMessageNotFound
		}
		return nil, err
	}
	return message, nil
}

// GetSessionMessages retrieves messages for a session
func (s *MessageService) GetSessionMessages(ctx context.Context, sessionID string, page, pageSize int) (*models.ListResponse, error) {
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
	messages, total, err := s.messageRepo.GetBySession(ctx, sessionID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.ListResponse{
		Items:      messages,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetRecentMessages retrieves recent messages for a session
func (s *MessageService) GetRecentMessages(ctx context.Context, sessionID string, limit int) ([]*models.Message, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return s.messageRepo.GetRecentBySession(ctx, sessionID, limit)
}

// DeleteMessage deletes a message
func (s *MessageService) DeleteMessage(ctx context.Context, id string) error {
	return s.messageRepo.Delete(ctx, id)
}