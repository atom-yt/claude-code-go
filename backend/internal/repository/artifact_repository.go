package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
)

// ArtifactRepository handles artifact data operations
type ArtifactRepository struct {
	db *pgxpool.Pool
}

// NewArtifactRepository creates a new artifact repository
func NewArtifactRepository(db *pgxpool.Pool) *ArtifactRepository {
	return &ArtifactRepository{db: db}
}

// Create creates a new artifact
func (r *ArtifactRepository) Create(ctx context.Context, artifact *models.Artifact) error {
	query := `
		INSERT INTO artifacts (id, user_id, session_id, title, content, file_type, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, session_id, title, content, file_type, tags, created_at, updated_at
	`

	now := timeNow()
	if artifact.ID == "" {
		artifact.ID = uuid.New().String()
	}
	artifact.CreatedAt = now
	artifact.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		artifact.ID,
		artifact.UserID,
		artifact.SessionID,
		artifact.Title,
		artifact.Content,
		artifact.FileType,
		artifact.Tags,
		artifact.CreatedAt,
		artifact.UpdatedAt,
	).Scan(
		&artifact.ID,
		&artifact.UserID,
		&artifact.SessionID,
		&artifact.Title,
		&artifact.Content,
		&artifact.FileType,
		&artifact.Tags,
		&artifact.CreatedAt,
		&artifact.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create artifact: %w", err)
	}

	return nil
}

// GetByID retrieves an artifact by ID
func (r *ArtifactRepository) GetByID(ctx context.Context, id string) (*models.Artifact, error) {
	query := `
		SELECT id, user_id, session_id, title, content, file_type, tags, created_at, updated_at
		FROM artifacts
		WHERE id = $1
	`

	artifact := &models.Artifact{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&artifact.ID,
		&artifact.UserID,
		&artifact.SessionID,
		&artifact.Title,
		&artifact.Content,
		&artifact.FileType,
		&artifact.Tags,
		&artifact.CreatedAt,
		&artifact.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get artifact: %w", err)
	}

	return artifact, nil
}

// GetByUser retrieves artifacts for a user with optional search and pagination
func (r *ArtifactRepository) GetByUser(ctx context.Context, userID string, search string, limit, offset int) ([]*models.Artifact, int64, error) {
	var total int64

	if search != "" {
		// Get total count with search filter
		countQuery := `SELECT COUNT(*) FROM artifacts WHERE user_id = $1 AND title ILIKE $2`
		searchPattern := "%" + search + "%"
		if err := r.db.QueryRow(ctx, countQuery, userID, searchPattern).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("failed to count artifacts: %w", err)
		}

		// Get artifacts with search filter
		query := `
			SELECT id, user_id, session_id, title, content, file_type, tags, created_at, updated_at
			FROM artifacts
			WHERE user_id = $1 AND title ILIKE $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4
		`

		rows, err := r.db.Query(ctx, query, userID, searchPattern, limit, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to list artifacts: %w", err)
		}
		defer rows.Close()

		artifacts := []*models.Artifact{}
		for rows.Next() {
			artifact := &models.Artifact{}
			err := rows.Scan(
				&artifact.ID,
				&artifact.UserID,
				&artifact.SessionID,
				&artifact.Title,
				&artifact.Content,
				&artifact.FileType,
				&artifact.Tags,
				&artifact.CreatedAt,
				&artifact.UpdatedAt,
			)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to scan artifact: %w", err)
			}
			artifacts = append(artifacts, artifact)
		}

		return artifacts, total, nil
	}

	// Get total count without search filter
	countQuery := `SELECT COUNT(*) FROM artifacts WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count artifacts: %w", err)
	}

	// Get artifacts without search filter
	query := `
		SELECT id, user_id, session_id, title, content, file_type, tags, created_at, updated_at
		FROM artifacts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list artifacts: %w", err)
	}
	defer rows.Close()

	artifacts := []*models.Artifact{}
	for rows.Next() {
		artifact := &models.Artifact{}
		err := rows.Scan(
			&artifact.ID,
			&artifact.UserID,
			&artifact.SessionID,
			&artifact.Title,
			&artifact.Content,
			&artifact.FileType,
			&artifact.Tags,
			&artifact.CreatedAt,
			&artifact.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan artifact: %w", err)
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts, total, nil
}

// Delete deletes an artifact
func (r *ArtifactRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM artifacts WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete artifact: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// GetStats returns the total count of artifacts for a user
func (r *ArtifactRepository) GetStats(ctx context.Context, userID string) (int64, error) {
	var total int64
	query := `SELECT COUNT(*) FROM artifacts WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, query, userID).Scan(&total); err != nil {
		return 0, fmt.Errorf("failed to get artifact stats: %w", err)
	}
	return total, nil
}
