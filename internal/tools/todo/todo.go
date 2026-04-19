// Package todo implements the TodoWrite tool for task list management.
// This tool helps track progress on complex multi-step tasks.
package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/atom-yt/claude-code-go/internal/tools"
)

// Task status constants
const (
	StatusPending    = "pending"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
)

// Task represents a single todo item.
type Task struct {
	ID          string `json:"id"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Status      string `json:"status"`
	ActiveForm  string `json:"activeForm,omitempty"`
	Owner       string `json:"owner,omitempty"`
}

// TaskList manages the global task list.
type TaskList struct {
	mu     sync.RWMutex
	tasks  map[string]*Task
	nextID int
}

// Global task list instance
var globalTaskList = &TaskList{
	tasks:  make(map[string]*Task),
	nextID: 1,
}

// Tool implements the TodoWrite tool.
type Tool struct{}

var _ tools.Tool = (*Tool)(nil)

func (t *Tool) Name() string           { return "TodoWrite" }
func (t *Tool) IsReadOnly() bool        { return false }
func (t *Tool) IsConcurrencySafe() bool { return false } // Stateful, needs serialization

func (t *Tool) Description() string {
	return "Use this tool to create a structured task list for your current coding session. " +
		"This helps you track progress, organize complex tasks, and demonstrate thoroughness to the user. " +
		"It also helps the user understand the progress of the task and overall progress of their requests."
}

func (t *Tool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"todos": map[string]any{
				"type":        "array",
				"description": "List of todo items. When updating an existing todo, the ID should match. For new todos, omit the ID or use empty string.",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{
							"type":        "string",
							"description": "Unique identifier for the todo. Auto-generated if not provided.",
						},
						"subject": map[string]any{
							"type":        "string",
							"description": "Brief title for the todo (imperative form, e.g., 'Fix authentication bug')",
						},
						"description": map[string]any{
							"type":        "string",
							"description": "Detailed description of what needs to be done",
						},
						"status": map[string]any{
							"type":        "string",
							"enum":        []string{"pending", "in_progress", "completed"},
							"description": "Status of the todo item",
						},
						"activeForm": map[string]any{
							"type":        "string",
							"description": "Present continuous form shown in spinner when in_progress (e.g., 'Running tests')",
						},
						"owner": map[string]any{
							"type":        "string",
							"description": "Agent ID if assigned to a sub-agent",
						},
					},
					"required": []string{"subject", "description", "status"},
				},
			},
		},
		"required": []string{"todos"},
	}
}

func (t *Tool) Call(ctx context.Context, input map[string]any) (tools.ToolResult, error) {
	todosRaw, ok := input["todos"]
	if !ok {
		return tools.ToolResult{Output: "Error: todos parameter is required", IsError: true}, nil
	}

	todos, ok := todosRaw.([]any)
	if !ok {
		return tools.ToolResult{Output: "Error: todos must be an array", IsError: true}, nil
	}

	globalTaskList.mu.Lock()
	defer globalTaskList.mu.Unlock()

	// Clear existing tasks and process new ones
	globalTaskList.tasks = make(map[string]*Task)
	globalTaskList.nextID = 1

	for _, todoRaw := range todos {
		todoMap, ok := todoRaw.(map[string]any)
		if !ok {
			continue
		}

		task := &Task{}

		// Parse fields
		if id, ok := todoMap["id"].(string); ok && id != "" {
			task.ID = id
		} else {
			task.ID = fmt.Sprintf("%d", globalTaskList.nextID)
			globalTaskList.nextID++
		}

		if subject, ok := todoMap["subject"].(string); ok {
			task.Subject = subject
		}
		if desc, ok := todoMap["description"].(string); ok {
			task.Description = desc
		}
		if status, ok := todoMap["status"].(string); ok {
			switch status {
			case StatusPending, StatusInProgress, StatusCompleted:
				task.Status = status
			default:
				task.Status = StatusPending
			}
		}
		if activeForm, ok := todoMap["activeForm"].(string); ok {
			task.ActiveForm = activeForm
		}
		if owner, ok := todoMap["owner"].(string); ok {
			task.Owner = owner
		}

		globalTaskList.tasks[task.ID] = task
	}

	// Format output
	return tools.ToolResult{Output: formatTaskList(globalTaskList.tasks)}, nil
}

// formatTaskList formats the task list for display
func formatTaskList(tasks map[string]*Task) string {
	if len(tasks) == 0 {
		return "Task list is empty."
	}

	var sb strings.Builder
	sb.WriteString("Task List:\n")
	sb.WriteString("───────────────────────────────────────────────────────────────────────\n")

	// Order: pending first, then in_progress, then completed
	order := []string{StatusInProgress, StatusPending, StatusCompleted}
	for _, status := range order {
		for _, task := range tasks {
			if task.Status != status {
				continue
			}

			var icon string
			switch task.Status {
			case StatusPending:
				icon = "○"
			case StatusInProgress:
				icon = "◐"
			case StatusCompleted:
				icon = "●"
			}

			sb.WriteString(fmt.Sprintf("%s [%s] %s\n", icon, task.ID, task.Subject))
			if task.Description != "" {
				sb.WriteString(fmt.Sprintf("  %s\n", task.Description))
			}
		}
	}

	sb.WriteString("───────────────────────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Total: %d tasks", len(tasks)))

	return sb.String()
}

// GetTaskList returns a copy of the current task list (for external use)
func GetTaskList() []Task {
	globalTaskList.mu.RLock()
	defer globalTaskList.mu.RUnlock()

	tasks := make([]Task, 0, len(globalTaskList.tasks))
	for _, t := range globalTaskList.tasks {
		tasks = append(tasks, *t)
	}
	return tasks
}

// MarshalJSON implements json.Marshaler for Task
func (t *Task) MarshalJSON() ([]byte, error) {
	type Alias Task
	return json.Marshal((*Alias)(t))
}