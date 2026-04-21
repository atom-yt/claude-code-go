// Package taskstore provides durable storage for tasks with file-based persistence.
// Tasks are stored in .claude/tasks.json and support CRUD operations with session association.
package taskstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TaskStatus represents the current status of a task.
type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusCompleted  TaskStatus = "completed"
	StatusBlocked    TaskStatus = "blocked"
	StatusDeleted   TaskStatus = "deleted"
)

// Task represents a single task with metadata.
type Task struct {
	ID          string      `json:"id"`
	Subject     string      `json:"subject"`
	Description string      `json:"description"`
	Status      TaskStatus  `json:"status"`
	ActiveForm  string      `json:"activeForm,omitempty"`
	SessionID   string      `json:"sessionId,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
	CompletedAt *time.Time  `json:"completedAt,omitempty"`

	// Task dependencies
	Blocks     []string `json:"blocks,omitempty"`     // Task IDs blocked by this task
	BlockedBy   []string `json:"blockedBy,omitempty"`  // Task IDs this task depends on

	// Custom metadata
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Store manages durable task storage.
type Store struct {
	mu     sync.RWMutex
	file   string
	tasks  map[string]*Task
	nextID int
}

// New creates a new task store with the given workspace root.
// The store will load existing tasks from .claude/tasks.json if it exists.
func New(workspaceRoot string) (*Store, error) {
	s := &Store{
		file:  filepath.Join(workspaceRoot, ".claude", "tasks.json"),
		tasks: make(map[string]*Task),
		nextID: 1,
	}

	if err := s.load(); err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	return s, nil
}

// load loads tasks from the storage file.
func (s *Store) load() error {
	data, err := os.ReadFile(s.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No existing tasks file
		}
		return err
	}

	var fileData struct {
		Tasks  map[string]*Task `json:"tasks"`
		NextID int              `json:"nextId"`
	}

	if err := json.Unmarshal(data, &fileData); err != nil {
		return err
	}

	s.tasks = fileData.Tasks
	if fileData.NextID > 0 {
		s.nextID = fileData.NextID
	}

	return nil
}

// save persists tasks to the storage file.
func (s *Store) save() error {
	dir := filepath.Dir(s.file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	fileData := struct {
		Tasks  map[string]*Task `json:"tasks"`
		NextID int              `json:"nextId"`
	}{
		Tasks:  s.tasks,
		NextID: s.nextID,
	}

	data, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.file, data, 0644)
}

// Create creates a new task with the given parameters.
func (s *Store) Create(subject, description string) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	task := &Task{
		ID:          fmt.Sprintf("task-%d", s.nextID),
		Subject:     subject,
		Description: description,
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.tasks[task.ID] = task
	s.nextID++

	if err := s.save(); err != nil {
		return nil, err
	}

	return task, nil
}

// Get retrieves a task by ID.
func (s *Store) Get(id string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	if !ok || task.Status == StatusDeleted {
		return nil, false
	}
	return task, true
}

// List returns all non-deleted tasks, optionally filtered by status or session.
func (s *Store) List(status TaskStatus, sessionID string) []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Task
	for _, task := range s.tasks {
		if task.Status == StatusDeleted {
			continue
		}
		if status != "" && task.Status != status {
			continue
		}
		if sessionID != "" && task.SessionID != sessionID {
			continue
		}
		result = append(result, task)
	}
	return result
}

// Update updates a task's fields.
func (s *Store) Update(id string, updates map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return fmt.Errorf("task not found: %s", id)
	}

	now := time.Now()

	// Apply updates
	if subject, ok := updates["subject"].(string); ok {
		task.Subject = subject
	}
	if description, ok := updates["description"].(string); ok {
		task.Description = description
	}
	if status, ok := updates["status"].(TaskStatus); ok {
		task.Status = status
		if status == StatusCompleted {
			task.CompletedAt = &now
		}
	}
	if activeForm, ok := updates["activeForm"].(string); ok {
		task.ActiveForm = activeForm
	}
	if sessionID, ok := updates["sessionId"].(string); ok {
		task.SessionID = sessionID
	}

	task.UpdatedAt = now

	return s.save()
}

// Delete marks a task as deleted.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return fmt.Errorf("task not found: %s", id)
	}

	task.Status = StatusDeleted
	task.UpdatedAt = time.Now()

	return s.save()
}

// SetBlockedBy adds dependency relationships.
func (s *Store) SetBlockedBy(id string, blockedBy []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return fmt.Errorf("task not found: %s", id)
	}

	task.BlockedBy = blockedBy
	task.UpdatedAt = time.Now()

	return s.save()
}

// SetBlocks adds blocking relationships.
func (s *Store) SetBlocks(id string, blocks []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return fmt.Errorf("task not found: %s", id)
	}

	task.Blocks = blocks
	task.UpdatedAt = time.Now()

	return s.save()
}

// SetSessionID associates a task with a session.
func (s *Store) SetSessionID(id, sessionID string) error {
	return s.Update(id, map[string]any{"sessionId": sessionID})
}

// GetPendingTasks returns all pending tasks with no blockers.
func (s *Store) GetPendingTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Task
	for _, task := range s.tasks {
		if task.Status == StatusPending {
			// Check if task has any pending blockers
			hasPendingBlocker := false
			for _, blockerID := range task.BlockedBy {
				if blocker, ok := s.tasks[blockerID]; ok && blocker.Status != StatusCompleted {
					hasPendingBlocker = true
					break
				}
			}
			if !hasPendingBlocker {
				result = append(result, task)
			}
		}
	}
	return result
}

// GetBlockedTasks returns all pending tasks that have blockers.
func (s *Store) GetBlockedTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Task
	for _, task := range s.tasks {
		if task.Status == StatusPending && len(task.BlockedBy) > 0 {
			result = append(result, task)
		}
	}
	return result
}

// GetInProgressTasks returns all tasks currently in progress.
func (s *Store) GetInProgressTasks() []*Task {
	return s.List(StatusInProgress, "")
}

// GetCompletedTasks returns all completed tasks.
func (s *Store) GetCompletedTasks() []*Task {
	return s.List(StatusCompleted, "")
}

// SetMetadata updates task metadata.
func (s *Store) SetMetadata(id string, metadata map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return fmt.Errorf("task not found: %s", id)
	}

	if task.Metadata == nil {
		task.Metadata = make(map[string]any)
	}

	// Merge metadata
	for k, v := range metadata {
		if v == nil {
			delete(task.Metadata, k)
		} else {
			task.Metadata[k] = v
		}
	}

	task.UpdatedAt = time.Now()

	return s.save()
}
