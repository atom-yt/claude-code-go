package task

import (
	"fmt"

	"github.com/atom-yt/claude-code-go/internal/taskstore"
)

var (
	// Global task manager instance
	globalManager *Manager
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
	ID          string       `json:"id"`
	Subject     string       `json:"subject"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status"`
	ActiveForm  string       `json:"activeForm,omitempty"`
	Owner       string       `json:"owner,omitempty"`
	CreatedAt   string       `json:"createdAt"`
	UpdatedAt   string       `json:"updatedAt"`

	// Task dependencies
	Blocks     []string `json:"blocks,omitempty"`    // Task IDs blocked by this task
	BlockedBy   []string `json:"blockedBy,omitempty"` // Task IDs this task depends on

	// Custom metadata
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Manager provides a task management interface backed by a durable store.
type Manager struct {
	store *taskstore.Store
}

// GetManager returns the global task manager instance.
func GetManager() *Manager {
	if globalManager == nil {
		globalManager = &Manager{}
	}
	return globalManager
}

// Initialize initializes the task manager with a workspace root.
// This should be called during TUI initialization.
func Initialize(workspaceRoot string) error {
	if globalManager == nil {
		globalManager = &Manager{}
	}
	store, err := taskstore.New(workspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to initialize task store: %w", err)
	}
	globalManager.store = store
	return nil
}

// Create creates a new task.
func (m *Manager) Create(subject, description, activeForm string) *Task {
	if m.store == nil {
		return &Task{
			ID:          "pending",
			Subject:     subject,
			Description: description,
			Status:      StatusPending,
			CreatedAt:   "now",
			UpdatedAt:   "now",
		}
	}

	storeTask, err := m.store.Create(subject, description)
	if err != nil {
		return &Task{
			ID:          "error",
			Subject:     subject,
			Description: fmt.Sprintf("Error creating task: %v", err),
			Status:      StatusPending,
		}
	}

	if activeForm != "" {
		_ = m.store.Update(storeTask.ID, map[string]any{"activeForm": activeForm})
		storeTask, _ = m.store.Get(storeTask.ID)
	}

	return toTask(storeTask)
}

// Get retrieves a task by ID.
func (m *Manager) Get(id string) (*Task, bool) {
	if m.store == nil {
		return nil, false
	}

	storeTask, ok := m.store.Get(id)
	if !ok {
		return nil, false
	}
	return toTask(storeTask), true
}

// List returns all tasks.
func (m *Manager) List() []*Task {
	if m.store == nil {
		return []*Task{}
	}

	storeTasks := m.store.List("", "")
	tasks := make([]*Task, len(storeTasks))
	for i, st := range storeTasks {
		tasks[i] = toTask(st)
	}
	return tasks
}

// Update updates a task with the given function.
func (m *Manager) Update(id string, fn func(*Task)) error {
	if m.store == nil {
		return fmt.Errorf("task store not initialized")
	}

	task, ok := m.store.Get(id)
	if !ok {
		return fmt.Errorf("task not found: %s", id)
	}

	// Convert to manager Task type
	mgrTask := toTask(task)

	// Apply user function
	fn(mgrTask)

	// Convert updates back to store
	updates := make(map[string]any)
	if mgrTask.Subject != "" && mgrTask.Subject != task.Subject {
		updates["subject"] = mgrTask.Subject
	}
	if mgrTask.Description != "" && mgrTask.Description != task.Description {
		updates["description"] = mgrTask.Description
	}
	if string(mgrTask.Status) != string(task.Status) {
		// Convert to taskstore.TaskStatus
		updates["status"] = taskstore.TaskStatus(mgrTask.Status)
	}
	if mgrTask.ActiveForm != "" && mgrTask.ActiveForm != task.ActiveForm {
		updates["activeForm"] = mgrTask.ActiveForm
	}

	return m.store.Update(id, updates)
}

// Delete deletes a task.
func (m *Manager) Delete(id string) error {
	if m.store == nil {
		return fmt.Errorf("task store not initialized")
	}
	return m.store.Delete(id)
}

// AddBlock adds a blocking relationship (task blocks otherTaskID).
func (m *Manager) AddBlock(taskID, otherTaskID string) {
	if m.store == nil {
		return
	}

	task, ok := m.store.Get(taskID)
	if ok {
		blocks := append(task.Blocks, otherTaskID)
		_ = m.store.SetBlocks(taskID, blocks)
	}

	// Add reverse dependency
	otherTask, ok := m.store.Get(otherTaskID)
	if ok {
		blockedBy := append(otherTask.BlockedBy, taskID)
		_ = m.store.SetBlockedBy(otherTaskID, blockedBy)
	}
}

// AddBlockedBy adds a dependency relationship (task depends on otherTaskID).
func (m *Manager) AddBlockedBy(taskID, otherTaskID string) {
	if m.store == nil {
		return
	}

	task, ok := m.store.Get(taskID)
	if ok {
		blockedBy := append(task.BlockedBy, otherTaskID)
		_ = m.store.SetBlockedBy(taskID, blockedBy)
	}

	// Add reverse dependency
	otherTask, ok := m.store.Get(otherTaskID)
	if ok {
		blocks := append(otherTask.Blocks, taskID)
		_ = m.store.SetBlocks(otherTaskID, blocks)
	}
}

// SetOwner sets the owner of a task.
func (m *Manager) SetOwner(taskID, owner string) error {
	if m.store == nil {
		return fmt.Errorf("task store not initialized")
	}
	return m.store.SetMetadata(taskID, map[string]any{"owner": owner})
}

// toTask converts a taskstore.Task to a manager Task.
func toTask(st *taskstore.Task) *Task {
	metadata := st.Metadata
	owner := ""
	if metadata != nil {
		if o, ok := metadata["owner"].(string); ok {
			owner = o
		}
	}

	return &Task{
		ID:          st.ID,
		Subject:     st.Subject,
		Description: st.Description,
		Status:      TaskStatus(st.Status),
		ActiveForm:  st.ActiveForm,
		Owner:       owner,
		CreatedAt:   st.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   st.UpdatedAt.Format("2006-01-02 15:04:05"),
		Blocks:      st.Blocks,
		BlockedBy:   st.BlockedBy,
		Metadata:    metadata,
	}
}
