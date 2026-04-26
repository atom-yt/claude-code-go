package repository

import (
	"context"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
)

// SessionRepositoryI defines the interface for session data operations
type SessionRepositoryI interface {
	Create(ctx context.Context, session *models.Session) error
	GetByID(ctx context.Context, id string) (*models.Session, error)
	GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Session, int64, error)
	Update(ctx context.Context, session *models.Session) error
	Delete(ctx context.Context, id string) error
	Archive(ctx context.Context, id string) error
	GetActiveSessions(ctx context.Context, userID string) ([]*models.Session, error)
}

// MessageRepositoryI defines the interface for message data operations
type MessageRepositoryI interface {
	Create(ctx context.Context, message *models.Message) error
	GetByID(ctx context.Context, id string) (*models.Message, error)
	GetBySession(ctx context.Context, sessionID string, limit, offset int) ([]*models.Message, int64, error)
	GetRecentBySession(ctx context.Context, sessionID string, limit int) ([]*models.Message, error)
	Delete(ctx context.Context, id string) error
	DeleteBySession(ctx context.Context, sessionID string) error
	Count(ctx context.Context, sessionID string) (int64, error)
}

// AgentRepositoryI defines the interface for agent data operations
type AgentRepositoryI interface {
	Create(ctx context.Context, agent *models.Agent) error
	GetByID(ctx context.Context, id string) (*models.Agent, error)
	GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Agent, int64, error)
	Update(ctx context.Context, agent *models.Agent) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.Agent, int64, error)
	GetDefaultAgent(ctx context.Context, userID string) (*models.Agent, error)
}

// SkillRepositoryI defines the interface for skill data operations
type SkillRepositoryI interface {
	Create(ctx context.Context, skill *models.Skill) error
	GetByID(ctx context.Context, id string) (*models.Skill, error)
	GetByUser(ctx context.Context, userID string, category string, limit, offset int) ([]*models.Skill, int64, error)
	Update(ctx context.Context, skill *models.Skill) error
	Delete(ctx context.Context, id string) error
	ToggleEnabled(ctx context.Context, id string, enabled bool) error
}

// ArtifactRepositoryI defines the interface for artifact data operations
type ArtifactRepositoryI interface {
	Create(ctx context.Context, artifact *models.Artifact) error
	GetByID(ctx context.Context, id string) (*models.Artifact, error)
	GetByUser(ctx context.Context, userID string, search string, limit, offset int) ([]*models.Artifact, int64, error)
	Delete(ctx context.Context, id string) error
	GetStats(ctx context.Context, userID string) (int64, error)
}

// ScheduleRepositoryI defines the interface for scheduled task data operations
type ScheduleRepositoryI interface {
	Create(ctx context.Context, task *models.ScheduledTask) error
	GetByID(ctx context.Context, id string) (*models.ScheduledTask, error)
	GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.ScheduledTask, int64, error)
	Update(ctx context.Context, task *models.ScheduledTask) error
	Delete(ctx context.Context, id string) error
	ToggleEnabled(ctx context.Context, id string, enabled bool) error
}

// KnowledgeRepositoryI defines the interface for knowledge base data operations
type KnowledgeRepositoryI interface {
	Create(ctx context.Context, kb *models.KnowledgeBase) error
	GetByID(ctx context.Context, id string) (*models.KnowledgeBase, error)
	GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.KnowledgeBase, int64, error)
	Update(ctx context.Context, kb *models.KnowledgeBase) error
	Delete(ctx context.Context, id string) error
}