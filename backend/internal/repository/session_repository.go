package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
)

// SessionRepository handles session data operations
type SessionRepository struct {
	db *pgxpool.Pool
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session
func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO chat_sessions (id, user_id, agent_id, title, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, agent_id, title, status, created_at, updated_at
	`

	now := timeNow()
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	session.CreatedAt = now
	session.UpdatedAt = now
	if session.Status == "" {
		session.Status = "active"
	}

	err := r.db.QueryRow(ctx, query,
		session.ID,
		session.UserID,
		session.AgentID,
		session.Title,
		session.Status,
		session.CreatedAt,
		session.UpdatedAt,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.AgentID,
		&session.Title,
		&session.Status,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByID retrieves a session by ID
func (r *SessionRepository) GetByID(ctx context.Context, id string) (*models.Session, error) {
	query := `
		SELECT id, user_id, agent_id, title, status, created_at, updated_at
		FROM chat_sessions
		WHERE id = $1
	`

	session := &models.Session{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.AgentID,
		&session.Title,
		&session.Status,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// GetByUser retrieves sessions for a user with pagination
func (r *SessionRepository) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Session, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM chat_sessions WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	// Get sessions
	query := `
		SELECT id, user_id, agent_id, title, status, created_at, updated_at
		FROM chat_sessions
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	sessions := []*models.Session{}
	for rows.Next() {
		session := &models.Session{}
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.AgentID,
			&session.Title,
			&session.Status,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, total, nil
}

// Update updates a session
func (r *SessionRepository) Update(ctx context.Context, session *models.Session) error {
	query := `
		UPDATE chat_sessions
		SET title = $2, status = $3, updated_at = $4
		WHERE id = $1
		RETURNING id, user_id, agent_id, title, status, created_at, updated_at
	`

	session.UpdatedAt = timeNow()

	err := r.db.QueryRow(ctx, query,
		session.ID,
		session.Title,
		session.Status,
		session.UpdatedAt,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.AgentID,
		&session.Title,
		&session.Status,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// Delete deletes a session
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM chat_sessions WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// Archive archives a session
func (r *SessionRepository) Archive(ctx context.Context, id string) error {
	query := `
		UPDATE chat_sessions
		SET status = 'archived', updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, timeNow())
	if err != nil {
		return fmt.Errorf("failed to archive session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// GetActiveSessions retrieves active sessions for a user
func (r *SessionRepository) GetActiveSessions(ctx context.Context, userID string) ([]*models.Session, error) {
	query := `
		SELECT id, user_id, agent_id, title, status, created_at, updated_at
		FROM chat_sessions
		WHERE user_id = $1 AND status = 'active'
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	defer rows.Close()

	sessions := []*models.Session{}
	for rows.Next() {
		session := &models.Session{}
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.AgentID,
			&session.Title,
			&session.Status,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}