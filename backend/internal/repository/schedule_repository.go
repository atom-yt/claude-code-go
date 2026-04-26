package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/atom-yt/atom-ai-platform/backend/internal/models"
)

// ScheduleRepository handles scheduled task data operations
type ScheduleRepository struct {
	db *pgxpool.Pool
}

// NewScheduleRepository creates a new schedule repository
func NewScheduleRepository(db *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

// Create creates a new scheduled task
func (r *ScheduleRepository) Create(ctx context.Context, task *models.ScheduledTask) error {
	query := `
		INSERT INTO scheduled_tasks (id, user_id, title, prompt, schedule_type, schedule_time, model, enabled, notify_on_done, execution_count, last_run_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, user_id, title, prompt, schedule_type, schedule_time, model, enabled, notify_on_done, execution_count, last_run_at, created_at, updated_at
	`

	now := timeNow()
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	task.CreatedAt = now
	task.UpdatedAt = now

	// Set default values
	if task.ScheduleType == "" {
		task.ScheduleType = "daily"
	}
	if task.Model == "" {
		task.Model = "auto"
	}
	task.Enabled = true
	task.NotifyOnDone = true

	err := r.db.QueryRow(ctx, query,
		task.ID,
		task.UserID,
		task.Title,
		task.Prompt,
		task.ScheduleType,
		task.ScheduleTime,
		task.Model,
		task.Enabled,
		task.NotifyOnDone,
		task.ExecutionCount,
		task.LastRunAt,
		task.CreatedAt,
		task.UpdatedAt,
	).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Prompt,
		&task.ScheduleType,
		&task.ScheduleTime,
		&task.Model,
		&task.Enabled,
		&task.NotifyOnDone,
		&task.ExecutionCount,
		&task.LastRunAt,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create scheduled task: %w", err)
	}

	return nil
}

// GetByID retrieves a scheduled task by ID
func (r *ScheduleRepository) GetByID(ctx context.Context, id string) (*models.ScheduledTask, error) {
	query := `
		SELECT id, user_id, title, prompt, schedule_type, schedule_time, model, enabled, notify_on_done, execution_count, last_run_at, created_at, updated_at
		FROM scheduled_tasks
		WHERE id = $1
	`

	task := &models.ScheduledTask{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Prompt,
		&task.ScheduleType,
		&task.ScheduleTime,
		&task.Model,
		&task.Enabled,
		&task.NotifyOnDone,
		&task.ExecutionCount,
		&task.LastRunAt,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled task: %w", err)
	}

	return task, nil
}

// GetByUser retrieves scheduled tasks for a user with pagination
func (r *ScheduleRepository) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*models.ScheduledTask, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM scheduled_tasks WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count scheduled tasks: %w", err)
	}

	// Get scheduled tasks
	query := `
		SELECT id, user_id, title, prompt, schedule_type, schedule_time, model, enabled, notify_on_done, execution_count, last_run_at, created_at, updated_at
		FROM scheduled_tasks
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list scheduled tasks: %w", err)
	}
	defer rows.Close()

	tasks := []*models.ScheduledTask{}
	for rows.Next() {
		task := &models.ScheduledTask{}
		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Prompt,
			&task.ScheduleType,
			&task.ScheduleTime,
			&task.Model,
			&task.Enabled,
			&task.NotifyOnDone,
			&task.ExecutionCount,
			&task.LastRunAt,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan scheduled task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, total, nil
}

// Update updates a scheduled task
func (r *ScheduleRepository) Update(ctx context.Context, task *models.ScheduledTask) error {
	query := `
		UPDATE scheduled_tasks
		SET title = $2, prompt = $3, schedule_type = $4, schedule_time = $5, model = $6, enabled = $7, notify_on_done = $8, updated_at = $9
		WHERE id = $1
		RETURNING id, user_id, title, prompt, schedule_type, schedule_time, model, enabled, notify_on_done, execution_count, last_run_at, created_at, updated_at
	`

	task.UpdatedAt = timeNow()

	err := r.db.QueryRow(ctx, query,
		task.ID,
		task.Title,
		task.Prompt,
		task.ScheduleType,
		task.ScheduleTime,
		task.Model,
		task.Enabled,
		task.NotifyOnDone,
		task.UpdatedAt,
	).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Prompt,
		&task.ScheduleType,
		&task.ScheduleTime,
		&task.Model,
		&task.Enabled,
		&task.NotifyOnDone,
		&task.ExecutionCount,
		&task.LastRunAt,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to update scheduled task: %w", err)
	}

	return nil
}

// Delete deletes a scheduled task
func (r *ScheduleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM scheduled_tasks WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete scheduled task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ToggleEnabled toggles the enabled status of a scheduled task
func (r *ScheduleRepository) ToggleEnabled(ctx context.Context, id string, enabled bool) error {
	query := `UPDATE scheduled_tasks SET enabled = $2, updated_at = $3 WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id, enabled, timeNow())
	if err != nil {
		return fmt.Errorf("failed to toggle scheduled task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
