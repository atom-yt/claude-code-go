package task

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestManager(t *testing.T) (*Manager, func()) {
	tmpDir := t.TempDir()

	// Initialize the global manager
	err := Initialize(tmpDir)
	require.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return GetManager(), cleanup
}

func TestManager_Create(t *testing.T) {
	m, _ := setupTestManager(t)
	task := m.Create("Test Task", "Test Description", "Testing")

	assert.Equal(t, "Test Task", task.Subject)
	assert.Equal(t, "Test Description", task.Description)
	assert.Equal(t, "Testing", task.ActiveForm)
	assert.Equal(t, StatusPending, task.Status)
	assert.Equal(t, "task-1", task.ID)
}

func TestManager_Get(t *testing.T) {
	m, _ := setupTestManager(t)
	task := m.Create("Test", "Desc", "Act")

	retrieved, ok := m.Get(task.ID)
	assert.True(t, ok)
	assert.Equal(t, task.ID, retrieved.ID)
	assert.Equal(t, "Test", retrieved.Subject)

	_, ok = m.Get("999")
	assert.False(t, ok)
}

func TestManager_List(t *testing.T) {
	m, _ := setupTestManager(t)
	m.Create("Task 1", "Desc 1", "")
	m.Create("Task 2", "Desc 2", "")

	tasks := m.List()
	assert.Len(t, tasks, 2)
}

func TestManager_Update(t *testing.T) {
	m, _ := setupTestManager(t)
	task := m.Create("Original", "Desc", "")

	err := m.Update(task.ID, func(t *Task) {
		t.Subject = "Updated"
		t.Status = StatusCompleted
	})

	require.NoError(t, err)

	updated, ok := m.Get(task.ID)
	require.True(t, ok)
	assert.Equal(t, "Updated", updated.Subject)
	assert.Equal(t, StatusCompleted, updated.Status)
}

func TestManager_Delete(t *testing.T) {
	m, _ := setupTestManager(t)
	task := m.Create("To Delete", "Desc", "")

	err := m.Delete(task.ID)
	require.NoError(t, err)

	_, ok := m.Get(task.ID)
	assert.False(t, ok)
}

func TestManager_AddBlock(t *testing.T) {
	m, _ := setupTestManager(t)
	task1 := m.Create("Task 1", "Desc 1", "")
	task2 := m.Create("Task 2", "Desc 2", "")

	m.AddBlock(task1.ID, task2.ID)

	t1, ok := m.Get(task1.ID)
	require.True(t, ok)
	assert.Contains(t, t1.Blocks, task2.ID)

	t2, ok := m.Get(task2.ID)
	require.True(t, ok)
	assert.Contains(t, t2.BlockedBy, task1.ID)
}

func TestManager_AddBlockedBy(t *testing.T) {
	m, _ := setupTestManager(t)
	task1 := m.Create("Task 1", "Desc 1", "")
	task2 := m.Create("Task 2", "Desc 2", "")

	m.AddBlockedBy(task2.ID, task1.ID)

	t1, ok := m.Get(task1.ID)
	require.True(t, ok)
	assert.Contains(t, t1.Blocks, task2.ID)

	t2, ok := m.Get(task2.ID)
	require.True(t, ok)
	assert.Contains(t, t2.BlockedBy, task1.ID)
}

func TestManager_SetOwner(t *testing.T) {
	m, _ := setupTestManager(t)
	task := m.Create("Task", "Desc", "")

	err := m.SetOwner(task.ID, "agent-1")
	require.NoError(t, err)

	updated, ok := m.Get(task.ID)
	require.True(t, ok)
	assert.Equal(t, "agent-1", updated.Owner)
}

func TestCreateTool(t *testing.T) {
	m, _ := setupTestManager(t)
	subject := "Fix authentication bug"
	description := "The login flow is failing for users with special characters in their password."
	activeForm := "Fixing authentication bug"

	task := m.Create(subject, description, activeForm)

	assert.Equal(t, "task-1", task.ID)
	assert.Equal(t, subject, task.Subject)
	assert.Equal(t, description, task.Description)
	assert.Equal(t, activeForm, task.ActiveForm)
	assert.Equal(t, StatusPending, task.Status)
}

func TestGetTool(t *testing.T) {
	m, _ := setupTestManager(t)
	task := m.Create("Test Task", "Test Description", "")

	retrieved, ok := m.Get(task.ID)
	assert.True(t, ok)
	assert.Equal(t, task.ID, retrieved.ID)
	assert.Equal(t, "Test Task", retrieved.Subject)
}
