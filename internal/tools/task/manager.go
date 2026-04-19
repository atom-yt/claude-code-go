// Package task implements Task tools for managing background tasks and sub-agents.
// These tools enable the agent to manage a queue of background tasks, track their
// progress, and retrieve their results.
package task

import (
	"fmt"
	"sync"
	"time"
)

// Task status constants
const (
	StatusPending    = "pending"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusDeleted    = "deleted"
)

// TaskStatus represents the status of a task.
type TaskStatus string

// Task represents a single task in the task list.
type Task struct {
	ID          string                 `json:"id"`
	Subject     string                 `json:"subject"`
	Description string                 `json:"description"`
	Status      TaskStatus             `json:"status"`
	ActiveForm  string                 `json:"activeForm,omitempty"`
	Owner       string                 `json:"owner,omitempty"`
	Blocks      []string               `json:"blocks,omitempty"`    // Task IDs this task blocks
	BlockedBy   []string               `json:"blockedBy,omitempty"` // Task IDs blocking this task
	Metadata    map[string]interface{} `json:"metadata,omitempty"`  // Additional metadata
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// TaskManager manages all tasks.
type TaskManager struct {
	mu     sync.RWMutex
	tasks  map[string]*Task
	nextID int
}

// Global task manager instance
var globalManager = &TaskManager{
	tasks:  make(map[string]*Task),
	nextID: 1,
}

// GetManager returns the global task manager.
func GetManager() *TaskManager {
	return globalManager
}

// Create creates a new task.
func (tm *TaskManager) Create(subject, description, activeForm string) *Task {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	task := &Task{
		ID:          fmt.Sprintf("%d", tm.nextID),
		Subject:     subject,
		Description: description,
		Status:      StatusPending,
		ActiveForm:  activeForm,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	tm.tasks[task.ID] = task
	tm.nextID++
	return task
}

// Get retrieves a task by ID.
func (tm *TaskManager) Get(id string) (*Task, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	task, ok := tm.tasks[id]
	return task, ok
}

// List returns all tasks.
func (tm *TaskManager) List() []*Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tasks := make([]*Task, 0, len(tm.tasks))
	for _, t := range tm.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// Update updates a task's status and optionally other fields.
func (tm *TaskManager) Update(id string, updates func(*Task)) (*Task, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	task, ok := tm.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	updates(task)
	task.UpdatedAt = time.Now()
	return task, nil
}

// Delete marks a task as deleted.
func (tm *TaskManager) Delete(id string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, ok := tm.tasks[id]; !ok {
		return fmt.Errorf("task not found: %s", id)
	}
	delete(tm.tasks, id)
	return nil
}

// AddBlock adds a task ID to the blocks list.
// This also adds the taskID to the blockedBy list of the blocksID task.
func (tm *TaskManager) AddBlock(taskID, blocksID string) {
	_, _ = tm.Update(taskID, func(t *Task) {
		for _, b := range t.Blocks {
			if b == blocksID {
				return
			}
		}
		t.Blocks = append(t.Blocks, blocksID)
	})
	_, _ = tm.Update(blocksID, func(t *Task) {
		for _, b := range t.BlockedBy {
			if b == taskID {
				return
			}
		}
		t.BlockedBy = append(t.BlockedBy, taskID)
	})
}

// AddBlockedBy adds a task ID to the blockedBy list.
// This also adds the taskID to the blocks list of the blockedByID task.
func (tm *TaskManager) AddBlockedBy(taskID, blockedByID string) {
	_, _ = tm.Update(taskID, func(t *Task) {
		for _, b := range t.BlockedBy {
			if b == blockedByID {
				return
			}
		}
		t.BlockedBy = append(t.BlockedBy, blockedByID)
	})
	_, _ = tm.Update(blockedByID, func(t *Task) {
		for _, b := range t.Blocks {
			if b == taskID {
				return
			}
		}
		t.Blocks = append(t.Blocks, taskID)
	})
}
