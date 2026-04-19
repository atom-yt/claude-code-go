package todo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTodoTool_Name(t *testing.T) {
	tool := &Tool{}
	assert.Equal(t, "TodoWrite", tool.Name())
}

func TestTodoTool_NotReadOnly(t *testing.T) {
	tool := &Tool{}
	assert.False(t, tool.IsReadOnly())
}

func TestTodoTool_NotConcurrencySafe(t *testing.T) {
	tool := &Tool{}
	assert.False(t, tool.IsConcurrencySafe())
}

func TestTodoTool_MissingTodos(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Output, "todos parameter is required")
}

func TestTodoTool_InvalidTodos(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"todos": "not an array",
	})
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Output, "must be an array")
}

func TestTodoTool_CreateTasks(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"todos": []any{
			map[string]any{
				"subject":     "Task 1",
				"description": "First task",
				"status":      "pending",
			},
			map[string]any{
				"subject":     "Task 2",
				"description": "Second task",
				"status":      "in_progress",
				"activeForm":  "Working on task 2",
			},
		},
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Task List:")
	assert.Contains(t, result.Output, "Task 1")
	assert.Contains(t, result.Output, "Task 2")
	assert.Contains(t, result.Output, "Total: 2 tasks")
}

func TestTodoTool_UpdateTaskStatus(t *testing.T) {
	tool := &Tool{}

	// Create initial task
	_, _ = tool.Call(context.Background(), map[string]any{
		"todos": []any{
			map[string]any{
				"subject":     "Task 1",
				"description": "First task",
				"status":      "pending",
			},
		},
	})

	// Update to completed
	result, err := tool.Call(context.Background(), map[string]any{
		"todos": []any{
			map[string]any{
				"id":          "1",
				"subject":     "Task 1",
				"description": "First task",
				"status":      "completed",
			},
		},
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "●")
}

func TestTodoTool_GetTaskList(t *testing.T) {
	tool := &Tool{}
	_, _ = tool.Call(context.Background(), map[string]any{
		"todos": []any{
			map[string]any{
				"subject":     "Test Task",
				"description": "Testing",
				"status":      "pending",
			},
		},
	})

	tasks := GetTaskList()
	assert.Len(t, tasks, 1)
	assert.Equal(t, "Test Task", tasks[0].Subject)
}

func TestTodoTool_AllStatuses(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"todos": []any{
			map[string]any{
				"subject":     "Pending Task",
				"description": "Not started",
				"status":      "pending",
			},
			map[string]any{
				"subject":     "In Progress Task",
				"description": "Working on it",
				"status":      "in_progress",
			},
			map[string]any{
				"subject":     "Completed Task",
				"description": "Done",
				"status":      "completed",
			},
		},
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "○ [1] Pending Task")
	assert.Contains(t, result.Output, "◐ [2] In Progress Task")
	assert.Contains(t, result.Output, "● [3] Completed Task")
}

func TestTodoTool_InvalidStatusDefaultsToPending(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"todos": []any{
			map[string]any{
				"subject":     "Task",
				"description": "With invalid status",
				"status":      "invalid_status",
			},
		},
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	tasks := GetTaskList()
	assert.Equal(t, StatusPending, tasks[0].Status)
}

func TestTodoTool_EmptyList(t *testing.T) {
	tool := &Tool{}
	result, err := tool.Call(context.Background(), map[string]any{
		"todos": []any{},
	})
	assert.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Output, "Task list is empty")
}
