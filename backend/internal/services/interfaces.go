package services

import (
	"context"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
)

// SessionServiceInterface defines the interface for session service
type SessionServiceInterface interface {
	CreateSession(ctx context.Context, userID string, req *models.CreateSessionRequest) (*models.Session, error)
	GetSession(ctx context.Context, id string) (*models.Session, error)
	GetUserSessions(ctx context.Context, userID string, page, pageSize int) (*models.ListResponse, error)
	UpdateSession(ctx context.Context, id string, req *models.UpdateSessionRequest) (*models.Session, error)
	DeleteSession(ctx context.Context, id string) error
	ArchiveSession(ctx context.Context, id string) error
	GetActiveSessions(ctx context.Context, userID string) ([]*models.Session, error)
}

// MessageServiceInterface defines the interface for message service
type MessageServiceInterface interface {
	CreateMessage(ctx context.Context, sessionID string, req *models.CreateMessageRequest) (*models.Message, error)
	GetMessage(ctx context.Context, id string) (*models.Message, error)
	GetSessionMessages(ctx context.Context, sessionID string, page, pageSize int) (*models.ListResponse, error)
	GetRecentMessages(ctx context.Context, sessionID string, limit int) ([]*models.Message, error)
	DeleteMessage(ctx context.Context, id string) error
}

// AgentServiceInterface defines the interface for agent service
type AgentServiceInterface interface {
	CreateAgent(ctx context.Context, userID string, req *models.CreateAgentRequest) (*models.Agent, error)
	GetAgent(ctx context.Context, id string) (*models.Agent, error)
	GetUserAgents(ctx context.Context, userID string, page, pageSize int) (*models.ListResponse, error)
	ListAgents(ctx context.Context, page, pageSize int) (*models.ListResponse, error)
	UpdateAgent(ctx context.Context, id string, req *models.UpdateAgentRequest) (*models.Agent, error)
	DeleteAgent(ctx context.Context, id string) error
	GetDefaultAgent(ctx context.Context, userID string) (*models.Agent, error)
}

// SkillServiceInterface defines the interface for skill service
type SkillServiceInterface interface {
	CreateSkill(ctx context.Context, userID string, req *models.CreateSkillRequest) (*models.Skill, error)
	GetSkill(ctx context.Context, id string) (*models.Skill, error)
	GetUserSkills(ctx context.Context, userID string, category string, page, pageSize int) (*models.ListResponse, error)
	UpdateSkill(ctx context.Context, id string, req *models.UpdateSkillRequest) (*models.Skill, error)
	DeleteSkill(ctx context.Context, id string) error
	ToggleSkill(ctx context.Context, id string, enabled bool) error
}

// ArtifactServiceInterface defines the interface for artifact service
type ArtifactServiceInterface interface {
	CreateArtifact(ctx context.Context, userID string, req *models.CreateArtifactRequest) (*models.Artifact, error)
	GetArtifact(ctx context.Context, id string) (*models.Artifact, error)
	GetUserArtifacts(ctx context.Context, userID string, search string, page, pageSize int) (*models.ListResponse, error)
	DeleteArtifact(ctx context.Context, id string) error
	GetArtifactStats(ctx context.Context, userID string) (map[string]interface{}, error)
}

// ScheduleServiceInterface defines the interface for schedule service
type ScheduleServiceInterface interface {
	CreateSchedule(ctx context.Context, userID string, req *models.CreateScheduledTaskRequest) (*models.ScheduledTask, error)
	GetSchedule(ctx context.Context, id string) (*models.ScheduledTask, error)
	GetUserSchedules(ctx context.Context, userID string, page, pageSize int) (*models.ListResponse, error)
	UpdateSchedule(ctx context.Context, id string, req *models.UpdateScheduledTaskRequest) (*models.ScheduledTask, error)
	DeleteSchedule(ctx context.Context, id string) error
	ToggleSchedule(ctx context.Context, id string, enabled bool) error
}

// KnowledgeServiceInterface defines the interface for knowledge service
type KnowledgeServiceInterface interface {
	CreateKnowledge(ctx context.Context, userID string, req *models.CreateKnowledgeBaseRequest) (*models.KnowledgeBase, error)
	GetKnowledge(ctx context.Context, id string) (*models.KnowledgeBase, error)
	GetUserKnowledge(ctx context.Context, userID string, page, pageSize int) (*models.ListResponse, error)
	UpdateKnowledge(ctx context.Context, id string, req *models.UpdateKnowledgeBaseRequest) (*models.KnowledgeBase, error)
	DeleteKnowledge(ctx context.Context, id string) error
}