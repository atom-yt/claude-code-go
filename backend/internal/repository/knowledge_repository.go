package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
)

// KnowledgeRepository handles knowledge base data operations
type KnowledgeRepository struct {
	db *pgxpool.Pool
}

// NewKnowledgeRepository creates a new knowledge repository
func NewKnowledgeRepository(db *pgxpool.Pool) *KnowledgeRepository {
	return &KnowledgeRepository{db: db}
}

// Create creates a new knowledge base
func (r *KnowledgeRepository) Create(ctx context.Context, kb *models.KnowledgeBase) error {
	query := `
		INSERT INTO knowledge_bases (id, user_id, name, description, type, source, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, name, description, type, source, created_at, updated_at
	`

	now := timeNow()
	if kb.ID == "" {
		kb.ID = uuid.New().String()
	}
	kb.CreatedAt = now
	kb.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		kb.ID,
		kb.UserID,
		kb.Name,
		kb.Description,
		kb.Type,
		kb.Source,
		kb.CreatedAt,
		kb.UpdatedAt,
	).Scan(
		&kb.ID,
		&kb.UserID,
		&kb.Name,
		&kb.Description,
		&kb.Type,
		&kb.Source,
		&kb.CreatedAt,
		&kb.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create knowledge base: %w", err)
	}

	return nil
}

// GetByID retrieves a knowledge base by ID
func (r *KnowledgeRepository) GetByID(ctx context.Context, id string) (*models.KnowledgeBase, error) {
	query := `
		SELECT id, user_id, name, description, type, source, created_at, updated_at
		FROM knowledge_bases
		WHERE id = $1
	`

	kb := &models.KnowledgeBase{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&kb.ID,
		&kb.UserID,
		&kb.Name,
		&kb.Description,
		&kb.Type,
		&kb.Source,
		&kb.CreatedAt,
		&kb.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge base: %w", err)
	}

	return kb, nil
}

// GetByUser retrieves knowledge bases for a user with pagination
func (r *KnowledgeRepository) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.KnowledgeBase, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM knowledge_bases WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count knowledge bases: %w", err)
	}

	// Get knowledge bases
	query := `
		SELECT id, user_id, name, description, type, source, created_at, updated_at
		FROM knowledge_bases
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list knowledge bases: %w", err)
	}
	defer rows.Close()

	kbs := []*models.KnowledgeBase{}
	for rows.Next() {
		kb := &models.KnowledgeBase{}
		err := rows.Scan(
			&kb.ID,
			&kb.UserID,
			&kb.Name,
			&kb.Description,
			&kb.Type,
			&kb.Source,
			&kb.CreatedAt,
			&kb.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan knowledge base: %w", err)
		}
		kbs = append(kbs, kb)
	}

	return kbs, total, nil
}

// Update updates a knowledge base
func (r *KnowledgeRepository) Update(ctx context.Context, kb *models.KnowledgeBase) error {
	query := `
		UPDATE knowledge_bases
		SET name = $2, description = $3, type = $4, source = $5, updated_at = $6
		WHERE id = $1
		RETURNING id, user_id, name, description, type, source, created_at, updated_at
	`

	kb.UpdatedAt = timeNow()

	err := r.db.QueryRow(ctx, query,
		kb.ID,
		kb.Name,
		kb.Description,
		kb.Type,
		kb.Source,
		kb.UpdatedAt,
	).Scan(
		&kb.ID,
		&kb.UserID,
		&kb.Name,
		&kb.Description,
		&kb.Type,
		&kb.Source,
		&kb.CreatedAt,
		&kb.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to update knowledge base: %w", err)
	}

	return nil
}

// Delete deletes a knowledge base
func (r *KnowledgeRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM knowledge_bases WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete knowledge base: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
