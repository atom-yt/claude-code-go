package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
)

// AgentRepository handles agent data operations
type AgentRepository struct {
	db *pgxpool.Pool
}

// NewAgentRepository creates a new agent repository
func NewAgentRepository(db *pgxpool.Pool) *AgentRepository {
	return &AgentRepository{db: db}
}

// Create creates a new agent
func (r *AgentRepository) Create(ctx context.Context, agent *models.Agent) error {
	query := `
		INSERT INTO agents (id, user_id, name, description, system_prompt, model, provider, temperature, max_tokens, tools, knowledge_ids, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, user_id, name, description, system_prompt, model, provider, temperature, max_tokens, tools, knowledge_ids, created_at, updated_at
	`

	now := timeNow()
	if agent.ID == "" {
		agent.ID = uuid.New().String()
	}
	agent.CreatedAt = now
	agent.UpdatedAt = now

	// Set default values
	if agent.Temperature == 0 {
		agent.Temperature = 0.7
	}
	if agent.MaxTokens == 0 {
		agent.MaxTokens = 4096
	}

	err := r.db.QueryRow(ctx, query,
		agent.ID,
		agent.UserID,
		agent.Name,
		agent.Description,
		agent.SystemPrompt,
		agent.Model,
		agent.Provider,
		agent.Temperature,
		agent.MaxTokens,
		agent.Tools,
		agent.KnowledgeIDs,
		agent.CreatedAt,
		agent.UpdatedAt,
	).Scan(
		&agent.ID,
		&agent.UserID,
		&agent.Name,
		&agent.Description,
		&agent.SystemPrompt,
		&agent.Model,
		&agent.Provider,
		&agent.Temperature,
		&agent.MaxTokens,
		&agent.Tools,
		&agent.KnowledgeIDs,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	return nil
}

// GetByID retrieves an agent by ID
func (r *AgentRepository) GetByID(ctx context.Context, id string) (*models.Agent, error) {
	query := `
		SELECT id, user_id, name, description, system_prompt, model, provider, temperature, max_tokens, tools, knowledge_ids, created_at, updated_at
		FROM agents
		WHERE id = $1
	`

	agent := &models.Agent{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&agent.ID,
		&agent.UserID,
		&agent.Name,
		&agent.Description,
		&agent.SystemPrompt,
		&agent.Model,
		&agent.Provider,
		&agent.Temperature,
		&agent.MaxTokens,
		&agent.Tools,
		&agent.KnowledgeIDs,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return agent, nil
}

// GetByUser retrieves agents for a user with pagination
func (r *AgentRepository) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Agent, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM agents WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count agents: %w", err)
	}

	// Get agents
	query := `
		SELECT id, user_id, name, description, system_prompt, model, provider, temperature, max_tokens, tools, knowledge_ids, created_at, updated_at
		FROM agents
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	agents := []*models.Agent{}
	for rows.Next() {
		agent := &models.Agent{}
		err := rows.Scan(
			&agent.ID,
			&agent.UserID,
			&agent.Name,
			&agent.Description,
			&agent.SystemPrompt,
			&agent.Model,
			&agent.Provider,
			&agent.Temperature,
			&agent.MaxTokens,
			&agent.Tools,
			&agent.KnowledgeIDs,
			&agent.CreatedAt,
			&agent.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan agent: %w", err)
		}
		agents = append(agents, agent)
	}

	return agents, total, nil
}

// Update updates an agent
func (r *AgentRepository) Update(ctx context.Context, agent *models.Agent) error {
	query := `
		UPDATE agents
		SET name = $2, description = $3, system_prompt = $4, model = $5, provider = $6, temperature = $7, max_tokens = $8, tools = $9, knowledge_ids = $10, updated_at = $11
		WHERE id = $1
		RETURNING id, user_id, name, description, system_prompt, model, provider, temperature, max_tokens, tools, knowledge_ids, created_at, updated_at
	`

	agent.UpdatedAt = timeNow()

	err := r.db.QueryRow(ctx, query,
		agent.ID,
		agent.Name,
		agent.Description,
		agent.SystemPrompt,
		agent.Model,
		agent.Provider,
		agent.Temperature,
		agent.MaxTokens,
		agent.Tools,
		agent.KnowledgeIDs,
		agent.UpdatedAt,
	).Scan(
		&agent.ID,
		&agent.UserID,
		&agent.Name,
		&agent.Description,
		&agent.SystemPrompt,
		&agent.Model,
		&agent.Provider,
		&agent.Temperature,
		&agent.MaxTokens,
		&agent.Tools,
		&agent.KnowledgeIDs,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to update agent: %w", err)
	}

	return nil
}

// Delete deletes an agent
func (r *AgentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM agents WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// List retrieves all agents with pagination
func (r *AgentRepository) List(ctx context.Context, limit, offset int) ([]*models.Agent, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM agents`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count agents: %w", err)
	}

	// Get agents
	query := `
		SELECT id, user_id, name, description, system_prompt, model, provider, temperature, max_tokens, tools, knowledge_ids, created_at, updated_at
		FROM agents
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	agents := []*models.Agent{}
	for rows.Next() {
		agent := &models.Agent{}
		err := rows.Scan(
			&agent.ID,
			&agent.UserID,
			&agent.Name,
			&agent.Description,
			&agent.SystemPrompt,
			&agent.Model,
			&agent.Provider,
			&agent.Temperature,
			&agent.MaxTokens,
			&agent.Tools,
			&agent.KnowledgeIDs,
			&agent.CreatedAt,
			&agent.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan agent: %w", err)
		}
		agents = append(agents, agent)
	}

	return agents, total, nil
}

// GetDefaultAgent returns the default agent for a user
func (r *AgentRepository) GetDefaultAgent(ctx context.Context, userID string) (*models.Agent, error) {
	// For now, return the first agent created by the user
	// In the future, we might have a "is_default" flag
	query := `
		SELECT id, user_id, name, description, system_prompt, model, provider, temperature, max_tokens, tools, knowledge_ids, created_at, updated_at
		FROM agents
		WHERE user_id = $1
		ORDER BY created_at ASC
		LIMIT 1
	`

	agent := &models.Agent{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&agent.ID,
		&agent.UserID,
		&agent.Name,
		&agent.Description,
		&agent.SystemPrompt,
		&agent.Model,
		&agent.Provider,
		&agent.Temperature,
		&agent.MaxTokens,
		&agent.Tools,
		&agent.KnowledgeIDs,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get default agent: %w", err)
	}

	return agent, nil
}