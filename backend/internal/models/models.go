package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	DisplayName  string    `json:"display_name" db:"display_name"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Session represents a chat session
type Session struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	AgentID   string    `json:"agent_id" db:"agent_id"`
	Title     string    `json:"title" db:"title"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Message represents a message in a session
type Message struct {
	ID        string                 `json:"id" db:"id"`
	SessionID string                 `json:"session_id" db:"session_id"`
	Role      string                 `json:"role" db:"role"`
	Content   map[string]interface{} `json:"content" db:"content"`
	ToolCalls map[string]interface{} `json:"tool_calls,omitempty" db:"tool_calls"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// Agent represents an AI agent configuration
type Agent struct {
	ID             string    `json:"id" db:"id"`
	UserID         string    `json:"user_id" db:"user_id"`
	Name           string    `json:"name" db:"name"`
	Description    string    `json:"description" db:"description"`
	SystemPrompt   string    `json:"system_prompt" db:"system_prompt"`
	Model          string    `json:"model" db:"model"`
	Provider       string    `json:"provider" db:"provider"`
	Temperature    float64   `json:"temperature" db:"temperature"`
	MaxTokens      int       `json:"max_tokens" db:"max_tokens"`
	Tools          []string  `json:"tools" db:"tools"`
	KnowledgeIDs   []string  `json:"knowledge_ids,omitempty" db:"knowledge_ids"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// KnowledgeBase represents a knowledge base
type KnowledgeBase struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"`
	Source      string    `json:"source" db:"source"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// KnowledgeDocument represents a document in a knowledge base
type KnowledgeDocument struct {
	ID             string    `json:"id" db:"id"`
	KnowledgeBaseID string   `json:"knowledge_base_id" db:"knowledge_base_id"`
	Title          string    `json:"title" db:"title"`
	Source         string    `json:"source" db:"source"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// KnowledgeChunk represents a chunk of text from a knowledge document
type KnowledgeChunk struct {
	ID             string                 `json:"id" db:"id"`
	DocumentID     string                 `json:"document_id" db:"document_id"`
	Content        string                 `json:"content" db:"content"`
	Embedding      []float64              `json:"embedding" db:"embedding"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
}

// CreateUserRequest is the request to create a user
type CreateUserRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
}

// UpdateUserRequest is the request to update a user
type UpdateUserRequest struct {
	DisplayName *string `json:"display_name" binding:"omitempty,min=2,max=100"`
}

// CreateSessionRequest is the request to create a session
type CreateSessionRequest struct {
	Title    string `json:"title" binding:"omitempty,max=255"`
	AgentID  string `json:"agent_id" binding:"required,uuid"`
}

// UpdateSessionRequest is the request to update a session
type UpdateSessionRequest struct {
	Title  *string `json:"title" binding:"omitempty,max=255"`
	Status *string `json:"status" binding:"omitempty,oneof=active archived deleted"`
}

// CreateMessageRequest is the request to create a message
type CreateMessageRequest struct {
	Role      string                 `json:"role" binding:"required,oneof=user assistant system"`
	Content   map[string]interface{} `json:"content" binding:"required"`
	ToolCalls map[string]interface{} `json:"tool_calls" binding:"omitempty"`
}

// CreateAgentRequest is the request to create an agent
type CreateAgentRequest struct {
	Name         string                 `json:"name" binding:"required,min=1,max=255"`
	Description  string                 `json:"description" binding:"omitempty,max=500"`
	SystemPrompt string                 `json:"system_prompt" binding:"required"`
	Model        string                 `json:"model" binding:"required,max=100"`
	Provider     string                 `json:"provider" binding:"required,max=50"`
	Temperature  *float64               `json:"temperature" binding:"omitempty,min=0,max=2"`
	MaxTokens    *int                   `json:"max_tokens" binding:"omitempty,min=1"`
	Tools        []string               `json:"tools" binding:"omitempty"`
	KnowledgeIDs []string               `json:"knowledge_ids" binding:"omitempty"`
	Config       map[string]interface{} `json:"config" binding:"omitempty"`
}

// UpdateAgentRequest is the request to update an agent
type UpdateAgentRequest struct {
	Name         *string                 `json:"name" binding:"omitempty,min=1,max=255"`
	Description  *string                 `json:"description" binding:"omitempty,max=500"`
	SystemPrompt *string                 `json:"system_prompt" binding:"omitempty"`
	Model        *string                 `json:"model" binding:"omitempty,max=100"`
	Provider     *string                 `json:"provider" binding:"omitempty,max=50"`
	Temperature  *float64               `json:"temperature" binding:"omitempty,min=0,max=2"`
	MaxTokens    *int                   `json:"max_tokens" binding:"omitempty,min=1"`
	Tools        *[]string              `json:"tools" binding:"omitempty"`
	KnowledgeIDs *[]string              `json:"knowledge_ids" binding:"omitempty"`
	Config       map[string]interface{} `json:"config" binding:"omitempty"`
}

// CreateKnowledgeBaseRequest is the request to create a knowledge base
type CreateKnowledgeBaseRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"omitempty,max=500"`
	Type        string `json:"type" binding:"required,max=50"`
	Source      string `json:"source" binding:"omitempty"`
}

// UpdateKnowledgeBaseRequest is the request to update a knowledge base
type UpdateKnowledgeBaseRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

// Skill represents a skill/tool that atom can use
type Skill struct {
	ID          string                 `json:"id" db:"id"`
	UserID      *string                `json:"user_id" db:"user_id"`
	TeamID      *string                `json:"team_id" db:"team_id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Category    string                 `json:"category" db:"category"`
	Icon        string                 `json:"icon" db:"icon"`
	Enabled     bool                   `json:"enabled" db:"enabled"`
	Config      map[string]interface{} `json:"config" db:"config"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// CreateSkillRequest is the request to create a skill
type CreateSkillRequest struct {
	Name        string                 `json:"name" binding:"required,min=1,max=255"`
	Description string                 `json:"description" binding:"omitempty,max=1000"`
	Category    string                 `json:"category" binding:"omitempty,oneof=personal team builtin"`
	Icon        string                 `json:"icon" binding:"omitempty,max=50"`
	Config      map[string]interface{} `json:"config" binding:"omitempty"`
}

// UpdateSkillRequest is the request to update a skill
type UpdateSkillRequest struct {
	Name        *string                 `json:"name" binding:"omitempty,min=1,max=255"`
	Description *string                 `json:"description" binding:"omitempty,max=1000"`
	Icon        *string                 `json:"icon" binding:"omitempty,max=50"`
	Enabled     *bool                   `json:"enabled" binding:"omitempty"`
	Config      map[string]interface{}  `json:"config" binding:"omitempty"`
}

// Artifact represents a generated artifact/deliverable
type Artifact struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	SessionID *string   `json:"session_id" db:"session_id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	FileType  string    `json:"file_type" db:"file_type"`
	Tags      []string  `json:"tags" db:"tags"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateArtifactRequest is the request to create an artifact
type CreateArtifactRequest struct {
	Title     string   `json:"title" binding:"required,min=1,max=500"`
	Content   string   `json:"content" binding:"required"`
	SessionID string   `json:"session_id" binding:"omitempty"`
	FileType  string   `json:"file_type" binding:"omitempty,max=20"`
	Tags      []string `json:"tags" binding:"omitempty"`
}

// ScheduledTask represents a scheduled AI task
type ScheduledTask struct {
	ID             string     `json:"id" db:"id"`
	UserID         string     `json:"user_id" db:"user_id"`
	Title          string     `json:"title" db:"title"`
	Prompt         string     `json:"prompt" db:"prompt"`
	ScheduleType   string     `json:"schedule_type" db:"schedule_type"`
	ScheduleTime   string     `json:"schedule_time" db:"schedule_time"`
	Model          string     `json:"model" db:"model"`
	Enabled        bool       `json:"enabled" db:"enabled"`
	NotifyOnDone   bool       `json:"notify_on_done" db:"notify_on_done"`
	ExecutionCount int        `json:"execution_count" db:"execution_count"`
	LastRunAt      *time.Time `json:"last_run_at" db:"last_run_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateScheduledTaskRequest is the request to create a scheduled task
type CreateScheduledTaskRequest struct {
	Title        string `json:"title" binding:"required,min=1,max=255"`
	Prompt       string `json:"prompt" binding:"required,max=5000"`
	ScheduleType string `json:"schedule_type" binding:"omitempty,oneof=daily weekly cron"`
	ScheduleTime string `json:"schedule_time" binding:"required,max=50"`
	Model        string `json:"model" binding:"omitempty,max=100"`
	NotifyOnDone *bool  `json:"notify_on_done" binding:"omitempty"`
}

// UpdateScheduledTaskRequest is the request to update a scheduled task
type UpdateScheduledTaskRequest struct {
	Title        *string `json:"title" binding:"omitempty,min=1,max=255"`
	Prompt       *string `json:"prompt" binding:"omitempty,max=5000"`
	ScheduleType *string `json:"schedule_type" binding:"omitempty,oneof=daily weekly cron"`
	ScheduleTime *string `json:"schedule_time" binding:"omitempty,max=50"`
	Model        *string `json:"model" binding:"omitempty,max=100"`
	Enabled      *bool   `json:"enabled" binding:"omitempty"`
	NotifyOnDone *bool   `json:"notify_on_done" binding:"omitempty"`
}

// ListResponse is a generic list response with pagination
type ListResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}