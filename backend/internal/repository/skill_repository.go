package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
)

// SkillRepository handles skill data operations
type SkillRepository struct {
	db *pgxpool.Pool
}

// NewSkillRepository creates a new skill repository
func NewSkillRepository(db *pgxpool.Pool) *SkillRepository {
	return &SkillRepository{db: db}
}

// Create creates a new skill
func (r *SkillRepository) Create(ctx context.Context, skill *models.Skill) error {
	query := `
		INSERT INTO skills (id, user_id, team_id, name, description, category, icon, enabled, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, user_id, team_id, name, description, category, icon, enabled, config, created_at, updated_at
	`

	now := timeNow()
	if skill.ID == "" {
		skill.ID = uuid.New().String()
	}
	skill.CreatedAt = now
	skill.UpdatedAt = now

	// Set default values
	if skill.Category == "" {
		skill.Category = "personal"
	}
	if !skill.Enabled {
		skill.Enabled = true
	}

	err := r.db.QueryRow(ctx, query,
		skill.ID,
		skill.UserID,
		skill.TeamID,
		skill.Name,
		skill.Description,
		skill.Category,
		skill.Icon,
		skill.Enabled,
		skill.Config,
		skill.CreatedAt,
		skill.UpdatedAt,
	).Scan(
		&skill.ID,
		&skill.UserID,
		&skill.TeamID,
		&skill.Name,
		&skill.Description,
		&skill.Category,
		&skill.Icon,
		&skill.Enabled,
		&skill.Config,
		&skill.CreatedAt,
		&skill.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create skill: %w", err)
	}

	return nil
}

// GetByID retrieves a skill by ID
func (r *SkillRepository) GetByID(ctx context.Context, id string) (*models.Skill, error) {
	query := `
		SELECT id, user_id, team_id, name, description, category, icon, enabled, config, created_at, updated_at
		FROM skills
		WHERE id = $1
	`

	skill := &models.Skill{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&skill.ID,
		&skill.UserID,
		&skill.TeamID,
		&skill.Name,
		&skill.Description,
		&skill.Category,
		&skill.Icon,
		&skill.Enabled,
		&skill.Config,
		&skill.CreatedAt,
		&skill.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get skill: %w", err)
	}

	return skill, nil
}

// GetByUser retrieves skills for a user with optional category filter and pagination
func (r *SkillRepository) GetByUser(ctx context.Context, userID string, category string, limit, offset int) ([]*models.Skill, int64, error) {
	var total int64

	if category != "" {
		// Get total count with category filter
		countQuery := `SELECT COUNT(*) FROM skills WHERE user_id = $1 AND category = $2`
		if err := r.db.QueryRow(ctx, countQuery, userID, category).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("failed to count skills: %w", err)
		}

		// Get skills with category filter
		query := `
			SELECT id, user_id, team_id, name, description, category, icon, enabled, config, created_at, updated_at
			FROM skills
			WHERE user_id = $1 AND category = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4
		`

		rows, err := r.db.Query(ctx, query, userID, category, limit, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to list skills: %w", err)
		}
		defer rows.Close()

		return scanSkills(rows, total)
	}

	// Get total count without category filter
	countQuery := `SELECT COUNT(*) FROM skills WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count skills: %w", err)
	}

	// Get skills without category filter
	query := `
		SELECT id, user_id, team_id, name, description, category, icon, enabled, config, created_at, updated_at
		FROM skills
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list skills: %w", err)
	}
	defer rows.Close()

	return scanSkills(rows, total)
}

// Update updates a skill
func (r *SkillRepository) Update(ctx context.Context, skill *models.Skill) error {
	query := `
		UPDATE skills
		SET name = $2, description = $3, category = $4, icon = $5, enabled = $6, config = $7, updated_at = $8
		WHERE id = $1
		RETURNING id, user_id, team_id, name, description, category, icon, enabled, config, created_at, updated_at
	`

	skill.UpdatedAt = timeNow()

	err := r.db.QueryRow(ctx, query,
		skill.ID,
		skill.Name,
		skill.Description,
		skill.Category,
		skill.Icon,
		skill.Enabled,
		skill.Config,
		skill.UpdatedAt,
	).Scan(
		&skill.ID,
		&skill.UserID,
		&skill.TeamID,
		&skill.Name,
		&skill.Description,
		&skill.Category,
		&skill.Icon,
		&skill.Enabled,
		&skill.Config,
		&skill.CreatedAt,
		&skill.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to update skill: %w", err)
	}

	return nil
}

// Delete deletes a skill
func (r *SkillRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM skills WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete skill: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ToggleEnabled updates the enabled status of a skill
func (r *SkillRepository) ToggleEnabled(ctx context.Context, id string, enabled bool) error {
	query := `UPDATE skills SET enabled = $2, updated_at = $3 WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id, enabled, timeNow())
	if err != nil {
		return fmt.Errorf("failed to toggle skill: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// scanSkills scans skill rows into a slice
func scanSkills(rows pgx.Rows, total int64) ([]*models.Skill, int64, error) {
	skills := []*models.Skill{}
	for rows.Next() {
		skill := &models.Skill{}
		err := rows.Scan(
			&skill.ID,
			&skill.UserID,
			&skill.TeamID,
			&skill.Name,
			&skill.Description,
			&skill.Category,
			&skill.Icon,
			&skill.Enabled,
			&skill.Config,
			&skill.CreatedAt,
			&skill.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan skill: %w", err)
		}
		skills = append(skills, skill)
	}

	return skills, total, nil
}
