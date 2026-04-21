package taskstore

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestStore(t *testing.T) (*Store, string) {
	tmpDir := t.TempDir()
	store, err := New(tmpDir)
	require.NoError(t, err)
	return store, tmpDir
}

func TestNew(t *testing.T) {
	store, tmpDir := setupTestStore(t)
	defer func() {
		os.RemoveAll(tmpDir)
	}()

	assert.NotNil(t, store)
	assert.Empty(t, store.List("", ""))
	assert.Equal(t, 1, store.nextID)
}

func TestNew_LoadsExisting(t *testing.T) {
	tmpDir := t.TempDir()
	// Create an existing tasks file
	tasksFile := filepath.Join(tmpDir, ".claude", "tasks.json")
	err := os.MkdirAll(filepath.Dir(tasksFile), 0755)
	require.NoError(t, err)

	existingData := `{
		"tasks": {
			"task-1": {
				"id": "task-1",
				"subject": "Test Task",
				"description": "Test description",
				"status": "pending",
				"createdAt": "2024-01-01T00:00:00Z",
				"updatedAt": "2024-01-01T00:00:00Z"
			}
		},
		"nextId": 5
	}`
	err = os.WriteFile(tasksFile, []byte(existingData), 0644)
	require.NoError(t, err)

	store, err := New(tmpDir)
	require.NoError(t, err)

	task, ok := store.Get("task-1")
	require.True(t, ok)
	assert.Equal(t, "Test Task", task.Subject)
	assert.Equal(t, 5, store.nextID)
}

func TestCreate(t *testing.T) {
	store, _ := setupTestStore(t)

	task, err := store.Create("Test Subject", "Test Description")
	require.NoError(t, err)

	assert.Equal(t, "task-1", task.ID)
	assert.Equal(t, "Test Subject", task.Subject)
	assert.Equal(t, "Test Description", task.Description)
	assert.Equal(t, StatusPending, task.Status)
	assert.False(t, task.CreatedAt.IsZero())
	assert.False(t, task.UpdatedAt.IsZero())

	// Verify it was persisted
	store2, _ := New(t.TempDir())
	task2, err := store2.Create("Another Task", "Description")
	require.NoError(t, err)
	assert.Equal(t, "task-1", task2.ID)
}

func TestGet(t *testing.T) {
	store, _ := setupTestStore(t)

	task, err := store.Create("Test", "Description")
	require.NoError(t, err)

	found, ok := store.Get(task.ID)
	require.True(t, ok)
	assert.Equal(t, task.ID, found.ID)

	_, ok = store.Get("non-existent")
	assert.False(t, ok)
}

func TestList(t *testing.T) {
	store, _ := setupTestStore(t)

	task1, _ := store.Create("Task 1", "Desc 1")
	task2, _ := store.Create("Task 2", "Desc 2")
	_ = store.Update(task1.ID, map[string]any{"status": StatusCompleted})

	allTasks := store.List("", "")
	assert.Len(t, allTasks, 2)

	pendingTasks := store.List(StatusPending, "")
	assert.Len(t, pendingTasks, 1)
	assert.Equal(t, task2.ID, pendingTasks[0].ID)

	completedTasks := store.List(StatusCompleted, "")
	assert.Len(t, completedTasks, 1)
	assert.Equal(t, task1.ID, completedTasks[0].ID)
}

func TestList_SessionScoped(t *testing.T) {
	store, _ := setupTestStore(t)

	task1, _ := store.Create("Task 1", "Desc 1")
	task2, _ := store.Create("Task 2", "Desc 2")
	_ = store.SetSessionID(task1.ID, "session-1")
	_ = store.SetSessionID(task2.ID, "session-2")

	session1Tasks := store.List("", "session-1")
	assert.Len(t, session1Tasks, 1)
	assert.Equal(t, task1.ID, session1Tasks[0].ID)
}

func TestUpdate(t *testing.T) {
	store, _ := setupTestStore(t)

	task, _ := store.Create("Original", "Description")
	err := store.Update(task.ID, map[string]any{
		"subject":     "Updated",
		"description": "New description",
		"status":      StatusCompleted,
		"activeForm":  "Completing",
	})
	require.NoError(t, err)

	updated, ok := store.Get(task.ID)
	require.True(t, ok)
	assert.Equal(t, "Updated", updated.Subject)
	assert.Equal(t, "New description", updated.Description)
	assert.Equal(t, StatusCompleted, updated.Status)
	assert.Equal(t, "Completing", updated.ActiveForm)
	assert.NotNil(t, updated.CompletedAt)
}

func TestDelete(t *testing.T) {
	store, _ := setupTestStore(t)

	task, _ := store.Create("To Delete", "Description")
	err := store.Delete(task.ID)
	require.NoError(t, err)

	_, ok := store.Get(task.ID)
	assert.False(t, ok)

	// Should not appear in list
	tasks := store.List("", "")
	assert.Len(t, tasks, 0)
}

func TestSetBlockedBy(t *testing.T) {
	store, _ := setupTestStore(t)

	task1, _ := store.Create("Task 1", "Desc 1")
	task2, _ := store.Create("Task 2", "Desc 2")

	err := store.SetBlockedBy(task2.ID, []string{task1.ID})
	require.NoError(t, err)

	t2, ok := store.Get(task2.ID)
	require.True(t, ok)
	assert.Len(t, t2.BlockedBy, 1)
	assert.Equal(t, task1.ID, t2.BlockedBy[0])
}

func TestSetBlocks(t *testing.T) {
	store, _ := setupTestStore(t)

	task1, _ := store.Create("Task 1", "Desc 1")
	task2, _ := store.Create("Task 2", "Desc 2")

	err := store.SetBlocks(task1.ID, []string{task2.ID})
	require.NoError(t, err)

	t1, ok := store.Get(task1.ID)
	require.True(t, ok)
	assert.Len(t, t1.Blocks, 1)
	assert.Equal(t, task2.ID, t1.Blocks[0])
}

