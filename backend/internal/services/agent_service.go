package services

import (
	"context"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
	"github.com/atom-yt/atom-ai-platform/backend/internal/repository"
)

// AgentService handles agent business logic
type AgentService struct {
	repo *repository.AgentRepository
}

// NewAgentService creates a new agent service
func NewAgentService(repo *repository.AgentRepository) *AgentService {
	return &AgentService{repo: repo}
}

// CreateAgent creates a new agent
func (s *AgentService) CreateAgent(ctx context.Context, userID string, req *models.CreateAgentRequest) (*models.Agent, error) {
	agent := &models.Agent{
		UserID:       userID,
		Name:         req.Name,
		Description:  req.Description,
		SystemPrompt: req.SystemPrompt,
		Model:        req.Model,
		Provider:     req.Provider,
		Tools:        req.Tools,
		KnowledgeIDs: req.KnowledgeIDs,
	}

	if req.Temperature != nil {
		agent.Temperature = *req.Temperature
	}
	if req.MaxTokens != nil {
		agent.MaxTokens = *req.MaxTokens
	}

	if err := s.repo.Create(ctx, agent); err != nil {
		return nil, err
	}

	return agent, nil
}

// GetAgent retrieves an agent by ID
func (s *AgentService) GetAgent(ctx context.Context, id string) (*models.Agent, error) {
	agent, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrAgentNotFound
		}
		return nil, err
	}
	return agent, nil
}

// GetUserAgents retrieves agents for a user
func (s *AgentService) GetUserAgents(ctx context.Context, userID string, page, pageSize int) (*models.ListResponse, error) {
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
	agents, total, err := s.repo.GetByUser(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.ListResponse{
		Items:      agents,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ListAgents lists all agents with pagination
func (s *AgentService) ListAgents(ctx context.Context, page, pageSize int) (*models.ListResponse, error) {
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
	agents, total, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.ListResponse{
		Items:      agents,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateAgent updates an agent
func (s *AgentService) UpdateAgent(ctx context.Context, id string, req *models.UpdateAgentRequest) (*models.Agent, error) {
	agent, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrAgentNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		agent.Name = *req.Name
	}
	if req.Description != nil {
		agent.Description = *req.Description
	}
	if req.SystemPrompt != nil {
		agent.SystemPrompt = *req.SystemPrompt
	}
	if req.Model != nil {
		agent.Model = *req.Model
	}
	if req.Provider != nil {
		agent.Provider = *req.Provider
	}
	if req.Temperature != nil {
		agent.Temperature = *req.Temperature
	}
	if req.MaxTokens != nil {
		agent.MaxTokens = *req.MaxTokens
	}
	if req.Tools != nil {
		agent.Tools = *req.Tools
	}
	if req.KnowledgeIDs != nil {
		agent.KnowledgeIDs = *req.KnowledgeIDs
	}

	if err := s.repo.Update(ctx, agent); err != nil {
		return nil, err
	}

	return agent, nil
}

// DeleteAgent deletes an agent
func (s *AgentService) DeleteAgent(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// GetDefaultAgent retrieves the default agent for a user
func (s *AgentService) GetDefaultAgent(ctx context.Context, userID string) (*models.Agent, error) {
	agent, err := s.repo.GetDefaultAgent(ctx, userID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, ErrAgentNotFound
		}
		return nil, err
	}
	return agent, nil
}