package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
)

// MessageRepository handles message data operations
type MessageRepository struct {
	db *pgxpool.Pool
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create creates a new message
func (r *MessageRepository) Create(ctx context.Context, message *models.Message) error {
	query := `
		INSERT INTO messages (id, session_id, role, content, tool_calls, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, session_id, role, content, tool_calls, created_at
	`

	now := timeNow()
	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	message.CreatedAt = now

	var contentBytes, toolCallsBytes []byte
	var err error

	contentBytes, err = jsonMarshal(message.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	if message.ToolCalls != nil {
		toolCallsBytes, err = jsonMarshal(message.ToolCalls)
		if err != nil {
			return fmt.Errorf("failed to marshal tool_calls: %w", err)
		}
	}

	err = r.db.QueryRow(ctx, query,
		message.ID,
		message.SessionID,
		message.Role,
		contentBytes,
		toolCallsBytes,
		message.CreatedAt,
	).Scan(
		&message.ID,
		&message.SessionID,
		&message.Role,
		&message.Content,
		&message.ToolCalls,
		&message.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// GetByID retrieves a message by ID
func (r *MessageRepository) GetByID(ctx context.Context, id string) (*models.Message, error) {
	query := `
		SELECT id, session_id, role, content, tool_calls, created_at
		FROM messages
		WHERE id = $1
	`

	message := &models.Message{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&message.ID,
		&message.SessionID,
		&message.Role,
		&message.Content,
		&message.ToolCalls,
		&message.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return message, nil
}

// GetBySession retrieves messages for a session with pagination
func (r *MessageRepository) GetBySession(ctx context.Context, sessionID string, limit, offset int) ([]*models.Message, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM messages WHERE session_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, sessionID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	// Get messages
	query := `
		SELECT id, session_id, role, content, tool_calls, created_at
		FROM messages
		WHERE session_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, sessionID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	messages := []*models.Message{}
	for rows.Next() {
		message := &models.Message{}
		err := rows.Scan(
			&message.ID,
			&message.SessionID,
			&message.Role,
			&message.Content,
			&message.ToolCalls,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, message)
	}

	return messages, total, nil
}

// GetRecentBySession retrieves recent messages for a session
func (r *MessageRepository) GetRecentBySession(ctx context.Context, sessionID string, limit int) ([]*models.Message, error) {
	query := `
		SELECT id, session_id, role, content, tool_calls, created_at
		FROM messages
		WHERE session_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent messages: %w", err)
	}
	defer rows.Close()

	messages := []*models.Message{}
	for rows.Next() {
		message := &models.Message{}
		err := rows.Scan(
			&message.ID,
			&message.SessionID,
			&message.Role,
			&message.Content,
			&message.ToolCalls,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, message)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// Delete deletes a message
func (r *MessageRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM messages WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteBySession deletes all messages for a session
func (r *MessageRepository) DeleteBySession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM messages WHERE session_id = $1`

	_, err := r.db.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}

	return nil
}

// Count returns the total message count for a session
func (r *MessageRepository) Count(ctx context.Context, sessionID string) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM messages WHERE session_id = $1`, sessionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}
	return count, nil
}

func jsonMarshal(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}