func TestGetPendingTasks(t *testing.T) {
	store, _ := setupTestStore(t)

	task1, _ := store.Create("Pending Task 1", "Desc 1")
	task2, _ := store.Create("Pending Task 2", "Desc 2")
	task3, _ := store.Create("Pending Task 3", "Desc 3")
	task4, _ := store.Create("Completed Task", "Desc 4")

	// Mark task1 as blocker for task2
	_ = store.SetBlockedBy(task2.ID, []string{task1.ID})
	// Mark task4 as completed
	_ = store.Update(task4.ID, map[string]any{"status": StatusCompleted})

	pending := store.GetPendingTasks()
	assert.Len(t, pending, 2)
	ids := []string{pending[0].ID, pending[1].ID}
	assert.Contains(t, ids, task1.ID)
	assert.Contains(t, ids, task3.ID)
}

func TestGetPendingTasks_WithCompletedBlocker(t *testing.T) {
	store, _ := setupTestStore(t)

	task1, _ := store.Create("Blocker Task", "Desc 1")
	task2, _ := store.Create("Blocked Task", "Desc 2")

	// task2 is blocked by task1
	_ = store.SetBlockedBy(task2.ID, []string{task1.ID})

	// Initially, only task1 is available
	pending := store.GetPendingTasks()
	assert.Len(t, pending, 1)
	assert.Equal(t, task1.ID, pending[0].ID)

	// Complete task1
	_ = store.Update(task1.ID, map[string]any{"status": StatusCompleted})

	// Now task2 should be available
	pending = store.GetPendingTasks()
	assert.Len(t, pending, 1)
	assert.Equal(t, task2.ID, pending[0].ID)
}

func TestGetBlockedTasks(t *testing.T) {
	store, _ := setupTestStore(t)

	task1, _ := store.Create("Unblocked Task", "Desc 1")
	task2, _ := store.Create("Blocked Task", "Desc 2")
	task3, _ := store.Create("Also Blocked", "Desc 3")

	_ = store.SetBlockedBy(task2.ID, []string{task1.ID})
	_ = store.SetBlockedBy(task3.ID, []string{task1.ID})

	blocked := store.GetBlockedTasks()
	assert.Len(t, blocked, 2)
	ids := []string{blocked[0].ID, blocked[1].ID}
	assert.Contains(t, ids, task2.ID)
	assert.Contains(t, ids, task3.ID)
}

func TestGetInProgressTasks(t *testing.T) {
	store, _ := setupTestStore(t)

	task1, _ := store.Create("Task 1", "Desc 1")
	task2, _ := store.Create("Task 2", "Desc 2")

	_ = store.Update(task1.ID, map[string]any{"status": StatusInProgress})
	_ = store.Update(task2.ID, map[string]any{"status": StatusCompleted})

	inProgress := store.GetInProgressTasks()
	assert.Len(t, inProgress, 1)
	assert.Equal(t, task1.ID, inProgress[0].ID)
}

func TestGetCompletedTasks(t *testing.T) {
	store, _ := setupTestStore(t)

	task1, _ := store.Create("Task 1", "Desc 1")
	task2, _ := store.Create("Task 2", "Desc 2")

	_ = store.Update(task1.ID, map[string]any{"status": StatusCompleted})
	_ = store.Update(task2.ID, map[string]any{"status": StatusInProgress})

	completed := store.GetCompletedTasks()
	assert.Len(t, completed, 1)
	assert.Equal(t, task1.ID, completed[0].ID)
}

func TestSetMetadata(t *testing.T) {
	store, _ := setupTestStore(t)

	task, _ := store.Create("Task", "Description")

	err := store.SetMetadata(task.ID, map[string]any{
		"key1": "value1",
		"key2": 123,
	})
	require.NoError(t, err)

	t1, ok := store.Get(task.ID)
	require.True(t, ok)
	assert.Equal(t, "value1", t1.Metadata["key1"])
	assert.Equal(t, 123, t1.Metadata["key2"])

	// Update with deletion
	err = store.SetMetadata(task.ID, map[string]any{
		"key1": nil,      // delete
		"key2": "updated", // update
		"key3": "new",    // add
	})
	require.NoError(t, err)

	t1, ok = store.Get(task.ID)
	require.True(t, ok)
	_, exists := t1.Metadata["key1"]
	assert.False(t, exists)
	assert.Equal(t, "updated", t1.Metadata["key2"])
	assert.Equal(t, "new", t1.Metadata["key3"])
}

func TestPersistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a store and add a task
	store1, err := New(tmpDir)
	require.NoError(t, err)

	task1, err := store1.Create("Task 1", "Description 1")
	require.NoError(t, err)

	_ = store1.SetBlockedBy(task1.ID, []string{"task-99"})

	// Create a new store instance
	store2, err := New(tmpDir)
	require.NoError(t, err)

	// Verify task was persisted
	task2, ok := store2.Get(task1.ID)
	require.True(t, ok)
	assert.Equal(t, "Task 1", task2.Subject)
	assert.Equal(t, "Description 1", task2.Description)
	assert.Equal(t, []string{"task-99"}, task2.BlockedBy)

	// Verify nextID was persisted
	assert.Equal(t, 2, store2.nextID)
}

func TestConcurrentAccess(t *testing.T) {
	store, _ := setupTestStore(t)

	// Concurrent creates
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			store.Create(fmt.Sprintf("Task %d", i), "Description")
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	tasks := store.List("", "")
	assert.Len(t, tasks, 10)
}
