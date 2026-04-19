package task

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_Create(t *testing.T) {
	m := &TaskManager{tasks: make(map[string]*Task), nextID: 1}
	task := m.Create("Test Task", "Test Description", "Testing")

	assert.Equal(t, "Test Task", task.Subject)
	assert.Equal(t, "Test Description", task.Description)
	assert.Equal(t, "Testing", task.ActiveForm)
	assert.Equal(t, StatusPending, string(task.Status))
	assert.Equal(t, "1", task.ID)
}

func TestManager_Get(t *testing.T) {
	m := &TaskManager{tasks: make(map[string]*Task), nextID: 1}
	task := m.Create("Test", "Desc", "Act")

	retrieved, ok := m.Get(task.ID)
	assert.True(t, ok)
	assert.Equal(t, task.ID, retrieved.ID)
	assert.Equal(t, "Test", retrieved.Subject)

	_, ok = m.Get("999")
	assert.False(t, ok)
}

func TestManager_List(t *testing.T) {
	m := &TaskManager{tasks: make(map[string]*Task), nextID: 1}
	m.Create("Task 1", "Desc 1", "")
	m.Create("Task 2", "Desc 2", "")

	tasks := m.List()
	assert.Len(t, tasks, 2)
}

func TestManager_Update(t *testing.T) {
	m := &TaskManager{tasks: make(map[string]*Task), nextID: 1}
	task := m.Create("Original", "Desc", "")

	updated, err := m.Update(task.ID, func(t *Task) {
		t.Subject = "Updated"
		t.Status = StatusCompleted
	})

	assert.NoError(t, err)
	assert.Equal(t, "Updated", updated.Subject)
	assert.Equal(t, StatusCompleted, string(updated.Status))
}

func TestManager_UpdateNotFound(t *testing.T) {
	m := &TaskManager{tasks: make(map[string]*Task), nextID: 1}
	_, err := m.Update("999", func(t *Task) {})
	assert.Error(t, err)
}

func TestManager_Delete(t *testing.T) {
	m := &TaskManager{tasks: make(map[string]*Task), nextID: 1}
	task := m.Create("Test", "Desc", "")

	err := m.Delete(task.ID)
	assert.NoError(t, err)

	_, ok := m.Get(task.ID)
	assert.False(t, ok)
}

func TestManager_DeleteNotFound(t *testing.T) {
	m := &TaskManager{tasks: make(map[string]*Task), nextID: 1}
	err := m.Delete("999")
	assert.Error(t, err)
}

func TestManager_AddBlock(t *testing.T) {
	m := &TaskManager{tasks: make(map[string]*Task), nextID: 1}
	task1 := m.Create("Task 1", "Desc", "")
	task2 := m.Create("Task 2", "Desc", "")

	m.AddBlock(task1.ID, task2.ID)

	updated1, _ := m.Get(task1.ID)
	assert.Contains(t, updated1.Blocks, task2.ID)

	updated2, _ := m.Get(task2.ID)
	assert.Contains(t, updated2.BlockedBy, task1.ID)
}

func TestTaskCreateTool(t *testing.T) {
	tool := &TaskCreateTool{}

	// Reset global manager
	*globalManager = TaskManager{tasks: make(map[string]*Task), nextID: 1}

	result, err := tool.Call(context.Background(), map[string]any{
		"subject":     "Test Task",
		"description": "Test Description",
		"activeForm":  "Testing",
	})

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Task #1 created")
	assert.Contains(t, result.Output, "Test Task")
}

func TestTaskCreateTool_MissingSubject(t *testing.T) {
	tool := &TaskCreateTool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"description": "Test",
	})

	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Output, "subject is required")
}

func TestTaskGetTool(t *testing.T) {
	tool := &TaskGetTool{}

	// Reset global manager
	*globalManager = TaskManager{tasks: make(map[string]*Task), nextID: 1}
	task := globalManager.Create("Test", "Description", "")

	result, err := tool.Call(context.Background(), map[string]any{
		"taskId": task.ID,
	})

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Task #"+task.ID)
	assert.Contains(t, result.Output, "Test")
}

func TaskListTool_Test(t *testing.T) {
	tool := &TaskListTool{}

	// Reset global manager
	*globalManager = TaskManager{tasks: make(map[string]*Task), nextID: 1}
	globalManager.Create("Task 1", "Desc", "")
	globalManager.Create("Task 2", "Desc", "")

	result, err := tool.Call(context.Background(), map[string]any{})

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Task List:")
	assert.Contains(t, result.Output, "Total: 2 tasks")
}

func TestTaskUpdateTool(t *testing.T) {
	tool := &TaskUpdateTool{}

	// Reset global manager
	*globalManager = TaskManager{tasks: make(map[string]*Task), nextID: 1}
	task := globalManager.Create("Original", "Desc", "")

	result, err := tool.Call(context.Background(), map[string]any{
		"taskId": task.ID,
		"status": "completed",
		"subject": "Updated",
	})

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Task #"+task.ID)
	assert.Contains(t, result.Output, "completed")
}

func TestTaskDeleteTool(t *testing.T) {
	tool := &TaskDeleteTool{}

	// Reset global manager
	*globalManager = TaskManager{tasks: make(map[string]*Task), nextID: 1}
	task := globalManager.Create("Test", "Desc", "")

	result, err := tool.Call(context.Background(), map[string]any{
		"taskId": task.ID,
	})

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "deleted")

	_, ok := globalManager.Get(task.ID)
	assert.False(t, ok)
}

func TestTaskOutputTool(t *testing.T) {
	tool := &TaskOutputTool{}

	// Reset global manager
	*globalManager = TaskManager{tasks: make(map[string]*Task), nextID: 1}
	task := globalManager.Create("Test", "Desc", "")
	task.Status = StatusCompleted

	result, err := tool.Call(context.Background(), map[string]any{
		"taskId": task.ID,
	})

	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Task #"+task.ID)
}

func TestFormatTaskListSummary(t *testing.T) {
	*globalManager = TaskManager{tasks: make(map[string]*Task), nextID: 1}
	globalManager.Create("Pending Task", "Desc", "")
	task2 := globalManager.Create("In Progress Task", "Desc", "")
	task2.Status = StatusInProgress
	task3 := globalManager.Create("Completed Task", "Desc", "")
	task3.Status = StatusCompleted

	result := formatTaskListSummary(globalManager.List())
	assert.Contains(t, result, "○ [#1]")
	assert.Contains(t, result, "◐ [#2]")
	assert.Contains(t, result, "● [#3]")
	assert.Contains(t, result, "Total: 3 tasks")
}

func TestFormatTaskDetail(t *testing.T) {
	task := &Task{
		ID:          "1",
		Subject:     "Test Task",
		Description: "Test Description",
		Status:      StatusPending,
		Owner:       "agent-1",
		Blocks:      []string{"2", "3"},
		BlockedBy:   []string{"4"},
	}

	result := formatTaskDetail(task)
	assert.Contains(t, result, "Task #1: Test Task")
	assert.Contains(t, result, "Status: pending")
	assert.Contains(t, result, "Owner: agent-1")
	assert.Contains(t, result, "Blocks: 2, 3")
	assert.Contains(t, result, "Blocked by: 4")
	assert.Contains(t, result, "Test Description")
}
