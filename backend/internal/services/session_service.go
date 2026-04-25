package services

import (
	"context"
	"errors"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrAgentNotFound   = errors.New("agent not found")
)

// SessionService handles session business logic
type SessionService struct {
	sessionRepo *repository.SessionRepository
	agentRepo   *repository.AgentRepository
}

// NewSessionService creates a new session service
func NewSessionService(
	sessionRepo *repository.SessionRepository,
	agentRepo *repository.AgentRepository,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		agentRepo:   agentRepo,
	}
}

// CreateSession creates a new session
func (s *SessionService) CreateSession(ctx context.Context, userID string, req *models.CreateSessionRequest) (*models.Session, error) {
	// Verify agent exists
	_, err := s.agentRepo.GetByID(ctx, req.AgentID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrAgentNotFound
		}
		return nil, err
	}

	session := &models.Session{
		UserID:  userID,
		AgentID: req.AgentID,
		Title:   req.Title,
		Status:  "active",
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(ctx context.Context, id string) (*models.Session, error) {
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	return session, nil
}

// GetUserSessions retrieves sessions for a user
func (s *SessionService) GetUserSessions(ctx context.Context, userID string, page, pageSize int) (*models.ListResponse, error) {
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
	sessions, total, err := s.sessionRepo.GetByUser(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.ListResponse{
		Items:      sessions,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateSession updates a session
func (s *SessionService) UpdateSession(ctx context.Context, id string, req *models.UpdateSessionRequest) (*models.Session, error) {
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	if req.Title != nil {
		session.Title = *req.Title
	}
	if req.Status != nil {
		session.Status = *req.Status
	}

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// DeleteSession deletes a session
func (s *SessionService) DeleteSession(ctx context.Context, id string) error {
	return s.sessionRepo.Delete(ctx, id)
}

// ArchiveSession archives a session
func (s *SessionService) ArchiveSession(ctx context.Context, id string) error {
	return s.sessionRepo.Archive(ctx, id)
}

// GetActiveSessions retrieves active sessions for a user
func (s *SessionService) GetActiveSessions(ctx context.Context, userID string) ([]*models.Session, error) {
	return s.sessionRepo.GetActiveSessions(ctx, userID)
}